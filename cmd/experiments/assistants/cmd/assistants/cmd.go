package assistants

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
)

var AssistantCmd = &cobra.Command{
	Use:   "assistant",
	Short: "Manage OpenAI Assistants",
}

func init() {
	listAssistantsCmd, err := NewListAssistantsCommand()
	if err != nil {
		panic(err)
	}
	cobraCommand, err := cli.BuildCobraCommandFromGlazeCommand(listAssistantsCmd)
	if err != nil {
		panic(err)
	}
	AssistantCmd.AddCommand(cobraCommand)

	createAssistantCmd, err := NewCreateAssistantCommand()
	if err != nil {
		panic(err)
	}
	cobraCommand, err = cli.BuildCobraCommandFromGlazeCommand(createAssistantCmd)
	if err != nil {
		panic(err)
	}
	AssistantCmd.AddCommand(cobraCommand)

	retrieveAssistantCmd, err := NewRetrieveAssistantCommand()
	if err != nil {
		panic(err)
	}
	cobraCommand, err = cli.BuildCobraCommandFromGlazeCommand(retrieveAssistantCmd)
	if err != nil {
		panic(err)
	}
	AssistantCmd.AddCommand(cobraCommand)

	modifyAssistantCmd, err := NewModifyAssistantCommand()
	if err != nil {
		panic(err)
	}
	cobraCommand, err = cli.BuildCobraCommandFromGlazeCommand(modifyAssistantCmd)
	if err != nil {
		panic(err)
	}
	AssistantCmd.AddCommand(cobraCommand)

	deleteAssistantCmd, err := NewDeleteAssistantCommand()
	if err != nil {
		panic(err)
	}
	cobraCommand, err = cli.BuildCobraCommandFromBareCommand(deleteAssistantCmd)
	if err != nil {
		panic(err)
	}
	AssistantCmd.AddCommand(cobraCommand)
}
