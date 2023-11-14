package assistants

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ListRunStepsRequestParams struct {
	ThreadID string
	RunID    string
	Limit    int    // Optional
	Order    string // Optional, "asc" or "desc"
	After    string // Optional
	Before   string // Optional
}

type GetRunStepRequestParams struct {
	ThreadID string
	RunID    string
	StepID   string
}

type ListRunStepsResponse struct {
	Object  string    `json:"object"`
	Data    []RunStep `json:"data"`
	FirstID string    `json:"first_id"`
	LastID  string    `json:"last_id"`
	HasMore bool      `json:"has_more"`
}

type RunStep struct {
	ID          string       `json:"id"`
	Object      string       `json:"object"`
	CreatedAt   int64        `json:"created_at"`
	RunID       string       `json:"run_id"`
	AssistantID string       `json:"assistant_id"`
	ThreadID    string       `json:"thread_id"`
	Type        string       `json:"type"`
	Status      string       `json:"status"`
	CancelledAt *int64       `json:"cancelled_at,omitempty"`
	CompletedAt *int64       `json:"completed_at,omitempty"`
	ExpiredAt   *int64       `json:"expired_at,omitempty"`
	FailedAt    *int64       `json:"failed_at,omitempty"`
	LastError   *string      `json:"last_error,omitempty"`
	StepDetails *StepDetails `json:"step_details"`
}

type StepDetails struct {
	Type            string           `json:"type"`
	MessageCreation *MessageCreation `json:"message_creation,omitempty"`
	// Other types of step details can be added here
}

type MessageCreation struct {
	MessageID string `json:"message_id"`
}

type RunStepDetails struct {
	Type            string           `json:"type"`
	MessageCreation *MessageCreation `json:"message_creation,omitempty"`
	ToolCalls       *ToolCalls       `json:"tool_calls,omitempty"`
}

type ToolCalls struct {
	Type      string     `json:"type"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

type ToolCall struct {
	ID              string            `json:"id"`
	Type            string            `json:"type"`
	CodeInterpreter *CodeInterpreter  `json:"code_interpreter,omitempty"`
	Retrieval       *Retrieval        `json:"retrieval,omitempty"`
	Function        *FunctionToolCall `json:"function,omitempty"`
}

type CodeInterpreter struct {
	Input   string                  `json:"input"`
	Outputs []CodeInterpreterOutput `json:"outputs"`
}

type CodeInterpreterOutput struct {
	Type  string       `json:"type"`
	Logs  *string      `json:"logs,omitempty"`
	Image *ImageOutput `json:"image,omitempty"`
}

type ImageOutput struct {
	FileID string `json:"file_id"`
}

type Retrieval struct {
	// Empty for now as per the specification
}

type FunctionToolCall struct {
	Name      string  `json:"name"`
	Arguments string  `json:"arguments"`
	Output    *string `json:"output,omitempty"` // Nullable
}

func ListRunSteps(client *http.Client, baseURL string, apiKey string, params ListRunStepsRequestParams) (*ListRunStepsResponse, error) {
	url := fmt.Sprintf("%s/threads/%s/runs/%s/steps", baseURL, params.ThreadID, params.RunID)
	if params.Limit > 0 || params.Order != "" || params.After != "" || params.Before != "" {
		// Add query parameters as needed
		// Example: url += fmt.Sprintf("?limit=%d&order=%s", params.Limit, params.Order)
	}

	req, err := http.NewRequest("GET", url, nil)
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

	var response ListRunStepsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func GetRunStep(client *http.Client, baseURL string, apiKey string, params GetRunStepRequestParams) (*RunStep, error) {
	url := fmt.Sprintf("%s/threads/%s/runs/%s/steps/%s", baseURL, params.ThreadID, params.RunID, params.StepID)

	req, err := http.NewRequest("GET", url, nil)
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

	var response RunStep
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
