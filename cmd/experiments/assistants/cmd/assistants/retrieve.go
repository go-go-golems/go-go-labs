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

type RetrieveAssistantCommand struct {
	*cmds.CommandDescription
}

func NewRetrieveAssistantCommand() (*RetrieveAssistantCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	assistantIDArg := parameters.NewParameterDefinition(
		"assistantID",
		parameters.ParameterTypeString,
		parameters.WithHelp("The ID of the assistant to retrieve"),
	)

	return &RetrieveAssistantCommand{
		CommandDescription: cmds.NewCommandDescription(
			"retrieve",
			cmds.WithShort("Retrieve an assistant"),
			cmds.WithArguments(assistantIDArg),
			cmds.WithLayers(glazedLayer),
		),
	}, nil
}

func (c *RetrieveAssistantCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	assistantID := ps["assistantID"].(string)

	assistant, err := assistants.RetrieveAssistant(apiKey, assistantID)
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
