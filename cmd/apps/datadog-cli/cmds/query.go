package cmds

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/client"
	datadog_layers "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/layers"
	dd_types "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// QueryCommand executes raw Datadog search queries
type QueryCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*QueryCommand)(nil)

// NewQueryCommand creates a new command for executing raw Datadog queries
func NewQueryCommand() (*QueryCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	datadogLayer, err := datadog_layers.NewDatadogParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create datadog parameter layer")
	}

	return &QueryCommand{
		CommandDescription: cmds.NewCommandDescription(
			"query",
			cmds.WithShort("Execute a raw Datadog search query"),
			cmds.WithLong(`Execute a raw Datadog search query without YAML templating.
This command allows you to run direct Datadog log search queries using the search syntax.

Examples:
  datadog-cli logs query "service:web-api AND status:error"
  datadog-cli logs query "host:prod-* AND @timestamp:[now-1h TO now]"
  datadog-cli logs query "@error.message:*timeout*" --limit 50
`),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"search-query",
					parameters.ParameterTypeString,
					parameters.WithHelp("Datadog search query string"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"from",
					parameters.ParameterTypeString,
					parameters.WithDefault("-1h"),
					parameters.WithHelp("Start time (relative like -1h or absolute)"),
				),
				parameters.NewParameterDefinition(
					"to",
					parameters.ParameterTypeString,
					parameters.WithDefault("now"),
					parameters.WithHelp("End time (relative like now or absolute)"),
				),
				parameters.NewParameterDefinition(
					"limit",
					parameters.ParameterTypeInteger,
					parameters.WithDefault(100),
					parameters.WithHelp("Number of results to return"),
				),
				parameters.NewParameterDefinition(
					"sort",
					parameters.ParameterTypeString,
					parameters.WithDefault("desc"),
					parameters.WithHelp("Sort order (asc or desc)"),
				),
			),
			cmds.WithLayersList(datadogLayer, glazedLayer),
		),
	}, nil
}

// QuerySettings represents the settings for the query command
type QuerySettings struct {
	SearchQuery string `glazed.parameter:"search-query"`
	From        string `glazed.parameter:"from"`
	To          string `glazed.parameter:"to"`
	Limit       int    `glazed.parameter:"limit"`
	Sort        string `glazed.parameter:"sort"`
}

// RunIntoGlazeProcessor executes the raw Datadog query
func (q *QueryCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	log.Info().
		Str("command", "raw_query").
		Msg("Starting raw Datadog query execution")
	
	// Extract settings
	log.Debug().Msg("Extracting query settings from parsed layers")
	settings := &QuerySettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to initialize query settings")
		return errors.Wrap(err, "failed to initialize query settings")
	}

	log.Debug().
		Str("search_query", settings.SearchQuery).
		Str("from", settings.From).
		Str("to", settings.To).
		Int("limit", settings.Limit).
		Str("sort", settings.Sort).
		Msg("Query settings extracted successfully")

	// Extract Datadog settings
	log.Debug().Msg("Extracting Datadog settings from parsed layers")
	ddSettings := &datadog_layers.DatadogSettings{}
	err = parsedLayers.InitializeStruct("datadog", ddSettings)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to initialize Datadog settings")
		return errors.Wrap(err, "failed to initialize Datadog settings")
	}

	// Validate Datadog settings
	log.Debug().Msg("Validating Datadog settings")
	err = ddSettings.Validate()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Datadog settings validation failed")
		return err
	}

	// Create Datadog client
	log.Debug().Msg("Creating Datadog API client")
	ddClient, err := client.NewDatadogClient(client.DatadogConfig{
		APIKey:  ddSettings.APIKey,
		AppKey:  ddSettings.AppKey,
		Site:    ddSettings.Site,
		RawHTTP: ddSettings.RawHTTP,
	})
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to create Datadog client")
		return errors.Wrap(err, "failed to create Datadog client")
	}
	log.Info().Msg("Datadog client created successfully")

	// Parse time parameters
	log.Debug().Str("from_param", settings.From).Msg("Parsing from time parameter")
	fromTime, err := utils.ParseTimeParameter(settings.From)
	if err != nil {
		log.Error().
			Err(err).
			Str("from_param", settings.From).
			Msg("Failed to parse from time parameter")
		return errors.Wrapf(err, "failed to parse from time: %s", settings.From)
	}
	log.Debug().Time("from_parsed", fromTime).Msg("From time parameter parsed successfully")

	log.Debug().Str("to_param", settings.To).Msg("Parsing to time parameter")
	toTime, err := utils.ParseTimeParameter(settings.To)
	if err != nil {
		log.Error().
			Err(err).
			Str("to_param", settings.To).
			Msg("Failed to parse to time parameter")
		return errors.Wrapf(err, "failed to parse to time: %s", settings.To)
	}
	log.Debug().Time("to_parsed", toTime).Msg("To time parameter parsed successfully")

	// Create the query
	log.Debug().Msg("Creating Datadog query structure")
	query := dd_types.DatadogQuery{
		Query: settings.SearchQuery,
		From:  fromTime,
		To:    toTime,
		Limit: settings.Limit,
		Sort:  settings.Sort,
	}
	
	log.Debug().
		Str("query", query.Query).
		Time("from", query.From).
		Time("to", query.To).
		Int("limit", query.Limit).
		Str("sort", query.Sort).
		Msg("Query structure created successfully")

	log.Info().
		Str("query", query.Query).
		Time("from", query.From).
		Time("to", query.To).
		Int("limit", query.Limit).
		Str("sort", query.Sort).
		Msg("Executing raw Datadog query")

	// Execute the search
	log.Info().
		Str("raw_query", query.Query).
		Time("from", query.From).
		Time("to", query.To).
		Int("limit", query.Limit).
		Str("sort", query.Sort).
		Msg("Executing raw Datadog query")
	
	err = client.ExecuteLogsSearch(ctx, ddClient, query, func(ctx context.Context, row types.Row) error {
		return gp.AddRow(ctx, row)
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("query", query.Query).
			Msg("Failed to execute raw logs search")
		return errors.Wrap(err, "failed to execute logs search")
	}

	log.Info().Msg("Raw Datadog query execution completed successfully")
	return nil
}


