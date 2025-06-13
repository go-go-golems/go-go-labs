package loaders

import (
	"io"
	"io/fs"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	datadog_cmds "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/cmds"
	dd_types "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// DatadogYAMLCommandLoader loads YAML files and converts them to DatadogQueryCommands
type DatadogYAMLCommandLoader struct{}

var _ loaders.CommandLoader = (*DatadogYAMLCommandLoader)(nil)

// LoadCommands loads Datadog query commands from YAML files
func (d *DatadogYAMLCommandLoader) LoadCommands(
	f fs.FS,
	entryName string,
	options []cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) ([]cmds.Command, error) {
	s, err := f.Open(entryName)
	if err != nil {
		return nil, err
	}
	defer func(s fs.File) {
		_ = s.Close()
	}(s)

	return loaders.LoadCommandOrAliasFromReader(
		s,
		d.loadDatadogCommand,
		options,
		aliasOptions,
	)
}

// loadDatadogCommand parses a YAML file into a DatadogQueryCommand
func (d *DatadogYAMLCommandLoader) loadDatadogCommand(
	s io.Reader,
	options []cmds.CommandDescriptionOption,
	_ []alias.Option,
) ([]cmds.Command, error) {
	buf, err := io.ReadAll(s)
	if err != nil {
		return nil, err
	}

	// Parse the full YAML to extract all data
	var yamlData map[string]interface{}
	err = yaml.Unmarshal(buf, &yamlData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse YAML")
	}

	// Parse into CommandDescription
	description := &cmds.CommandDescription{}
	err = yaml.Unmarshal(buf, description)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse command description")
	}

	// Initialize Layers if nil (required for AppendLayers to work)
	if description.Layers == nil {
		description.Layers = layers.NewParameterLayers()
	}

	// Apply additional options
	for _, option := range options {
		option(description)
	}

	// Extract query
	query, ok := yamlData["query"].(string)
	if !ok {
		return nil, errors.New("query field is required in YAML file")
	}

	// Extract subqueries metadata
	var subqueries dd_types.QueryMetadata
	if subqueriesData, exists := yamlData["subqueries"]; exists {
		subqueriesBytes, err := yaml.Marshal(subqueriesData)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal subqueries")
		}
		err = yaml.Unmarshal(subqueriesBytes, &subqueries)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal subqueries")
		}
	}

	// Create the Datadog query command
	datadogCmd, err := datadog_cmds.NewDatadogQueryCommand(description, query, subqueries)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create datadog query command")
	}

	return []cmds.Command{datadogCmd}, nil
}

// IsFileSupported checks if the file is a supported YAML file
func (d *DatadogYAMLCommandLoader) IsFileSupported(f fs.FS, fileName string) bool {
	return strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml")
}

// NewDatadogYAMLCommandLoader creates a new DatadogYAMLCommandLoader
func NewDatadogYAMLCommandLoader() loaders.CommandLoader {
	return &DatadogYAMLCommandLoader{}
}
