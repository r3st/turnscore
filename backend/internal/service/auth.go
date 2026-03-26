package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/r3st/turnscore/config"
	"github.com/r3st/turnscore/internal/domain"
)

const googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

// AuthService handles Google OAuth and rater login flows.
type AuthService struct {
	jwtSvc      *JWTService
	userRepo    domain.UserRepository
	raterRepo   domain.RaterRepository
	tourneyRepo domain.TournamentRepository
	oauthCfg    *oauth2.Config
}

// NewAuthService wires up the auth service.
func NewAuthService(
	cfg config.AuthConfig,
	jwtSvc *JWTService,
	userRepo domain.UserRepository,
	raterRepo domain.RaterRepository,
	tourneyRepo domain.TournamentRepository,
) *AuthService {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.OAuthCallbackURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
	return &AuthService{
		jwtSvc:      jwtSvc,
		userRepo:    userRepo,
		raterRepo:   raterRepo,
		tourneyRepo: tourneyRepo,
		oauthCfg:    oauthCfg,
	}
}

// OAuthRedirectURL returns the Google OAuth consent screen URL.
func (s *AuthService) OAuthRedirectURL(state string) string {
	return s.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// HandleCallback exchanges the OAuth code for a user session.
// Returns tokens and the raw refresh token (to be stored hashed).
func (s *AuthService) HandleCallback(ctx context.Context, code string) (*domain.TokenPair, error) {
	oauthToken, err := s.oauthCfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: exchanging oauth code", domain.ErrUnauthorized)
	}

	googleUser, err := s.fetchGoogleUserInfo(ctx, oauthToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.UpsertByGoogleSub(ctx,
		googleUser.Sub, googleUser.Email, googleUser.Name, googleUser.Picture)
	if err != nil {
		return nil, fmt.Errorf("upserting user: %w", err)
	}

	tokenPair, rawRefresh, err := s.jwtSvc.IssueUserTokenPair(user.ID)
	if err != nil {
		return nil, err
	}

	// Store hashed refresh token — raw token only leaves the server once
	hashed := hashToken(rawRefresh)
	expiresAt := time.Now().Add(s.jwtSvc.cfg.RefreshExpiry)
	if err := s.userRepo.SaveRefreshToken(ctx, user.ID, hashed, expiresAt); err != nil {
		return nil, fmt.Errorf("saving refresh token: %w", err)
	}

	return tokenPair, nil
}

// RefreshTokens validates a refresh token and issues a new token pair.
func (s *AuthService) RefreshTokens(ctx context.Context, rawRefresh string) (*domain.TokenPair, error) {
	hashed := hashToken(rawRefresh)
	user, err := s.userRepo.FindByRefreshToken(ctx, hashed)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	// Rotate refresh token: invalidate old, issue new
	if err := s.userRepo.DeleteRefreshToken(ctx, hashed); err != nil {
		return nil, fmt.Errorf("deleting old refresh token: %w", err)
	}

	tokenPair, newRawRefresh, err := s.jwtSvc.IssueUserTokenPair(user.ID)
	if err != nil {
		return nil, err
	}

	newHashed := hashToken(newRawRefresh)
	expiresAt := time.Now().Add(s.jwtSvc.cfg.RefreshExpiry)
	if err := s.userRepo.SaveRefreshToken(ctx, user.ID, newHashed, expiresAt); err != nil {
		return nil, fmt.Errorf("saving new refresh token: %w", err)
	}

	return tokenPair, nil
}

// RaterLogin validates nickname + code and issues a rater JWT.
func (s *AuthService) RaterLogin(ctx context.Context, input domain.RaterLoginInput) (*domain.TokenPair, error) {
	tournament, err := s.tourneyRepo.FindBySlug(ctx, input.TournamentSlug)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	rater, err := s.raterRepo.FindByNicknameAndCode(ctx, tournament.ID, input.Nickname, input.Code)
	if err != nil {
		return nil, domain.ErrInvalidCode
	}

	// Token lifetime = voting_end + 24h buffer (or 24h from now if voting_end not set)
	expiry := time.Now().Add(24 * time.Hour)
	if tournament.VotingEnd != nil {
		expiry = tournament.VotingEnd.Add(24 * time.Hour)
	}

	token, err := s.jwtSvc.IssueRaterToken(rater.ID, tournament.ID, expiry)
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken: token,
		ExpiresIn:   int(time.Until(expiry).Seconds()),
	}, nil
}

// fetchGoogleUserInfo calls Google's UserInfo API and returns minimal user data.
// ⚠️ Email and name are only used for DB storage — never in JWT claims.
func (s *AuthService) fetchGoogleUserInfo(ctx context.Context, token *oauth2.Token) (*domain.GoogleUserInfo, error) {
	client := s.oauthCfg.Client(ctx, token)
	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("fetching google userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: google userinfo returned %d", domain.ErrUnauthorized, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return nil, fmt.Errorf("reading google userinfo response: %w", err)
	}

	var raw struct {
		Sub     string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing google userinfo: %w", err)
	}

	return &domain.GoogleUserInfo{
		Sub:     raw.Sub,
		Email:   raw.Email,
		Name:    raw.Name,
		Picture: raw.Picture,
	}, nil
}

// hashToken returns a SHA-256 hex digest of the raw token.
// Refresh tokens are stored hashed — the raw value only leaves the server once.
func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

