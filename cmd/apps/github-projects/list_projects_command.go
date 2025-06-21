package main

import (
	"context"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/github-projects/pkg/github"
)

// ListProjectsCommand lists projects for a user or organization
type ListProjectsCommand struct {
	*cmds.CommandDescription
}

// ListProjectsSettings holds the command settings
type ListProjectsSettings struct {
	Owner *string `glazed.parameter:"owner"`
	First int     `glazed.parameter:"first"`
	After *string `glazed.parameter:"after"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &ListProjectsCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *ListProjectsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	start := time.Now()

	// Parse settings
	s := &ListProjectsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Create contextual logger
	logger := log.With().
		Str("function", "RunIntoGlazeProcessor").
		Str("owner", func() string {
			if s.Owner != nil {
				return *s.Owner
			}
			return "nil"
		}()).
		Int("first", s.First).
		Logger()

	logger.Debug().Msg("starting list projects command")

	// Create GitHub client
	clientStart := time.Now()
	client, err := github.NewClient()
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(clientStart)).
			Msg("failed to create GitHub client")
		return errors.Wrap(err, "failed to create GitHub client")
	}

	logger.Trace().
		Dur("duration", time.Since(clientStart)).
		Msg("GitHub client created")

	// Get projects
	var response *github.ListProjectsResponse
	apiStart := time.Now()

	if s.Owner != nil {
		logger.Debug().
			Str("api_type", "organization_projects").
			Msg("fetching organization projects from GitHub API")

		// List organization projects
		response, err = client.ListOrganizationProjects(ctx, *s.Owner, s.First, s.After)
	} else {
		logger.Debug().
			Str("api_type", "user_projects").
			Msg("fetching user projects from GitHub API")

		// List user projects
		response, err = client.ListUserProjects(ctx, s.First, s.After)
	}

	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(apiStart)).
			Msg("GitHub API call failed")
		return errors.Wrap(err, "failed to list projects")
	}

	logger.Debug().
		Dur("duration", time.Since(apiStart)).
		Int("projects_count", len(response.Projects)).
		Bool("has_next_page", response.HasNextPage).
		Msg("GitHub API call completed")

	// Process projects into rows
	processStart := time.Now()
	logger.Debug().
		Int("projects_to_process", len(response.Projects)).
		Msg("processing projects")

	processedRows := 0
	for i, project := range response.Projects {
		projectLogger := logger.With().
			Int("project_index", i).
			Str("project_id", project.ID).
			Int("project_number", project.Number).
			Logger()

		projectLogger.Trace().
			Str("project_title", project.Title).
			Bool("project_public", project.Public).
			Bool("project_closed", project.Closed).
			Msg("processing project")

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
			projectLogger.Error().
				Err(err).
				Msg("failed to add row to processor")
			return errors.Wrap(err, "failed to add row")
		}

		processedRows++
		projectLogger.Trace().Msg("project processed")
	}

	logger.Debug().
		Dur("processing_duration", time.Since(processStart)).
		Int("total_rows_processed", processedRows).
		Dur("total_duration", time.Since(start)).
		Bool("has_next_page", response.HasNextPage).
		Msg("list projects command completed")

	return nil
}

// NewListProjectsCommand creates a new list projects command
func NewListProjectsCommand() (*ListProjectsCommand, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "NewListProjectsCommand").
		Logger()

	logger.Trace().Msg("creating list projects command")

	// Create Glazed layer for output formatting
	glazedStart := time.Now()
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(glazedStart)).
			Msg("failed to create glazed parameter layers")
		return nil, err
	}

	logger.Trace().
		Dur("duration", time.Since(glazedStart)).
		Msg("glazed parameter layers created")

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
				parameters.WithDefault(GetDefaultOwner()),
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
		),
		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	command := &ListProjectsCommand{
		CommandDescription: cmdDesc,
	}

	logger.Trace().
		Dur("total_duration", time.Since(start)).
		Msg("list projects command created")

	return command, nil
}
