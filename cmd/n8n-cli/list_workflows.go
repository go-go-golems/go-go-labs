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

// ListWorkflowsCommand lists all workflows
type ListWorkflowsCommand struct {
	*cmds.CommandDescription
}

// Settings for ListWorkflowsCommand
type ListWorkflowsSettings struct {
	Active bool `glazed.parameter:"active"`
	Limit  int  `glazed.parameter:"limit"`
	Offset int  `glazed.parameter:"offset"`
}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *ListWorkflowsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &ListWorkflowsSettings{}
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

	// List workflows
	workflows, err := client.ListWorkflows(s.Active, s.Limit, s.Offset)
	if err != nil {
		return err
	}

	// Output as rows
	for _, workflow := range workflows {
		row := types.NewRowFromMap(workflow)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// Ensure the interface is implemented
var _ cmds.GlazeCommand = &ListWorkflowsCommand{}

// NewListWorkflowsCommand creates a new ListWorkflowsCommand
func NewListWorkflowsCommand() (*ListWorkflowsCommand, error) {
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
		"list-workflows",
		cmds.WithShort("List all workflows in the n8n instance"),
		cmds.WithLong("List all workflows from the n8n instance with optional filtering."),
		
		// Define flags (parameters)
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"active",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Filter by active workflows only"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of workflows to return"),
				parameters.WithDefault(50),
			),
			parameters.NewParameterDefinition(
				"offset",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Offset for pagination"),
				parameters.WithDefault(0),
			),
		),
		
		// Add parameter layers
		cmds.WithLayersList(glazedLayer, apiLayer),
	)
	


	// Return the command instance
	return &ListWorkflowsCommand{
		CommandDescription: cmdDesc,
	}, nil
	}