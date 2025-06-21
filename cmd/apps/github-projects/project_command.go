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

// ProjectCommand shows project information
type ProjectCommand struct {
	*cmds.CommandDescription
}

// ProjectSettings holds the command settings
type ProjectSettings struct {
	Owner    string `glazed.parameter:"owner"`
	Number   int    `glazed.parameter:"number"`
	LogLevel string `glazed.parameter:"log-level"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &ProjectCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *ProjectCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings
	s := &ProjectSettings{}
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

	// Get project
	project, err := client.GetProject(ctx, s.Owner, s.Number)
	if err != nil {
		return errors.Wrap(err, "failed to get project")
	}

	// Create row from project data
	row := types.NewRow(
		types.MRP("id", project.ID),
		types.MRP("title", project.Title),
		types.MRP("public", project.Public),
		types.MRP("short_description", project.ShortDescription),
		types.MRP("closed", project.Closed),
		types.MRP("total_items", project.Items.TotalCount),
	)

	return gp.AddRow(ctx, row)
}

// NewProjectCommand creates a new project command
func NewProjectCommand() (*ProjectCommand, error) {
	// Create Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"project",
		cmds.WithShort("Get project information"),
		cmds.WithLong(`
Get information about a GitHub Project v2 by owner and number.

Examples:
  github-graphql-cli project --owner=myorg --number=5
  github-graphql-cli project --owner=myorg --number=5 --output=json
  github-graphql-cli project --owner=myorg --number=5 --fields=title,total_items
		`),
		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"owner",
				parameters.ParameterTypeString,
				parameters.WithHelp("Organization or user name that owns the project"),
				parameters.WithDefault(githubConfig.Owner),
			),
			parameters.NewParameterDefinition(
				"number",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Project number"),
				parameters.WithDefault(githubConfig.ProjectNumber),
			),
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

	return &ProjectCommand{
		CommandDescription: cmdDesc,
	}, nil
}
