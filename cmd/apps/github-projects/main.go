package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

// GitHubConfig holds configuration for GitHub integration
type GitHubConfig struct {
	Token         string
	Owner         string
	ProjectNumber int
}

// LoadGitHubConfig loads configuration from environment variables
func LoadGitHubConfig() (*GitHubConfig, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	owner := os.Getenv("GITHUB_OWNER")
	if owner == "" {
		return nil, fmt.Errorf("GITHUB_OWNER environment variable is required")
	}

	projectNumberStr := os.Getenv("GITHUB_PROJECT_NUMBER")
	if projectNumberStr == "" {
		return nil, fmt.Errorf("GITHUB_PROJECT_NUMBER environment variable is required")
	}

	projectNumber, err := strconv.Atoi(projectNumberStr)
	if err != nil {
		return nil, fmt.Errorf("invalid GITHUB_PROJECT_NUMBER: %v", err)
	}

	return &GitHubConfig{
		Token:         token,
		Owner:         owner,
		ProjectNumber: projectNumber,
	}, nil
}

// Global configuration instance
var githubConfig *GitHubConfig

func main() {
	// Load GitHub configuration from environment variables
	var err error
	githubConfig, err = LoadGitHubConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Required environment variables:\n")
		fmt.Fprintf(os.Stderr, "  GITHUB_TOKEN - GitHub personal access token\n")
		fmt.Fprintf(os.Stderr, "  GITHUB_OWNER - GitHub organization or user\n")
		fmt.Fprintf(os.Stderr, "  GITHUB_PROJECT_NUMBER - Project number (integer)\n")
		os.Exit(1)
	}

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "github-graphql-cli",
		Short: "GitHub GraphQL CLI for Projects v2",
		Long: fmt.Sprintf(`A command-line tool for interacting with GitHub's GraphQL API,
specifically designed for Projects v2 (Beta). Supports querying projects,
managing project items, and updating custom fields.

Current configuration:
  GitHub Owner: %s
  Project Number: %d
  Token: %s...`, githubConfig.Owner, githubConfig.ProjectNumber, githubConfig.Token[:8]),
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
