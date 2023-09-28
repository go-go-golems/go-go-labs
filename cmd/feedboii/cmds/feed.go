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
	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"
	"strings"
)

type FeedCommand struct {
	*cmds.CommandDescription
}

func NewFeedCommand() (*FeedCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &FeedCommand{
		CommandDescription: cmds.NewCommandDescription(
			"feed",
			cmds.WithShort("Fetch RSS feed and format as rows"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"feed-url",
					parameters.ParameterTypeStringList,
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (f *FeedCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	feedUrls, ok := ps["feed-url"].([]string)
	if !ok {
		return fmt.Errorf("feed-url is not a string list")
	}

	fp := gofeed.NewParser()

	for _, url := range feedUrls {
		feed, err := fp.ParseURL(url)
		if err != nil {
			return errors.Wrapf(err, "Error fetching the RSS feed from URL: %s", url)
		}

		// Convert feed into structured data
		row := types.NewRow()
		row.Set("feedTitle", strings.TrimSpace(feed.Title))
		row.Set("feedLink", strings.TrimSpace(feed.Link))
		row.Set("feedDescription", strings.TrimSpace(feed.Description))

		for _, item := range feed.Items {
			row.Set("title", strings.TrimSpace(item.Title))
			link := item.Link
			if link == "" {
				// look for the first enclosure
				for _, enclosure := range item.Enclosures {
					if enclosure.URL != "" {
						link = enclosure.URL
						break
					}
				}
			}
			row.Set("link", strings.TrimSpace(link))
			row.Set("description", strings.TrimSpace(item.Description))

			err = gp.AddRow(ctx, row)
			if err != nil {
				return errors.Wrapf(err, "Error processing RSS feed from URL: %s", url)
			}
		}
	}

	return nil
}
