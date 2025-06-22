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

// FieldsCommand lists project fields
type FieldsCommand struct {
	*cmds.CommandDescription
}

// FieldsSettings holds the command settings
type FieldsSettings struct {
	Owner  string `glazed.parameter:"owner"`
	Number int    `glazed.parameter:"number"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &FieldsCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *FieldsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	start := time.Now()

	// Parse settings
	s := &FieldsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Create contextual logger
	logger := log.With().
		Str("function", "RunIntoGlazeProcessor").
		Str("owner", s.Owner).
		Int("number", s.Number).
		Logger()

	logger.Debug().Msg("starting fields command")

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
	projectStart := time.Now()
	logger.Debug().Msg("fetching project from GitHub API")
	project, err := client.GetProject(ctx, s.Owner, s.Number)
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(projectStart)).
			Msg("failed to get project")
		return errors.Wrap(err, "failed to get project")
	}

	projectLogger := logger.With().
		Str("project_id", project.ID).
		Str("project_title", project.Title).
		Logger()

	projectLogger.Debug().
		Dur("duration", time.Since(projectStart)).
		Msg("project fetched")

	// Get project fields
	fieldsStart := time.Now()
	projectLogger.Debug().Msg("fetching project fields from GitHub API")
	fields, err := client.GetProjectFields(ctx, project.ID)
	if err != nil {
		projectLogger.Error().
			Err(err).
			Dur("duration", time.Since(fieldsStart)).
			Msg("failed to get project fields")
		return errors.Wrap(err, "failed to get project fields")
	}
	projectLogger.Debug().
		Int("field_count", len(fields)).
		Dur("duration", time.Since(fieldsStart)).
		Msg("project fields fetched")

	// Create rows for each field
	processingStart := time.Now()
	projectLogger.Debug().
		Int("field_count", len(fields)).
		Msg("processing fields")

	for i, field := range fields {
		fieldStart := time.Now()
		fieldLogger := projectLogger.With().
			Int("field_index", i).
			Str("field_id", field.ID).
			Str("field_name", field.Name).
			Str("field_type", field.Typename).
			Logger()

		fieldLogger.Trace().Msg("processing field")

		row := types.NewRow(
			types.MRP("id", field.ID),
			types.MRP("name", field.Name),
			types.MRP("type", field.Typename),
		)

		// Add options if it's a single-select field
		if field.Typename == "ProjectV2SingleSelectField" && len(field.Options) > 0 {
			fieldLogger.Trace().
				Int("option_count", len(field.Options)).
				Msg("processing single-select field options")

			var optionNames []string
			var optionIDs []string
			for _, option := range field.Options {
				optionNames = append(optionNames, option.Name)
				optionIDs = append(optionIDs, option.ID)
			}
			row.Set("option_names", optionNames)
			row.Set("option_ids", optionIDs)

			fieldLogger.Trace().
				Int("option_count", len(field.Options)).
				Msg("single-select field options processed")
		}

		// Add iterations if it's an iteration field
		if field.Typename == "ProjectV2IterationField" && field.Configuration != nil && len(field.Configuration.Iterations) > 0 {
			fieldLogger.Trace().
				Int("iteration_count", len(field.Configuration.Iterations)).
				Msg("processing iteration field iterations")

			var iterationTitles []string
			var iterationIDs []string
			var iterationDates []string
			for _, iteration := range field.Configuration.Iterations {
				iterationTitles = append(iterationTitles, iteration.Title)
				iterationIDs = append(iterationIDs, iteration.ID)
				iterationDates = append(iterationDates, iteration.StartDate)
			}
			row.Set("iteration_titles", iterationTitles)
			row.Set("iteration_ids", iterationIDs)
			row.Set("iteration_dates", iterationDates)

			fieldLogger.Trace().
				Int("iteration_count", len(field.Configuration.Iterations)).
				Msg("iteration field iterations processed")
		}

		// Add row to processor
		if err := gp.AddRow(ctx, row); err != nil {
			fieldLogger.Error().
				Err(err).
				Dur("duration", time.Since(fieldStart)).
				Msg("failed to add row to processor")
			return err
		}

		fieldLogger.Trace().
			Dur("duration", time.Since(fieldStart)).
			Msg("field processed")
	}

	projectLogger.Debug().
		Int("field_count", len(fields)).
		Dur("processing_duration", time.Since(processingStart)).
		Dur("total_duration", time.Since(start)).
		Msg("fields command completed")

	return nil
}

// NewFieldsCommand creates a new fields command
func NewFieldsCommand() (*FieldsCommand, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "NewFieldsCommand").
		Logger()

	logger.Trace().Msg("creating fields command")

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
		"fields",
		cmds.WithShort("List project fields"),
		cmds.WithLong(`
List all fields for a GitHub Project v2, including their types and options.

Examples:
  github-graphql-cli fields --owner=myorg --number=5
  github-graphql-cli fields --owner=myorg --number=5 --output=json
  github-graphql-cli fields --owner=myorg --number=5 --fields=name,type
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

	logger.Trace().
		Dur("total_duration", time.Since(start)).
		Msg("fields command created")

	return &FieldsCommand{
		CommandDescription: cmdDesc,
	}, nil
}
