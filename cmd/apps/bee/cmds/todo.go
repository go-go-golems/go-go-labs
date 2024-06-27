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

// ListTodosCommand
type ListTodosCommand struct {
	*cmds.CommandDescription
}

type ListTodosSettings struct {
	Page  int `glazed.parameter:"page"`
	Limit int `glazed.parameter:"limit"`
}

func NewListTodosCommand() (*ListTodosCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &ListTodosCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List todos"),
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
					parameters.WithHelp("Number of todos per page"),
					parameters.WithDefault(10),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *ListTodosCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ListTodosSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	todos, err := client.GetTodos(ctx, "me", s.Page, s.Limit)
	if err != nil {
		return fmt.Errorf("failed to get todos: %w", err)
	}

	for _, todo := range todos.Todos {
		row := types.NewRow(
			types.MRP("id", todo.ID),
			types.MRP("text", todo.Text),
			types.MRP("alarm_at", todo.AlarmAt),
			types.MRP("completed", todo.Completed),
			types.MRP("created_at", todo.CreatedAt),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// CreateTodoCommand
type CreateTodoCommand struct {
	*cmds.CommandDescription
}

type CreateTodoSettings struct {
	Text      string `glazed.parameter:"text"`
	Completed bool   `glazed.parameter:"completed"`
}

func NewCreateTodoCommand() (*CreateTodoCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &CreateTodoCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create",
			cmds.WithShort("Create a new todo"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"text",
					parameters.ParameterTypeString,
					parameters.WithHelp("Text of the todo"),
				),
				parameters.NewParameterDefinition(
					"completed",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Whether the todo is completed"),
					parameters.WithDefault(false),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *CreateTodoCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &CreateTodoSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	todo, err := client.CreateTodo(ctx, "me", bee.TodoInput{
		Text:      s.Text,
		Completed: s.Completed,
	})
	if err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}

	row := types.NewRow(
		types.MRP("id", todo.ID),
		types.MRP("text", todo.Text),
		types.MRP("alarm_at", todo.AlarmAt),
		types.MRP("completed", todo.Completed),
		types.MRP("created_at", todo.CreatedAt),
	)
	return gp.AddRow(ctx, row)
}

// GetTodoCommand
type GetTodoCommand struct {
	*cmds.CommandDescription
}

type GetTodoSettings struct {
	TodoID int `glazed.parameter:"todo_id"`
}

func NewGetTodoCommand() (*GetTodoCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &GetTodoCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get",
			cmds.WithShort("Get a specific todo"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"todo_id",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("ID of the todo to retrieve"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *GetTodoCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &GetTodoSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	todo, err := client.GetTodo(ctx, "me", s.TodoID)
	if err != nil {
		return fmt.Errorf("failed to get todo: %w", err)
	}

	row := types.NewRow(
		types.MRP("id", todo.ID),
		types.MRP("text", todo.Text),
		types.MRP("alarm_at", todo.AlarmAt),
		types.MRP("completed", todo.Completed),
		types.MRP("created_at", todo.CreatedAt),
	)
	return gp.AddRow(ctx, row)
}

// UpdateTodoCommand
type UpdateTodoCommand struct {
	*cmds.CommandDescription
}

type UpdateTodoSettings struct {
	TodoID    int    `glazed.parameter:"todo_id"`
	Text      string `glazed.parameter:"text"`
	Completed bool   `glazed.parameter:"completed"`
}

func NewUpdateTodoCommand() (*UpdateTodoCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &UpdateTodoCommand{
		CommandDescription: cmds.NewCommandDescription(
			"update",
			cmds.WithShort("Update a specific todo"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"todo_id",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("ID of the todo to update"),
				),
			),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"text",
					parameters.ParameterTypeString,
					parameters.WithHelp("New text for the todo"),
				),
				parameters.NewParameterDefinition(
					"completed",
					parameters.ParameterTypeBool,
					parameters.WithHelp("New completed status for the todo"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *UpdateTodoCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &UpdateTodoSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	todo, err := client.UpdateTodo(ctx, "me", s.TodoID, bee.TodoInput{
		Text:      s.Text,
		Completed: s.Completed,
	})
	if err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	row := types.NewRow(
		types.MRP("id", todo.ID),
		types.MRP("text", todo.Text),
		types.MRP("alarm_at", todo.AlarmAt),
		types.MRP("completed", todo.Completed),
		types.MRP("created_at", todo.CreatedAt),
	)
	return gp.AddRow(ctx, row)
}

// DeleteTodoCommand
type DeleteTodoCommand struct {
	*cmds.CommandDescription
}

type DeleteTodoSettings struct {
	TodoID int `glazed.parameter:"todo_id"`
}

func NewDeleteTodoCommand() (*DeleteTodoCommand, error) {
	glazedParameterLayer, err := createGlazedParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &DeleteTodoCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete",
			cmds.WithShort("Delete a specific todo"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"todo_id",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("ID of the todo to delete"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *DeleteTodoCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	s := &DeleteTodoSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	client := bee.NewClient(os.Getenv("BEE_API_KEY"))
	err := client.DeleteTodo(ctx, "me", s.TodoID)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	fmt.Printf("Todo %d deleted successfully\n", s.TodoID)
	return nil
}
