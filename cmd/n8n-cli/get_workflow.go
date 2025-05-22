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

// GetWorkflowCommand gets a workflow by ID
type GetWorkflowCommand struct {
	*cmds.CommandDescription
}

// Settings for GetWorkflowCommand
type GetWorkflowSettings struct {
	ID         string `glazed.parameter:"id"`
	SaveToFile string `glazed.parameter:"save-to-file"`
}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *GetWorkflowCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &GetWorkflowSettings{}
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

	// Get workflow
	workflow, err := client.GetWorkflow(s.ID)
	if err != nil {
		return err
	}

	// Write to output file if specified
	if s.SaveToFile != "" {
		if err := WriteJSONFile(s.SaveToFile, workflow); err != nil {
			return err
		}
		fmt.Printf("Workflow written to %s\n", s.SaveToFile)
	}

	// Output as row
	row := types.NewRowFromMap(workflow)
	return gp.AddRow(ctx, row)
}

// Ensure the interface is implemented
var _ cmds.GlazeCommand = &GetWorkflowCommand{}

// NewGetWorkflowCommand creates a new GetWorkflowCommand
func NewGetWorkflowCommand() (*GetWorkflowCommand, error) {
	// Create the standard Glazed output layer
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	apiLayer, err := NewN8NAPILayer()
	if err != nil {
		return nil, err
	}

	// Create the command description
	cmdDesc := cmds.NewCommandDescription(
		"get-workflow",
		cmds.WithShort("Get a workflow by ID"),
		cmds.WithLong("Get a workflow by ID from the n8n instance."),

		// Define flags (parameters)
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"id",
				parameters.ParameterTypeString,
				parameters.WithHelp("Workflow ID"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"save-to-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Save workflow to JSON file (optional)"),
				parameters.WithDefault(""),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(glazedLayer, apiLayer),
	)

	// Return the command instance
	return &GetWorkflowCommand{
		CommandDescription: cmdDesc,
	}, nil
}
