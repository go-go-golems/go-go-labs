package cmds

import (
	"bufio"
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/pkg/errors"
	"os"
	"regexp"
	"strconv"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type RegexpMatchCommand struct {
	description *cmds.CommandDescription
	regexMap    map[string]*regexp.Regexp
}

func NewRegexpMatchCommand(regexMap map[string]*regexp.Regexp) (*RegexpMatchCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &RegexpMatchCommand{
		description: cmds.NewCommandDescription(
			"regexpMatch",
			cmds.WithShort("Regexp Match command"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"inputFile",
					parameters.ParameterTypeString,
					parameters.WithHelp("Input file to apply regexes"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
		regexMap: regexMap,
	}, nil
}

func (c *RegexpMatchCommand) Description() *cmds.CommandDescription {
	return c.description
}

func (c *RegexpMatchCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	inputFile := ps["inputFile"].(string)

	file, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		for name, regex := range c.regexMap {
			matches := regex.FindStringSubmatch(line)
			if matches != nil {
				row := types.NewRow(
					types.MRP("regex", name),
				)

				for i, match := range matches[1:] {
					row.Set("group"+strconv.Itoa(i+1), match)
				}

				if err := gp.AddRow(ctx, row); err != nil {
					return err
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
