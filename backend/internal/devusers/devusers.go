// Package devusers provides predefined test users for local development and e2e testing.
// This package must only be used when the server runs in debug mode.
package devusers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/r3st/turnscore/internal/domain"
)

// User holds the fixed definition of a dev user.
type User struct {
	ID               uuid.UUID
	Name             string
	Password         string // same as Name, lowercase
	Email            string
	GoogleSub        string
	HelperInviteCode string
}

// All returns the five predefined dev users.
var All = []User{
	{
		ID:               uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:             "Alice",
		Password:         "alice",
		Email:            "alice@dev.local",
		GoogleSub:        "dev-alice",
		HelperInviteCode: "DEV-ALICE",
	},
	{
		ID:               uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		Name:             "Bob",
		Password:         "bob",
		Email:            "bob@dev.local",
		GoogleSub:        "dev-bob",
		HelperInviteCode: "DEV-BOB",
	},
	{
		ID:               uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		Name:             "Carol",
		Password:         "carol",
		Email:            "carol@dev.local",
		GoogleSub:        "dev-carol",
		HelperInviteCode: "DEV-CAROL",
	},
	{
		ID:               uuid.MustParse("00000000-0000-0000-0000-000000000004"),
		Name:             "Dave",
		Password:         "dave",
		Email:            "dave@dev.local",
		GoogleSub:        "dev-dave",
		HelperInviteCode: "DEV-DAVE",
	},
	{
		ID:               uuid.MustParse("00000000-0000-0000-0000-000000000005"),
		Name:             "Eve",
		Password:         "eve",
		Email:            "eve@dev.local",
		GoogleSub:        "dev-eve",
		HelperInviteCode: "DEV-EVE",
	},
}

// FindByPassword returns the dev user matching the given name/password (case-insensitive).
func FindByPassword(password string) *User {
	for i := range All {
		if All[i].Password == password {
			return &All[i]
		}
	}
	return nil
}

// Seed upserts all dev users into the database. Safe to call on every startup.
func Seed(ctx context.Context, db *gorm.DB) error {
	for _, u := range All {
		row := domain.User{
			ID:               u.ID,
			GoogleSub:        u.GoogleSub,
			Email:            u.Email,
			Name:             u.Name,
			HelperInviteCode: u.HelperInviteCode,
			Theme:            "scifi",
			ColorMode:        "dark",
		}
		result := db.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "email", "google_sub", "helper_invite_code"}),
			}).
			Create(&row)
		if result.Error != nil {
			return fmt.Errorf("seed dev user %s: %w", u.Name, result.Error)
		}
	}
	return nil
}
