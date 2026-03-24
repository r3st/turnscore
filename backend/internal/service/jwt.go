// Package service contains the business logic for TurnScore.
package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/r3st/turnscore/internal/domain"
)

// jwtClaims is the JWT payload. Minimal — IDs only, no PII.
type jwtClaims struct {
	jwt.RegisteredClaims
	Role         string `json:"role"`
	TournamentID string `json:"tid,omitempty"` // raters only
}

// JWTService handles JWT issuance and validation.
type JWTService struct {
	cfg domain.JWTConfig
}

// NewJWTService creates a JWTService. Panics if the secret is shorter than 32 bytes.
func NewJWTService(cfg domain.JWTConfig) *JWTService {
	if len(cfg.Secret) < 32 {
		panic("JWT_SECRET must be at least 32 characters — generate one with: openssl rand -hex 32")
	}
	return &JWTService{cfg: cfg}
}

// IssueUserTokenPair issues an access + refresh token for an organizer/helper user.
// JWT role is always "user" — tournament-level role is checked per-request in the service layer.
func (s *JWTService) IssueUserTokenPair(userID uuid.UUID) (*domain.TokenPair, string, error) {
	access, err := s.sign(jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Role: "user",
		// ❌ no name, no email, no tournament role
	})
	if err != nil {
		return nil, "", fmt.Errorf("signing access token: %w", err)
	}

	rawRefresh, err := generateOpaqueToken()
	if err != nil {
		return nil, "", fmt.Errorf("generating refresh token: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  access,
		RefreshToken: rawRefresh,
		ExpiresIn:    int(s.cfg.AccessExpiry.Seconds()),
	}, rawRefresh, nil
}

// IssueRaterToken issues a short-lived token for a rater.
// Lifetime is set by the caller (typically voting_end + 24h).
func (s *JWTService) IssueRaterToken(raterID uuid.UUID, tournamentID uuid.UUID, expiry time.Time) (string, error) {
	token, err := s.sign(jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   raterID.String(),
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Role:         "rater",
		TournamentID: tournamentID.String(),
		// ❌ no nickname in JWT
	})
	if err != nil {
		return "", fmt.Errorf("signing rater token: %w", err)
	}
	return token, nil
}

// ValidateClaims parses and validates a JWT, returning its claims.
// Returns domain.ErrUnauthorized for any validation failure.
func (s *JWTService) ValidateClaims(tokenStr string) (*jwtClaims, error) {
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	}, jwt.WithExpirationRequired())

	if err != nil || !token.Valid {
		return nil, domain.ErrUnauthorized
	}
	return claims, nil
}

// sign creates a signed JWT string from the given claims.
func (s *JWTService) sign(claims jwtClaims) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.Secret))
}

// generateOpaqueToken creates a cryptographically random 32-byte hex string
// used as a refresh token. It is stored hashed in the DB.
func generateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
