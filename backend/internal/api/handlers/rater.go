package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"github.com/r3st/turnscore/internal/api/generated"
	"github.com/r3st/turnscore/internal/domain"
	"github.com/r3st/turnscore/internal/service"
)

const (
	msgMissingRaterID      = "missing rater ID"
	msgMissingTournamentID = "missing tournament ID"
	msgRaterConflict       = "nickname already taken"
	msgDuplicateRating     = "you have already rated this table"
	msgVotingNotActive     = "voting is not currently active for this tournament"
	msgForbiddenRater      = "you are not allowed to submit ratings for this tournament"
	msgForbiddenResults    = "only organizers and helpers can view results"
	msgForbiddenConfig     = "only organizers can update result config"
	msgUserProfileNotFound = "user profile not found"
)

// ListRaters returns all raters for a tournament (organizer or helper only).
func (h *Handlers) ListRaters(ctx context.Context, req generated.ListRatersRequestObject) (generated.ListRatersResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.ListRaters401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	raters, err := h.raterSvc.ListRaters(ctx, userID, req.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.ListRaters404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.ListRaters403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenResults},
			}, nil
		}
		return nil, err
	}

	result := make(generated.ListRaters200JSONResponse, 0, len(raters))
	for _, r := range raters {
		result = append(result, generated.Rater{
			Id:       r.ID,
			Nickname: r.Nickname,
			Code:     r.Code,
		})
	}
	return result, nil
}

// ExportRatersPdf returns a PDF listing all raters with their access codes (organizer or helper only).
func (h *Handlers) ExportRatersPdf(ctx context.Context, req generated.ExportRatersPdfRequestObject) (generated.ExportRatersPdfResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.ExportRatersPdf401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	pdfBytes, err := h.raterSvc.ExportRatersPDF(ctx, userID, req.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.ExportRatersPdf404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.ExportRatersPdf403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenResults},
			}, nil
		}
		return nil, err
	}

	return generated.ExportRatersPdf200ApplicationpdfResponse{
		Body:          bytes.NewReader(pdfBytes),
		ContentLength: int64(len(pdfBytes)),
	}, nil
}

// CreateRater registers a new rater for a tournament (organizer or helper only).
func (h *Handlers) CreateRater(ctx context.Context, req generated.CreateRaterRequestObject) (generated.CreateRaterResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.CreateRater401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}
	if req.Body == nil {
		return generated.CreateRater400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgBodyRequired},
		}, nil
	}

	rater, rawCode, err := h.raterSvc.CreateRater(ctx, userID, req.Slug, req.Body.Nickname, req.Body.Code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.CreateRater404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.CreateRater403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenResults},
			}, nil
		}
		if errors.Is(err, domain.ErrConflict) {
			return generated.CreateRater409JSONResponse{
				ConflictJSONResponse: generated.ConflictJSONResponse{Code: "conflict", Message: msgRaterConflict},
			}, nil
		}
		return nil, err
	}

	return generated.CreateRater201JSONResponse{
		Id:       rater.ID,
		Nickname: rater.Nickname,
		Code:     rawCode,
	}, nil
}

// SubmitRating submits a rating for a table (rater only).
func (h *Handlers) SubmitRating(ctx context.Context, req generated.SubmitRatingRequestObject) (generated.SubmitRatingResponseObject, error) {
	raterID, ok := getUserID(ctx)
	if !ok {
		return generated.SubmitRating401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingRaterID},
		}, nil
	}
	tournamentID, ok := getTournamentID(ctx)
	if !ok {
		return generated.SubmitRating401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingTournamentID},
		}, nil
	}
	if req.Body == nil {
		return generated.SubmitRating400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgBodyRequired},
		}, nil
	}

	err := h.raterSvc.SubmitRating(ctx, service.SubmitRatingInput{
		RaterID:      raterID,
		TournamentID: tournamentID,
		Slug:         req.Slug,
		TableNumber:  req.Number,
		Scores:       map[string]int(req.Body.CriteriaScores),
		PlayedZone:   string(req.Body.PlayedZone),
		Comment:      req.Body.Comment,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.SubmitRating404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.SubmitRating403JSONResponse{Code: "forbidden", Message: msgForbiddenRater}, nil
		}
		if errors.Is(err, domain.ErrVotingNotActive) {
			return generated.SubmitRating403JSONResponse{Code: "forbidden", Message: msgVotingNotActive}, nil
		}
		if errors.Is(err, domain.ErrDuplicateRating) {
			return generated.SubmitRating409JSONResponse{Code: "conflict", Message: msgDuplicateRating}, nil
		}
		return nil, err
	}
	return generated.SubmitRating201Response{}, nil
}

