package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/r3st/turnscore/internal/domain"
)

func newTestJWTService(t *testing.T) *JWTService {
	t.Helper()
	return NewJWTService(domain.JWTConfig{
		Secret:        "test-secret-that-is-at-least-32-chars!!",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	})
}

func TestNewJWTService_PanicsOnShortSecret(t *testing.T) {
	assert.Panics(t, func() {
		NewJWTService(domain.JWTConfig{Secret: "tooshort"})
	})
}

func TestIssueUserTokenPair_ValidToken(t *testing.T) {
	svc := newTestJWTService(t)
	userID := uuid.New()

	pair, rawRefresh, err := svc.IssueUserTokenPair(userID)

	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, rawRefresh)
	assert.Equal(t, rawRefresh, pair.RefreshToken)
	assert.Equal(t, int((15 * time.Minute).Seconds()), pair.ExpiresIn)
}

func TestIssueUserTokenPair_ClaimsAreCorrect(t *testing.T) {
	svc := newTestJWTService(t)
	userID := uuid.New()

	pair, _, err := svc.IssueUserTokenPair(userID)
	require.NoError(t, err)

	claims, err := svc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)

	assert.Equal(t, userID.String(), claims.Subject)
	assert.Equal(t, "user", claims.Role)
	assert.Empty(t, claims.TournamentID, "user tokens must not contain a tournament ID")
}

func TestIssueRaterToken_ClaimsAreCorrect(t *testing.T) {
	svc := newTestJWTService(t)
	raterID := uuid.New()
	tournamentID := uuid.New()
	expiry := time.Now().Add(24 * time.Hour)

	token, err := svc.IssueRaterToken(raterID, tournamentID, expiry)
	require.NoError(t, err)

	claims, err := svc.ValidateClaims(token)
	require.NoError(t, err)

	assert.Equal(t, raterID.String(), claims.Subject)
	assert.Equal(t, "rater", claims.Role)
	assert.Equal(t, tournamentID.String(), claims.TournamentID)
}

func TestValidateClaims_RejectsExpiredToken(t *testing.T) {
	svc := newTestJWTService(t)
	raterID := uuid.New()
	tournamentID := uuid.New()

	// Issue token that already expired
	token, err := svc.IssueRaterToken(raterID, tournamentID, time.Now().Add(-1*time.Minute))
	require.NoError(t, err)

	_, err = svc.ValidateClaims(token)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestValidateClaims_RejectsTamperedToken(t *testing.T) {
	svc := newTestJWTService(t)
	userID := uuid.New()

	pair, _, err := svc.IssueUserTokenPair(userID)
	require.NoError(t, err)

	tampered := pair.AccessToken + "x"
	_, err = svc.ValidateClaims(tampered)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestValidateClaims_RejectsTokenSignedWithDifferentSecret(t *testing.T) {
	svc1 := NewJWTService(domain.JWTConfig{
		Secret:       "secret-one-that-is-at-least-32-chars!!",
		AccessExpiry: 15 * time.Minute,
	})
	svc2 := NewJWTService(domain.JWTConfig{
		Secret:       "secret-two-that-is-at-least-32-chars!!",
		AccessExpiry: 15 * time.Minute,
	})

	pair, _, err := svc1.IssueUserTokenPair(uuid.New())
	require.NoError(t, err)

	_, err = svc2.ValidateClaims(pair.AccessToken)
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestGenerateOpaqueToken_IsUnique(t *testing.T) {
	t1, err := generateOpaqueToken()
	require.NoError(t, err)
	t2, err := generateOpaqueToken()
	require.NoError(t, err)

	assert.NotEqual(t, t1, t2)
	assert.Len(t, t1, 64) // 32 bytes → 64 hex chars
}
