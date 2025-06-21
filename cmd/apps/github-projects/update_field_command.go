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
	startTime := time.Now()

	// Parse settings
	s := &UpdateFieldSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	log.Debug().
		Str("function", "RunIntoGlazeProcessor").
		Str("owner", s.Owner).
		Int("number", s.Number).
		Str("item_id", s.ItemID).
		Str("field_id", s.FieldID).
		Str("text_value", s.TextValue).
		Float64("number_value", s.NumberValue).
		Str("date_value", s.DateValue).
		Str("single_select_option", s.SingleSelectOption).
		Str("iteration_id", s.IterationID).
		Msg("Function entry - parameters parsed")

	// Create GitHub client
	log.Debug().Msg("Creating GitHub client")
	clientStartTime := time.Now()
	client, err := github.NewClient()
	if err != nil {
		log.Error().
			Err(err).
			Dur("duration", time.Since(clientStartTime)).
			Msg("Failed to create GitHub client")
		return errors.Wrap(err, "failed to create GitHub client")
	}
	log.Debug().
		Dur("duration", time.Since(clientStartTime)).
		Msg("GitHub client created successfully")

	// Get project to get project ID
	log.Debug().
		Str("owner", s.Owner).
		Int("number", s.Number).
		Msg("Fetching project details")
	projectStartTime := time.Now()
	project, err := client.GetProject(ctx, s.Owner, s.Number)
	if err != nil {
		log.Error().
			Err(err).
			Str("owner", s.Owner).
			Int("number", s.Number).
			Dur("duration", time.Since(projectStartTime)).
			Msg("Failed to get project")
		return errors.Wrap(err, "failed to get project")
	}
	log.Debug().
		Str("project_id", project.ID).
		Str("project_title", project.Title).
		Dur("duration", time.Since(projectStartTime)).
		Msg("Project details fetched successfully")

	// Determine the field value to set based on which parameter was provided
	log.Debug().Msg("Determining field value type")
	var fieldValue interface{}
	var valueType string

	if s.TextValue != "" {
		fieldValue = map[string]interface{}{"text": s.TextValue}
		valueType = "text"
		log.Debug().
			Str("value_type", valueType).
			Str("text_value", s.TextValue).
			Msg("Selected text field value")
	} else if s.NumberValue != 0 {
		fieldValue = map[string]interface{}{"number": s.NumberValue}
		valueType = "number"
		log.Debug().
			Str("value_type", valueType).
			Float64("number_value", s.NumberValue).
			Msg("Selected number field value")
	} else if s.DateValue != "" {
		fieldValue = map[string]interface{}{"date": s.DateValue}
		valueType = "date"
		log.Debug().
			Str("value_type", valueType).
			Str("date_value", s.DateValue).
			Msg("Selected date field value")
	} else if s.SingleSelectOption != "" {
		fieldValue = map[string]interface{}{"singleSelectOptionId": s.SingleSelectOption}
		valueType = "single-select"
		log.Debug().
			Str("value_type", valueType).
			Str("single_select_option_id", s.SingleSelectOption).
			Msg("Selected single-select field value")
	} else if s.IterationID != "" {
		fieldValue = map[string]interface{}{"iterationId": s.IterationID}
		valueType = "iteration"
		log.Debug().
			Str("value_type", valueType).
			Str("iteration_id", s.IterationID).
			Msg("Selected iteration field value")
	} else {
		log.Error().
			Bool("text_value_empty", s.TextValue == "").
			Bool("number_value_zero", s.NumberValue == 0).
			Bool("date_value_empty", s.DateValue == "").
			Bool("single_select_option_empty", s.SingleSelectOption == "").
			Bool("iteration_id_empty", s.IterationID == "").
			Msg("No field value provided")
		return errors.New("must provide one of: --text-value, --number-value, --date-value, --single-select-option, or --iteration-id")
	}

	log.Debug().
		Interface("field_value_structured", fieldValue).
		Msg("Field value structure prepared")

	log.Info().
		Str("projectID", project.ID).
		Str("itemID", s.ItemID).
		Str("fieldID", s.FieldID).
		Str("valueType", valueType).
		Interface("value", fieldValue).
		Msg("Updating field value")

	// Update the field value
	log.Debug().
		Str("project_id", project.ID).
		Str("item_id", s.ItemID).
		Str("field_id", s.FieldID).
		Msg("Initiating GitHub API call to update field value")
	updateStartTime := time.Now()
	err = client.UpdateFieldValue(ctx, project.ID, s.ItemID, s.FieldID, fieldValue)
	updateDuration := time.Since(updateStartTime)
	if err != nil {
		log.Error().
			Err(err).
			Str("project_id", project.ID).
			Str("item_id", s.ItemID).
			Str("field_id", s.FieldID).
			Interface("field_value", fieldValue).
			Dur("duration", updateDuration).
			Msg("Failed to update field value via GitHub API")
		return errors.Wrap(err, "failed to update field value")
	}
	log.Debug().
		Dur("api_call_duration", updateDuration).
		Msg("Field value updated successfully via GitHub API")

	// Return success
	log.Debug().Msg("Preparing success response row")
	row := types.NewRow(
		types.MRP("project_id", project.ID),
		types.MRP("item_id", s.ItemID),
		types.MRP("field_id", s.FieldID),
		types.MRP("value_type", valueType),
		types.MRP("status", "updated"),
	)

	log.Debug().
		Interface("response_row", row).
		Msg("Success response row created")

	totalDuration := time.Since(startTime)
	log.Debug().
		Str("function", "RunIntoGlazeProcessor").
		Dur("total_duration", totalDuration).
		Str("status", "success").
		Msg("Function exit - operation completed successfully")

	return gp.AddRow(ctx, row)
}

// NewUpdateFieldCommand creates a new update-field command
func NewUpdateFieldCommand() (*UpdateFieldCommand, error) {
	startTime := time.Now()

	log.Debug().
		Str("function", "NewUpdateFieldCommand").
		Msg("Function entry - initializing update-field command")

	// Create Glazed layer for output formatting
	log.Debug().Msg("Creating Glazed parameter layers")
	layerStartTime := time.Now()
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		log.Error().
			Err(err).
			Dur("duration", time.Since(layerStartTime)).
			Msg("Failed to create Glazed parameter layers")
		return nil, err
	}
	log.Debug().
		Dur("duration", time.Since(layerStartTime)).
		Msg("Glazed parameter layers created successfully")

	// Create command description
	log.Debug().Msg("Creating command description with parameters")
	cmdDescStartTime := time.Now()
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
				parameters.WithDefault(githubConfig.Owner),
			),
			parameters.NewParameterDefinition(
				"number",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Project number"),
				parameters.WithDefault(githubConfig.ProjectNumber),
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

	log.Debug().
		Dur("duration", time.Since(cmdDescStartTime)).
		Int("parameter_count", 9).
		Msg("Command description created with all parameters")

	// Create command instance
	log.Debug().Msg("Creating UpdateFieldCommand instance")
	cmd := &UpdateFieldCommand{
		CommandDescription: cmdDesc,
	}

	totalDuration := time.Since(startTime)
	log.Debug().
		Str("function", "NewUpdateFieldCommand").
		Dur("total_duration", totalDuration).
		Str("status", "success").
		Msg("Function exit - update-field command initialized successfully")

	return cmd, nil
}
