package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-labs/pkg/snakemake/doc"
	"github.com/spf13/cobra"
)

func main() {
	viewCmd, err := NewSnakemakeViewerCommand()
	if err != nil {
		fmt.Printf("Error creating view command: %v\n", err)
		os.Exit(1)
	}

	jobReportsCmd, err := NewJobReportsCommand()
	if err != nil {
		fmt.Printf("Error creating job-reports command: %v\n", err)
		os.Exit(1)
	}

	// Create a new help system
	helpSystem := help.NewHelpSystem()

	// Add documentation to the help system
	err = doc.AddDocToHelpSystem(helpSystem)
	if err != nil {
		fmt.Printf("Error adding documentation to help system: %v\n", err)
		os.Exit(1)
	}

	glazedViewCmd, err := cli.BuildCobraCommandFromGlazeCommand(viewCmd)
	if err != nil {
		fmt.Printf("Error creating Glazed view command: %v\n", err)
		os.Exit(1)
	}

	glazedJobReportsCmd, err := cli.BuildCobraCommandFromGlazeCommand(jobReportsCmd)
	if err != nil {
		fmt.Printf("Error creating Glazed job-reports command: %v\n", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{Use: "snakemake-viewer-cli"}
	rootCmd.AddCommand(glazedViewCmd)
	rootCmd.AddCommand(glazedJobReportsCmd)

	helpSystem.SetupCobraRootCommand(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
