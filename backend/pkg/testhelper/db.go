// Package testhelper provides shared utilities for backend tests.
package testhelper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewTestDB opens a connection to the test database and wraps the test in a
// transaction that is rolled back when the test ends. This keeps tests fast and
// isolated — each test starts with a clean state.
//
// Requires TEST_DATABASE_URL environment variable:
//
//	TEST_DATABASE_URL=postgres://dev:dev@localhost:5432/turnscore_dev?sslmode=disable
func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	require.NotEmpty(t, dsn, "TEST_DATABASE_URL must be set for integration tests")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Discard, // silence GORM output during tests
	})
	require.NoError(t, err, "opening test database connection")

	tx := db.Begin()
	require.NoError(t, tx.Error, "beginning test transaction")

	t.Cleanup(func() {
		tx.Rollback()
	})

	return tx
}
