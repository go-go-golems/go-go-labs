package main

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/n8n-cli/pkg/n8n"
)

// CreateWorkflowCommand creates a new workflow
type CreateWorkflowCommand struct {
	*cmds.CommandDescription
}

// Settings for CreateWorkflowCommand
type CreateWorkflowSettings struct {
	Name   string `glazed.parameter:"name"`
	File   string `glazed.parameter:"file"`
	Active bool   `glazed.parameter:"active"`
}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *CreateWorkflowCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from layers
	s := &CreateWorkflowSettings{}
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

	// Prepare workflow data
	var workflowData map[string]interface{}

	if s.File != "" {
		// Read from JSON file
		if err := n8n.ReadJSONFile(s.File, &workflowData); err != nil {
			return err
		}

		// Override name and active status
		workflowData["name"] = s.Name
		workflowData["active"] = s.Active
	} else {
		// Create a minimal workflow
		workflowData = map[string]interface{}{
			"name":        s.Name,
			"active":      s.Active,
			"nodes":       []interface{}{},
			"connections": map[string]interface{}{},
		}
	}

	// Create workflow
	result, err := client.CreateWorkflow(workflowData)
	if err != nil {
		return err
	}

	// Output as row
	row := types.NewRowFromMap(result)
	return gp.AddRow(ctx, row)
}

// Ensure the interface is implemented
var _ cmds.GlazeCommand = &CreateWorkflowCommand{}

// NewCreateWorkflowCommand creates a new CreateWorkflowCommand
func NewCreateWorkflowCommand() (*CreateWorkflowCommand, error) {
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
		"create-workflow",
		cmds.WithShort("Create a new workflow"),
		cmds.WithLong("Create a new workflow in the n8n instance."),

		// Define flags (parameters)
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Workflow name"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"file",
				parameters.ParameterTypeString,
				parameters.WithHelp("JSON file with workflow definition"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"active",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Set workflow as active"),
				parameters.WithDefault(false),
			),
		),

		// Add parameter layers
		cmds.WithLayersList(glazedLayer, apiLayer),
	)

	// Return the command instance
	return &CreateWorkflowCommand{
		CommandDescription: cmdDesc,
	}, nil
}
