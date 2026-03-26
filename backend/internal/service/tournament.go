package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/r3st/turnscore/internal/domain"
)

const (
	roleOrganizer = "organizer"
	roleHelper    = "helper"

	defaultListLimit = 5
	maxSlugAttempts  = 10
	maxSlugLength    = 90

	defaultCriteriaJSON = `["balance","aesthetics","terrain_density","labeling","overall","zone_a","zone_b"]`
	defaultResultConfig = `{"show_comments":false,"visible_comment_criteria":[]}`
)

var defaultActiveCriteria = []string{
	"balance", "aesthetics", "terrain_density", "labeling", "overall", "zone_a", "zone_b",
}

var slugReplacer = regexp.MustCompile(`[^a-z0-9]+`)

// TournamentLink is the service-layer representation of a link associated with a tournament.
type TournamentLink struct {
	URL   string `json:"url"`
	Label string `json:"label"`
}

// CreateTournamentInput holds the validated input for creating a tournament.
type CreateTournamentInput struct {
	Name           string
	Type           string
	Description    *string
	Links          []TournamentLink
	Location       *string
	EventDate      *time.Time
	TableCount     int
	VotingStart    *time.Time
	VotingEnd      *time.Time
	ActiveCriteria []string
}

// UpdateTournamentInput holds the validated input for updating a tournament.
// Nil pointer fields are left unchanged.
type UpdateTournamentInput struct {
	Name           *string
	Type           *string
	TableCount     *int
	Description    *string
	Links          *[]TournamentLink
	Location       *string
	EventDate      *time.Time
	Status         *string
	VotingStart    *time.Time
	VotingEnd      *time.Time
	ActiveCriteria *[]string
}

// TournamentService contains business logic for tournament management.
type TournamentService struct {
	tourneyRepo domain.TournamentRepository
	memberRepo  domain.TournamentMemberRepository
	userRepo    domain.UserRepository
}

// NewTournamentService creates a new TournamentService.
func NewTournamentService(
	tourneyRepo domain.TournamentRepository,
	memberRepo domain.TournamentMemberRepository,
	userRepo domain.UserRepository,
) *TournamentService {
	return &TournamentService{
		tourneyRepo: tourneyRepo,
		memberRepo:  memberRepo,
		userRepo:    userRepo,
	}
}

// List returns up to 5 upcoming and 5 past tournaments.
func (s *TournamentService) List(ctx context.Context) (upcoming, past []domain.Tournament, err error) {
	upcoming, err = s.tourneyRepo.ListUpcoming(ctx, defaultListLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("listing upcoming tournaments: %w", err)
	}
	past, err = s.tourneyRepo.ListPast(ctx, defaultListLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("listing past tournaments: %w", err)
	}
	return upcoming, past, nil
}

// Create creates a new tournament and registers the caller as organizer.
func (s *TournamentService) Create(ctx context.Context, userID uuid.UUID, input CreateTournamentInput) (*domain.Tournament, error) {
	slug, err := s.uniqueSlug(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("generating slug: %w", err)
	}

	criteria := input.ActiveCriteria
	if len(criteria) == 0 {
		criteria = defaultActiveCriteria
	}
	criteriaJSON, err := json.Marshal(criteria)
	if err != nil {
		return nil, fmt.Errorf("marshaling active_criteria: %w", err)
	}

	linksJSON, err := json.Marshal(input.Links)
	if err != nil {
		return nil, fmt.Errorf("marshaling links: %w", err)
	}

	t := &domain.Tournament{
		ID:             uuid.New(),
		Slug:           slug,
		Name:           input.Name,
		Type:           input.Type,
		Description:    input.Description,
		Links:          string(linksJSON),
		Location:       input.Location,
		EventDate:      input.EventDate,
		OrganizerID:    userID,
		TableCount:     input.TableCount,
		Status:         "draft",
		VotingStart:    input.VotingStart,
		VotingEnd:      input.VotingEnd,
		ActiveCriteria: string(criteriaJSON),
		ResultConfig:   defaultResultConfig,
	}

	if err := s.tourneyRepo.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("creating tournament: %w", err)
	}

	if err := s.memberRepo.AddMember(ctx, t.ID, userID, roleOrganizer); err != nil {
		return nil, fmt.Errorf("adding organizer membership: %w", err)
	}

	return t, nil
}

// GetBySlug returns a tournament by its slug.
func (s *TournamentService) GetBySlug(ctx context.Context, slug string) (*domain.Tournament, error) {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("GetBySlug: %w", err)
	}
	return t, nil
}

