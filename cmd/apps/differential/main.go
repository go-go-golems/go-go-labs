package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Change represents a single change in the DSL.
type Change struct {
	Comment          string `json:"comment"`
	Action           string `json:"action"`
	Old              string `json:"old,omitempty"`
	New              string `json:"new,omitempty"`
	Content          string `json:"content,omitempty"`
	DestinationAbove string `json:"destination_above,omitempty"`
	DestinationBelow string `json:"destination_below,omitempty"`
}

// DSL represents the entire DSL document.
type DSL struct {
	Path    string   `json:"path"`
	Changes []Change `json:"changes"`
}

// findLocation searches for the location of a specific block of code within the source code.
// It returns the index of the first line of the code block in the source, or an error if not found.
func findLocation(sourceLines []string, locationLines []string) (int, error) {
	if len(locationLines) == 0 {
		return -1, errors.New("specified code block not found in the source")
	}

	// Join lines for easier matching and find the context in the source.
	sourceText := strings.Join(sourceLines, "\n")
	locationText := strings.Join(locationLines, "\n")

	locationIndex := strings.Index(sourceText, locationText)

	if locationIndex == -1 {
		return -1, errors.New("specified code block not found in the source")
	}

	// Calculate the line number in the source code.
	// Lines are 1-indexed, so we add 1 to the result.
	lineNumber := strings.Count(sourceText[:locationIndex], "\n")

	return lineNumber, nil
}

// applyChange applies a single change to the source lines based on the action specified in the change.
func applyChange(sourceLines []string, change Change) ([]string, error) {
	switch change.Action {
	case "replace", "delete", "move":
		contentLines := strings.Split(change.Old, "\n")
		if change.Action != "replace" {
			contentLines = strings.Split(change.Content, "\n")
		}
		startIdx, err := findLocation(sourceLines, contentLines)
		startIdx += 1
		if err != nil {
			return nil, err
		}
		endIdx := startIdx + len(contentLines)

		if change.Action == "replace" {
			newLines := strings.Split(change.New, "\n")
			sourceLines = append(sourceLines[:startIdx], append(newLines, sourceLines[endIdx:]...)...)
		} else if change.Action == "delete" {
			sourceLines = append(sourceLines[:startIdx], sourceLines[endIdx:]...)
		} else if change.Action == "move" {
			destination := change.DestinationAbove
			if destination == "" {
				destination = change.DestinationBelow
			}
			destLines := strings.Split(destination, "\n")
			moveIdx, err := findLocation(sourceLines, destLines)
			if err != nil {
				return nil, err
			}
			moveIdx += 1
			if change.DestinationBelow != "" {
				moveIdx += len(destLines)
			}
			segment := make([]string, endIdx-startIdx)
			copy(segment, sourceLines[startIdx:endIdx])
			sourceLines = append(sourceLines[:startIdx], sourceLines[endIdx:]...)
			sourceLines = append(sourceLines[:moveIdx], append(segment, sourceLines[moveIdx:]...)...)
		}

	case "insert":
		contentLines := strings.Split(change.Content, "\n")
		destination := change.DestinationAbove
		if destination == "" {
			destination = change.DestinationBelow
		}
		destLines := strings.Split(destination, "\n")
		insertIdx, err := findLocation(sourceLines, destLines)
		if err != nil {
			return nil, err
		}
		insertIdx += 1
		if change.DestinationBelow != "" {
			insertIdx += len(destLines)
		}
		sourceLines = append(sourceLines[:insertIdx], append(contentLines, sourceLines[insertIdx:]...)...)

	default:
		return nil, errors.New("unsupported action: " + change.Action)
	}

	return sourceLines, nil
}

// applyDSL reads the DSL, applies all changes, and writes the result back to the file.
func applyDSL(dslJSON string) error {
	var dsl DSL

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
		sourceLines, err = applyChange(sourceLines, change)
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
