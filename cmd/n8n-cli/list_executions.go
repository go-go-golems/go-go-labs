package main

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

// ListExecutionsCommand lists workflow executions
type ListExecutionsCommand struct {
	*cmds.CommandDescription
}

// Settings for ListExecutionsCommand
type ListExecutionsSettings struct {
	WorkflowID string `glazed.parameter:"workflow-id"`
	Status     string `glazed.parameter:"status"`
	Limit      int    `glazed.parameter:"limit"`
	Offset     int    `glazed.parameter:"offset"`
}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *ListExecutionsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &ListExecutionsSettings{}
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

	// Build query parameters
	params := make(map[string]string)
	if s.WorkflowID != "" {
		params["workflowId"] = s.WorkflowID
	}
	if s.Status != "" {
		params["status"] = s.Status
	}
	if s.Limit > 0 {
		params["limit"] = fmt.Sprintf("%d", s.Limit)
	}
	if s.Offset > 0 {
		params["offset"] = fmt.Sprintf("%d", s.Offset)
	}

	// Get executions
	executions, err := client.ListExecutions(params)
	if err != nil {
		return err
	}

	// Output as rows
	for _, exec := range executions {
		row := types.NewRowFromMap(exec)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// Ensure the interface is implemented
var _ cmds.GlazeCommand = &ListExecutionsCommand{}

// NewListExecutionsCommand creates a new ListExecutionsCommand
func NewListExecutionsCommand() (*ListExecutionsCommand, error) {
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
		"list-executions",
		cmds.WithShort("List workflow executions"),
		cmds.WithLong("List executions of workflows in the n8n instance."),

		// Define flags (parameters)
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"workflow-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Filter by workflow ID"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"status",
				parameters.ParameterTypeString,
				parameters.WithHelp("Filter by status (success, error, waiting)"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of executions to return"),
				parameters.WithDefault(20),
			),
			parameters.NewParameterDefinition(
				"offset",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Number of executions to skip"),
				parameters.WithDefault(0),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(glazedLayer, apiLayer),
	)

	// Return the command instance
	return &ListExecutionsCommand{
		CommandDescription: cmdDesc,
	}, nil
}