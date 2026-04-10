// Package api wires together the HTTP router and registered handlers.
package api

import (
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/r3st/turnscore/config"
	"github.com/r3st/turnscore/internal/api/generated"
	"github.com/r3st/turnscore/internal/api/handlers"
	"github.com/r3st/turnscore/internal/api/middleware"
	"github.com/r3st/turnscore/internal/devusers"
	"github.com/r3st/turnscore/internal/service"
	"github.com/r3st/turnscore/internal/static"
	"github.com/r3st/turnscore/internal/storage"
)

// NewRouter builds and returns a configured Gin engine.
// Authentication is enforced per-handler; OptionalAuth globally extracts the token when present.
func NewRouter(cfg *config.Config, jwtSvc *service.JWTService, h *handlers.Handlers, stor storage.Storage) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(cfg.Server.FrontendURL))
	r.Use(middleware.RequestLogger())
	r.Use(middleware.OptionalAuth(jwtSvc))

	// Serve uploaded files via the storage backend (local filesystem or S3 proxy).
	r.GET("/uploads/*key", uploadsHandler(stor))

	// Serve legal text files (imprint.{lang}.md, privacy.{lang}.md) from mounted directory.
	if cfg.Legal.Path != "" {
		r.Static("/legal", cfg.Legal.Path)
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
// Unknown paths serve index.html directly so the React router handles client-side navigation.
//
// NOTE: the SPA fallback must NOT use c.FileFromFS("index.html", ...) or http.FileServer
// with path "/index.html" — http.FileServer redirects /index.html → / (301), which causes
// an infinite redirect loop because every redirect comes back here.
func spaHandler(distFS fs.FS) gin.HandlerFunc {
	fileServer := http.FileServer(http.FS(distFS))
	return func(c *gin.Context) {
		reqPath := c.Request.URL.Path
		// For real static assets (not "/" and not a directory) serve via fileServer
		// so it can set correct Content-Type, cache headers, etc.
		if reqPath != "/" {
			name := reqPath[1:] // strip leading "/"
			f, err := distFS.Open(name)
			if err == nil {
				fi, statErr := f.Stat()
				f.Close()
				if statErr == nil && !fi.IsDir() {
					fileServer.ServeHTTP(c.Writer, c.Request)
					c.Abort()
					return
				}
			}
		}
		// SPA fallback: read and write index.html bytes directly.
		// Bypasses http.FileServer to avoid the /index.html → / redirect loop.
		data, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			c.Status(http.StatusNotFound)
			c.Abort()
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		c.Abort()
	}
}

// uploadsHandler streams files from the storage backend under /uploads/*key.
// For local storage this reads from the filesystem; for S3 it proxies via GetObject.
func uploadsHandler(stor storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := strings.TrimPrefix(c.Param("key"), "/")
		rc, err := stor.Open(c.Request.Context(), key)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		defer rc.Close()
		contentType := mime.TypeByExtension(filepath.Ext(key))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.DataFromReader(http.StatusOK, -1, contentType, rc, nil)
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
