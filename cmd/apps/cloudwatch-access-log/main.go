package main

import (
	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

func createRootCmd() *cobra.Command {
	helpSystem := help.NewHelpSystem()

	rootCmd := &cobra.Command{
		Use:   "log-parse",
		Short: "Parse log files",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			err := clay.InitLogger()
			cobra.CheckErr(err)
		},
	}

	helpSystem.SetupCobraRootCommand(rootCmd)

	err := clay.InitViper("clay", rootCmd)
	cobra.CheckErr(err)
	err = clay.InitLogger()
	cobra.CheckErr(err)

	return rootCmd
}

func main() {
	rootCmd := createRootCmd()

	// Register the LogParserCommand
	logParserCommand, err := NewLogParserCommand()
	cobra.CheckErr(err)
	logParserCmd, err := cli.BuildCobraCommandFromGlazeCommand(logParserCommand)
	cobra.CheckErr(err)
	rootCmd.AddCommand(logParserCmd)

	initSQLiteCommand, err := NewInitSQLiteCommand()
	cobra.CheckErr(err)
	initSQLiteCmd, err := cli.BuildCobraCommandFromWriterCommand(initSQLiteCommand)
	cobra.CheckErr(err)
	rootCmd.AddCommand(initSQLiteCmd)

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
