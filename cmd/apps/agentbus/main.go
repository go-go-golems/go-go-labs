package main

import (
	"context"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-labs/cmd/apps/agentbus/pkg/commands"
	"github.com/go-go-golems/go-go-labs/cmd/apps/agentbus/pkg/doc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "agentbus",
	Short: "Redis-backed CLI tool for coordinating coding sub-agents",
	Long: `AgentBus provides a Redis-backed communication layer for coding sub-agents to coordinate their work.

It offers three main coordination primitives:
- Chat streams (speak/overhear) for real-time communication
- Knowledge snippets (jot/recall) for shared documentation and TIL notes  
- Coordination flags (announce/await/satisfy) for dependency management

Each agent identifies itself with AGENT_ID (via --agent flag or env var).
All state is namespaced by agent ID to prevent conflicts.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set up logging
		level, _ := cmd.Flags().GetString("log-level")
		logLevel, err := zerolog.ParseLevel(level)
		if err != nil {
			log.Warn().Str("level", level).Msg("Invalid log level, using info")
			logLevel = zerolog.InfoLevel
		}
		zerolog.SetGlobalLevel(logLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().String("agent", "", "Agent ID (can also be set via AGENT_ID env var)")
	rootCmd.PersistentFlags().String("redis-url", "redis://localhost:6379", "Redis connection URL")
	rootCmd.PersistentFlags().String("format", "json", "Output format (json, text)")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")

	// Bind flags to environment variables
	viper.BindPFlag("agent", rootCmd.PersistentFlags().Lookup("agent"))
	viper.BindPFlag("redis-url", rootCmd.PersistentFlags().Lookup("redis-url"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))

	viper.BindEnv("agent", "AGENT_ID")
	viper.BindEnv("redis-url", "REDIS_URL")
}

func main() {
	// Set up help system
	helpSystem := help.NewHelpSystem()
	err := doc.AddDocToHelpSystem(helpSystem)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load help system")
	}

	// Create all commands
	speakCmd, err := commands.NewSpeakCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create speak command")
	}

	overhearCmd, err := commands.NewOverhearCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create overhear command")
	}

	jotCmd, err := commands.NewJotCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create jot command")
	}

	recallCmd, err := commands.NewRecallCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create recall command")
	}

	listCmd, err := commands.NewListCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create list command")
	}

	announceCmd, err := commands.NewAnnounceCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create announce command")
	}

	awaitCmd, err := commands.NewAwaitCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create await command")
	}

	satisfyCmd, err := commands.NewSatisfyCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create satisfy command")
	}

	// Convert to cobra commands
	speakCobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(speakCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build speak cobra command")
	}

	overhearCobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(overhearCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build overhear cobra command")
	}

	jotCobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(jotCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build jot cobra command")
	}

	recallCobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(recallCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build recall cobra command")
	}

	listCobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(listCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build list cobra command")
	}

	announceCobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(announceCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build announce cobra command")
	}

	awaitCobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(awaitCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build await cobra command")
	}

	satisfyCobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(satisfyCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build satisfy cobra command")
	}

	// Add commands to root
	rootCmd.AddCommand(speakCobraCmd)
	rootCmd.AddCommand(overhearCobraCmd)
	rootCmd.AddCommand(jotCobraCmd)
	rootCmd.AddCommand(recallCobraCmd)
	rootCmd.AddCommand(listCobraCmd)
	rootCmd.AddCommand(announceCobraCmd)
	rootCmd.AddCommand(awaitCobraCmd)
	rootCmd.AddCommand(satisfyCobraCmd)

	// Execute
	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
