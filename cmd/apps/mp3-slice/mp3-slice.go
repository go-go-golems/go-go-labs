package main

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/apps/mp3-slice/mp3lib"
	"github.com/pkg/errors"
	"path/filepath"
)

// SliceCommand is the command structure for the slice command
type SliceCommand struct {
	*cmds.CommandDescription
}

// NewSliceCommand initializes a new SliceCommand
func NewSliceCommand() (*SliceCommand, error) {
	// Create glazed parameter layer
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	// Define flags
	mp3FilePath := parameters.NewParameterDefinition(
		"file",
		parameters.ParameterTypeString,
		parameters.WithHelp("Path to the mp3 file to slice"),
		parameters.WithRequired(true),
	)

	duration := parameters.NewParameterDefinition(
		"duration",
		parameters.ParameterTypeInteger,
		parameters.WithHelp("Duration of each slice in seconds"),
		parameters.WithRequired(true),
	)

	outputDir := parameters.NewParameterDefinition(
		"output-dir",
		parameters.ParameterTypeString,
		parameters.WithHelp("Output directory for sliced mp3 segments"),
		parameters.WithDefault("."),
	)

	// Assemble the command
	return &SliceCommand{
		CommandDescription: cmds.NewCommandDescription(
			"slice",
			cmds.WithShort("Slice an mp3 file into segments"),
			cmds.WithFlags(mp3FilePath, duration, outputDir),
			cmds.WithLayers(glazedParameterLayer),
		),
	}, nil
}

// Ensure SliceCommand satisfies the GlazeCommand interface
var _ cmds.GlazeCommand = &SliceCommand{}

func (c *SliceCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {

	// Extract flag values
	mp3FilePath, _, err := cast.GetAndCast[string](ps, "file", "")
	if err != nil {
		return errors.Wrap(err, "Error getting mp3 file path")
	}
	duration, _, err := cast.GetAndCast[int](ps, "duration", 250)
	if err != nil {
		return errors.Wrap(err, "Error getting duration")
	}
	outputDir, _, err := cast.GetAndCast[string](ps, "output-dir", ".")
	if err != nil {
		return errors.Wrap(err, "Error getting output directory")
	}

	// Ensure the output directory exists
	if err := ensureDirExists(outputDir); err != nil {
		return errors.Wrap(err, "Error ensuring output directory exists")
	}

	// Get the length of the MP3 file
	length, err := mp3lib.GetLengthSeconds(mp3FilePath)
	if err != nil {
		return errors.Wrap(err, "Error getting mp3 file length")
	}

	// Calculate the number of slices
	numSlices := length / duration
	if length%duration != 0 {
		numSlices++
	}

	// Start slicing the mp3 file
	for i := 0; i < numSlices; i++ {
		startSec := i * duration
		endSec := startSec + duration
		if endSec > length {
			endSec = length
		}

		outputFilePath := filepath.Join(outputDir, fmt.Sprintf("slice_%.2d.mp3", i+1))
		err := mp3lib.ExtractSectionToFile(mp3FilePath, outputFilePath, startSec, endSec)
		if err != nil {
			return errors.Wrapf(err, "Error extracting segment from %d to %d seconds", startSec, endSec)
		}

		// Create a row for the result and send it via the GlazeProcessor
		row := types.NewRow(
			types.MRP("segment_number", i+1),
			types.MRP("start_sec", startSec),
			types.MRP("end_sec", endSec),
			types.MRP("output_file", outputFilePath),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
