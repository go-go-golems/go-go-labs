package main

import (
	"context"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/github-projects/pkg/github"
)

// ViewerCommand shows current user info
type ViewerCommand struct {
	*cmds.CommandDescription
}

// ViewerSettings holds the command settings
type ViewerSettings struct {
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &ViewerCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *ViewerCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	startTime := time.Now()

	// Parse settings
	s := &ViewerSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Use global logger for debugging
	logger := log.Logger

	logger.Debug().
		Str("function", "RunIntoGlazeProcessor").
		Time("start_time", startTime).
		Msg("entering viewer command execution")

	defer func() {
		duration := time.Since(startTime)
		logger.Debug().
			Str("function", "RunIntoGlazeProcessor").
			Dur("execution_time", duration).
			Msg("exiting viewer command execution")
	}()

	// Log parsed settings
	logger.Debug().
		Interface("settings", s).
		Msg("parsed viewer command settings")

	// Create GitHub client
	logger.Debug().Msg("creating GitHub client")
	clientStartTime := time.Now()
	client, err := github.NewClient()
	clientDuration := time.Since(clientStartTime)

	if err != nil {
		logger.Error().
			Err(err).
			Dur("client_creation_time", clientDuration).
			Msg("failed to create GitHub client")
		return errors.Wrap(err, "failed to create GitHub client")
	}

	logger.Debug().
		Dur("client_creation_time", clientDuration).
		Msg("successfully created GitHub client")

	// Get viewer information
	logger.Debug().Msg("requesting viewer information from GitHub API")
	viewerStartTime := time.Now()
	viewer, err := client.GetViewer(ctx)
	viewerDuration := time.Since(viewerStartTime)

	if err != nil {
		logger.Error().
			Err(err).
			Dur("api_call_duration", viewerDuration).
			Msg("failed to get viewer information from GitHub API")
		return errors.Wrap(err, "failed to get viewer")
	}

	logger.Debug().
		Dur("api_call_duration", viewerDuration).
		Str("viewer_login", viewer.Login).
		Str("viewer_name", viewer.Name).
		Str("viewer_email", viewer.Email).
		Msg("successfully retrieved viewer information from GitHub API")

	// Create row from viewer data
	logger.Debug().Msg("creating output row from viewer data")
	rowFields := map[string]interface{}{
		"login": viewer.Login,
		"name":  viewer.Name,
		"email": viewer.Email,
	}

	logger.Debug().
		Interface("row_fields", rowFields).
		Msg("preparing row data for output")

	row := types.NewRow(
		types.MRP("login", viewer.Login),
		types.MRP("name", viewer.Name),
		types.MRP("email", viewer.Email),
	)

	logger.Debug().Msg("adding row to glazed processor")
	if err := gp.AddRow(ctx, row); err != nil {
		logger.Error().
			Err(err).
			Interface("row_data", rowFields).
			Msg("failed to add row to glazed processor")
		return errors.Wrap(err, "failed to add row to processor")
	}

	logger.Debug().
		Interface("row_data", rowFields).
		Msg("successfully added row to glazed processor")

	return nil
}

// NewViewerCommand creates a new viewer command
func NewViewerCommand() (*ViewerCommand, error) {
	startTime := time.Now()

	// Use a basic logger for initialization logging since we don't have access to configured logger yet
	logger := log.Logger

	logger.Debug().
		Str("function", "NewViewerCommand").
		Time("start_time", startTime).
		Msg("entering viewer command initialization")

	// Create Glazed layer for output formatting
	logger.Debug().Msg("creating glazed parameter layers")
	glazedLayerStartTime := time.Now()
	glazedLayer, err := settings.NewGlazedParameterLayers()
	glazedLayerDuration := time.Since(glazedLayerStartTime)

	if err != nil {
		logger.Error().
			Err(err).
			Dur("glazed_layer_creation_time", glazedLayerDuration).
			Msg("failed to create glazed parameter layers")
		return nil, err
	}

	logger.Debug().
		Dur("glazed_layer_creation_time", glazedLayerDuration).
		Msg("successfully created glazed parameter layers")

	// Create command description
	logger.Debug().Msg("creating command description")
	cmdDescStartTime := time.Now()

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
		// Add parameter layers
		cmds.WithLayersList(glazedLayer),
	)

	cmdDescDuration := time.Since(cmdDescStartTime)
	logger.Debug().
		Dur("command_description_creation_time", cmdDescDuration).
		Str("command_name", "viewer").
		Msg("successfully created command description")

	viewerCmd := &ViewerCommand{
		CommandDescription: cmdDesc,
	}

	logger.Debug().
		Interface("command_structure", map[string]interface{}{
			"name":        "viewer",
			"has_glazed":  glazedLayer != nil,
			"param_count": 0,
		}).
		Msg("successfully created viewer command")

	return viewerCmd, nil
}