// SubmitEventRating submits a tournament-level event rating (optional criteria + comment).
func (h *Handlers) SubmitEventRating(ctx context.Context, req generated.SubmitEventRatingRequestObject) (generated.SubmitEventRatingResponseObject, error) {
	raterID, ok := getUserID(ctx)
	if !ok {
		return generated.SubmitEventRating401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingRaterID},
		}, nil
	}
	tournamentID, ok := getTournamentID(ctx)
	if !ok {
		return generated.SubmitEventRating401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingTournamentID},
		}, nil
	}
	if req.Body == nil {
		return generated.SubmitEventRating400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgBodyRequired},
		}, nil
	}

	scores := map[string]int(req.Body.CriteriaScores)

	err := h.raterSvc.SubmitEventRating(ctx, raterID, tournamentID, req.Slug, scores, req.Body.Comment)
	if err != nil {
		if errors.Is(err, domain.ErrVotingNotActive) {
			return generated.SubmitEventRating403JSONResponse{Code: "forbidden", Message: msgVotingNotActive}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.SubmitEventRating403JSONResponse{Code: "forbidden", Message: msgForbiddenRater}, nil
		}
		if errors.Is(err, domain.ErrDuplicateRating) {
			return generated.SubmitEventRating409JSONResponse{Code: "conflict", Message: "you have already submitted an event rating for this tournament"}, nil
		}
		return nil, err
	}
	return generated.SubmitEventRating201Response{}, nil
}

// GetEventRatingStatus returns whether the rater has already submitted an event rating.
func (h *Handlers) GetEventRatingStatus(ctx context.Context, req generated.GetEventRatingStatusRequestObject) (generated.GetEventRatingStatusResponseObject, error) {
	raterID, ok := getUserID(ctx)
	if !ok {
		return generated.GetEventRatingStatus401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingRaterID},
		}, nil
	}
	tournamentID, ok := getTournamentID(ctx)
	if !ok {
		return generated.GetEventRatingStatus401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingTournamentID},
		}, nil
	}

	submitted, err := h.raterSvc.HasEventRating(ctx, raterID, tournamentID, req.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return generated.GetEventRatingStatus401JSONResponse{
				UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingTournamentID},
			}, nil
		}
		return nil, err
	}
	return generated.GetEventRatingStatus200JSONResponse{Submitted: submitted}, nil
}

// GetMyRatings returns the table numbers already rated by the current rater in this tournament.
func (h *Handlers) GetMyRatings(ctx context.Context, req generated.GetMyRatingsRequestObject) (generated.GetMyRatingsResponseObject, error) {
	raterID, ok := getUserID(ctx)
	if !ok {
		return generated.GetMyRatings401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingRaterID},
		}, nil
	}
	tournamentID, ok := getTournamentID(ctx)
	if !ok {
		return generated.GetMyRatings401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingTournamentID},
		}, nil
	}

	numbers, err := h.raterSvc.GetRatedTableNumbers(ctx, raterID, tournamentID, req.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return generated.GetMyRatings401JSONResponse{
				UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingTournamentID},
			}, nil
		}
		return nil, err
	}
	return generated.GetMyRatings200JSONResponse{RatedTableNumbers: numbers}, nil
}

