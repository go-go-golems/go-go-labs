package main

import (
	"encoding/json"
	"fmt"
	json2 "github.com/go-go-golems/glazed/pkg/helpers/json"
	yaml2 "github.com/go-go-golems/glazed/pkg/helpers/yaml"
	"github.com/go-go-golems/go-go-labs/cmd/apps/differential/pkg"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

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

func tryLoading(dslJSON string) (*pkg.DSL, error) {
	var dsl pkg.DSL
	dslJSON_ := json2.SanitizeJSONString(dslJSON, true)
	if dslJSON_ != "" {
		err := json.Unmarshal([]byte(dslJSON_), &dsl)
		if err == nil {
			return &dsl, nil
		}
	}

	dslJSON_ = json2.SanitizeJSONString(dslJSON, false)
	if dslJSON_ != "" {
		err := json.Unmarshal([]byte(dslJSON_), &dsl)
		if err == nil {
			return &dsl, nil
		}
	}

	dslYAML := yaml2.Clean(dslJSON, true)
	if dslYAML != "" {
		err := yaml.Unmarshal([]byte(dslYAML), &dsl)
		if err == nil {
			return &dsl, nil
		}
	}

	dslYAML = yaml2.Clean(dslJSON, false)
	if dslYAML != "" {
		err := yaml.Unmarshal([]byte(dslYAML), &dsl)
		if err == nil {
			return &dsl, nil
		}
	}

	return nil, fmt.Errorf("could not parse DSL")
}

// applyDSL reads the DSL, applies all changes, and writes the result back to the file.
func applyDSL(dslJSON string, options Options) error {
	dsl, err := tryLoading(dslJSON)
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

	d := pkg.NewDifferential(sourceLines)
	// Apply each change
	for _, change := range dsl.Changes {
		err = d.ApplyChange(change)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error applying change: %s\n%s\n", err, change.String())

			return err // Or handle the error as you prefer
		}
	}

	// Join the lines back into a single string
	updatedContent := strings.Join(d.SourceLines, "\n")

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

var rootCmd = &cobra.Command{
	Use:   "myprogram [flags] FILE [FILE...]",
	Short: "A brief description of your program",
	Long:  `A longer description of your program...`,
	Run:   run,
	Args:  cobra.MinimumNArgs(1), // Require at least one argument
}

var (
	dryRun             bool
	showDiff           bool
	saveBackup         bool
	askForConfirmation bool
	dontOverWrite      bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "dry run")
	rootCmd.PersistentFlags().BoolVar(&showDiff, "show-diff", true, "show diff")
	rootCmd.PersistentFlags().BoolVar(&saveBackup, "save-backup", true, "save backup")
	rootCmd.PersistentFlags().BoolVar(&askForConfirmation, "ask-for-confirmation", true, "ask for confirmation")
	rootCmd.PersistentFlags().BoolVar(&dontOverWrite, "dont-overwrite", false, "don't overwrite")
}

func run(cmd *cobra.Command, args []string) {
	for _, dslFile := range args {
		dslJSON, err := os.ReadFile(dslFile)
		if err != nil {
			fmt.Printf("Error reading DSL file %s: %s\n", dslFile, err)
			continue // Skip to the next file on error
		}

		options := Options{
			DryRun:             dryRun,
			ShowDiff:           showDiff,
			SaveBackup:         saveBackup,
			AskForConfirmation: askForConfirmation,
			DontOverWrite:      dontOverWrite,
		}
		err = applyDSL(string(dslJSON), options)
		if err != nil {
			fmt.Printf("Error applying DSL from file %s: %s\n", dslFile, err)
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
