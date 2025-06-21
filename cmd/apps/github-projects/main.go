package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

// EnsureGitHubConfig loads config if not already loaded
func EnsureGitHubConfig() error {
	if githubConfig != nil {
		return nil
	}

	var err error
	githubConfig, err = LoadGitHubConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %v\nRequired environment variables:\n  GITHUB_TOKEN - GitHub personal access token\n  GITHUB_OWNER - GitHub organization or user\n  GITHUB_PROJECT_NUMBER - Project number (integer)", err)
	}
	return nil
}

// GetDefaultOwner returns default owner from env var
func GetDefaultOwner() string {
	return os.Getenv("GITHUB_OWNER")
}

// GetDefaultProjectNumber returns default project number from env var
func GetDefaultProjectNumber() int {
	projectNumberStr := os.Getenv("GITHUB_PROJECT_NUMBER")
	if projectNumberStr == "" {
		return 0
	}
	projectNumber, err := strconv.Atoi(projectNumberStr)
	if err != nil {
		return 0
	}
	return projectNumber
}

// Global configuration instance
var githubConfig *GitHubConfig

func main() {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "github-graphql-cli",
		Short: "GitHub GraphQL CLI for Projects v2",
		Long: `A command-line tool for interacting with GitHub's GraphQL API,
specifically designed for Projects v2 (Beta). Supports querying projects,
managing project items, and updating custom fields.

Configuration is loaded from environment variables when needed:
  GITHUB_TOKEN - GitHub personal access token
  GITHUB_OWNER - GitHub organization or user
  GITHUB_PROJECT_NUMBER - Project number (integer)`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			log.Info().Msg("Initializing logger")
			err := logging.InitLoggerFromViper()
			if err != nil {
				return err
			}
			log.Info().Msg("Logger initialized")
			return nil
		},
	}
	err := logging.AddLoggingLayerToRootCommand(rootCmd, "github-projects")
	cobra.CheckErr(err)

	logging.InitViper("github-projects", rootCmd)

	err = viper.BindPFlags(rootCmd.PersistentFlags())
	cobra.CheckErr(err)

	err = logging.InitLoggerFromViper()
	cobra.CheckErr(err)

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
