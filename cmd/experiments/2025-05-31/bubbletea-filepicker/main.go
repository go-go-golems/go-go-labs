package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Get starting path from command line or use current directory
	startPath := "."
	if len(os.Args) > 1 {
		startPath = os.Args[1]
	}

	// Create file picker
	picker := NewFilePicker(startPath)

	// Create the program
	program := tea.NewProgram(picker, tea.WithAltScreen())

	// Run the program
	model, err := program.Run()
	if err != nil {
		fmt.Printf("Error running file picker: %v\n", err)
		os.Exit(1)
	}

	// Get the result
	fp := model.(*FilePicker)
	
	// Check for errors
	if err := fp.GetError(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Show result
	if selected, ok := fp.GetSelected(); ok {
		if len(selected) == 1 {
			fmt.Printf("Selected file: %s\n", selected[0])
		} else {
			fmt.Printf("Selected %d files:\n", len(selected))
			for _, file := range selected {
				fmt.Printf("  %s\n", file)
			}
		}
	} else {
		fmt.Println("No file selected (cancelled)")
	}
}
