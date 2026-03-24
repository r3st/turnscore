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

func TestRaterLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tournamentID := uuid.New()
	raterID := uuid.New()
	votingEnd := time.Now().Add(48 * time.Hour)

	mockTourneyRepo := mocks.NewMockTournamentRepository(ctrl)
	mockRaterRepo := mocks.NewMockRaterRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	mockTourneyRepo.EXPECT().
		FindBySlug(gomock.Any(), "essen-open-2025").
		Return(&domain.Tournament{
			ID:        tournamentID,
			Slug:      "essen-open-2025",
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
		TournamentSlug: "essen-open-2025",
		Nickname:       "Alice",
		Code:           "1234",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)

	// Validate the issued token
	jwtSvc := newTestJWTService(t)
	claims, err := jwtSvc.ValidateClaims(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, raterID.String(), claims.Subject)
	assert.Equal(t, "rater", claims.Role)
	assert.Equal(t, tournamentID.String(), claims.TournamentID)
}

func TestRaterLogin_TournamentNotFound(t *testing.T) {
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

func TestRaterLogin_InvalidCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tournamentID := uuid.New()

	mockTourneyRepo := mocks.NewMockTournamentRepository(ctrl)
	mockTourneyRepo.EXPECT().
		FindBySlug(gomock.Any(), "essen-open-2025").
		Return(&domain.Tournament{ID: tournamentID, Slug: "essen-open-2025"}, nil)

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
		TournamentSlug: "essen-open-2025",
		Nickname:       "Alice",
		Code:           "9999",
	})

	assert.ErrorIs(t, err, domain.ErrInvalidCode)
}

func TestRaterLogin_TokenExpiryWithoutVotingEnd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tournamentID := uuid.New()
	raterID := uuid.New()

	mockTourneyRepo := mocks.NewMockTournamentRepository(ctrl)
	mockTourneyRepo.EXPECT().
		FindBySlug(gomock.Any(), "essen-open-2025").
		Return(&domain.Tournament{
			ID:        tournamentID,
			Slug:      "essen-open-2025",
			VotingEnd: nil, // no voting end set
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
		TournamentSlug: "essen-open-2025",
		Nickname:       "Bob",
		Code:           "5678",
	})

	require.NoError(t, err)
	// ExpiresIn should be ~24h
	assert.InDelta(t, (24 * time.Hour).Seconds(), float64(pair.ExpiresIn), 5)
}

func TestHashToken_IsDeterministic(t *testing.T) {
	h1 := hashToken("my-raw-token")
	h2 := hashToken("my-raw-token")
	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 64) // SHA-256 hex = 64 chars
}

func TestHashToken_DiffersForDiffInputs(t *testing.T) {
	h1 := hashToken("token-a")
	h2 := hashToken("token-b")
	assert.NotEqual(t, h1, h2)
}
