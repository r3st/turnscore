---
name: backend
description: >
  Use this skill for all Go backend development tasks in TurnScore. Triggers include:
  implementing API handlers (Gin + oapi-codegen), writing GORM repository code, setting up
  Google OAuth, handling file uploads (local/S3), generating QR codes, writing Cobra/Viper
  configuration, zerolog logging, golang-migrate migrations, writing unit or integration tests,
  or any time Go code needs to be written or reviewed. Always use this skill when writing
  backend code — it defines all patterns, conventions, and module choices.
---

# Backend Skill — Go Development

## Core Dependencies (go.mod)

```go
require (
    // HTTP & API
    github.com/gin-gonic/gin v1.10.0
    github.com/oapi-codegen/runtime v1.1.1

    // Database
    gorm.io/gorm v1.25.11
    gorm.io/driver/postgres v1.5.9
    github.com/golang-migrate/migrate/v4 v4.17.1

    // Configuration & CLI
    github.com/spf13/cobra v1.8.1
    github.com/spf13/viper v1.19.0

    // Auth
    github.com/golang-jwt/jwt/v5 v5.2.1
    golang.org/x/oauth2 v0.23.0

    // Logging
    github.com/rs/zerolog v1.33.0

    // Storage
    github.com/aws/aws-sdk-go-v2 v1.30.0
    github.com/aws/aws-sdk-go-v2/service/s3 v1.63.0

    // QR & PDF
    github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
    github.com/jung-kurt/gofpdf v1.16.2

    // Utilities
    github.com/google/uuid v1.6.0

    // Testing
    github.com/stretchr/testify v1.9.0
    go.uber.org/mock v0.4.0           // gomock
)
```

---

## oapi-codegen: Implementing the Handler Interface

Code in `internal/api/generated/` is **never edited manually**.
Handlers implement the generated `StrictServerInterface`:

```go
// internal/api/handlers/handlers.go

// Compile-time check: Handlers implements the interface
var _ generated.StrictServerInterface = (*Handlers)(nil)

type Handlers struct {
    tournamentSvc service.TournamentService
    tableSvc      service.TableService
    ratingSvc     service.RatingService
    qrSvc         service.QRCodeService
}

// Example: create tournament
func (h *Handlers) CreateTournament(
    ctx context.Context,
    req generated.CreateTournamentRequestObject,
) (generated.CreateTournamentResponseObject, error) {

    userID := getUserIDFromContext(ctx)
    tournament, err := h.tournamentSvc.Create(ctx, userID, req.Body)
    if err != nil {
        return mapErrorToResponse(err), nil
    }
    return generated.CreateTournament201JSONResponse(tournament), nil
}

// Map domain errors to API responses
func mapErrorToResponse(err error) generated.CreateTournamentResponseObject {
    switch {
    case errors.Is(err, domain.ErrNotFound):
        return generated.CreateTournament404JSONResponse{Message: "not found"}
    case errors.Is(err, domain.ErrForbidden):
        return generated.CreateTournament403JSONResponse{Message: "forbidden"}
    default:
        return generated.CreateTournament500JSONResponse{Message: "internal error"}
    }
}
```

---

## Gin Router Setup

