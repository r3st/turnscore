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
	slugTableTournament = "battle-open-2025"
	tableCountDefault   = 5
)

// newTableService wires up mocks into a TableService for testing.
func newTableService(ctrl *gomock.Controller) (
	*TableService,
	*mocks.MockTableRepository,
	*mocks.MockTournamentMemberRepository,
	*mocks.MockTournamentRepository,
	*mocks.MockPhotoRepository,
	*mocks.MockStorage,
) {
	tableRepo := mocks.NewMockTableRepository(ctrl)
	memberRepo := mocks.NewMockTournamentMemberRepository(ctrl)
	tourneyRepo := mocks.NewMockTournamentRepository(ctrl)
	photoRepo := mocks.NewMockPhotoRepository(ctrl)
	stor := mocks.NewMockStorage(ctrl)
	svc := NewTableService(tableRepo, memberRepo, tourneyRepo, photoRepo, stor)
	return svc, tableRepo, memberRepo, tourneyRepo, photoRepo, stor
}

// fakeTournament returns a minimal Tournament with tableCount tables.
func fakeTournament(tableCount int) *domain.Tournament {
	return &domain.Tournament{
		ID:             uuid.New(),
		Slug:           slugTableTournament,
		Name:           "Battle Open 2025",
		Type:           "fantasy",
		TableCount:     tableCount,
		Status:         "draft",
		ActiveCriteria: `["balance","aesthetics"]`,
		ResultConfig:   `{}`,
		Links:          `[]`,
	}
}

func TestTableServiceCreateForTournamentSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, tableRepo, memberRepo, tourneyRepo, _, _ := newTableService(ctrl)

	ctx := context.Background()
	organizerID := uuid.New()
	tourney := fakeTournament(tableCountDefault)

	created := make([]domain.Table, tableCountDefault)
	for i := range created {
		created[i] = domain.Table{ID: uuid.New(), TournamentID: tourney.ID, Number: i + 1}
	}

	tourneyRepo.EXPECT().FindBySlug(ctx, slugTableTournament).Return(tourney, nil)
	memberRepo.EXPECT().GetRole(ctx, tourney.ID, organizerID).Return(roleOrganizer, nil)
	// First call: no existing tables
	tableRepo.EXPECT().ListByTournamentID(ctx, tourney.ID).Return([]domain.Table{}, nil)
	tableRepo.EXPECT().CreateBatch(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, tables []domain.Table) error {
			assert.Len(t, tables, tableCountDefault)
			for i, tbl := range tables {
				assert.Equal(t, i+1, tbl.Number)
				assert.Equal(t, tourney.ID, tbl.TournamentID)
			}
			return nil
		},
	)
	// Second call: return final list
	tableRepo.EXPECT().ListByTournamentID(ctx, tourney.ID).Return(created, nil)

	tables, err := svc.CreateForTournament(ctx, organizerID, slugTableTournament)
	require.NoError(t, err)
	assert.Len(t, tables, tableCountDefault)
	for i, tbl := range tables {
		assert.Equal(t, i+1, tbl.Number)
	}
}

func TestTableServiceCreateForTournamentAdditive(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, tableRepo, memberRepo, tourneyRepo, _, _ := newTableService(ctrl)

	ctx := context.Background()
	organizerID := uuid.New()
	tourney := fakeTournament(tableCountDefault)

	// Tables 1 and 2 already exist; 3–5 are missing.
	existing := []domain.Table{
		{ID: uuid.New(), TournamentID: tourney.ID, Number: 1},
		{ID: uuid.New(), TournamentID: tourney.ID, Number: 2},
	}
	all := append(existing,
		domain.Table{ID: uuid.New(), TournamentID: tourney.ID, Number: 3},
		domain.Table{ID: uuid.New(), TournamentID: tourney.ID, Number: 4},
		domain.Table{ID: uuid.New(), TournamentID: tourney.ID, Number: 5},
	)

	tourneyRepo.EXPECT().FindBySlug(ctx, slugTableTournament).Return(tourney, nil)
	memberRepo.EXPECT().GetRole(ctx, tourney.ID, organizerID).Return(roleOrganizer, nil)
	tableRepo.EXPECT().ListByTournamentID(ctx, tourney.ID).Return(existing, nil)
	tableRepo.EXPECT().CreateBatch(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, tables []domain.Table) error {
			assert.Len(t, tables, 3)
			assert.Equal(t, 3, tables[0].Number)
			assert.Equal(t, 4, tables[1].Number)
			assert.Equal(t, 5, tables[2].Number)
			return nil
		},
	)
	tableRepo.EXPECT().ListByTournamentID(ctx, tourney.ID).Return(all, nil)

	tables, err := svc.CreateForTournament(ctx, organizerID, slugTableTournament)
	require.NoError(t, err)
	assert.Len(t, tables, tableCountDefault)
}

func TestTableServiceCreateForTournamentForbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, _, memberRepo, tourneyRepo, _, _ := newTableService(ctrl)

	ctx := context.Background()
	nonMemberID := uuid.New()
	tourney := fakeTournament(tableCountDefault)

	tourneyRepo.EXPECT().FindBySlug(ctx, slugTableTournament).Return(tourney, nil)
	memberRepo.EXPECT().GetRole(ctx, tourney.ID, nonMemberID).Return("", domain.ErrNotFound)

	tables, err := svc.CreateForTournament(ctx, nonMemberID, slugTableTournament)
	assert.Nil(t, tables)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestTableServiceUpdateSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc, tableRepo, memberRepo, tourneyRepo, _, _ := newTableService(ctrl)

	ctx := context.Background()
	helperID := uuid.New()
	tourney := fakeTournament(tableCountDefault)

	existingName := "Old Name"
	table := &domain.Table{
		ID:           uuid.New(),
		TournamentID: tourney.ID,
		Number:       3,
		Name:         &existingName,
	}

	newName := "New Name"
	newDesc := "A great table"

	tourneyRepo.EXPECT().FindBySlug(ctx, slugTableTournament).Return(tourney, nil)
	memberRepo.EXPECT().GetRole(ctx, tourney.ID, helperID).Return(roleHelper, nil)
	tableRepo.EXPECT().FindByTournamentAndNumber(ctx, tourney.ID, 3).Return(table, nil)
	tableRepo.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(
		func(_ context.Context, tbl *domain.Table) error {
			assert.Equal(t, &newName, tbl.Name)
			assert.Equal(t, &newDesc, tbl.Description)
			return nil
		},
	)

	updated, err := svc.Update(ctx, helperID, slugTableTournament, 3, &newName, &newDesc)
	require.NoError(t, err)
	assert.Equal(t, &newName, updated.Name)
	assert.Equal(t, &newDesc, updated.Description)
}
