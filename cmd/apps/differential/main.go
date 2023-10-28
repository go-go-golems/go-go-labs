package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/apps/differential/pkg"
	"os"
)

// applyDSL reads the DSL, applies all changes, and writes the result back to the file.
func applyDSL(dslJSON string) error {
	var dsl pkg.DSL

	// Parse the JSON input
	err := json.Unmarshal([]byte(dslJSON), &dsl)
	if err != nil {
		return err
	}

	// Read the source file
	fileContent, err := os.ReadFile(dsl.Path)
	if err != nil {
		return err
	}

	// Split the file content into lines
	sourceLines := splitIntoLines(string(fileContent))

	// Apply each change
	for _, change := range dsl.Changes {
		sourceLines, err = pkg.ApplyChange(sourceLines, change)
		if err != nil {
			return err // Or handle the error as you prefer
		}
	}

	// Join the lines back into a single string
	updatedContent := joinLines(sourceLines)

	// Save the modified source file
	err = os.WriteFile(dsl.Path, []byte(updatedContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

// splitIntoLines splits content into lines.
func splitIntoLines(content string) []string {
	// TODO: Implement splitting content into lines.
	return nil
}

// joinLines joins lines into content.
func joinLines(lines []string) string {
	// TODO: Implement joining lines into a single string.
	return ""
}

func main() {
	// Assuming the DSL is passed as a file argument to the program
	if len(os.Args) < 2 {
		fmt.Println("Error: No DSL file provided")
		return
	}

	dslFile := os.Args[1]
	dslJSON, err := os.ReadFile(dslFile)
	if err != nil {
		fmt.Printf("Error reading DSL file: %s\n", err)
		return
	}

	err = applyDSL(string(dslJSON))
	if err != nil {
		fmt.Printf("Error applying DSL: %s\n", err)
	}
}
