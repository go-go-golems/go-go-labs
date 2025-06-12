package cmds

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"

	datadog "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/client"
	datadog_layers "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/layers"
	dd_types "github.com/go-go-golems/go-go-labs/cmd/apps/datadog-cli/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// AuthTestCommand tests authentication with Datadog API
type AuthTestCommand struct {
	*cmds.CommandDescription
}

// AuthTestSettings holds parameters for the auth test command
type AuthTestSettings struct {
	Verbose bool `glazed.parameter:"verbose"`
}

var _ cmds.GlazeCommand = (*AuthTestCommand)(nil)

func NewAuthTestCommand() (*AuthTestCommand, error) {
	log.Debug().Msg("Creating auth test command")
	
	description := cmds.NewCommandDescription("auth-test",
		cmds.WithShort("Test Datadog API authentication"),
		cmds.WithLong("Test the authentication credentials with Datadog API by making a simple request"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"verbose",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable verbose output"),
				parameters.WithDefault(false),
			),
		),
	)

	// Create parameter layers
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	datadogLayer, err := datadog_layers.NewDatadogParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create datadog parameter layer")
	}

	description.Layers.AppendLayers(datadogLayer, glazedLayer)

	return &AuthTestCommand{
		CommandDescription: description,
	}, nil
}

func (c *AuthTestCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	log.Info().Msg("Starting Datadog authentication test")

	// Extract Datadog settings
	ddSettings := &datadog_layers.DatadogSettings{}
	err := parsedLayers.InitializeStruct("datadog", ddSettings)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract Datadog settings")
		return errors.Wrap(err, "failed to initialize Datadog settings")
	}

	// Extract test command settings
	testSettings := &AuthTestSettings{}
	err = parsedLayers.InitializeStruct(layers.DefaultSlug, testSettings)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract test settings")
		return errors.Wrap(err, "failed to initialize test settings")
	}

	log.Info().
		Bool("api_key_set", ddSettings.APIKey != "").
		Bool("app_key_set", ddSettings.AppKey != "").
		Str("site", ddSettings.Site).
		Bool("verbose", testSettings.Verbose).
		Msg("Testing authentication with provided credentials")

	// Validate settings
	err = ddSettings.Validate()
	if err != nil {
		log.Error().Err(err).Msg("Datadog settings validation failed")
		return gp.AddRow(ctx, types.NewRow(
			types.MRP("status", "failed"),
			types.MRP("error", err.Error()),
			types.MRP("check", "validation"),
		))
	}

	// Create client
	ddClient, err := client.NewDatadogClient(client.DatadogConfig{
		APIKey:  ddSettings.APIKey,
		AppKey:  ddSettings.AppKey,
		Site:    ddSettings.Site,
		RawHTTP: ddSettings.RawHTTP,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Datadog client")
		return gp.AddRow(ctx, types.NewRow(
			types.MRP("status", "failed"),
			types.MRP("error", err.Error()),
			types.MRP("check", "client_creation"),
		))
	}

	// Test authentication with a simple API call
	log.Info().Msg("Testing authentication with Datadog API")
	
	auth := context.WithValue(
		ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {Key: ddSettings.APIKey},
			"appKeyAuth": {Key: ddSettings.AppKey},
		},
	)

	// Test with a simple logs query
	logsApi := datadogV2.NewLogsApi(ddClient)
	testRequest := datadogV2.LogsListRequest{
		Filter: &datadogV2.LogsQueryFilter{
			Query: datadog.PtrString("*"),
		},
		Page: &datadogV2.LogsListRequestPage{
			Limit: datadog.PtrInt32(1),
		},
	}

	opts := datadogV2.NewListLogsOptionalParameters().WithBody(testRequest)
	resp, httpResp, err := logsApi.ListLogs(auth, *opts)
	if err != nil {
		logResponseBodyOnError(httpResp, err, "authentication_test")
		return gp.AddRow(ctx, types.NewRow(
			types.MRP("status", "failed"),
			types.MRP("error", err.Error()),
			types.MRP("check", "api_call"),
			types.MRP("status_code", httpResp.StatusCode),
		))
	}

	log.Info().
		Int("status_code", httpResp.StatusCode).
		Msg("Authentication test successful")

	resultRow := types.NewRow(
		types.MRP("status", "success"),
		types.MRP("check", "api_call"),
		types.MRP("status_code", httpResp.StatusCode),
		types.MRP("api_key_prefix", maskKey(ddSettings.APIKey)),
		types.MRP("app_key_prefix", maskKey(ddSettings.AppKey)),
		types.MRP("site", ddSettings.Site),
	)

	if testSettings.Verbose && resp.Data != nil {
		resultRow.Set("logs_returned", len(resp.Data))
		if len(resp.Data) > 0 {
			resultRow.Set("first_log_id", *resp.Data[0].Id)
		}
	}

	return gp.AddRow(ctx, resultRow)
}

