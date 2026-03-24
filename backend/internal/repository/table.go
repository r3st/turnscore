package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/r3st/turnscore/internal/domain"
)

// TableRepository implements domain.TableRepository using GORM.
type TableRepository struct {
	db *gorm.DB
}

// NewTableRepository creates a new TableRepository backed by the given DB.
func NewTableRepository(db *gorm.DB) *TableRepository {
	return &TableRepository{db: db}
}

// ListByTournamentID returns all tables for a tournament ordered by number ASC.
func (r *TableRepository) ListByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]domain.Table, error) {
	var tables []domain.Table
	err := r.db.WithContext(ctx).
		Preload("Photos").
		Where("tournament_id = ?", tournamentID).
		Order("number ASC").
		Find(&tables).Error
	if err != nil {
		return nil, fmt.Errorf("ListByTournamentID: %w", err)
	}
	return tables, nil
}

// FindByID returns a table by its primary key UUID.
func (r *TableRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Table, error) {
	var t domain.Table
	err := r.db.WithContext(ctx).
		Preload("Photos").
		First(&t, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("FindByID table: %w", err)
	}
	return &t, nil
}

// FindByTournamentAndNumber returns a single table by tournament ID and table number.
func (r *TableRepository) FindByTournamentAndNumber(ctx context.Context, tournamentID uuid.UUID, number int) (*domain.Table, error) {
	var t domain.Table
	err := r.db.WithContext(ctx).
		Preload("Photos").
		Where("tournament_id = ? AND number = ?", tournamentID, number).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("FindByTournamentAndNumber: %w", err)
	}
	return &t, nil
}

// CountByTournamentID returns the number of tables for a tournament.
func (r *TableRepository) CountByTournamentID(ctx context.Context, tournamentID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Table{}).
		Where("tournament_id = ?", tournamentID).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("CountByTournamentID: %w", err)
	}
	return count, nil
}

// CreateBatch persists a slice of tables to the database.
func (r *TableRepository) CreateBatch(ctx context.Context, tables []domain.Table) error {
	if err := r.db.WithContext(ctx).Create(&tables).Error; err != nil {
		return fmt.Errorf("CreateBatch tables: %w", err)
	}
	return nil
}

// Update saves all fields of an existing table.
func (r *TableRepository) Update(ctx context.Context, t *domain.Table) error {
	if err := r.db.WithContext(ctx).Save(t).Error; err != nil {
		return fmt.Errorf("Update table: %w", err)
	}
	return nil
}

// UpdateQRCodePath sets the qr_code_path for the given table.
func (r *TableRepository) UpdateQRCodePath(ctx context.Context, tableID uuid.UUID, path string) error {
	err := r.db.WithContext(ctx).
		Model(&domain.Table{}).
		Where("id = ?", tableID).
		Update("qr_code_path", path).Error
	if err != nil {
		return fmt.Errorf("UpdateQRCodePath: %w", err)
	}
	return nil
}
