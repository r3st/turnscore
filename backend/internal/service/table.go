package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // PNG decoder
	"io"

	"github.com/google/uuid"
	_ "golang.org/x/image/webp" // WebP decoder

	"golang.org/x/image/draw"

	"github.com/r3st/turnscore/internal/domain"
	"github.com/r3st/turnscore/internal/storage"
)

const (
	maxPhotoSize       = 10 * 1024 * 1024 // 10 MB
	thumbnailMaxDim    = 400
	mimeJPEG           = "image/jpeg"
	mimePNG            = "image/png"
	mimeWebP           = "image/webp"
)

var allowedPhotoTypes = map[string]bool{
	mimeJPEG: true,
	mimePNG:  true,
	mimeWebP: true,
}

// UploadPhotoInput groups the file metadata for UploadPhoto to keep the parameter count low.
type UploadPhotoInput struct {
	File        io.Reader
	Filename    string
	ContentType string
	Category    string
	Size        int64
}

// TableService contains business logic for table and photo management.
type TableService struct {
	tableRepo   domain.TableRepository
	memberRepo  domain.TournamentMemberRepository
	tourneyRepo domain.TournamentRepository
	photoRepo   domain.PhotoRepository
	storage     storage.Storage
}

// NewTableService creates a new TableService.
func NewTableService(
	tableRepo domain.TableRepository,
	memberRepo domain.TournamentMemberRepository,
	tourneyRepo domain.TournamentRepository,
	photoRepo domain.PhotoRepository,
	stor storage.Storage,
) *TableService {
	return &TableService{
		tableRepo:   tableRepo,
		memberRepo:  memberRepo,
		tourneyRepo: tourneyRepo,
		photoRepo:   photoRepo,
		storage:     stor,
	}
}

// Storage returns the underlying storage backend (used by handlers for URL mapping).
func (s *TableService) Storage() storage.Storage {
	return s.storage
}

// ListBySlug returns all tables for the tournament identified by slug.
func (s *TableService) ListBySlug(ctx context.Context, slug string) ([]domain.Table, error) {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("ListBySlug find tournament: %w", err)
	}
	tables, err := s.tableRepo.ListByTournamentID(ctx, t.ID)
	if err != nil {
		return nil, fmt.Errorf("ListBySlug list tables: %w", err)
	}
	return tables, nil
}

// GetBySlugAndNumber returns the table and its tournament by slug and table number.
func (s *TableService) GetBySlugAndNumber(ctx context.Context, slug string, number int) (*domain.Table, *domain.Tournament, error) {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, nil, fmt.Errorf("GetBySlugAndNumber find tournament: %w", err)
	}
	table, err := s.tableRepo.FindByTournamentAndNumber(ctx, t.ID, number)
	if err != nil {
		return nil, nil, fmt.Errorf("GetBySlugAndNumber find table: %w", err)
	}
	return table, t, nil
}

// CreateForTournament creates tables 1..TableCount for the tournament.
// Caller must be organizer or helper; returns ErrConflict if tables already exist.
func (s *TableService) CreateForTournament(ctx context.Context, userID uuid.UUID, slug string) ([]domain.Table, error) {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("CreateForTournament find tournament: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil || (role != roleOrganizer && role != roleHelper) {
		return nil, domain.ErrForbidden
	}

	count, err := s.tableRepo.CountByTournamentID(ctx, t.ID)
	if err != nil {
		return nil, fmt.Errorf("CreateForTournament count: %w", err)
	}
	if count > 0 {
		return nil, domain.ErrConflict
	}

	tables := make([]domain.Table, 0, t.TableCount)
	for i := 1; i <= t.TableCount; i++ {
		tables = append(tables, domain.Table{
			ID:           uuid.New(),
			TournamentID: t.ID,
			Number:       i,
		})
	}

	if err := s.tableRepo.CreateBatch(ctx, tables); err != nil {
		return nil, fmt.Errorf("CreateForTournament create batch: %w", err)
	}

	return tables, nil
}

// Update updates the name and/or description of a table.
// Caller must be organizer or helper.
func (s *TableService) Update(ctx context.Context, userID uuid.UUID, slug string, number int, name, description *string) (*domain.Table, error) {
	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("Update find tournament: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil || (role != roleOrganizer && role != roleHelper) {
		return nil, domain.ErrForbidden
	}

	table, err := s.tableRepo.FindByTournamentAndNumber(ctx, t.ID, number)
	if err != nil {
		return nil, fmt.Errorf("Update find table: %w", err)
	}

	if name != nil {
		table.Name = name
	}
	if description != nil {
		table.Description = description
	}

	if err := s.tableRepo.Update(ctx, table); err != nil {
		return nil, fmt.Errorf("Update save table: %w", err)
	}

	return table, nil
}

