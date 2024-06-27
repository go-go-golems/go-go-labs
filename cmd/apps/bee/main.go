package main

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-labs/cmd/apps/bee/cmds"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "bee",
	Short: "Bee API CLI tool",
}

func main() {
	ctx := context.Background()

	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Conversation commands
	conversationCmd := &cobra.Command{
		Use:   "conversation",
		Short: "Manage conversations",
	}
	rootCmd.AddCommand(conversationCmd)

	listConversationsCmd, err := cmds.NewListConversationsCommand()
	cobra.CheckErr(err)
	glazedListConversationsCmd, err := cli.BuildCobraCommandFromGlazeCommand(listConversationsCmd)
	cobra.CheckErr(err)
	conversationCmd.AddCommand(glazedListConversationsCmd)

	getConversationCmd, err := cmds.NewGetConversationCommand()
	cobra.CheckErr(err)
	glazedGetConversationCmd, err := cli.BuildCobraCommandFromGlazeCommand(getConversationCmd)
	cobra.CheckErr(err)
	conversationCmd.AddCommand(glazedGetConversationCmd)

	deleteConversationCmd, err := cmds.NewDeleteConversationCommand()
	cobra.CheckErr(err)
	glazedDeleteConversationCmd, err := cli.BuildCobraCommandFromBareCommand(deleteConversationCmd)
	cobra.CheckErr(err)
	conversationCmd.AddCommand(glazedDeleteConversationCmd)

	endConversationCmd, err := cmds.NewEndConversationCommand()
	cobra.CheckErr(err)
	glazedEndConversationCmd, err := cli.BuildCobraCommandFromBareCommand(endConversationCmd)
	cobra.CheckErr(err)
	conversationCmd.AddCommand(glazedEndConversationCmd)

	retryConversationCmd, err := cmds.NewRetryConversationCommand()
	cobra.CheckErr(err)
	glazedRetryConversationCmd, err := cli.BuildCobraCommandFromBareCommand(retryConversationCmd)
	cobra.CheckErr(err)
	conversationCmd.AddCommand(glazedRetryConversationCmd)

	// Fact commands
	factCmd := &cobra.Command{
		Use:   "fact",
		Short: "Manage facts",
	}
	rootCmd.AddCommand(factCmd)

	listFactsCmd, err := cmds.NewListFactsCommand()
	cobra.CheckErr(err)
	glazedListFactsCmd, err := cli.BuildCobraCommandFromGlazeCommand(listFactsCmd)
	cobra.CheckErr(err)
	factCmd.AddCommand(glazedListFactsCmd)

	createFactCmd, err := cmds.NewCreateFactCommand()
	cobra.CheckErr(err)
	glazedCreateFactCmd, err := cli.BuildCobraCommandFromGlazeCommand(createFactCmd)
	cobra.CheckErr(err)
	factCmd.AddCommand(glazedCreateFactCmd)

	getFactCmd, err := cmds.NewGetFactCommand()
	cobra.CheckErr(err)
	glazedGetFactCmd, err := cli.BuildCobraCommandFromGlazeCommand(getFactCmd)
	cobra.CheckErr(err)
	factCmd.AddCommand(glazedGetFactCmd)

	updateFactCmd, err := cmds.NewUpdateFactCommand()
	cobra.CheckErr(err)
	glazedUpdateFactCmd, err := cli.BuildCobraCommandFromGlazeCommand(updateFactCmd)
	cobra.CheckErr(err)
	factCmd.AddCommand(glazedUpdateFactCmd)

	deleteFactCmd, err := cmds.NewDeleteFactCommand()
	cobra.CheckErr(err)
	glazedDeleteFactCmd, err := cli.BuildCobraCommandFromBareCommand(deleteFactCmd)
	cobra.CheckErr(err)
	factCmd.AddCommand(glazedDeleteFactCmd)

	// Todo commands
	todoCmd := &cobra.Command{
		Use:   "todo",
		Short: "Manage todos",
	}
	rootCmd.AddCommand(todoCmd)

	listTodosCmd, err := cmds.NewListTodosCommand()
	cobra.CheckErr(err)
	glazedListTodosCmd, err := cli.BuildCobraCommandFromGlazeCommand(listTodosCmd)
	cobra.CheckErr(err)
	todoCmd.AddCommand(glazedListTodosCmd)

	createTodoCmd, err := cmds.NewCreateTodoCommand()
	cobra.CheckErr(err)
	glazedCreateTodoCmd, err := cli.BuildCobraCommandFromGlazeCommand(createTodoCmd)
	cobra.CheckErr(err)
	todoCmd.AddCommand(glazedCreateTodoCmd)

	getTodoCmd, err := cmds.NewGetTodoCommand()
	cobra.CheckErr(err)
	glazedGetTodoCmd, err := cli.BuildCobraCommandFromGlazeCommand(getTodoCmd)
	cobra.CheckErr(err)
	todoCmd.AddCommand(glazedGetTodoCmd)

	updateTodoCmd, err := cmds.NewUpdateTodoCommand()
	cobra.CheckErr(err)
	glazedUpdateTodoCmd, err := cli.BuildCobraCommandFromGlazeCommand(updateTodoCmd)
	cobra.CheckErr(err)
	todoCmd.AddCommand(glazedUpdateTodoCmd)

	deleteTodoCmd, err := cmds.NewDeleteTodoCommand()
	cobra.CheckErr(err)
	glazedDeleteTodoCmd, err := cli.BuildCobraCommandFromBareCommand(deleteTodoCmd)
	cobra.CheckErr(err)
	todoCmd.AddCommand(glazedDeleteTodoCmd)

	err = rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}
