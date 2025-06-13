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
	"github.com/rs/zerolog/log"
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
	log.Info().
		Str("command", "run_yaml").
		Msg("Starting YAML query file execution")

	// Get the query file path
	log.Debug().Msg("Extracting query file path from parameters")
	params := parsedLayers.GetDataMap()
	queryFile, ok := params["query-file"]
	if !ok {
		log.Error().Msg("query-file parameter not found in parsed layers")
		return errors.New("query-file parameter not found")
	}

	queryFilePath, ok := queryFile.(string)
	if !ok {
		log.Error().
			Interface("query_file_value", queryFile).
			Msg("query-file parameter is not a string")
		return errors.New("query-file must be a string")
	}

	log.Debug().Str("query_file_path", queryFilePath).Msg("Query file path extracted")

	// Check if file exists
	log.Debug().Str("file_path", queryFilePath).Msg("Checking if query file exists")
	if _, err := os.Stat(queryFilePath); os.IsNotExist(err) {
		log.Error().
			Str("file_path", queryFilePath).
			Msg("Query file does not exist")
		return errors.Errorf("query file does not exist: %s", queryFilePath)
	}
	log.Debug().Str("file_path", queryFilePath).Msg("Query file exists")

	// Read the YAML file
	log.Debug().Str("file_path", queryFilePath).Msg("Reading YAML query file")
	data, err := ioutil.ReadFile(queryFilePath)
	if err != nil {
		log.Error().
			Err(err).
			Str("file_path", queryFilePath).
			Msg("Failed to read YAML query file")
		return errors.Wrapf(err, "failed to read query file: %s", queryFilePath)
	}
	log.Debug().
		Str("file_path", queryFilePath).
		Int("file_size", len(data)).
		Msg("YAML query file read successfully")

	// Parse the YAML into a map to extract all data
	log.Debug().Msg("Parsing YAML data into map")
	var yamlData map[string]interface{}
	err = yaml.Unmarshal(data, &yamlData)
	if err != nil {
		log.Error().
			Err(err).
			Str("file_path", queryFilePath).
			Msg("Failed to parse YAML file into map")
		return errors.Wrapf(err, "failed to parse YAML file: %s", queryFilePath)
	}

	// Log available keys in YAML for debugging
	yamlKeys := make([]string, 0, len(yamlData))
	for key := range yamlData {
		yamlKeys = append(yamlKeys, key)
	}
	log.Debug().
		Strs("yaml_keys", yamlKeys).
		Msg("YAML data parsed successfully")

	// Create command description from YAML
	log.Debug().Msg("Creating command description from YAML")
	commandDesc := &cmds.CommandDescription{}
	err = yaml.Unmarshal(data, commandDesc)
	if err != nil {
		log.Error().
			Err(err).
			Str("file_path", queryFilePath).
			Msg("Failed to create command description from YAML")
		return errors.Wrap(err, "failed to create command description from YAML")
	}
	log.Debug().
		Str("command_name", commandDesc.Name).
		Msg("Command description created from YAML")

	// Extract query and subqueries
	log.Debug().Msg("Extracting query field from YAML")
	query, ok := yamlData["query"].(string)
	if !ok {
		log.Error().
			Str("file_path", queryFilePath).
			Msg("Query field is missing or not a string in YAML file")
		return errors.New("query field is required in YAML file")
	}
	log.Debug().
		Str("query", query).
		Msg("Query field extracted from YAML")

	log.Debug().Msg("Extracting subqueries metadata from YAML")
	var subqueries dd_types.QueryMetadata
	if subqueriesData, ok := yamlData["subqueries"]; ok {
		log.Debug().Msg("Subqueries field found, processing")
		subqueriesBytes, err := yaml.Marshal(subqueriesData)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to marshal subqueries data")
			return errors.Wrap(err, "failed to marshal subqueries")
		}
		err = yaml.Unmarshal(subqueriesBytes, &subqueries)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to unmarshal subqueries data")
			return errors.Wrap(err, "failed to unmarshal subqueries")
		}
		log.Debug().
			Str("sort", subqueries.Sort).
			Strs("group_by", subqueries.GroupBy).
			Int("aggs_count", len(subqueries.Aggs)).
			Msg("Subqueries metadata extracted successfully")
	} else {
		log.Debug().Msg("No subqueries field found in YAML")
	}

	// Create and execute the Datadog query command
	log.Debug().Msg("Creating Datadog query command from YAML data")
	datadogCmd, err := NewDatadogQueryCommand(commandDesc, query, subqueries)
	if err != nil {
		log.Error().
			Err(err).
			Str("file_path", queryFilePath).
			Msg("Failed to create Datadog query command")
		return errors.Wrap(err, "failed to create Datadog query command")
	}
	log.Debug().Msg("Datadog query command created successfully")

	// Execute the command
	log.Info().
		Str("file_path", queryFilePath).
		Str("query", query).
		Msg("Executing Datadog query command from YAML file")

	err = datadogCmd.RunIntoGlazeProcessor(ctx, parsedLayers, gp)
	if err != nil {
		log.Error().
			Err(err).
			Str("file_path", queryFilePath).
			Msg("Failed to execute Datadog query command")
		return err
	}

	log.Info().
		Str("file_path", queryFilePath).
		Msg("YAML query file execution completed successfully")
	return nil
}
