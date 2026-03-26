package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/r3st/turnscore/internal/api/generated"
	"github.com/r3st/turnscore/internal/domain"
	"github.com/r3st/turnscore/internal/service"
)

const (
	msgMissingUserID        = "missing user ID"
	msgBodyRequired         = "request body is required"
	msgTournamentNotFound   = "tournament not found"
	msgForbiddenOrganizerDelete  = "only the organizer can delete a tournament"
	msgForbiddenOrganizerAdd     = "only the organizer can add members"
	msgForbiddenOrganizerRemove  = "only the organizer can remove members"
	msgForbiddenLeave            = "Organizers cannot leave their own tournament"
	msgForbiddenUpdate           = "you do not have permission to update this tournament"
	msgTournamentOrMemberNotFound = "tournament or member not found"
	msgUserNotFound              = "user with that invite code not found"
	msgAlreadyMember             = "user is already a member of this tournament"
	msgSlugConflict              = "slug already exists"
)

// ListTournaments returns 5 upcoming and 5 past tournaments.
func (h *Handlers) ListTournaments(ctx context.Context, req generated.ListTournamentsRequestObject) (generated.ListTournamentsResponseObject, error) {
	upcoming, past, err := h.tournamentSvc.List(ctx)
	if err != nil {
		return nil, err
	}

	return generated.ListTournaments200JSONResponse{
		Upcoming: mapTournamentSummaries(upcoming),
		Past:     mapTournamentSummaries(past),
	}, nil
}

// CreateTournament creates a new tournament and registers the caller as organizer.
func (h *Handlers) CreateTournament(ctx context.Context, req generated.CreateTournamentRequestObject) (generated.CreateTournamentResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.CreateTournament401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}
	if req.Body == nil {
		return generated.CreateTournament400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgBodyRequired},
		}, nil
	}

	input := buildCreateInput(req.Body)

	t, err := h.tournamentSvc.Create(ctx, userID, input)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return generated.CreateTournament409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse{Code: "conflict", Message: msgSlugConflict},
			}, nil
		}
		return nil, err
	}

	detail, err := h.mapTournamentDetail(ctx, t)
	if err != nil {
		return nil, err
	}
	return generated.CreateTournament201JSONResponse(detail), nil
}

// GetTournament returns tournament details by slug (public).
func (h *Handlers) GetTournament(ctx context.Context, req generated.GetTournamentRequestObject) (generated.GetTournamentResponseObject, error) {
	t, err := h.tournamentSvc.GetBySlug(ctx, req.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.GetTournament404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentNotFound},
			}, nil
		}
		return nil, err
	}

	detail, err := h.mapTournamentDetail(ctx, t)
	if err != nil {
		return nil, err
	}
	return generated.GetTournament200JSONResponse(detail), nil
}

// UpdateTournament updates a tournament. Caller must be organizer or helper.
func (h *Handlers) UpdateTournament(ctx context.Context, req generated.UpdateTournamentRequestObject) (generated.UpdateTournamentResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.UpdateTournament401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}
	if req.Body == nil {
		return generated.UpdateTournament400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgBodyRequired},
		}, nil
	}

	input := buildUpdateInput(req.Body)

	t, err := h.tournamentSvc.Update(ctx, userID, req.Slug, input)
	if err != nil {
		return mapUpdateError(err)
	}

	detail, err := h.mapTournamentDetail(ctx, t)
	if err != nil {
		return nil, err
	}
	return generated.UpdateTournament200JSONResponse(detail), nil
}

// DeleteTournament deletes a tournament. Only the organizer may delete.
func (h *Handlers) DeleteTournament(ctx context.Context, req generated.DeleteTournamentRequestObject) (generated.DeleteTournamentResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.DeleteTournament401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	if err := h.tournamentSvc.Delete(ctx, userID, req.Slug); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.DeleteTournament404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.DeleteTournament403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenOrganizerDelete},
			}, nil
		}
		return nil, err
	}
	return generated.DeleteTournament204Response{}, nil
}

