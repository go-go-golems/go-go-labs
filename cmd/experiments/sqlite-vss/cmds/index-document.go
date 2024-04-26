package cmds

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sqlite-vss/pkg"
	"github.com/pkg/errors"
	"time"
)

type IndexDocumentCommand struct {
	*cmds.CommandDescription
	embedder *pkg.Embedder
}

type IndexDocumentSettings struct {
	File  parameters.FileData `glazed.parameter:"file"`
	Title string              `glazed.parameter:"title"`
	Body  string              `glazed.parameter:"body"`
}

func NewIndexDocumentCommand(embedder *pkg.Embedder) (*IndexDocumentCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &IndexDocumentCommand{
		CommandDescription: cmds.NewCommandDescription(
			"index",
			cmds.WithShort("Index a document"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeFile,
					parameters.WithHelp("File to index"),
				),
				parameters.NewParameterDefinition(
					"title",
					parameters.ParameterTypeString,
					parameters.WithHelp("Title of the document"),
				),
				parameters.NewParameterDefinition(
					"body",
					parameters.ParameterTypeString,
					parameters.WithHelp("Body of the document"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
		embedder: embedder,
	}, nil
}

var _ cmds.GlazeCommand = (*IndexDocumentCommand)(nil)

func (c *IndexDocumentCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &IndexDocumentSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	var title, body string
	var modifiedAt time.Time

	if s.File.Path != "" {
		// Index from file
		title = s.File.BaseName
		body = s.File.StringContent
		modifiedAt = s.File.LastModifiedTime
	} else {
		// Index from title and body
		title = s.Title
		body = s.Body
		modifiedAt = time.Now()
	}

	if title == "" || body == "" {
		return fmt.Errorf("title and body are required")
	}

	err := c.embedder.IndexDocument(ctx, title, body, modifiedAt)
	if err != nil {
		return err
	}

	fmt.Printf("Indexed document: %s\n", title)

	return nil
}
