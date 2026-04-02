// Package api wires together the HTTP router and registered handlers.
package api

import (
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/r3st/turnscore/config"
	"github.com/r3st/turnscore/internal/api/generated"
	"github.com/r3st/turnscore/internal/api/handlers"
	"github.com/r3st/turnscore/internal/api/middleware"
	"github.com/r3st/turnscore/internal/devusers"
	"github.com/r3st/turnscore/internal/service"
	"github.com/r3st/turnscore/internal/static"
)

// NewRouter builds and returns a configured Gin engine.
// Authentication is enforced per-handler; OptionalAuth globally extracts the token when present.
func NewRouter(cfg *config.Config, jwtSvc *service.JWTService, h *handlers.Handlers) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(cfg.Server.FrontendURL))
	r.Use(middleware.RequestLogger())
	r.Use(middleware.OptionalAuth(jwtSvc))

	// Serve uploaded files for local storage backend.
	if cfg.Storage.Backend == "local" {
		r.Static("/uploads", cfg.Storage.Local.Path)
	}

	strictHandler := generated.NewStrictHandler(h, nil)
	generated.RegisterHandlers(r.Group("/api/v1"), strictHandler)

	// Dev-only login endpoint — only registered outside of release mode.
	if cfg.Server.Mode != "release" {
		r.POST("/api/v1/auth/dev-login", devLoginHandler(jwtSvc))
	}

	// Serve the embedded React SPA for all non-API routes.
	// Static assets (JS/CSS/images) are served directly; everything else falls back to index.html.
	distFS, err := fs.Sub(static.FS, "dist")
	if err == nil {
		r.Use(spaHandler(distFS))
	}

	return r
}

// spaHandler returns a Gin middleware that serves static files from the embedded dist FS.
// Unknown paths return index.html so the React router handles client-side navigation.
func spaHandler(distFS fs.FS) gin.HandlerFunc {
	fileServer := http.FileServer(http.FS(distFS))
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		// Check whether the file exists in the embedded FS.
		f, err := distFS.Open(path[1:]) // strip leading "/"
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}
		// Unknown path → serve index.html and let the React router take over.
		c.FileFromFS("index.html", http.FS(distFS))
		c.Abort()
	}
}

// devLoginHandler returns a Gin handler for the dev-only login endpoint.
// Accepts { "password": "<name>" } and issues a token pair for the matching dev user.
func devLoginHandler(jwtSvc *service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || body.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": "bad_request", "message": "password required"})
			return
		}
		u := devusers.FindByPassword(body.Password)
		if u == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "unknown dev user"})
			return
		}
		tokenPair, _, err := jwtSvc.IssueUserTokenPair(u.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal", "message": "token error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
		})
	}
}
