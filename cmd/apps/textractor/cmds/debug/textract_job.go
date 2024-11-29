package debug

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func addTextractJobCommand(debugCmd *cobra.Command) {
	debugCmd.AddCommand(&cobra.Command{
		Use:   "textract-job [jobId]",
		Short: "Check status of a Textract job",
		Long:  "Shows the current status and details of a Textract document analysis job",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			jobID := args[0]
			fmt.Printf("🔍 Checking status of Textract job: %s\n", jobID)

			// Get job status
			err = runAWSCommand("textract", "get-document-analysis",
				"--job-id", jobID)
			if err != nil {
				log.Printf("Failed to get job status: %v", err)
			}

			// List all results pages
			fmt.Println("\nListing all result pages:")
			err = runAWSCommand("textract", "get-document-analysis",
				"--job-id", jobID,
				"--max-results", "1000")
			if err != nil {
				log.Printf("Failed to list results: %v", err)
			}
		},
	})
}
