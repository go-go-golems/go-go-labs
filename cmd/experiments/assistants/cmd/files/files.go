package files

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/pkg"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

type ListFilesCommand struct {
	*cmds.CommandDescription
}

func NewListFilesCommand() (*ListFilesCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	return &ListFilesCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list-files",
			cmds.WithShort("List all files"),
			cmds.WithLayers(glazedLayer),
		),
	}, nil
}

func (c *ListFilesCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := &http.Client{}

	filesResponse, err := pkg.ListFiles(client, "https://api.openai.com/v1/", apiKey)
	if err != nil {
		return err
	}

	for _, file := range filesResponse.Data {
		row := types.NewRow(
			types.MRP("id", file.ID),
			types.MRP("bytes", file.Bytes),
			types.MRP("created_at", file.CreatedAt),
			types.MRP("filename", file.Filename),
			types.MRP("object", file.Object),
			types.MRP("purpose", file.Purpose),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

type CreateFileCommand struct {
	*cmds.CommandDescription
}

func NewCreateFileCommand() (*CreateFileCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	return &CreateFileCommand{
		CommandDescription: cmds.NewCommandDescription(
			"create-file",
			cmds.WithShort("Create a new file"),
			cmds.WithLayers(glazedLayer),
			cmds.WithFlags(
			// Define flags for file content and purpose
			),
		),
	}, nil
}

func (c *CreateFileCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := &http.Client{}

	// Retrieve flags for file content and purpose
	fileContent := ps["fileContent"].([]byte)
	purpose := ps["purpose"].(string)

	file, err := pkg.CreateFile(client, "https://api.openai.com/v1/", apiKey, pkg.CreateFileRequest{
		File:    fileContent,
		Purpose: purpose,
	})
	if err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("id", file.ID),
		types.MRP("bytes", file.Bytes),
		types.MRP("created_at", file.CreatedAt),
		types.MRP("filename", file.Filename),
		types.MRP("object", file.Object),
		types.MRP("purpose", file.Purpose),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}

	return nil
}

type DeleteFileCommand struct {
	*cmds.CommandDescription
}

func NewDeleteFileCommand() (*DeleteFileCommand, error) {
	return &DeleteFileCommand{
		CommandDescription: cmds.NewCommandDescription(
			"delete-file",
			cmds.WithShort("Delete a file"),
			cmds.WithArguments(
			// Define argument for file ID
			),
		),
	}, nil
}

func (c *DeleteFileCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
) error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := &http.Client{}

	// Retrieve argument for file ID
	fileID := ps["fileID"].(string)

	deleteResponse, err := pkg.DeleteFile(client, "https://api.openai.com/v1/", apiKey, fileID)
	if err != nil {
		return err
	}

	if !deleteResponse.Deleted {
		return fmt.Errorf("failed to delete file with ID: %s", fileID)
	}

	fmt.Printf("File with ID %s deleted successfully\n", fileID)
	return nil
}

var FilesCmd = &cobra.Command{
	Use:   "files",
	Short: "Manage files",
}

func init() {
	listCommand, err := NewListFilesCommand()
	if err != nil {
		panic(err)
	}
	cobraCommand, err := cli.BuildCobraCommandFromGlazeCommand(listCommand)
	if err != nil {
		panic(err)
	}
	FilesCmd.AddCommand(cobraCommand)

	createCommand, err := NewCreateFileCommand()
	if err != nil {
		panic(err)
	}
	cobraCommand, err = cli.BuildCobraCommandFromGlazeCommand(createCommand)
	if err != nil {
		panic(err)
	}

	FilesCmd.AddCommand(cobraCommand)

	deleteCommand, err := NewDeleteFileCommand()
	if err != nil {
		panic(err)
	}
	cobraCommand, err = cli.BuildCobraCommandFromBareCommand(deleteCommand)
	if err != nil {
		panic(err)
	}
	FilesCmd.AddCommand(cobraCommand)
}