// Update applies non-nil fields from input to the tournament identified by slug.
// Caller must be organizer or helper.
func (s *TournamentService) Update(ctx context.Context, userID uuid.UUID, slug string, input UpdateTournamentInput) (*domain.Tournament, error) {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("Update: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil || (role != roleOrganizer && role != roleHelper) {
		return nil, domain.ErrForbidden
	}

	if input.Name != nil {
		t.Name = *input.Name
	}
	if input.Type != nil {
		t.Type = *input.Type
	}
	if input.TableCount != nil && t.Status == "draft" {
		t.TableCount = *input.TableCount
	}
	if input.Description != nil {
		t.Description = input.Description
	}
	if input.Location != nil {
		t.Location = input.Location
	}
	if input.EventDate != nil {
		t.EventDate = input.EventDate
	}
	if input.Status != nil {
		t.Status = *input.Status
	}
	if input.VotingStart != nil {
		t.VotingStart = input.VotingStart
	}
	if input.VotingEnd != nil {
		t.VotingEnd = input.VotingEnd
	}
	if input.Links != nil {
		linksJSON, err := json.Marshal(*input.Links)
		if err != nil {
			return nil, fmt.Errorf("marshaling links: %w", err)
		}
		t.Links = string(linksJSON)
	}
	if input.ActiveCriteria != nil {
		criteriaJSON, err := json.Marshal(*input.ActiveCriteria)
		if err != nil {
			return nil, fmt.Errorf("marshaling active_criteria: %w", err)
		}
		t.ActiveCriteria = string(criteriaJSON)
	}

	if err := s.tourneyRepo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("Update tournament: %w", err)
	}
	return t, nil
}

// Delete removes a tournament. Only the organizer may delete.
func (s *TournamentService) Delete(ctx context.Context, userID uuid.UUID, slug string) error {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil || role != roleOrganizer {
		return domain.ErrForbidden
	}

	if err := s.tourneyRepo.Delete(ctx, t.ID); err != nil {
		return fmt.Errorf("Delete tournament: %w", err)
	}
	return nil
}

// AddMember adds a user as helper by their invite code.
// Caller must be organizer.
func (s *TournamentService) AddMember(ctx context.Context, userID uuid.UUID, slug, inviteCode string) (*domain.User, error) {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("AddMember: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil || role != roleOrganizer {
		return nil, domain.ErrForbidden
	}

	target, err := s.userRepo.FindByHelperInviteCode(ctx, inviteCode)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("FindByHelperInviteCode: %w", err)
	}

	if err := s.memberRepo.AddMember(ctx, t.ID, target.ID, roleHelper); err != nil {
		return nil, fmt.Errorf("AddMember membership: %w", err)
	}

	return target, nil
}

// RemoveMember removes a helper from the tournament.
// Caller must be organizer; the organizer cannot remove themselves.
func (s *TournamentService) RemoveMember(ctx context.Context, userID uuid.UUID, slug string, targetUserID uuid.UUID) error {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return fmt.Errorf("RemoveMember: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil || role != roleOrganizer {
		return domain.ErrForbidden
	}

	if targetUserID == t.OrganizerID {
		return domain.ErrForbidden
	}

	if err := s.memberRepo.RemoveMember(ctx, t.ID, targetUserID); err != nil {
		return fmt.Errorf("RemoveMember membership: %w", err)
	}
	return nil
}

// LeaveTournament allows a helper to remove themselves from a tournament.
// The organizer cannot leave their own tournament.
func (s *TournamentService) LeaveTournament(ctx context.Context, userID uuid.UUID, slug string) error {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return fmt.Errorf("LeaveTournament: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("LeaveTournament GetRole: %w", err)
	}

	if role == roleOrganizer {
		return domain.ErrForbidden
	}

	if err := s.memberRepo.RemoveMember(ctx, t.ID, userID); err != nil {
		return fmt.Errorf("LeaveTournament RemoveMember: %w", err)
	}
	return nil
}

// generateSlug converts a tournament name to a URL-safe slug.
func generateSlug(name string) string {
	s := strings.ToLower(name)
	s = slugReplacer.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > maxSlugLength {
		s = s[:maxSlugLength]
	}
	return s
}

// uniqueSlug finds an available slug for the given name, trying up to maxSlugAttempts suffixes.
func (s *TournamentService) uniqueSlug(ctx context.Context, name string) (string, error) {
	base := generateSlug(name)
	candidate := base

	for i := 2; i <= maxSlugAttempts+1; i++ {
		exists, err := s.tourneyRepo.ExistsBySlug(ctx, candidate)
		if err != nil {
			return "", fmt.Errorf("checking slug existence: %w", err)
		}
		if !exists {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	return "", fmt.Errorf("could not generate unique slug after %d attempts", maxSlugAttempts)
}
