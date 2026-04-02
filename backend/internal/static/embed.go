// Package static embeds the compiled React frontend so the Go binary can serve it.
// The dist/ directory is populated by running `npm run build` in the frontend workspace
// before `go build`. In development the frontend is served by Vite directly.
package static

import "embed"

//go:embed all:dist
var FS embed.FS
