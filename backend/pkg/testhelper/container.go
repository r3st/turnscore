package testhelper

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewMigrationDB starts a fresh PostgreSQL container, runs all migrations up,
// and returns a GORM connection. The container is stopped after the test.
//
// Use this only for migration tests — it is slower than NewTestDB because
// it starts a real container (~5–10s). For repository tests, use NewTestDB.
func NewMigrationDB(t *testing.T, migrationsPath string) *gorm.DB {
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
	require.NoError(t, err, "starting PostgreSQL test container")

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("warning: failed to terminate test container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "getting container connection string")

	// Run migrations up
	m, err := migrate.New(fmt.Sprintf("file://%s", migrationsPath), dsn)
	require.NoError(t, err, "creating migrate instance")

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err, "running migrations up")
	}

	db, err := gorm.Open(pgdriver.Open(dsn), &gorm.Config{
		Logger: logger.Discard,
	})
	require.NoError(t, err, "opening GORM connection to test container")

	return db
}
