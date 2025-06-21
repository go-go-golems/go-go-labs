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
	start := time.Now()

	// Parse settings
	s := &ViewerSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Create contextual logger
	logger := log.With().
		Str("function", "RunIntoGlazeProcessor").
		Logger()

	logger.Debug().Msg("starting viewer command")

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

	// Get viewer information
	viewerStart := time.Now()
	logger.Debug().Msg("fetching viewer information from GitHub API")
	viewer, err := client.GetViewer(ctx)
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(viewerStart)).
			Msg("failed to get viewer information")
		return errors.Wrap(err, "failed to get viewer")
	}

	logger.Debug().
		Dur("duration", time.Since(viewerStart)).
		Str("viewer_login", viewer.Login).
		Str("viewer_name", viewer.Name).
		Msg("viewer information retrieved")

	// Create row from viewer data
	row := types.NewRow(
		types.MRP("login", viewer.Login),
		types.MRP("name", viewer.Name),
		types.MRP("email", viewer.Email),
	)

	if err := gp.AddRow(ctx, row); err != nil {
		logger.Error().
			Err(err).
			Msg("failed to add row to processor")
		return errors.Wrap(err, "failed to add row to processor")
	}

	logger.Debug().
		Dur("total_duration", time.Since(start)).
		Msg("viewer command completed")

	return nil
}

// NewViewerCommand creates a new viewer command
func NewViewerCommand() (*ViewerCommand, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "NewViewerCommand").
		Logger()

	logger.Trace().Msg("creating viewer command")

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

	viewerCmd := &ViewerCommand{
		CommandDescription: cmdDesc,
	}

	logger.Trace().
		Dur("total_duration", time.Since(start)).
		Msg("viewer command created")

	return viewerCmd, nil
}
