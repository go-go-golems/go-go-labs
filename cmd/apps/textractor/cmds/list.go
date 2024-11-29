package cmds

import (
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/pkg"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Textract jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse flags
			since, _ := cmd.Flags().GetString("since")
			status, _ := cmd.Flags().GetString("status")

			stateLoader := pkg.NewStateLoader()
			resources, err := stateLoader.LoadStateFromCommand(cmd)
			if err != nil {
				return fmt.Errorf("failed to load state: %w", err)
			}

			// Initialize AWS session
			sess := session.Must(session.NewSession(&aws.Config{
				Region: aws.String(resources.Region),
			}))

			jobClient := pkg.NewJobClient(sess, resources.JobsTable)

			var opts pkg.ListJobsOptions
			if since != "" {
				// Try RFC3339 format first
				t, err := time.Parse(time.RFC3339, since)
				if err != nil {
					// If that fails, try simple date format
					t, err = time.Parse("2006-01-02", since)
					if err != nil {
						return fmt.Errorf("invalid time format for --since flag, use YYYY-MM-DD or RFC3339 format: %w", err)
					}
				}
				opts.Since = &t
			}
			opts.Status = status

			jobs, err := jobClient.ListJobs(opts)
			if err != nil {
				return fmt.Errorf("failed to list jobs: %w", err)
			}

			// Sort jobs by submission time
			sort.Slice(jobs, func(i, j int) bool {
				return jobs[i].SubmittedAt.After(jobs[j].SubmittedAt)
			})

			// Print jobs
			for _, job := range jobs {
				fmt.Printf("Job ID: %s\n", job.JobID)
				fmt.Printf("  Document: %s\n", job.DocumentKey)
				fmt.Printf("  Status: %s\n", job.Status)
				fmt.Printf("  Submitted: %s\n", job.SubmittedAt.Format(time.RFC3339))
				fmt.Printf("  Textract ID: %s\n", job.TextractID)
				fmt.Printf("  Result Key: %s\n", job.ResultKey)
				if job.CompletedAt != nil {
					fmt.Printf("  Completed: %s\n", job.CompletedAt.Format(time.RFC3339))
				}
				if job.Error != "" {
					fmt.Printf("  Error: %s\n", job.Error)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().String("since", "", "Only show jobs submitted after this time (RFC3339 format)")
	cmd.Flags().String("status", "", "Filter by job status (UPLOADING, SUBMITTED, PROCESSING, COMPLETED, FAILED, ERROR)")
	return cmd
}
