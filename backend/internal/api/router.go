// Package api wires together the HTTP router and registered handlers.
package api

import (
	"github.com/gin-gonic/gin"

	"github.com/r3st/turnscore/config"
	"github.com/r3st/turnscore/internal/api/generated"
	"github.com/r3st/turnscore/internal/api/handlers"
	"github.com/r3st/turnscore/internal/api/middleware"
	"github.com/r3st/turnscore/internal/service"
)

// NewRouter builds and returns a configured Gin engine.
// Authentication is enforced per-handler; OptionalAuth globally extracts the token when present.
func NewRouter(cfg *config.Config, jwtSvc *service.JWTService, h *handlers.Handlers) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.OptionalAuth(jwtSvc))

	strictHandler := generated.NewStrictHandler(h, nil)
	generated.RegisterHandlers(r.Group("/api/v1"), strictHandler)

	return r
}
