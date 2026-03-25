package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/r3st/turnscore/internal/domain"
	"github.com/r3st/turnscore/pkg/mocks"
)

const (
	slugVotingTournament = "essen-open-voting"
	slugDraftTournament  = "essen-open-draft"
	nicknameAlpha        = "Alpha"
	codeFixed            = "1234"
	statusVoting         = "voting"
)

// newRaterService is a test helper that wires up mocks into a RaterService.
func newRaterService(ctrl *gomock.Controller) (
	*RaterService,
	*mocks.MockRaterRepository,
	*mocks.MockRatingRepository,
	*mocks.MockTableRepository,
	*mocks.MockTournamentMemberRepository,
	*mocks.MockTournamentRepository,
	*mocks.MockUserRepository,
) {
	raterRepo := mocks.NewMockRaterRepository(ctrl)
	ratingRepo := mocks.NewMockRatingRepository(ctrl)
	tableRepo := mocks.NewMockTableRepository(ctrl)
	memberRepo := mocks.NewMockTournamentMemberRepository(ctrl)
	tourneyRepo := mocks.NewMockTournamentRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewRaterService(raterRepo, ratingRepo, tableRepo, memberRepo, tourneyRepo, userRepo)
	return svc, raterRepo, ratingRepo, tableRepo, memberRepo, tourneyRepo, userRepo
}

func TestCreateRaterSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, raterRepo, _, _, memberRepo, tourneyRepo, _ := newRaterService(ctrl)

	ctx := context.Background()
	organizerID := uuid.New()
	tournamentID := uuid.New()

	tournament := &domain.Tournament{
		ID:     tournamentID,
		Slug:   slugVotingTournament,
		Status: statusVoting,
	}

	tourneyRepo.EXPECT().FindBySlug(ctx, slugVotingTournament).Return(tournament, nil)
	memberRepo.EXPECT().GetRole(ctx, tournamentID, organizerID).Return(roleOrganizer, nil)
	raterRepo.EXPECT().ExistsByCode(ctx, tournamentID, gomock.Any()).Return(false, nil)
	raterRepo.EXPECT().ExistsByNickname(ctx, tournamentID, nicknameAlpha).Return(false, nil)
	raterRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

	rater, rawCode, err := svc.CreateRater(ctx, organizerID, slugVotingTournament, nicknameAlpha, nil)

	require.NoError(t, err)
	assert.Equal(t, nicknameAlpha, rater.Nickname)
	assert.Equal(t, tournamentID, rater.TournamentID)
	assert.Len(t, rawCode, 4)
	assert.Regexp(t, `^\d{4}$`, rawCode)
}

func TestCreateRaterWithProvidedCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, raterRepo, _, _, memberRepo, tourneyRepo, _ := newRaterService(ctrl)

	ctx := context.Background()
	organizerID := uuid.New()
	tournamentID := uuid.New()
	code := codeFixed

	tournament := &domain.Tournament{
		ID:     tournamentID,
		Slug:   slugVotingTournament,
		Status: statusVoting,
	}

	tourneyRepo.EXPECT().FindBySlug(ctx, slugVotingTournament).Return(tournament, nil)
	memberRepo.EXPECT().GetRole(ctx, tournamentID, organizerID).Return(roleOrganizer, nil)
	raterRepo.EXPECT().ExistsByNickname(ctx, tournamentID, nicknameAlpha).Return(false, nil)
	raterRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

	rater, rawCode, err := svc.CreateRater(ctx, organizerID, slugVotingTournament, nicknameAlpha, &code)

	require.NoError(t, err)
	assert.Equal(t, nicknameAlpha, rater.Nickname)
	assert.Equal(t, codeFixed, rawCode)
}

func TestCreateRaterNicknameConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, raterRepo, _, _, memberRepo, tourneyRepo, _ := newRaterService(ctrl)

	ctx := context.Background()
	organizerID := uuid.New()
	tournamentID := uuid.New()
	code := codeFixed

	tournament := &domain.Tournament{
		ID:     tournamentID,
		Slug:   slugVotingTournament,
		Status: statusVoting,
	}

	tourneyRepo.EXPECT().FindBySlug(ctx, slugVotingTournament).Return(tournament, nil)
	memberRepo.EXPECT().GetRole(ctx, tournamentID, organizerID).Return(roleOrganizer, nil)
	raterRepo.EXPECT().ExistsByNickname(ctx, tournamentID, nicknameAlpha).Return(true, nil)

	_, _, err := svc.CreateRater(ctx, organizerID, slugVotingTournament, nicknameAlpha, &code)

	require.ErrorIs(t, err, domain.ErrConflict)
}

func TestSubmitRatingVotingNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, _, _, _, _, tourneyRepo, _ := newRaterService(ctrl)

	ctx := context.Background()
	raterID := uuid.New()
	tournamentID := uuid.New()

	tournament := &domain.Tournament{
		ID:     tournamentID,
		Slug:   slugDraftTournament,
		Status: statusDraft,
	}

	tourneyRepo.EXPECT().FindBySlug(ctx, slugDraftTournament).Return(tournament, nil)

	err := svc.SubmitRating(ctx, SubmitRatingInput{
		RaterID:      raterID,
		TournamentID: tournamentID,
		Slug:         slugDraftTournament,
		TableNumber:  1,
		Scores:       map[string]int{"balance": 2},
		PlayedZone:   "none",
	})

	require.ErrorIs(t, err, domain.ErrVotingNotActive)
}
