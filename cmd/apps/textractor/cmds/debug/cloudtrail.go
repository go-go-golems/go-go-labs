package debug

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
)

func newCloudTrailCommand() *cobra.Command {
	cloudTrailDebugCmd := &cobra.Command{
		Use:   "cloudtrail",
		Short: "Debug CloudTrail logs",
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			fmt.Printf("🔍 Debugging CloudTrail logs\n")
			fmt.Printf("CloudTrail Log Group: %s\n", resources.CloudTrailLogGroup)

			// Get the last hour of CloudTrail logs
			endTime := time.Now()
			startTime := endTime.Add(-1 * time.Hour)

			err = runAWSCommand("logs", "filter-log-events",
				"--log-group-name", resources.CloudTrailLogGroup,
				"--start-time", fmt.Sprintf("%d", startTime.UnixNano()/1000000),
				"--end-time", fmt.Sprintf("%d", endTime.UnixNano()/1000000))
			if err != nil {
				log.Printf("Failed to get CloudTrail logs: %v", err)
			}
		},
	}

	// Add lookup subcommand
	cloudTrailDebugCmd.AddCommand(&cobra.Command{
		Use:   "lookup [event-name]",
		Short: "Look up specific CloudTrail events",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			eventName := args[0]
			fmt.Printf("🔍 Looking up CloudTrail events: %s\n", eventName)

			// Get the last 24 hours of events
			endTime := time.Now()
			startTime := endTime.Add(-24 * time.Hour)

			err = runAWSCommand("logs", "filter-log-events",
				"--log-group-name", resources.CloudTrailLogGroup,
				"--start-time", fmt.Sprintf("%d", startTime.UnixNano()/1000000),
				"--end-time", fmt.Sprintf("%d", endTime.UnixNano()/1000000),
				"--filter-pattern", fmt.Sprintf(`{ $.eventName = "%s" }`, eventName))
			if err != nil {
				log.Printf("Failed to look up CloudTrail events: %v", err)
			}
		},
	})

	return cloudTrailDebugCmd
}
