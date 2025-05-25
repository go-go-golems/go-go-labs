package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/n8n-cli/pkg/n8n"
)

// AddNodeCommand adds a node to an existing workflow
type AddNodeCommand struct {
	*cmds.CommandDescription
}

// Settings for AddNodeCommand
type AddNodeSettings struct {
	WorkflowID string `glazed.parameter:"workflow-id"`
	NodeType   string `glazed.parameter:"node-type"`
	NodeName   string `glazed.parameter:"node-name"`
	Parameters string `glazed.parameter:"parameters"`
	PositionX  int    `glazed.parameter:"position-x"`
	PositionY  int    `glazed.parameter:"position-y"`
}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *AddNodeCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &AddNodeSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Get API settings
	apiSettings, err := n8n.GetN8NAPISettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	// Create API client
	client := n8n.NewN8NClient(apiSettings.BaseURL, apiSettings.APIKey)

	// Get current workflow
	workflow, err := client.GetWorkflow(s.WorkflowID)
	if err != nil {
		return err
	}

	// Parse parameters JSON
	var parameters map[string]interface{}
	if s.Parameters != "" {
		if err := json.Unmarshal([]byte(s.Parameters), &parameters); err != nil {
			return fmt.Errorf("invalid parameters JSON: %w", err)
		}
	} else {
		parameters = map[string]interface{}{}
	}

	// Create new node
	newNode := map[string]interface{}{
		"name":        s.NodeName,
		"type":        s.NodeType,
		"typeVersion": 1,
		"position":    []int{s.PositionX, s.PositionY},
		"parameters":  parameters,
	}

	// Add node to workflow
	workflow["nodes"] = append(workflow["nodes"].([]interface{}), newNode)

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
var _ cmds.GlazeCommand = &AddNodeCommand{}

// NewAddNodeCommand creates a new AddNodeCommand
func NewAddNodeCommand() (*AddNodeCommand, error) {
	// Create the standard Glazed output layer
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Add the n8n API layer
	apiLayer, err := n8n.NewN8NAPILayer()
	if err != nil {
		return nil, err
	}

	// Create the command description
	cmdDesc := cmds.NewCommandDescription(
		"add-node",
		cmds.WithShort("Add a node to a workflow"),
		cmds.WithLong("Add a new node to an existing workflow in the n8n instance."),

		// Define flags (parameters)
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"workflow-id",
				parameters.ParameterTypeString,
				parameters.WithHelp("ID of the workflow to add the node to"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"node-type",
				parameters.ParameterTypeString,
				parameters.WithHelp("Type of node to add (e.g., n8n-nodes-base.webhook)"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"node-name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Name for the new node"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"parameters",
				parameters.ParameterTypeString,
				parameters.WithHelp("JSON string with node parameters"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"position-x",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("X position in workflow editor"),
				parameters.WithDefault(200),
			),
			parameters.NewParameterDefinition(
				"position-y",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Y position in workflow editor"),
				parameters.WithDefault(300),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(glazedLayer, apiLayer),
	)

	// Return the command instance
	return &AddNodeCommand{
		CommandDescription: cmdDesc,
	}, nil
}
