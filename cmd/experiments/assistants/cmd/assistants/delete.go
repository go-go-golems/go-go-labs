package assistants

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/pkg/assistants"
	"os"
)

type DeleteAssistantCommand struct {
	*cmds.CommandDescription
}

func NewDeleteAssistantCommand() (*DeleteAssistantCommand, error) {
	return &DeleteAssistantCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete",
			cmds.WithShort("Delete an assistant"),
			cmds.WithFlags(
				parameters.NewParameterDefinition("id",
					parameters.ParameterTypeString,
					parameters.WithHelp("ID of the assistant"),
					parameters.WithRequired(true),
				),
			),
		),
	}, nil
}

func (c *DeleteAssistantCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
) error {
	assistantID, ok := ps["id"].(string)
	if !ok {
		return fmt.Errorf("id is required")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	err := assistants.DeleteAssistant(apiKey, assistantID)
	if err != nil {
		fmt.Println("Error deleting assistant:", err)
		return err
	}
	fmt.Println("Assistant deleted successfully")
	return nil
}
