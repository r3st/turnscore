package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Table is the GORM model mapped to the tables table.
type Table struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	TournamentID uuid.UUID `gorm:"type:uuid;not null"`
	Number       int       `gorm:"not null"`
	Name         *string
	Description  *string
	QRCodePath   *string  `gorm:"column:qr_code_path"`
	Photos       []Photo  `gorm:"foreignKey:TableID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Photo is the GORM model mapped to the photos table.
type Photo struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TableID      uuid.UUID  `gorm:"type:uuid;not null"`
	URL          string     `gorm:"not null"`
	ThumbnailURL *string    `gorm:"column:thumbnail_url"`
	Category     string     `gorm:"not null;default:general"`
	UploadedBy   *uuid.UUID `gorm:"type:uuid;column:uploaded_by"`
	CreatedAt    time.Time
}

// TableRepository defines data access for tables.
type TableRepository interface {
	ListByTournamentID(ctx context.Context, tournamentID uuid.UUID) ([]Table, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Table, error)
	FindByTournamentAndNumber(ctx context.Context, tournamentID uuid.UUID, number int) (*Table, error)
	CountByTournamentID(ctx context.Context, tournamentID uuid.UUID) (int64, error)
	CreateBatch(ctx context.Context, tables []Table) error
	Update(ctx context.Context, t *Table) error
	UpdateQRCodePath(ctx context.Context, tableID uuid.UUID, path string) error
}

// PhotoRepository defines data access for photos.
type PhotoRepository interface {
	Create(ctx context.Context, p *Photo) error
	FindByID(ctx context.Context, id uuid.UUID) (*Photo, error)
	FindByTableID(ctx context.Context, tableID uuid.UUID) ([]Photo, error)
	Delete(ctx context.Context, id uuid.UUID) error
	BelongsToTournament(ctx context.Context, photoID, tournamentID uuid.UUID) (bool, error)
}
