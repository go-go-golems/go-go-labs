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

// FieldsCommand lists project fields
type FieldsCommand struct {
	*cmds.CommandDescription
}

// FieldsSettings holds the command settings
type FieldsSettings struct {
	Owner    string `glazed.parameter:"owner"`
	Number   int    `glazed.parameter:"number"`
	LogLevel string `glazed.parameter:"log-level"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &FieldsCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *FieldsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings
	s := &FieldsSettings{}
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

	// Get project fields
	fields, err := client.GetProjectFields(ctx, project.ID)
	if err != nil {
		return errors.Wrap(err, "failed to get project fields")
	}

	// Create rows for each field
	for _, field := range fields {
		row := types.NewRow(
			types.MRP("id", field.ID),
			types.MRP("name", field.Name),
			types.MRP("type", field.Typename),
		)

		// Add options if it's a single-select field
		if field.Typename == "ProjectV2SingleSelectField" && len(field.Options) > 0 {
			var optionNames []string
			var optionIDs []string
			for _, option := range field.Options {
				optionNames = append(optionNames, option.Name)
				optionIDs = append(optionIDs, option.ID)
			}
			row.Set("option_names", optionNames)
			row.Set("option_ids", optionIDs)
		}

		// Add iterations if it's an iteration field
		if field.Typename == "ProjectV2IterationField" && field.Configuration != nil && len(field.Configuration.Iterations) > 0 {
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
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// NewFieldsCommand creates a new fields command
func NewFieldsCommand() (*FieldsCommand, error) {
	// Create Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

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
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"number",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Project number"),
				parameters.WithRequired(true),
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

	return &FieldsCommand{
		CommandDescription: cmdDesc,
	}, nil
}
