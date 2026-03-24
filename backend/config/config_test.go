package config_test

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/r3st/turnscore/config"
)

// Tests run with working directory = backend/config/, so config.yaml is directly accessible.

func TestLoadDefaults(t *testing.T) {
	viper.Reset()
	viper.SetConfigFile("config.yaml")
	require.NoError(t, viper.ReadInConfig())

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "release", cfg.Server.Mode)
	assert.Equal(t, "local", cfg.Storage.Backend)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
}

func TestLoadENVOverridesConfig(t *testing.T) {
	viper.Reset()
	viper.SetConfigFile("config.yaml")
	require.NoError(t, viper.ReadInConfig())

	t.Setenv("APP_SERVER_PORT", "9090")
	t.Setenv("APP_LOGGING_LEVEL", "debug")
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Logging.Level)
}

func TestDatabaseConfigDSN(t *testing.T) {
	db := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "turnscore",
		User:     "app",
		Password: "secret",
		SSLMode:  "disable",
	}
	dsn := db.DSN()
	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "dbname=turnscore")
	assert.Contains(t, dsn, "user=app")
	assert.Contains(t, dsn, "sslmode=disable")
	// Password present in DSN but never logged (checked by convention, not test)
	assert.Contains(t, dsn, "password=secret")
}
