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
				parameters.NewParameterDefinition(
					"check-all-unused",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Check if all classes in a file are unused"),
					parameters.WithDefault(false),
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

	// Structures to keep track of counts
	fileToTotalClasses := make(map[string]int)
	fileToUsedClasses := make(map[string]map[string]interface{})
	definedClassesToFiles := make(map[string][]string)

	// Count the total number of classes per file
	for _, entry := range defined {
		file := entry["file"].(string)
		fileToTotalClasses[file]++
		class := entry["class"].(string)
		definedClassesToFiles[class] = append(definedClassesToFiles[class], file)
	}

	filterUsingFiles := ps["filter-using-files"].([]string)
	// Count the number of used classes per file
	for _, entry := range used {
		if len(filterUsingFiles) > 0 {
			if !containsGlob(filterUsingFiles, entry["file"].(string)) {
				continue
			}
		}
		class_ := entry["class"].(string)
		for _, file := range definedClassesToFiles[class_] {
			if _, found := fileToUsedClasses[file]; !found {
				fileToUsedClasses[file] = make(map[string]interface{})
			}
			fileToUsedClasses[file][class_] = true
		}
	}

	checkAllUnused := ps["check-all-unused"].(bool)

	filterDefiningFiles := ps["filter-defining-files"].([]string)
	filterClasses := ps["filter-classes"].([]string)

	for _, entry := range defined {
		class := entry["class"].(string)
		if _, found := usedMap[class]; !found {
			filename := entry["file"].(string)
			if len(filterDefiningFiles) > 0 {
				if !containsGlob(filterDefiningFiles, filename) {
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
				types.MRP("file", filename),
			)

			if checkAllUnused {
				totalClasses := fileToTotalClasses[filename]
				usedClasses := len(fileToUsedClasses[filename])
				row.Set("total_classes", totalClasses)
				row.Set("used_classes", usedClasses)
				row.Set("all_unused", usedClasses == 0)
			}
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}
	return nil
}
