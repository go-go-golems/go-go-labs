package cmds

import (
	"context"
	"io"
	"reflect"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// GenerateSettings defines parameters for the generate command
type GenerateSettings struct {
	InputFile *parameters.FileData `glazed.parameter:"input-file"`
}

// GenerateCommand processes an input YAML using Emrichen and outputs the resulting YAML.
type GenerateCommand struct {
	*cmds.CommandDescription
}

var _ cmds.WriterCommand = (*GenerateCommand)(nil)

// NewGenerateCommand creates a new GenerateCommand instance.
func NewGenerateCommand() (*GenerateCommand, error) {
	return &GenerateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"generate",
			cmds.WithShort("Generate YAML output from an Emrichen YAML file with faker tags"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"input-file",
					parameters.ParameterTypeFile,
					parameters.WithHelp("Input YAML file to process"),
					parameters.WithRequired(true),
				),
			),
			// No Glazed layer needed for WriterCommand that outputs raw YAML
		),
	}, nil
}

// RunIntoWriter executes the generate command.
func (c *GenerateCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	s := &GenerateSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings struct")
	}

	if s.InputFile == nil || s.InputFile.Path == "" {
		return errors.New("input file parameter is required and must be specified")
	}

	// 1. Get the shared Faker Interpreter
	ei, err := NewFakerInterpreter()
	if err != nil {
		return err // Error already wrapped in factory
	}

	// 2. Decode, Process YAML, and Encode output
	decoder := yaml.NewDecoder(strings.NewReader(s.InputFile.Content))
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2) // Standard YAML indentation

	firstDoc := true
	for {
		var outputData interface{}
		// Use ei.CreateDecoder to process tags during decoding
		err = decoder.Decode(ei.CreateDecoder(&outputData))
		if err == io.EOF {
			break // End of file
		}
		if err != nil {
			return errors.Wrapf(err, "failed to decode or process YAML from %s", s.InputFile.Path)
		}

		// Handle documents that evaluate to nil (like !Defaults only)
		if outputData == nil || (reflect.ValueOf(outputData).Kind() == reflect.Map && reflect.ValueOf(outputData).IsNil()) {
			continue
		}

		// Add YAML document separator --- if not the first document
		if !firstDoc {
			_, err = w.Write([]byte("---\n"))
			if err != nil {
				return errors.Wrap(err, "failed to write document separator")
			}
		}
		firstDoc = false

		// Encode the processed data back to YAML
		if err := encoder.Encode(outputData); err != nil {
			return errors.Wrap(err, "failed to encode processed data to YAML")
		}
	}

	if err := encoder.Close(); err != nil {
		return errors.Wrap(err, "failed to close YAML encoder")
	}

	return nil
}
