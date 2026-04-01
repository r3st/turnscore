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
	Theme            string    `gorm:"not null;default:scifi"`
	ColorMode        string    `gorm:"not null;default:dark"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// UpdatePreferencesInput holds the user-editable preference fields.
type UpdatePreferencesInput struct {
	Theme     *string // "scifi" | "fantasy"
	ColorMode *string // "light" | "dark"
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

	// UpdatePreferences persists theme and color_mode for the given user.
	UpdatePreferences(ctx interface{}, userID uuid.UUID, input UpdatePreferencesInput) (*User, error)

	// DeleteUser permanently removes a user record. Cascade in DB handles
	// refresh_tokens and tournament_memberships (as helper).
	DeleteUser(ctx interface{}, userID uuid.UUID) error
}
