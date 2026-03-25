// Package migrations embeds all SQL migration files for use with golang-migrate.
package migrations

import "embed"

// Migrations contains all *.sql migration files embedded at compile time.
//
//go:embed *.sql
var Migrations embed.FS
