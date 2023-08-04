package main

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-labs/cmd/gtm/cmds"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gtm",
	Short: "gtm is a tool to manage tags, triggers and variables in Google Tag Manager",
}

func main() {
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	tagsCommand, err := cmds.NewTagsCommand()
	cobra.CheckErr(err)
	command, err := cli.BuildCobraCommandFromGlazeCommand(tagsCommand)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	variablesCommand, err := cmds.NewVariablesCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommandFromGlazeCommand(variablesCommand)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	triggersCommand, err := cmds.NewTriggersCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommandFromGlazeCommand(triggersCommand)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	_ = rootCmd.Execute()
}
