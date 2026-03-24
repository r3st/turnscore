---
name: architecture
description: >
  Use this skill whenever architectural decisions need to be made or documented for TurnScore.
  Triggers include: designing or discussing API endpoints, database schema design,
  folder/package structure, Docker setup, authentication flow, file storage strategy,
  OpenAPI code generation, configuration management, test strategy, license management,
  accessibility, or any time the user asks about system structure, database schema, API design,
  architecture, configuration, or testing strategy. Always consult this skill before creating
  new packages, API routes, or database tables.
---

# Architecture Skill — TurnScore

## Stack Overview

| Layer | Technology | Reason |
|---|---|---|
| Backend API | **Go 1.24+** with **Gin** | Performant, great oapi-codegen support |
| Database | **PostgreSQL 18** | Proven, JSONB for flexible criteria |
| ORM / Migrations | **GORM** + **golang-migrate** | GORM for queries, migrate for schema versioning |
| Configuration | **Cobra** (CLI) + **Viper** (config) | Hierarchical config: file → ENV → CLI flag |
| Logging | **zerolog** | Structured JSON logging, zero-allocation |
| Auth | **Google OAuth 2.0** + **JWT** | No custom password flow needed |
| File Storage | **Local storage** (default) / **S3-compatible** (optional) | Interface-based, switchable via config |
| QR Codes | **go-qrcode** + **gofpdf** | PNG + PDF export |
| API Spec | **OpenAPI 3.1** → **oapi-codegen** | Contract-first, type-safe handler interfaces |
| Frontend | **React 18** + **TypeScript** + **Vite** | |
| UI Library | **shadcn/ui** + **Tailwind CSS** | Accessibility built-in (Radix UI primitives) |
| API Client | **TanStack Query** + **openapi-typescript** | Types generated directly from OpenAPI spec |
| i18n | **react-i18next** | EN/DE, no text hardcoding |
| Backend Testing | **Go testing** + **testify** + **gomock** | Unit + integration tests |
| Frontend Testing | **Vitest** + **React Testing Library** + **Playwright** | Unit + E2E |
| Containerization | **Docker** + **Docker Compose** | |

---

## Monorepo Structure

```
turnscore/
├── docker-compose.yml
├── docker-compose.dev.yml
├── .env.example
├── Makefile
│
├── api/
│   └── openapi.yaml            # OpenAPI 3.1 spec (single source of truth)
│
├── backend/
│   ├── Dockerfile
│   ├── go.mod / go.sum
│   ├── main.go
│   ├── cmd/
│   │   └── server/
│   │       └── root.go         # Cobra root command
│   ├── config/
│   │   ├── config.go           # Config struct + Viper bindings
│   │   ├── config.yaml         # Default configuration
│   │   └── config.dev.yaml     # Dev overrides (gitignored)
│   ├── internal/
│   │   ├── api/
│   │   │   ├── generated/      # oapi-codegen output (do not edit)
│   │   │   ├── handlers/       # Handler implementations
│   │   │   └── middleware/
│   │   │       ├── auth.go
│   │   │       ├── cors.go
│   │   │       └── logging.go  # zerolog middleware (no PII!)
│   │   ├── domain/             # Interfaces + domain types + errors
│   │   ├── service/            # Business logic
│   │   ├── repository/         # GORM repositories
│   │   └── storage/
│   │       ├── interface.go    # Storage interface
│   │       ├── local.go        # Local filesystem storage
│   │       └── s3.go           # S3-compatible storage (AWS/MinIO)
│   ├── db/
│   │   └── migrations/         # golang-migrate SQL files
│   └── pkg/
│       ├── logger/             # zerolog setup
│       └── testhelper/         # Shared test utilities
│
├── frontend/
│   ├── Dockerfile
│   ├── package.json
│   ├── vite.config.ts
│   ├── playwright.config.ts
│   ├── src/
│   │   ├── api/
│   │   │   ├── generated/      # openapi-typescript output (do not edit)
│   │   │   └── client.ts       # Axios + TanStack Query setup
│   │   ├── components/
│   │   ├── pages/
│   │   ├── hooks/
│   │   ├── stores/
│   │   ├── i18n/
│   │   │   ├── en.json
│   │   │   └── de.json
│   │   └── themes/
│   │       ├── fantasy.css
│   │       └── scifi.css
│   └── e2e/                    # Playwright E2E tests
│
└── docs/
    └── decisions/              # Architecture Decision Records (ADRs)
```

