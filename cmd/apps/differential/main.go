package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/apps/differential/pkg"
	"github.com/sergi/go-diff/diffmatchpatch"

	"os"
	"strings"
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

	// Create a new diffmatchpatch object.
	dmp := diffmatchpatch.New()

	// Calculate the difference between the two texts.
	diffs := dmp.DiffMain(string(fileContent), updatedContent, false)

	// Display the differences.
	fmt.Println(dmp.DiffPrettyText(diffs))

	// store original in backup file
	//name := dsl.Path + ".bak"
	//i := 0
	//for {
	//	// check that name does not exist
	//	_, err = os.Stat(name)
	//	if err == nil {
	//		return fmt.Errorf("backup file %s already exists", name)
	//	}
	//
	//	if os.IsNotExist(err) {
	//		break
	//	}
	//
	//	name = dsl.Path + fmt.Sprintf(".bak%d", i)
	//	i++
	//
	//	if i > 100 {
	//		return fmt.Errorf("could not find a backup file name")
	//	}
	//}
	//
	//err = os.WriteFile(name, fileContent, 0644)
	//if err != nil {
	//	return err
	//}

	err = os.WriteFile(dsl.Path+".new", []byte(updatedContent), 0644)
	if err != nil {
		return err
	}

	return nil
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
