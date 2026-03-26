package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/r3st/turnscore/internal/api/generated"
	"github.com/r3st/turnscore/internal/domain"
	"github.com/r3st/turnscore/internal/service"
	"github.com/r3st/turnscore/internal/storage"
)

const (
	msgTableNotFound = "table not found"
	msgPhotoNotFound = "photo not found"
	msgTablesExist   = "tables already exist for this tournament"
	msgInvalidFile   = "invalid file type or file too large"
)

// --- Table handlers ---

// ListTables returns all tables for a tournament.
func (h *Handlers) ListTables(ctx context.Context, req generated.ListTablesRequestObject) (generated.ListTablesResponseObject, error) {
	tables, err := h.tableSvc.ListBySlug(ctx, req.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.ListTables404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentNotFound},
			}, nil
		}
		return nil, err
	}
	return generated.ListTables200JSONResponse(mapTableSummaries(tables, h.tableSvc.Storage())), nil
}

// CreateTables creates all tables (1..TableCount) for a tournament.
func (h *Handlers) CreateTables(ctx context.Context, req generated.CreateTablesRequestObject) (generated.CreateTablesResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.CreateTables401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	tables, err := h.tableSvc.CreateForTournament(ctx, userID, req.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.CreateTables404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.CreateTables403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenUpdate},
			}, nil
		}
		if errors.Is(err, domain.ErrConflict) {
			return generated.CreateTables409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse{Code: "conflict", Message: msgTablesExist},
			}, nil
		}
		return nil, err
	}
	return generated.CreateTables201JSONResponse(mapTableSummaries(tables, h.tableSvc.Storage())), nil
}

// GetTable returns details for a single table.
func (h *Handlers) GetTable(ctx context.Context, req generated.GetTableRequestObject) (generated.GetTableResponseObject, error) {
	table, tourney, err := h.tableSvc.GetBySlugAndNumber(ctx, req.Slug, req.Number)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.GetTable404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTableNotFound},
			}, nil
		}
		return nil, err
	}

	activeCriteria, parseErr := parseCriteria(tourney.ActiveCriteria)
	if parseErr != nil {
		return nil, parseErr
	}

	stor := h.tableSvc.Storage()
	detail := generated.TableDetail{
		Id:             table.ID,
		Number:         table.Number,
		Name:           table.Name,
		Description:    table.Description,
		Photos:         mapPhotos(table.Photos, stor),
		TournamentSlug: tourney.Slug,
		TournamentType: generated.TournamentType(tourney.Type),
		TournamentName: tourney.Name,
		ActiveCriteria: activeCriteria,
	}
	if table.QRCodePath != nil {
		u := stor.PublicURL(*table.QRCodePath)
		detail.QrCodeUrl = &u
	}

	return generated.GetTable200JSONResponse(detail), nil
}

// UpdateTable updates the name and/or description of a table.
func (h *Handlers) UpdateTable(ctx context.Context, req generated.UpdateTableRequestObject) (generated.UpdateTableResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.UpdateTable401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	var name, description *string
	if req.Body != nil {
		name = req.Body.Name
		description = req.Body.Description
	}

	table, err := h.tableSvc.Update(ctx, userID, req.Slug, req.Number, name, description)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.UpdateTable404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTableNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.UpdateTable403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenUpdate},
			}, nil
		}
		return nil, err
	}
	return generated.UpdateTable200JSONResponse(mapTableSummary(table, h.tableSvc.Storage())), nil
}

// UploadPhoto handles multipart photo upload for a table.
// req.Body is a *multipart.Reader; we iterate NextPart() to find "category" and "photo" fields.
func (h *Handlers) UploadPhoto(ctx context.Context, req generated.UploadPhotoRequestObject) (generated.UploadPhotoResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.UploadPhoto401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	if req.Body == nil {
		return generated.UploadPhoto400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgBodyRequired},
		}, nil
	}

	var category string
	var fileData []byte
	var filename string

	for {
		part, partErr := req.Body.NextPart()
		if partErr == io.EOF {
			break
		}
		if partErr != nil {
			return generated.UploadPhoto400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgInvalidFile},
			}, nil
		}

		fieldName := part.FormName()
		switch fieldName {
		case "category":
			raw, readErr := io.ReadAll(part)
			if readErr != nil {
				return generated.UploadPhoto400JSONResponse{
					BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgInvalidFile},
				}, nil
			}
			category = string(raw)
		case "photo":
			filename = part.FileName()
			raw, readErr := io.ReadAll(part)
			if readErr != nil {
				return generated.UploadPhoto400JSONResponse{
					BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgInvalidFile},
				}, nil
			}
			fileData = raw
		}
	}

	if len(fileData) == 0 {
		return generated.UploadPhoto400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgInvalidFile},
		}, nil
	}

	input := service.UploadPhotoInput{
		File:        bytes.NewReader(fileData),
		Filename:    filename,
		ContentType: filenameToMIME(filename),
		Category:    category,
		Size:        int64(len(fileData)),
	}

	photo, err := h.tableSvc.UploadPhoto(ctx, userID, req.Slug, req.Number, input)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.UploadPhoto404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTableNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.UploadPhoto403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenUpdate},
			}, nil
		}
		if errors.Is(err, domain.ErrFileTooLarge) || errors.Is(err, domain.ErrInvalidFileType) {
			return generated.UploadPhoto400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgInvalidFile},
			}, nil
		}
		return nil, err
	}
	return generated.UploadPhoto201JSONResponse(mapPhoto(photo, h.tableSvc.Storage())), nil
}

