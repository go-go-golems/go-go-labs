package main

import (
	"github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "kagi",
	Short:   "kagi is a tool to format structured data",
	Version: version,
}

func main() {
	helpSystem := help.NewHelpSystem()

	helpSystem.SetupCobraRootCommand(rootCmd)

	// load the pinocchio config file
	err := pkg.InitViper("pinocchio", rootCmd)
	cobra.CheckErr(err)

	enrichCmd, err := NewEnrichWebCommand()
	cobra.CheckErr(err)
	command, err := cli.BuildCobraCommandFromGlazeCommand(enrichCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	summarizeCmd, err := NewSummarizeCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommandFromWriterCommand(summarizeCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	_ = rootCmd.Execute()
}
