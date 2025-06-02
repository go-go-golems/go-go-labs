package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cmd := NewWorkspaceManagerCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func NewWorkspaceManagerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace-manager",
		Short: "A tool for managing multi-repository workspaces",
		Long: `Workspace Manager helps you work with multiple related git repositories 
simultaneously by automating workspace setup, git operations, and status tracking.`,
	}

	cmd.AddCommand(
		NewDiscoverCommand(),
		NewListCommand(),
		NewCreateCommand(),
		NewStatusCommand(),
		NewTUICommand(),
		NewCommitCommand(),
		NewSyncCommand(),
		NewBranchCommand(),
		NewDiffCommand(),
		NewLogCommand(),
	)

	return cmd
}
