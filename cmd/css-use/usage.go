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
	"github.com/rs/zerolog/log"
)

type FindUsageClassesCommand struct {
	*cmds.CommandDescription
}

func NewFindUsageClassesCommand() (*FindUsageClassesCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &FindUsageClassesCommand{
		CommandDescription: cmds.NewCommandDescription(
			"usage",
			cmds.WithShort("Find classes usage command"),
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
				parameters.NewParameterDefinition(
					"filter-defining-files",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of files to filter on (can be glob)"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"filter-using-files",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of files to filter on (can be glob)"),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"filter-classes",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of classes to filter on (can be glob)"),
					parameters.WithDefault([]string{}),
				),
			),
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (c *FindUsageClassesCommand) Run(
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

	definedClassesToFiles := make(map[string][]string)
	fileToDefinedClasses := make(map[string][]string)

	for _, entry := range defined {
		file := entry["file"].(string)
		class := entry["class"].(string)
		definedClassesToFiles[class] = append(definedClassesToFiles[class], file)

		fileToDefinedClasses[file] = append(fileToDefinedClasses[file], class)

		if len(definedClassesToFiles[class]) > 1 {
			log.Warn().Str("class", class).
				Strs("files", definedClassesToFiles[class]).
				Msg("class defined in multiple files")
		}
	}

	usedClassesToFiles := make(map[string][]string)

	for _, entry := range used {
		class_ := entry["class"].(string)
		file := entry["file"].(string)
		usedClassesToFiles[class_] = append(usedClassesToFiles[class_], file)
	}

	filterDefiningFiles := ps["filter-defining-files"].([]string)
	filterUsingFiles := ps["filter-using-files"].([]string)
	filterClasses := ps["filter-classes"].([]string)

	for file, classes := range fileToDefinedClasses {
		for _, class := range classes {
			filesUsingClass := usedClassesToFiles[class]
			if len(filterDefiningFiles) > 0 {
				if !containsGlob(filterDefiningFiles, file) {
					continue
				}
			}
			if len(filterUsingFiles) > 0 {
				if !containsAnyGlob(filterUsingFiles, filesUsingClass) {
					continue
				}
			}
			if len(filterClasses) > 0 {
				if !containsGlob(filterClasses, class) {
					continue
				}
			}
			row := types.NewRow(
				types.MRP("class", class),
				types.MRP("file", file),
				types.MRP("used", filesUsingClass),
				types.MRP("unused", len(filesUsingClass) == 0),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}
	return nil
}
