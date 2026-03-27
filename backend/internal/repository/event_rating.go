package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/r3st/turnscore/internal/domain"
)

// EventRatingRepository persists tournament-level event ratings.
type EventRatingRepository struct {
	db *gorm.DB
}

// NewEventRatingRepository creates a new EventRatingRepository.
func NewEventRatingRepository(db *gorm.DB) *EventRatingRepository {
	return &EventRatingRepository{db: db}
}

// Create inserts a new event rating. Returns domain.ErrDuplicateRating on unique violation.
func (r *EventRatingRepository) Create(ctx context.Context, er *domain.EventRating) error {
	result := r.db.WithContext(ctx).Create(er)
	if result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) && pgErr.Code == pgErrDuplicateKey {
			return domain.ErrDuplicateRating
		}
		return result.Error
	}
	return nil
}

// ExistsByTournamentAndRater returns true if the rater already submitted an event rating.
func (r *EventRatingRepository) ExistsByTournamentAndRater(ctx context.Context, tournamentID, raterID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.EventRating{}).
		Where("tournament_id = ? AND rater_id = ?", tournamentID, raterID).
		Count(&count).Error
	return count > 0, err
}
