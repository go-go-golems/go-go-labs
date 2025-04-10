package cmds

import (
	"context"
	"io"
	"reflect"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	// Import faker library
)

// --- Command Definition ---

type ProcessSettings struct {
	InputFile *parameters.FileData `glazed.parameter:"input-file"`
}

type ProcessCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*ProcessCommand)(nil)

func NewProcessCommand() (*ProcessCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &ProcessCommand{
		CommandDescription: cmds.NewCommandDescription(
			"process",
			cmds.WithShort("Process an Emrichen YAML file with faker tags into structured data"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"input-file",
					parameters.ParameterTypeFile,
					parameters.WithHelp("Input YAML file to process"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

// processOutput processes the unmarshalled data and adds it to the Glaze processor.
func processOutput(ctx context.Context, gp middlewares.Processor, data interface{}) error {
	// Handle different types of output data
	switch v := data.(type) {
	case map[string]interface{}:
		// Single object -> single row
		row := types.NewRowFromMap(v)
		if err := gp.AddRow(ctx, row); err != nil {
			return errors.Wrap(err, "failed to add single row")
		}
	case []interface{}:
		// List of items
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				// If item is an object -> row
				row := types.NewRowFromMap(itemMap)
				if err := gp.AddRow(ctx, row); err != nil {
					return errors.Wrap(err, "failed to add list row")
				}
			} else {
				// If item is scalar, wrap it
				row := types.NewRow(types.MRP("value", item))
				if err := gp.AddRow(ctx, row); err != nil {
					return errors.Wrap(err, "failed to add scalar list item row")
				}
			}
		}
	default:
		// Handle scalar or other types: wrap in a single row
		row := types.NewRow(types.MRP("value", v))
		if err := gp.AddRow(ctx, row); err != nil {
			return errors.Wrap(err, "failed to add default type row")
		}
	}
	return nil
}

func (c *ProcessCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ProcessSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings struct")
	}

	if s.InputFile == nil || s.InputFile.Path == "" {
		// Re-check since FileData might be initialized but empty
		return errors.New("input file parameter is required and must be specified")
	}

	// 1. Get the shared Faker Interpreter
	ei, err := NewFakerInterpreter()
	if err != nil {
		return err // Error already wrapped in factory
	}

	// 2. Decode and Process YAML from FileData Content
	decoder := yaml.NewDecoder(strings.NewReader(s.InputFile.Content))

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

		// 4. Process outputData and add to GlazeProcessor (gp)
		if err := processOutput(ctx, gp, outputData); err != nil {
			return err
		}
	}

	return nil
}
