package main

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

// GetExecutionCommand gets details of a specific workflow execution
type GetExecutionCommand struct {
	*cmds.CommandDescription
}

// Settings for GetExecutionCommand
type GetExecutionSettings struct {
	ExecutionID string `glazed.parameter:"execution-id"`
}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *GetExecutionCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &GetExecutionSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Get API settings
	apiSettings, err := GetN8NAPISettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	// Create API client
	client := NewN8NClient(apiSettings.BaseURL, apiSettings.APIKey)

	// Get execution details
	execution, err := client.GetExecution(s.ExecutionID)
	if err != nil {
		return err
	}

	// Output as row
	row := types.NewRowFromMap(execution)
	return gp.AddRow(ctx, row)
}

// Ensure the interface is implemented
var _ cmds.GlazeCommand = &GetExecutionCommand{}

// NewGetExecutionCommand creates a new GetExecutionCommand
func NewGetExecutionCommand() (*GetExecutionCommand, error) {
	// Create the standard Glazed output layer
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Add the n8n API layer
	apiLayer, err := NewN8NAPILayer()
	if err != nil {
		return nil, err
	}

	// Create the command description
	cmdDesc := cmds.NewCommandDescription(
		"get-execution",
		cmds.WithShort("Get execution details"),
		cmds.WithLong("Get details of a specific workflow execution in the n8n instance."),

		// Define flags (parameters)
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"execution-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("ID of the execution to retrieve"),
				parameters.WithRequired(true),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(glazedLayer, apiLayer),
	)

	// Return the command instance
	return &GetExecutionCommand{
		CommandDescription: cmdDesc,
	}, nil
}