```go
// internal/api/router.go
func NewRouter(cfg *config.Config, handlers *Handlers) *gin.Engine {
    if cfg.Server.Mode == "release" {
        gin.SetMode(gin.ReleaseMode)
    }

    r := gin.New()
    r.Use(middleware.ZerologLogger())   // Structured logging (no PII!)
    r.Use(gin.Recovery())
    r.Use(middleware.CORS(cfg.Server.FrontendURL))
    r.Use(middleware.RequestID())

    // Public routes (no auth)
    public := r.Group("/api/v1")
    {
        public.GET("/auth/google", handlers.Auth.Redirect)
        public.GET("/auth/google/callback", handlers.Auth.Callback)
        public.POST("/auth/refresh", handlers.Auth.Refresh)
        public.POST("/auth/rater", handlers.Auth.RaterLogin)
        public.GET("/tournaments", handlers.Tournament.ListPublic)
        public.GET("/tournaments/:slug", handlers.Tournament.GetPublic)
        public.GET("/tournaments/:slug/tables/:num", handlers.Table.GetPublic)
    }

    // Rater routes (JWT: role=rater)
    rater := r.Group("/api/v1")
    rater.Use(middleware.Auth(cfg.Auth.JWTSecret))
    rater.Use(middleware.RequireRole("rater"))
    {
        rater.POST("/tournaments/:slug/tables/:num/ratings", handlers.Rating.Create)
        rater.GET("/tournaments/:slug/tables", handlers.Table.ListForRater)
    }

    // Organizer + Helper (JWT: role=organizer|helper)
    staff := r.Group("/api/v1")
    staff.Use(middleware.Auth(cfg.Auth.JWTSecret))
    staff.Use(middleware.RequireRole("organizer", "helper"))
    {
        staff.POST("/tournaments", handlers.Tournament.Create)
        staff.PUT("/tournaments/:slug", handlers.Tournament.Update)
        staff.POST("/tournaments/:slug/tables", handlers.Table.Create)
        staff.PUT("/tournaments/:slug/tables/:num", handlers.Table.Update)
        staff.POST("/tournaments/:slug/tables/:num/photos", handlers.Photo.Upload)
        staff.DELETE("/photos/:id", handlers.Photo.Delete)
        staff.POST("/tournaments/:slug/raters", handlers.Rater.Create)
        staff.GET("/tournaments/:slug/raters", handlers.Rater.List)
        staff.POST("/tournaments/:slug/qrcodes", handlers.QRCode.ExportPDF)
    }

    // Organizer only (JWT: role=organizer)
    organizer := r.Group("/api/v1")
    organizer.Use(middleware.Auth(cfg.Auth.JWTSecret))
    organizer.Use(middleware.RequireRole("organizer"))
    {
        organizer.DELETE("/tournaments/:slug", handlers.Tournament.Delete)
        organizer.POST("/tournaments/:slug/members", handlers.Member.Add)
        organizer.DELETE("/tournaments/:slug/members/:userId", handlers.Member.Remove)
        organizer.GET("/tournaments/:slug/results", handlers.Rating.Results)
        organizer.PUT("/tournaments/:slug/result-config", handlers.Tournament.UpdateResultConfig)
    }

    return r
}
```

---

## Gin Auth Middleware

```go
// internal/api/middleware/auth.go
func Auth(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractBearerToken(c.GetHeader("Authorization"))
        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
            return
        }
        claims, err := validateJWT(token, jwtSecret)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }
        // Only IDs in context — no name, no email
        c.Set("userID", claims.Subject)
        c.Set("role", claims.Role)
        c.Set("tournamentID", claims.TournamentID) // only set for raters
        c.Next()
    }
}

func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role := c.GetString("role")
        for _, r := range roles {
            if role == r {
                c.Next()
                return
            }
        }
        c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
    }
}
```

---

## zerolog Logging Middleware

```go
// internal/api/middleware/logging.go
func ZerologLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        reqID := c.GetString("requestID")

        c.Next()

        // ✅ Allowed: technical request metadata
        log.Info().
            Str("request_id", reqID).
            Str("method", c.Request.Method).
            Str("path", c.Request.URL.Path).   // ⚠️ no query string (may contain codes)
            Int("status", c.Writer.Status()).
            Dur("latency", time.Since(start)).
            Msg("request")

        // ❌ NEVER log:
        // c.Request.URL.RawQuery  → may contain nickname/code
        // c.GetString("userID")   → PII (internal user ID)
        // c.Request.Body          → may contain passwords/codes
    }
}
```

---

## GORM Repository Pattern

```go
// internal/repository/tournament.go
type TournamentRepository struct {
    db *gorm.DB
}

func (r *TournamentRepository) FindBySlug(ctx context.Context, slug string) (*domain.Tournament, error) {
    var t domain.Tournament
    result := r.db.WithContext(ctx).
        Preload("Tables").
        Where("slug = ?", slug).
        First(&t)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return nil, domain.ErrNotFound
    }
    return &t, result.Error
}

// For complex aggregations: raw SQL via GORM
func (r *RatingRepository) GetTableResults(ctx context.Context, tournamentID string) ([]domain.TableResult, error) {
    var results []domain.TableResult
    return results, r.db.WithContext(ctx).Raw(`
        SELECT
            t.id, t.number, t.name,
            COUNT(ra.id)                                         AS rating_count,
            AVG((ra.criteria_scores->>'balance')::float)         AS avg_balance,
            AVG((ra.criteria_scores->>'aesthetics')::float)      AS avg_aesthetics,
            AVG((ra.criteria_scores->>'terrain_density')::float) AS avg_terrain_density,
            AVG((ra.criteria_scores->>'zone_a')::float)          AS avg_zone_a,
            AVG((ra.criteria_scores->>'zone_b')::float)          AS avg_zone_b,
            AVG((ra.criteria_scores->>'overall')::float)         AS avg_overall
        FROM tables t
        LEFT JOIN ratings ra ON t.id = ra.table_id
        WHERE t.tournament_id = ?
        GROUP BY t.id, t.number, t.name
        ORDER BY avg_overall ASC NULLS LAST  -- ASC! school grades: 1 = best
    `, tournamentID).Scan(&results).Error
}
```

