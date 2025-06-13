package client

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"

	datadog "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// logResponseBodyOnError logs the HTTP response body for debugging when an error occurs
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

// DatadogConfig holds Datadog API configuration
type DatadogConfig struct {
	APIKey string
	AppKey string
	Site   string
	// RawHTTP enables raw, un-redacted HTTP debug output for the Datadog client
	RawHTTP bool
}

// NewDatadogClient creates a new Datadog API client with authentication
func NewDatadogClient(config DatadogConfig) (*datadog.APIClient, error) {
	log.Debug().
		Bool("api_key_set", config.APIKey != "").
		Str("api_key_prefix", maskKey(config.APIKey)).
		Bool("app_key_set", config.AppKey != "").
		Str("app_key_prefix", maskKey(config.AppKey)).
		Str("site", config.Site).
		Msg("Creating Datadog client")

	// Validate required fields
	if config.APIKey == "" {
		log.Error().Msg("API key is empty - cannot create Datadog client")
		return nil, errors.New("DATADOG_CLI_API_KEY is required")
	}
	if config.AppKey == "" {
		log.Error().Msg("App key is empty - cannot create Datadog client")
		return nil, errors.New("DATADOG_CLI_APP_KEY is required")
	}

	// Set default site if not provided
	if config.Site == "" {
		config.Site = "datadoghq.com"
		log.Debug().Str("site", config.Site).Msg("Using default Datadog site")
	}

	log.Debug().
		Str("site", config.Site).
		Msg("Initializing Datadog API client configuration")

	// Create configuration
	configuration := datadog.NewConfiguration()
	configuration.SetUnstableOperationEnabled("v2.SearchLogs", true)

	// Sanitize the site value in case it already contains a scheme or an api. prefix
	sanitizeSite := func(site string) string {
		site = strings.TrimSpace(site)
		// Remove scheme if present
		site = strings.TrimPrefix(site, "https://")
		site = strings.TrimPrefix(site, "http://")
		// Remove leading api. if present
		site = strings.TrimPrefix(site, "api.")
		// Remove any trailing slash
		site = strings.TrimSuffix(site, "/")
		return site
	}

	// Set the host based on the sanitized site (host only, without scheme)
	if config.Site != "" {
		normalizedSite := sanitizeSite(config.Site)
		host := "api." + normalizedSite
		configuration.Host = host
		log.Debug().Str("host", host).Msg("Set Datadog API host")
	}

	// Enable debug mode only if raw HTTP debug is requested
	configuration.Debug = config.RawHTTP
	log.Debug().Bool("raw_http_debug", config.RawHTTP).Msg("Enabled raw HTTP debug mode")

	// Create client
	apiClient := datadog.NewAPIClient(configuration)
	log.Debug().Msg("Datadog API client created successfully")

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

	log.Debug().Msg("Authentication context configured")

	// Test authentication by making a simple API call
	log.Debug().Msg("Testing authentication with Datadog API")
	testRequest := datadogV2.LogsListRequest{
		Page: &datadogV2.LogsListRequestPage{
			Limit: datadog.PtrInt32(1),
		},
	}
	opts := datadogV2.NewListLogsOptionalParameters().WithBody(testRequest)
	_, httpResp, err := datadogV2.NewLogsApi(apiClient).ListLogs(auth, *opts)
	if err != nil {
		logResponseBodyOnError(httpResp, err, "datadog_auth_test")
		return nil, errors.Wrap(err, "failed to authenticate with Datadog API")
	}

	log.Info().
		Int("status_code", getHTTPStatusCode(httpResp)).
		Msg("Datadog API authentication successful")

	return apiClient, nil
}

// maskKey masks API keys for logging, showing only prefix
func maskKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-4)
}

// getHTTPStatusCode safely extracts status code from HTTP response
func getHTTPStatusCode(resp interface{}) int {
	if resp == nil {
		return 0
	}
	// This is a simplified approach - the actual response type might vary
	return 0
}

// GetDatadogConfigFromEnv creates DatadogConfig from environment variables
func GetDatadogConfigFromEnv() DatadogConfig {
	// Check both DATADOG_CLI_* and DATADOG_* variants
	apiKey := os.Getenv("DATADOG_CLI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("DATADOG_API_KEY")
		log.Debug().Msg("Using DATADOG_API_KEY fallback")
	}

	appKey := os.Getenv("DATADOG_CLI_APP_KEY")
	if appKey == "" {
		appKey = os.Getenv("DATADOG_APP_KEY")
		log.Debug().Msg("Using DATADOG_APP_KEY fallback")
	}

	site := os.Getenv("DATADOG_CLI_SITE")
	if site == "" {
		site = os.Getenv("DATADOG_SITE")
		log.Debug().Msg("Using DATADOG_SITE fallback")
	}

	log.Debug().
		Bool("api_key_set", apiKey != "").
		Str("api_key_prefix", maskKey(apiKey)).
		Bool("app_key_set", appKey != "").
		Str("app_key_prefix", maskKey(appKey)).
		Str("site", site).
		Msg("Loading Datadog configuration from environment variables")

	// Log which environment variables were found
	if os.Getenv("DATADOG_CLI_API_KEY") != "" {
		log.Debug().Msg("Found DATADOG_CLI_API_KEY")
	}
	if os.Getenv("DATADOG_API_KEY") != "" {
		log.Debug().Msg("Found DATADOG_API_KEY")
	}
	if os.Getenv("DATADOG_CLI_APP_KEY") != "" {
		log.Debug().Msg("Found DATADOG_CLI_APP_KEY")
	}
	if os.Getenv("DATADOG_APP_KEY") != "" {
		log.Debug().Msg("Found DATADOG_APP_KEY")
	}
	if os.Getenv("DATADOG_CLI_SITE") != "" {
		log.Debug().Msg("Found DATADOG_CLI_SITE")
	}
	if os.Getenv("DATADOG_SITE") != "" {
		log.Debug().Msg("Found DATADOG_SITE")
	}

	return DatadogConfig{
		APIKey: apiKey,
		AppKey: appKey,
		Site:   site,
	}
}
