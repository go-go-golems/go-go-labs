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

// ListProjectsCommand lists projects for a user or organization
type ListProjectsCommand struct {
	*cmds.CommandDescription
}

// ListProjectsSettings holds the command settings
type ListProjectsSettings struct {
	Owner    *string `glazed.parameter:"owner"`
	First    int     `glazed.parameter:"first"`
	After    *string `glazed.parameter:"after"`
	LogLevel string  `glazed.parameter:"log-level"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &ListProjectsCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *ListProjectsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings
	s := &ListProjectsSettings{}
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

	// Get projects
	var response *github.ListProjectsResponse
	if s.Owner != nil {
		// List organization projects
		response, err = client.ListOrganizationProjects(ctx, *s.Owner, s.First, s.After)
	} else {
		// List user projects
		response, err = client.ListUserProjects(ctx, s.First, s.After)
	}

	if err != nil {
		return errors.Wrap(err, "failed to list projects")
	}

	// Create rows from project data
	for _, project := range response.Projects {
		row := types.NewRow(
			types.MRP("id", project.ID),
			types.MRP("number", project.Number),
			types.MRP("title", project.Title),
			types.MRP("public", project.Public),
			types.MRP("closed", project.Closed),
			types.MRP("shortDescription", project.ShortDescription),
			types.MRP("url", project.URL),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return errors.Wrap(err, "failed to add row")
		}
	}

	return nil
}

// NewListProjectsCommand creates a new list projects command
func NewListProjectsCommand() (*ListProjectsCommand, error) {
	// Create Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"list-projects",
		cmds.WithShort("List GitHub projects"),
		cmds.WithLong(`
List GitHub Projects v2 for the authenticated user or for a specific organization.

If --owner is specified, lists projects for that organization.
If --owner is not specified, lists projects for the authenticated user.

Examples:
  github-graphql-cli list-projects
  github-graphql-cli list-projects --owner=myorg
  github-graphql-cli list-projects --first=10
  github-graphql-cli list-projects --owner=myorg --output=json
		`),
		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"owner",
				parameters.ParameterTypeString,
				parameters.WithHelp("Organization name to list projects for (if not specified, lists user projects)"),
			),
			parameters.NewParameterDefinition(
				"first",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Number of projects to fetch"),
				parameters.WithDefault(20),
			),
			parameters.NewParameterDefinition(
				"after",
				parameters.ParameterTypeString,
				parameters.WithHelp("Cursor for pagination"),
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

	return &ListProjectsCommand{
		CommandDescription: cmdDesc,
	}, nil
}
