package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go-go-golems/go-go-labs/cmd/github-projects/cmds"
	"github.com/go-go-golems/go-go-labs/cmd/github-projects/config"
	"github.com/go-go-golems/go-go-labs/pkg/github/mcp"
)

// EnsureGitHubConfig loads config if not already loaded
func EnsureGitHubConfig() error {
	if githubConfig != nil {
		return nil
	}

	var err error
	githubConfig, err = config.LoadGitHubConfig()
	if err != nil {
		return fmt.Errorf("configuration error: %v\nRequired environment variables:\n  GITHUB_TOKEN - GitHub personal access token\n  GITHUB_OWNER - GitHub organization or user\n  GITHUB_PROJECT_NUMBER - Project number (integer)", err)
	}
	return nil
}

// Global configuration instance
var githubConfig *config.GitHubConfig

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
		func() error { return addProjectInfoCommand(rootCmd) },
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
	cmd, err := cmds.NewProjectCommand()
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
	cmd, err := cmds.NewFieldsCommand()
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
	cmd, err := cmds.NewItemsCommand()
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
	cmd, err := cmds.NewIssueCommand()
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
	cmd, err := cmds.NewListProjectsCommand()
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
	cmd, err := cmds.NewViewerCommand()
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
	cmd, err := cmds.NewUpdateIssueCommand()
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
	cmd, err := cmds.NewUpdateFieldCommand()
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

func addProjectInfoCommand(rootCmd *cobra.Command) error {
	cmd, err := cmds.NewProjectInfoCommand()
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

func addMCPCommand(rootCmd *cobra.Command) error {
	handlers := &mcp.ToolHandlers{
		ReadProjectItems:      ReadProjectItemsHandler,
		AddProjectItem:        AddProjectItemHandler,
		UpdateProjectItem:     UpdateProjectItemHandler,
		AddProjectItemComment: AddProjectItemCommentHandler,
		GetProjectInfo:        GetProjectInfoHandler,
	}
	return mcp.AddMCPCommand(rootCmd, handlers)
}