// LogsV1TestCommand tests Datadog Logs API v1
type LogsV1TestCommand struct {
	*cmds.CommandDescription
}

// LogsV1TestSettings holds parameters for the logs v1 test command
type LogsV1TestSettings struct {
	Query   string `glazed.parameter:"query"`
	Limit   int    `glazed.parameter:"limit"`
	Verbose bool   `glazed.parameter:"verbose"`
}

var _ cmds.GlazeCommand = (*LogsV1TestCommand)(nil)

func NewLogsV1TestCommand() (*LogsV1TestCommand, error) {
	log.Debug().Msg("Creating logs v1 test command")
	
	description := cmds.NewCommandDescription("logs-v1-test",
		cmds.WithShort("Test Datadog Logs API v1"),
		cmds.WithLong("Test the Datadog Logs API v1 with a simple query"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"query",
				parameters.ParameterTypeString,
				parameters.WithHelp("Log query to execute"),
				parameters.WithDefault("*"),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of logs to return"),
				parameters.WithDefault(5),
			),
			parameters.NewParameterDefinition(
				"verbose",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable verbose output"),
				parameters.WithDefault(false),
			),
		),
	)

	// Create parameter layers
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	datadogLayer, err := datadog_layers.NewDatadogParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create datadog parameter layer")
	}

	description.Layers.AppendLayers(datadogLayer, glazedLayer)

	return &LogsV1TestCommand{
		CommandDescription: description,
	}, nil
}

