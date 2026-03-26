package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Tournament is the GORM model mapped to the tournaments table.
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
	Status         string    `gorm:"not null;default:draft"`
	VotingStart    *time.Time
	VotingEnd      *time.Time
	ActiveCriteria string `gorm:"type:jsonb;not null"`
	ResultConfig   string `gorm:"type:jsonb;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TournamentMembership is the GORM model for tournament_members.
type TournamentMembership struct {
	TournamentID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	Role         string    `gorm:"not null"` // "organizer" | "helper"
	JoinedAt     time.Time
}

func (TournamentMembership) TableName() string { return "tournament_members" }

// TournamentRepository defines data access for tournaments.
type TournamentRepository interface {
	FindBySlug(ctx context.Context, slug string) (*Tournament, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Tournament, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	ListUpcoming(ctx context.Context, limit int) ([]Tournament, error)
	ListPast(ctx context.Context, limit int) ([]Tournament, error)
	Create(ctx context.Context, t *Tournament) error
	Update(ctx context.Context, t *Tournament) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// MembershipWithTournament is a read-model joining membership with tournament info.
type MembershipWithTournament struct {
	TournamentID   uuid.UUID
	TournamentName string
	TournamentSlug string
	Role           string
}

// TournamentMemberRepository defines data access for tournament membership.
type TournamentMemberRepository interface {
	GetRole(ctx context.Context, tournamentID, userID uuid.UUID) (string, error)
	AddMember(ctx context.Context, tournamentID, userID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, tournamentID, userID uuid.UUID) error

	// ListWithTournamentsByUserID returns memberships with tournament info for a user.
	ListWithTournamentsByUserID(ctx context.Context, userID uuid.UUID) ([]MembershipWithTournament, error)
}
