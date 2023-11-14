package assistants

type Assistant struct {
	ID           string                 `json:"id,omitempty"`
	Object       string                 `json:"object,omitempty"`
	CreatedAt    int64                  `json:"created_at,omitempty"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Model        string                 `json:"model"`
	Instructions string                 `json:"instructions"`
	Tools        []Tool                 `json:"tools,omitempty"`
	FileIDs      []string               `json:"file_ids,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type Tool struct {
	// Type is one of "code_interpreter", "retrieval", "function"
	Type     string    `json:"type"`
	Function *Function `json:"function"`
}

type Function struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	// Parameters are a json schema, no parameters : {"type": "object", "properties": {}}
	Parameters map[string]interface{} `json:"parameters"`
}

type File struct {
	ID        string `json:"id"`
	Bytes     int    `json:"bytes"`
	CreatedAt int64  `json:"created_at"` // Unix timestamp in seconds
	Filename  string `json:"filename"`
	Object    string `json:"object"`
	// Supported values are "fine-tune", "fine-tune-results", "assistants", "assistants-output"
	Purpose string `json:"purpose"`
}

type Message struct {
	ID          string            `json:"id,omitempty"`
	Object      string            `json:"object,omitempty"`
	CreatedAt   int64             `json:"created_at"` // Unix timestamp in seconds
	ThreadID    string            `json:"thread_id,omitempty"`
	Role        string            `json:"role"`
	Content     []ContentObject   `json:"content"`
	AssistantID string            `json:"assistant_id,omitempty"`
	RunID       string            `json:"run_id,omitempty"`
	FileIDs     []string          `json:"file_ids,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type ContentObject struct {
	// type can be "text" or "image_file"
	Type      string            `json:"type"`
	Text      *TextContent      `json:"text,omitempty"`
	ImageFile *ImageFileContent `json:"image_file,omitempty"`
}

type TextContent struct {
	Value       string       `json:"value"`
	Annotations []Annotation `json:"annotations,omitempty"`
}

type ImageFileContent struct {
	FileID string `json:"file_id"`
}

type Annotation struct {
	// type can be "file_citation" or "file_path"
	Type         string        `json:"type"`
	Text         string        `json:"text"`
	FileCitation *FileCitation `json:"file_citation,omitempty"`
	FilePath     *FilePath     `json:"file_path,omitempty"`
}

type FileCitation struct {
	FileID     string `json:"file_id"`
	Quote      string `json:"quote"`
	StartIndex int    `json:"start_index"`
	EndIndex   int    `json:"end_index"`
}

type FilePath struct {
	FileID     string `json:"file_id"`
	StartIndex int    `json:"start_index"`
	EndIndex   int    `json:"end_index"`
}

type Thread struct {
	ID        string            `json:"id"`
	Object    string            `json:"object"`
	CreatedAt int64             `json:"created_at"` // Unix timestamp in seconds
	Metadata  map[string]string `json:"metadata"`   // Map of up to 16 key-value pairs
}
