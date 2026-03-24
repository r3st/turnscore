# CLAUDE.md — TurnScore
> Tabletop Tournament Table Rating App
> GitHub: https://github.com/[USERNAME]/turnscore (placeholder)

---

## What is this project?

A web application for tabletop miniature game tournaments (Fantasy & Sci-Fi systems).
Organizers present their game tables online; participants rate them based on defined
criteria via QR code and smartphone.

**Key terms Claude must know:**
- **Rater** = Tournament participant without a classic account — login via nickname + code
- **Organizer** = Tournament host — Google OAuth login, full permissions
- **Helper** = Assistant — same rights as Organizer except results and tournament deletion
- **Deployment Zone A/B** = the two deployment areas on a game table
- **Criteria Score** = School grade 1–6 (1 = best, 6 = worst → rankings sorted ASC!)
- **Slug** = human-readable URL identifier for tournaments (e.g. "essen-open-2025")
- **Tournament type** = `fantasy` or `scifi` — determines UI theme

---

## Project Structure

```
turnscore/
├── CLAUDE.md                   # This file
├── LICENSE                     # MIT License (always present)
├── THIRD-PARTY-LICENSES.md     # All dependency licenses (auto-generated)
├── README.md                   # Project overview, setup, contributing
├── .claude/
│   └── skills/                 # All .skill files go here
├── Makefile                    # All common commands
├── docker-compose.yml          # Production
├── docker-compose.dev.yml      # Development
├── .env.example                # All ENV variables documented, NO real values
├── .gitignore                  # .env, uploads/, *.secret never committed
├── .github/
│   └── pull_request_template.md
├── api/
│   └── openapi.yaml            # API spec (single source of truth)
├── assets/
│   ├── logo.svg                # Main logo (abstract, trademark-free)
│   ├── logo-fantasy.svg        # Fantasy theme variant
│   └── logo-scifi.svg          # Sci-Fi theme variant
├── backend/                    # Go 1.24+
│   ├── config/
│   │   └── config.yaml         # Default configuration
│   ├── cmd/server/             # Cobra CLI entry point
│   ├── internal/
│   │   ├── api/
│   │   │   ├── generated/      # ⛔ oapi-codegen output — NEVER edit manually
│   │   │   ├── handlers/
│   │   │   └── middleware/
│   │   ├── domain/
│   │   ├── repository/
│   │   ├── service/
│   │   └── storage/
│   └── db/migrations/
└── frontend/                   # React 18 + TypeScript + Vite
    ├── src/
    │   ├── api/
    │   │   └── generated/      # ⛔ openapi-typescript output — NEVER edit manually
    │   ├── components/
    │   ├── pages/
    │   ├── themes/             # fantasy.css / scifi.css
    │   └── i18n/               # en.json / de.json
    └── e2e/                    # Playwright tests
```

---

## Skills — When to Use Which

| Skill | When to use |
|---|---|
| **requirements** | Planning new features, user stories, domain questions, "what should X do?" |
| **architecture** | DB schema, API endpoints, new packages, Docker, configuration |
| **backend** | Any Go code: handlers, services, repositories, tests, migrations |
| **frontend** | Any React/TS code: components, hooks, tests, themes, i18n |

---

## Makefile Commands

```bash
make dev           # Start Docker Compose dev stack
make generate      # OpenAPI → Go + TypeScript types
make test          # All tests: backend + frontend + e2e
make test-be       # Backend tests with coverage
make test-fe       # Frontend tests with coverage
make test-e2e      # Playwright E2E tests only
make migrate-up    # Run DB migrations
make migrate-down  # Roll back last migration
make lint          # golangci-lint + eslint
make licenses      # Regenerate THIRD-PARTY-LICENSES.md
```

---

## Absolute Rules (NEVER break)

### Privacy & Anonymity
- ❌ NEVER expose `rater_id`, nickname, or rater identity in API responses visible to Organizers/Helpers
- ❌ NEVER log email addresses, names, or nicknames in zerolog
- ❌ NEVER log request bodies (may contain codes or passwords)
- ❌ NEVER log full query parameters

### Code Generation
- ❌ NEVER manually edit `backend/internal/api/generated/`
- ❌ NEVER manually edit `frontend/src/api/generated/`
- ❌ NEVER use `gorm.AutoMigrate()` in production — always use golang-migrate SQL files

### Security & Git
- ❌ NEVER commit secrets to code or `.env` — always use `.env.example` with placeholders
- ❌ NEVER include name or email in JWT claims — only `sub` (UserID) and `role`
- ❌ NEVER commit the `uploads/` directory

### Rating Scale
- ⚠️ Criteria scores are school grades 1–6: **1 = best grade**
- ⚠️ Rankings and sorting must always be **ascending (ASC)**

### Licenses (GitHub publication)
- ✅ For every new Go dependency: check license, update `THIRD-PARTY-LICENSES.md`
- ✅ For every new npm package: check license, update `THIRD-PARTY-LICENSES.md`
- ✅ Run `make licenses` and include it in the same commit as the new dependency
- ❌ Do NOT use GPL-licensed dependencies (incompatible with MIT for SaaS)
- ❌ Do NOT use proprietary/commercial dependencies without review

### Trademark & Assets
- ❌ NEVER use names, logos, or assets from game systems
  (Warhammer, 40K, Infinity, Age of Sigmar, Star Wars Legion, etc. are protected trademarks)
- ❌ NEVER use fonts characteristic of a game system
- ✅ Logo is abstract and trademark-free (geometric grid + star symbol)
- ✅ Only use Google Fonts or other freely licensed typefaces
- ✅ Icons only from free-license sources (Lucide Icons, Heroicons — MIT licensed)

