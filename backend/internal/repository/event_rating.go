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

// GetAveragesForTournament computes the average score per criterion key across all event ratings
// for the given tournament. Returns nil when no event ratings exist.
func (r *EventRatingRepository) GetAveragesForTournament(ctx context.Context, tournamentID uuid.UUID) (map[string]float64, error) {
	var rows []struct {
		Key string
		Avg float64
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT key, AVG(value::float) AS avg
		FROM event_ratings,
		     jsonb_each_text(criteria_scores) AS kv(key, value)
		WHERE tournament_id = ?
		GROUP BY key
	`, tournamentID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("GetAveragesForTournament: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	result := make(map[string]float64, len(rows))
	for _, row := range rows {
		result[row.Key] = row.Avg
	}
	return result, nil
}
