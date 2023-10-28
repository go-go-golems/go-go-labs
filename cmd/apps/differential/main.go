package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/apps/differential/kmp"
	"os"
	"strings"
)

// Change represents a single change in the DSL.
type Change struct {
	Comment          string `json:"comment"`
	Action           Action `json:"action"`
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

type Action string

const (
	ActionReplace Action = "replace"
	ActionDelete  Action = "delete"
	ActionMove    Action = "move"
	ActionInsert  Action = "insert"
)

type ErrCodeBlock struct{}

func (e *ErrCodeBlock) Error() string {
	return "specified code block not found in the source"
}

// findLocation is a function that identifies the position of a specific block of
// code within a given source code. It takes two parameters: sourceLines and
// locationLines, both of which are slices of strings.
//
// sourceLines represents the entire source code split into lines.
// locationLines represents the block of code whose location is to be found.
//
// The function returns two values: the line number (or -1 if not found), and an error
// if the string was not found.
func findLocation(sourceLines []string, locationLines []string) (int, error) {
	if len(locationLines) == 0 {
		return -1, &ErrCodeBlock{}
	}

	l := kmp.KMPSearch(sourceLines, locationLines)
	if l == -1 {
		return -1, &ErrCodeBlock{}
	}

	return l, nil
}

// applyChange applies a specified change to a given set of source lines.
//
// It takes two parameters:
// - sourceLines: A slice of strings representing the source lines to be modified.
// - change: A Change struct detailing the change to be applied.
//
// The function supports four types of actions specified in the Change struct:
// - ActionReplace: Replaces the old content with the new content in the source lines.
// - ActionDelete: Removes the old content from the source lines.
// - ActionMove: Moves the old content to a new location in the source lines.
// - ActionInsert: Inserts new content at a specified location in the source lines.
//
// The function returns a slice of strings representing the modified source lines,
// and an error if the action is unsupported or if there is an issue locating the
// content or destination in the source lines.
func applyChange(sourceLines []string, change Change) ([]string, error) {
	switch change.Action {
	case ActionReplace, ActionDelete, ActionMove:
		contentLines := strings.Split(change.Old, "\n")
		if change.Action != ActionReplace {
			contentLines = strings.Split(change.Content, "\n")
		}
		startIdx, err := findLocation(sourceLines, contentLines)
		if err != nil {
			return nil, err
		}
		endIdx := startIdx + len(contentLines)

		if change.Action == ActionReplace {
			newLines := strings.Split(change.New, "\n")
			sourceLines = append(sourceLines[:startIdx], append(newLines, sourceLines[endIdx:]...)...)
		} else if change.Action == ActionDelete {
			sourceLines = append(sourceLines[:startIdx], sourceLines[endIdx:]...)
		} else if change.Action == ActionMove {
			destination := change.DestinationAbove
			if destination == "" {
				destination = change.DestinationBelow
			}
			destLines := strings.Split(destination, "\n")
			moveIdx, err := findLocation(sourceLines, destLines)
			if err != nil {
				return nil, err
			}
			if change.DestinationBelow != "" {
				moveIdx += len(destLines)
			}
			segment := make([]string, endIdx-startIdx)
			copy(segment, sourceLines[startIdx:endIdx])
			sourceLines = append(sourceLines[:startIdx], sourceLines[endIdx:]...)
			sourceLines = append(sourceLines[:moveIdx], append(segment, sourceLines[moveIdx:]...)...)
		}

	case ActionInsert:
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
		if change.DestinationBelow != "" {
			insertIdx += len(destLines)
		}
		sourceLines = append(sourceLines[:insertIdx], append(contentLines, sourceLines[insertIdx:]...)...)

	default:
		return nil, errors.New("unsupported action: " + string(change.Action))
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
