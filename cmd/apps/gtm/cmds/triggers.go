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
	"strings"
)

type TriggersCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*TriggersCommand)(nil)

func NewTriggersCommand() (*TriggersCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &TriggersCommand{
		CommandDescription: cmds.NewCommandDescription(
			"triggers",
			cmds.WithShort("Output triggers from GTM file"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to GTM file"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}, nil
}

type TriggersSettings struct {
	File string `glazed.parameter:"file"`
}

func (c *TriggersCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &TriggersSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(s.File)
	if err != nil {
		return errors.Wrap(err, "failed to read GTM file")
	}

	var gtmExport pkg.GTMExport
	if err := json.Unmarshal(data, &gtmExport); err != nil {
		return errors.Wrap(err, "failed to parse GTM file")
	}

	for _, trigger := range gtmExport.ContainerVersion.Trigger {
		filters := make([]string, len(trigger.Filter))
		for i, filter := range trigger.Filter {
			filters[i] = pkg.FilterToString(filter)
		}

		customEventFilters := make([]string, len(trigger.CustomEventFilter))
		for i, filter := range trigger.CustomEventFilter {
			customEventFilters[i] = pkg.FilterToString(filter)
		}

		row := types.NewRow(
			types.MRP("accountId", trigger.AccountID),
			types.MRP("containerId", trigger.ContainerID),
			types.MRP("triggerId", trigger.TriggerID),
			types.MRP("name", trigger.Name),
			types.MRP("type", trigger.Type),
			types.MRP("fingerprint", trigger.Fingerprint),
		)

		if len(filters) > 0 {
			row.Set("filters", strings.Join(filters, ", "))
		}

		if len(customEventFilters) > 0 {
			row.Set("customEventFilters", strings.Join(customEventFilters, ", "))
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
