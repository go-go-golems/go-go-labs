package cmds

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

	"github.com/go-go-golems/go-go-labs/cmd/github-projects/config"
	"github.com/go-go-golems/go-go-labs/pkg/github"
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
	start := time.Now()

	// Parse settings
	s := &ProjectSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Create contextual logger
	logger := log.With().
		Str("function", "RunIntoGlazeProcessor").
		Str("owner", s.Owner).
		Int("number", s.Number).
		Logger()

	logger.Debug().Msg("starting project command")

	// Validate parameters
	if s.Owner == "" {
		return errors.New("owner parameter is required")
	}
	if s.Number <= 0 {
		return errors.New("project number must be positive")
	}

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

	// Get project
	apiStart := time.Now()
	logger.Debug().Msg("fetching project from GitHub API")

	project, err := client.GetProject(ctx, s.Owner, s.Number)
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(apiStart)).
			Msg("failed to get project")
		return errors.Wrap(err, "failed to get project")
	}

	logger.Debug().
		Str("project_id", project.ID).
		Str("project_title", project.Title).
		Int("total_items", project.Items.TotalCount).
		Dur("duration", time.Since(apiStart)).
		Msg("project retrieved")

	// Create row from project data
	row := types.NewRow(
		types.MRP("id", project.ID),
		types.MRP("title", project.Title),
		types.MRP("public", project.Public),
		types.MRP("short_description", project.ShortDescription),
		types.MRP("closed", project.Closed),
		types.MRP("total_items", project.Items.TotalCount),
	)

	// Add row to processor
	if err := gp.AddRow(ctx, row); err != nil {
		logger.Error().
			Err(err).
			Msg("failed to add row to processor")
		return errors.Wrap(err, "failed to add row to processor")
	}

	logger.Debug().
		Dur("total_duration", time.Since(start)).
		Msg("project command completed")

	return nil
}

// NewProjectCommand creates a new project command
func NewProjectCommand() (*ProjectCommand, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "NewProjectCommand").
		Logger()

	logger.Trace().Msg("creating project command")

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
				parameters.WithDefault(config.GetDefaultOwner()),
			),
			parameters.NewParameterDefinition(
				"number",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Project number"),
				parameters.WithDefault(config.GetDefaultProjectNumber()),
			),
		),
		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	command := &ProjectCommand{
		CommandDescription: cmdDesc,
	}

	logger.Trace().
		Dur("total_duration", time.Since(start)).
		Msg("project command created")

	return command, nil
}
