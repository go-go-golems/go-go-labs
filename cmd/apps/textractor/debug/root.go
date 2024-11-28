package debug

import (
	"github.com/spf13/cobra"
)

// NewDebugCommand creates a new debug command with all subcommands
func NewDebugCommand() *cobra.Command {
	debugCmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug Textractor resources and processing",
	}

	// Add all debug subcommands
	debugCmd.AddCommand(
		newLambdaCommand(),
		newQueueCommand(),
		newS3Command(),
		newSNSCommand(),
		newOutputS3Command(),
		newMetricsCommand(),
		newTestCommand(),
		newCloudTrailCommand(),
		newNotificationsCommand(),
	)

	return debugCmd
}

// AddDebugCommands adds the debug command to the root command
func AddDebugCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(NewDebugCommand())
}
