package layers

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func BuildCobraCommandWithDatadogMiddlewares(
	cmd cmds.Command,
	options ...cli.CobraParserOption,
) (*cobra.Command, error) {
	options_ := append([]cli.CobraParserOption{
		cli.WithCobraMiddlewaresFunc(GetCobraCommandDatadogMiddlewares),
		cli.WithCobraShortHelpLayers(layers.DefaultSlug, DatadogSlug),
		cli.WithCreateCommandSettingsLayer(),
		cli.WithProfileSettingsLayer(),
	}, options...)

	return cli.BuildCobraCommandFromCommand(cmd, options_...)
}

func GetCobraCommandDatadogMiddlewares(
	parsedCommandLayers *layers.ParsedLayers,
	cmd *cobra.Command,
	args []string,
) ([]middlewares.Middleware, error) {
	log.Debug().
		Str("command", cmd.Name()).
		Strs("args", args).
		Msg("Building Datadog command middlewares")

	// Start with cobra-specific middlewares
	middlewares_ := []middlewares.Middleware{
		middlewares.ParseFromCobraCommand(cmd,
			parameters.WithParseStepSource("cobra"),
		),
		middlewares.GatherArguments(args,
			parameters.WithParseStepSource("arguments"),
		),
	}
	log.Debug().Msg("Added cobra and arguments middlewares")

	// Get the common datadog middlewares
	additionalMiddlewares, err := GetDatadogMiddlewares(parsedCommandLayers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Datadog middlewares")
		return nil, err
	}
	middlewares_ = append(middlewares_, additionalMiddlewares...)
	log.Debug().
		Int("total_middlewares", len(middlewares_)).
		Msg("Built complete middleware chain")

	return middlewares_, nil
}

// GetDatadogMiddlewares returns the common middleware chain used by datadog commands
func GetDatadogMiddlewares(
	parsedCommandLayers *layers.ParsedLayers,
) ([]middlewares.Middleware, error) {
	log.Debug().Msg("Creating Datadog middleware chain")
	
	commandSettings := &cli.CommandSettings{}
	err := parsedCommandLayers.InitializeStruct(cli.CommandSettingsSlug, commandSettings)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize command settings")
		return nil, err
	}
	log.Debug().
		Str("load_parameters_from_file", commandSettings.LoadParametersFromFile).
		Msg("Command settings initialized")
	
	middlewares_ := []middlewares.Middleware{}

	if commandSettings.LoadParametersFromFile != "" {
		log.Debug().
			Str("file", commandSettings.LoadParametersFromFile).
			Msg("Adding load parameters from file middleware")
		middlewares_ = append(middlewares_,
			middlewares.LoadParametersFromFile(commandSettings.LoadParametersFromFile))
	}

	profileSettings := &cli.ProfileSettings{}
	err = parsedCommandLayers.InitializeStruct(cli.ProfileSettingsSlug, profileSettings)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize profile settings")
		return nil, err
	}
	log.Debug().
		Str("profile_file", profileSettings.ProfileFile).
		Str("profile", profileSettings.Profile).
		Msg("Profile settings initialized")

	xdgConfigPath, err := os.UserConfigDir()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user config directory")
		return nil, err
	}

	defaultProfileFile := fmt.Sprintf("%s/datadog-cli/profiles.yaml", xdgConfigPath)
	if profileSettings.ProfileFile == "" {
		profileSettings.ProfileFile = defaultProfileFile
	}
	if profileSettings.Profile == "" {
		profileSettings.Profile = "default"
	}
	log.Debug().
		Str("default_profile_file", defaultProfileFile).
		Str("final_profile_file", profileSettings.ProfileFile).
		Str("final_profile", profileSettings.Profile).
		Msg("Adding profiles middleware")
	
	middlewares_ = append(middlewares_,
		middlewares.GatherFlagsFromProfiles(
			defaultProfileFile,
			profileSettings.ProfileFile,
			profileSettings.Profile,
			parameters.WithParseStepSource("profiles"),
			parameters.WithParseStepMetadata(map[string]interface{}{
				"profileFile": profileSettings.ProfileFile,
				"profile":     profileSettings.Profile,
			}),
		),
	)

	log.Debug().
		Strs("whitelisted_layers", []string{DatadogSlug}).
		Msg("Adding viper middleware with whitelisted layers")
	
	middlewares_ = append(middlewares_,
		middlewares.WrapWithWhitelistedLayers(
			[]string{
				DatadogSlug,
			},
			middlewares.GatherFlagsFromViper(parameters.WithParseStepSource("viper")),
		),
		middlewares.SetFromDefaults(parameters.WithParseStepSource("defaults")),
	)

	log.Debug().
		Int("total_middlewares", len(middlewares_)).
		Msg("Datadog middleware chain completed")
	return middlewares_, nil
}
