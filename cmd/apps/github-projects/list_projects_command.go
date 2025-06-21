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
	startTime := time.Now()

	// Parse settings
	s := &ListProjectsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	log.Debug().
		Str("function", "RunIntoGlazeProcessor").
		Str("event", "entry").
		Str("owner", func() string {
			if s.Owner != nil {
				return *s.Owner
			}
			return "nil"
		}()).
		Int("first", s.First).
		Str("after", func() string {
			if s.After != nil {
				return *s.After
			}
			return "nil"
		}()).
		Msg("Starting list projects command")

	// Create GitHub client
	clientStart := time.Now()
	log.Debug().
		Str("event", "client_creation_start").
		Msg("Creating GitHub client")

	client, err := github.NewClient()
	if err != nil {
		log.Error().
			Err(err).
			Dur("duration", time.Since(clientStart)).
			Str("event", "client_creation_failed").
			Msg("Failed to create GitHub client")
		return errors.Wrap(err, "failed to create GitHub client")
	}

	log.Debug().
		Dur("duration", time.Since(clientStart)).
		Str("event", "client_creation_success").
		Msg("GitHub client created successfully")

	// Get projects
	var response *github.ListProjectsResponse
	apiStart := time.Now()

	if s.Owner != nil {
		log.Debug().
			Str("event", "api_call_start").
			Str("api_type", "organization_projects").
			Str("owner", *s.Owner).
			Int("first", s.First).
			Str("after", func() string {
				if s.After != nil {
					return *s.After
				}
				return "nil"
			}()).
			Msg("Starting GitHub API call for organization projects")

		// List organization projects
		response, err = client.ListOrganizationProjects(ctx, *s.Owner, s.First, s.After)
	} else {
		log.Debug().
			Str("event", "api_call_start").
			Str("api_type", "user_projects").
			Int("first", s.First).
			Str("after", func() string {
				if s.After != nil {
					return *s.After
				}
				return "nil"
			}()).
			Msg("Starting GitHub API call for user projects")

		// List user projects
		response, err = client.ListUserProjects(ctx, s.First, s.After)
	}

	if err != nil {
		log.Error().
			Err(err).
			Dur("api_duration", time.Since(apiStart)).
			Str("event", "api_call_failed").
			Msg("GitHub API call failed")
		return errors.Wrap(err, "failed to list projects")
	}

	log.Debug().
		Dur("api_duration", time.Since(apiStart)).
		Int("projects_count", len(response.Projects)).
		Bool("has_next_page", response.HasNextPage).
		Str("end_cursor", func() string {
			if response.EndCursor != nil {
				return *response.EndCursor
			}
			return "nil"
		}()).
		Str("event", "api_call_success").
		Msg("GitHub API call completed successfully")

	// Process projects into rows
	rowProcessingStart := time.Now()
	log.Debug().
		Str("event", "row_processing_start").
		Int("projects_to_process", len(response.Projects)).
		Msg("Starting to process projects into rows")

	processedRows := 0
	for i, project := range response.Projects {
		log.Debug().
			Str("event", "project_processing").
			Int("project_index", i).
			Str("project_id", project.ID).
			Int("project_number", project.Number).
			Str("project_title", project.Title).
			Bool("project_public", project.Public).
			Bool("project_closed", project.Closed).
			Str("project_url", project.URL).
			Msg("Processing project")

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
			log.Error().
				Err(err).
				Int("project_index", i).
				Str("project_id", project.ID).
				Str("event", "row_add_failed").
				Msg("Failed to add row to processor")
			return errors.Wrap(err, "failed to add row")
		}

		processedRows++
		log.Debug().
			Str("event", "project_processed").
			Int("project_index", i).
			Str("project_id", project.ID).
			Int("processed_rows", processedRows).
			Msg("Project processed successfully")
	}

	log.Debug().
		Dur("row_processing_duration", time.Since(rowProcessingStart)).
		Int("total_rows_processed", processedRows).
		Str("event", "row_processing_complete").
		Msg("All projects processed into rows")

	log.Debug().
		Str("function", "RunIntoGlazeProcessor").
		Str("event", "exit").
		Dur("total_duration", time.Since(startTime)).
		Int("total_projects", len(response.Projects)).
		Int("total_rows", processedRows).
		Bool("has_next_page", response.HasNextPage).
		Msg("List projects command completed successfully")

	return nil
}

// NewListProjectsCommand creates a new list projects command
func NewListProjectsCommand() (*ListProjectsCommand, error) {
	startTime := time.Now()

	log.Debug().
		Str("function", "NewListProjectsCommand").
		Str("event", "entry").
		Msg("Creating new list projects command")

	// Create Glazed layer for output formatting
	glazedLayerStart := time.Now()
	log.Debug().
		Str("event", "glazed_layer_creation_start").
		Msg("Creating glazed parameter layers")

	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		log.Error().
			Err(err).
			Dur("duration", time.Since(glazedLayerStart)).
			Str("event", "glazed_layer_creation_failed").
			Msg("Failed to create glazed parameter layers")
		return nil, err
	}

	log.Debug().
		Dur("duration", time.Since(glazedLayerStart)).
		Str("event", "glazed_layer_creation_success").
		Msg("Glazed parameter layers created successfully")

	// Create command description
	cmdDescStart := time.Now()
	log.Debug().
		Str("event", "command_description_creation_start").
		Msg("Creating command description")

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
				parameters.WithDefault(githubConfig.Owner),
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

	log.Debug().
		Dur("duration", time.Since(cmdDescStart)).
		Str("event", "command_description_creation_success").
		Msg("Command description created successfully")

	command := &ListProjectsCommand{
		CommandDescription: cmdDesc,
	}

	log.Debug().
		Str("function", "NewListProjectsCommand").
		Str("event", "exit").
		Dur("total_duration", time.Since(startTime)).
		Msg("List projects command created successfully")

	return command, nil
}
