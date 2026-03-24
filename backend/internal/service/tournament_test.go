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
	slugEssenOpen    = "essen-open-2025"
	slugEssenOpen2   = "essen-open-2025-2"
	slugMyTournament = "my-tournament"
	slugSomeTournament = "some-tournament"
	inviteCodeABC    = "ABC123"
	statusDraft      = "draft"
)

// newTournamentService is a test helper that wires up mocks into a TournamentService.
func newTournamentService(
	ctrl *gomock.Controller,
) (
	*TournamentService,
	*mocks.MockTournamentRepository,
	*mocks.MockTournamentMemberRepository,
	*mocks.MockUserRepository,
) {
	tourneyRepo := mocks.NewMockTournamentRepository(ctrl)
	memberRepo := mocks.NewMockTournamentMemberRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	svc := NewTournamentService(tourneyRepo, memberRepo, userRepo)
	return svc, tourneyRepo, memberRepo, userRepo
}

func TestCreateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, tourneyRepo, memberRepo, _ := newTournamentService(ctrl)

	organizerID := uuid.New()
	ctx := context.Background()

	input := CreateTournamentInput{
		Name:       "Essen Open 2025",
		Type:       "fantasy",
		TableCount: 20,
	}

	tourneyRepo.EXPECT().ExistsBySlug(ctx, slugEssenOpen).Return(false, nil)
	tourneyRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	memberRepo.EXPECT().AddMember(ctx, gomock.Any(), organizerID, roleOrganizer).Return(nil)

	tournament, err := svc.Create(ctx, organizerID, input)
	require.NoError(t, err)
	assert.Equal(t, slugEssenOpen, tournament.Slug)
	assert.Equal(t, statusDraft, tournament.Status)
	assert.Equal(t, organizerID, tournament.OrganizerID)
}

func TestCreateSlugDedup(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, tourneyRepo, memberRepo, _ := newTournamentService(ctrl)

	organizerID := uuid.New()
	ctx := context.Background()

	input := CreateTournamentInput{
		Name:       "Essen Open 2025",
		Type:       "fantasy",
		TableCount: 10,
	}

	// first slug taken, second available
	tourneyRepo.EXPECT().ExistsBySlug(ctx, slugEssenOpen).Return(true, nil)
	tourneyRepo.EXPECT().ExistsBySlug(ctx, slugEssenOpen2).Return(false, nil)
	tourneyRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	memberRepo.EXPECT().AddMember(ctx, gomock.Any(), organizerID, roleOrganizer).Return(nil)

	tournament, err := svc.Create(ctx, organizerID, input)
	require.NoError(t, err)
	assert.Equal(t, slugEssenOpen2, tournament.Slug)
}

func TestDeleteForbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, tourneyRepo, memberRepo, _ := newTournamentService(ctrl)

	helperID := uuid.New()
	tournamentID := uuid.New()
	ctx := context.Background()

	tourneyRepo.EXPECT().
		FindBySlug(ctx, slugSomeTournament).
		Return(&domain.Tournament{ID: tournamentID, OrganizerID: uuid.New()}, nil)

	memberRepo.EXPECT().
		GetRole(ctx, tournamentID, helperID).
		Return(roleHelper, nil)

	err := svc.Delete(ctx, helperID, slugSomeTournament)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestAddMemberSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, tourneyRepo, memberRepo, userRepo := newTournamentService(ctrl)

	organizerID := uuid.New()
	helperID := uuid.New()
	tournamentID := uuid.New()
	ctx := context.Background()

	helper := &domain.User{ID: helperID, Name: "Helper User", HelperInviteCode: inviteCodeABC}

	tourneyRepo.EXPECT().
		FindBySlug(ctx, slugMyTournament).
		Return(&domain.Tournament{ID: tournamentID, OrganizerID: organizerID}, nil)

	memberRepo.EXPECT().
		GetRole(ctx, tournamentID, organizerID).
		Return(roleOrganizer, nil)

	userRepo.EXPECT().
		FindByHelperInviteCode(ctx, inviteCodeABC).
		Return(helper, nil)

	memberRepo.EXPECT().
		AddMember(ctx, tournamentID, helperID, roleHelper).
		Return(nil)

	user, err := svc.AddMember(ctx, organizerID, slugMyTournament, inviteCodeABC)
	require.NoError(t, err)
	assert.Equal(t, helperID, user.ID)
}

func TestAddMemberNotOrganizer(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, tourneyRepo, memberRepo, _ := newTournamentService(ctrl)

	helperID := uuid.New()
	tournamentID := uuid.New()
	ctx := context.Background()

	tourneyRepo.EXPECT().
		FindBySlug(ctx, slugMyTournament).
		Return(&domain.Tournament{ID: tournamentID, OrganizerID: uuid.New()}, nil)

	memberRepo.EXPECT().
		GetRole(ctx, tournamentID, helperID).
		Return(roleHelper, nil)

	_, err := svc.AddMember(ctx, helperID, slugMyTournament, "SOMECODE")
	assert.ErrorIs(t, err, domain.ErrForbidden)
}
