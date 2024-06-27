package cmds

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/apps/bee/pkg/bee"
	"os"
)

// GetConversationCommand
type GetConversationCommand struct {
	*cmds.CommandDescription
}

type GetConversationSettings struct {
	ConversationID int `glazed.parameter:"conversation_id"`
}

func NewGetConversationCommand() (*GetConversationCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &GetConversationCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get",
			cmds.WithShort("Get a specific conversation"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"conversation_id",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("ID of the conversation to retrieve"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *GetConversationCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &GetConversationSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	conversation, err := client.GetConversation(ctx, "me", s.ConversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	row := types.NewRow(
		types.MRP("id", conversation.ID),
		types.MRP("start_time", conversation.StartTime),
		types.MRP("end_time", conversation.EndTime),
		types.MRP("device_type", conversation.DeviceType),
		types.MRP("summary", conversation.Summary),
		types.MRP("short_summary", conversation.ShortSummary),
		types.MRP("state", conversation.State),
		types.MRP("created_at", conversation.CreatedAt),
		types.MRP("updated_at", conversation.UpdatedAt),
		types.MRP("primary_location", conversation.PrimaryLocation),
		types.MRP("transcriptions", conversation.Transcriptions),
		types.MRP("suggested_links", conversation.SuggestedLinks),
	)

	return gp.AddRow(ctx, row)
}

// DeleteConversationCommand
type DeleteConversationCommand struct {
	*cmds.CommandDescription
}

type DeleteConversationSettings struct {
	ConversationID int `glazed.parameter:"conversation_id"`
}

func NewDeleteConversationCommand() (*DeleteConversationCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &DeleteConversationCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete",
			cmds.WithShort("Delete a specific conversation"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"conversation_id",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("ID of the conversation to delete"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *DeleteConversationCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	s := &DeleteConversationSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	err := client.DeleteConversation(ctx, "me", s.ConversationID)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	fmt.Printf("Conversation %d deleted successfully\n", s.ConversationID)
	return nil
}

// EndConversationCommand
type EndConversationCommand struct {
	*cmds.CommandDescription
}

type EndConversationSettings struct {
	ConversationID int `glazed.parameter:"conversation_id"`
}

func NewEndConversationCommand() (*EndConversationCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &EndConversationCommand{
		CommandDescription: cmds.NewCommandDescription(
			"end",
			cmds.WithShort("End a specific conversation"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"conversation_id",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("ID of the conversation to end"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *EndConversationCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	s := &EndConversationSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	err := client.EndConversation(ctx, "me", s.ConversationID)
	if err != nil {
		return fmt.Errorf("failed to end conversation: %w", err)
	}

	fmt.Printf("Conversation %d ended successfully\n", s.ConversationID)
	return nil
}

// RetryConversationCommand
type RetryConversationCommand struct {
	*cmds.CommandDescription
}

type RetryConversationSettings struct {
	ConversationID int `glazed.parameter:"conversation_id"`
}

func NewRetryConversationCommand() (*RetryConversationCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &RetryConversationCommand{
		CommandDescription: cmds.NewCommandDescription(
			"retry",
			cmds.WithShort("Retry a specific conversation"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"conversation_id",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("ID of the conversation to retry"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *RetryConversationCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	s := &RetryConversationSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	err := client.RetryConversation(ctx, "me", s.ConversationID)
	if err != nil {
		return fmt.Errorf("failed to retry conversation: %w", err)
	}

	fmt.Printf("Conversation %d retried successfully\n", s.ConversationID)
	return nil
}

type ListConversationsCommand struct {
	*cmds.CommandDescription
}

func NewListConversationsCommand() (*ListConversationsCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &ListConversationsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List all conversations"),
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
					parameters.WithHelp("Number of conversations per page"),
					parameters.WithDefault(10),
				),
			),
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

type ListConversationsSettings struct {
	Page  int `glazed.parameter:"page"`
	Limit int `glazed.parameter:"limit"`
}

func (c *ListConversationsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ListConversationsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Assuming you have a client that can fetch conversations
	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	conversations, err := client.GetConversations(ctx, "me", s.Page, s.Limit)
	if err != nil {
		return fmt.Errorf("failed to get conversations: %w", err)
	}

	for _, conversation := range conversations.Conversations {
		row := types.NewRow(
			types.MRP("id", conversation.ID),
			types.MRP("start_time", conversation.StartTime),
			types.MRP("end_time", conversation.EndTime),
			// 	DeviceType      string          `json:"device_type"`
			//	Summary         *string         `json:"summary"`
			//	ShortSummary    *string         `json:"short_summary"`
			//	State           string          `json:"state"`
			//	CreatedAt       time.Time       `json:"created_at"`
			//	UpdatedAt       time.Time       `json:"updated_at"`
			//	PrimaryLocation *Location       `json:"primary_location"`
			//	Transcriptions  []Transcription `json:"transcriptions"`
			//	SuggestedLinks  []string        `json:"suggested_links"`
			types.MRP("short_summary", ""),
			types.MRP("summary", ""),
			types.MRP("state", conversation.State),
			types.MRP("created_at", conversation.CreatedAt),
			types.MRP("updated_at", conversation.UpdatedAt),
			types.MRP("suggested_links", conversation.SuggestedLinks),
			types.MRP("transcriptions", conversation.Transcriptions),
			types.MRP("device_type", conversation.DeviceType),
		)
		if conversation.Summary != nil {
			row.Set("summary", *conversation.Summary)
		}
		if conversation.ShortSummary != nil {
			row.Set("short_summary", *conversation.ShortSummary)
		}
		if conversation.PrimaryLocation != nil {
			row.Set("primary_location", conversation.PrimaryLocation)
		}
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
