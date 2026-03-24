package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/r3st/turnscore/config"
	"github.com/r3st/turnscore/internal/domain"
	"github.com/r3st/turnscore/pkg/mocks"
)

const testTournamentSlug = "essen-open-2025"

func newTestAuthService(
	t *testing.T,
	userRepo domain.UserRepository,
	raterRepo domain.RaterRepository,
	tourneyRepo domain.TournamentRepository,
) *AuthService {
	t.Helper()
	jwtSvc := newTestJWTService(t)
	return NewAuthService(
		config.AuthConfig{
			GoogleClientID:     "test-client-id",
			GoogleClientSecret: "test-client-secret",
		},
		jwtSvc,
		userRepo,
		raterRepo,
		tourneyRepo,
	)
}

func TestRaterLoginSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tournamentID := uuid.New()
	raterID := uuid.New()
	votingEnd := time.Now().Add(48 * time.Hour)

	mockTourneyRepo := mocks.NewMockTournamentRepository(ctrl)
	mockRaterRepo := mocks.NewMockRaterRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	mockTourneyRepo.EXPECT().
		FindBySlug(gomock.Any(), testTournamentSlug).
		Return(&domain.Tournament{
			ID:        tournamentID,
			Slug:      testTournamentSlug,
			VotingEnd: &votingEnd,
		}, nil)

	mockRaterRepo.EXPECT().
		FindByNicknameAndCode(gomock.Any(), tournamentID, "Alice", "1234").
		Return(&domain.Rater{
			ID:           raterID,
			TournamentID: tournamentID,
			Nickname:     "Alice",
		}, nil)

	svc := newTestAuthService(t, mockUserRepo, mockRaterRepo, mockTourneyRepo)

	pair, err := svc.RaterLogin(context.Background(), domain.RaterLoginInput{
		TournamentSlug: testTournamentSlug,
		Nickname:       "Alice",
		Code:           "1234",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)

	// Validate the issued token claims
	jwtSvc := newTestJWTService(t)
	claims, err := jwtSvc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, raterID.String(), claims.Subject)
	assert.Equal(t, "rater", claims.Role)
	assert.Equal(t, tournamentID.String(), claims.TournamentID)
}

func TestRaterLoginTournamentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTourneyRepo := mocks.NewMockTournamentRepository(ctrl)
	mockTourneyRepo.EXPECT().
		FindBySlug(gomock.Any(), "nonexistent").
		Return(nil, domain.ErrNotFound)

	svc := newTestAuthService(t,
		mocks.NewMockUserRepository(ctrl),
		mocks.NewMockRaterRepository(ctrl),
		mockTourneyRepo,
	)

	_, err := svc.RaterLogin(context.Background(), domain.RaterLoginInput{
		TournamentSlug: "nonexistent",
		Nickname:       "Alice",
		Code:           "1234",
	})

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestRaterLoginInvalidCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tournamentID := uuid.New()

	mockTourneyRepo := mocks.NewMockTournamentRepository(ctrl)
	mockTourneyRepo.EXPECT().
		FindBySlug(gomock.Any(), testTournamentSlug).
		Return(&domain.Tournament{ID: tournamentID, Slug: testTournamentSlug}, nil)

	mockRaterRepo := mocks.NewMockRaterRepository(ctrl)
	mockRaterRepo.EXPECT().
		FindByNicknameAndCode(gomock.Any(), tournamentID, "Alice", "9999").
		Return(nil, domain.ErrNotFound)

	svc := newTestAuthService(t,
		mocks.NewMockUserRepository(ctrl),
		mockRaterRepo,
		mockTourneyRepo,
	)

	_, err := svc.RaterLogin(context.Background(), domain.RaterLoginInput{
		TournamentSlug: testTournamentSlug,
		Nickname:       "Alice",
		Code:           "9999",
	})

	assert.ErrorIs(t, err, domain.ErrInvalidCode)
}

func TestRaterLoginTokenExpiryWithoutVotingEnd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tournamentID := uuid.New()
	raterID := uuid.New()

	mockTourneyRepo := mocks.NewMockTournamentRepository(ctrl)
	mockTourneyRepo.EXPECT().
		FindBySlug(gomock.Any(), testTournamentSlug).
		Return(&domain.Tournament{
			ID:        tournamentID,
			Slug:      testTournamentSlug,
			VotingEnd: nil,
		}, nil)

	mockRaterRepo := mocks.NewMockRaterRepository(ctrl)
	mockRaterRepo.EXPECT().
		FindByNicknameAndCode(gomock.Any(), tournamentID, "Bob", "5678").
		Return(&domain.Rater{ID: raterID, TournamentID: tournamentID}, nil)

	svc := newTestAuthService(t,
		mocks.NewMockUserRepository(ctrl),
		mockRaterRepo,
		mockTourneyRepo,
	)

	pair, err := svc.RaterLogin(context.Background(), domain.RaterLoginInput{
		TournamentSlug: testTournamentSlug,
		Nickname:       "Bob",
		Code:           "5678",
	})

	require.NoError(t, err)
	assert.InDelta(t, (24 * time.Hour).Seconds(), float64(pair.ExpiresIn), 5)
}

func TestOAuthRedirectURLContainsState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := newTestAuthService(t,
		mocks.NewMockUserRepository(ctrl),
		mocks.NewMockRaterRepository(ctrl),
		mocks.NewMockTournamentRepository(ctrl),
	)

	url := svc.OAuthRedirectURL("my-state-value")
	assert.Contains(t, url, "state=my-state-value")
	assert.Contains(t, url, "accounts.google.com")
}

func TestRefreshTokensSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	rawToken := "some-raw-refresh-token-value-here"
	hashed := hashToken(rawToken)

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		FindByRefreshToken(gomock.Any(), hashed).
		Return(&domain.User{ID: userID}, nil)
	mockUserRepo.EXPECT().
		DeleteRefreshToken(gomock.Any(), hashed).
		Return(nil)
	mockUserRepo.EXPECT().
		SaveRefreshToken(gomock.Any(), userID, gomock.Any(), gomock.Any()).
		Return(nil)

	svc := newTestAuthService(t,
		mockUserRepo,
		mocks.NewMockRaterRepository(ctrl),
		mocks.NewMockTournamentRepository(ctrl),
	)

	pair, err := svc.RefreshTokens(context.Background(), rawToken)

	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.NotEqual(t, rawToken, pair.RefreshToken, "refresh token must be rotated")
}

func TestRefreshTokensInvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().
		FindByRefreshToken(gomock.Any(), gomock.Any()).
		Return(nil, domain.ErrNotFound)

	svc := newTestAuthService(t,
		mockUserRepo,
		mocks.NewMockRaterRepository(ctrl),
		mocks.NewMockTournamentRepository(ctrl),
	)

	_, err := svc.RefreshTokens(context.Background(), "unknown-token")
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestHashTokenIsDeterministic(t *testing.T) {
	h1 := hashToken("my-raw-token")
	h2 := hashToken("my-raw-token")
	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 64) // SHA-256 hex = 64 chars
}

func TestHashTokenDiffersForDiffInputs(t *testing.T) {
	h1 := hashToken("token-a")
	h2 := hashToken("token-b")
	assert.NotEqual(t, h1, h2)
}