---

## Google OAuth Flow

```go
// internal/service/auth.go
func (s *AuthService) HandleCallback(ctx context.Context, code string) (*TokenPair, error) {
    oauthToken, err := s.oauthConfig.Exchange(ctx, code)
    if err != nil {
        return nil, domain.ErrUnauthorized
    }

    // Fetch Google UserInfo
    client := s.oauthConfig.Client(ctx, oauthToken)
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    // parse response → googleUser{Sub, Email, Name, Picture}

    // Upsert user in DB
    user, err := s.userRepo.UpsertByGoogleSub(ctx,
        googleUser.Sub, googleUser.Email, googleUser.Name, googleUser.Picture)

    // Issue JWT — no name/email in token, IDs only!
    return s.issueTokenPair(user.ID, user.Role)
}

// JWT claims: minimal, no PII
type Claims struct {
    jwt.RegisteredClaims
    Role         string `json:"role"`
    TournamentID string `json:"tid,omitempty"` // rater only
}

func (s *AuthService) issueTokenPair(userID, role string) (*TokenPair, error) {
    access := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   userID,  // ✅ ID only
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
        },
        Role: role,
        // ❌ no name, no email
    })
    // ...
}
```

---

## Storage Interface (local / S3)

```go
// internal/storage/interface.go
type Storage interface {
    Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
    Delete(ctx context.Context, key string) error
    PublicURL(key string) string
}

// Local storage (default)
type LocalStorage struct {
    basePath string
    baseURL  string
}

func (s *LocalStorage) Put(ctx context.Context, key string, r io.Reader, _ int64, _ string) error {
    fullPath := filepath.Join(s.basePath, key)
    if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
        return err
    }
    f, err := os.Create(fullPath)
    if err != nil {
        return err
    }
    defer f.Close()
    _, err = io.Copy(f, r)
    return err
}

// Factory: selected via config
func NewStorage(cfg config.StorageConfig) (Storage, error) {
    switch cfg.Backend {
    case "s3":
        return newS3Storage(cfg.S3)
    default: // "local"
        return newLocalStorage(cfg.Local)
    }
}
```

---

## Photo Upload Service

```go
// internal/service/photo.go
func (s *PhotoService) Upload(
    ctx context.Context,
    tableID uuid.UUID,
    category string, // "general" | "zone_a" | "zone_b"
    file multipart.File,
    header *multipart.FileHeader,
) (*domain.Photo, error) {
    if header.Size > 10*1024*1024 {
        return nil, domain.ErrFileTooLarge
    }
    if !isAllowedImageType(header.Header.Get("Content-Type")) {
        return nil, domain.ErrInvalidFileType
    }

    thumbnail, thumbSize, err := s.generateThumbnail(file)
    if err != nil {
        return nil, err
    }
    file.Seek(0, io.SeekStart)

    // Key pattern: tables/{tableID}/{category}/{uuid}.jpg
    key := fmt.Sprintf("tables/%s/%s/%s.jpg", tableID, category, uuid.New())
    thumbKey := fmt.Sprintf("tables/%s/%s/thumb_%s.jpg", tableID, category, uuid.New())

    if err := s.storage.Put(ctx, key, file, header.Size, "image/jpeg"); err != nil {
        return nil, fmt.Errorf("upload failed: %w", err)
    }
    if err := s.storage.Put(ctx, thumbKey, thumbnail, thumbSize, "image/jpeg"); err != nil {
        return nil, fmt.Errorf("thumbnail upload failed: %w", err)
    }

    return s.repo.CreatePhoto(ctx, domain.CreatePhotoParams{
        TableID:      tableID,
        Category:     category,
        URL:          s.storage.PublicURL(key),
        ThumbnailURL: s.storage.PublicURL(thumbKey),
    })
}
```

---

## QR Code PDF Export

