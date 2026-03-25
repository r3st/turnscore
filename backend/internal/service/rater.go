package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"regexp"

	"github.com/google/uuid"

	"github.com/r3st/turnscore/internal/domain"
)

const (
	codePattern      = `^\d{4,6}$`
	maxCodeAttempts  = 10
	codeMin          = 1000
	codeRange        = 9000 // 1000..9999
)

var codeRegexp = regexp.MustCompile(codePattern)

// RaterService contains business logic for rater and rating management.
type RaterService struct {
	raterRepo   domain.RaterRepository
	ratingRepo  domain.RatingRepository
	tableRepo   domain.TableRepository
	memberRepo  domain.TournamentMemberRepository
	tourneyRepo domain.TournamentRepository
	userRepo    domain.UserRepository
}

// NewRaterService creates a new RaterService.
func NewRaterService(
	raterRepo domain.RaterRepository,
	ratingRepo domain.RatingRepository,
	tableRepo domain.TableRepository,
	memberRepo domain.TournamentMemberRepository,
	tourneyRepo domain.TournamentRepository,
	userRepo domain.UserRepository,
) *RaterService {
	return &RaterService{
		raterRepo:   raterRepo,
		ratingRepo:  ratingRepo,
		tableRepo:   tableRepo,
		memberRepo:  memberRepo,
		tourneyRepo: tourneyRepo,
		userRepo:    userRepo,
	}
}

// ListRaters returns all raters for a tournament.
// Caller must be organizer or helper.
func (s *RaterService) ListRaters(ctx context.Context, userID uuid.UUID, slug string) ([]domain.Rater, error) {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("ListRaters find tournament: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil || (role != roleOrganizer && role != roleHelper) {
		return nil, domain.ErrForbidden
	}

	raters, err := s.raterRepo.ListByTournamentID(ctx, t.ID)
	if err != nil {
		return nil, fmt.Errorf("ListRaters list: %w", err)
	}
	return raters, nil
}

// CreateRater creates a new rater for a tournament.
// Caller must be organizer or helper.
// Returns (rater, rawCode, error). rawCode is only returned at creation.
func (s *RaterService) CreateRater(ctx context.Context, userID uuid.UUID, slug, nickname string, code *string) (*domain.Rater, string, error) {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, "", fmt.Errorf("CreateRater find tournament: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil || (role != roleOrganizer && role != roleHelper) {
		return nil, "", domain.ErrForbidden
	}

	// Validate or generate code.
	var rawCode string
	if code != nil {
		if !codeRegexp.MatchString(*code) {
			return nil, "", domain.ErrConflict // reuse conflict for bad input – caller maps to 400
		}
		rawCode = *code
	} else {
		rawCode, err = s.generateUniqueCode(ctx, t.ID)
		if err != nil {
			return nil, "", fmt.Errorf("CreateRater generate code: %w", err)
		}
	}

	// Check nickname uniqueness.
	exists, err := s.raterRepo.ExistsByNickname(ctx, t.ID, nickname)
	if err != nil {
		return nil, "", fmt.Errorf("CreateRater check nickname: %w", err)
	}
	if exists {
		return nil, "", domain.ErrConflict
	}

	rater := &domain.Rater{
		ID:           uuid.New(),
		TournamentID: t.ID,
		Nickname:     nickname,
		Code:         rawCode,
	}

	if err := s.raterRepo.Create(ctx, rater); err != nil {
		return nil, "", fmt.Errorf("CreateRater create: %w", err)
	}

	return rater, rawCode, nil
}

// SubmitRatingInput groups all parameters for SubmitRating to keep the signature compact.
type SubmitRatingInput struct {
	RaterID      uuid.UUID
	TournamentID uuid.UUID
	Slug         string
	TableNumber  int
	Scores       map[string]int
	PlayedZone   string
	Comment      *string
}

