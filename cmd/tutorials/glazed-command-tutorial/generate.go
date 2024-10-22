package main

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type GenerateCommand struct {
	*cmds.CommandDescription
}

func NewGenerateCommand() (*GenerateCommand, error) {
	return &GenerateCommand{
		CommandDescription: cmds.NewCommandDescription(
			"generate",
			cmds.WithShort("Generate user records"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"count",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Number of users to generate"),
					parameters.WithDefault(10),
				),
				parameters.NewParameterDefinition(
					"verbose",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Enable verbose output"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"prefix",
					parameters.ParameterTypeString,
					parameters.WithHelp("Prefix for usernames"),
					parameters.WithDefault("User"),
				),
			),
			cmds.WithArguments(),
		),
	}, nil
}

type GenerateSettings struct {
	Count   int    `glazed.parameter:"count"`
	Verbose bool   `glazed.parameter:"verbose"`
	Prefix  string `glazed.parameter:"prefix"`
}

func (c *GenerateCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	settings := &GenerateSettings{}
	if err := parsedLayers.InitializeStruct("default", settings); err != nil {
		return err
	}

	for i := 1; i <= settings.Count; i++ {
		user := types.NewRow(
			types.MRP("id", i),
			types.MRP("name", settings.Prefix+"-"+strconv.Itoa(i)),
			types.MRP("email", "user"+strconv.Itoa(i)+"@example.com"),
		)

		if settings.Verbose {
			user.Set("debug", "Verbose mode enabled")
		}

		if err := gp.AddRow(ctx, user); err != nil {
			return err
		}
	}

	return nil
}

func (c *GenerateCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
	settings := &GenerateSettings{}
	if err := parsedLayers.InitializeStruct("default", settings); err != nil {
		return err
	}

	for i := 1; i <= settings.Count; i++ {
		output := fmt.Sprintf("%s %d: %s-%d <user%d@example.com>\n", settings.Prefix, i, settings.Prefix, i, i)
		if settings.Verbose {
			output += "Debug: Verbose mode enabled\n"
		}
		_, err := w.Write([]byte(output))
		if err != nil {
			return err
		}
	}

	return nil
}
