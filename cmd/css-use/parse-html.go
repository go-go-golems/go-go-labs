package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ParseHTMLCommand struct {
	*cmds.CommandDescription
}

func NewParseHTMLCommand() (*ParseHTMLCommand, error) {
	argURL := parameters.NewParameterDefinition(
		"url",
		parameters.ParameterTypeString,
		parameters.WithHelp("The URL of the HTML page to be parsed."),
	)

	return &ParseHTMLCommand{
		CommandDescription: cmds.NewCommandDescription(
			"parse-html",
			cmds.WithShort("Parses an HTML page and lists all CSS classes used in it."),
			cmds.WithArguments(argURL),
		),
	}, nil
}

func (c *ParseHTMLCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	url := ps["url"].(string)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
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