// SubmitRating submits a rating for a table.
// RaterID and TournamentID come from the rater's JWT via context.
func (s *RaterService) SubmitRating(ctx context.Context, input SubmitRatingInput) error {
	t, err := s.tourneyRepo.FindBySlug(ctx, input.Slug)
	if err != nil {
		return fmt.Errorf("SubmitRating find tournament: %w", err)
	}

	// Verify tournament is in voting status.
	if t.Status != "voting" {
		return domain.ErrVotingNotActive
	}

	// Verify rater belongs to this tournament.
	if t.ID != input.TournamentID {
		return domain.ErrForbidden
	}

	// Find the target table.
	table, err := s.tableRepo.FindByTournamentAndNumber(ctx, t.ID, input.TableNumber)
	if err != nil {
		return fmt.Errorf("SubmitRating find table: %w", err)
	}

	// Duplicate check.
	dup, err := s.ratingRepo.ExistsByTableAndRater(ctx, table.ID, input.RaterID)
	if err != nil {
		return fmt.Errorf("SubmitRating duplicate check: %w", err)
	}
	if dup {
		return domain.ErrDuplicateRating
	}

	// Marshal criteria scores.
	scoresJSON, err := json.Marshal(input.Scores)
	if err != nil {
		return fmt.Errorf("SubmitRating marshal scores: %w", err)
	}

	rating := &domain.Rating{
		ID:             uuid.New(),
		TableID:        table.ID,
		RaterID:        input.RaterID,
		CriteriaScores: string(scoresJSON),
		PlayedZone:     input.PlayedZone,
		Comment:        input.Comment,
	}

	if err := s.ratingRepo.Create(ctx, rating); err != nil {
		return fmt.Errorf("SubmitRating create: %w", err)
	}

	return nil
}

// GetResults returns tournament results for organizer/helper.
// Also returns a map of comments per table when show_comments is enabled in result config.
func (s *RaterService) GetResults(ctx context.Context, userID uuid.UUID, slug string) (*domain.Tournament, []domain.RatingResult, map[uuid.UUID][]string, error) {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("GetResults find tournament: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil || (role != roleOrganizer && role != roleHelper) {
		return nil, nil, nil, domain.ErrForbidden
	}

	results, err := s.ratingRepo.GetResultsForTournament(ctx, t.ID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("GetResults fetch results: %w", err)
	}

	// Fetch comments if show_comments is enabled in result config.
	var commentsMap map[uuid.UUID][]string
	var resultCfg struct {
		ShowComments bool `json:"show_comments"`
	}
	if jsonErr := json.Unmarshal([]byte(t.ResultConfig), &resultCfg); jsonErr == nil && resultCfg.ShowComments {
		commentsMap, err = s.ratingRepo.GetCommentsForTournament(ctx, t.ID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("GetResults fetch comments: %w", err)
		}
	}

	return t, results, commentsMap, nil
}

// UpdateResultConfig updates the result visibility config for a tournament.
// Caller must be organizer (not helper).
func (s *RaterService) UpdateResultConfig(ctx context.Context, userID uuid.UUID, slug string, showComments bool, visibleCriteria []string) (*domain.Tournament, error) {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("UpdateResultConfig find tournament: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil || role != roleOrganizer {
		return nil, domain.ErrForbidden
	}

	type resultConfig struct {
		ShowComments           bool     `json:"show_comments"`
		VisibleCommentCriteria []string `json:"visible_comment_criteria"`
	}
	cfg := resultConfig{
		ShowComments:           showComments,
		VisibleCommentCriteria: visibleCriteria,
	}
	if cfg.VisibleCommentCriteria == nil {
		cfg.VisibleCommentCriteria = []string{}
	}

	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("UpdateResultConfig marshal: %w", err)
	}
	t.ResultConfig = string(cfgJSON)

	if err := s.tourneyRepo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("UpdateResultConfig update: %w", err)
	}

	return t, nil
}

// generateUniqueCode tries up to maxCodeAttempts to find an unused 4-digit code.
func (s *RaterService) generateUniqueCode(ctx context.Context, tournamentID uuid.UUID) (string, error) {
	for i := 0; i < maxCodeAttempts; i++ {
		code := fmt.Sprintf("%04d", codeMin+rand.IntN(codeRange))
		exists, err := s.raterRepo.ExistsByCode(ctx, tournamentID, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", fmt.Errorf("could not generate unique code after %d attempts", maxCodeAttempts)
}
