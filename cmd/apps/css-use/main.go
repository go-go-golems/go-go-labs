package main

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "css",
	Short:   "CSS is a tool to work with CSS files",
	Version: version,
}

func main() {
	helpSystem := help.NewHelpSystem()

	helpSystem.SetupCobraRootCommand(rootCmd)

	parseHtmlCmd, err := NewUsedCommand()
	cobra.CheckErr(err)
	command, err := cli.BuildCobraCommandFromGlazeCommand(parseHtmlCmd)
	cobra.CheckErr(err)

	rootCmd.AddCommand(command)

	extractCSSClassesCmd, err := NewDefinedCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommandFromGlazeCommand(extractCSSClassesCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	FindUnusedClassesCmd, err := NewFindUnusedClassesCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommandFromGlazeCommand(FindUnusedClassesCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	UsageCmd, err := NewFindUsageClassesCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommandFromGlazeCommand(UsageCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	_ = rootCmd.Execute()
}
