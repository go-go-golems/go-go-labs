package client

import (
	"context"
	"time"

	datadog "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/go-go-golems/glazed/pkg/types"
	dd_types "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ExecuteLogsSearch executes a Datadog logs search and streams results to a processor
func ExecuteLogsSearch(
	ctx context.Context,
	client *datadog.APIClient,
	query dd_types.DatadogQuery,
	processor func(context.Context, types.Row) error,
) error {
	logsApi := datadogV2.NewLogsApi(client)

	// Build the search request
	searchRequest := datadogV2.LogsListRequest{
		Filter: &datadogV2.LogsQueryFilter{
			Query: &query.Query,
		},
	}

	// Set time range if provided
	if !query.From.IsZero() || !query.To.IsZero() {
		if searchRequest.Filter == nil {
			searchRequest.Filter = &datadogV2.LogsQueryFilter{}
		}
		if !query.From.IsZero() {
			from := query.From.Format(time.RFC3339)
			searchRequest.Filter.From = &from
		}
		if !query.To.IsZero() {
			to := query.To.Format(time.RFC3339)
			searchRequest.Filter.To = &to
		}
	}

	// Set limit if provided
	if query.Limit > 0 {
		limit := int32(query.Limit)
		searchRequest.Page = &datadogV2.LogsListRequestPage{
			Limit: &limit,
		}
	}

	// Set sort if provided
	if query.Sort != "" {
		searchRequest.Sort = (*datadogV2.LogsSort)(&query.Sort)
	}

	log.Debug().
		Str("query", query.Query).
		Time("from", query.From).
		Time("to", query.To).
		Int("limit", query.Limit).
		Str("sort", query.Sort).
		Msg("Executing Datadog logs search")

	// Create auth context
	auth := context.WithValue(
		ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {Key: ""}, // Will be set by client
			"appKeyAuth": {Key: ""}, // Will be set by client
		},
	)

	// Execute the search with pagination
	var cursor *string
	totalProcessed := 0

	for {
		// Set cursor for pagination
		if cursor != nil {
			if searchRequest.Page == nil {
				searchRequest.Page = &datadogV2.LogsListRequestPage{}
			}
			searchRequest.Page.Cursor = cursor
		}

		// Execute request
		opts := datadogV2.NewListLogsOptionalParameters().WithBody(searchRequest)
		resp, httpResp, err := logsApi.ListLogs(auth, *opts)
		if err != nil {
			return errors.Wrap(err, "failed to execute logs search")
		}

		if httpResp.StatusCode != 200 {
			return errors.Errorf("API request failed with status %d", httpResp.StatusCode)
		}

		// Process logs
		if resp.Data != nil {
			for _, logEntry := range resp.Data {
				row := convertLogToRow(logEntry)
				err := processor(ctx, row)
				if err != nil {
					return errors.Wrap(err, "failed to process log entry")
				}
				totalProcessed++

				// Check for context cancellation
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
			}
		}

		// Check if we have more pages
		if resp.Meta == nil || resp.Meta.Page == nil || resp.Meta.Page.After == nil {
			break
		}

		cursor = resp.Meta.Page.After

		// If we have a limit and reached it, stop
		if query.Limit > 0 && totalProcessed >= query.Limit {
			break
		}
	}

	log.Info().Int("totalProcessed", totalProcessed).Msg("Finished processing logs")
	return nil
}

// convertLogToRow converts a Datadog log entry to a Glazed row
func convertLogToRow(logEntry datadogV2.Log) types.Row {
	row := types.NewRow()

	// Set log ID
	if logEntry.Id != nil {
		row.Set("id", *logEntry.Id)
	}

	// Set log type
	if logEntry.Type != nil {
		row.Set("type", string(*logEntry.Type))
	}

	// Process attributes
	if logEntry.Attributes != nil {
		attrs := logEntry.Attributes

		// Timestamp
		if attrs.Timestamp != nil {
			row.Set("timestamp", attrs.Timestamp.Format(time.RFC3339))
		}

		// Status
		if attrs.Status != nil {
			row.Set("status", *attrs.Status)
		}

		// Message
		if attrs.Message != nil {
			row.Set("message", *attrs.Message)
		}

		// Hostname
		if attrs.Host != nil {
			row.Set("host", *attrs.Host)
		}

		// Service
		if attrs.Service != nil {
			row.Set("service", *attrs.Service)
		}

		// Source (from general attributes)
		if attrs.Attributes != nil {
			if source, ok := attrs.Attributes["ddsource"]; ok {
				row.Set("source", source)
			}
		}

		// Tags
		if attrs.Tags != nil {
			for _, tag := range attrs.Tags {
				row.Set("tag", tag)
			}
		}

		// Custom attributes (from AdditionalProperties)
		if attrs.AdditionalProperties != nil {
			for key, value := range attrs.AdditionalProperties {
				row.Set("custom_"+key, value)
			}
		}
	}

	return row
}
