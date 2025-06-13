package cmds

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/client"
	datadog_layers "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/layers"
	"github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/render"
	dd_types "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// DatadogQueryCommand implements a YAML-driven Datadog logs query command
type DatadogQueryCommand struct {
	*cmds.CommandDescription
	Query      string                 `yaml:"query"`
	Subqueries dd_types.QueryMetadata `yaml:"subqueries"`
}

var _ cmds.GlazeCommand = (*DatadogQueryCommand)(nil)

// NewDatadogQueryCommand creates a new DatadogQueryCommand from a CommandDescription
func NewDatadogQueryCommand(description *cmds.CommandDescription, query string, subqueries dd_types.QueryMetadata) (*DatadogQueryCommand, error) {
	// Initialize Layers if nil (required for AppendLayers to work)
	if description.Layers == nil {
		description.Layers = layers.NewParameterLayers()
	}

	// Create parameter layers
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	datadogLayer, err := datadog_layers.NewDatadogParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create datadog parameter layer")
	}

	// Add layers to command description
	description.Layers.AppendLayers(datadogLayer, glazedLayer)

	return &DatadogQueryCommand{
		CommandDescription: description,
		Query:              query,
		Subqueries:         subqueries,
	}, nil
}

// RunIntoGlazeProcessor executes the Datadog query and streams results
func (d *DatadogQueryCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	log.Info().
		Str("command", "datadog_query").
		Msg("Starting Datadog query command execution")

	// Extract Datadog settings
	log.Debug().Msg("Extracting Datadog settings from parsed layers")
	ddSettings := &datadog_layers.DatadogSettings{}
	err := parsedLayers.InitializeStruct("datadog", ddSettings)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to initialize Datadog settings from layers")
		return errors.Wrap(err, "failed to initialize Datadog settings")
	}

	log.Debug().
		Bool("api_key_set", ddSettings.APIKey != "").
		Bool("app_key_set", ddSettings.AppKey != "").
		Str("site", ddSettings.Site).
		Msg("Datadog settings extracted successfully")

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

	// Get all parameters for template rendering
	log.Debug().Msg("Extracting parameters for template rendering")
	params := parsedLayers.GetDataMap()

	// Log parameter keys for debugging (without values to avoid exposing sensitive data)
	paramKeys := make([]string, 0, len(params))
	for key := range params {
		paramKeys = append(paramKeys, key)
	}
	log.Debug().
		Strs("parameter_keys", paramKeys).
		Msg("Parameters extracted for template rendering")

	// Parse time parameters if they're strings
	if fromStr, ok := params["from"].(string); ok && fromStr != "" {
		log.Debug().Str("from_param", fromStr).Msg("Parsing from time parameter")
		if fromTime, err := utils.ParseTimeParameter(fromStr); err == nil {
			params["from"] = fromTime
			log.Debug().Time("from_parsed", fromTime).Msg("From time parameter parsed successfully")
		} else {
			log.Warn().Err(err).Str("from_param", fromStr).Msg("Failed to parse from time parameter")
		}
	}
	if toStr, ok := params["to"].(string); ok && toStr != "" {
		log.Debug().Str("to_param", toStr).Msg("Parsing to time parameter")
		if toTime, err := utils.ParseTimeParameter(toStr); err == nil {
			params["to"] = toTime
			log.Debug().Time("to_parsed", toTime).Msg("To time parameter parsed successfully")
		} else {
			log.Warn().Err(err).Str("to_param", toStr).Msg("Failed to parse to time parameter")
		}
	}

	// Render the query
	log.Debug().
		Str("query_template", d.Query).
		Msg("Rendering query template")
	renderedQuery, err := render.RenderDatadogQuery(ctx, d.Query, params)
	if err != nil {
		log.Error().
			Err(err).
			Str("query_template", d.Query).
			Msg("Failed to render query template")
		return errors.Wrap(err, "failed to render query")
	}

	log.Debug().
		Str("rendered_query", renderedQuery.Query).
		Msg("Query template rendered successfully")

	// Apply subquery metadata
	log.Debug().Msg("Applying subquery metadata")
	if d.Subqueries.Sort != "" {
		renderedQuery.Sort = d.Subqueries.Sort
		log.Debug().Str("sort", d.Subqueries.Sort).Msg("Applied subquery sort")
	}
	if len(d.Subqueries.GroupBy) > 0 {
		renderedQuery.GroupBy = d.Subqueries.GroupBy
		log.Debug().Strs("group_by", d.Subqueries.GroupBy).Msg("Applied subquery group by")
	}
	if len(d.Subqueries.Aggs) > 0 {
		renderedQuery.Aggs = d.Subqueries.Aggs
		log.Debug().Int("aggs_count", len(d.Subqueries.Aggs)).Msg("Applied subquery aggregations")
	}

	// Validate the rendered query
	log.Debug().Str("query", renderedQuery.Query).Msg("Validating rendered query")
	err = render.ValidateQuery(renderedQuery.Query)
	if err != nil {
		log.Error().
			Err(err).
			Str("query", renderedQuery.Query).
			Msg("Query validation failed")
		return errors.Wrap(err, "invalid rendered query")
	}
	log.Debug().Msg("Query validation successful")

	log.Info().
		Str("query", renderedQuery.Query).
		Time("from", renderedQuery.From).
		Time("to", renderedQuery.To).
		Int("limit", renderedQuery.Limit).
		Msg("Executing Datadog logs search")

	// Execute the search
	log.Info().
		Str("final_query", renderedQuery.Query).
		Time("from", renderedQuery.From).
		Time("to", renderedQuery.To).
		Int("limit", renderedQuery.Limit).
		Str("sort", renderedQuery.Sort).
		Msg("Executing final Datadog logs search")

	err = client.ExecuteLogsSearch(ctx, ddClient, renderedQuery, func(ctx context.Context, row types.Row) error {
		return gp.AddRow(ctx, row)
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("query", renderedQuery.Query).
			Msg("Failed to execute logs search")
		return errors.Wrap(err, "failed to execute logs search")
	}

	log.Info().Msg("Datadog query command execution completed successfully")
	return nil
}
