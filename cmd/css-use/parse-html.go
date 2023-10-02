package main

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
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

	argURL := parameters.NewParameterDefinition(
		"url",
		parameters.ParameterTypeString,
		parameters.WithHelp("The URL of the HTML page to be parsed."),
	)

	return &UsedCommand{
		CommandDescription: cmds.NewCommandDescription(
			"used",
			cmds.WithShort("Parses an HTML page and lists all CSS classes used in it."),
			cmds.WithArguments(argURL),
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
	url := ps["url"].(string)

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

	fmt.Println("All CSS classes used in the HTML:")
	alphabeticalKeys := make([]string, 0, len(classesMap))
	for key := range classesMap {
		alphabeticalKeys = append(alphabeticalKeys, key)
	}
	sort.Strings(alphabeticalKeys)
	for _, key := range alphabeticalKeys {
		fmt.Println(key)
	}

	return nil
}
