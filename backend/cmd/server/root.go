// Package server provides the Cobra CLI entry point for the TurnScore backend.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/r3st/turnscore/config"
	dbmigrations "github.com/r3st/turnscore/db/migrations"
	"github.com/r3st/turnscore/internal/api"
	"github.com/r3st/turnscore/internal/api/handlers"
	"github.com/r3st/turnscore/internal/domain"
	"github.com/r3st/turnscore/internal/repository"
	"github.com/r3st/turnscore/internal/service"
	"github.com/r3st/turnscore/internal/storage"
	"github.com/r3st/turnscore/pkg/logger"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "turnscore",
	Short: "TurnScore — Tabletop Tournament Table Rating Server",
	RunE:  runServer,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: config/config.yaml)")
	rootCmd.Flags().String("port", "", "server port (overrides config)")
	rootCmd.Flags().String("log-level", "", "log level: debug|info|warn|error (overrides config)")
	rootCmd.Flags().String("storage-backend", "", "storage backend: local|s3 (overrides config)")

	viper.BindPFlag("server.port", rootCmd.Flags().Lookup("port"))
	viper.BindPFlag("logging.level", rootCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("storage.backend", rootCmd.Flags().Lookup("storage-backend"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("config")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	// ENV overrides: APP_SERVER_PORT maps to server.port, APP_DATABASE_HOST to database.host, etc.
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintln(os.Stderr, "error reading config:", err)
			os.Exit(1)
		}
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	log := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	log.Info().Str("version", "0.1.0").Msg("starting turnscore server")

	// --- Storage ---
	stor, err := storage.New(cfg.Storage)
	if err != nil {
		return fmt.Errorf("initializing storage: %w", err)
	}
	log.Info().Str("backend", cfg.Storage.Backend).Msg("storage initialized")

	// --- Database ---
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	log.Info().Str("host", cfg.Database.Host).Int("port", cfg.Database.Port).Msg("database connected")

	// --- Migrations ---
	if err := runMigrations(cfg.Database.MigrateDSN()); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	log.Info().Msg("migrations applied")

	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	tourneyRepo := repository.NewTournamentRepository(db)
	memberRepo := repository.NewMembershipRepository(db)
	tableRepo := repository.NewTableRepository(db)
	photoRepo := repository.NewPhotoRepository(db)
	raterRepo := repository.NewRaterRepository(db)
	ratingRepo := repository.NewRatingRepository(db)

	// --- JWT Service ---
	accessExpiry, err := time.ParseDuration(cfg.Auth.JWTExpiry)
	if err != nil {
		return fmt.Errorf("parsing jwt_expiry %q: %w", cfg.Auth.JWTExpiry, err)
	}
	refreshExpiry, err := time.ParseDuration(cfg.Auth.RefreshExpiry)
	if err != nil {
		return fmt.Errorf("parsing refresh_expiry %q: %w", cfg.Auth.RefreshExpiry, err)
	}
	jwtSvc := service.NewJWTService(domain.JWTConfig{
		Secret:        cfg.Auth.JWTSecret,
		AccessExpiry:  accessExpiry,
		RefreshExpiry: refreshExpiry,
	})

	// --- Services ---
	authSvc := service.NewAuthService(cfg.Auth, jwtSvc, userRepo, raterRepo, tourneyRepo)
	tournamentSvc := service.NewTournamentService(tourneyRepo, memberRepo, userRepo)
	tableSvc := service.NewTableService(tableRepo, memberRepo, tourneyRepo, photoRepo, stor)
	qrSvc := service.NewQRCodeService(tableRepo, memberRepo, tourneyRepo, stor, cfg.Server.FrontendURL)
	raterSvc := service.NewRaterService(raterRepo, ratingRepo, tableRepo, memberRepo, tourneyRepo, userRepo)
	userSvc := service.NewUserService(userRepo, memberRepo, tourneyRepo)

	// --- Handlers & Router ---
	h := handlers.New(tournamentSvc, tableSvc, qrSvc, raterSvc, userSvc, authSvc)
	router := api.NewRouter(cfg, jwtSvc, h)

	// --- HTTP Server with graceful shutdown ---
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background
	serverErr := make(chan error, 1)
	go func() {
		log.Info().Str("addr", addr).Msg("HTTP server listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	// Wait for interrupt or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		log.Info().Str("signal", sig.String()).Msg("shutdown signal received")
	}

	// Graceful shutdown: give in-flight requests up to 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	log.Info().Msg("server stopped")
	return nil
}

// runMigrations applies all pending database migrations using the embedded SQL files.
func runMigrations(dsn string) error {
	src, err := iofs.New(dbmigrations.Migrations, ".")
	if err != nil {
		return fmt.Errorf("creating migration source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("applying migrations: %w", err)
	}
	return nil
}
