// Package server provides the Cobra CLI entry point for the TurnScore backend.
package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/r3st/turnscore/config"
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

	// TODO: wire up DB, storage, services, router and start HTTP server
	_ = cfg
	return nil
}
