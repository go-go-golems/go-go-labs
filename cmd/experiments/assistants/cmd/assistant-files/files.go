package assistant_files

import (
	"context"
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
)

const baseURL = "https://api.openai.com/v1"

// ListAssistantFilesCommand is a GlazedCommand for listing assistant files.
type ListAssistantFilesCommand struct {
	*cmds.CommandDescription
}

func NewListAssistantFilesCommand() (*ListAssistantFilesCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	return &ListAssistantFilesCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list-files",
			cmds.WithShort("List all assistant files"),
			cmds.WithLayers(glazedLayer),
		),
	}, nil
}

// CreateAssistantFileCommand is a GlazedCommand for creating an assistant file.
type CreateAssistantFileCommand struct {
	*cmds.CommandDescription
}

func NewCreateAssistantFileCommand() (*CreateAssistantFileCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	return &CreateAssistantFileCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create-file",
			cmds.WithShort("Create an assistant file"),
			cmds.WithLayers(glazedLayer),
		),
	}, nil
}

// RetrieveAssistantFileCommand is a GlazedCommand for retrieving an assistant file.
type RetrieveAssistantFileCommand struct {
	*cmds.CommandDescription
}

func NewRetrieveAssistantFileCommand() (*RetrieveAssistantFileCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	return &RetrieveAssistantFileCommand{
		CommandDescription: cmds.NewCommandDescription(
			"get-file",
			cmds.WithShort("Retrieve an assistant file"),
			cmds.WithLayers(glazedLayer),
		),
	}, nil
}

// DeleteAssistantFileCommand is a BareCommand for deleting an assistant file.
type DeleteAssistantFileCommand struct {
	*cmds.CommandDescription
}

func NewDeleteAssistantFileCommand() (*DeleteAssistantFileCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	return &DeleteAssistantFileCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete-file",
			cmds.WithShort("Delete an assistant file"),
			cmds.WithLayers(glazedLayer),
		),
	}, nil
}

func (c *ListAssistantFilesCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := &http.Client{}
	assistantID := ps["assistant_id"].(string) // Assuming assistant_id is a flag

	response, err := assistants.ListAssistantFiles(client, baseURL, apiKey, assistantID)
	if err != nil {
		return err
	}

	for _, file := range response.Data {
		row := types.NewRow(
			types.MRP("id", file.ID),
			types.MRP("object", file.Object),
			types.MRP("created_at", file.CreatedAt),
			types.MRP("assistant_id", file.AssistantID),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// Implement the Run method for CreateAssistantFileCommand
func (c *CreateAssistantFileCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := &http.Client{}
	assistantID := ps["assistant_id"].(string) // Assuming assistant_id is a flag
	fileID := ps["file_id"].(string)           // Assuming file_id is a flag

	request := assistants.CreateAssistantFileRequest{FileID: fileID}
	response, err := assistants.CreateAssistantFile(client, baseURL, apiKey, assistantID, request)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("id", response.ID),
		types.MRP("object", response.Object),
		types.MRP("created_at", response.CreatedAt),
		types.MRP("assistant_id", response.AssistantID),
	)
	return gp.AddRow(ctx, row)
}

// Implement the Run method for RetrieveAssistantFileCommand
func (c *RetrieveAssistantFileCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := &http.Client{}
	assistantID := ps["assistant_id"].(string) // Assuming assistant_id is a flag
	fileID := ps["file_id"].(string)           // Assuming file_id is a flag

	response, err := assistants.GetAssistantFile(client, baseURL, apiKey, assistantID, fileID)
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("id", response.ID),
		types.MRP("object", response.Object),
		types.MRP("created_at", response.CreatedAt),
		types.MRP("assistant_id", response.AssistantID),
	)
	return gp.AddRow(ctx, row)
}

// Implement the Run method for DeleteAssistantFileCommand
func (c *DeleteAssistantFileCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := &http.Client{}
	assistantID := ps["assistant_id"].(string) // Assuming assistant_id is a flag
	fileID := ps["file_id"].(string)           // Assuming file_id is a flag

	response, err := assistants.DeleteAssistantFile(client, baseURL, apiKey, assistantID, fileID)
	if err != nil {
		return err
	}

	// Output the result of the deletion
	fmt.Printf("File %s deleted: %t\n", response.ID, response.Deleted)
	return nil
}

var AssistantFilesCmd = &cobra.Command{
	Use:   "assistant-files",
	Short: "Manage OpenAI Assistant files",
}

func init() {
	listFilesCmd, err := NewListAssistantFilesCommand()
	cobra.CheckErr(err)
	command, err := cli.BuildCobraCommandFromGlazeCommand(listFilesCmd)
	cobra.CheckErr(err)
	AssistantFilesCmd.AddCommand(command)

	createFileCmd, err := NewCreateAssistantFileCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommandFromGlazeCommand(createFileCmd)
	cobra.CheckErr(err)
	AssistantFilesCmd.AddCommand(command)

	retrieveFileCmd, err := NewRetrieveAssistantFileCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommandFromGlazeCommand(retrieveFileCmd)
	cobra.CheckErr(err)
	AssistantFilesCmd.AddCommand(command)

	deleteFileCmd, err := NewDeleteAssistantFileCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommandFromBareCommand(deleteFileCmd)
	cobra.CheckErr(err)
	AssistantFilesCmd.AddCommand(command)
}
