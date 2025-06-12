package cmds

import (
	"context"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/client"
	datadog_layers "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/layers"
	"github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/render"
	dd_types "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/types"
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
	// Extract Datadog settings
	ddSettings := &datadog_layers.DatadogSettings{}
	err := parsedLayers.InitializeStruct("datadog", ddSettings)
	if err != nil {
		return errors.Wrap(err, "failed to initialize Datadog settings")
	}

	// Validate Datadog settings
	err = ddSettings.Validate()
	if err != nil {
		return err
	}

	// Create Datadog client
	ddClient, err := client.NewDatadogClient(client.DatadogConfig{
		APIKey: ddSettings.APIKey,
		AppKey: ddSettings.AppKey,
		Site:   ddSettings.Site,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create Datadog client")
	}

	// Get all parameters for template rendering
	params := parsedLayers.GetDataMap()

	// Parse time parameters if they're strings
	if fromStr, ok := params["from"].(string); ok && fromStr != "" {
		if fromTime, err := parseTimeParameter(fromStr); err == nil {
			params["from"] = fromTime
		}
	}
	if toStr, ok := params["to"].(string); ok && toStr != "" {
		if toTime, err := parseTimeParameter(toStr); err == nil {
			params["to"] = toTime
		}
	}

	// Render the query
	renderedQuery, err := render.RenderDatadogQuery(ctx, d.Query, params)
	if err != nil {
		return errors.Wrap(err, "failed to render query")
	}

	// Apply subquery metadata
	if d.Subqueries.Sort != "" {
		renderedQuery.Sort = d.Subqueries.Sort
	}
	if len(d.Subqueries.GroupBy) > 0 {
		renderedQuery.GroupBy = d.Subqueries.GroupBy
	}
	if len(d.Subqueries.Aggs) > 0 {
		renderedQuery.Aggs = d.Subqueries.Aggs
	}

	// Validate the rendered query
	err = render.ValidateQuery(renderedQuery.Query)
	if err != nil {
		return errors.Wrap(err, "invalid rendered query")
	}

	log.Info().
		Str("query", renderedQuery.Query).
		Time("from", renderedQuery.From).
		Time("to", renderedQuery.To).
		Int("limit", renderedQuery.Limit).
		Msg("Executing Datadog logs search")

	// Execute the search
	err = client.ExecuteLogsSearch(ctx, ddClient, renderedQuery, func(ctx context.Context, row types.Row) error {
		return gp.AddRow(ctx, row)
	})
	if err != nil {
		return errors.Wrap(err, "failed to execute logs search")
	}

	return nil
}

// parseTimeParameter parses various time formats including relative times
func parseTimeParameter(timeStr string) (time.Time, error) {
	// Handle "now"
	if timeStr == "now" {
		return time.Now(), nil
	}

	// Handle relative times like "-1h", "-30m", "-1d"
	if timeStr[0] == '-' || timeStr[0] == '+' {
		duration, err := time.ParseDuration(timeStr)
		if err != nil {
			// Try parsing as days (not supported by time.ParseDuration)
			if timeStr[len(timeStr)-1] == 'd' {
				days := timeStr[:len(timeStr)-1]
				if d, err := time.ParseDuration(days + "h"); err == nil {
					duration = d * 24
				} else {
					return time.Time{}, err
				}
			} else {
				return time.Time{}, err
			}
		}
		return time.Now().Add(duration), nil
	}

	// Try various time formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.Errorf("unable to parse time: %s", timeStr)
}
