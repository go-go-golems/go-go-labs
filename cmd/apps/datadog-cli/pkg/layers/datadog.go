package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
)

// NewDatadogParameterLayer creates a parameter layer for Datadog authentication
func NewDatadogParameterLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"datadog",
		"Datadog Configuration",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"api-key",
				parameters.ParameterTypeString,
				parameters.WithHelp("Datadog API key (env: DATADOG_CLI_API_KEY)"),
			),
			parameters.NewParameterDefinition(
				"app-key",
				parameters.ParameterTypeString,
				parameters.WithHelp("Datadog application key (env: DATADOG_CLI_APP_KEY)"),
			),
			parameters.NewParameterDefinition(
				"site",
				parameters.ParameterTypeString,
				parameters.WithHelp("Datadog site (env: DATADOG_CLI_SITE)"),
				parameters.WithDefault("datadoghq.com"),
			),
		),
	)
}

// DatadogSettings represents Datadog configuration parameters
type DatadogSettings struct {
	APIKey string `glazed.parameter:"api-key"`
	AppKey string `glazed.parameter:"app-key"`
	Site   string `glazed.parameter:"site"`
}

// Validate checks if the Datadog settings are valid
func (d *DatadogSettings) Validate() error {
	if d.APIKey == "" {
		return errors.New("api-key is required (set DATADOG_CLI_API_KEY environment variable)")
	}
	if d.AppKey == "" {
		return errors.New("app-key is required (set DATADOG_CLI_APP_KEY environment variable)")
	}
	return nil
}
