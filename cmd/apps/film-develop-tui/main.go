package main

import (
	"context"
	"os"

	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	logLevel string
	version  = "dev"
)

func main() {
	setupLogging()
	
	rootCmd := &cobra.Command{
		Use:   "film-develop-tui",
		Short: "Film development calculator TUI",
		Long:  "A terminal user interface for calculating B&W film development parameters using ILFOSOL 3, ILFOSTOP, and Sprint Fixer.",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the log level (debug, info, warn, error)")
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))

	rootCmd.AddCommand(cmd.NewStartCommand())

	if err := rootCmd.ExecuteContext(context.Background()); err != nil {
		log.Error().Err(err).Msg("Command execution failed")
		os.Exit(1)
	}
}

func setupLogging() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
