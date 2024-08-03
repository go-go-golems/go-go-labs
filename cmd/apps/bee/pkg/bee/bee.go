package bee

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"net/url"
)

const (
	baseURL = "https://api.bee.computer/v1"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	limiter    *rate.Limiter
}

type ClientOption func(*Client)

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func WithRateLimit(r rate.Limit, b int) ClientOption {
	return func(c *Client) {
		c.limiter = rate.NewLimiter(r, b)
	}
}

func NewClient(apiKey string, options ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{},
		apiKey:     apiKey,
		baseURL:    baseURL,
	}

	for _, option := range options {
		option(c)
	}

	return c
}

// doRequest performs an HTTP request with context and rate limiting
func (c *Client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	if c.limiter != nil {
		err := c.limiter.Wait(ctx)
		if err != nil {
			return nil, fmt.Errorf("rate limit exceeded: %w", err)
		}
	}

	req = req.WithContext(ctx)
	req.Header.Set("x-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	if err := checkResponseForError(resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) getRequest(ctx context.Context, endpoint string, params url.Values, result interface{}) error {
	url := fmt.Sprintf("%s%s?%s", c.baseURL, endpoint, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func (c *Client) postPutRequest(ctx context.Context, method, endpoint string, input, result interface{}) error {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func (c *Client) deleteRequest(ctx context.Context, endpoint string) error {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	return nil
}

func (c *Client) GetConversations(ctx context.Context, userID string, page, limit int) (*ConversationsResponse, error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("limit", fmt.Sprintf("%d", limit))

	var result ConversationsResponse
	err := c.getRequest(ctx, fmt.Sprintf("/%s/conversations", userID), params, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	return &result, nil
}

func (c *Client) GetConversation(ctx context.Context, userID string, conversationID int) (*Conversation, error) {
	var result struct {
		Conversation Conversation `json:"conversation"`
	}
	err := c.getRequest(ctx, fmt.Sprintf("/%s/conversations/%d", userID, conversationID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return &result.Conversation, nil
}

func (c *Client) DeleteConversation(ctx context.Context, userID string, conversationID int) error {
	err := c.deleteRequest(ctx, fmt.Sprintf("/%s/conversations/%d", userID, conversationID))
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	return nil
}

func (c *Client) EndConversation(ctx context.Context, userID string, conversationID int) error {
	err := c.postPutRequest(ctx, "POST", fmt.Sprintf("/%s/conversations/%d/end", userID, conversationID), nil, nil)
	if err != nil {
		return fmt.Errorf("failed to end conversation: %w", err)
	}

	return nil
}

func (c *Client) RetryConversation(ctx context.Context, userID string, conversationID int) error {
	err := c.postPutRequest(ctx, "POST", fmt.Sprintf("/%s/conversations/%d/retry", userID, conversationID), nil, nil)
	if err != nil {
		return fmt.Errorf("failed to retry conversation: %w", err)
	}

	return nil
}

func (c *Client) GetFacts(ctx context.Context, userID string, page, limit int, confirmed *bool) (*FactsResponse, error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("limit", fmt.Sprintf("%d", limit))
	if confirmed != nil {
		params.Set("confirmed", fmt.Sprintf("%t", *confirmed))
	}

	var result FactsResponse
	err := c.getRequest(ctx, fmt.Sprintf("/%s/facts", userID), params, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get facts: %w", err)
	}

	return &result, nil
}

func (c *Client) CreateFact(ctx context.Context, userID string, input FactInput) (*Fact, error) {
	var result Fact
	err := c.postPutRequest(ctx, "POST", fmt.Sprintf("/%s/facts", userID), input, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create fact: %w", err)
	}

	return &result, nil
}

func (c *Client) GetFact(ctx context.Context, userID string, factID int) (*Fact, error) {
	var result Fact
	err := c.getRequest(ctx, fmt.Sprintf("/%s/facts/%d", userID, factID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get fact: %w", err)
	}

	return &result, nil
}

func (c *Client) UpdateFact(ctx context.Context, userID string, factID int, input FactInput) (*Fact, error) {
	var result Fact
	err := c.postPutRequest(ctx, "PUT", fmt.Sprintf("/%s/facts/%d", userID, factID), input, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update fact: %w", err)
	}

	return &result, nil
}

func (c *Client) DeleteFact(ctx context.Context, userID string, factID int) error {
	err := c.deleteRequest(ctx, fmt.Sprintf("/%s/facts/%d", userID, factID))
	if err != nil {
		return fmt.Errorf("failed to delete fact: %w", err)
	}

	return nil
}

func (c *Client) GetTodos(ctx context.Context, userID string, page, limit int) (*TodosResponse, error) {
	params := url.Values{}
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("limit", fmt.Sprintf("%d", limit))

	var result TodosResponse
	err := c.getRequest(ctx, fmt.Sprintf("/%s/todos", userID), params, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}

	return &result, nil
}

func (c *Client) CreateTodo(ctx context.Context, userID string, input TodoInput) (*Todo, error) {
	var result Todo
	err := c.postPutRequest(ctx, "POST", fmt.Sprintf("/%s/todos", userID), input, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	return &result, nil
}

func (c *Client) GetTodo(ctx context.Context, userID string, todoID int) (*Todo, error) {
	var result Todo
	err := c.getRequest(ctx, fmt.Sprintf("/%s/todos/%d", userID, todoID), nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	return &result, nil
}

func (c *Client) UpdateTodo(ctx context.Context, userID string, todoID int, input TodoInput) (*Todo, error) {
	var result Todo
	err := c.postPutRequest(ctx, "PUT", fmt.Sprintf("/%s/todos/%d", userID, todoID), input, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	return &result, nil
}

func (c *Client) DeleteTodo(ctx context.Context, userID string, todoID int) error {
	err := c.deleteRequest(ctx, fmt.Sprintf("/%s/todos/%d", userID, todoID))
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	return nil
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

func checkResponseForError(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}
	return nil
}
