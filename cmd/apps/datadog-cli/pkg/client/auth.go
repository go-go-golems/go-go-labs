package client

import (
	"context"
	"os"

	datadog "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/pkg/errors"
)

// DatadogConfig holds Datadog API configuration
type DatadogConfig struct {
	APIKey string
	AppKey string
	Site   string
}

// NewDatadogClient creates a new Datadog API client with authentication
func NewDatadogClient(config DatadogConfig) (*datadog.APIClient, error) {
	// Validate required fields
	if config.APIKey == "" {
		return nil, errors.New("DATADOG_CLI_API_KEY is required")
	}
	if config.AppKey == "" {
		return nil, errors.New("DATADOG_CLI_APP_KEY is required")
	}

	// Set default site if not provided
	if config.Site == "" {
		config.Site = "datadoghq.com"
	}

	// Create configuration
	configuration := datadog.NewConfiguration()
	configuration.SetUnstableOperationEnabled("v2.SearchLogs", true)

	// Create client
	apiClient := datadog.NewAPIClient(configuration)

	// Set authentication
	auth := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: config.APIKey,
			},
			"appKeyAuth": {
				Key: config.AppKey,
			},
		},
	)

	// Test authentication by making a simple API call
	testRequest := datadogV2.LogsListRequest{
		Page: &datadogV2.LogsListRequestPage{
			Limit: datadog.PtrInt32(1),
		},
	}
	opts := datadogV2.NewListLogsOptionalParameters().WithBody(testRequest)
	_, _, err := datadogV2.NewLogsApi(apiClient).ListLogs(auth, *opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to authenticate with Datadog API")
	}

	return apiClient, nil
}

// GetDatadogConfigFromEnv creates DatadogConfig from environment variables
func GetDatadogConfigFromEnv() DatadogConfig {
	return DatadogConfig{
		APIKey: os.Getenv("DATADOG_CLI_API_KEY"),
		AppKey: os.Getenv("DATADOG_CLI_APP_KEY"),
		Site:   os.Getenv("DATADOG_CLI_SITE"),
	}
}
