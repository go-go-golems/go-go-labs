package main

import (
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	formpkg "github.com/go-go-golems/go-go-labs/cmd/apps/form-generator/pkg"
	uhoh_doc "github.com/go-go-golems/uhoh/pkg/doc"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "form-generator",
	Short: "Generate a Google Form from a Uhoh Wizard DSL file",
}

func main() {
	helpSystem := help.NewHelpSystem()
	if err := uhoh_doc.AddDocToHelpSystem(helpSystem); err != nil {
		cobra.CheckErr(err)
	}
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	genCmd, err := formpkg.NewGenerateCommand()
	cobra.CheckErr(err)
	rootCmd.AddCommand(genCmd)

	fetchCmd, err := formpkg.NewFetchCommand()
	cobra.CheckErr(err)
	rootCmd.AddCommand(fetchCmd)

	fetchSubmissionsCmd, err := formpkg.NewFetchSubmissionsCommand()
	cobra.CheckErr(err)
	rootCmd.AddCommand(fetchSubmissionsCmd)

	_ = rootCmd.Execute()
}
