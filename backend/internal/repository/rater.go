package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/r3st/turnscore/internal/domain"
)

// RaterRepository implements domain.RaterRepository using GORM.
type RaterRepository struct {
	db *gorm.DB
}

// NewRaterRepository creates a new RaterRepository backed by the given DB.
func NewRaterRepository(db *gorm.DB) *RaterRepository {
	return &RaterRepository{db: db}
}

// FindByNicknameAndCode looks up a rater by tournament, nickname, and code.
func (r *RaterRepository) FindByNicknameAndCode(ctx context.Context, tournamentID uuid.UUID, nickname, code string) (*domain.Rater, error) {
	var rater domain.Rater
	err := r.db.WithContext(ctx).
		Where("tournament_id = ? AND nickname = ? AND code = ?", tournamentID, nickname, code).
		First(&rater).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("FindByNicknameAndCode: %w", err)
	}
	return &rater, nil
}

// FindByID returns a rater by their internal UUID.
func (r *RaterRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Rater, error) {
	var rater domain.Rater
	err := r.db.WithContext(ctx).First(&rater, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("FindByID rater: %w", err)
	}
	return &rater, nil
}

// ListByTournamentID returns all raters for a tournament ordered by nickname ASC.
func (r *RaterRepository) ListByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]domain.Rater, error) {
	var raters []domain.Rater
	err := r.db.WithContext(ctx).
		Where("tournament_id = ?", tournamentID).
		Order("nickname ASC").
		Find(&raters).Error
	if err != nil {
		return nil, fmt.Errorf("ListByTournamentID raters: %w", err)
	}
	return raters, nil
}

// Create persists a new rater. Returns domain.ErrConflict on unique violation.
func (r *RaterRepository) Create(ctx context.Context, rater *domain.Rater) error {
	err := r.db.WithContext(ctx).Create(rater).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgErrDuplicateKey {
			return domain.ErrConflict
		}
		return fmt.Errorf("Create rater: %w", err)
	}
	return nil
}

// ExistsByNickname returns true if a rater with the given nickname exists in the tournament.
func (r *RaterRepository) ExistsByNickname(ctx context.Context, tournamentID uuid.UUID, nickname string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Rater{}).
		Where("tournament_id = ? AND nickname = ?", tournamentID, nickname).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("ExistsByNickname: %w", err)
	}
	return count > 0, nil
}

// ExistsByCode returns true if a rater with the given code exists in the tournament.
func (r *RaterRepository) ExistsByCode(ctx context.Context, tournamentID uuid.UUID, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Rater{}).
		Where("tournament_id = ? AND code = ?", tournamentID, code).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("ExistsByCode: %w", err)
	}
	return count > 0, nil
}
