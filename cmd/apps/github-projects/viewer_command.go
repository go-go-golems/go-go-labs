package main

import (
	"context"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/go-go-golems/go-go-labs/cmd/apps/github-projects/pkg/github"
)

// ViewerCommand shows current user info
type ViewerCommand struct {
	*cmds.CommandDescription
}

// ViewerSettings holds the command settings
type ViewerSettings struct {
	LogLevel string `glazed.parameter:"log-level"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &ViewerCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *ViewerCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings
	s := &ViewerSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Set up logger
	level, err := zerolog.ParseLevel(s.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	logger := zerolog.New(os.Stderr).Level(level).With().Timestamp().Logger()

	// Create GitHub client
	client, err := github.NewClient(logger)
	if err != nil {
		return errors.Wrap(err, "failed to create GitHub client")
	}

	// Get viewer
	viewer, err := client.GetViewer(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get viewer")
	}

	// Create row from viewer data
	row := types.NewRow(
		types.MRP("login", viewer.Login),
		types.MRP("name", viewer.Name),
		types.MRP("email", viewer.Email),
	)

	return gp.AddRow(ctx, row)
}

// NewViewerCommand creates a new viewer command
func NewViewerCommand() (*ViewerCommand, error) {
	// Create Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"viewer",
		cmds.WithShort("Show current authenticated user information"),
		cmds.WithLong(`
Show information about the currently authenticated GitHub user.
This is useful for testing authentication and verifying your token works.

Examples:
  github-graphql-cli viewer
  github-graphql-cli viewer --output=json
		`),
		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"log-level",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Log level"),
				parameters.WithDefault("info"),
				parameters.WithChoices("trace", "debug", "info", "warn", "error"),
			),
		),
		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &ViewerCommand{
		CommandDescription: cmdDesc,
	}, nil
}
