package assistants

import (
	"context"
	"encoding/json"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/pkg/assistants"
	"os"
)

type ModifyAssistantCommand struct {
	*cmds.CommandDescription
}

func NewModifyAssistantCommand() (*ModifyAssistantCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &ModifyAssistantCommand{
		CommandDescription: cmds.NewCommandDescription(
			"modify",
			cmds.WithShort("Modify an assistant"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("id", parameters.ParameterTypeString, parameters.WithHelp("ID of the assistant")),
				parameters.NewParameterDefinition("object", parameters.ParameterTypeString, parameters.WithHelp("Object type of the assistant")),
				parameters.NewParameterDefinition("name", parameters.ParameterTypeString, parameters.WithHelp("Name of the assistant")),
				parameters.NewParameterDefinition("description", parameters.ParameterTypeString, parameters.WithHelp("Description of the assistant")),
				parameters.NewParameterDefinition("model", parameters.ParameterTypeString, parameters.WithHelp("Model of the assistant")),
				parameters.NewParameterDefinition("instructions", parameters.ParameterTypeString, parameters.WithHelp("Instructions for the assistant")),
				parameters.NewParameterDefinition("tools", parameters.ParameterTypeObjectListFromFile, parameters.WithHelp("Tools used by the assistant")),
				parameters.NewParameterDefinition("file_ids", parameters.ParameterTypeStringList, parameters.WithHelp("File IDs associated with the assistant")),
				parameters.NewParameterDefinition("metadata", parameters.ParameterTypeKeyValue, parameters.WithHelp("Metadata for the assistant")),
			),
			cmds.WithLayers(glazedLayer),
		),
	}, nil
}

func (c *ModifyAssistantCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	assistantID := ps["id"].(string)

	// go through all flags and check if set and then update
	updateData := assistants.Assistant{}

	if ps["name"] != nil {
		updateData.Name = ps["name"].(string)
	}
	if ps["model"] != nil {
		updateData.Model = ps["model"].(string)
	}
	if ps["instructions"] != nil {
		updateData.Instructions = ps["instructions"].(string)
	}
	if ps["tools"] != nil {
		// deserialize reserialize to Tool
		var res []assistants.Tool

		s, err := json.Marshal(ps["tools"])
		if err != nil {
			return err
		}

		err = json.Unmarshal(s, &res)
		if err != nil {
			return err
		}

		updateData.Tools = res
	}
	if ps["file_ids"] != nil {
		updateData.FileIDs = ps["file_ids"].([]string)
	}
	if ps["metadata"] != nil {
		updateData.Metadata = ps["metadata"].(map[string]interface{})
	}

	assistant, err := assistants.ModifyAssistant(apiKey, assistantID, updateData)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("id", assistant.ID),
		types.MRP("object", assistant.Object),
		types.MRP("created_at", assistant.CreatedAt),
		types.MRP("name", assistant.Name),
		types.MRP("description", assistant.Description),
		types.MRP("model", assistant.Model),
		types.MRP("instructions", assistant.Instructions),
		types.MRP("tools", assistant.Tools),
		types.MRP("file_ids", assistant.FileIDs),
		types.MRP("metadata", assistant.Metadata),
	)

	return gp.AddRow(ctx, row)
}
