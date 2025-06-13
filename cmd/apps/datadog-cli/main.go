package main

import (
	"embed"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	clay "github.com/go-go-golems/clay/pkg"
	clay_commandmeta "github.com/go-go-golems/clay/pkg/cmds/commandmeta"
	clay_profiles "github.com/go-go-golems/clay/pkg/cmds/profiles"
	clay_repositories "github.com/go-go-golems/clay/pkg/cmds/repositories"
	clay_doc "github.com/go-go-golems/clay/pkg/doc"
	"github.com/go-go-golems/clay/pkg/repositories"
	"github.com/go-go-golems/glazed/pkg/cli"
	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/types"
	datadog_cmds "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/cmds"
	datadog_layers "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/layers"
	datadog_loaders "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/loaders"
	"github.com/pkg/errors"
	"github.com/pkg/profile"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version = "dev"
var profiler interface {
	Stop()
}

//go:embed doc/*
var docFS embed.FS

//go:embed queries/*
var queriesFS embed.FS

var rootCmd = &cobra.Command{
	Use:   "datadog-cli",
	Short: "YAML-driven CLI for Datadog Logs API",
	Long: `A composable CLI for querying Datadog logs using YAML command definitions.
Supports streaming results, pagination, and all Glazed output formats.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := logging.InitLoggerFromViper()
		cobra.CheckErr(err)

		memProfile, _ := cmd.Flags().GetBool("mem-profile")
		if memProfile {
			log.Info().Msg("Starting memory profiler")
			profiler = profile.Start(profile.MemProfile)

			// on SIGHUP, restart the profiler
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGHUP)
			go func() {
				for range sigCh {
					log.Info().Msg("Restarting memory profiler")
					profiler.Stop()
					profiler = profile.Start(profile.MemProfile)
				}
			}()
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if profiler != nil {
			log.Info().Msg("Stopping memory profiler")
			profiler.Stop()
		}
	},
	Version: version,
}

func main() {
	// Handle run-command specially like sqleton does
	if len(os.Args) >= 3 && os.Args[1] == "run-command" && os.Args[2] != "--help" {
		// load the command
		loader := datadog_loaders.NewDatadogYAMLCommandLoader()
		fs_, filePath, err := loaders.FileNameToFsFilePath(os.Args[2])
		if err != nil {
			fmt.Printf("Could not get absolute path: %v\n", err)
			os.Exit(1)
		}
		cmds, err := loader.LoadCommands(fs_, filePath, []glazed_cmds.CommandDescriptionOption{}, []alias.Option{})
		if err != nil {
			fmt.Printf("Could not load command: %v\n", err)
			os.Exit(1)
		}
		if len(cmds) != 1 {
			fmt.Printf("Expected exactly one command, got %d", len(cmds))
		}

		glazeCommand, ok := cmds[0].(glazed_cmds.GlazeCommand)
		if !ok {
			fmt.Printf("Expected GlazeCommand, got %T", cmds[0])
			os.Exit(1)
		}

		cobraCommand, err := datadog_layers.BuildCobraCommandWithDatadogMiddlewares(glazeCommand)
		if err != nil {
			fmt.Printf("Could not build cobra command: %v\n", err)
			os.Exit(1)
		}

		_, err = initRootCmd()
		cobra.CheckErr(err)

		rootCmd.AddCommand(cobraCommand)
		restArgs := os.Args[3:]
		os.Args = append([]string{os.Args[0], cobraCommand.Use}, restArgs...)
	} else {
		helpSystem, err := initRootCmd()
		cobra.CheckErr(err)

		err = initAllCommands(helpSystem)
		cobra.CheckErr(err)
	}

	err := rootCmd.Execute()
	cobra.CheckErr(err)
}

var runCommandCmd = &cobra.Command{
	Use:   "run-command",
	Short: "Run a command from a file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		panic(errors.Errorf("not implemented"))
	},
}

func initRootCmd() (*help.HelpSystem, error) {
	helpSystem := help.NewHelpSystem()
	err := helpSystem.LoadSectionsFromFS(docFS, ".")
	cobra.CheckErr(err)

	err = clay_doc.AddDocToHelpSystem(helpSystem)
	cobra.CheckErr(err)

	helpSystem.SetupCobraRootCommand(rootCmd)

	err = clay.InitViper("datadog-cli", rootCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(runCommandCmd)

	return helpSystem, nil
}

func initAllCommands(helpSystem *help.HelpSystem) error {
	// Add logs command group
	logsCmd := &cobra.Command{
		Use:   "logs",
		Short: "Datadog logs commands",
	}

	// Add run command for ad-hoc YAML queries
	runCmd, err := datadog_cmds.NewRunCommand()
	if err != nil {
		return err
	}
	runCobraCmd, err := datadog_layers.BuildCobraCommandWithDatadogMiddlewares(runCmd)
	if err != nil {
		return err
	}
	logsCmd.AddCommand(runCobraCmd)

	// Add raw query command
	queryCmd, err := datadog_cmds.NewQueryCommand()
	if err != nil {
		return err
	}
	queryCobraCmd, err := datadog_layers.BuildCobraCommandWithDatadogMiddlewares(queryCmd)
	if err != nil {
		return err
	}
	logsCmd.AddCommand(queryCobraCmd)

	rootCmd.AddCommand(logsCmd)

	// Add test commands for debugging
	rootCmd.AddCommand(datadog_cmds.TestCmd)

	// Setup repositories
	repositoryPaths := viper.GetStringSlice("repositories")

	defaultDirectory := "$HOME/.datadog-cli/queries"
	_, err = os.Stat(os.ExpandEnv(defaultDirectory))
	if err == nil {
		repositoryPaths = append(repositoryPaths, os.ExpandEnv(defaultDirectory))
	}

	loader := datadog_loaders.NewDatadogYAMLCommandLoader()
	directories := []repositories.Directory{
		{
			FS:               queriesFS,
			RootDirectory:    "queries",
			RootDocDirectory: "queries/doc",
			Name:             "datadog-cli",
			SourcePrefix:     "embed",
		}}

	for _, repositoryPath := range repositoryPaths {
		dir := os.ExpandEnv(repositoryPath)
		// check if dir exists
		if fi, err := os.Stat(dir); os.IsNotExist(err) || !fi.IsDir() {
			continue
		}
		directories = append(directories, repositories.Directory{
			FS:               os.DirFS(dir),
			RootDirectory:    ".",
			RootDocDirectory: "doc",
			Name:             dir,
			SourcePrefix:     "file",
		})
	}

	repositories_ := []*repositories.Repository{
		repositories.NewRepository(
			repositories.WithDirectories(directories...),
			repositories.WithCommandLoader(loader),
		),
	}

	allCommands, err := repositories.LoadRepositories(
		helpSystem,
		logsCmd, // Add to logs command instead of root
		repositories_,
		cli.WithCreateCommandSettingsLayer(),
		cli.WithProfileSettingsLayer(),
	)
	if err != nil {
		return err
	}

	// Create and add the unified command management group
	commandManagementCmd, err := clay_commandmeta.NewCommandManagementCommandGroup(
		allCommands,
		clay_commandmeta.WithListAddCommandToRowFunc(func(
			command glazed_cmds.Command,
			row types.Row,
			parsedLayers *layers.ParsedLayers,
		) ([]types.Row, error) {
			switch c := command.(type) {
			case *datadog_cmds.DatadogQueryCommand:
				row.Set("query", c.Query)
				row.Set("type", "datadog")
			case *alias.CommandAlias:
				row.Set("type", "alias")
				row.Set("aliasFor", c.AliasFor)
			default:
				if _, ok := row.Get("type"); !ok {
					row.Set("type", "unknown")
				}
			}
			return []types.Row{row}, nil
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize command management commands: %w", err)
	}
	rootCmd.AddCommand(commandManagementCmd)

	// Create and add the profiles command
	profilesCmd, err := clay_profiles.NewProfilesCommand("datadog-cli", datadogInitialProfilesContent)
	if err != nil {
		return fmt.Errorf("failed to initialize profiles command: %w", err)
	}
	rootCmd.AddCommand(profilesCmd)

	// Create and add the repositories command group
	rootCmd.AddCommand(clay_repositories.NewRepositoriesGroupCommand())

	rootCmd.PersistentFlags().Bool("mem-profile", false, "Enable memory profiling")

	return nil
}

// datadogInitialProfilesContent provides the default YAML content for a new datadog-cli profiles file.
func datadogInitialProfilesContent() string {
	return `# Datadog CLI Profiles Configuration
#
# This file allows defining profiles to override default Datadog connection
# settings or query parameters for datadog-cli commands.
#
# Profiles are selected using the --profile <profile-name> flag.
# Settings within a profile override the default values for the specified layer.
#
# Example:
#
# production:
#   # Override settings for the 'datadog' layer
#   datadog:
#     api-key: "[REDACTED:api-key]"
#     app-key: "[REDACTED:app-key]" 
#     site: "datadoghq.com"
#
# eu-instance:
#   datadog:
#     api-key: "[REDACTED:api-key]"
#     app-key: "[REDACTED:app-key]"
#     site: "datadoghq.eu"
#
# You can manage this file using the 'datadog-cli profiles' commands:
# - list: List all profiles
# - get: Get profile settings
# - set: Set a profile setting
# - delete: Delete a profile, layer, or setting
# - edit: Open this file in your editor
# - init: Create this file if it doesn't exist
# - duplicate: Copy an existing profile
`
}
