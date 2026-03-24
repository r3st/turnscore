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
	tableSvc      *service.TableService
	qrSvc         *service.QRCodeService
}

// New creates a new Handlers instance wired to the given services.
func New(
	tournamentSvc *service.TournamentService,
	tableSvc *service.TableService,
	qrSvc *service.QRCodeService,
) *Handlers {
	return &Handlers{
		tournamentSvc: tournamentSvc,
		tableSvc:      tableSvc,
		qrSvc:         qrSvc,
	}
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

const msgNotImplemented = "not implemented"

// --- Stub implementations for endpoints not yet implemented in this issue ---

// GoogleOAuthRedirect is not yet implemented.
func (h *Handlers) GoogleOAuthRedirect(ctx context.Context, req generated.GoogleOAuthRedirectRequestObject) (generated.GoogleOAuthRedirectResponseObject, error) {
	return nil, errors.New(msgNotImplemented)
}

// GoogleOAuthCallback is not yet implemented.
func (h *Handlers) GoogleOAuthCallback(ctx context.Context, req generated.GoogleOAuthCallbackRequestObject) (generated.GoogleOAuthCallbackResponseObject, error) {
	return nil, errors.New(msgNotImplemented)
}

// RaterLogin is not yet implemented.
func (h *Handlers) RaterLogin(ctx context.Context, req generated.RaterLoginRequestObject) (generated.RaterLoginResponseObject, error) {
	return nil, errors.New(msgNotImplemented)
}

// RefreshToken is not yet implemented.
func (h *Handlers) RefreshToken(ctx context.Context, req generated.RefreshTokenRequestObject) (generated.RefreshTokenResponseObject, error) {
	return nil, errors.New(msgNotImplemented)
}

// GetMe is not yet implemented.
func (h *Handlers) GetMe(ctx context.Context, req generated.GetMeRequestObject) (generated.GetMeResponseObject, error) {
	return nil, errors.New(msgNotImplemented)
}

// ListRaters is not yet implemented.
func (h *Handlers) ListRaters(ctx context.Context, req generated.ListRatersRequestObject) (generated.ListRatersResponseObject, error) {
	return nil, errors.New(msgNotImplemented)
}

// CreateRater is not yet implemented.
func (h *Handlers) CreateRater(ctx context.Context, req generated.CreateRaterRequestObject) (generated.CreateRaterResponseObject, error) {
	return nil, errors.New(msgNotImplemented)
}

// UpdateResultConfig is not yet implemented.
func (h *Handlers) UpdateResultConfig(ctx context.Context, req generated.UpdateResultConfigRequestObject) (generated.UpdateResultConfigResponseObject, error) {
	return nil, errors.New(msgNotImplemented)
}

// GetTournamentResults is not yet implemented.
func (h *Handlers) GetTournamentResults(ctx context.Context, req generated.GetTournamentResultsRequestObject) (generated.GetTournamentResultsResponseObject, error) {
	return nil, errors.New(msgNotImplemented)
}

// SubmitRating is not yet implemented.
func (h *Handlers) SubmitRating(ctx context.Context, req generated.SubmitRatingRequestObject) (generated.SubmitRatingResponseObject, error) {
	return nil, errors.New(msgNotImplemented)
}
