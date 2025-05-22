package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	// Setup viper for configuration
	viper.SetEnvPrefix("N8N_CLI")
	viper.AutomaticEnv()

	// Create the root command
	rootCmd := &cobra.Command{
		Use:   "n8n-cli",
		Short: "CLI for interacting with n8n REST API",
		Long: `A command line tool for managing n8n workflows via the REST API.

This tool allows you to create, modify, and manage workflows programmatically,
including adding nodes, connecting them, and setting node parameters.

All commands require an API key, which can be obtained from the n8n UI under
Settings → n8n API → "Create an API key".
`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Bind viper to cobra flags
			_ = viper.BindPFlags(cmd.Flags())

			// Initialize the logger from viper settings
			err := logging.InitLoggerFromViper()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
			}
		},
	}

	// Add logging layer to root command
	err := logging.AddLoggingLayerToRootCommand(rootCmd, "n8n-cli")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding logging layer: %v\n", err)
		os.Exit(1)
	}

	// Initialize help system
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Create all commands
	commands, err := createCommands()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating commands: %v\n", err)
		os.Exit(1)
	}

	// Add commands to root
	for _, cmd := range commands {
		// BuildCobraCommandFromCommand handles the dispatch to the appropriate builder
		var cobraCmd *cobra.Command
		var err error

		// Fall back to attempting direct conversion
		cobraCmd, err = cli.BuildCobraCommandFromCommand(cmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building command: %v\n", err)
			os.Exit(1)
		}

		rootCmd.AddCommand(cobraCmd)
	}

	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func createCommands() ([]cmds.Command, error) {
	var commands []cmds.Command

	// NEW STYLE COMMANDS
	// List workflows command
	listWorkflowsCmd, err := NewListWorkflowsCommand()
	if err != nil {
		return nil, fmt.Errorf("error creating list-workflows command: %w", err)
	}
	commands = append(commands, listWorkflowsCmd)

	// Get workflow command
	getWorkflowCmd, err := NewGetWorkflowCommand()
	if err != nil {
		return nil, fmt.Errorf("error creating get-workflow command: %w", err)
	}
	commands = append(commands, getWorkflowCmd)

	// Create workflow command
	createWorkflowCmd, err := NewCreateWorkflowCommand()
	if err != nil {
		return nil, fmt.Errorf("error creating create-workflow command: %w", err)
	}
	commands = append(commands, createWorkflowCmd)

	// Get available node types command
	getNodesCmd, err := NewGetNodesCommand()
	if err != nil {
		return nil, fmt.Errorf("error creating get-nodes command: %w", err)
	}
	commands = append(commands, getNodesCmd)

	// Add node to workflow command
	addNodeCmd, err := NewAddNodeCommand()
	if err != nil {
		return nil, fmt.Errorf("error creating add-node command: %w", err)
	}
	commands = append(commands, addNodeCmd)

	// Connect nodes in workflow command
	connectNodesCmd, err := NewConnectNodesCommand()
	if err != nil {
		return nil, fmt.Errorf("error creating connect-nodes command: %w", err)
	}
	commands = append(commands, connectNodesCmd)

	// List executions command
	listExecutionsCmd, err := NewListExecutionsCommand()
	if err != nil {
		return nil, fmt.Errorf("error creating list-executions command: %w", err)
	}
	commands = append(commands, listExecutionsCmd)

	// Get execution details command
	getExecutionCmd, err := NewGetExecutionCommand()
	if err != nil {
		return nil, fmt.Errorf("error creating get-execution command: %w", err)
	}
	commands = append(commands, getExecutionCmd)

	return commands, nil
}