// GetTournamentResults returns tournament results (organizer or helper only).
func (h *Handlers) GetTournamentResults(ctx context.Context, req generated.GetTournamentResultsRequestObject) (generated.GetTournamentResultsResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.GetTournamentResults401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	tournament, results, commentsMap, err := h.raterSvc.GetResults(ctx, userID, req.Slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.GetTournamentResults404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.GetTournamentResults403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenResults},
			}, nil
		}
		return nil, err
	}

	ranking := buildRanking(results, commentsMap)

	// Parse active_criteria from tournament JSON field.
	var activeCriteria []string
	if err := json.Unmarshal([]byte(tournament.ActiveCriteria), &activeCriteria); err != nil {
		activeCriteria = []string{}
	}
	apiCriteria := make([]generated.CriteriaKey, 0, len(activeCriteria))
	for _, k := range activeCriteria {
		apiCriteria = append(apiCriteria, generated.CriteriaKey(k))
	}

	// Parse show_comments from result config.
	var resultCfg struct {
		ShowComments bool `json:"show_comments"`
	}
	_ = json.Unmarshal([]byte(tournament.ResultConfig), &resultCfg)

	return generated.GetTournamentResults200JSONResponse(generated.TournamentResults{
		TournamentName: tournament.Name,
		ActiveCriteria: apiCriteria,
		ShowComments:   resultCfg.ShowComments,
		Ranking:        ranking,
	}), nil
}

// buildRanking converts domain results into generated TableResult slice.
// commentsMap may be nil when show_comments is disabled.
func buildRanking(results []domain.RatingResult, commentsMap map[uuid.UUID][]string) []generated.TableResult {
	ranking := make([]generated.TableResult, 0, len(results))
	for _, r := range results {
		avgs := make(map[string]float32, len(r.CriteriaAverages))
		for k, v := range r.CriteriaAverages {
			avgs[k] = float32(v)
		}

		tr := generated.TableResult{
			TableNumber:  r.TableNumber,
			TableName:    r.TableName,
			RatingCount:  int(r.RatingCount),
			AverageScores: avgs,
			ZoneDistribution: struct {
				None  int `json:"none"`
				ZoneA int `json:"zone_a"`
				ZoneB int `json:"zone_b"`
			}{
				None:  int(r.ZoneNone),
				ZoneA: int(r.ZoneA),
				ZoneB: int(r.ZoneB),
			},
		}

		if commentsMap != nil {
			if comments, exists := commentsMap[r.TableID]; exists && len(comments) > 0 {
				commentItems := make([]struct {
					Comment string `json:"comment"`
				}, 0, len(comments))
				for _, c := range comments {
					commentItems = append(commentItems, struct {
						Comment string `json:"comment"`
					}{Comment: c})
				}
				tr.Comments = &commentItems
			}
		}

		ranking = append(ranking, tr)
	}
	return ranking
}

// UpdateResultConfig updates the result visibility config (organizer only).
func (h *Handlers) UpdateResultConfig(ctx context.Context, req generated.UpdateResultConfigRequestObject) (generated.UpdateResultConfigResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.UpdateResultConfig401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}
	if req.Body == nil {
		return generated.UpdateResultConfig400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgBodyRequired},
		}, nil
	}

	visibleCriteria := make([]string, 0, len(req.Body.VisibleCommentCriteria))
	for _, c := range req.Body.VisibleCommentCriteria {
		visibleCriteria = append(visibleCriteria, string(c))
	}

	_, err := h.raterSvc.UpdateResultConfig(ctx, userID, req.Slug, req.Body.ShowComments, visibleCriteria)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.UpdateResultConfig404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{Code: "not_found", Message: msgTournamentNotFound},
			}, nil
		}
		if errors.Is(err, domain.ErrForbidden) {
			return generated.UpdateResultConfig403JSONResponse{
				ForbiddenJSONResponse: generated.ForbiddenJSONResponse{Code: "forbidden", Message: msgForbiddenConfig},
			}, nil
		}
		return nil, err
	}

	return generated.UpdateResultConfig200JSONResponse(generated.ResultConfig{
		ShowComments:           req.Body.ShowComments,
		VisibleCommentCriteria: req.Body.VisibleCommentCriteria,
	}), nil
}

