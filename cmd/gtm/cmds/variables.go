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
	handlers "github.com/go-go-golems/go-go-labs/cmd/gtm/pkg"
	"github.com/pkg/errors"
	"os"
)

type VariablesCommand struct {
	description *cmds.CommandDescription
}

func NewVariablesCommand() (*VariablesCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &VariablesCommand{
		description: cmds.NewCommandDescription(
			"variables",
			cmds.WithShort("Outputs the variables in a GTM file as a table"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to the GTM export file"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (c *VariablesCommand) Description() *cmds.CommandDescription {
	return c.description
}

func (c *VariablesCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	filePath := ps["file"].(string)

	file, err := os.Open(filePath)
	if err != nil {
		return errors.Wrap(err, "could not open file")
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	decoder := json.NewDecoder(file)
	gtmExport := handlers.GTMExport{}
	err = decoder.Decode(&gtmExport)
	if err != nil {
		return errors.Wrap(err, "could not decode GTM export")
	}

	for _, variable := range gtmExport.ContainerVersion.Variable {
		row := types.NewRow(
			types.MRP("accountId", variable.AccountID),
			types.MRP("containerId", variable.ContainerID),
			types.MRP("variableId", variable.VariableID),
			types.MRP("name", variable.Name),
			types.MRP("type", variable.Type),
			types.MRP("fingerprint", variable.Fingerprint),
		)

		switch variable.Type {
		case "c":
			row.Set("type", "Constant")
			row.Set("value", variable.Parameter[0].Value)
		case "v":
			row.Set("type", "Data Layer Variable")
			row.Set("dataLayerVariable", variable.Parameter[2].Value)
		case "jsm":
			row.Set("type", "JavaScript Variable")
			row.Set("javascript", variable.Parameter[0].Value)
		case "ed":
			row.Set("type", "Element Data Variable")
			// find the template entry in parameters
			for _, parameter := range variable.Parameter {
				if parameter.Key == "keyPath" {
					row.Set("elementPath", parameter.Value)
				}
			}
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
