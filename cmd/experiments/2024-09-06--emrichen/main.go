package main

import (
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/go-go-golems/go-emrichen/pkg/emrichen"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: simple_emrichen <input_file>")
		os.Exit(1)
	}

	inputFile := os.Args[1]

	// Initialize emrichen interpreter
	ei, err := emrichen.NewInterpreter(
		emrichen.WithFuncMap(sprig.TxtFuncMap()),
	)
	if err != nil {
		fmt.Printf("Error creating interpreter: %v\n", err)
		os.Exit(1)
	}

	// Process the input file
	err = processFile(ei, inputFile, os.Stdout)
	if err != nil {
		fmt.Printf("Error processing file: %v\n", err)
		os.Exit(1)
	}
}

func processFile(interpreter *emrichen.Interpreter, filePath string, w io.Writer) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)

	for {
		var document interface{}

		err = decoder.Decode(interpreter.CreateDecoder(&document))
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Skip empty documents
		if document == nil {
			continue
		}

		processedYAML, err := yaml.Marshal(&document)
		if err != nil {
			return err
		}

		fmt.Println("Processed YAML:")
		_, err = w.Write(processedYAML)
		if err != nil {
			return err
		}
	}

	return nil
}
