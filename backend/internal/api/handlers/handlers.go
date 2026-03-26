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
	raterSvc      *service.RaterService
	userSvc       *service.UserService
	authSvc       *service.AuthService
	frontendURL   string // used to build OAuth callback redirect URLs
}

// New creates a new Handlers instance wired to the given services.
func New(
	tournamentSvc *service.TournamentService,
	tableSvc *service.TableService,
	qrSvc *service.QRCodeService,
	raterSvc *service.RaterService,
	userSvc *service.UserService,
	authSvc *service.AuthService,
	frontendURL string,
) *Handlers {
	return &Handlers{
		tournamentSvc: tournamentSvc,
		tableSvc:      tableSvc,
		qrSvc:         qrSvc,
		raterSvc:      raterSvc,
		userSvc:       userSvc,
		authSvc:       authSvc,
		frontendURL:   frontendURL,
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

// getTournamentID extracts and parses the tournamentID from the request context (set by RaterAuth middleware).
func getTournamentID(ctx context.Context) (uuid.UUID, bool) {
	raw, ok := ctx.Value(middleware.ContextKeyTournamentID).(string)
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
