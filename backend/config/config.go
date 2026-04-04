// Package config provides configuration loading via Viper.
// Priority (lowest → highest): config.yaml → ENV variables (APP_*) → CLI flags.
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config is the root configuration structure.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Legal    LegalConfig    `mapstructure:"legal"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Mode        string `mapstructure:"mode"`         // debug | release
	FrontendURL string `mapstructure:"frontend_url"` // for CORS
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
}

// DSN returns the PostgreSQL connection string in key=value format (used by GORM).
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		d.Host, d.Port, d.Name, d.User, d.Password, d.SSLMode,
	)
}

// MigrateDSN returns the PostgreSQL URL used by golang-migrate.
func (d DatabaseConfig) MigrateDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// AuthConfig holds Google OAuth and JWT settings.
type AuthConfig struct {
	GoogleClientID     string `mapstructure:"google_client_id"`
	GoogleClientSecret string `mapstructure:"google_client_secret"`
	OAuthCallbackURL   string `mapstructure:"oauth_callback_url"` // e.g. "http://localhost:8080/api/v1/auth/google/callback"
	JWTSecret          string `mapstructure:"jwt_secret"`
	JWTExpiry          string `mapstructure:"jwt_expiry"`    // e.g. "15m"
	RefreshExpiry      string `mapstructure:"refresh_expiry"` // e.g. "168h"
}

// StorageConfig holds file storage settings.
type StorageConfig struct {
	Backend string            `mapstructure:"backend"` // local | s3
	Local   LocalStorageConfig `mapstructure:"local"`
	S3      S3StorageConfig    `mapstructure:"s3"`
}

// LocalStorageConfig holds local filesystem storage settings.
type LocalStorageConfig struct {
	Path    string `mapstructure:"path"`
	BaseURL string `mapstructure:"base_url"`
}

// S3StorageConfig holds S3-compatible storage settings.
type S3StorageConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	Bucket    string `mapstructure:"bucket"`
	Region    string `mapstructure:"region"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
}

// LegalConfig holds paths to the mounted legal text files.
type LegalConfig struct {
	Path string `mapstructure:"path"` // directory containing imprint.{lang}.md and privacy.{lang}.md
}

// LoggingConfig holds zerolog settings.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`  // debug | info | warn | error
	Format string `mapstructure:"format"` // json | pretty
}

// Load reads configuration from Viper (already initialized by Cobra).
func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	return &cfg, nil
}
