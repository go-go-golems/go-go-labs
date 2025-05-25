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
	"io"
	"net/http"
)

// ListWorkflowsCommand lists all workflows
type ListWorkflowsCommand struct {
	*cmds.CommandDescription
}

// Settings for ListWorkflowsCommand
type ListWorkflowsSettings struct {
	Active            bool   `glazed.parameter:"active"`
	Limit             int    `glazed.parameter:"limit"`
	Cursor            string `glazed.parameter:"cursor"`
	WithNodes         bool   `glazed.parameter:"with-nodes"`
	WithNodePositions bool   `glazed.parameter:"with-node-positions"`
	MermaidOutput     bool   `glazed.parameter:"mermaid-output"`
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
	apiSettings, err := n8n.GetN8NAPISettingsFromParsedLayers(parsedLayers)
	if err != nil {
		return err
	}

	// Create API client
	client := n8n.NewN8NClient(apiSettings.BaseURL, apiSettings.APIKey)

	// List workflows with cursor-based pagination
	endpoint := fmt.Sprintf("workflows?limit=%d", s.Limit)
	if s.Active {
		endpoint += "&active=true"
	}
	if s.Cursor != "" {
		endpoint += fmt.Sprintf("&cursor=%s", s.Cursor)
	}
	// Note: withNodes parameter is not supported by the API
	// Instead, we'll handle the filtering client-side
	if s.MermaidOutput {
		// If mermaid output is requested, always get nodes
		s.WithNodes = true
	}

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

	// Parse the response which should include data and nextCursor
	var result struct {
		Data       []map[string]interface{} `json:"data"`
		NextCursor string                   `json:"nextCursor"`
	}

	var workflows []map[string]interface{}
	var nextCursor string

	if err := json.Unmarshal(data, &result); err != nil {
		// Try parsing as direct array for older n8n versions
		if jsonErr := json.Unmarshal(data, &workflows); jsonErr != nil {
			return fmt.Errorf("failed to parse JSON response: %w", err)
		}
	} else {
		workflows = result.Data
		nextCursor = result.NextCursor
	}

	// Remove node positions if requested
	if s.WithNodes && s.WithNodePositions == false {
		for i := range workflows {
			if nodes, ok := workflows[i]["nodes"].([]interface{}); ok {
				for j := range nodes {
					if node, ok := nodes[j].(map[string]interface{}); ok {
						delete(node, "position")
					}
				}
			}
		}
	}

	// If not requesting nodes, remove them from the response
	if !s.WithNodes {
		for i := range workflows {
			delete(workflows[i], "nodes")
			delete(workflows[i], "connections")
		}
	}

	// Check if we should output as mermaid
	if s.MermaidOutput {
		// For each workflow with nodes, output a mermaid diagram
		for _, workflow := range workflows {
			// Skip workflows without nodes (in case withNodes=false)
			_, hasNodes := workflow["nodes"]
			if !hasNodes {
				continue
			}

			// Generate mermaid diagram and extract sticky notes
			result := n8n.WorkflowToMermaid(workflow)

			// Print header with workflow name/id
			name, _ := workflow["name"].(string)
			id, _ := workflow["id"].(string)
			fmt.Printf("\n# Workflow: %s (ID: %s)\n\n", name, id)

			// Print sticky notes as markdown
			for _, note := range result.Notes {
				fmt.Printf("> %s\n\n", note)
			}

			// Print the mermaid diagram
			fmt.Printf("```mermaid\n%s```\n\n", result.MermaidCode)
		}

		// Return empty for the processor
		return nil
	}

	// Add nextCursor as metadata if present
	if nextCursor != "" {
		// Add to context or metadata
		ctx = context.WithValue(ctx, "nextCursor", nextCursor)
	}

	// Output as rows
	for i, workflow := range workflows {
		// Add nextCursor as a special field in the first row if present
		if nextCursor != "" && i == 0 {
			workflow["__nextCursor"] = nextCursor
		}

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

	apiLayer, err := n8n.NewN8NAPILayer()
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
				"cursor",
				parameters.ParameterTypeString,
				parameters.WithHelp("Cursor for pagination"),
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
				parameters.WithHelp("Output workflows as mermaid diagrams to stdout"),
				parameters.WithDefault(false),
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
