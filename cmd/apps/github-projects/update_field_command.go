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

// UpdateFieldCommand updates a field value for a project item
type UpdateFieldCommand struct {
	*cmds.CommandDescription
}

// UpdateFieldSettings holds the command settings
type UpdateFieldSettings struct {
	Owner              string  `glazed.parameter:"owner"`
	Number             int     `glazed.parameter:"number"`
	ItemID             string  `glazed.parameter:"item-id"`
	FieldID            string  `glazed.parameter:"field-id"`
	TextValue          string  `glazed.parameter:"text-value"`
	NumberValue        float64 `glazed.parameter:"number-value"`
	DateValue          string  `glazed.parameter:"date-value"`
	SingleSelectOption string  `glazed.parameter:"single-select-option"`
	IterationID        string  `glazed.parameter:"iteration-id"`
	LogLevel           string  `glazed.parameter:"log-level"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &UpdateFieldCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *UpdateFieldCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings
	s := &UpdateFieldSettings{}
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

	// Get project to get project ID
	project, err := client.GetProject(ctx, s.Owner, s.Number)
	if err != nil {
		return errors.Wrap(err, "failed to get project")
	}

	// Determine the field value to set based on which parameter was provided
	var fieldValue interface{}
	var valueType string

	if s.TextValue != "" {
		fieldValue = map[string]interface{}{"text": s.TextValue}
		valueType = "text"
	} else if s.NumberValue != 0 {
		fieldValue = map[string]interface{}{"number": s.NumberValue}
		valueType = "number"
	} else if s.DateValue != "" {
		fieldValue = map[string]interface{}{"date": s.DateValue}
		valueType = "date"
	} else if s.SingleSelectOption != "" {
		fieldValue = map[string]interface{}{"singleSelectOptionId": s.SingleSelectOption}
		valueType = "single-select"
	} else if s.IterationID != "" {
		fieldValue = map[string]interface{}{"iterationId": s.IterationID}
		valueType = "iteration"
	} else {
		return errors.New("must provide one of: --text-value, --number-value, --date-value, --single-select-option, or --iteration-id")
	}

	logger.Info().
		Str("projectID", project.ID).
		Str("itemID", s.ItemID).
		Str("fieldID", s.FieldID).
		Str("valueType", valueType).
		Interface("value", fieldValue).
		Msg("Updating field value")

	// Update the field value
	err = client.UpdateFieldValue(ctx, project.ID, s.ItemID, s.FieldID, fieldValue)
	if err != nil {
		return errors.Wrap(err, "failed to update field value")
	}

	// Return success
	row := types.NewRow(
		types.MRP("project_id", project.ID),
		types.MRP("item_id", s.ItemID),
		types.MRP("field_id", s.FieldID),
		types.MRP("value_type", valueType),
		types.MRP("status", "updated"),
	)

	return gp.AddRow(ctx, row)
}

// NewUpdateFieldCommand creates a new update-field command
func NewUpdateFieldCommand() (*UpdateFieldCommand, error) {
	// Create Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"update-field",
		cmds.WithShort("Update a field value for a project item"),
		cmds.WithLong(`
Update a field value for a project item. You must specify exactly one value type.

Examples:
  # Update text field
  github-graphql-cli update-field --owner=myorg --number=5 --item-id=ITEM_ID --field-id=FIELD_ID --text-value="New text"
  
  # Update number field  
  github-graphql-cli update-field --owner=myorg --number=5 --item-id=ITEM_ID --field-id=FIELD_ID --number-value=42
  
  # Update date field
  github-graphql-cli update-field --owner=myorg --number=5 --item-id=ITEM_ID --field-id=FIELD_ID --date-value="2025-07-01"
  
  # Update single select field (use option ID)
  github-graphql-cli update-field --owner=myorg --number=5 --item-id=ITEM_ID --field-id=FIELD_ID --single-select-option="f75ad846"
  
  # Update iteration field 
  github-graphql-cli update-field --owner=myorg --number=5 --item-id=ITEM_ID --field-id=FIELD_ID --iteration-id="ITERATION_ID"
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
				"item-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Project item ID (from items command)"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"field-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Field ID (from fields command)"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"text-value",
				parameters.ParameterTypeString,
				parameters.WithHelp("Text value to set (for text fields)"),
			),
			parameters.NewParameterDefinition(
				"number-value",
				parameters.ParameterTypeFloat,
				parameters.WithHelp("Number value to set (for number fields)"),
			),
			parameters.NewParameterDefinition(
				"date-value",
				parameters.ParameterTypeString,
				parameters.WithHelp("Date value to set in YYYY-MM-DD format (for date fields)"),
			),
			parameters.NewParameterDefinition(
				"single-select-option",
				parameters.ParameterTypeString,
				parameters.WithHelp("Single select option ID to set (for single select fields)"),
			),
			parameters.NewParameterDefinition(
				"iteration-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Iteration ID to set (for iteration fields)"),
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

	return &UpdateFieldCommand{
		CommandDescription: cmdDesc,
	}, nil
}
