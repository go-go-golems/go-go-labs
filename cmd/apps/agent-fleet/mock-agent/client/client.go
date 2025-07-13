package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// Client represents an HTTP client for the Agent Fleet API
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// New creates a new API client
func New(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest performs an HTTP request with authentication
func (c *Client) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal request body")
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + "/v1" + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	log.Debug().Str("method", method).Str("url", url).Msg("Making API request")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}

	return resp, nil
}

// parseResponse parses an HTTP response into a target struct
func (c *Client) parseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorResp models.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return fmt.Errorf("API error (%d): %s - %s", resp.StatusCode, errorResp.Error.Code, errorResp.Error.Message)
		}
		return fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return errors.Wrap(err, "failed to decode response")
		}
	}

	return nil
}

// Agent operations

func (c *Client) CreateAgent(req models.CreateAgentRequest) (*models.Agent, error) {
	resp, err := c.makeRequest("POST", "/agents", req)
	if err != nil {
		return nil, err
	}

	var agent models.Agent
	if err := c.parseResponse(resp, &agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

func (c *Client) GetAgent(agentID string) (*models.Agent, error) {
	path := fmt.Sprintf("/agents/%s", agentID)
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var agent models.Agent
	if err := c.parseResponse(resp, &agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

func (c *Client) UpdateAgent(agentID string, req models.UpdateAgentRequest) (*models.Agent, error) {
	path := fmt.Sprintf("/agents/%s", agentID)
	resp, err := c.makeRequest("PATCH", path, req)
	if err != nil {
		return nil, err
	}

	var agent models.Agent
	if err := c.parseResponse(resp, &agent); err != nil {
		return nil, err
	}

	return &agent, nil
}

func (c *Client) DeleteAgent(agentID string) error {
	path := fmt.Sprintf("/agents/%s", agentID)
	resp, err := c.makeRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	return c.parseResponse(resp, nil)
}

// Event operations

func (c *Client) CreateEvent(agentID string, req models.CreateEventRequest) (*models.Event, error) {
	path := fmt.Sprintf("/agents/%s/events", agentID)
	resp, err := c.makeRequest("POST", path, req)
	if err != nil {
		return nil, err
	}

	var event models.Event
	if err := c.parseResponse(resp, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

func (c *Client) ListEvents(agentID string, eventType string, limit, offset int) ([]models.Event, error) {
	path := fmt.Sprintf("/agents/%s/events", agentID)

	// Add query parameters
	params := url.Values{}
	if eventType != "" {
		params.Add("type", eventType)
	}
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		params.Add("offset", strconv.Itoa(offset))
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response models.EventsListResponse
	if err := c.parseResponse(resp, &response); err != nil {
		return nil, err
	}

	return response.Events, nil
}

// Todo operations

func (c *Client) CreateTodo(agentID string, req models.CreateTodoRequest) (*models.TodoItem, error) {
	path := fmt.Sprintf("/agents/%s/todos", agentID)
	resp, err := c.makeRequest("POST", path, req)
	if err != nil {
		return nil, err
	}

	var todo models.TodoItem
	if err := c.parseResponse(resp, &todo); err != nil {
		return nil, err
	}

	return &todo, nil
}

func (c *Client) ListTodos(agentID string) ([]models.TodoItem, error) {
	path := fmt.Sprintf("/agents/%s/todos", agentID)
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response models.TodosListResponse
	if err := c.parseResponse(resp, &response); err != nil {
		return nil, err
	}

	return response.Todos, nil
}

func (c *Client) UpdateTodo(agentID, todoID string, req models.UpdateTodoRequest) (*models.TodoItem, error) {
	path := fmt.Sprintf("/agents/%s/todos/%s", agentID, todoID)
	resp, err := c.makeRequest("PATCH", path, req)
	if err != nil {
		return nil, err
	}

	var todo models.TodoItem
	if err := c.parseResponse(resp, &todo); err != nil {
		return nil, err
	}

	return &todo, nil
}

func (c *Client) DeleteTodo(agentID, todoID string) error {
	path := fmt.Sprintf("/agents/%s/todos/%s", agentID, todoID)
	resp, err := c.makeRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	return c.parseResponse(resp, nil)
}

// Command operations

func (c *Client) ListCommands(agentID string, status string, limit int) ([]models.Command, error) {
	path := fmt.Sprintf("/agents/%s/commands", agentID)

	// Add query parameters
	params := url.Values{}
	if status != "" {
		params.Add("status", status)
	}
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response models.CommandsListResponse
	if err := c.parseResponse(resp, &response); err != nil {
		return nil, err
	}

	return response.Commands, nil
}

func (c *Client) UpdateCommand(agentID, commandID string, req models.UpdateCommandRequest) (*models.Command, error) {
	path := fmt.Sprintf("/agents/%s/commands/%s", agentID, commandID)
	resp, err := c.makeRequest("PATCH", path, req)
	if err != nil {
		return nil, err
	}

	var command models.Command
	if err := c.parseResponse(resp, &command); err != nil {
		return nil, err
	}

	return &command, nil
}

// Task operations

func (c *Client) ListTasks(status, priority string, limit, offset int) ([]models.Task, error) {
	path := "/tasks"

	// Add query parameters
	params := url.Values{}
	if status != "" {
		params.Add("status", status)
	}
	if priority != "" {
		params.Add("priority", priority)
	}
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		params.Add("offset", strconv.Itoa(offset))
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response models.TasksListResponse
	if err := c.parseResponse(resp, &response); err != nil {
		return nil, err
	}

	return response.Tasks, nil
}

func (c *Client) GetTask(taskID string) (*models.Task, error) {
	path := fmt.Sprintf("/tasks/%s", taskID)
	resp, err := c.makeRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var task models.Task
	if err := c.parseResponse(resp, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

func (c *Client) UpdateTask(taskID string, req models.UpdateTaskRequest) (*models.Task, error) {
	path := fmt.Sprintf("/tasks/%s", taskID)
	resp, err := c.makeRequest("PATCH", path, req)
	if err != nil {
		return nil, err
	}

	var task models.Task
	if err := c.parseResponse(resp, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// Fleet operations

func (c *Client) GetFleetStatus() (*models.FleetStatus, error) {
	resp, err := c.makeRequest("GET", "/fleet/status", nil)
	if err != nil {
		return nil, err
	}

	var status models.FleetStatus
	if err := c.parseResponse(resp, &status); err != nil {
		return nil, err
	}

	return &status, nil
}
