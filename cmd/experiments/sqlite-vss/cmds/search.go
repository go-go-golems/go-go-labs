package cmds

import (
	"context"
	"fmt"
	geppetto_cmds "github.com/go-go-golems/geppetto/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmds_middlewares "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sqlite-vss/pkg"
	"github.com/pkg/errors"
	"os"
)

type SearchCommand struct {
	*cmds.CommandDescription
	embedder              *pkg.Embedder
	AnswerQuestionCommand *geppetto_cmds.GeppettoCommand
}

type SearchSettings struct {
	Query  string `glazed.parameter:"query"`
	Answer bool   `glazed.parameter:"answer"`
}

func NewSearchCommand(embedder *pkg.Embedder) (*SearchCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	answerCmd, err := NewAnswerQuestionCommand()
	if err != nil {
		return nil, errors.Wrap(err, "could not create AnswerQuestion command")
	}

	layerList := []layers.ParameterLayer{
		glazedParameterLayer,
	}
	err = answerCmd.Layers.ForEachE(func(slug string, layer layers.ParameterLayer) error {
		if slug != layers.DefaultSlug {
			layerList = append(layerList, layer)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not get additional layers")
	}

	// ideally we would bubble up and merge the answer question parameter layers here, so maybe
	// this can be done by instantiating the answer command here already, and we take everything except existing layers
	// (meaning, we still have the default layer to fill, but at that point it's there)

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
				parameters.NewParameterDefinition(
					"answer",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Answer the question"),
				),
			),
			cmds.WithLayersList(layerList...),
		),
		embedder:              embedder,
		AnswerQuestionCommand: answerCmd,
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

	if s.Answer {
		if len(results) == 0 {
			fmt.Println("No results found")
			return nil
		}

		// NOTE(manuel, 2024-04-26)
		// this could potentially be more elegant or solvable with codegen
		parsedLayers_ := parsedLayers.Clone()
		parsedLayers_.Delete(layers.DefaultSlug)

		mw := cmds_middlewares.UpdateFromMap(map[string]map[string]interface{}{
			layers.DefaultSlug: {
				"question": s.Query,
				"document": results[0].Title + "\n" + results[0].Body,
			},
		})
		_ = mw
		err = cmds_middlewares.ExecuteMiddlewares(c.AnswerQuestionCommand.Layers, parsedLayers_, mw)
		if err != nil {
			return err
		}

		err = c.AnswerQuestionCommand.RunIntoWriter(ctx, parsedLayers_, os.Stdout)
		if err != nil {
			return err
		}

		return nil
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
