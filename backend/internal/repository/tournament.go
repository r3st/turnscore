package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/r3st/turnscore/internal/domain"
)

// TournamentRepository implements domain.TournamentRepository using GORM.
type TournamentRepository struct {
	db *gorm.DB
}

// NewTournamentRepository creates a new TournamentRepository backed by the given DB.
func NewTournamentRepository(db *gorm.DB) *TournamentRepository {
	return &TournamentRepository{db: db}
}

// FindBySlug returns a tournament by its URL slug.
func (r *TournamentRepository) FindBySlug(ctx context.Context, slug string) (*domain.Tournament, error) {
	var t domain.Tournament
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("FindBySlug: %w", err)
	}
	return &t, nil
}

// FindByID returns a tournament by its internal UUID.
func (r *TournamentRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tournament, error) {
	var t domain.Tournament
	err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("FindByID: %w", err)
	}
	return &t, nil
}

// ExistsBySlug returns true if a tournament with the given slug exists.
func (r *TournamentRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Tournament{}).
		Where("slug = ?", slug).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("ExistsBySlug: %w", err)
	}
	return count > 0, nil
}

// ListUpcoming returns tournaments that are not yet archived or are in the future,
// ordered by event_date ASC NULLS LAST.
func (r *TournamentRepository) ListUpcoming(ctx context.Context, limit int) ([]domain.Tournament, error) {
	var ts []domain.Tournament
	err := r.db.WithContext(ctx).
		Where("status IN ('draft','active') AND (event_date >= CURRENT_DATE OR event_date IS NULL)").
		Order("event_date ASC NULLS LAST").
		Limit(limit).
		Find(&ts).Error
	if err != nil {
		return nil, fmt.Errorf("ListUpcoming: %w", err)
	}
	return ts, nil
}

// ListPast returns archived tournaments or those with a past event_date,
// ordered by event_date DESC.
func (r *TournamentRepository) ListPast(ctx context.Context, limit int) ([]domain.Tournament, error) {
	var ts []domain.Tournament
	err := r.db.WithContext(ctx).
		Where("status = 'archived' OR event_date < CURRENT_DATE").
		Order("event_date DESC").
		Limit(limit).
		Find(&ts).Error
	if err != nil {
		return nil, fmt.Errorf("ListPast: %w", err)
	}
	return ts, nil
}

// Create persists a new tournament to the database.
func (r *TournamentRepository) Create(ctx context.Context, t *domain.Tournament) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return fmt.Errorf("Create tournament: %w", err)
	}
	return nil
}

// Update saves all fields of an existing tournament.
func (r *TournamentRepository) Update(ctx context.Context, t *domain.Tournament) error {
	if err := r.db.WithContext(ctx).Save(t).Error; err != nil {
		return fmt.Errorf("Update tournament: %w", err)
	}
	return nil
}

// Delete removes a tournament by ID.
func (r *TournamentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Tournament{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("Delete tournament: %w", err)
	}
	return nil
}

// ListByOrganizerID returns all tournaments owned by the given user.
func (r *TournamentRepository) ListByOrganizerID(ctx context.Context, organizerID uuid.UUID) ([]domain.Tournament, error) {
	var tournaments []domain.Tournament
	if err := r.db.WithContext(ctx).Where("organizer_id = ?", organizerID).Find(&tournaments).Error; err != nil {
		return nil, fmt.Errorf("ListByOrganizerID: %w", err)
	}
	return tournaments, nil
}