func (c *LogsV1TestCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	log.Info().Msg("Starting Datadog Logs API v1 test")

	// Extract Datadog settings
	ddSettings := &datadog_layers.DatadogSettings{}
	err := parsedLayers.InitializeStruct("datadog", ddSettings)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract Datadog settings")
		return errors.Wrap(err, "failed to initialize Datadog settings")
	}

	// Extract test command settings
	testSettings := &LogsV1TestSettings{}
	err = parsedLayers.InitializeStruct(layers.DefaultSlug, testSettings)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract test settings")
		return errors.Wrap(err, "failed to initialize test settings")
	}

	log.Info().
		Str("query", testSettings.Query).
		Int("limit", testSettings.Limit).
		Bool("verbose", testSettings.Verbose).
		Msg("Testing Logs API v1 with parameters")

	// Validate settings
	err = ddSettings.Validate()
	if err != nil {
		log.Error().Err(err).Msg("Datadog settings validation failed")
		return errors.Wrap(err, "Datadog settings validation failed")
	}

	// Create client
	ddClient, err := client.NewDatadogClient(client.DatadogConfig{
		APIKey:  ddSettings.APIKey,
		AppKey:  ddSettings.AppKey,
		Site:    ddSettings.Site,
		RawHTTP: ddSettings.RawHTTP,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Datadog client")
		return errors.Wrap(err, "failed to create Datadog client")
	}

	// Test Logs API v1
	auth := context.WithValue(
		ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {Key: ddSettings.APIKey},
			"appKeyAuth": {Key: ddSettings.AppKey},
		},
	)

	logsV1Api := datadogV1.NewLogsApi(ddClient)
	
	// Create time range for last hour
	to := time.Now()
	from := to.Add(-1 * time.Hour)
	
	log.Debug().
		Time("from", from).
		Time("to", to).
		Str("query", testSettings.Query).
		Msg("Making Logs API v1 request")

	body := datadogV1.LogsListRequest{
		Query: datadog.PtrString(testSettings.Query),
		Time: datadogV1.LogsListRequestTime{
			From: from,
			To:   to,
		},
		Limit: datadog.PtrInt32(int32(testSettings.Limit)),
		Sort:  (*datadogV1.LogsSort)(datadog.PtrString("desc")),
	}

	resp, httpResp, err := logsV1Api.ListLogs(auth, body)
	if err != nil {
		logResponseBodyOnError(httpResp, err, "logs_v1_test")
		return gp.AddRow(ctx, types.NewRow(
			types.MRP("api_version", "v1"),
			types.MRP("status", "failed"),
			types.MRP("error", err.Error()),
			types.MRP("status_code", httpResp.StatusCode),
		))
	}

	log.Info().
		Int("status_code", httpResp.StatusCode).
		Int("logs_count", len(resp.Logs)).
		Msg("Logs API v1 request successful")

	// Add summary row
	summaryRow := types.NewRow(
		types.MRP("api_version", "v1"),
		types.MRP("status", "success"),
		types.MRP("status_code", httpResp.StatusCode),
		types.MRP("logs_returned", len(resp.Logs)),
		types.MRP("query", testSettings.Query),
		types.MRP("time_range", fmt.Sprintf("%s to %s", from.Format(time.RFC3339), to.Format(time.RFC3339))),
	)

	if testSettings.Verbose && resp.NextLogId.IsSet() {
		summaryRow.Set("next_log_id", resp.NextLogId.Get())
	}

	err = gp.AddRow(ctx, summaryRow)
	if err != nil {
		return err
	}

	// Add individual log rows if verbose
	if testSettings.Verbose {
		for i, logEntry := range resp.Logs {
			if i >= 3 { // Limit to first 3 for verbose output
				break
			}
			logRow := types.NewRow(
				types.MRP("api_version", "v1"),
				types.MRP("type", "log_entry"),
				types.MRP("index", i),
			)
			
			if logEntry.Id != nil {
				logRow.Set("id", *logEntry.Id)
			}
			if logEntry.Content != nil && logEntry.Content.Message != nil {
				logRow.Set("message", *logEntry.Content.Message)
			}
			if logEntry.Content != nil && logEntry.Content.Timestamp != nil {
				logRow.Set("timestamp", logEntry.Content.Timestamp.Format(time.RFC3339))
			}
			if logEntry.Content != nil && logEntry.Content.Service != nil {
				logRow.Set("service", *logEntry.Content.Service)
			}
			if logEntry.Content != nil && logEntry.Content.Host != nil {
				logRow.Set("host", *logEntry.Content.Host)
			}

			err = gp.AddRow(ctx, logRow)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// LogsV2TestCommand tests Datadog Logs API v2
type LogsV2TestCommand struct {
	*cmds.CommandDescription
}

// LogsV2TestSettings holds parameters for the logs v2 test command
type LogsV2TestSettings struct {
	Query   string `glazed.parameter:"query"`
	Limit   int    `glazed.parameter:"limit"`
	Verbose bool   `glazed.parameter:"verbose"`
}

var _ cmds.GlazeCommand = (*LogsV2TestCommand)(nil)

func NewLogsV2TestCommand() (*LogsV2TestCommand, error) {
	log.Debug().Msg("Creating logs v2 test command")
	
	description := cmds.NewCommandDescription("logs-v2-test",
		cmds.WithShort("Test Datadog Logs API v2"),
		cmds.WithLong("Test the Datadog Logs API v2 with a simple query"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"query",
				parameters.ParameterTypeString,
				parameters.WithHelp("Log query to execute"),
				parameters.WithDefault("*"),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of logs to return"),
				parameters.WithDefault(5),
			),
			parameters.NewParameterDefinition(
				"verbose",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Enable verbose output"),
				parameters.WithDefault(false),
			),
		),
	)

	// Create parameter layers
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	datadogLayer, err := datadog_layers.NewDatadogParameterLayer()
	if err != nil {
		return nil, errors.Wrap(err, "could not create datadog parameter layer")
	}

	description.Layers.AppendLayers(datadogLayer, glazedLayer)

	return &LogsV2TestCommand{
		CommandDescription: description,
	}, nil
}

func (c *LogsV2TestCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	log.Info().Msg("Starting Datadog Logs API v2 test")

	// Extract Datadog settings
	ddSettings := &datadog_layers.DatadogSettings{}
	err := parsedLayers.InitializeStruct("datadog", ddSettings)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract Datadog settings")
		return errors.Wrap(err, "failed to initialize Datadog settings")
	}

	// Extract test command settings
	testSettings := &LogsV2TestSettings{}
	err = parsedLayers.InitializeStruct(layers.DefaultSlug, testSettings)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract test settings")
		return errors.Wrap(err, "failed to initialize test settings")
	}

	log.Info().
		Str("query", testSettings.Query).
		Int("limit", testSettings.Limit).
		Bool("verbose", testSettings.Verbose).
		Msg("Testing Logs API v2 with parameters")

	// Validate settings
	err = ddSettings.Validate()
	if err != nil {
		log.Error().Err(err).Msg("Datadog settings validation failed")
		return errors.Wrap(err, "Datadog settings validation failed")
	}

	// Create client
	ddClient, err := client.NewDatadogClient(client.DatadogConfig{
		APIKey:  ddSettings.APIKey,
		AppKey:  ddSettings.AppKey,
		Site:    ddSettings.Site,
		RawHTTP: ddSettings.RawHTTP,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Datadog client")
		return errors.Wrap(err, "failed to create Datadog client")
	}

	// Test using our existing v2 implementation
	renderedQuery := dd_types.DatadogQuery{
		Query: testSettings.Query,
		From:  time.Now().Add(-1 * time.Hour),
		To:    time.Now(),
		Limit: testSettings.Limit,
		Sort:  "desc",
	}

	log.Debug().
		Str("query", renderedQuery.Query).
		Time("from", renderedQuery.From).
		Time("to", renderedQuery.To).
		Int("limit", renderedQuery.Limit).
		Msg("Executing Logs API v2 test query")

	// Track results
	logsCount := 0
	err = client.ExecuteLogsSearch(ctx, ddClient, renderedQuery, func(ctx context.Context, row types.Row) error {
		logsCount++
		
		if testSettings.Verbose && logsCount <= 3 { // Show first 3 logs if verbose
			logRow := types.NewRow(
				types.MRP("api_version", "v2"),
				types.MRP("type", "log_entry"),
				types.MRP("index", logsCount-1),
			)
			
			// Copy relevant fields from the row
			fieldNames := []string{"id", "message", "timestamp", "service", "host"}
			for _, key := range fieldNames {
				if value, exists := row.Get(key); exists {
					logRow.Set(key, value)
				}
			}
			
			err := gp.AddRow(ctx, logRow)
			if err != nil {
				return err
			}
		}
		
		return nil
	})
	
	if err != nil {
		log.Error().Err(err).Str("operation", "logs_v2_test").Msg("Logs API v2 request failed")
		return gp.AddRow(ctx, types.NewRow(
			types.MRP("api_version", "v2"),
			types.MRP("status", "failed"),
			types.MRP("error", err.Error()),
		))
	}

	log.Info().
		Int("logs_count", logsCount).
		Msg("Logs API v2 request successful")

	// Add summary row
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("api_version", "v2"),
		types.MRP("status", "success"),
		types.MRP("logs_returned", logsCount),
		types.MRP("query", testSettings.Query),
		types.MRP("time_range", fmt.Sprintf("%s to %s", 
			renderedQuery.From.Format(time.RFC3339), 
			renderedQuery.To.Format(time.RFC3339))),
	))
}

