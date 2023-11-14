package assistants

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/pkg/assistants"
	"os"
)

type CreateAssistantCommand struct {
	*cmds.CommandDescription
}

func NewCreateAssistantCommand() (*CreateAssistantCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &CreateAssistantCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Create a new assistant"),
			cmds.WithFlags(
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

func (c *CreateAssistantCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	assistantData := assistants.Assistant{
		Name:         ps["name"].(string),
		Model:        ps["model"].(string),
		Instructions: ps["instructions"].(string),
		// Set other fields from flags
	}
	assistant, err := assistants.CreateAssistant(apiKey, assistantData)
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
