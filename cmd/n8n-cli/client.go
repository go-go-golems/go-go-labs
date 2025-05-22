package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
)

// N8NClient represents a client for interacting with the n8n REST API
type N8NClient struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

// NewN8NClient creates a new n8n API client
func NewN8NClient(baseURL, apiKey string) *N8NClient {
	return &N8NClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client:  &http.Client{},
	}
}

// DoRequest makes an HTTP request to the n8n API
func (c *N8NClient) DoRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s/api/v1/%s", c.BaseURL, endpoint)
	log.Debug().Str("method", method).Str("url", url).Msg("API request")

	// Copy and log request body at TRACE level
	var bodyToSend io.Reader
	if body != nil {
		// Read the entire body
		bodyBytes, err := io.ReadAll(body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read request body")
			return nil, err
		}

		// Log at TRACE level
		log.Trace().Str("body", string(bodyBytes)).Msg("Request body")

		// Create a new reader with the same content
		bodyToSend = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyToSend)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create request")
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-N8N-API-KEY", c.APIKey)

	// Log all request headers at TRACE level
	log.Trace().Interface("headers", req.Header).Msg("Request headers")

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("HTTP request failed")
		return nil, err
	}

	// Always read the response body so we can log it and make it available for reuse
	bodyBytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Error().Err(err).Msg("Failed to read response body")
		return nil, err
	}

	// Recreate response body
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Log response status
	logEvent := log.Debug().Int("status", resp.StatusCode).Str("status_text", resp.Status)

	// For non-2xx responses, log the body at DEBUG level
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logEvent = log.Error().Int("status", resp.StatusCode).Str("status_text", resp.Status)
		logEvent.Str("body", string(bodyBytes)).Msg("Error response")
	} else {
		// For successful responses, log body at TRACE level
		logEvent.Msg("Response received")
		log.Trace().Str("body", string(bodyBytes)).Msg("Response body")
	}

	log.Trace().Interface("headers", resp.Header).Msg("Response headers")

	return resp, nil
}

// ReadFile reads a JSON file and unmarshals it into the provided interface
func ReadJSONFile(filePath string, v interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// WriteJSONFile writes an interface to a JSON file with pretty formatting
func WriteJSONFile(filePath string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// GetWorkflow retrieves a workflow by ID
func (c *N8NClient) GetWorkflow(id string) (map[string]interface{}, error) {
	resp, err := c.DoRequest("GET", fmt.Sprintf("workflows/%s", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var workflow map[string]interface{}
	if err := json.Unmarshal(data, &workflow); err != nil {
		return nil, err
	}

	return workflow, nil
}

// ListWorkflows retrieves all workflows with optional filtering
func (c *N8NClient) ListWorkflows(active bool, limit int, cursor string) ([]map[string]interface{}, string, error) {
	// Build the endpoint with the correct pagination parameters (cursor-based, not offset-based)
	endpoint := fmt.Sprintf("workflows?limit=%d", limit)
	if active {
		endpoint += "&active=true"
	}
	if cursor != "" {
		endpoint += fmt.Sprintf("&cursor=%s", cursor)
	}

	log.Debug().Bool("active", active).Int("limit", limit).Str("cursor", cursor).Str("endpoint", endpoint).Msg("Listing workflows")
	resp, err := c.DoRequest("GET", endpoint, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list workflows")
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Int("status", resp.StatusCode).Str("endpoint", endpoint).Msg("List workflows failed with non-200 status")
		return nil, "", fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read response body")
		return nil, "", err
	}

	// Always log response body at TRACE level
	log.Trace().Msgf("Response body: %s", string(data))

	// Log the response size and content type
	log.Debug().Int("body_size", len(data)).Str("content_type", resp.Header.Get("Content-Type")).Msg("Received response")

	// Parse the response which should include data and nextCursor
	var result struct {
		Data       []map[string]interface{} `json:"data"`
		NextCursor string                   `json:"nextCursor"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		log.Debug().Err(err).Str("raw_data", string(data)).Msg("Failed to parse response as {data:[]} structure, trying direct array")
		// Try parsing as direct array for older n8n versions
		var workflows []map[string]interface{}
		if jsonErr := json.Unmarshal(data, &workflows); jsonErr != nil {
			log.Error().Err(jsonErr).Str("raw_data", string(data)).Msg("Failed to parse response as direct array")
			return nil, "", fmt.Errorf("failed to parse JSON response: %w, raw response: %s", err, string(data))
		}
		log.Debug().Int("count", len(workflows)).Msg("Successfully parsed as direct array")
		return workflows, "", nil
	}

	log.Debug().Int("count", len(result.Data)).Str("next_cursor", result.NextCursor).Msg("Successfully parsed workflows")
	return result.Data, result.NextCursor, nil
}

// CreateWorkflow creates a new workflow
func (c *N8NClient) CreateWorkflow(workflowData map[string]interface{}) (map[string]interface{}, error) {
	workflowJSON, err := json.Marshal(workflowData)
	if err != nil {
		return nil, err
	}

	resp, err := c.DoRequest("POST", "workflows", bytes.NewBuffer(workflowJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API returned non-200/201 status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateWorkflow updates an existing workflow
func (c *N8NClient) UpdateWorkflow(id string, workflowData map[string]interface{}) (map[string]interface{}, error) {
	workflowJSON, err := json.Marshal(workflowData)
	if err != nil {
		return nil, err
	}

	resp, err := c.DoRequest("PUT", fmt.Sprintf("workflows/%s", id), bytes.NewBuffer(workflowJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteWorkflow deletes a workflow by ID
func (c *N8NClient) DeleteWorkflow(id string) error {
	resp, err := c.DoRequest("DELETE", fmt.Sprintf("workflows/%s", id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

// ListExecutions lists workflow executions with optional filters
func (c *N8NClient) ListExecutions(params map[string]string) ([]map[string]interface{}, error) {
	// Build query string
	endpoint := "executions"
	if len(params) > 0 {
		first := true
		for k, v := range params {
			if first {
				endpoint += "?"
				first = false
			} else {
				endpoint += "&"
			}
			endpoint += fmt.Sprintf("%s=%s", k, v)
		}
	}

	resp, err := c.DoRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		// Try parsing as direct array for older n8n versions
		var executions []map[string]interface{}
		if jsonErr := json.Unmarshal(data, &executions); jsonErr != nil {
			return nil, err
		}
		return executions, nil
	}

	return result.Data, nil
}

// GetExecution gets details of a specific execution by ID
func (c *N8NClient) GetExecution(id string) (map[string]interface{}, error) {
	resp, err := c.DoRequest("GET", fmt.Sprintf("executions/%s", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var execution map[string]interface{}
	if err := json.Unmarshal(data, &execution); err != nil {
		return nil, err
	}

	return execution, nil
}

// GetNodes gets available node types in n8n
func (c *N8NClient) GetNodes() ([]map[string]interface{}, error) {
	resp, err := c.DoRequest("GET", "nodes", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		// Try parsing as direct array for older n8n versions
		var nodes []map[string]interface{}
		if jsonErr := json.Unmarshal(data, &nodes); jsonErr != nil {
			return nil, err
		}
		return nodes, nil
	}

	return result.Data, nil
}