// Helper function to mask API keys (same as in layers package)
func maskKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****"
}

// Helper function to log HTTP response body for debugging
func logResponseBodyOnError(httpResp *http.Response, err error, operation string) {
	if httpResp == nil {
		log.Error().Err(err).Str("operation", operation).Msg("HTTP response is nil")
		return
	}

	log.Error().
		Err(err).
		Str("operation", operation).
		Int("status_code", httpResp.StatusCode).
		Str("status", httpResp.Status).
		Msg("HTTP request failed")

	if httpResp.Body != nil {
		// Read response body for debugging
		bodyBytes, readErr := io.ReadAll(httpResp.Body)
		if readErr != nil {
			log.Error().Err(readErr).Msg("Failed to read response body")
		} else {
			log.Error().
				Str("response_body", string(bodyBytes)).
				Msg("Full response body for debugging")
		}
	}
}

// TestCmd is the parent command for all test commands
var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test commands for debugging Datadog CLI",
	Long:  "Collection of test commands to help debug authentication and API issues with the Datadog CLI",
}

func init() {
	log.Debug().Msg("Initializing test commands")
	
	// Create auth test command
	authTestCmd, err := NewAuthTestCommand()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create auth test command")
		fmt.Fprintf(os.Stderr, "Failed to create auth test command: %v\n", err)
		return
	}
	authTestCobraCmd, err := datadog_layers.BuildCobraCommandWithDatadogMiddlewares(authTestCmd)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build auth test cobra command")
		fmt.Fprintf(os.Stderr, "Failed to build auth test cobra command: %v\n", err)
		return
	}
	TestCmd.AddCommand(authTestCobraCmd)

	// Create logs v1 test command
	logsV1TestCmd, err := NewLogsV1TestCommand()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create logs v1 test command")
		fmt.Fprintf(os.Stderr, "Failed to create logs v1 test command: %v\n", err)
		return
	}
	logsV1TestCobraCmd, err := datadog_layers.BuildCobraCommandWithDatadogMiddlewares(logsV1TestCmd)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build logs v1 test cobra command")
		fmt.Fprintf(os.Stderr, "Failed to build logs v1 test cobra command: %v\n", err)
		return
	}
	TestCmd.AddCommand(logsV1TestCobraCmd)

	// Create logs v2 test command
	logsV2TestCmd, err := NewLogsV2TestCommand()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create logs v2 test command")
		fmt.Fprintf(os.Stderr, "Failed to create logs v2 test command: %v\n", err)
		return
	}
	logsV2TestCobraCmd, err := datadog_layers.BuildCobraCommandWithDatadogMiddlewares(logsV2TestCmd)
	if err != nil {
		log.Error().Err(err).Msg("Failed to build logs v2 test cobra command")
		fmt.Fprintf(os.Stderr, "Failed to build logs v2 test cobra command: %v\n", err)
		return
	}
	TestCmd.AddCommand(logsV2TestCobraCmd)

	log.Debug().Msg("Test commands initialized successfully")
}
