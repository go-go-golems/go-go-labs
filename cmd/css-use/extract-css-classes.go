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
	"io"
	"os"
	"sort"
	"strings"
)

type DefinedCommand struct {
	*cmds.CommandDescription
}

func NewDefinedCommand() (*DefinedCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &DefinedCommand{
		CommandDescription: cmds.NewCommandDescription(
			"defined",
			cmds.WithShort("Parse provided HTML files and extract CSS classes."),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"files",
					parameters.ParameterTypeFileList,
					parameters.WithHelp("List of HTML files to parse for CSS classes."),
					parameters.WithDefault([]*parameters.FileData{}),
				),
				parameters.NewParameterDefinition(
					"urls",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of URLs to parse for CSS classes."),
					parameters.WithDefault([]string{}),
				),
				parameters.NewParameterDefinition(
					"with_selector",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Include CSS selectors in output."),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"with_rules",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Include CSS rules in output."),
					parameters.WithDefault(false),
				),
			),
			cmds.WithLayers(glazedParameterLayer),
		),
	}, nil
}

func (c *DefinedCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	withSelector := ps["with_selector"].(bool)
	withRules := ps["with_rules"].(bool)

	fileList, ok := cast.CastList2[*parameters.FileData, interface{}](ps["files"])
	if !ok {
		return fmt.Errorf("files argument is not a list of files")
	}

	for _, fileData := range fileList {
		fileContent, err := os.ReadFile(fileData.Path)
		if err != nil {
			return fmt.Errorf("Error reading file %s: %w", fileData.Path, err)
		}

		reader := strings.NewReader(string(fileContent))
		err2 := outputDefinedCSSClassesFromHTML(ctx, reader, fileData.Path, gp, withSelector, withRules)
		if err2 != nil {
			return err2
		}
	}

	urls := ps["urls"].([]string)

	for _, url := range urls {
		reader, err := ReaderUrlOrFile(url)
		if err != nil {
			return err
		}
		defer func() {
			_ = reader.Close()
		}()

		// if url ends with css, go straight to parsing CSS
		if strings.HasSuffix(url, ".css") {
			cssContent, err := io.ReadAll(reader)
			if err != nil {
				return fmt.Errorf("error reading CSS from %s: %w", url, err)
			}

			err = outputDefinedCSSClassesFromCSS(ctx, string(cssContent), url, gp, withSelector, withRules)
			if err != nil {
				return err
			}
			continue
		}

		err = outputDefinedCSSClassesFromHTML(ctx, reader, url, gp, withSelector, withRules)
		if err != nil {
			return err
		}
	}

	return nil
}

func outputDefinedCSSClassesFromHTML(
	ctx context.Context,
	reader io.Reader,
	path string,
	gp middlewares.Processor,
	withSelector bool,
	withRules bool,
) error {
	doc, err := html.Parse(reader)
	if err != nil {
		return fmt.Errorf("error parsing HTML from %s: %w", path, err)
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
		err = outputDefinedCSSClassesFromCSS(ctx, cssContent, path, gp, withSelector, withRules)
	}
	return nil
}

func outputDefinedCSSClassesFromCSS(
	ctx context.Context,
	cssContent string,
	path string,
	gp middlewares.Processor,
	withSelectors bool,
	withRules bool,
) error {
	selectors := parseCSS(cssContent)

	classes := map[string]interface{}{}

	for p := selectors.Oldest(); p != nil; p = p.Next() {
		selector, rules := p.Key, p.Value
		if withSelectors {
			row := types.NewRow(
				types.MRP("file", path),
				types.MRP("selector", selector),
			)
			if withRules {
				row.Set("rules", rules)
			}
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
			continue
		}

		// split selector and filter out .class_names
		selectorParts := strings.Split(selector, " ")
		for _, selectorPart := range selectorParts {
			if strings.HasPrefix(selectorPart, ".") {
				classes[selectorPart[1:]] = true
			}
		}
	}

	if !withSelectors {
		classes_ := []string{}
		for class := range classes {
			classes_ = append(classes_, class)
		}

		sort.Strings(classes_)

		for _, class := range classes_ {
			row := types.NewRow(
				types.MRP("file", path),
				types.MRP("class", class),
			)
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}

	return nil
}

type CSSRule struct {
	Selector string
	Rules    string
}

func GetDefinedCSSRules(cssContent string) ([]CSSRule, error) {
	selectors := parseCSS(cssContent)

	var cssRules []CSSRule
	for p := selectors.Oldest(); p != nil; p = p.Next() {
		selector, rules := p.Key, p.Value
		cssRules = append(cssRules, CSSRule{
			Selector: selector,
			Rules:    rules,
		})
	}

	return cssRules, nil
}