// AddTournamentMember adds a helper to a tournament by invite code.
func (h *Handlers) AddTournamentMember(ctx context.Context, req generated.AddTournamentMemberRequestObject) (generated.AddTournamentMemberResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.AddTournamentMember401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}
	if req.Body == nil {
		return generated.AddTournamentMember400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgBodyRequired},
		}, nil
	}

	user, err := h.tournamentSvc.AddMember(ctx, userID, req.Slug, req.Body.InviteCode)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return generated.AddTournamentMember403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenOrganizerAdd},
			}, nil
		}
		if errors.Is(err, domain.ErrNotFound) {
			return generated.AddTournamentMember404JSONResponse{Code: "not_found", Message: msgUserNotFound}, nil
		}
		if errors.Is(err, domain.ErrConflict) {
			return generated.AddTournamentMember409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse{Code: "conflict", Message: msgAlreadyMember},
			}, nil
		}
		return nil, err
	}

	return generated.AddTournamentMember200JSONResponse(generated.MemberPreview{
		Id:        user.ID,
		Name:      user.Name,
		AvatarUrl: user.AvatarURL,
	}), nil
}

// RemoveTournamentMember removes a helper from a tournament. Organizer only.
func (h *Handlers) RemoveTournamentMember(ctx context.Context, req generated.RemoveTournamentMemberRequestObject) (generated.RemoveTournamentMemberResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.RemoveTournamentMember401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	if err := h.tournamentSvc.RemoveMember(ctx, userID, req.Slug, req.UserId); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.RemoveTournamentMember404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentOrMemberNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.RemoveTournamentMember403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenOrganizerRemove},
			}, nil
		}
		return nil, err
	}
	return generated.RemoveTournamentMember204Response{}, nil
}

// LeaveTournament allows a helper to leave a tournament voluntarily.
func (h *Handlers) LeaveTournament(ctx context.Context, req generated.LeaveTournamentRequestObject) (generated.LeaveTournamentResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.LeaveTournament401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	if err := h.tournamentSvc.LeaveTournament(ctx, userID, req.Slug); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.LeaveTournament404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.LeaveTournament403JSONResponse{Code: "forbidden", Message: msgForbiddenLeave}, nil
		}
		return nil, err
	}
	return generated.LeaveTournament204Response{}, nil
}

// --- Input builders ---

func buildCreateInput(body *generated.CreateTournamentRequest) service.CreateTournamentInput {
	input := service.CreateTournamentInput{
		Name:        body.Name,
		Type:        string(body.Type),
		Description: body.Description,
		Location:    body.Location,
		TableCount:  body.TableCount,
		VotingStart: body.VotingStart,
		VotingEnd:   body.VotingEnd,
	}
	if body.EventDate != nil {
		t := body.EventDate.Time
		input.EventDate = &t
	}
	if body.Links != nil {
		links := make([]service.TournamentLink, 0, len(*body.Links))
		for _, l := range *body.Links {
			links = append(links, service.TournamentLink{URL: l.Url, Label: l.Label})
		}
		input.Links = links
	}
	if body.ActiveCriteria != nil {
		criteria := make([]string, 0, len(*body.ActiveCriteria))
		for _, c := range *body.ActiveCriteria {
			criteria = append(criteria, string(c))
		}
		input.ActiveCriteria = criteria
	}
	return input
}

func buildUpdateInput(body *generated.UpdateTournamentRequest) service.UpdateTournamentInput {
	input := service.UpdateTournamentInput{
		Name:        body.Name,
		Description: body.Description,
		Location:    body.Location,
		VotingStart: body.VotingStart,
		VotingEnd:   body.VotingEnd,
	}
	if body.Type != nil {
		s := string(*body.Type)
		input.Type = &s
	}
	if body.TableCount != nil {
		tc := *body.TableCount
		input.TableCount = &tc
	}
	if body.EventDate != nil {
		t := body.EventDate.Time
		input.EventDate = &t
	}
	if body.Status != nil {
		s := string(*body.Status)
		input.Status = &s
	}
	if body.Links != nil {
		links := make([]service.TournamentLink, 0, len(*body.Links))
		for _, l := range *body.Links {
			links = append(links, service.TournamentLink{URL: l.Url, Label: l.Label})
		}
		input.Links = &links
	}
	if body.ActiveCriteria != nil {
		criteria := make([]string, 0, len(*body.ActiveCriteria))
		for _, c := range *body.ActiveCriteria {
			criteria = append(criteria, string(c))
		}
		input.ActiveCriteria = &criteria
	}
	return input
}

