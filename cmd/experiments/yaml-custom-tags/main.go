package main

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Resolver interface {
	Process(node *yaml.Node) (*yaml.Node, error)
}

func main() {
	// Check if a file path is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: myyamltool <yaml-file>")
		os.Exit(1)
	}

	// Open the file provided in the command line
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	// Create a YAML decoder
	decoder := yaml.NewDecoder(f)

	// Initialize the Emrichen interpreter
	interpreter, err := NewEmrichenInterpreter()
	if err != nil {
		panic(err)
	}

	for {
		// Declare a variable to hold the decoded data
		var document interface{}

		// Decode the YAML content into the document
		err = decoder.Decode(interpreter.CreateDecoder(&document))
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		// skip a document that was probably used to set !Defaults
		if document == nil {
			continue
		}
		// Marshal the processed document back to YAML

		processedYAML, err := yaml.Marshal(&document)
		if err != nil {
			panic(err)
		}

		// Print the processed YAML
		fmt.Println(string(processedYAML))
	}

}
