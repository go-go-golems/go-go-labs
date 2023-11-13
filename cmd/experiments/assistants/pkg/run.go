package pkg

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

type ToolCall struct {
	ID       string        `json:"id"`
	Type     string        `json:"type"`
	Function *FunctionCall `json:"function,omitempty"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