// UploadPhoto validates, thumbnails, stores, and records a photo for a table.
// Caller must be organizer or helper.
func (s *TableService) UploadPhoto(
	ctx context.Context,
	userID uuid.UUID,
	slug string,
	number int,
	input UploadPhotoInput,
) (*domain.Photo, error) {
	if input.Size > maxPhotoSize {
		return nil, domain.ErrFileTooLarge
	}
	if !allowedPhotoTypes[input.ContentType] {
		return nil, domain.ErrInvalidFileType
	}

	t, err := s.tourneyRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("UploadPhoto find tournament: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, t.ID, userID)
	if err != nil || (role != roleOrganizer && role != roleHelper) {
		return nil, domain.ErrForbidden
	}

	table, err := s.tableRepo.FindByTournamentAndNumber(ctx, t.ID, number)
	if err != nil {
		return nil, fmt.Errorf("UploadPhoto find table: %w", err)
	}

	// Read full file bytes for processing.
	data, err := io.ReadAll(input.File)
	if err != nil {
		return nil, fmt.Errorf("UploadPhoto read file: %w", err)
	}
	if int64(len(data)) > maxPhotoSize {
		return nil, domain.ErrFileTooLarge
	}

	// Decode image for thumbnail generation.
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("UploadPhoto decode image: %w", err)
	}

	// Generate storage keys.
	photoID := uuid.New()
	origKey := fmt.Sprintf("tables/%s/%s/%s.jpg", table.ID, input.Category, photoID)
	thumbKey := fmt.Sprintf("tables/%s/%s/thumb_%s.jpg", table.ID, input.Category, photoID)

	// Store original as JPEG.
	var origBuf bytes.Buffer
	if err := jpeg.Encode(&origBuf, src, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("UploadPhoto encode original: %w", err)
	}
	if err := s.storage.Put(ctx, origKey, &origBuf, int64(origBuf.Len()), mimeJPEG); err != nil {
		return nil, fmt.Errorf("UploadPhoto store original: %w", err)
	}

	// Generate and store thumbnail.
	thumb := generateThumbnail(src, thumbnailMaxDim)
	var thumbBuf bytes.Buffer
	if err := jpeg.Encode(&thumbBuf, thumb, &jpeg.Options{Quality: 80}); err != nil {
		return nil, fmt.Errorf("UploadPhoto encode thumbnail: %w", err)
	}
	if err := s.storage.Put(ctx, thumbKey, &thumbBuf, int64(thumbBuf.Len()), mimeJPEG); err != nil {
		return nil, fmt.Errorf("UploadPhoto store thumbnail: %w", err)
	}

	// Persist the photo record. URL and ThumbnailURL store storage keys.
	p := &domain.Photo{
		ID:           photoID,
		TableID:      table.ID,
		URL:          origKey,
		ThumbnailURL: &thumbKey,
		Category:     input.Category,
		UploadedBy:   &userID,
	}
	if err := s.photoRepo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("UploadPhoto create record: %w", err)
	}

	return p, nil
}

// DeletePhoto removes a photo from storage and the database.
// Caller must be organizer or helper of the tournament that owns the photo.
func (s *TableService) DeletePhoto(ctx context.Context, userID uuid.UUID, photoID uuid.UUID) error {
	photo, err := s.photoRepo.FindByID(ctx, photoID)
	if err != nil {
		return fmt.Errorf("DeletePhoto find photo: %w", err)
	}

	table, err := s.tableRepo.FindByID(ctx, photo.TableID)
	if err != nil {
		return fmt.Errorf("DeletePhoto find table: %w", err)
	}

	role, err := s.memberRepo.GetRole(ctx, table.TournamentID, userID)
	if err != nil || (role != roleOrganizer && role != roleHelper) {
		return domain.ErrForbidden
	}

	// Delete from storage.
	if delErr := s.storage.Delete(ctx, photo.URL); delErr != nil {
		return fmt.Errorf("DeletePhoto delete original: %w", delErr)
	}
	if photo.ThumbnailURL != nil {
		if delErr := s.storage.Delete(ctx, *photo.ThumbnailURL); delErr != nil {
			return fmt.Errorf("DeletePhoto delete thumbnail: %w", delErr)
		}
	}

	if err := s.photoRepo.Delete(ctx, photoID); err != nil {
		return fmt.Errorf("DeletePhoto delete record: %w", err)
	}

	return nil
}

// generateThumbnail scales src so that the longest side is at most maxDim.
func generateThumbnail(src image.Image, maxDim int) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return src
	}
	if w > h {
		h = h * maxDim / w
		w = maxDim
	} else {
		w = w * maxDim / h
		h = maxDim
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}
