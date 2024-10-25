package main

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

func createDemoParsedLayers() *layers.ParsedLayers {
	// Create ParsedLayers
	parsedLayers := layers.NewParsedLayers()

	// Layer 1: User Configuration
	userConfigLayer, _ := layers.NewParameterLayer("user-config", "User Configuration",
		layers.WithDescription("User-specific configuration options"),
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("username", parameters.ParameterTypeString,
				parameters.WithHelp("User's username"),
				parameters.WithDefault("guest"),
			),
			parameters.NewParameterDefinition("theme", parameters.ParameterTypeString,
				parameters.WithHelp("UI theme"),
				parameters.WithDefault("light"),
				parameters.WithChoices("light", "dark"),
			),
		),
	)

	parsedUserConfigLayer, _ := layers.NewParsedLayer(userConfigLayer,
		layers.WithParsedParameterValue("username", "john_doe"),
		layers.WithParsedParameterValue("theme", "dark"),
	)
	parsedLayers.Set("user-config", parsedUserConfigLayer)

	// Layer 2: Application Settings
	appSettingsLayer, _ := layers.NewParameterLayer("app-settings", "Application Settings",
		layers.WithDescription("General application settings"),
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("verbose", parameters.ParameterTypeBool,
				parameters.WithHelp("Enable verbose logging"),
				parameters.WithDefault(false),
			),
			parameters.NewParameterDefinition("max-connections", parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of concurrent connections"),
				parameters.WithDefault(10),
			),
		),
	)

	parsedAppSettingsLayer, _ := layers.NewParsedLayer(appSettingsLayer,
		layers.WithParsedParameterValue("verbose", true),
		layers.WithParsedParameterValue("max-connections", 20),
	)
	parsedLayers.Set("app-settings", parsedAppSettingsLayer)

	// Layer 3: Output Configuration
	outputConfigLayer, _ := layers.NewParameterLayer("output-config", "Output Configuration",
		layers.WithDescription("Output formatting and destination options"),
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("format", parameters.ParameterTypeString,
				parameters.WithHelp("Output format"),
				parameters.WithDefault("json"),
				parameters.WithChoices("json", "yaml", "csv"),
			),
			parameters.NewParameterDefinition("output-file", parameters.ParameterTypeString,
				parameters.WithHelp("Path to output file"),
				parameters.WithDefault("output.txt"),
			),
		),
	)

	parsedOutputConfigLayer, _ := layers.NewParsedLayer(outputConfigLayer,
		layers.WithParsedParameterValue("format", "yaml"),
		layers.WithParsedParameterValue("output-file", "result.yaml"),
	)
	parsedLayers.Set("output-config", parsedOutputConfigLayer)

	return parsedLayers
}