---

## Configuration (Cobra + Viper)

### Priority (lowest → highest)
```
config.yaml (defaults) → config.dev.yaml → ENV variables → CLI flags
```

### config.yaml
```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"          # debug | release
  frontend_url: "http://localhost:3000"

database:
  host: "localhost"
  port: 5432
  name: "turnscore"
  user: "app"
  password: ""             # Set via ENV!
  sslmode: "disable"

auth:
  google_client_id: ""     # Set via ENV!
  google_client_secret: "" # Set via ENV!
  jwt_secret: ""           # Set via ENV!
  jwt_expiry: "15m"
  refresh_expiry: "168h"   # 7 days

storage:
  backend: "local"         # local | s3
  local:
    path: "./uploads"
    base_url: "http://localhost:8080"
  s3:
    endpoint: ""
    bucket: ""
    region: "eu-central-1"
    access_key: ""
    secret_key: ""

logging:
  level: "info"            # debug | info | warn | error
  format: "json"           # json | pretty (pretty for dev only)
```

### ENV Variables (override config)
```bash
# Pattern: APP_ + key path with _ separator (Viper AutomaticEnv)
APP_SERVER_PORT=9090
APP_DATABASE_HOST=my-db-host
APP_DATABASE_PASSWORD=secret
APP_AUTH_JWT_SECRET=my-jwt-secret
APP_STORAGE_BACKEND=s3
APP_LOGGING_LEVEL=debug
```

### CLI Flags (override everything)
```bash
./server --port 9090 --log-level debug --storage-backend local
./server --config /path/to/custom-config.yaml
```

---

## OpenAPI + Code Generation

### Workflow
```
api/openapi.yaml
    ↓ oapi-codegen
backend/internal/api/generated/   (Go server interfaces + types)
    ↓ openapi-typescript
frontend/src/api/generated/        (TypeScript types)
```

### Makefile Targets
```makefile
generate:
	oapi-codegen -config backend/oapi-codegen.yaml api/openapi.yaml
	cd frontend && npx openapi-typescript ../api/openapi.yaml -o src/api/generated/api.ts

test-backend:
	go test ./... -coverprofile=coverage.out -covermode=atomic

test-frontend:
	cd frontend && npx vitest run --coverage

test-e2e:
	cd frontend && npx playwright test

test: test-backend test-frontend test-e2e
```

### oapi-codegen Config
```yaml
# backend/oapi-codegen.yaml
package: generated
generate:
  gin-server: true
  models: true
  strict-server: true     # Forces implementation of all handlers
output: internal/api/generated/api.gen.go
```

---

## Database Schema (GORM + golang-migrate)

Schema managed via golang-migrate SQL files — never via `gorm.AutoMigrate` in production.

