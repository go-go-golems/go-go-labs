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
	"golang.org/x/net/html"
	"os"
	"strings"
)

type ExtractCSSClassesCommand struct {
	*cmds.CommandDescription
	fileList []parameters.FileData
}

func NewExtractCSSClassesCommand() (*ExtractCSSClassesCommand, error) {
	fileParameter := parameters.NewParameterDefinition(
		"files",
		parameters.ParameterTypeFileList,
		parameters.WithHelp("List of HTML files to parse for CSS classes."),
	)

	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &ExtractCSSClassesCommand{
		CommandDescription: cmds.NewCommandDescription(
			"extract-classes",
			cmds.WithShort("Parse provided HTML files and extract CSS classes."),
			cmds.WithArguments(fileParameter),
			cmds.WithLayers(glazedParameterLayer),
		),
	}, nil
}

func (c *ExtractCSSClassesCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	fileList, ok := cast.CastList2[*parameters.FileData, interface{}](ps["files"])
	if !ok {
		return fmt.Errorf("files argument is not a list of files")
	}

	for _, fileData := range fileList {
		fileContent, err := os.ReadFile(fileData.Path)
		if err != nil {
			return fmt.Errorf("Error reading file %s: %w", fileData.Path, err)
		}

		doc, err := html.Parse(strings.NewReader(string(fileContent)))
		if err != nil {
			return fmt.Errorf("Error parsing HTML from %s: %w", fileData.Path, err)
		}

		var cssContents []string
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "style" {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.TextNode {
						cssContents = append(cssContents, c.Data)
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(doc)

		for _, cssContent := range cssContents {
			selectors := parseCSS(cssContent)

			for p := selectors.Oldest(); p != nil; p = p.Next() {
				selector, rules := p.Key, p.Value
				row := types.NewRow(
					types.MRP("file", fileData.Path),
					types.MRP("selector", selector),
					types.MRP("rules", rules),
				)
				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}

		}
	}

	return nil
}