// DeletePhoto removes a photo by ID.
func (h *Handlers) DeletePhoto(ctx context.Context, req generated.DeletePhotoRequestObject) (generated.DeletePhotoResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.DeletePhoto401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	if err := h.tableSvc.DeletePhoto(ctx, userID, req.PhotoId); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.DeletePhoto404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgPhotoNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.DeletePhoto403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenUpdate},
			}, nil
		}
		return nil, err
	}
	return generated.DeletePhoto204Response{}, nil
}

// DeleteTable removes a table (organizer only, draft status).
func (h *Handlers) DeleteTable(ctx context.Context, req generated.DeleteTableRequestObject) (generated.DeleteTableResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.DeleteTable401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	if err := h.tableSvc.DeleteTable(ctx, userID, req.Slug, req.Number); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.DeleteTable404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTableNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.DeleteTable403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenUpdate},
			}, nil
		}
		return nil, err
	}
	return generated.DeleteTable204Response{}, nil
}

// GenerateQRCode generates a QR code PNG for a table and returns it as image/png.
func (h *Handlers) GenerateQRCode(ctx context.Context, req generated.GenerateQRCodeRequestObject) (generated.GenerateQRCodeResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.GenerateQRCode401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	pngData, err := h.qrSvc.GenerateForTable(ctx, userID, req.Slug, req.Number)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.GenerateQRCode404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTableNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.GenerateQRCode403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenUpdate},
			}, nil
		}
		return nil, err
	}
	return generated.GenerateQRCode200ImagepngResponse{
		Body:          bytes.NewReader(pngData),
		ContentLength: int64(len(pngData)),
	}, nil
}

// ExportQRCodesPDF generates a PDF with QR cards for all tables.
func (h *Handlers) ExportQRCodesPDF(ctx context.Context, req generated.ExportQRCodesPDFRequestObject) (generated.ExportQRCodesPDFResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.ExportQRCodesPDF401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	pdfData, err := h.qrSvc.ExportPDF(ctx, userID, req.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.ExportQRCodesPDF404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.ExportQRCodesPDF403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenUpdate},
			}, nil
		}
		return nil, err
	}
	return generated.ExportQRCodesPDF200ApplicationpdfResponse{
		Body:          bytes.NewReader(pdfData),
		ContentLength: int64(len(pdfData)),
	}, nil
}

// --- Mapping helpers ---

func mapTableSummaries(tables []domain.Table, stor storage.Storage) []generated.TableSummary {
	result := make([]generated.TableSummary, 0, len(tables))
	for i := range tables {
		result = append(result, mapTableSummary(&tables[i], stor))
	}
	return result
}

func mapTableSummary(t *domain.Table, stor storage.Storage) generated.TableSummary {
	s := generated.TableSummary{
		Id:     t.ID,
		Number: t.Number,
		Name:   t.Name,
		Photos: mapPhotos(t.Photos, stor),
	}
	if t.QRCodePath != nil {
		u := stor.PublicURL(*t.QRCodePath)
		s.QrCodeUrl = &u
	}
	return s
}

func mapPhotos(photos []domain.Photo, stor storage.Storage) []generated.Photo {
	result := make([]generated.Photo, 0, len(photos))
	for i := range photos {
		result = append(result, mapPhoto(&photos[i], stor))
	}
	return result
}

func mapPhoto(p *domain.Photo, stor storage.Storage) generated.Photo {
	gp := generated.Photo{
		Id:       p.ID,
		Url:      stor.PublicURL(p.URL),
		Category: generated.PhotoCategory(p.Category),
	}
	if p.ThumbnailURL != nil {
		u := stor.PublicURL(*p.ThumbnailURL)
		gp.ThumbnailUrl = &u
	}
	return gp
}

// filenameToMIME maps a filename extension to a MIME type, defaulting to image/jpeg.
func filenameToMIME(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	default:
		return "image/jpeg"
	}
}
