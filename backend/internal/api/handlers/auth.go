package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/r3st/turnscore/internal/api/generated"
	"github.com/r3st/turnscore/internal/domain"
)

const (
	msgInvalidCredentials = "invalid nickname or code"
	msgInvalidToken       = "invalid or expired token"
	msgOAuthFailed        = "oauth flow failed"
)

// oauthRedirectResponse implements GoogleOAuthRedirectResponseObject and sets a Location header.
type oauthRedirectResponse struct{ redirectURL string }

func (r oauthRedirectResponse) VisitGoogleOAuthRedirectResponse(w http.ResponseWriter) error {
	w.Header().Set("Location", r.redirectURL)
	w.WriteHeader(http.StatusFound)
	return nil
}

// callbackRedirectResponse implements GoogleOAuthCallbackResponseObject and redirects the browser.
type callbackRedirectResponse struct{ redirectURL string }

func (r callbackRedirectResponse) VisitGoogleOAuthCallbackResponse(w http.ResponseWriter) error {
	w.Header().Set("Location", r.redirectURL)
	w.WriteHeader(http.StatusFound)
	return nil
}

// GoogleOAuthRedirect initiates the Google OAuth flow by redirecting to Google.
func (h *Handlers) GoogleOAuthRedirect(ctx context.Context, req generated.GoogleOAuthRedirectRequestObject) (generated.GoogleOAuthRedirectResponseObject, error) {
	state, err := generateOAuthState()
	if err != nil {
		return nil, err
	}
	return oauthRedirectResponse{redirectURL: h.authSvc.OAuthRedirectURL(state)}, nil
}

// GoogleOAuthCallback handles the Google OAuth callback and redirects the browser to the frontend with tokens.
func (h *Handlers) GoogleOAuthCallback(ctx context.Context, req generated.GoogleOAuthCallbackRequestObject) (generated.GoogleOAuthCallbackResponseObject, error) {
	tokenPair, err := h.authSvc.HandleCallback(ctx, req.Params.Code)
	if err != nil {
		return callbackRedirectResponse{redirectURL: h.frontendURL + "/login?error=oauth_failed"}, nil
	}
	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s&refresh=%s",
		h.frontendURL,
		url.QueryEscape(tokenPair.AccessToken),
		url.QueryEscape(tokenPair.RefreshToken),
	)
	return callbackRedirectResponse{redirectURL: redirectURL}, nil
}

// RaterLogin validates nickname + code and issues a rater JWT.
func (h *Handlers) RaterLogin(ctx context.Context, req generated.RaterLoginRequestObject) (generated.RaterLoginResponseObject, error) {
	if req.Body == nil {
		return generated.RaterLogin401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgBodyRequired},
		}, nil
	}

	tokenPair, err := h.authSvc.RaterLogin(ctx, domain.RaterLoginInput{
		TournamentSlug: req.Body.TournamentSlug,
		Nickname:       req.Body.Nickname,
		Code:           req.Body.Code,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return generated.RaterLogin404JSONResponse{
				NotFoundJSONResponse: generated.NotFoundJSONResponse{
					Code:    "not_found",
					Message: msgTournamentNotFound,
				},
			}, nil
		}
		// ErrInvalidCode and any other auth failure → 401
		return generated.RaterLogin401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{
				Code:    "unauthorized",
				Message: msgInvalidCredentials,
			},
		}, nil
	}
	return generated.RaterLogin200JSONResponse(generated.AuthTokens{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}), nil
}

// RefreshToken validates a refresh token and issues a new token pair.
func (h *Handlers) RefreshToken(ctx context.Context, req generated.RefreshTokenRequestObject) (generated.RefreshTokenResponseObject, error) {
	if req.Body == nil {
		return generated.RefreshToken401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{Code: "unauthorized", Message: msgBodyRequired},
		}, nil
	}

	tokenPair, err := h.authSvc.RefreshTokens(ctx, req.Body.RefreshToken)
	if err != nil {
		return generated.RefreshToken401JSONResponse{
			UnauthorizedJSONResponse: generated.UnauthorizedJSONResponse{
				Code:    "unauthorized",
				Message: msgInvalidToken,
			},
		}, nil
	}
	return generated.RefreshToken200JSONResponse(generated.AuthTokens{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}), nil
}

// generateOAuthState generates a random hex string for OAuth CSRF protection.
func generateOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
