package n8n

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// N8NAPISettings holds the common settings for n8n API calls
type N8NAPISettings struct {
	BaseURL string `glazed.parameter:"base-url"`
	APIKey  string `glazed.parameter:"api-key"`
}

// GetN8NAPISettingsFromParsedLayers extracts API settings from parsed layers
func GetN8NAPISettingsFromParsedLayers(parsedLayers *layers.ParsedLayers) (*N8NAPISettings, error) {
	s := &N8NAPISettings{}
	if err := parsedLayers.InitializeStruct("n8n-api", s); err != nil {
		return nil, err
	}
	return s, nil
}

// NewN8NAPILayer creates a new parameter layer for n8n API parameters
func NewN8NAPILayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(
		"n8n-api",
		"n8n API connection parameters",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"base-url",
				parameters.ParameterTypeString,
				parameters.WithHelp("Base URL of n8n instance"),
				parameters.WithDefault("http://localhost:5678"),
			),
			parameters.NewParameterDefinition(
				"api-key",
				parameters.ParameterTypeString,
				parameters.WithHelp("API Key for n8n instance"),
				parameters.WithRequired(true),
			),
		),
	)
}
