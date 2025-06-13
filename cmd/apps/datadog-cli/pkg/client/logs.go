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
	log.Debug().
		Str("query", query.Query).
		Time("from", query.From).
		Time("to", query.To).
		Int("limit", query.Limit).
		Str("sort", query.Sort).
		Msg("Starting Datadog logs search execution")

	logsApi := datadogV2.NewLogsApi(client)
	log.Debug().Msg("Datadog Logs API client initialized")

	// Build the search request
	searchRequest := datadogV2.LogsListRequest{
		Filter: &datadogV2.LogsQueryFilter{
			Query: &query.Query,
		},
	}
	log.Debug().Str("base_query", query.Query).Msg("Base search request created")

	// Set time range if provided
	if !query.From.IsZero() || !query.To.IsZero() {
		log.Debug().
			Time("from", query.From).
			Time("to", query.To).
			Msg("Setting time range for search")

		if searchRequest.Filter == nil {
			searchRequest.Filter = &datadogV2.LogsQueryFilter{}
		}
		if !query.From.IsZero() {
			from := query.From.Format(time.RFC3339)
			searchRequest.Filter.From = &from
			log.Debug().Str("from_formatted", from).Msg("Set from time")
		}
		if !query.To.IsZero() {
			to := query.To.Format(time.RFC3339)
			searchRequest.Filter.To = &to
			log.Debug().Str("to_formatted", to).Msg("Set to time")
		}
	}

	// Set limit if provided
	if query.Limit > 0 {
		limit := int32(query.Limit)
		searchRequest.Page = &datadogV2.LogsListRequestPage{
			Limit: &limit,
		}
		log.Debug().Int("limit", query.Limit).Msg("Set query limit")
	}

	// Set sort if provided
	if query.Sort != "" {
		searchRequest.Sort = (*datadogV2.LogsSort)(&query.Sort)
		log.Debug().Str("sort", query.Sort).Msg("Set query sort order")
	}

	log.Debug().
		Str("query", query.Query).
		Time("from", query.From).
		Time("to", query.To).
		Int("limit", query.Limit).
		Str("sort", query.Sort).
		Msg("Executing Datadog logs search")

	// Create auth context
	log.Debug().Msg("Creating authentication context for API request")
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
	pageNumber := 1

	log.Info().Msg("Starting paginated Datadog logs search")

	for {
		// Set cursor for pagination
		if cursor != nil {
			if searchRequest.Page == nil {
				searchRequest.Page = &datadogV2.LogsListRequestPage{}
			}
			searchRequest.Page.Cursor = cursor
			log.Debug().
				Str("cursor", *cursor).
				Int("page", pageNumber).
				Msg("Setting pagination cursor")
		}

		// Execute request
		log.Debug().
			Int("page", pageNumber).
			Int("total_processed", totalProcessed).
			Msg("Making API request to Datadog")

		opts := datadogV2.NewListLogsOptionalParameters().WithBody(searchRequest)
		resp, httpResp, err := logsApi.ListLogs(auth, *opts)
		if err != nil {
			logResponseBodyOnError(httpResp, err, "logs_search_v2")
			return errors.Wrap(err, "failed to execute logs search")
		}

		log.Debug().
			Int("status_code", httpResp.StatusCode).
			Int("page", pageNumber).
			Msg("API request completed")

		if httpResp.StatusCode != 200 {
			// Create a synthetic error for non-200 status codes
			statusErr := errors.Errorf("API request failed with status %d", httpResp.StatusCode)
			logResponseBodyOnError(httpResp, statusErr, "logs_search_v2_non_200")
			return statusErr
		}

		// Process logs
		pageLogsCount := 0
		if resp.Data != nil {
			pageLogsCount = len(resp.Data)
			log.Debug().
				Int("page", pageNumber).
				Int("logs_in_page", pageLogsCount).
				Msg("Processing logs from current page")

			for _, logEntry := range resp.Data {
				row := convertLogToRow(logEntry)
				err := processor(ctx, row)
				if err != nil {
					log.Error().
						Err(err).
						Int("total_processed", totalProcessed).
						Msg("Failed to process log entry")
					return errors.Wrap(err, "failed to process log entry")
				}
				totalProcessed++

				// Check for context cancellation
				select {
				case <-ctx.Done():
					log.Warn().
						Int("total_processed", totalProcessed).
						Msg("Context cancelled, stopping log processing")
					return ctx.Err()
				default:
				}
			}
		} else {
			log.Debug().Int("page", pageNumber).Msg("No logs returned in this page")
		}

		log.Debug().
			Int("page", pageNumber).
			Int("logs_in_page", pageLogsCount).
			Int("total_processed", totalProcessed).
			Msg("Completed processing page")

		// Check if we have more pages
		if resp.Meta == nil || resp.Meta.Page == nil || resp.Meta.Page.After == nil {
			log.Debug().Msg("No more pages available, stopping pagination")
			break
		}

		cursor = resp.Meta.Page.After
		log.Debug().
			Str("next_cursor", *cursor).
			Msg("Retrieved cursor for next page")

		// If we have a limit and reached it, stop
		if query.Limit > 0 && totalProcessed >= query.Limit {
			log.Debug().
				Int("total_processed", totalProcessed).
				Int("limit", query.Limit).
				Msg("Reached query limit, stopping pagination")
			break
		}

		pageNumber++
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