// GetMe returns the authenticated user's profile with tournament memberships.
func (h *Handlers) GetMe(ctx context.Context, req generated.GetMeRequestObject) (generated.GetMeResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.GetMe401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}

	user, memberships, err := h.userSvc.GetProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.GetMe401JSONResponse{
				UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgUserProfileNotFound},
			}, nil
		}
		return nil, err
	}

	apiMemberships := make([]generated.TournamentMembership, 0, len(memberships))
	for _, m := range memberships {
		apiMemberships = append(apiMemberships, generated.TournamentMembership{
			TournamentId:   m.TournamentID,
			TournamentName: m.TournamentName,
			TournamentSlug: m.TournamentSlug,
			Role:           generated.TournamentMembershipRole(m.Role),
		})
	}

	return generated.GetMe200JSONResponse(generated.UserProfile{
		Id:               user.ID,
		Name:             user.Name,
		AvatarUrl:        user.AvatarURL,
		HelperInviteCode: user.HelperInviteCode,
		Memberships:      apiMemberships,
		Theme:            generated.UserProfileTheme(user.Theme),
		ColorMode:        generated.UserProfileColorMode(user.ColorMode),
	}), nil
}

// PatchMe updates the authenticated user's theme and color_mode preferences.
func (h *Handlers) PatchMe(ctx context.Context, req generated.PatchMeRequestObject) (generated.PatchMeResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.PatchMe401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}
	if req.Body == nil {
		return generated.PatchMe400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgBodyRequired},
		}, nil
	}

	input := domain.UpdatePreferencesInput{}
	if req.Body.Theme != nil {
		t := string(*req.Body.Theme)
		input.Theme = &t
	}
	if req.Body.ColorMode != nil {
		c := string(*req.Body.ColorMode)
		input.ColorMode = &c
	}

	if _, err := h.userSvc.UpdatePreferences(ctx, userID, input); err != nil {
		return nil, err
	}

	user, memberships, err := h.userSvc.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	apiMemberships := make([]generated.TournamentMembership, 0, len(memberships))
	for _, m := range memberships {
		apiMemberships = append(apiMemberships, generated.TournamentMembership{
			TournamentId:   m.TournamentID,
			TournamentName: m.TournamentName,
			TournamentSlug: m.TournamentSlug,
			Role:           generated.TournamentMembershipRole(m.Role),
		})
	}

	return generated.PatchMe200JSONResponse(generated.UserProfile{
		Id:               user.ID,
		Name:             user.Name,
		AvatarUrl:        user.AvatarURL,
		HelperInviteCode: user.HelperInviteCode,
		Memberships:      apiMemberships,
		Theme:            generated.UserProfileTheme(user.Theme),
		ColorMode:        generated.UserProfileColorMode(user.ColorMode),
	}), nil
}

// DeleteMe permanently deletes the authenticated user's account and all owned data.
func (h *Handlers) DeleteMe(ctx context.Context, req generated.DeleteMeRequestObject) (generated.DeleteMeResponseObject, error) {
	userID, ok := getUserID(ctx)
	if !ok {
		return generated.DeleteMe401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgMissingUserID},
		}, nil
	}
	if req.Body == nil {
		return generated.DeleteMe400JSONResponse{
			BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "bad_request", Message: msgBodyRequired},
		}, nil
	}

	if err := h.userSvc.DeleteAccount(ctx, userID, string(req.Body.ConfirmEmail)); err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return generated.DeleteMe400JSONResponse{
				BadRequestJSONResponse: generated.BadRequestJSONResponse{Code: "email_mismatch", Message: "email does not match"},
			}, nil
		}
		return nil, err
	}
	return generated.DeleteMe204Response{}, nil
}
