package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	totalTasks    int
	remainingTasks int
)

var rootCmd = &cobra.Command{
	Use:   "sla-table",
	Short: "Generates an SLA status table",
	Long:  `Generates a formatted table displaying team SLA status based on total and remaining tasks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if totalTasks <= 0 {
			return errors.New("--total must be a positive integer")
		}
		if remainingTasks < 0 {
			return errors.New("--remaining must be a non-negative integer")
		}
        if remainingTasks > totalTasks {
            return errors.Errorf("--remaining (%d) cannot be greater than --total (%d)", remainingTasks, totalTasks)
        }

		// Calculate SLA Status
        // Avoid division by zero, already checked totalTasks > 0
		slaPercentage := float64(totalTasks-remainingTasks) / float64(totalTasks) * 100
        slaString := fmt.Sprintf("%.0f%%", slaPercentage) // Format as integer percentage

		// Prepare table data
		data := [][]string{
			{"MyTeam", strconv.Itoa(totalTasks), strconv.Itoa(remainingTasks), slaString},
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Team Name", "Total Task", "Remaining Task", "SLA Status %"})
        table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
        table.SetAlignment(tablewriter.ALIGN_CENTER) // Default alignment
        table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER}) // Align first column left
        table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
        table.SetCenterSeparator("|")
        table.SetRowLine(true) // Enable row line
        table.SetHeaderLine(true) // Enable header line separator


		table.AppendBulk(data)
		table.Render()

		return nil
	},
}

func init() {
	rootCmd.Flags().IntVar(&totalTasks, "total", 0, "Total number of tasks (required)")
	rootCmd.Flags().IntVar(&remainingTasks, "remaining", 0, "Number of remaining tasks (required)")
	rootCmd.MarkFlagRequired("total")
	rootCmd.MarkFlagRequired("remaining")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}

