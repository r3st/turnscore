package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/r3st/turnscore/internal/api/middleware"
	"github.com/r3st/turnscore/internal/domain"
	"github.com/r3st/turnscore/internal/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newJWTSvc(t *testing.T) *service.JWTService {
	t.Helper()
	return service.NewJWTService(domain.JWTConfig{
		Secret:       "test-secret-that-is-at-least-32-chars!!",
		AccessExpiry: 15 * time.Minute,
	})
}

func mustUUID(t *testing.T) uuid.UUID {
	t.Helper()
	id, err := uuid.NewRandom()
	require.NoError(t, err)
	return id
}

func setupRouter(t *testing.T, jwtSvc *service.JWTService, extraMiddleware ...gin.HandlerFunc) *gin.Engine {
	t.Helper()
	r := gin.New()
	okHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"userID": c.GetString(middleware.ContextKeyUserID),
			"role":   c.GetString(middleware.ContextKeyRole),
		})
	}
	handlers := append([]gin.HandlerFunc{middleware.Auth(jwtSvc)}, extraMiddleware...)
	handlers = append(handlers, okHandler)
	r.GET("/test", handlers...)
	return r
}

func TestAuthMissingToken(t *testing.T) {
	r := setupRouter(t, newJWTSvc(t))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthInvalidToken(t *testing.T) {
	r := setupRouter(t, newJWTSvc(t))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthValidUserToken(t *testing.T) {
	jwtSvc := newJWTSvc(t)
	r := setupRouter(t, jwtSvc)

	pair, _, err := jwtSvc.IssueUserTokenPair(mustUUID(t))
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthSetsContextValues(t *testing.T) {
	jwtSvc := newJWTSvc(t)

	var capturedUserID, capturedRole string
	r := gin.New()
	r.GET("/test", middleware.Auth(jwtSvc), func(c *gin.Context) {
		capturedUserID = c.GetString(middleware.ContextKeyUserID)
		capturedRole = c.GetString(middleware.ContextKeyRole)
		c.Status(http.StatusOK)
	})

	pair, _, err := jwtSvc.IssueUserTokenPair(mustUUID(t))
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	r.ServeHTTP(w, req)

	assert.NotEmpty(t, capturedUserID)
	assert.Equal(t, "user", capturedRole)
}

func TestRequireUserAllowsUserRole(t *testing.T) {
	jwtSvc := newJWTSvc(t)
	r := setupRouter(t, jwtSvc, middleware.RequireUser())

	pair, _, err := jwtSvc.IssueUserTokenPair(mustUUID(t))
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireUserRejectsRaterRole(t *testing.T) {
	jwtSvc := newJWTSvc(t)
	r := setupRouter(t, jwtSvc, middleware.RequireUser())

	raterToken, err := jwtSvc.IssueRaterToken(mustUUID(t), mustUUID(t), time.Now().Add(time.Hour))
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+raterToken)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRaterAllowsRaterRole(t *testing.T) {
	jwtSvc := newJWTSvc(t)
	r := setupRouter(t, jwtSvc, middleware.RequireRater())

	token, err := jwtSvc.IssueRaterToken(mustUUID(t), mustUUID(t), time.Now().Add(time.Hour))
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRaterRejectsUserRole(t *testing.T) {
	jwtSvc := newJWTSvc(t)
	r := setupRouter(t, jwtSvc, middleware.RequireRater())

	pair, _, err := jwtSvc.IssueUserTokenPair(mustUUID(t))
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
