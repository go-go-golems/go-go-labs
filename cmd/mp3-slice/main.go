package main

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
	"os"
)

func ensureDirExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mp3slicer",
		Short: "A tool to slice MP3 files into segments.",
	}

	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Convert SliceCommand from glazed to a Cobra command
	sliceCmdInstance, err := NewSliceCommand() // Assuming you've created this function
	if err != nil {
		cobra.CheckErr(err)
	}

	command, err := cli.BuildCobraCommandFromGlazeCommand(sliceCmdInstance)
	if err != nil {
		cobra.CheckErr(err)
	}

	rootCmd.AddCommand(command)

	// If there are other commands, repeat the process
	// ...

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
