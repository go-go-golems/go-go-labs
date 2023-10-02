package main

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
)

type FindUnusedClassesCommand struct {
	*cmds.CommandDescription
}

func NewFindUnusedClassesCommand() (*FindUnusedClassesCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &FindUnusedClassesCommand{
		CommandDescription: cmds.NewCommandDescription(
			"find-unused",
			cmds.WithShort("Find unused classes command"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"used",
					parameters.ParameterTypeObjectListFromFile,
					parameters.WithHelp("Path to the used.json file"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"defined",
					parameters.ParameterTypeObjectListFromFile,
					parameters.WithHelp("Path to the defined.json file"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (c *FindUnusedClassesCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	used_ := ps["used"].([]interface{})
	used, ok := cast.CastList2[map[string]interface{}, interface{}](used_)
	if !ok {
		return errors.New("could not cast used")
	}
	defined_ := ps["defined"].([]interface{})
	defined, ok := cast.CastList2[map[string]interface{}, interface{}](defined_)
	if !ok {
		return errors.New("could not cast defined")
	}

	usedMap := make(map[string]bool)
	for _, entry := range used {
		class := entry["class"].(string)
		usedMap[class] = true
	}

	for _, entry := range defined {
		class := entry["class"].(string)
		if _, found := usedMap[class]; !found {
			row := types.NewRow(
				types.MRP("class", class),
				types.MRP("file", entry["file"].(string)),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}
	return nil
}
