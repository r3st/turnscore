package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/r3st/turnscore/internal/domain"
)

// generateInviteCode creates a cryptographically random 8-byte hex string (16 chars).
func generateInviteCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// refreshToken is the GORM model for the refresh_tokens table.
type refreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	TokenHash string    `gorm:"column:token_hash;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
}

func (refreshToken) TableName() string { return "refresh_tokens" }

// UserRepository implements domain.UserRepository using GORM.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository backed by the given DB.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// toContext coerces ctx interface{} to context.Context, falling back to Background.
func toContext(ctx interface{}) context.Context {
	if c, ok := ctx.(context.Context); ok {
		return c
	}
	return context.Background()
}

// UpsertByGoogleSub creates or updates a user identified by their Google subject ID.
// On first creation a random helper_invite_code is generated.
func (r *UserRepository) UpsertByGoogleSub(ctx interface{}, googleSub, email, name, avatarURL string) (*domain.User, error) {
	c := toContext(ctx)

	var u domain.User
	err := r.db.WithContext(c).
		Where("google_sub = ?", googleSub).
		First(&u).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// New user — generate invite code and create
		code, genErr := generateInviteCode()
		if genErr != nil {
			return nil, fmt.Errorf("UpsertByGoogleSub generate code: %w", genErr)
		}
		u = domain.User{
			ID:               uuid.New(),
			GoogleSub:        googleSub,
			Email:            email,
			Name:             name,
			HelperInviteCode: code,
		}
		if avatarURL != "" {
			u.AvatarURL = &avatarURL
		}
		if createErr := r.db.WithContext(c).Create(&u).Error; createErr != nil {
			return nil, fmt.Errorf("UpsertByGoogleSub create: %w", createErr)
		}
		return &u, nil
	}
	if err != nil {
		return nil, fmt.Errorf("UpsertByGoogleSub find: %w", err)
	}

	// Existing user — update mutable fields
	updates := map[string]interface{}{
		"email": email,
		"name":  name,
	}
	if avatarURL != "" {
		updates["avatar_url"] = avatarURL
	}
	if updateErr := r.db.WithContext(c).Model(&u).Updates(updates).Error; updateErr != nil {
		return nil, fmt.Errorf("UpsertByGoogleSub update: %w", updateErr)
	}
	return &u, nil
}

// FindByID returns a user by their internal UUID.
func (r *UserRepository) FindByID(ctx interface{}, id uuid.UUID) (*domain.User, error) {
	c := toContext(ctx)
	var u domain.User
	err := r.db.WithContext(c).First(&u, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("FindByID: %w", err)
	}
	return &u, nil
}

// FindByHelperInviteCode looks up a user by their invite code.
func (r *UserRepository) FindByHelperInviteCode(ctx interface{}, code string) (*domain.User, error) {
	c := toContext(ctx)
	var u domain.User
	err := r.db.WithContext(c).Where("helper_invite_code = ?", code).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("FindByHelperInviteCode: %w", err)
	}
	return &u, nil
}

// FindByRefreshToken validates a non-expired refresh token hash and returns the user.
func (r *UserRepository) FindByRefreshToken(ctx interface{}, tokenHash string) (*domain.User, error) {
	c := toContext(ctx)

	var rt refreshToken
	err := r.db.WithContext(c).
		Where("token_hash = ? AND expires_at > NOW()", tokenHash).
		First(&rt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUnauthorized
	}
	if err != nil {
		return nil, fmt.Errorf("FindByRefreshToken lookup: %w", err)
	}

	return r.FindByID(c, rt.UserID)
}

// SaveRefreshToken stores a hashed refresh token for a user.
func (r *UserRepository) SaveRefreshToken(ctx interface{}, userID uuid.UUID, hashedToken string, expiresAt time.Time) error {
	c := toContext(ctx)
	rt := refreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hashedToken,
		ExpiresAt: expiresAt,
	}
	if err := r.db.WithContext(c).Create(&rt).Error; err != nil {
		return fmt.Errorf("SaveRefreshToken: %w", err)
	}
	return nil
}

// DeleteRefreshToken removes a refresh token by its hash (logout / rotation).
func (r *UserRepository) DeleteRefreshToken(ctx interface{}, hashedToken string) error {
	c := toContext(ctx)
	if err := r.db.WithContext(c).
		Where("token_hash = ?", hashedToken).
		Delete(&refreshToken{}).Error; err != nil {
		return fmt.Errorf("DeleteRefreshToken: %w", err)
	}
	return nil
}
