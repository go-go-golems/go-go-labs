package cmds

import (
	"context"
	"encoding/json"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/apps/gtm/pkg"
	"github.com/pkg/errors"
	"os"
)

type TagsCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*TagsCommand)(nil)

func NewTagsCommand() (*TagsCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &TagsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"tags",
			cmds.WithShort("Outputs the tags in a GTM file as a table"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to the GTM export file"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

type TagsSettings struct {
	File string `glazed.parameter:"file"`
}

func (c *TagsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &TagsSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	file, err := os.Open(s.File)
	if err != nil {
		return errors.Wrap(err, "could not open file")
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	decoder := json.NewDecoder(file)
	gtmExport := pkg.GTMExport{}
	err = decoder.Decode(&gtmExport)
	if err != nil {
		return errors.Wrap(err, "could not decode GTM export")
	}

	for _, tag := range gtmExport.ContainerVersion.Tag {
		row := types.NewRow(
			types.MRP("accountId", tag.AccountID),
			types.MRP("containerId", tag.ContainerID),
			types.MRP("tagId", tag.TagID),
			types.MRP("name", tag.Name),
			types.MRP("type", tag.Type),
			types.MRP("fingerprint", tag.Fingerprint),
			types.MRP("tagFiringOption", tag.TagFiringOption),
		)

		switch tag.Type {
		case "html":
			row.Set("html", tag.Parameter[0].Value)
		case "http_request":
			for _, parameter := range tag.Parameter {
				switch parameter.Key {
				case "url":
					row.Set("url", parameter.Value)
				case "httpMethod":
					row.Set("httpMethod", parameter.Value)
				case "requestBody":
					row.Set("requestBody", parameter.Value)
				case "headers":
					headers := make(map[string]string)
					for _, header := range parameter.List {
						keyParameter := header.Map[0]
						valueParameter := header.Map[1]
						headers[keyParameter.Value] = valueParameter.Value
					}
					row.Set("headers", headers)
				}
			}

		}
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
