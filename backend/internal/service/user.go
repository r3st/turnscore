package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/r3st/turnscore/internal/domain"
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
}

// NewUserService creates a new UserService.
func NewUserService(
	userRepo domain.UserRepository,
	memberRepo domain.TournamentMemberRepository,
	tourneyRepo domain.TournamentRepository,
) *UserService {
	return &UserService{
		userRepo:    userRepo,
		memberRepo:  memberRepo,
		tourneyRepo: tourneyRepo,
	}
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
