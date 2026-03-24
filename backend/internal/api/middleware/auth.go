// Package middleware provides Gin middleware for TurnScore.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/r3st/turnscore/internal/service"
)

const (
	ContextKeyUserID       = "userID"
	ContextKeyRole         = "role"
	ContextKeyTournamentID = "tournamentID" // raters only
)

// Auth validates the Bearer JWT and sets userID + role in the Gin context.
// Does not restrict by role — use RequireUser or RequireRater for that.
func Auth(jwtSvc *service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := extractBearerToken(c.GetHeader("Authorization"))
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "missing_token",
				"message": "Authorization header with Bearer token required",
			})
			return
		}

		claims, err := jwtSvc.ValidateClaims(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "invalid_token",
				"message": "Token is invalid or expired",
			})
			return
		}

		// Store only IDs and role — no name, no email
		c.Set(ContextKeyUserID, claims.Subject)
		c.Set(ContextKeyRole, claims.Role)
		if claims.TournamentID != "" {
			c.Set(ContextKeyTournamentID, claims.TournamentID)
		}
		c.Next()
	}
}

// RequireUser rejects rater tokens — allows only role="user" (organizers/helpers).
func RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString(ContextKeyRole) != "user" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    "forbidden",
				"message": "This endpoint requires an organizer or helper account",
			})
			return
		}
		c.Next()
	}
}

// RequireRater rejects non-rater tokens.
func RequireRater() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString(ContextKeyRole) != "rater" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    "forbidden",
				"message": "This endpoint requires a rater token",
			})
			return
		}
		c.Next()
	}
}

// OptionalAuth parses the Bearer token if present but does not reject missing or invalid tokens.
// It sets userID, role (and optionally tournamentID) in the Gin context when a valid token is found.
func OptionalAuth(jwtSvc *service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := extractBearerToken(c.GetHeader("Authorization"))
		if raw != "" {
			if claims, err := jwtSvc.ValidateClaims(raw); err == nil {
				c.Set(ContextKeyUserID, claims.Subject)
				c.Set(ContextKeyRole, claims.Role)
				if claims.TournamentID != "" {
					c.Set(ContextKeyTournamentID, claims.TournamentID)
				}
			}
		}
		c.Next()
	}
}

// extractBearerToken strips the "Bearer " prefix from the Authorization header.
func extractBearerToken(header string) string {
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(header, "Bearer ")
}
