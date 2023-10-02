package main

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type UsedCommand struct {
	*cmds.CommandDescription
}

func NewUsedCommand() (*UsedCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &UsedCommand{
		CommandDescription: cmds.NewCommandDescription(
			"used",
			cmds.WithShort("Parses an HTML page and lists all CSS classes used in it."),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"files",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of URLs to parse for CSS classes."),
					parameters.WithDefault([]string{}),
				),
			),
			cmds.WithLayers(glazedLayer),
		),
	}, nil
}

func (c *UsedCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	files := ps["files"].([]string)

	for _, url := range files {
		reader, err2 := ReaderUrlOrFile(url)
		if err2 != nil {
			return err2
		}
		defer func() {
			_ = reader.Close()
		}()

		doc, err := goquery.NewDocumentFromReader(reader)
		if err != nil {
			return err
		}

		classesMap := make(map[string]bool)
		doc.Find("*").Each(func(index int, element *goquery.Selection) {
			class, exists := element.Attr("class")
			if exists {
				for _, cls := range strings.Fields(class) {
					classesMap[cls] = true
				}
			}
		})

		alphabeticalKeys := make([]string, 0, len(classesMap))
		for key := range classesMap {
			alphabeticalKeys = append(alphabeticalKeys, key)
		}
		sort.Strings(alphabeticalKeys)
		for _, key := range alphabeticalKeys {
			row := types.NewRow(
				types.MRP("file", url),
				types.MRP("class", key),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	return nil
}