---

## License Strategy

**Project license: MIT**
File `LICENSE` in project root — must always be present.

**Allowed dependency licenses (✅ compatible with MIT):**
- MIT, BSD-2-Clause, BSD-3-Clause, ISC, Apache 2.0 → fine, just document

**Not allowed (❌):**
- GPL, AGPL → copyleft would apply to entire project
- LGPL → review required
- Proprietary/Commercial → not without explicit permission

**`make licenses` auto-generates `THIRD-PARTY-LICENSES.md`:**
```bash
# Backend
go-licenses report ./... > THIRD-PARTY-LICENSES-go.txt

# Frontend
npx license-checker --production --out THIRD-PARTY-LICENSES-npm.txt
```

---

## Logo Guidelines (trademark-free)

TurnScore logo is built exclusively from:
- Geometric shapes (grid/raster = game board viewed from above)
- Star/rating symbol (universal, not protected)
- Free Google Fonts: **Cinzel** (Fantasy) / **Orbitron** (Sci-Fi) / **Inter** (Neutral)

**What MUST NEVER appear in the logo:**
- Miniature silhouettes resembling any known game system
- Faction symbols or colors (Aquila, Chaos Star, etc.)
- Typefaces characteristic of a game system

**AI image generation prompt (Option A — abstract logo):**
> *"Minimalist app logo, top-down view of a hexagonal game board grid, small star rating icon integrated in center, flat vector design, two color variants: warm brown and gold (fantasy), dark navy and cyan neon (scifi), no text, SVG style, clean geometric shapes only"*

---

## Git Workflow (GitHub Flow)

### Principle
- `main` is always deployable — no direct pushes
- Every feature / bugfix gets its own branch + PR
- PRs are squash-merged into `main` → clean history

### Branch Naming
```
feat/[short-description]       # New feature
fix/[short-description]        # Bug fix
chore/[short-description]      # Config, dependencies, tooling
docs/[short-description]       # Documentation only
refactor/[short-description]   # Refactoring without behavior change
```

Examples:
```
feat/openapi-spec
feat/backend-google-oauth
feat/frontend-rating-form
fix/duplicate-rating-check
chore/license-and-dependencies
```

### Step-by-Step Workflow
```bash
# 1. Always start from up-to-date main
git checkout main && git pull

# 2. Create feature branch
git checkout -b feat/[description]

# 3. Develop + commit (Conventional Commits)
git add -p                          # Stage selectively, never blind "git add ."
git commit -m "feat: [description]"

# 4. Push + open PR
git push -u origin feat/[description]

# 5. After review: Squash & Merge into main
# 6. Delete branch (enable auto-delete in GitHub Settings)
```

### Conventional Commits — Required
Format: `type(scope): description`

| Type | When |
|---|---|
| `feat` | New feature |
| `fix` | Bug fix |
| `chore` | Build, dependencies, configuration |
| `docs` | Documentation only |
| `test` | Tests added/changed only |
| `refactor` | Code change without new behavior |

Examples:
```
feat(backend): add google oauth handler
feat(frontend): implement rating form with 1-6 scale
fix(backend): prevent duplicate ratings on concurrent requests
chore(deps): update gin to v1.10.0
```

### Rules for Claude Code in Git
- ✅ Before every new feature: switch to main and pull
- ✅ Create a new feature branch before writing any code
- ✅ Use `git add -p` instead of `git add .`
- ✅ Run `make licenses` and commit it when adding new dependencies
- ✅ Commit generated files (`generated/`) — they belong in the repo
- ❌ Never push directly to `main`
- ❌ Never commit `.env` or secrets
- ❌ Never commit `uploads/`

### PR Template (`.github/pull_request_template.md`)
```markdown
## What was done?
[Short description]

## Why?
[Reference to user story / feature]

## Checklist
- [ ] Tests written and passing
- [ ] `make licenses` run (if new dependencies added)
- [ ] No secrets in code
- [ ] CLAUDE.md rules followed
```

---

## Key Architectural Decisions (do not change without ADR)

| Decision | Reason |
|---|---|
| Go + Gin | Performant, works well with oapi-codegen |
| GORM + golang-migrate | GORM for queries, migrate for versioned schema changes |
| Google OAuth only | No SMTP, no 2FA, no password reset needed |
| OpenAPI contract-first | Frontend + backend always in sync, types generated |
| Local storage as default | No MinIO/S3 needed for local development |
| PostgreSQL 18 | JSONB for flexible criteria configuration |
| School grades 1–6 | Well-known system in European tabletop community |
| MIT license | Compatible with all dependencies used |

For changes: create an ADR in `docs/decisions/`.

---

## Recommended Development Order

1. `api/openapi.yaml` — write the spec first
2. `make generate` — generate server + client code
3. Backend: Migration → Repository → Service → Handler → Test
4. Frontend: Types already exist → Hook → Component → Page → Test
5. `make licenses` — after every new dependency

---

## Open TODOs Before GitHub Publication

- [ ] Fill in `LICENSE` (MIT — add year + name)
- [ ] Run `make licenses` to generate initial `THIRD-PARTY-LICENSES.md`
- [ ] Write `README.md` with screenshot, setup, contributing guide
- [ ] Create Google OAuth Client ID (Google Cloud Console)
- [ ] Verify `.gitignore`: `.env`, `uploads/`, `*.secret` excluded?
- [ ] Create logo (abstract, geometric — see logo guidelines above)
- [ ] Decide on tournament type badges (🏰 Fantasy / 🚀 Sci-Fi — or different icons?)
- [ ] Clarify domain / hosting
