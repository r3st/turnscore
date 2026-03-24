package migrations_test

import (
	"context"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

const migrationsSource = "file://."

func newPostgresContainer(t *testing.T) (dsn string, cleanup func()) {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("turnscore_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "starting PostgreSQL container")

	dsn, err = container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "getting container DSN")

	return dsn, func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("warning: failed to stop container: %v", err)
		}
	}
}

// TestMigrationsUp verifies that all migrations apply cleanly to a fresh database.
func TestMigrationsUp(t *testing.T) {
	dsn, cleanup := newPostgresContainer(t)
	defer cleanup()

	m, err := migrate.New(migrationsSource, dsn)
	require.NoError(t, err)

	err = m.Up()
	assert.NoError(t, err, "all migrations should apply without error")

	version, dirty, err := m.Version()
	require.NoError(t, err)
	assert.False(t, dirty, "database should not be in dirty state after migrations")
	assert.Equal(t, uint(7), version, "should be at migration version 7")
}

// TestMigrationsUpDown verifies that every migration can be rolled back cleanly.
func TestMigrationsUpDown(t *testing.T) {
	dsn, cleanup := newPostgresContainer(t)
	defer cleanup()

	m, err := migrate.New(migrationsSource, dsn)
	require.NoError(t, err)

	require.NoError(t, m.Up(), "migrations up should succeed")

	// Roll back all migrations at once
	err = m.Down()
	require.NoError(t, err, "rolling back all migrations should not error")

	_, _, err = m.Version()
	assert.ErrorIs(t, err, migrate.ErrNilVersion, "all migrations rolled back: version should be nil")
}

// TestMigrationsIdempotent verifies that running Up twice does not error.
func TestMigrationsIdempotent(t *testing.T) {
	dsn, cleanup := newPostgresContainer(t)
	defer cleanup()

	m, err := migrate.New(migrationsSource, dsn)
	require.NoError(t, err)

	require.NoError(t, m.Up())

	err = m.Up()
	assert.ErrorIs(t, err, migrate.ErrNoChange, "second Up() should return ErrNoChange")
}
