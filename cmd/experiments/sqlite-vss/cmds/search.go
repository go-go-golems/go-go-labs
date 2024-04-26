package cmds

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sqlite-vss/pkg"
	"github.com/pkg/errors"
)

type SearchCommand struct {
	*cmds.CommandDescription
	embedder *pkg.Embedder
}

type SearchSettings struct {
	Query string `glazed.parameter:"query"`
}

func NewSearchCommand(embedder *pkg.Embedder) (*SearchCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &SearchCommand{
		CommandDescription: cmds.NewCommandDescription(
			"search",
			cmds.WithShort("Search for documents"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"query",
					parameters.ParameterTypeString,
					parameters.WithHelp("Search query"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
		embedder: embedder,
	}, nil
}

func (c *SearchCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &SearchSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	if s.Query == "" {
		return fmt.Errorf("query is required")
	}

	results, err := c.embedder.Search(ctx, s.Query)
	if err != nil {
		return err
	}

	for _, result := range results {
		row := types.NewRow(
			types.MRP("id", result.ID),
			types.MRP("distance", result.Distance),
			types.MRP("title", result.Title),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
