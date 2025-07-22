package cmd

import (
	"fmt"
)

// showProjectTitle displays the project title in a consistent format
func showProjectTitle(tm *TaskManager) {
	project, err := tm.GetDefaultProject()
	if err == nil {
		fmt.Printf("═══════════════════════════════════════════════════════════════════════════════\n")
		fmt.Printf("📋 PROJECT: %s\n", project.Name)
		fmt.Printf("═══════════════════════════════════════════════════════════════════════════════\n")
		if project.Guidelines != "" {
			fmt.Printf("Guidelines: %s\n\n", project.Guidelines)
		} else {
			fmt.Println()
		}
	}
}

// showTILNotesReminders displays helpful reminders for TIL and notes functionality
func showTILNotesReminders() {
	fmt.Printf("\nQuick Notes & Learning:\n")
	fmt.Printf("• Add insights: amp-tasks til create \"Title\" --content \"Learning\"\n")
	fmt.Printf("• Take notes: amp-tasks notes add <task-id> \"Progress note\"\n")
	fmt.Printf("• View project insights: amp-tasks til list\n")
}
