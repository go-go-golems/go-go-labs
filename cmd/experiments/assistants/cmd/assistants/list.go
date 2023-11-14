package assistants

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/pkg/assistants"
	"os"
	// Import other necessary packages
)

type ListAssistantsCommand struct {
	*cmds.CommandDescription
}

func NewListAssistantsCommand() (*ListAssistantsCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	return &ListAssistantsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List all assistants"),
			cmds.WithLayers(glazedLayer),
		),
	}, nil
}

func (c *ListAssistantsCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	after := "" // Initialize after to an empty string
	limit := 20 // Set default limit or get from flags

	for {
		assistants, hasMore, err := assistants.ListAssistants(apiKey, after, limit)
		if err != nil {
			return err
		}

		for _, assistant := range assistants {
			row := types.NewRow(
				types.MRP("id", assistant.ID),
				types.MRP("name", assistant.Name),
				types.MRP("object", assistant.Object),
				types.MRP("created_at", assistant.CreatedAt),
				types.MRP("description", assistant.Description),
				types.MRP("model", assistant.Model),
				types.MRP("instructions", assistant.Instructions),
				types.MRP("tools", assistant.Tools),
				types.MRP("file_ids", assistant.FileIDs),
				types.MRP("metadata", assistant.Metadata),

				// Add other assistant fields as necessary
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}

		if hasMore {
			after = assistants[len(assistants)-1].ID
		} else {
			break
		}

		// Add logic for interactive pagination if necessary
	}

	return nil
}
