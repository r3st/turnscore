# TurnScore

Tabletop tournament table rating app — built for Fantasy & Sci-Fi miniature game events.

Organizers set up their game tables online. Participants rate them (1–6 school grade scale)
via QR code and smartphone. Results are visible to organizers after the tournament.

---

## Quick Start — Local Development

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) + Docker Compose
- A Google OAuth Client ID ([Google Cloud Console](https://console.cloud.google.com/))
  - Authorized redirect URI: `http://localhost:8080/api/v1/auth/google/callback`

### 1. Clone and configure

```bash
git clone https://github.com/r3st/turnscore.git
cd turnscore

cp .env.example .env
# Edit .env — fill in GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET
```

The `.env` file only needs two values for local development:

```dotenv
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
```

All other settings (database, JWT secret, storage) are pre-configured in
`docker-compose.dev.yml` for local use.

### 2. Start the stack

```bash
make dev
```

This starts three services with hot-reload:

| Service    | URL                        | Details                          |
|------------|----------------------------|----------------------------------|
| Frontend   | http://localhost:3000      | React + Vite (HMR)               |
| Backend    | http://localhost:8080      | Go + air (live reload)           |
| PostgreSQL | localhost:5432             | `turnscore_dev` / `dev` / `dev`  |

Migrations run automatically on backend startup.

### 3. Stop the stack

```bash
make dev-down
```

---

## Make Targets

| Command           | Description                                        |
|-------------------|----------------------------------------------------|
| `make dev`        | Start full dev stack (Docker Compose)              |
| `make dev-down`   | Stop dev stack                                     |
| `make generate`   | Regenerate OpenAPI types (Go + TypeScript)         |
| `make test`       | Run all tests (backend + frontend + e2e)           |
| `make test-be`    | Backend unit + integration tests with coverage     |
| `make test-fe`    | Frontend Vitest tests with coverage                |
| `make test-e2e`   | Playwright end-to-end tests                        |
| `make lint`       | golangci-lint + ESLint                             |
| `make licenses`   | Regenerate THIRD-PARTY-LICENSES (run after new deps)|

---

## Architecture

```
turnscore/
├── api/openapi.yaml        # API contract (single source of truth)
├── backend/                # Go 1.25+ — Gin, GORM, oapi-codegen
│   ├── cmd/server/         # Cobra entry point
│   ├── internal/
│   │   ├── api/            # Handlers + middleware (generated/ never edited)
│   │   ├── domain/         # Entities + error types
│   │   ├── repository/     # GORM database access
│   │   ├── service/        # Business logic
│   │   └── storage/        # Local / S3 file storage
│   └── db/migrations/      # golang-migrate SQL files
└── frontend/               # React 18 + TypeScript + Vite
    └── src/
        ├── api/generated/  # openapi-typescript output (never edited)
        ├── components/
        ├── pages/
        ├── themes/         # fantasy.css / scifi.css (CSS custom properties)
        └── i18n/           # en.json / de.json
```

Key decisions:
- **Contract-first**: OpenAPI spec drives both Go handler interfaces and TypeScript types
- **Google OAuth only**: No passwords, no SMTP needed
- **School grades 1–6**: 1 = best — rankings always sorted ASC
- **MIT license**: All dependencies are MIT/Apache/BSD-compatible

---

## Contributing

1. Fork + create a feature branch: `feat/my-feature`
2. Follow [Conventional Commits](https://www.conventionalcommits.org/)
3. Run `make test` before opening a PR
4. Run `make licenses` if you added new dependencies
5. Open a PR — the template guides the rest

See [CLAUDE.md](CLAUDE.md) for full development guidelines.

---

## License

[MIT](LICENSE) — © 2025 TurnScore Contributors
