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
	*cmds.CommandDescription
	regexMap map[string]*regexp.Regexp
}

var _ cmds.GlazeCommand = (*RegexpMatchCommand)(nil)

type RegexpMatchSettings struct {
	InputFile string `json:"inputFile"`
}

func NewRegexpMatchCommand(regexMap map[string]*regexp.Regexp) (*RegexpMatchCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &RegexpMatchCommand{
		CommandDescription: cmds.NewCommandDescription(
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

func (c *RegexpMatchCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &RegexpMatchSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	file, err := os.Open(s.InputFile)
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
