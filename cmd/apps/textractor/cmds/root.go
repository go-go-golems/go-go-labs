package cmds

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "textractor",
	Short: "A CLI tool for managing and analyzing text data",
	Long:  `textractor is a CLI tool for managing and analyzing text data`,
}

func Execute() {
	rootCmd.AddCommand(NewSubmitCommand())
	rootCmd.AddCommand(NewStatusCommand())
	rootCmd.AddCommand(NewFetchCommand())

	if err := rootCmd.Execute(); err != nil {
		println(err)
		return
	}
}