```sql
-- 000001_create_users.up.sql
CREATE TABLE users (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    google_sub         VARCHAR(255) UNIQUE NOT NULL,  -- Google OAuth subject ID
    email              VARCHAR(255) UNIQUE NOT NULL,
    name               VARCHAR(200) NOT NULL,
    avatar_url         VARCHAR(500),
    helper_invite_code VARCHAR(20) UNIQUE NOT NULL,   -- permanent, shareable; organizers use this to add helpers
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    updated_at         TIMESTAMPTZ DEFAULT NOW()
    -- NO password_hash — login via Google only
    -- NO global role — role is per tournament in tournament_members
);

-- 000002_create_tournaments.up.sql
CREATE TABLE tournaments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            VARCHAR(100) UNIQUE NOT NULL,
    name            VARCHAR(200) NOT NULL,
    type            VARCHAR(20) NOT NULL CHECK (type IN ('fantasy', 'scifi')),
    description     TEXT,
    links           JSONB DEFAULT '[]',        -- [{url, label}]
    location        VARCHAR(200),
    event_date      DATE,
    organizer_id    UUID NOT NULL REFERENCES users(id),
    table_count     INT NOT NULL CHECK (table_count > 0),
    status          VARCHAR(20) DEFAULT 'draft'
                    CHECK (status IN ('draft','active','voting','archived')),
    voting_start    TIMESTAMPTZ,
    voting_end      TIMESTAMPTZ,
    active_criteria JSONB DEFAULT '["balance","aesthetics","terrain_density",
                                    "labeling","overall","zone_a","zone_b"]',
    result_config   JSONB DEFAULT '{"show_comments":false,
                                    "visible_comment_criteria":[]}',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 000003_create_tournament_members.up.sql
CREATE TABLE tournament_members (
    tournament_id UUID REFERENCES tournaments(id) ON DELETE CASCADE,
    user_id       UUID REFERENCES users(id) ON DELETE CASCADE,
    role          VARCHAR(20) NOT NULL CHECK (role IN ('organizer', 'helper')),
    joined_at     TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (tournament_id, user_id)
    -- organizer_id in tournaments table is the creator; this table tracks all members incl. helpers
);

-- 000004_create_tables.up.sql
CREATE TABLE tables (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    number        INT NOT NULL CHECK (number > 0),
    name          VARCHAR(200),
    description   TEXT,
    qr_code_path  VARCHAR(500),
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (tournament_id, number)
);

-- 000005_create_photos.up.sql
CREATE TABLE photos (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id      UUID NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
    url           VARCHAR(500) NOT NULL,
    thumbnail_url VARCHAR(500),
    category      VARCHAR(20) NOT NULL DEFAULT 'general'
                  CHECK (category IN ('general', 'zone_a', 'zone_b')),
    uploaded_by   UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

-- 000006_create_raters.up.sql
CREATE TABLE raters (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    nickname      VARCHAR(100) NOT NULL,
    code          VARCHAR(10) NOT NULL,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (tournament_id, nickname),
    UNIQUE (tournament_id, code)
    -- No name, no email → GDPR compliant
);

-- 000007_create_ratings.up.sql
CREATE TABLE ratings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id        UUID NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
    rater_id        UUID NOT NULL REFERENCES raters(id) ON DELETE CASCADE,
    criteria_scores JSONB NOT NULL,  -- {"balance":2,"aesthetics":1,...} scale 1-6
    played_zone     VARCHAR(20) DEFAULT 'none'
                    CHECK (played_zone IN ('none', 'zone_a', 'zone_b')),
    comment         TEXT CHECK (char_length(comment) <= 1000),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (table_id, rater_id)     -- Duplicate rating protection at DB level
);
```

---

## Authentication Flow

### Google OAuth (Organizer/Helper)
```
1. Frontend: redirect → Google OAuth consent screen
2. Google: callback with authorization code → /api/v1/auth/google/callback
3. Backend: exchange code for access token → fetch Google UserInfo
4. Backend: upsert user in DB (google_sub as identifier)
   → if new user: generate unique helper_invite_code and store it
5. Backend: issue JWT (15min) + refresh token (7d)
6. JWT payload: { sub: userId, role: "user", exp: ... }
   ⚠️ No name, no email in JWT — IDs only
   ⚠️ No tournament role in JWT — role is per-tournament, checked in middleware against DB
```

### Rater Login (Nickname + Code)
```
1. POST /api/v1/auth/rater { tournamentSlug, nickname, code }
2. Backend: look up rater in DB (tournament_id + nickname + code)
3. On match: issue JWT (lifetime = voting_end + 24h)
4. JWT payload: { sub: raterId, role: "rater", tournamentId: "...", exp: ... }
   ⚠️ No nickname in JWT
```

---

## Storage Interface (local / S3 switchable)

```go
// internal/storage/interface.go
type Storage interface {
    Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
    Delete(ctx context.Context, key string) error
    PublicURL(key string) string
}

// Selected via configuration:
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

## API Endpoints Overview

```
# Auth
GET  /api/v1/auth/google             # OAuth redirect
GET  /api/v1/auth/google/callback    # OAuth callback → JWT
POST /api/v1/auth/refresh            # Refresh JWT
POST /api/v1/auth/rater              # Rater login (nickname + code)

# Current user profile (JWT: user)
GET /api/v1/me                       # Get own profile incl. helper_invite_code + tournament memberships

# Public
GET  /api/v1/tournaments             # Home page: 5 upcoming + 5 past
GET  /api/v1/tournaments/:slug       # Tournament details (public)
GET  /api/v1/tournaments/:slug/tables/:num   # Table detail (public, for QR)

