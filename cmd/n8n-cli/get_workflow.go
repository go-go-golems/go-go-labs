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
	"io"
	"net/http"
)

// GetWorkflowCommand gets a workflow by ID
type GetWorkflowCommand struct {
	*cmds.CommandDescription
}

// Settings for GetWorkflowCommand
type GetWorkflowSettings struct {
	ID                string `glazed.parameter:"id"`
	SaveToFile        string `glazed.parameter:"save-to-file"`
	WithNodes         bool   `glazed.parameter:"with-nodes"`
	WithNodePositions bool   `glazed.parameter:"with-node-positions"`
	MermaidOutput     bool   `glazed.parameter:"mermaid-output"`
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
	endpoint := fmt.Sprintf("workflows/%s", s.ID)
	// Note: n8n API always returns the complete workflow with nodes
	// We'll handle the filtering client-side based on the with-nodes flag

	// Execute the request
	resp, err := client.DoRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	// Parse the response
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse the workflow data
	var workflow map[string]interface{}
	if err := json.Unmarshal(data, &workflow); err != nil {
		return err
	}

	// Remove node positions if requested
	if s.WithNodes && !s.WithNodePositions {
		if nodes, ok := workflow["nodes"].([]interface{}); ok {
			for i := range nodes {
				if node, ok := nodes[i].(map[string]interface{}); ok {
					delete(node, "position")
				}
			}
		}
	}

	// If not requesting nodes, remove them from the response
	if !s.WithNodes {
		delete(workflow, "nodes")
		delete(workflow, "connections")
	}

	// Check if we should output as mermaid
	if s.MermaidOutput {
		// For mermaid, we need the nodes and connections
		// No need to refetch since n8n API always includes nodes

		// Generate mermaid diagram and extract sticky notes
		result := WorkflowToMermaid(workflow)

		// Print sticky notes as markdown
		for _, note := range result.Notes {
			fmt.Printf("> %s\n\n", note)
		}

		// Print the mermaid diagram
		fmt.Println("```mermaid")
		fmt.Print(result.MermaidCode)
		fmt.Println("```")

		// Return empty for the processor
		return nil
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
			parameters.NewParameterDefinition(
				"with-nodes",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Include workflow nodes in response (client-side filtering)"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition(
				"with-node-positions",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Include node position data (only applies when with-nodes is true)"),
				parameters.WithDefault(true),
			),
			parameters.NewParameterDefinition(
				"mermaid-output",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Output workflow as mermaid diagram to stdout"),
				parameters.WithDefault(false),
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