```go
// internal/service/qrcode.go
func (s *QRCodeService) ExportPDF(ctx context.Context, tournamentSlug string, tables []domain.Table) ([]byte, error) {
    pdf := gofpdf.New("P", "mm", "A4", "")
    const (
        cardsPerRow  = 3
        cardsPerPage = 6
        cardW        = 63.0
        cardH        = 90.0
        marginLeft   = 10.5
        marginTop    = 10.0
    )

    for i, table := range tables {
        if i%cardsPerPage == 0 {
            pdf.AddPage()
        }
        row := (i % cardsPerPage) / cardsPerRow
        col := (i % cardsPerPage) % cardsPerRow
        x := marginLeft + float64(col)*cardW
        y := marginTop + float64(row)*cardH

        pdf.Rect(x, y, cardW-2, cardH-2, "D")

        qrURL := fmt.Sprintf("%s/rate/%s/%d", s.baseURL, tournamentSlug, table.Number)
        qrPNG, err := qrcode.Encode(qrURL, qrcode.High, 200)
        if err != nil {
            return nil, err
        }
        tmpFile, _ := os.CreateTemp("", "qr-*.png")
        tmpFile.Write(qrPNG)
        tmpFile.Close()
        defer os.Remove(tmpFile.Name())

        pdf.ImageOptions(tmpFile.Name(), x+4, y+4, cardW-10, cardW-10, false,
            gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

        pdf.SetFont("Arial", "B", 11)
        pdf.Text(x+2, y+cardW-2, fmt.Sprintf("Table %d", table.Number))
        if table.Name != "" {
            pdf.SetFont("Arial", "", 9)
            pdf.Text(x+2, y+cardW+4, table.Name)
        }
    }

    var buf bytes.Buffer
    if err := pdf.Output(&buf); err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}
```

---

## Domain Errors

```go
// internal/domain/errors.go
var (
    ErrNotFound        = errors.New("not found")
    ErrUnauthorized    = errors.New("unauthorized")
    ErrForbidden       = errors.New("forbidden")
    ErrDuplicateRating = errors.New("already rated this table")
    ErrVotingNotActive = errors.New("voting period is not active")
    ErrInvalidCode     = errors.New("invalid nickname or code")
    ErrFileTooLarge    = errors.New("file exceeds 10MB limit")
    ErrInvalidFileType = errors.New("only image files allowed")
)
```

---

## Migration Convention

```
db/migrations/
  000001_create_users.up.sql        ← no password_hash (Google OAuth only)
  000001_create_users.down.sql
  000002_create_tournaments.up.sql
  000002_create_tournaments.down.sql
  ...  (one file pair per table, numbered sequentially)
```

---

## Test Patterns

### Unit Test (service with mocked repository)

```go
// internal/service/tournament_test.go
func TestCreateTournament_SlugCollision(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockTournamentRepository(ctrl)
    mockRepo.EXPECT().
        ExistsBySlug(gomock.Any(), "essen-open-2025").
        Return(true, nil)
    mockRepo.EXPECT().
        ExistsBySlug(gomock.Any(), "essen-open-2025-2").
        Return(false, nil)
    mockRepo.EXPECT().
        Create(gomock.Any(), gomock.Any()).
        Return(&domain.Tournament{Slug: "essen-open-2025-2"}, nil)

    svc := NewTournamentService(mockRepo)
    result, err := svc.Create(context.Background(), "userID", CreateTournamentInput{
        Name: "Essen Open 2025",
    })

    assert.NoError(t, err)
    assert.Equal(t, "essen-open-2025-2", result.Slug)
}
```

### Integration Test (repository with test DB)

```go
// pkg/testhelper/db.go
func NewTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    dsn := os.Getenv("TEST_DATABASE_URL")
    db, err := gorm.Open(postgres.Open(dsn))
    require.NoError(t, err)

    // Each test in its own transaction → rollback at end
    tx := db.Begin()
    t.Cleanup(func() { tx.Rollback() })
    return tx
}

// internal/repository/rating_test.go
func TestCreateRating_PreventsDuplicate(t *testing.T) {
    db := testhelper.NewTestDB(t)
    repo := NewRatingRepository(db)
    // create fixtures, first rating → OK, second → ErrDuplicateRating
}
```

### API Handler Test

```go
// internal/api/handlers/tournament_test.go
func TestCreateTournament_Returns201(t *testing.T) {
    router := setupTestRouter(t) // with mocked services
    body := `{"name":"Test Tournament","type":"scifi","tableCount":8}`

    w := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodPost, "/api/v1/tournaments", strings.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+testJWT(t, "organizer"))
    req.Header.Set("Content-Type", "application/json")

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
    var resp map[string]any
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.Equal(t, "Test Tournament", resp["name"])
}
```

### Check Coverage

```bash
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | tail -1   # total coverage
# Target: >80% overall, >90% for service/
```

---

## Conventions & Rules

- **Never** use `gorm.AutoMigrate` in production — always golang-migrate SQL files
- **Never** log PII (email, name, nickname) in zerolog
- **Never** expose `rater_id` or nickname in API responses visible to organizers
- Repository interface in `domain/`, implementation in `repository/` → easily mockable
- All exported functions have a godoc comment
- `ctx context.Context` is always the first parameter for DB/service calls
- Always wrap errors with `fmt.Errorf("operation: %w", err)`
- Run `make licenses` and commit the update whenever a new dependency is added
