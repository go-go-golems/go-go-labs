package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/apps/differential/pkg"
	"github.com/sergi/go-diff/diffmatchpatch"

	"os"
	"strings"
)

type Options struct {
	DryRun             bool
	ShowDiff           bool
	SaveBackup         bool
	AskForConfirmation bool
	DontOverWrite      bool
}

// applyDSL reads the DSL, applies all changes, and writes the result back to the file.
func applyDSL(dslJSON string, options Options) error {
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
	sourceLines := strings.Split(string(fileContent), "\n")

	// Apply each change
	for _, change := range dsl.Changes {
		sourceLines, err = pkg.ApplyChange(sourceLines, change)
		if err != nil {
			return err // Or handle the error as you prefer
		}
	}

	// Join the lines back into a single string
	updatedContent := strings.Join(sourceLines, "\n")

	if options.ShowDiff || options.AskForConfirmation {
		// Create a new diffmatchpatch object.
		dmp := diffmatchpatch.New()

		// Calculate the difference between the two texts.
		diffs := dmp.DiffMain(string(fileContent), updatedContent, false)

		// Display the differences.
		fmt.Println(dmp.DiffPrettyText(diffs))
	}

	if options.AskForConfirmation {
		fmt.Println("Do you want to apply this change? (y/n)")
		var answer string
		_, err := fmt.Scanln(&answer)
		if err != nil {
			return err
		}

		if answer != "y" {
			return nil
		}
	}

	if options.DryRun {
		return nil
	}

	if options.SaveBackup {
		// store original in backup file
		name, err := findNonexistentFile(dsl.Path, ".bak")
		if err != nil {
			return err
		}

		err = os.WriteFile(name, fileContent, 0644)
		if err != nil {
			return err
		}
	}

	path := dsl.Path
	if options.DontOverWrite {
		path, err = findNonexistentFile(dsl.Path, ".new")
		if err != nil {
			return err
		}
	}
	err = os.WriteFile(path, []byte(updatedContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

func findNonexistentFile(path string, suffix string) (string, error) {
	var name string
	i := 0
	for {
		name = path + fmt.Sprintf("%s.%03d", suffix, i)
		i++

		if i > 100 {
			return "", fmt.Errorf("could not find a backup file name")
		}

		// check that name does not exist
		_, err := os.Stat(name)
		if err == nil {
			continue
		}

		if os.IsNotExist(err) {
			break
		}

	}
	return name, nil
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

	options := Options{
		DryRun:             false,
		ShowDiff:           true,
		SaveBackup:         false,
		AskForConfirmation: true,
		DontOverWrite:      true,
	}
	err = applyDSL(string(dslJSON), options)
	if err != nil {
		fmt.Printf("Error applying DSL: %s\n", err)
	}
}
