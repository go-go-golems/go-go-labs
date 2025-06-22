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

// ProjectInfoCommand shows detailed project information including fields and labels
type ProjectInfoCommand struct {
	*cmds.CommandDescription
}

// ProjectInfoSettings holds the command settings
type ProjectInfoSettings struct {
	Owner  string `glazed.parameter:"owner"`
	Number int    `glazed.parameter:"number"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &ProjectInfoCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *ProjectInfoCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	start := time.Now()

	// Parse settings
	s := &ProjectInfoSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Create contextual logger
	logger := log.With().
		Str("function", "RunIntoGlazeProcessor").
		Str("owner", s.Owner).
		Int("number", s.Number).
		Logger()

	logger.Debug().Msg("starting project info command")

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

	// Get project fields
	fieldsStart := time.Now()
	logger.Debug().Msg("fetching project fields from GitHub API")

	fields, err := client.GetProjectFields(ctx, project.ID)
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(fieldsStart)).
			Msg("failed to get project fields")
		return errors.Wrap(err, "failed to get project fields")
	}

	logger.Debug().
		Int("field_count", len(fields)).
		Dur("duration", time.Since(fieldsStart)).
		Msg("project fields retrieved")

	// Create basic project info row
	row := types.NewRow(
		types.MRP("id", project.ID),
		types.MRP("title", project.Title),
		types.MRP("public", project.Public),
		types.MRP("short_description", project.ShortDescription),
		types.MRP("closed", project.Closed),
		types.MRP("total_items", project.Items.TotalCount),
		types.MRP("total_fields", len(fields)),
	)

	// Add field information
	fieldNames := make([]string, 0, len(fields))
	fieldTypes := make([]string, 0, len(fields))
	fieldOptionsCount := make([]int, 0, len(fields))

	for _, field := range fields {
		fieldNames = append(fieldNames, field.Name)
		fieldTypes = append(fieldTypes, field.Typename)
		fieldOptionsCount = append(fieldOptionsCount, len(field.Options))
	}

	row.Set("field_names", fieldNames)
	row.Set("field_types", fieldTypes)
	row.Set("field_options_count", fieldOptionsCount)

	// Add detailed field options information
	fieldOptionsInfo := make(map[string][]string)
	for _, field := range fields {
		if len(field.Options) > 0 {
			optionNames := make([]string, 0, len(field.Options))
			for _, option := range field.Options {
				optionNames = append(optionNames, option.Name)
			}
			fieldOptionsInfo[field.Name] = optionNames
		}
	}

	row.Set("field_options", fieldOptionsInfo)

	// Add row to processor
	if err := gp.AddRow(ctx, row); err != nil {
		logger.Error().
			Err(err).
			Msg("failed to add row to processor")
		return errors.Wrap(err, "failed to add row to processor")
	}

	logger.Debug().
		Dur("total_duration", time.Since(start)).
		Msg("project info command completed")

	return nil
}

// NewProjectInfoCommand creates a new project info command
func NewProjectInfoCommand() (*ProjectInfoCommand, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "NewProjectInfoCommand").
		Logger()

	logger.Trace().Msg("creating project info command")

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
		"project-info",
		cmds.WithShort("Get detailed project information including fields and labels"),
		cmds.WithLong(`
Get detailed information about a GitHub Project v2 including all defined fields and their options.

This command provides comprehensive project metadata including:
- Basic project information (title, description, status)
- All project fields and their types
- Field options for single-select fields
- Total item count

Examples:
  github-projects project-info --owner=myorg --number=5
  github-projects project-info --owner=myorg --number=5 --output=json
  github-projects project-info --owner=myorg --number=5 --fields=title,field_names
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

	command := &ProjectInfoCommand{
		CommandDescription: cmdDesc,
	}

	logger.Trace().
		Dur("total_duration", time.Since(start)).
		Msg("project info command created")

	return command, nil
}
