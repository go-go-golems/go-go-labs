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

// ConnectNodesCommand connects nodes in a workflow
type ConnectNodesCommand struct {
	*cmds.CommandDescription
}

// Settings for ConnectNodesCommand
type ConnectNodesSettings struct {
	WorkflowID string `glazed.parameter:"workflow-id"`
	SourceNode string `glazed.parameter:"source-node"`
	TargetNode string `glazed.parameter:"target-node"`
	OutputIndex int    `glazed.parameter:"output-index"`
	InputIndex  int    `glazed.parameter:"input-index"`
}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *ConnectNodesCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &ConnectNodesSettings{}
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

	// Get current workflow
	workflow, err := client.GetWorkflow(s.WorkflowID)
	if err != nil {
		return err
	}

	// Get or create connections map
	connections, ok := workflow["connections"].(map[string]interface{})
	if !ok {
		connections = make(map[string]interface{})
		workflow["connections"] = connections
	}

	// Get or create source node connections
	sourceConnections, ok := connections[s.SourceNode].(map[string]interface{})
	if !ok {
		sourceConnections = make(map[string]interface{})
		connections[s.SourceNode] = sourceConnections
	}

	// Get or create main output array
	mainOutputs, ok := sourceConnections["main"].([]interface{})
	if !ok {
		mainOutputs = make([]interface{}, 0)
		sourceConnections["main"] = mainOutputs
	}

	// Ensure output index exists
	for len(mainOutputs) <= s.OutputIndex {
		mainOutputs = append(mainOutputs, make([]interface{}, 0))
	}
	sourceConnections["main"] = mainOutputs

	// Create the connection
	connection := map[string]interface{}{
		"node":  s.TargetNode,
		"type":  "main",
		"index": s.InputIndex,
	}

	// Add connection to outputs
	outputArray := mainOutputs[s.OutputIndex].([]interface{})
	outputArray = append(outputArray, connection)
	mainOutputs[s.OutputIndex] = outputArray

	// Update workflow
	result, err := client.UpdateWorkflow(s.WorkflowID, workflow)
	if err != nil {
		return err
	}

	// Output as row
	row := types.NewRowFromMap(result)
	return gp.AddRow(ctx, row)
}

// Ensure the interface is implemented
var _ cmds.GlazeCommand = &ConnectNodesCommand{}

// NewConnectNodesCommand creates a new ConnectNodesCommand
func NewConnectNodesCommand() (*ConnectNodesCommand, error) {
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
		"connect-nodes",
		cmds.WithShort("Connect nodes in a workflow"),
		cmds.WithLong("Connect two nodes in an existing workflow in the n8n instance."),

		// Define flags (parameters)
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"workflow-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("ID of the workflow"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"source-node",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name of the source node"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"target-node",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name of the target node"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"output-index",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Output index of the source node (0 for first output)"),
				parameters.WithDefault(0),
			),
			parameters.NewParameterDefinition(
				"input-index",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Input index of the target node (0 for first input)"),
				parameters.WithDefault(0),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(glazedLayer, apiLayer),
	)

	// Return the command instance
	return &ConnectNodesCommand{
		CommandDescription: cmdDesc,
	}, nil
}