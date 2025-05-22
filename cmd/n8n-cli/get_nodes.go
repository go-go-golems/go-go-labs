package main

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

// GetNodesCommand lists available node types in n8n
type GetNodesCommand struct {
	*cmds.CommandDescription
}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *GetNodesCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Get API settings
	apiSettings, err := GetN8NAPISettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	// Create API client
	client := NewN8NClient(apiSettings.BaseURL, apiSettings.APIKey)

	// Get available nodes
	nodes, err := client.GetNodes()
	if err != nil {
		return err
	}

	// Output as rows
	for _, node := range nodes {
		row := types.NewRowFromMap(node)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// Ensure the interface is implemented
var _ cmds.GlazeCommand = &GetNodesCommand{}

// NewGetNodesCommand creates a new GetNodesCommand
func NewGetNodesCommand() (*GetNodesCommand, error) {
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
		"get-nodes",
		cmds.WithShort("Get available node types"),
		cmds.WithLong("List all available node types in the n8n instance."),

		// Add parameter layers
		cmds.WithLayersList(glazedLayer, apiLayer),
	)

	// Return the command instance
	return &GetNodesCommand{
		CommandDescription: cmdDesc,
	}, nil
}