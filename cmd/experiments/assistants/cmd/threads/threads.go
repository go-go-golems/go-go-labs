package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/pkg/assistants"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	// Import other necessary packages
)

// Define the base URL and the environment variable for the API key
const baseURL = "https://api.openai.com/v1"
const apiKeyEnvVar = "OPENAI_API_KEY"

// ListThreadsCommand uses GlazeCommand to output structured data
type ListThreadsCommand struct {
	*cmds.CommandDescription
}

func NewListThreadsCommand() (*ListThreadsCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	return &ListThreadsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list-threads",
			cmds.WithShort("List all threads"),
			cmds.WithLayers(glazedLayer),
		),
	}, nil
}

func (c *ListThreadsCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	apiKey := os.Getenv(apiKeyEnvVar)
	client := &http.Client{}

	threadsResponse, err := assistants.ListThreads(client, baseURL, apiKey)
	if err != nil {
		return err
	}

	for _, thread := range threadsResponse.Data {
		row := types.NewRow(
			types.MRP("id", thread.ID),
			types.MRP("object", thread.Object),
			types.MRP("created_at", thread.CreatedAt),
			types.MRP("metadata", thread.Metadata),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// CreateThreadCommand uses BareCommand as it does not need to output structured data
type CreateThreadCommand struct {
	*cmds.CommandDescription
}

func NewCreateThreadCommand() (*CreateThreadCommand, error) {
	return &CreateThreadCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create-thread",
			cmds.WithShort("Create a new thread"),
			// Define flags and arguments as needed
		),
	}, nil
}

func (c *CreateThreadCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
) error {
	apiKey := os.Getenv(apiKeyEnvVar)
	client := &http.Client{}

	// Construct the request body from flags/arguments
	request := assistants.CreateThreadRequest{
		// Populate the request fields
	}

	thread, err := assistants.CreateThread(client, baseURL, apiKey, request)
	if err != nil {
		return err
	}

	// Output the result
	output, err := json.Marshal(thread)
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

// ModifyThreadCommand uses BareCommand for simple execution without structured output
type ModifyThreadCommand struct {
	*cmds.CommandDescription
}

func NewModifyThreadCommand() (*ModifyThreadCommand, error) {
	return &ModifyThreadCommand{
		CommandDescription: cmds.NewCommandDescription(
			"modify-thread",
			cmds.WithShort("Modify an existing thread"),
			// Define flags and arguments as needed
		),
	}, nil
}

func (c *ModifyThreadCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
) error {
	apiKey := os.Getenv(apiKeyEnvVar)
	client := &http.Client{}

	threadID := ps["threadID"].(string) // Retrieve thread ID from flags/arguments
	// Construct the request body from flags/arguments
	request := assistants.ModifyThreadRequest{
		// Populate the request fields
	}

	thread, err := assistants.ModifyThread(client, baseURL, apiKey, threadID, request)
	if err != nil {
		return err
	}

	// Output the result
	output, err := json.Marshal(thread)
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

// DeleteThreadCommand uses BareCommand for simple execution without structured output
type DeleteThreadCommand struct {
	*cmds.CommandDescription
}

func NewDeleteThreadCommand() (*DeleteThreadCommand, error) {
	return &DeleteThreadCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete-thread",
			cmds.WithShort("Delete an existing thread"),
			// Define flags and arguments as needed
		),
	}, nil
}

func (c *DeleteThreadCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
) error {
	apiKey := os.Getenv(apiKeyEnvVar)
	client := &http.Client{}

	threadID := ps["threadID"].(string) // Retrieve thread ID from flags/arguments

	deleteResponse, err := assistants.DeleteThread(client, baseURL, apiKey, threadID)
	if err != nil {
		return err
	}

	// Output the result
	output, err := json.Marshal(deleteResponse)
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

var ThreadCmd = &cobra.Command{
	Use:   "thread",
	Short: "Manage OpenAI Threads",
}

func init() {
	listThreadsCmd, err := NewListThreadsCommand()
	if err != nil {
		panic(err)
	}
	listCobraCommand, err := cli.BuildCobraCommandFromGlazeCommand(listThreadsCmd)
	if err != nil {
		panic(err)
	}
	ThreadCmd.AddCommand(listCobraCommand)

	createThreadCmd, err := NewCreateThreadCommand()
	if err != nil {
		panic(err)
	}
	createCobraCommand, err := cli.BuildCobraCommandFromBareCommand(createThreadCmd)
	if err != nil {
		panic(err)
	}
	ThreadCmd.AddCommand(createCobraCommand)

	// Define and add ModifyThreadCommand
	modifyThreadCmd, err := NewModifyThreadCommand()
	if err != nil {
		panic(err)
	}
	modifyCobraCommand, err := cli.BuildCobraCommandFromBareCommand(modifyThreadCmd)
	if err != nil {
		panic(err)
	}
	ThreadCmd.AddCommand(modifyCobraCommand)

	// Define and add DeleteThreadCommand
	deleteThreadCmd, err := NewDeleteThreadCommand()
	if err != nil {
		panic(err)
	}
	deleteCobraCommand, err := cli.BuildCobraCommandFromBareCommand(deleteThreadCmd)
	if err != nil {
		panic(err)
	}
	ThreadCmd.AddCommand(deleteCobraCommand)
}
