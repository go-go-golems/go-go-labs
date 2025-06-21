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

// ProjectCommand shows project information
type ProjectCommand struct {
	*cmds.CommandDescription
}

// ProjectSettings holds the command settings
type ProjectSettings struct {
	Owner  string `glazed.parameter:"owner"`
	Number int    `glazed.parameter:"number"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &ProjectCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *ProjectCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	startTime := time.Now()

	// Parse settings
	s := &ProjectSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	log.Debug().
		Str("function", "RunIntoGlazeProcessor").
		Str("owner", s.Owner).
		Int("number", s.Number).
		Msg("function entry - processing project command")

	defer func() {
		duration := time.Since(startTime)
		log.Debug().
			Str("function", "RunIntoGlazeProcessor").
			Dur("duration", duration).
			Msg("function exit - completed project command processing")
	}()

	// Log parameter validation
	if s.Owner == "" {
		log.Debug().Msg("owner parameter is empty")
		return errors.New("owner parameter is required")
	}
	if s.Number <= 0 {
		log.Debug().Int("number", s.Number).Msg("invalid project number")
		return errors.New("project number must be positive")
	}

	// Create GitHub client
	clientStartTime := time.Now()
	log.Debug().Msg("creating GitHub client")

	client, err := github.NewClient()
	if err != nil {
		log.Debug().
			Err(err).
			Dur("client_creation_duration", time.Since(clientStartTime)).
			Msg("failed to create GitHub client")
		return errors.Wrap(err, "failed to create GitHub client")
	}

	log.Debug().
		Dur("client_creation_duration", time.Since(clientStartTime)).
		Msg("GitHub client created successfully")

	// Get project
	apiStartTime := time.Now()
	log.Debug().
		Str("owner", s.Owner).
		Int("number", s.Number).
		Msg("initiating GitHub API call to get project")

	project, err := client.GetProject(ctx, s.Owner, s.Number)
	apiDuration := time.Since(apiStartTime)

	if err != nil {
		log.Debug().
			Err(err).
			Str("owner", s.Owner).
			Int("number", s.Number).
			Dur("api_call_duration", apiDuration).
			Msg("GitHub API call failed")
		return errors.Wrap(err, "failed to get project")
	}

	log.Debug().
		Str("project_id", project.ID).
		Str("project_title", project.Title).
		Bool("project_public", project.Public).
		Str("project_short_description", project.ShortDescription).
		Bool("project_closed", project.Closed).
		Int("total_items", project.Items.TotalCount).
		Dur("api_call_duration", apiDuration).
		Msg("GitHub API call completed successfully")

	// Create row from project data
	rowStartTime := time.Now()
	log.Debug().Msg("creating output row from project data")

	row := types.NewRow(
		types.MRP("id", project.ID),
		types.MRP("title", project.Title),
		types.MRP("public", project.Public),
		types.MRP("short_description", project.ShortDescription),
		types.MRP("closed", project.Closed),
		types.MRP("total_items", project.Items.TotalCount),
	)

	log.Debug().
		Dur("row_creation_duration", time.Since(rowStartTime)).
		Msg("output row created successfully")

	// Add row to processor
	processorStartTime := time.Now()
	log.Debug().Msg("adding row to glazed processor")

	err = gp.AddRow(ctx, row)
	if err != nil {
		log.Debug().
			Err(err).
			Dur("processor_duration", time.Since(processorStartTime)).
			Msg("failed to add row to processor")
		return errors.Wrap(err, "failed to add row to processor")
	}

	log.Debug().
		Dur("processor_duration", time.Since(processorStartTime)).
		Msg("row added to processor successfully")

	return nil
}

// NewProjectCommand creates a new project command
func NewProjectCommand() (*ProjectCommand, error) {
	startTime := time.Now()

	log.Debug().
		Str("function", "NewProjectCommand").
		Msg("function entry - creating new project command")

	defer func() {
		duration := time.Since(startTime)
		log.Debug().
			Str("function", "NewProjectCommand").
			Dur("duration", duration).
			Msg("function exit - project command creation completed")
	}()

	// Create Glazed layer for output formatting
	layerStartTime := time.Now()
	log.Debug().Msg("creating glazed parameter layers")

	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		log.Debug().
			Err(err).
			Dur("layer_creation_duration", time.Since(layerStartTime)).
			Msg("failed to create glazed parameter layers")
		return nil, err
	}

	log.Debug().
		Dur("layer_creation_duration", time.Since(layerStartTime)).
		Msg("glazed parameter layers created successfully")

	// Create command description
	descStartTime := time.Now()
	log.Debug().Msg("creating command description with parameters")

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
		),
		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	log.Debug().
		Str("command_name", "project").
		Int("parameter_count", 2).
		Dur("description_creation_duration", time.Since(descStartTime)).
		Msg("command description created successfully")

	command := &ProjectCommand{
		CommandDescription: cmdDesc,
	}

	log.Debug().
		Str("command_type", "ProjectCommand").
		Msg("project command instance created successfully")

	return command, nil
}
