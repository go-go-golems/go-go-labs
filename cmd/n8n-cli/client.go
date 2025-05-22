package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-N8N-API-KEY", c.APIKey)

	return c.Client.Do(req)
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
func (c *N8NClient) ListWorkflows(active bool, limit, offset int) ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("workflows?limit=%d&offset=%d", limit, offset)
	if active {
		endpoint += "&active=true"
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
		var workflows []map[string]interface{}
		if jsonErr := json.Unmarshal(data, &workflows); jsonErr != nil {
			return nil, err
		}
		return workflows, nil
	}

	return result.Data, nil
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