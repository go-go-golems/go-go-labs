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
	"github.com/rs/zerolog/log"
)

type FindUsageClassesCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*FindUsageClassesCommand)(nil)

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
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

type FindUsageClassesSettings struct {
	Used                []map[string]interface{} `glazed.parameter:"used"`
	Defined             []map[string]interface{} `glazed.parameter:"defined"`
	FilterDefiningFiles []string                 `glazed.parameter:"filter-defining-files"`
	FilterUsingFiles    []string                 `glazed.parameter:"filter-using-files"`
	FilterClasses       []string                 `glazed.parameter:"filter-classes"`
}

func (c *FindUsageClassesCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &FindUsageClassesSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	definedClassesToFiles := make(map[string][]string)
	fileToDefinedClasses := make(map[string][]string)

	for _, entry := range s.Defined {
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

	for _, entry := range s.Used {
		class_ := entry["class"].(string)
		file := entry["file"].(string)
		usedClassesToFiles[class_] = append(usedClassesToFiles[class_], file)
	}

	for file, classes := range fileToDefinedClasses {
		for _, class := range classes {
			filesUsingClass := usedClassesToFiles[class]
			if len(s.FilterDefiningFiles) > 0 {
				if !containsGlob(s.FilterDefiningFiles, file) {
					continue
				}
			}
			if len(s.FilterUsingFiles) > 0 {
				if !containsAnyGlob(s.FilterUsingFiles, filesUsingClass) {
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
