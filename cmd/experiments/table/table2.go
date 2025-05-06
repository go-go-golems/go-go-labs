package main

import (
	"flag"
	"fmt"
	"os"
	
	"github.com/olekukonko/tablewriter"
)

func main() {
	var (
		total     int
		remaining int
	)

	flag.IntVar(&total, "total", 0, "Total number of tasks")
	flag.IntVar(&remaining, "remaining", 0, "Remaining tasks to complete")
	flag.Parse()

	if total == 0 {
		fmt.Println("Error: --total must be greater than 0")
		os.Exit(1)
	}

	slaPercentage := (total - remaining) * 100 / total

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Team Name", "Total Task", "Remaining Task", "SLA Status %"})
	table.SetBorders(tablewriter.Border{
		Left:   true,
		Top:    true,
		Right:  true,
		Bottom: true,
	})
	table.SetCenterSeparator("|")
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,   // Team Name
		tablewriter.ALIGN_CENTER, // Total Task
		tablewriter.ALIGN_CENTER, // Remaining Task
		tablewriter.ALIGN_CENTER, // SLA Status %
	})

	data := []string{
		"MyTeam",
		fmt.Sprintf("%6d", total),
		fmt.Sprintf("%8d", remaining),
		fmt.Sprintf("%6d%%", slaPercentage),
	}
	table.Append(data)
	table.Render()
}
