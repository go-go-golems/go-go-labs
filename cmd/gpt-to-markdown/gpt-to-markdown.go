package main

import (
	"context"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
	"strings"
)

type GPTToMarkdownCommand struct {
	*cmds.CommandDescription
}

func NewGPTToMarkdownCommand() (*GPTToMarkdownCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &GPTToMarkdownCommand{
		CommandDescription: cmds.NewCommandDescription(
			"gpt-to-markdown",
			cmds.WithShort("Converts GPT HTML to markdown"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"urls",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Path to HTML files or URLs"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithFlags(

				parameters.NewParameterDefinition(
					"concise",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Concise output"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"with-metadata",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Include metadata"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"rename-roles",
					parameters.ParameterTypeKeyValue,
					parameters.WithHelp("Rename roles"),
					parameters.WithDefault(map[string]string{
						"user":      "john",
						"assistant": "claire",
						"system":    "george",
					}),
				),
			),
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}, nil

}

func (cmd *GPTToMarkdownCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	w io.Writer,
) error {
	// Extracting arguments and flags
	urls := ps["urls"].([]string)
	concise := ps["concise"].(bool)
	withMetadata := ps["with-metadata"].(bool)
	renameRoles := ps["rename-roles"].(map[string]string)

	if len(urls) == 0 {
		return errors.New("No URLs provided")
	}

	for _, url := range urls {
		var htmlContent []byte
		var err error

		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			htmlContent_, err := getContent(url)
			if err != nil {
				return err
			}
			htmlContent = htmlContent_
		} else {
			htmlContent, err = os.ReadFile(url)
			if err != nil {
				return err
			}
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(htmlContent)))
		if err != nil {
			return err
		}

		scriptContent := doc.Find("#__NEXT_DATA__").Text()
		var data NextData
		err = json.Unmarshal([]byte(scriptContent), &data)
		if err != nil {
			return err
		}

		renderer := &Renderer{
			RenameRoles:  renameRoles,
			Concise:      concise,
			WithMetadata: withMetadata,
		}

		linearConversation := data.Props.PageProps.ServerResponse.LinearConversation

		renderer.PrintConversation(url, data.Props.PageProps.ServerResponse.ServerResponseData, linearConversation)
	}

	return nil
}

func getContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	htmlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return htmlContent, nil
}
