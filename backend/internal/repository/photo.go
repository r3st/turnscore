package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/r3st/turnscore/internal/domain"
)

// PhotoRepository implements domain.PhotoRepository using GORM.
type PhotoRepository struct {
	db *gorm.DB
}

// NewPhotoRepository creates a new PhotoRepository backed by the given DB.
func NewPhotoRepository(db *gorm.DB) *PhotoRepository {
	return &PhotoRepository{db: db}
}

// Create persists a new photo to the database.
func (r *PhotoRepository) Create(ctx context.Context, p *domain.Photo) error {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return fmt.Errorf("Create photo: %w", err)
	}
	return nil
}

// FindByID returns a photo by its UUID.
func (r *PhotoRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Photo, error) {
	var p domain.Photo
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("FindByID photo: %w", err)
	}
	return &p, nil
}

// FindByTableID returns all photos for a table ordered by created_at ASC.
func (r *PhotoRepository) FindByTableID(ctx context.Context, tableID uuid.UUID) ([]domain.Photo, error) {
	var photos []domain.Photo
	err := r.db.WithContext(ctx).
		Where("table_id = ?", tableID).
		Order("created_at ASC").
		Find(&photos).Error
	if err != nil {
		return nil, fmt.Errorf("FindByTableID photos: %w", err)
	}
	return photos, nil
}

// Delete removes a photo by its UUID.
func (r *PhotoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Photo{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("Delete photo: %w", err)
	}
	return nil
}

// BelongsToTournament checks whether a photo belongs to a table in the given tournament.
func (r *PhotoRepository) BelongsToTournament(ctx context.Context, photoID, tournamentID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Raw(
			`SELECT COUNT(*) FROM photos p
			 JOIN tables t ON p.table_id = t.id
			 WHERE p.id = ? AND t.tournament_id = ?`,
			photoID, tournamentID,
		).
		Scan(&count).Error
	if err != nil {
		return false, fmt.Errorf("BelongsToTournament: %w", err)
	}
	return count > 0, nil
}
