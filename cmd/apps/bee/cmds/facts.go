package cmds

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/apps/bee/pkg/bee"
	"os"
)

// ListFactsCommand
type ListFactsCommand struct {
	*cmds.CommandDescription
}

type ListFactsSettings struct {
	Page      int  `glazed.parameter:"page"`
	Limit     int  `glazed.parameter:"limit"`
	Confirmed bool `glazed.parameter:"confirmed"`
}

func NewListFactsCommand() (*ListFactsCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &ListFactsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List facts"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"page",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Page number for pagination"),
					parameters.WithDefault(1),
				),
				parameters.NewParameterDefinition(
					"limit",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Number of facts per page"),
					parameters.WithDefault(10),
				),
				parameters.NewParameterDefinition(
					"confirmed",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Filter by confirmed status"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *ListFactsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ListFactsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	facts, err := client.GetFacts(ctx, "me", s.Page, s.Limit, &s.Confirmed)
	if err != nil {
		return fmt.Errorf("failed to get facts: %w", err)
	}

	for _, fact := range facts.Facts {
		row := types.NewRow(
			types.MRP("id", fact.ID),
			types.MRP("text", fact.Text),
			types.MRP("tags", fact.Tags),
			types.MRP("visibility", fact.Visibility),
			types.MRP("confirmed", fact.Confirmed),
			types.MRP("user_id", fact.UserID),
			types.MRP("updated_at", fact.UpdatedAt),
			types.MRP("created_at", fact.CreatedAt),
			types.MRP("topic", fact.Topic),
			types.MRP("source", fact.Source),
			types.MRP("score", fact.Score),
			types.MRP("embedding", fact.Embedding),
			types.MRP("fts", fact.FTS),
			types.MRP("conversation_id", fact.ConversationID),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// CreateFactCommand
type CreateFactCommand struct {
	*cmds.CommandDescription
}

type CreateFactSettings struct {
	Text      string `glazed.parameter:"text"`
	Confirmed bool   `glazed.parameter:"confirmed"`
}

func NewCreateFactCommand() (*CreateFactCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &CreateFactCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Create a new fact"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"text",
					parameters.ParameterTypeString,
					parameters.WithHelp("Text of the fact"),
				),
				parameters.NewParameterDefinition(
					"confirmed",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Whether the fact is confirmed"),
					parameters.WithDefault(false),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *CreateFactCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &CreateFactSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	fact, err := client.CreateFact(ctx, "me", bee.FactInput{
		Text:      s.Text,
		Confirmed: s.Confirmed,
	})
	if err != nil {
		return fmt.Errorf("failed to create fact: %w", err)
	}

	row := types.NewRow(
		types.MRP("id", fact.ID),
		types.MRP("text", fact.Text),
		types.MRP("tags", fact.Tags),
		types.MRP("visibility", fact.Visibility),
		types.MRP("confirmed", fact.Confirmed),
		types.MRP("user_id", fact.UserID),
		types.MRP("updated_at", fact.UpdatedAt),
		types.MRP("created_at", fact.CreatedAt),
		types.MRP("topic", fact.Topic),
		types.MRP("source", fact.Source),
		types.MRP("score", fact.Score),
		types.MRP("embedding", fact.Embedding),
		types.MRP("fts", fact.FTS),
		types.MRP("conversation_id", fact.ConversationID),
	)
	return gp.AddRow(ctx, row)
}

// GetFactCommand
type GetFactCommand struct {
	*cmds.CommandDescription
}

type GetFactSettings struct {
	FactID int `glazed.parameter:"fact_id"`
}

func NewGetFactCommand() (*GetFactCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &GetFactCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get",
			cmds.WithShort("Get a specific fact"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"fact_id",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("ID of the fact to retrieve"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *GetFactCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &GetFactSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	fact, err := client.GetFact(ctx, "me", s.FactID)
	if err != nil {
		return fmt.Errorf("failed to get fact: %w", err)
	}

	row := types.NewRow(
		types.MRP("id", fact.ID),
		types.MRP("text", fact.Text),
		types.MRP("tags", fact.Tags),
		types.MRP("visibility", fact.Visibility),
		types.MRP("confirmed", fact.Confirmed),
		types.MRP("user_id", fact.UserID),
		types.MRP("updated_at", fact.UpdatedAt),
		types.MRP("created_at", fact.CreatedAt),
		types.MRP("topic", fact.Topic),
		types.MRP("source", fact.Source),
		types.MRP("score", fact.Score),
		types.MRP("embedding", fact.Embedding),
		types.MRP("fts", fact.FTS),
		types.MRP("conversation_id", fact.ConversationID),
	)
	return gp.AddRow(ctx, row)
}

// UpdateFactCommand
type UpdateFactCommand struct {
	*cmds.CommandDescription
}

type UpdateFactSettings struct {
	FactID    int    `glazed.parameter:"fact_id"`
	Text      string `glazed.parameter:"text"`
	Confirmed bool   `glazed.parameter:"confirmed"`
}

func NewUpdateFactCommand() (*UpdateFactCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &UpdateFactCommand{
		CommandDescription: cmds.NewCommandDescription(
			"update",
			cmds.WithShort("Update a specific fact"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"fact_id",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("ID of the fact to update"),
				),
			),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"text",
					parameters.ParameterTypeString,
					parameters.WithHelp("New text for the fact"),
				),
				parameters.NewParameterDefinition(
					"confirmed",
					parameters.ParameterTypeBool,
					parameters.WithHelp("New confirmed status for the fact"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *UpdateFactCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &UpdateFactSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	fact, err := client.UpdateFact(ctx, "me", s.FactID, bee.FactInput{
		Text:      s.Text,
		Confirmed: s.Confirmed,
	})
	if err != nil {
		return fmt.Errorf("failed to update fact: %w", err)
	}

	row := types.NewRow(
		types.MRP("id", fact.ID),
		types.MRP("text", fact.Text),
		types.MRP("tags", fact.Tags),
		types.MRP("visibility", fact.Visibility),
		types.MRP("confirmed", fact.Confirmed),
		types.MRP("user_id", fact.UserID),
		types.MRP("updated_at", fact.UpdatedAt),
		types.MRP("created_at", fact.CreatedAt),
		types.MRP("topic", fact.Topic),
		types.MRP("source", fact.Source),
		types.MRP("score", fact.Score),
		types.MRP("embedding", fact.Embedding),
		types.MRP("fts", fact.FTS),
		types.MRP("conversation_id", fact.ConversationID),
	)
	return gp.AddRow(ctx, row)
}

// DeleteFactCommand
type DeleteFactCommand struct {
	*cmds.CommandDescription
}

type DeleteFactSettings struct {
	FactID int `glazed.parameter:"fact_id"`
}

func NewDeleteFactCommand() (*DeleteFactCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &DeleteFactCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete",
			cmds.WithShort("Delete a specific fact"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"fact_id",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("ID of the fact to delete"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *DeleteFactCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	s := &DeleteFactSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	err := client.DeleteFact(ctx, "me", s.FactID)
	if err != nil {
		return fmt.Errorf("failed to delete fact: %w", err)
	}

	fmt.Printf("Fact %d deleted successfully\n", s.FactID)
	return nil
}
