package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents an organizer or helper account.
// Role is NOT stored here — it lives in TournamentMembership (per-tournament).
type User struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	GoogleSub        string    `gorm:"uniqueIndex;not null"`
	Email            string    `gorm:"uniqueIndex;not null"`
	Name             string    `gorm:"not null"`
	AvatarURL        *string
	HelperInviteCode string    `gorm:"uniqueIndex;not null"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// UserRepository defines data access for users.
type UserRepository interface {
	// UpsertByGoogleSub creates or updates a user based on their Google subject ID.
	// Generates a helper_invite_code on first creation.
	UpsertByGoogleSub(ctx interface{}, googleSub, email, name, avatarURL string) (*User, error)

	// FindByID returns a user by their internal ID.
	FindByID(ctx interface{}, id uuid.UUID) (*User, error)

	// FindByHelperInviteCode looks up a user by their invite code.
	FindByHelperInviteCode(ctx interface{}, code string) (*User, error)

	// FindByRefreshToken validates a refresh token and returns the associated user.
	FindByRefreshToken(ctx interface{}, token string) (*User, error)

	// SaveRefreshToken stores a hashed refresh token for a user.
	SaveRefreshToken(ctx interface{}, userID uuid.UUID, hashedToken string, expiresAt time.Time) error

	// DeleteRefreshToken removes a refresh token (logout).
	DeleteRefreshToken(ctx interface{}, hashedToken string) error
}
