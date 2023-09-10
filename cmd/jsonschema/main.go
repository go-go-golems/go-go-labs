package main

import (
	"fmt"
	"os"

	"github.com/xeipuuv/gojsonschema"
)

func validateJSON(schemaFile, jsonFile string) error {
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaFile)
	documentLoader := gojsonschema.NewReferenceLoader("file://" + jsonFile)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if result.Valid() {
		fmt.Println("The JSON is valid against the schema.")
	} else {
		fmt.Println("The JSON is NOT valid. See errors:")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
	}

	return nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: validator <schema.json> <document.json>")
		os.Exit(1)
	}

	schemaFile := os.Args[1]
	for _, jsonFile := range os.Args[2:] {

		fmt.Printf("Validating %s against %s\n", jsonFile, schemaFile)
		if err := validateJSON(schemaFile, jsonFile); err != nil {
			fmt.Printf("Error during validation: %s\n", err)
		}
	}
}
