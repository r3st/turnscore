// Package handlers implements the StrictServerInterface for TurnScore.
package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/r3st/turnscore/internal/api/generated"
	"github.com/r3st/turnscore/internal/api/middleware"
	"github.com/r3st/turnscore/internal/domain"
	"github.com/r3st/turnscore/internal/service"
)

// Compile-time assertion that Handlers implements StrictServerInterface.
var _ generated.StrictServerInterface = (*Handlers)(nil)

// Handlers is the central handler struct implementing all API operations.
type Handlers struct {
	tournamentSvc *service.TournamentService
}

// New creates a new Handlers instance wired to the given services.
func New(tournamentSvc *service.TournamentService) *Handlers {
	return &Handlers{tournamentSvc: tournamentSvc}
}

// getUserID extracts and parses the userID from the request context (set by OptionalAuth middleware).
func getUserID(ctx context.Context) (uuid.UUID, bool) {
	raw, ok := ctx.Value(middleware.ContextKeyUserID).(string)
	if !ok || raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// errCode maps a domain sentinel error to an API error code string.
func errCode(err error) string {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return "not_found"
	case errors.Is(err, domain.ErrForbidden):
		return "forbidden"
	case errors.Is(err, domain.ErrConflict):
		return "conflict"
	case errors.Is(err, domain.ErrUnauthorized):
		return "unauthorized"
	default:
		return "internal_error"
	}
}

// --- Stub implementations for endpoints not yet implemented in this issue ---

// GoogleOAuthRedirect is not yet implemented.
func (h *Handlers) GoogleOAuthRedirect(ctx context.Context, req generated.GoogleOAuthRedirectRequestObject) (generated.GoogleOAuthRedirectResponseObject, error) {
	return nil, errors.New("not implemented")
}

// GoogleOAuthCallback is not yet implemented.
func (h *Handlers) GoogleOAuthCallback(ctx context.Context, req generated.GoogleOAuthCallbackRequestObject) (generated.GoogleOAuthCallbackResponseObject, error) {
	return nil, errors.New("not implemented")
}

// RaterLogin is not yet implemented.
func (h *Handlers) RaterLogin(ctx context.Context, req generated.RaterLoginRequestObject) (generated.RaterLoginResponseObject, error) {
	return nil, errors.New("not implemented")
}

// RefreshToken is not yet implemented.
func (h *Handlers) RefreshToken(ctx context.Context, req generated.RefreshTokenRequestObject) (generated.RefreshTokenResponseObject, error) {
	return nil, errors.New("not implemented")
}

// GetMe is not yet implemented.
func (h *Handlers) GetMe(ctx context.Context, req generated.GetMeRequestObject) (generated.GetMeResponseObject, error) {
	return nil, errors.New("not implemented")
}

// DeletePhoto is not yet implemented.
func (h *Handlers) DeletePhoto(ctx context.Context, req generated.DeletePhotoRequestObject) (generated.DeletePhotoResponseObject, error) {
	return nil, errors.New("not implemented")
}

// ExportQRCodesPDF is not yet implemented.
func (h *Handlers) ExportQRCodesPDF(ctx context.Context, req generated.ExportQRCodesPDFRequestObject) (generated.ExportQRCodesPDFResponseObject, error) {
	return nil, errors.New("not implemented")
}

// ListRaters is not yet implemented.
func (h *Handlers) ListRaters(ctx context.Context, req generated.ListRatersRequestObject) (generated.ListRatersResponseObject, error) {
	return nil, errors.New("not implemented")
}

// CreateRater is not yet implemented.
func (h *Handlers) CreateRater(ctx context.Context, req generated.CreateRaterRequestObject) (generated.CreateRaterResponseObject, error) {
	return nil, errors.New("not implemented")
}

// UpdateResultConfig is not yet implemented.
func (h *Handlers) UpdateResultConfig(ctx context.Context, req generated.UpdateResultConfigRequestObject) (generated.UpdateResultConfigResponseObject, error) {
	return nil, errors.New("not implemented")
}

// GetTournamentResults is not yet implemented.
func (h *Handlers) GetTournamentResults(ctx context.Context, req generated.GetTournamentResultsRequestObject) (generated.GetTournamentResultsResponseObject, error) {
	return nil, errors.New("not implemented")
}

// ListTables is not yet implemented.
func (h *Handlers) ListTables(ctx context.Context, req generated.ListTablesRequestObject) (generated.ListTablesResponseObject, error) {
	return nil, errors.New("not implemented")
}

// CreateTables is not yet implemented.
func (h *Handlers) CreateTables(ctx context.Context, req generated.CreateTablesRequestObject) (generated.CreateTablesResponseObject, error) {
	return nil, errors.New("not implemented")
}

// GetTable is not yet implemented.
func (h *Handlers) GetTable(ctx context.Context, req generated.GetTableRequestObject) (generated.GetTableResponseObject, error) {
	return nil, errors.New("not implemented")
}

// UpdateTable is not yet implemented.
func (h *Handlers) UpdateTable(ctx context.Context, req generated.UpdateTableRequestObject) (generated.UpdateTableResponseObject, error) {
	return nil, errors.New("not implemented")
}

// UploadPhoto is not yet implemented.
func (h *Handlers) UploadPhoto(ctx context.Context, req generated.UploadPhotoRequestObject) (generated.UploadPhotoResponseObject, error) {
	return nil, errors.New("not implemented")
}

// GenerateQRCode is not yet implemented.
func (h *Handlers) GenerateQRCode(ctx context.Context, req generated.GenerateQRCodeRequestObject) (generated.GenerateQRCodeResponseObject, error) {
	return nil, errors.New("not implemented")
}

// SubmitRating is not yet implemented.
func (h *Handlers) SubmitRating(ctx context.Context, req generated.SubmitRatingRequestObject) (generated.SubmitRatingResponseObject, error) {
	return nil, errors.New("not implemented")
}
