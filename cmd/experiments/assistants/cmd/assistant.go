package cmd

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/assistants/cmd/assistants"
	"github.com/spf13/cobra"
)

var AssistantCmd = &cobra.Command{
	Use:   "assistant",
	Short: "Manage OpenAI Assistants",
}

func init() {
	listAssistantsCmd, err := assistants.NewListAssistantsCommand()
	if err != nil {
		panic(err)
	}
	cobraCommand, err := cli.BuildCobraCommandFromGlazeCommand(listAssistantsCmd)
	if err != nil {
		panic(err)
	}
	AssistantCmd.AddCommand(cobraCommand)

	createAssistantCmd, err := assistants.NewCreateAssistantCommand()
	if err != nil {
		panic(err)
	}
	cobraCommand, err = cli.BuildCobraCommandFromGlazeCommand(createAssistantCmd)
	if err != nil {
		panic(err)
	}
	AssistantCmd.AddCommand(cobraCommand)

	retrieveAssistantCmd, err := assistants.NewRetrieveAssistantCommand()
	if err != nil {
		panic(err)
	}
	cobraCommand, err = cli.BuildCobraCommandFromGlazeCommand(retrieveAssistantCmd)
	if err != nil {
		panic(err)
	}
	AssistantCmd.AddCommand(cobraCommand)

	modifyAssistantCmd, err := assistants.NewModifyAssistantCommand()
	if err != nil {
		panic(err)
	}
	cobraCommand, err = cli.BuildCobraCommandFromGlazeCommand(modifyAssistantCmd)
	if err != nil {
		panic(err)
	}
	AssistantCmd.AddCommand(cobraCommand)

	deleteAssistantCmd, err := assistants.NewDeleteAssistantCommand()
	if err != nil {
		panic(err)
	}
	cobraCommand, err = cli.BuildCobraCommandFromBareCommand(deleteAssistantCmd)
	if err != nil {
		panic(err)
	}
	AssistantCmd.AddCommand(cobraCommand)
}
