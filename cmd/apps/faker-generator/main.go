package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"

	// Import your command package
	"github.com/go-go-golems/go-go-labs/cmd/apps/faker-generator/cmds"
)

var rootCmd = &cobra.Command{
	Use:   "faker-generator",
	Short: "Processes Emrichen YAML with faker tags",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// If you need any global setup, put it here
	},
}

func main() {
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Create and add the process command
	processCmdInstance, err := cmds.NewProcessCommand()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error creating process command: %v\n", err)
		os.Exit(1)
	}
	cobraProcessCmd, err := cli.BuildCobraCommandFromGlazeCommand(processCmdInstance)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error converting process command to Cobra: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(cobraProcessCmd)

	// Create and add the generate command
	generateCmdInstance, err := cmds.NewGenerateCommand()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error creating generate command: %v\n", err)
		os.Exit(1)
	}
	cobraGenerateCmd, err := cli.BuildCobraCommandFromWriterCommand(generateCmdInstance)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error converting generate command to Cobra: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(cobraGenerateCmd)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
