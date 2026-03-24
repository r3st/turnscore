package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tournament represents a tabletop game tournament.
type Tournament struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Slug           string     `gorm:"uniqueIndex;not null"`
	Name           string     `gorm:"not null"`
	Type           string     `gorm:"not null"` // "fantasy" | "scifi"
	Description    *string
	Links          string    `gorm:"type:jsonb;not null;default:'[]'"`
	Location       *string
	EventDate      *time.Time
	OrganizerID    uuid.UUID `gorm:"type:uuid;not null"`
	TableCount     int       `gorm:"not null"`
	Status         string    `gorm:"not null;default:draft"` // draft|active|voting|archived
	VotingStart    *time.Time
	VotingEnd      *time.Time
	ActiveCriteria string `gorm:"type:jsonb;not null"`
	ResultConfig   string `gorm:"type:jsonb;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TournamentRepository defines data access for tournaments.
type TournamentRepository interface {
	// FindBySlug returns a tournament by its URL slug.
	FindBySlug(ctx interface{}, slug string) (*Tournament, error)

	// FindByID returns a tournament by its internal ID.
	FindByID(ctx interface{}, id uuid.UUID) (*Tournament, error)
}
