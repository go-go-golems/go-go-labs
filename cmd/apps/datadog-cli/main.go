package main

import (
	"embed"
	"os"
	"time"

	"github.com/go-go-golems/clay/pkg/repositories"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	datadog_cmds "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/cmds"
	"github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/loaders"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//go:embed queries
var embeddedQueries embed.FS

var rootCmd = &cobra.Command{
	Use:   "datadog-cli",
	Short: "YAML-driven CLI for Datadog Logs API",
	Long: `A composable CLI for querying Datadog logs using YAML command definitions.
Supports streaming results, pagination, and all Glazed output formats.`,
}

func main() {
	// Setup logging
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	// Create help system
	helpSystem := help.NewHelpSystem()

	// Setup repository for embedded queries
	repository := repositories.NewRepository(
		repositories.WithName("datadog-cli"),
		repositories.WithDirectories(
			repositories.Directory{
				FS:            embeddedQueries,
				RootDirectory: "queries",
				Name:          "builtin-queries",
			},
		),
		repositories.WithCommandLoader(loaders.NewDatadogYAMLCommandLoader()),
	)

	// Load embedded commands
	err := repository.LoadCommands(helpSystem)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load embedded commands")
	}

	// Add logs command group
	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "Datadog logs commands",
	}

	// Add run command for ad-hoc YAML queries
	runCmd, err := datadog_cmds.NewRunCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create run command")
	}
	runCobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(runCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build run cobra command")
	}
	logsCmd.AddCommand(runCobraCmd)

	// Load all repository commands as subcommands
	allCommands := repository.CollectCommands([]string{}, true)
	for _, command := range allCommands {
		if datadogCmd, ok := command.(*datadog_cmds.DatadogQueryCommand); ok {
			cobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(datadogCmd)
			if err != nil {
				log.Error().Err(err).Str("command", datadogCmd.Name).Msg("Failed to build cobra command")
				continue
			}
			logsCmd.AddCommand(cobraCmd)
		}
	}

	rootCmd.AddCommand(logsCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Error executing command")
	}
}
