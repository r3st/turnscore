package domain

import (
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

// RaterRepository defines data access for raters.
type RaterRepository interface {
	// FindByNicknameAndCode looks up a rater by tournament, nickname, and code.
	// Returns ErrNotFound if no matching rater exists.
	FindByNicknameAndCode(ctx interface{}, tournamentID uuid.UUID, nickname, code string) (*Rater, error)
}
