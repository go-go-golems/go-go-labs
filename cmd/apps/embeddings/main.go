package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	clay "github.com/go-go-golems/clay/pkg"
	embeddings_config "github.com/go-go-golems/geppetto/pkg/embeddings/config"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// Declare the embeddings_commands variable that will be appended to in server.go
var embeddings_commands []cmds.Command

func main() {
	// Initialize logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Create and configure root command
	rootCmd := &cobra.Command{
		Use:   "embeddings",
		Short: "Embeddings and similarity computation tool",
		Long: `A tool for generating embeddings and computing similarity between texts.
It provides both CLI commands and a web server with APIs for:
- Computing embeddings for texts
- Computing similarity scores between texts
- Web UI for comparing up to three texts`,
	}

	// Initialize Viper for config file support
	err := clay.InitViper("embeddings", rootCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Viper")
	}

	// Set up help system
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Register all commands using the parser
	err = cli.AddCommandsToRootCommand(rootCmd, embeddings_commands, []*alias.CommandAlias{},
		cli.WithCobraMiddlewaresFunc(GetEmbeddingsMiddlewares),
		cli.WithProfileSettingsLayer())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to add commands to root command")
	}

	// Create context with signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Execute the root command with context
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatal().Err(err).Msg("Error executing command")
	}

	log.Info().Msg("Embeddings application exiting")
}

// GetEmbeddingsLayers returns all parameter layers used by the embeddings commands
func GetEmbeddingsLayers() ([]layers.ParameterLayer, error) {
	// Create embeddings parameter layers
	embeddingsLayer, err := embeddings_config.NewEmbeddingsParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create embeddings parameter layer")
	}

	embeddingsApiKey, err := embeddings_config.NewEmbeddingsApiKeyParameter()
	if err != nil {
		return nil, errors.Wrap(err, "could not create embeddings API key parameter layer")
	}

	// Command settings layer (for --config flag)
	commandSettingsLayer, err := cli.NewCommandSettingsLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create command settings parameter layer")
	}

	return []layers.ParameterLayer{
		embeddingsLayer,
		embeddingsApiKey,
		commandSettingsLayer,
	}, nil
}

// GetEmbeddingsMiddlewares returns the middleware stack for processing commands
func GetEmbeddingsMiddlewares(
	parsedLayers *layers.ParsedLayers,
	cmd *cobra.Command,
	args []string,
) ([]middlewares.Middleware, error) {
	// Parse command settings
	commandSettings := &cli.CommandSettings{}
	err := parsedLayers.InitializeStruct(cli.CommandSettingsSlug, commandSettings)
	if err != nil {
		return nil, err
	}

	// Parse profile settings
	profileSettings := &cli.ProfileSettings{}
	err = parsedLayers.InitializeStruct(cli.ProfileSettingsSlug, profileSettings)
	if err != nil {
		return nil, err
	}

	// Create middleware chain
	middlewareStack := []middlewares.Middleware{
		// Parse command-line flags
		middlewares.ParseFromCobraCommand(cmd,
			parameters.WithParseStepSource("cobra"),
		),
		// Gather command-line arguments
		middlewares.GatherArguments(args,
			parameters.WithParseStepSource("arguments"),
		),
	}

	// Add config file loading if specified
	if commandSettings.LoadParametersFromFile != "" {
		middlewareStack = append(middlewareStack,
			middlewares.LoadParametersFromFile(commandSettings.LoadParametersFromFile))
	}

	// Add profile support
	xdgConfigPath, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	defaultProfileFile := fmt.Sprintf("%s/embeddings/profiles.yaml", xdgConfigPath)
	if profileSettings.ProfileFile == "" {
		profileSettings.ProfileFile = defaultProfileFile
	}
	if profileSettings.Profile == "" {
		profileSettings.Profile = "default"
	}

	// Add profile middleware
	middlewareStack = append(middlewareStack,
		middlewares.GatherFlagsFromProfiles(
			defaultProfileFile,
			profileSettings.ProfileFile,
			profileSettings.Profile,
			parameters.WithParseStepSource("profiles"),
			parameters.WithParseStepMetadata(map[string]interface{}{
				"profileFile": profileSettings.ProfileFile,
				"profile":     profileSettings.Profile,
			}),
		),
	)

	// Add viper and defaults
	middlewareStack = append(middlewareStack,
		middlewares.WrapWithWhitelistedLayers(
			[]string{
				embeddings_config.EmbeddingsSlug,
				embeddings_config.EmbeddingsApiKeySlug,
				cli.CommandSettingsSlug,
				cli.ProfileSettingsSlug,
			},
			middlewares.GatherFlagsFromViper(parameters.WithParseStepSource("viper")),
		),
		middlewares.SetFromDefaults(parameters.WithParseStepSource("defaults")),
	)

	return middlewareStack, nil
}
