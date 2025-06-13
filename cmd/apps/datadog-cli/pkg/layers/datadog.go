package layers

import (
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const DatadogSlug = "datadog"

// NewDatadogParameterLayer creates a parameter layer for Datadog authentication
func NewDatadogParameterLayer() (layers.ParameterLayer, error) {
	log.Debug().Msg("Creating Datadog parameter layer")

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
			parameters.NewParameterDefinition(
				"raw-http-debug",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable raw, un-redacted HTTP debug output"),
			),
		),
	)
}

// maskKey masks API keys for logging, showing only prefix
func maskKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-4)
}

// DatadogSettings represents Datadog configuration parameters
type DatadogSettings struct {
	APIKey  string `glazed.parameter:"api-key"`
	AppKey  string `glazed.parameter:"app-key"`
	Site    string `glazed.parameter:"site"`
	RawHTTP bool   `glazed.parameter:"raw-http-debug"`
}

// Validate checks if the Datadog settings are valid
func (d *DatadogSettings) Validate() error {
	log.Debug().
		Bool("api_key_set", d.APIKey != "").
		Str("api_key_prefix", maskKey(d.APIKey)).
		Bool("app_key_set", d.AppKey != "").
		Str("app_key_prefix", maskKey(d.AppKey)).
		Str("site", d.Site).
		Msg("Validating Datadog settings")

	if d.APIKey == "" {
		log.Error().
			Msg("API key is missing - check DATADOG_CLI_API_KEY or DATADOG_API_KEY environment variables")
		return errors.New("api-key is required (set DATADOG_CLI_API_KEY or DATADOG_API_KEY environment variable)")
	}
	if d.AppKey == "" {
		log.Error().
			Msg("App key is missing - check DATADOG_CLI_APP_KEY or DATADOG_APP_KEY environment variables")
		return errors.New("app-key is required (set DATADOG_CLI_APP_KEY or DATADOG_APP_KEY environment variable)")
	}

	if d.Site == "" {
		log.Warn().Msg("Site is empty, using default datadoghq.com")
		d.Site = "datadoghq.com"
	}

	log.Debug().Msg("Datadog settings validation successful")
	return nil
}
