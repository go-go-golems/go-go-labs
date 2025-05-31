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
		fmt.Printf("Selected file: %s\n", selected)
	} else {
		fmt.Println("No file selected (cancelled)")
	}
}
