package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

func main() {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "github-graphql-cli",
		Short: "GitHub GraphQL CLI for Projects v2",
		Long: `A command-line tool for interacting with GitHub's GraphQL API,
specifically designed for Projects v2 (Beta). Supports querying projects,
managing project items, and updating custom fields.`,
	}

	// Initialize help system
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Create and add commands
	commands := []func() error{
		func() error { return addViewerCommand(rootCmd) },
		func() error { return addProjectCommand(rootCmd) },
		func() error { return addFieldsCommand(rootCmd) },
		func() error { return addItemsCommand(rootCmd) },
		func() error { return addIssueCommand(rootCmd) },
		func() error { return addUpdateIssueCommand(rootCmd) },
		func() error { return addListProjectsCommand(rootCmd) },
		func() error { return addUpdateFieldCommand(rootCmd) },
		func() error { return addMCPCommand(rootCmd) },
	}

	for _, addCmd := range commands {
		if err := addCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating command: %v\n", err)
			os.Exit(1)
		}
	}

	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func addProjectCommand(rootCmd *cobra.Command) error {
	cmd, err := NewProjectCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(cobraCmd)
	return nil
}

func addFieldsCommand(rootCmd *cobra.Command) error {
	cmd, err := NewFieldsCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(cobraCmd)
	return nil
}

func addItemsCommand(rootCmd *cobra.Command) error {
	cmd, err := NewItemsCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(cobraCmd)
	return nil
}

func addIssueCommand(rootCmd *cobra.Command) error {
	cmd, err := NewIssueCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(cobraCmd)
	return nil
}

func addListProjectsCommand(rootCmd *cobra.Command) error {
	cmd, err := NewListProjectsCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(cobraCmd)
	return nil
}

func addViewerCommand(rootCmd *cobra.Command) error {
	cmd, err := NewViewerCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(cobraCmd)
	return nil
}

func addUpdateIssueCommand(rootCmd *cobra.Command) error {
	cmd, err := NewUpdateIssueCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(cobraCmd)
	return nil
}

func addUpdateFieldCommand(rootCmd *cobra.Command) error {
	cmd, err := NewUpdateFieldCommand()
	if err != nil {
		return err
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	if err != nil {
		return err
	}

	rootCmd.AddCommand(cobraCmd)
	return nil
}
