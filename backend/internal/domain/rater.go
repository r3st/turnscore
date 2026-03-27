package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Rater represents a tournament participant who rates tables.
// No email or full name — GDPR compliant.
type Rater struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	TournamentID uuid.UUID `gorm:"type:uuid;not null"`
	Nickname     string    `gorm:"not null"`
	Code         string    `gorm:"not null"`
	CreatedAt    time.Time
}

// Rating represents a rater's evaluation of a single table.
// rater_id is NEVER exposed externally (privacy rule).
type Rating struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	TableID        uuid.UUID `gorm:"type:uuid;not null"`
	RaterID        uuid.UUID `gorm:"type:uuid;not null"`
	CriteriaScores string    `gorm:"type:jsonb;not null"` // JSON: {"balance":2,...}
	PlayedZone     string    `gorm:"not null;default:none"`
	Comment        *string
	CreatedAt      time.Time
}

// RatingResult is a read-model for results aggregation (not persisted).
type RatingResult struct {
	TableID          uuid.UUID
	TableNumber      int
	TableName        *string
	RatingCount      int64
	CriteriaAverages map[string]float64 // key=criterion, value=avg score
	ZoneNone         int64
	ZoneA            int64
	ZoneB            int64
}

// EventRating is a rater's one-time evaluation of the tournament as a whole
// (optional criteria: catering, venue, organization). Not tied to a specific table.
type EventRating struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	TournamentID   uuid.UUID `gorm:"type:uuid;not null"`
	RaterID        uuid.UUID `gorm:"type:uuid;not null"`
	CriteriaScores string    `gorm:"type:jsonb;not null;default:'{}'"`
	Comment        *string
	CreatedAt      time.Time
}

// EventRatingRepository defines data access for tournament-level event ratings.
type EventRatingRepository interface {
	// Create persists a new event rating. Returns ErrDuplicateRating if already submitted.
	Create(ctx context.Context, r *EventRating) error

	// ExistsByTournamentAndRater returns true if the rater already submitted an event rating.
	ExistsByTournamentAndRater(ctx context.Context, tournamentID, raterID uuid.UUID) (bool, error)
}

// RaterRepository defines data access for raters.
type RaterRepository interface {
	// FindByNicknameAndCode looks up a rater by tournament, nickname, and code.
	// Returns ErrNotFound if no matching rater exists.
	FindByNicknameAndCode(ctx context.Context, tournamentID uuid.UUID, nickname, code string) (*Rater, error)

	// FindByID returns a rater by their internal ID.
	FindByID(ctx context.Context, id uuid.UUID) (*Rater, error)

	// ListByTournamentID returns all raters for a tournament ordered by nickname ASC.
	ListByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]Rater, error)

	// Create persists a new rater. Returns ErrConflict on duplicate nickname or code.
	Create(ctx context.Context, r *Rater) error

	// ExistsByNickname returns true if a rater with the given nickname exists in the tournament.
	ExistsByNickname(ctx context.Context, tournamentID uuid.UUID, nickname string) (bool, error)

	// ExistsByCode returns true if a rater with the given code exists in the tournament.
	ExistsByCode(ctx context.Context, tournamentID uuid.UUID, code string) (bool, error)
}

// RatingRepository defines data access for ratings.
type RatingRepository interface {
	// Create persists a new rating. Returns ErrDuplicateRating on unique violation.
	Create(ctx context.Context, r *Rating) error

	// ExistsByTableAndRater returns true if the rater already rated the table.
	ExistsByTableAndRater(ctx context.Context, tableID, raterID uuid.UUID) (bool, error)

	// GetResultsForTournament returns aggregate results — one entry per table.
	GetResultsForTournament(ctx context.Context, tournamentID uuid.UUID) ([]RatingResult, error)

	// GetCommentsForTournament returns all non-empty comments per table for a tournament.
	// Returns map[tableID][]string of comments — fully anonymized, no rater identity.
	GetCommentsForTournament(ctx context.Context, tournamentID uuid.UUID) (map[uuid.UUID][]string, error)
}
