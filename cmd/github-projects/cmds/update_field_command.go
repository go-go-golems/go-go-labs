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
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &UpdateFieldCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *UpdateFieldCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	start := time.Now()

	// Parse settings
	s := &UpdateFieldSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Create contextual logger
	logger := log.With().
		Str("function", "RunIntoGlazeProcessor").
		Str("owner", s.Owner).
		Int("number", s.Number).
		Str("item_id", s.ItemID).
		Str("field_id", s.FieldID).
		Logger()

	logger.Debug().Msg("starting update field command")

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

	// Get project to get project ID
	projectStart := time.Now()
	logger.Debug().Msg("fetching project details")
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
		Msg("project details fetched")

	// Determine the field value to set based on which parameter was provided
	var fieldValue interface{}
	var valueType string

	if s.TextValue != "" {
		fieldValue = map[string]interface{}{"text": s.TextValue}
		valueType = "text"
		projectLogger.Trace().
			Str("value_type", valueType).
			Str("text_value", s.TextValue).
			Msg("selected text field value")
	} else if s.NumberValue != 0 {
		fieldValue = map[string]interface{}{"number": s.NumberValue}
		valueType = "number"
		projectLogger.Trace().
			Str("value_type", valueType).
			Float64("number_value", s.NumberValue).
			Msg("selected number field value")
	} else if s.DateValue != "" {
		fieldValue = map[string]interface{}{"date": s.DateValue}
		valueType = "date"
		projectLogger.Trace().
			Str("value_type", valueType).
			Str("date_value", s.DateValue).
			Msg("selected date field value")
	} else if s.SingleSelectOption != "" {
		fieldValue = map[string]interface{}{"singleSelectOptionId": s.SingleSelectOption}
		valueType = "single-select"
		projectLogger.Trace().
			Str("value_type", valueType).
			Str("single_select_option_id", s.SingleSelectOption).
			Msg("selected single-select field value")
	} else if s.IterationID != "" {
		fieldValue = map[string]interface{}{"iterationId": s.IterationID}
		valueType = "iteration"
		projectLogger.Trace().
			Str("value_type", valueType).
			Str("iteration_id", s.IterationID).
			Msg("selected iteration field value")
	} else {
		logger.Error().Msg("no field value provided")
		return errors.New("must provide one of: --text-value, --number-value, --date-value, --single-select-option, or --iteration-id")
	}

	projectLogger.Info().
		Str("valueType", valueType).
		Interface("value", fieldValue).
		Msg("updating field value")

	// Update the field value
	updateStart := time.Now()
	projectLogger.Debug().Msg("updating field value via GitHub API")
	err = client.UpdateFieldValue(ctx, project.ID, s.ItemID, s.FieldID, fieldValue)
	if err != nil {
		projectLogger.Error().
			Err(err).
			Interface("field_value", fieldValue).
			Dur("duration", time.Since(updateStart)).
			Msg("failed to update field value")
		return errors.Wrap(err, "failed to update field value")
	}
	projectLogger.Debug().
		Dur("duration", time.Since(updateStart)).
		Msg("field value updated successfully")

	// Return success
	row := types.NewRow(
		types.MRP("project_id", project.ID),
		types.MRP("item_id", s.ItemID),
		types.MRP("field_id", s.FieldID),
		types.MRP("value_type", valueType),
		types.MRP("status", "updated"),
	)

	logger.Debug().
		Dur("total_duration", time.Since(start)).
		Msg("update field command completed")

	return gp.AddRow(ctx, row)
}

// NewUpdateFieldCommand creates a new update-field command
func NewUpdateFieldCommand() (*UpdateFieldCommand, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "NewUpdateFieldCommand").
		Logger()

	logger.Trace().Msg("creating update-field command")

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
				parameters.WithDefault(config.GetDefaultOwner()),
			),
			parameters.NewParameterDefinition(
				"number",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Project number"),
				parameters.WithDefault(config.GetDefaultProjectNumber()),
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
		),
		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	// Create command instance
	cmd := &UpdateFieldCommand{
		CommandDescription: cmdDesc,
	}

	logger.Trace().
		Dur("total_duration", time.Since(start)).
		Msg("update-field command created")

	return cmd, nil
}
