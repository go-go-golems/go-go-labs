package cmd

import (
	"fmt"

	"github.com/go-go-golems/go-go-labs/cmd/apps/filter-binaries-from-git-history/pkg/ui"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	baseRef       string
	compareRef    string
	sizeThreshold int64
)

var (
	logLevel string
)

var rootCmd = &cobra.Command{
	Use:   "filter-binaries-from-git-history",
	Short: "Interactive tool to find and remove large binary files from git history",
	Long: `An interactive tool using charmbracelet TUI to analyze git history,
identify large binary files, and selectively remove them from git history.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Configure zerolog based on log level flag
		level, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			return fmt.Errorf("invalid log level: %w", err)
		}
		zerolog.SetGlobalLevel(level)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return ui.StartTUI(baseRef, compareRef, sizeThreshold)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the log level (debug, info, warn, error)")
	rootCmd.Flags().StringVar(&baseRef, "base", "origin/main", "Base reference to compare against")
	rootCmd.Flags().StringVar(&compareRef, "compare", "main", "Reference to compare with base")
	rootCmd.Flags().Int64Var(&sizeThreshold, "size-threshold", 1024*1024, "Size threshold in bytes for large files (default: 1MB)")
}
