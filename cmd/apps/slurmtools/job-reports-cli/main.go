package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

func main() {
	cmd, err := NewJobReportsCommand()
	if err != nil {
		fmt.Printf("Error creating command: %v\n", err)
		os.Exit(1)
	}

	// Create a new help system
	helpSystem := help.NewHelpSystem()

	glazedCmd, err := cli.BuildCobraCommandFromGlazeCommand(cmd)
	if err != nil {
		fmt.Printf("Error creating Glazed command: %v\n", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{Use: "job-reports-cli"}
	rootCmd.AddCommand(glazedCmd)

	helpSystem.SetupCobraRootCommand(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
