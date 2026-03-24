.PHONY: dev dev-down generate test test-be test-fe test-e2e migrate-up migrate-down lint licenses

# ---------------------------------------------------------------------------
# Development
# ---------------------------------------------------------------------------

dev:
	docker compose -f docker-compose.dev.yml up --build

dev-down:
	docker compose -f docker-compose.dev.yml down

# ---------------------------------------------------------------------------
# Code Generation (OpenAPI → Go + TypeScript)
# ---------------------------------------------------------------------------

generate:
	cd backend && go generate ./...
	cd backend && oapi-codegen -config oapi-codegen.yaml ../api/openapi.yaml
	cd frontend && npx openapi-typescript ../api/openapi.yaml -o src/api/generated/api.ts

# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------

test: test-be test-fe test-e2e

test-be:
	cd backend && go test ./... -coverprofile=coverage.out -covermode=atomic
	cd backend && go tool cover -func=coverage.out | tail -1

test-fe:
	cd frontend && npx vitest run --coverage

test-e2e:
	cd frontend && npx playwright test

# ---------------------------------------------------------------------------
# Database Migrations
# ---------------------------------------------------------------------------

migrate-up:
	cd backend && migrate -path db/migrations -database "$$DATABASE_URL" up

migrate-down:
	cd backend && migrate -path db/migrations -database "$$DATABASE_URL" down 1

# ---------------------------------------------------------------------------
# Lint
# ---------------------------------------------------------------------------

lint:
	cd backend && golangci-lint run ./...
	cd frontend && npx eslint src/

# ---------------------------------------------------------------------------
# Licenses (run after every new dependency!)
# ---------------------------------------------------------------------------

licenses:
	cd backend && go-licenses report ./... > ../THIRD-PARTY-LICENSES-go.txt 2>/dev/null || true
	cd frontend && npx license-checker --production --out ../THIRD-PARTY-LICENSES-npm.txt 2>/dev/null || true
	@echo "✅ License files updated. Review and commit THIRD-PARTY-LICENSES.md"
