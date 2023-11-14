package assistants

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type Run struct {
	ID             string          `json:"id"`
	Object         string          `json:"object"`
	CreatedAt      int             `json:"created_at"`
	ThreadID       string          `json:"thread_id"`
	AssistantID    string          `json:"assistant_id"`
	Status         string          `json:"status"`
	RequiredAction *RequiredAction `json:"required_action,omitempty"`
	LastError      *LastError      `json:"last_error,omitempty"`
	ExpiresAt      *int            `json:"expires_at,omitempty"`
	StartedAt      *int            `json:"started_at,omitempty"`
	CancelledAt    *int            `json:"cancelled_at,omitempty"`
	FailedAt       *int            `json:"failed_at,omitempty"`
	CompletedAt    *int            `json:"completed_at,omitempty"`
	Model          string          `json:"model"`
	Instructions   string          `json:"instructions"`
	Tools          []Tool          `json:"tools"`
	FileIDs        []string        `json:"file_ids"`
}

type RequiredAction struct {
	Type              string             `json:"type"`
	SubmitToolOutputs *SubmitToolOutputs `json:"submit_tool_outputs,omitempty"`
}
type SubmitToolOutputs struct {
	ToolCalls []ToolCall `json:"tool_calls"`
}

type CreateThreadAndRunRequest struct {
	AssistantID string        `json:"assistant_id"`
	Thread      ThreadRequest `json:"thread"`
}

type ThreadRequest struct {
	Messages []MessageRequest `json:"messages"`
}

type MessageRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ListRunsResponse struct {
	Object  string `json:"object"`
	Data    []Run  `json:"data"`
	FirstID string `json:"first_id"`
	LastID  string `json:"last_id"`
	HasMore bool   `json:"has_more"`
}

type CreateRunRequest struct {
	AssistantID  string            `json:"assistant_id"`
	Model        string            `json:"model"`
	Instructions string            `json:"instructions"`
	Tools        []Tool            `json:"tools"`
	FileIDs      []string          `json:"file_ids"`
	Metadata     map[string]string `json:"metadata"`
}

type ModifyRunRequest struct {
	Metadata map[string]string `json:"metadata"`
	// Other fields can be added here as per API specification
}

type SubmitToolOutputsRunRequest struct {
	ToolOutputs []ToolOutput `json:"tool_outputs"`
}

type ToolOutput struct {
	ToolCallID string `json:"tool_call_id"`
	Output     string `json:"output"`
}

func setCommonHeaders(req *http.Request, apiKey string) {
	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v1")
}

func buildQueryString(params ...string) string {
	if len(params) == 0 {
		return ""
	}
	return "?" + strings.Join(params, "&")
}

func CreateThreadAndRun(client *http.Client, baseURL, apiKey string, request CreateThreadAndRunRequest) (*Run, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/threads/runs", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var run Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, err
	}

	return &run, nil
}

func ListRuns(client *http.Client, baseURL, apiKey, threadID string, queryParams ...string) (*ListRunsResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/threads/"+threadID+"/runs"+buildQueryString(queryParams...), nil)
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var listResponse ListRunsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResponse); err != nil {
		return nil, err
	}

	return &listResponse, nil
}

func CreateRun(client *http.Client, baseURL, apiKey, threadID string, request CreateRunRequest) (*Run, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/threads/"+threadID+"/runs", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var run Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, err
	}

	return &run, nil
}

func GetRun(client *http.Client, baseURL, apiKey, threadID, runID string) (*Run, error) {
	req, err := http.NewRequest("GET", baseURL+"/threads/"+threadID+"/runs/"+runID, nil)
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var run Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, err
	}

	return &run, nil
}

func ModifyRun(client *http.Client, baseURL, apiKey, threadID, runID string, request ModifyRunRequest) (*Run, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/threads/"+threadID+"/runs/"+runID, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var run Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, err
	}

	return &run, nil
}

func SubmitToolOutputsToRun(client *http.Client, baseURL, apiKey, threadID, runID string, request SubmitToolOutputsRunRequest) (*Run, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/threads/"+threadID+"/runs/"+runID+"/submit_tool_outputs", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var run Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, err
	}

	return &run, nil
}

func CancelRun(client *http.Client, baseURL, apiKey, threadID, runID string) (*Run, error) {
	req, err := http.NewRequest("POST", baseURL+"/threads/"+threadID+"/runs/"+runID+"/cancel", nil)
	if err != nil {
		return nil, err
	}

	setCommonHeaders(req, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	var run Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, err
	}

	return &run, nil
}
