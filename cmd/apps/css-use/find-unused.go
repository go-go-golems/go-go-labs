package main

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
)

type FindUnusedClassesCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*FindUnusedClassesCommand)(nil)

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
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

type FindUnusedClassesSettings struct {
	Used                []map[string]interface{} `glazed.parameter:"used"`
	Defined             []map[string]interface{} `glazed.parameter:"defined"`
	CheckAllUnused      bool                     `glazed.parameter:"check-all-unused"`
	FilterDefiningFiles []string                 `glazed.parameter:"filter-defining-files"`
	FilterUsingFiles    []string                 `glazed.parameter:"filter-using-files"`
	FilterClasses       []string                 `glazed.parameter:"filter-classes"`
}

func (c *FindUnusedClassesCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &FindUnusedClassesSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	usedMap := make(map[string]bool)
	for _, entry := range s.Used {
		class := entry["class"].(string)
		usedMap[class] = true
	}

	// Structures to keep track of counts
	fileToTotalClasses := make(map[string]int)
	fileToUsedClasses := make(map[string]map[string]interface{})
	definedClassesToFiles := make(map[string][]string)

	// Count the total number of classes per file
	for _, entry := range s.Defined {
		file := entry["file"].(string)
		fileToTotalClasses[file]++
		class := entry["class"].(string)
		definedClassesToFiles[class] = append(definedClassesToFiles[class], file)
	}

	// Count the number of used classes per file
	for _, entry := range s.Used {
		if len(s.FilterUsingFiles) > 0 {
			if !containsGlob(s.FilterUsingFiles, entry["file"].(string)) {
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

	for _, entry := range s.Defined {
		class := entry["class"].(string)
		if _, found := usedMap[class]; !found {
			filename := entry["file"].(string)
			if len(s.FilterDefiningFiles) > 0 {
				if !containsGlob(s.FilterDefiningFiles, filename) {
					continue
				}
			}
			if len(s.FilterClasses) > 0 {
				if !containsGlob(s.FilterClasses, class) {
					continue
				}
			}
			row := types.NewRow(
				types.MRP("class", class),
				types.MRP("file", filename),
			)

			if s.CheckAllUnused {
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
