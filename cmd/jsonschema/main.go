package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xeipuuv/gojsonschema"
)

var onlyInvalid bool
var printDetails bool
var verbose bool

func validateJSON(schemaFile, jsonFile string) error {
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaFile)
	documentLoader := gojsonschema.NewReferenceLoader("file://" + jsonFile)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if result.Valid() {
		if !onlyInvalid {
			fmt.Printf("%s is valid\n", jsonFile)
			fmt.Println("The JSON is valid against the schema.")
		}
	} else {
		fmt.Printf("%s is invalid\n", jsonFile)
		fmt.Println("The JSON is NOT valid. See errors:")
		for _, desc := range result.Errors() {
			v := desc.Value()
			// serialize value to JSON
			b, err := json.Marshal(v)
			var s string
			if err != nil {
				s = fmt.Sprintf("%v", v)
			} else {
				s = string(b)
			}

			if printDetails {
				fmt.Printf("- %s, value: %v\n", desc, s)
			} else {
				fmt.Printf("- %s\n", desc)
			}
		}
		fmt.Println()
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "validator <schema.json> <document.json...>",
	Short: "Validates JSON files against a schema",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		schemaFile := args[0]
		for _, jsonFile := range args[1:] {
			if verbose {
				fmt.Printf("Validating %s against %s\n", jsonFile, schemaFile)
			}
			if err := validateJSON(schemaFile, jsonFile); err != nil {
				fmt.Printf("Error during validation: %s\n", err)
			}
		}
	},
}

func main() {
	rootCmd.Flags().BoolVar(&onlyInvalid, "only-invalid", false, "Only print invalid JSON results")
	rootCmd.Flags().BoolVar(&printDetails, "print-details", false, "Print details of validation errors")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
