# ── Stage 1: Build React frontend ────────────────────────────────────────────
FROM node:24-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ── Stage 2: Build Go binary ──────────────────────────────────────────────────
FROM golang:1.26-alpine AS backend-builder
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# Copy compiled frontend into the embed path before go build.
COPY --from=frontend-builder /app/frontend/dist ./internal/static/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /turnscore ./cmd/server

# ── Stage 3: Minimal runtime image ───────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot
# /usr/local/bin is on PATH and executable by all users in distroless:nonroot.
# Copying to / causes "permission denied" because the nonroot user cannot exec from /.
COPY --from=backend-builder /turnscore /usr/local/bin/turnscore
# Copy default config — individual values are overridden via environment variables.
COPY --from=backend-builder /app/backend/config/config.yaml /config/config.yaml
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/turnscore"]
