package main

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"golang.org/x/net/html"
	"io"
	"sort"
	"strings"
)

type DefinedCommand struct {
	*cmds.CommandDescription
}

// This would be good to transform into something that keeps track of the output
// processor, the flags, probably as a helper class, to avoid having to pass arguments all
// over.

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
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"files",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("List of HTML files (or URLs) to parse for CSS classes."),
					parameters.WithDefault([]string{}),
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

	urls := ps["files"].([]string)

	for _, url := range urls {
		reader, err := ReaderUrlOrFile(url)
		if err != nil {
			return err
		}
		defer func() {
			_ = reader.Close()
		}()

		err = ParseAndOutputFile(
			ctx,
			url,
			reader,
			gp,
			withSelector,
			withRules,
		)
		if err != nil {
			return err
		}

	}

	return nil
}

func ParseAndOutputFile(
	ctx context.Context,
	url string,
	reader io.Reader,
	gp middlewares.Processor,
	withSelector bool,
	withRules bool,
) error {
	// if url ends with css, go straight to parsing CSS
	if strings.HasSuffix(url, ".css") {
		cssContent, err := io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("error reading CSS from %s: %w", url, err)
		}
		rules, err := GetRules(string(cssContent))
		if err != nil {
			return err
		}

		err = outputRules(ctx, rules, url, gp, withSelector, withRules)
		if err != nil {
			return err
		}

		return nil
	}

	rules, err := GetRulesFromHTML(reader)
	if err != nil {
		return err
	}

	err = outputRules(ctx, rules, url, gp, withSelector, withRules)
	if err != nil {
		return err
	}

	return nil
}

func GetRulesFromHTML(
	reader io.Reader,
) ([]CSSRule, error) {
	ret := []CSSRule{}

	doc, err := html.Parse(reader)
	if err != nil {
		return nil, err
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
		rules, err := GetRules(cssContent)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rules...)
	}

	return ret, nil
}

func outputRules(
	ctx context.Context,
	rules []CSSRule,
	path string,
	gp middlewares.Processor,
	withSelectors bool,
	withRules bool,
) error {
	classes := map[string]interface{}{}

	for _, rule := range rules {
		if withSelectors {
			row := types.NewRow(
				types.MRP("file", path),
				types.MRP("selector", rule.Selector),
			)
			if withRules {
				row.Set("rules", rule.Rules)
			}
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
			continue
		}

		// split selector and filter out .class_names
		selectorParts := strings.Split(rule.Selector, " ")
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

func GetRules(cssContent string) ([]CSSRule, error) {
	selectors := parseCSS(cssContent)

	var cssRules []CSSRule
	for p := selectors.Oldest(); p != nil; p = p.Next() {
		selector, rules := p.Key, p.Value
		ruleBody := ""
		for r := rules.Oldest(); r != nil; r = r.Next() {
			ruleBody += fmt.Sprintf("%s: %s; ", r.Key, r.Value)
		}
		cssRules = append(cssRules, CSSRule{
			Selector: selector,
			Rules:    ruleBody,
		})
	}

	return cssRules, nil
}