# Organizer + Helper (JWT: user, membership role checked per tournament)
POST   /api/v1/tournaments                          # Create tournament (creator becomes organizer)
PUT    /api/v1/tournaments/:slug                    # Update tournament
POST   /api/v1/tournaments/:slug/tables             # Create table
PUT    /api/v1/tournaments/:slug/tables/:num        # Update table
POST   /api/v1/tournaments/:slug/tables/:num/photos # Upload photo
DELETE /api/v1/photos/:id                           # Delete photo
POST   /api/v1/tournaments/:slug/raters             # Create rater
GET    /api/v1/tournaments/:slug/raters             # List raters
POST   /api/v1/tournaments/:slug/qrcodes            # Export QR PDF

# Organizer only (JWT: user, membership role = organizer required)
DELETE /api/v1/tournaments/:slug                    # Delete tournament
POST   /api/v1/tournaments/:slug/members            # Add helper by invite_code { invite_code: "X7K392" }
DELETE /api/v1/tournaments/:slug/members/:userId    # Remove a helper (organizer removes others)
GET    /api/v1/tournaments/:slug/results            # View results
PUT    /api/v1/tournaments/:slug/result-config      # Update result config

# Helper self-removal (JWT: user, must be member with role=helper)
DELETE /api/v1/tournaments/:slug/members/me         # Leave tournament as helper

# Rater (JWT: rater)
POST /api/v1/tournaments/:slug/tables/:num/ratings  # Submit rating
GET  /api/v1/tournaments/:slug/tables               # Table overview
```

---

## Docker Compose

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:18-alpine
    volumes: [postgres_data:/var/lib/postgresql/data]
    environment:
      POSTGRES_DB: turnscore
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER}"]
      interval: 5s

  backend:
    build: ./backend
    depends_on:
      postgres: { condition: service_healthy }
    environment:
      APP_DATABASE_HOST: postgres
      APP_DATABASE_USER: ${DB_USER}
      APP_DATABASE_PASSWORD: ${DB_PASSWORD}
      APP_AUTH_GOOGLE_CLIENT_ID: ${GOOGLE_CLIENT_ID}
      APP_AUTH_GOOGLE_CLIENT_SECRET: ${GOOGLE_CLIENT_SECRET}
      APP_AUTH_JWT_SECRET: ${JWT_SECRET}
      APP_STORAGE_BACKEND: ${STORAGE_BACKEND:-local}
    volumes:
      - uploads_data:/app/uploads
    ports: ["8080:8080"]

  frontend:
    build: ./frontend
    ports: ["3000:80"]
    environment:
      VITE_API_URL: ${API_URL:-http://localhost:8080}

volumes:
  postgres_data:
  uploads_data:
```

---

## Test Strategy

### Backend — Target: >80% coverage

| Type | Tool | What is tested |
|---|---|---|
| Unit tests | `go test` + `testify` | Service layer, business logic, helpers |
| Mock tests | `gomock` | Repositories mocked, services tested in isolation |
| Integration tests | `go test` + test DB | Repository layer against real PostgreSQL |
| API tests | `httptest` + `testify` | Handler tests with real HTTP request/response |

**Test DB pattern:** Each test runs in its own transaction → rollback at end (fast, isolated).

### Frontend — Target: >75% coverage

| Type | Tool | What is tested |
|---|---|---|
| Unit tests | `Vitest` + `React Testing Library` | Components, hooks, utils |
| Accessibility tests | `jest-axe` in RTL tests | Detect WCAG violations automatically |
| E2E tests | `Playwright` | Critical user flows (rating, login, QR scan) |

---

## Accessibility (WCAG 2.1 AA)

shadcn/ui is built on **Radix UI Primitives** — many a11y features are built in:
- Correct ARIA roles for all interactive elements
- Keyboard navigation (Tab, Enter, Space, Escape)
- Focus management in modals and dialogs
- Screen-reader-compatible announcements

**Additional requirements:**
- All images: `alt` texts (table photos: table name + category as alt text)
- Color contrast: min. 4.5:1 for body text (verify Tailwind palette)
- Rating sliders: keyboard alternative (arrow keys)
- Form labels: always visible, not just as placeholders
- Error messages: with `role="alert"` and ARIA description

**Theme requirement:** Both themes (Fantasy + Sci-Fi) must meet contrast requirements — decorative elements must not impair readability.

