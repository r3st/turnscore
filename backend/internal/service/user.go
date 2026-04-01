package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/r3st/turnscore/internal/domain"
	"github.com/r3st/turnscore/internal/storage"
)

// UserMembership is the service-layer view of a tournament membership.
type UserMembership struct {
	TournamentID   uuid.UUID
	TournamentName string
	TournamentSlug string
	Role           string
}

// UserService contains business logic for user profile management.
type UserService struct {
	userRepo    domain.UserRepository
	memberRepo  domain.TournamentMemberRepository
	tourneyRepo domain.TournamentRepository
	tableRepo   domain.TableRepository
	photoRepo   domain.PhotoRepository
	storage     storage.Storage
}

// NewUserService creates a new UserService.
func NewUserService(
	userRepo domain.UserRepository,
	memberRepo domain.TournamentMemberRepository,
	tourneyRepo domain.TournamentRepository,
	tableRepo domain.TableRepository,
	photoRepo domain.PhotoRepository,
	stor storage.Storage,
) *UserService {
	return &UserService{
		userRepo:    userRepo,
		memberRepo:  memberRepo,
		tourneyRepo: tourneyRepo,
		tableRepo:   tableRepo,
		photoRepo:   photoRepo,
		storage:     stor,
	}
}

// UpdatePreferences saves theme and color_mode for a user.
func (s *UserService) UpdatePreferences(ctx context.Context, userID uuid.UUID, input domain.UpdatePreferencesInput) (*domain.User, error) {
	user, err := s.userRepo.UpdatePreferences(ctx, userID, input)
	if err != nil {
		return nil, fmt.Errorf("UpdatePreferences: %w", err)
	}
	return user, nil
}

// GetProfile returns the user's profile together with their tournament memberships.
func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, []UserMembership, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("GetProfile find user: %w", err)
	}

	rawMemberships, err := s.memberRepo.ListWithTournamentsByUserID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("GetProfile list memberships: %w", err)
	}

	memberships := make([]UserMembership, 0, len(rawMemberships))
	for _, m := range rawMemberships {
		memberships = append(memberships, UserMembership{
			TournamentID:   m.TournamentID,
			TournamentName: m.TournamentName,
			TournamentSlug: m.TournamentSlug,
			Role:           m.Role,
		})
	}

	return user, memberships, nil
}

// DeleteAccount permanently deletes the user and all owned data.
// confirmEmail must match the stored email address.
// Cascade order:
//  1. Delete photo storage files for all owned tournaments
//  2. Delete each owned tournament (DB CASCADE handles tables, photos, raters, ratings, members)
//  3. Delete user record (DB CASCADE handles refresh_tokens, helper memberships)
func (s *UserService) DeleteAccount(ctx context.Context, userID uuid.UUID, confirmEmail string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("DeleteAccount find user: %w", err)
	}
	if !strings.EqualFold(user.Email, confirmEmail) {
		return domain.ErrForbidden
	}

	tournaments, err := s.tourneyRepo.ListByOrganizerID(ctx, userID)
	if err != nil {
		return fmt.Errorf("DeleteAccount list tournaments: %w", err)
	}

	for _, t := range tournaments {
		if err := s.deleteTournamentStorageFiles(ctx, t.ID); err != nil {
			return err
		}
		if err := s.tourneyRepo.Delete(ctx, t.ID); err != nil {
			return fmt.Errorf("DeleteAccount delete tournament: %w", err)
		}
	}

	if err := s.userRepo.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("DeleteAccount delete user: %w", err)
	}
	return nil
}

// deleteTournamentStorageFiles removes all photo files from storage for a tournament.
// DB records are cleaned up by CASCADE when the tournament is deleted.
func (s *UserService) deleteTournamentStorageFiles(ctx context.Context, tournamentID uuid.UUID) error {
	tables, err := s.tableRepo.ListByTournamentID(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("deleteTournamentStorageFiles list tables: %w", err)
	}
	for _, tbl := range tables {
		if err := s.deleteTableStorageFiles(ctx, tbl.ID); err != nil {
			return err
		}
	}
	return nil
}

// deleteTableStorageFiles removes all photo files from storage for a single table.
func (s *UserService) deleteTableStorageFiles(ctx context.Context, tableID uuid.UUID) error {
	photos, err := s.photoRepo.FindByTableID(ctx, tableID)
	if err != nil {
		return fmt.Errorf("deleteTableStorageFiles list photos: %w", err)
	}
	for _, p := range photos {
		if err := s.storage.Delete(ctx, p.URL); err != nil {
			return fmt.Errorf("deleteTableStorageFiles delete photo: %w", err)
		}
		if p.ThumbnailURL != nil {
			if err := s.storage.Delete(ctx, *p.ThumbnailURL); err != nil {
				return fmt.Errorf("deleteTableStorageFiles delete thumbnail: %w", err)
			}
		}
	}
	return nil
}
