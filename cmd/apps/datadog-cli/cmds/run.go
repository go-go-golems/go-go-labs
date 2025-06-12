package cmds

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	datadog_layers "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/layers"
	dd_types "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// RunCommand allows running ad-hoc YAML query files
type RunCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*RunCommand)(nil)

// NewRunCommand creates a new command for running YAML query files
func NewRunCommand() (*RunCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	datadogLayer, err := datadog_layers.NewDatadogParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create datadog parameter layer")
	}

	return &RunCommand{
		CommandDescription: cmds.NewCommandDescription(
			"run",
			cmds.WithShort("Run a YAML query file"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"query-file",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to the YAML query file"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(datadogLayer, glazedLayer),
		),
	}, nil
}

// RunIntoGlazeProcessor loads and executes a YAML query file
func (r *RunCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Get the query file path
	params := parsedLayers.GetDataMap()
	queryFile, ok := params["query-file"]
	if !ok {
		return errors.New("query-file parameter not found")
	}

	queryFilePath, ok := queryFile.(string)
	if !ok {
		return errors.New("query-file must be a string")
	}

	// Check if file exists
	if _, err := os.Stat(queryFilePath); os.IsNotExist(err) {
		return errors.Errorf("query file does not exist: %s", queryFilePath)
	}

	// Read the YAML file
	data, err := ioutil.ReadFile(queryFilePath)
	if err != nil {
		return errors.Wrapf(err, "failed to read query file: %s", queryFilePath)
	}

	// Parse the YAML into a map to extract all data
	var yamlData map[string]interface{}
	err = yaml.Unmarshal(data, &yamlData)
	if err != nil {
		return errors.Wrapf(err, "failed to parse YAML file: %s", queryFilePath)
	}

	// Create command description from YAML
	commandDesc := &cmds.CommandDescription{}
	err = yaml.Unmarshal(data, commandDesc)
	if err != nil {
		return errors.Wrap(err, "failed to create command description from YAML")
	}

	// Extract query and subqueries
	query, ok := yamlData["query"].(string)
	if !ok {
		return errors.New("query field is required in YAML file")
	}

	var subqueries dd_types.QueryMetadata
	if subqueriesData, ok := yamlData["subqueries"]; ok {
		subqueriesBytes, err := yaml.Marshal(subqueriesData)
		if err != nil {
			return errors.Wrap(err, "failed to marshal subqueries")
		}
		err = yaml.Unmarshal(subqueriesBytes, &subqueries)
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal subqueries")
		}
	}

	// Create and execute the Datadog query command
	datadogCmd, err := NewDatadogQueryCommand(commandDesc, query, subqueries)
	if err != nil {
		return errors.Wrap(err, "failed to create Datadog query command")
	}

	// Execute the command
	return datadogCmd.RunIntoGlazeProcessor(ctx, parsedLayers, gp)
}