---

## License Management

### Project License: MIT
`LICENSE` file always in project root.

### Allowed Dependency Licenses (✅ compatible with MIT)
- MIT, BSD-2-Clause, BSD-3-Clause, ISC → no issues
- Apache 2.0 → OK (patent clause, but MIT-compatible)

### Not Allowed (❌)
- GPL v2/v3, AGPL → copyleft applies to entire project
- LGPL → review required
- Proprietary/Commercial → not without written permission

### Current Dependencies & Licenses

| Package | License |
|---|---|
| gin-gonic/gin | MIT |
| gorm.io/gorm | MIT |
| golang-migrate | MIT |
| spf13/cobra | Apache 2.0 |
| spf13/viper | MIT |
| rs/zerolog | MIT |
| golang-jwt/jwt | MIT |
| golang.org/x/oauth2 | BSD-3-Clause |
| aws-sdk-go-v2 | Apache 2.0 |
| go-qrcode | MIT |
| jung-kurt/gofpdf | MIT |
| google/uuid | BSD-3-Clause |
| stretchr/testify | MIT |
| uber-go/mock | Apache 2.0 |
| React | MIT |
| Vite | MIT |
| shadcn/ui | MIT |
| Tailwind CSS | MIT |
| TanStack Query | MIT |
| Zustand | MIT |
| react-i18next | MIT |
| Playwright | Apache 2.0 |
| Vitest | MIT |
| jest-axe | MIT |
| lucide-react | ISC |
| PostgreSQL | PostgreSQL License (BSD-like) |

### Rule: Adding a New Dependency
Always in this order:
1. Check the license of the new dependency
2. Confirm it is MIT-compatible (see table above)
3. Run `make licenses`
4. Include `THIRD-PARTY-LICENSES.md` in the same commit as the new dependency

---

## Architecture Decision Records

### ADR-001: GORM instead of sqlc
**Decision:** GORM as ORM, golang-migrate for schema migrations.
**Reason:** Faster development, good GORM hook integration. golang-migrate for versioned SQL migrations (no AutoMigrate in prod).

### ADR-002: Local storage as default, S3 optional
**Decision:** Storage abstracted via interface; default = local filesystem; S3 activatable via config.
**Reason:** Simple start without external services; easy switch to AWS S3/Cloudflare R2 without code changes.

### ADR-003: Contract-first API (OpenAPI + oapi-codegen)
**Decision:** OpenAPI spec as single source of truth; Go server and TypeScript client are generated.
**Reason:** Frontend and backend always in sync; API docs automatically available; type safety across the entire stack boundary.

### ADR-004: Google OAuth only (Phase 1)
**Decision:** No email/password login in Phase 1.
**Reason:** No SMTP setup, no 2FA concept, no password reset needed. Significantly reduces security risks.

### ADR-005: Slug instead of UUID in URLs
**Decision:** Tournaments identified via readable slug (e.g. "essen-open-2025").
**Reason:** QR code URLs stay readable and easy to share.

### ADR-006: JSONB for criteria_scores and links
**Decision:** Flexible fields stored as JSONB, not normalized tables.
**Reason:** Criteria and links are configurable per tournament; avoids unnecessary schema migrations.

### ADR-007: Gin instead of Chi or stdlib
**Decision:** Gin as HTTP framework.
**Reason:** Well-supported by oapi-codegen (gin-server generator); proven middleware ecosystem; good performance.

### ADR-008: Helper role per-tournament, not global
**Decision:** No global `role` field on `User`. Roles (`organizer`/`helper`) live in `tournament_members`. JWT for authenticated users contains only `role: "user"` — tournament-specific permissions checked in middleware against DB.
**Reason:** A user can be organizer of one tournament and helper of another simultaneously. A global role would be incorrect and would require re-issuing JWTs on every membership change.
**Consequence:** Every protected endpoint that checks organizer/helper permission needs one DB lookup against `tournament_members`.

### ADR-009: Helper invite via permanent personal code
**Decision:** Each user has a permanent, unique `helper_invite_code` generated on first login. Organizers enter this code to add a helper to their tournament. No email-based invitation, no time-limited tokens.
**Reason:** Simple, no email infrastructure required, helper can share the code verbally or in a chat. Code is permanent so the user only needs to share it once with recurring collaborators.