func mapUpdateError(err error) (generated.UpdateTournamentResponseObject, error) {
	if errors.Is(err, domain.ErrNotFound) {
		return generated.UpdateTournament404JSONResponse{
			NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentNotFound},
		}, nil
	}
	if errors.Is(err, domain.ErrForbidden) {
		return generated.UpdateTournament403JSONResponse{
			ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenUpdate},
		}, nil
	}
	return nil, err
}

// --- Mapping helpers ---

// mapTournamentDetail converts a domain.Tournament to generated.TournamentDetail,
// fetching live table data from tableSvc.
func (h *Handlers) mapTournamentDetail(ctx context.Context, t *domain.Tournament) (generated.TournamentDetail, error) {
	links, err := parseLinks(t.Links)
	if err != nil {
		return generated.TournamentDetail{}, fmt.Errorf("parsing links: %w", err)
	}

	criteria, err := parseCriteria(t.ActiveCriteria)
	if err != nil {
		return generated.TournamentDetail{}, fmt.Errorf("parsing active_criteria: %w", err)
	}

	tables, err := h.tableSvc.ListBySlug(ctx, t.Slug)
	if err != nil {
		return generated.TournamentDetail{}, fmt.Errorf("fetching tables: %w", err)
	}

	detail := generated.TournamentDetail{
		Id:             t.ID,
		Slug:           t.Slug,
		Name:           t.Name,
		Type:           generated.TournamentType(t.Type),
		Status:         generated.TournamentStatus(t.Status),
		Location:       t.Location,
		Description:    t.Description,
		TableCount:     t.TableCount,
		Links:          links,
		ActiveCriteria: criteria,
		VotingStart:    t.VotingStart,
		VotingEnd:      t.VotingEnd,
		Tables:         mapTableSummaries(tables, h.tableSvc.Storage()),
	}

	if t.EventDate != nil {
		d := openapi_types.Date{Time: *t.EventDate}
		detail.EventDate = &d
	}

	return detail, nil
}

// mapTournamentSummaries converts a slice of domain.Tournament to []generated.TournamentSummary.
func mapTournamentSummaries(ts []domain.Tournament) []generated.TournamentSummary {
	result := make([]generated.TournamentSummary, 0, len(ts))
	for i := range ts {
		t := &ts[i]
		summary := generated.TournamentSummary{
			Id:         t.ID,
			Slug:       t.Slug,
			Name:       t.Name,
			Type:       generated.TournamentType(t.Type),
			Status:     generated.TournamentStatus(t.Status),
			Location:   t.Location,
			TableCount: t.TableCount,
		}
		if t.EventDate != nil {
			d := openapi_types.Date{Time: *t.EventDate}
			summary.EventDate = &d
		}
		result = append(result, summary)
	}
	return result
}

// parseLinks deserializes the JSONB links string to []generated.TournamentLink.
func parseLinks(raw string) ([]generated.TournamentLink, error) {
	if raw == "" || raw == "null" {
		return []generated.TournamentLink{}, nil
	}
	var links []struct {
		URL   string `json:"url"`
		Label string `json:"label"`
	}
	if err := json.Unmarshal([]byte(raw), &links); err != nil {
		return nil, err
	}
	result := make([]generated.TournamentLink, 0, len(links))
	for _, l := range links {
		result = append(result, generated.TournamentLink{Url: l.URL, Label: l.Label})
	}
	return result, nil
}

// parseCriteria deserializes the JSONB active_criteria string to []generated.CriteriaKey.
func parseCriteria(raw string) ([]generated.CriteriaKey, error) {
	if raw == "" || raw == "null" {
		return []generated.CriteriaKey{}, nil
	}
	var keys []string
	if err := json.Unmarshal([]byte(raw), &keys); err != nil {
		return nil, err
	}
	result := make([]generated.CriteriaKey, 0, len(keys))
	for _, k := range keys {
		result = append(result, generated.CriteriaKey(k))
	}
	return result, nil
}
