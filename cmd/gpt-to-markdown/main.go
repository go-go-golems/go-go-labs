package main

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

// TODO add flag for only exporting the assistant responses
// TODO add flag for exporting the source blocks
// TODO add flag for adding the messages as comments in the source blocks (if we can detect their type, for example)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "gpt-to-markdown",
	Short:   "CSS is a tool to work with CSS files",
	Version: version,
}

func main() {
	helpSystem := help.NewHelpSystem()

	helpSystem.SetupCobraRootCommand(rootCmd)

	gptToMarkdownCmd, err := NewRenderCommand()
	cobra.CheckErr(err)
	command, err := cli.BuildCobraCommandFromWriterCommand(gptToMarkdownCmd)
	cobra.CheckErr(err)

	rootCmd.AddCommand(command)

	_ = rootCmd.Execute()
}
