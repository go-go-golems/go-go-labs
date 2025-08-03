package models

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"
	"time"
)

// generateSlug creates a slug from a title and adds a random suffix for uniqueness
func generateSlug(title string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	slug := strings.ToLower(title)
	// Replace non-alphanumeric chars with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")
	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")
	// Limit length
	if len(slug) > 50 {
		slug = slug[:50]
	}

	// Add random suffix for uniqueness
	suffix := generateRandomSuffix()
	return slug + "-" + suffix
}

// generateRandomSuffix generates a 6-character random suffix
func generateRandomSuffix() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 6)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

// Agent represents an AI coding agent
type Agent struct {
	ID              string     `json:"id" db:"id"`
	Name            string     `json:"name" db:"name"`
	Status          string     `json:"status" db:"status"` // active|idle|waiting_feedback|error
	CurrentTask     *string    `json:"current_task" db:"current_task"`
	Worktree        string     `json:"worktree" db:"worktree"`
	FilesChanged    int        `json:"files_changed" db:"files_changed"`
	LinesAdded      int        `json:"lines_added" db:"lines_added"`
	LinesRemoved    int        `json:"lines_removed" db:"lines_removed"`
	LastCommit      *time.Time `json:"last_commit" db:"last_commit"`
	Progress        int        `json:"progress" db:"progress"` // 0-100
	PendingQuestion *string    `json:"pending_question" db:"pending_question"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// Event represents an event in the agent's activity log
type Event struct {
	ID        string                 `json:"id" db:"id"`
	AgentID   string                 `json:"agent_id" db:"agent_id"`
	Type      string                 `json:"type" db:"type"` // start|commit|question|success|error|info|command
	Message   string                 `json:"message" db:"message"`
	Metadata  map[string]interface{} `json:"metadata" db:"metadata"`
	Timestamp time.Time              `json:"timestamp" db:"timestamp"`
}

// TodoItem represents a task item in an agent's todo list
type TodoItem struct {
	ID          string     `json:"id" db:"id"`
	AgentID     string     `json:"agent_id" db:"agent_id"`
	Text        string     `json:"text" db:"text"`
	Completed   bool       `json:"completed" db:"completed"`
	Current     bool       `json:"current" db:"current"`
	Order       int        `json:"order" db:"order_num"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
}

// Task represents a task that can be assigned to agents
type Task struct {
	ID              string     `json:"id" db:"id"`
	Title           string     `json:"title" db:"title"`
	Description     string     `json:"description" db:"description"`
	AssignedAgentID *string    `json:"assigned_agent_id" db:"assigned_agent_id"`
	Status          string     `json:"status" db:"status"`     // pending|assigned|in_progress|completed|failed
	Priority        string     `json:"priority" db:"priority"` // low|medium|high|urgent
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	AssignedAt      *time.Time `json:"assigned_at" db:"assigned_at"`
	CompletedAt     *time.Time `json:"completed_at" db:"completed_at"`
}

// Command represents a command sent to an agent
type Command struct {
	ID          string     `json:"id" db:"id"`
	AgentID     string     `json:"agent_id" db:"agent_id"`
	Content     string     `json:"content" db:"content"`
	Type        string     `json:"type" db:"type"` // instruction|feedback|question
	Response    *string    `json:"response" db:"response"`
	Status      string     `json:"status" db:"status"` // sent|acknowledged|completed
	SentAt      time.Time  `json:"sent_at" db:"sent_at"`
	RespondedAt *time.Time `json:"responded_at" db:"responded_at"`
}

// FleetStatus represents the overall status of the agent fleet
type FleetStatus struct {
	TotalAgents           int `json:"total_agents"`
	ActiveAgents          int `json:"active_agents"`
	PendingTasks          int `json:"pending_tasks"`
	AgentsNeedingFeedback int `json:"agents_needing_feedback"`
	TotalFilesChanged     int `json:"total_files_changed"`
	TotalCommitsToday     int `json:"total_commits_today"`
}

// AgentStatus represents valid agent status values
type AgentStatus string

const (
	AgentStatusActive          AgentStatus = "active"
	AgentStatusIdle            AgentStatus = "idle"
	AgentStatusWaitingFeedback AgentStatus = "waiting_feedback"
	AgentStatusError           AgentStatus = "error"
)

// EventType represents valid event types
type EventType string

const (
	EventTypeStart    EventType = "start"
	EventTypeCommit   EventType = "commit"
	EventTypeQuestion EventType = "question"
	EventTypeSuccess  EventType = "success"
	EventTypeError    EventType = "error"
	EventTypeInfo     EventType = "info"
	EventTypeCommand  EventType = "command"
)

// TaskStatus represents valid task status values
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusAssigned   TaskStatus = "assigned"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

// TaskPriority represents valid task priority values
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// CommandType represents valid command types
type CommandType string

const (
	CommandTypeInstruction CommandType = "instruction"
	CommandTypeFeedback    CommandType = "feedback"
	CommandTypeQuestion    CommandType = "question"
)

// CommandStatus represents valid command status values
type CommandStatus string

const (
	CommandStatusSent         CommandStatus = "sent"
	CommandStatusAcknowledged CommandStatus = "acknowledged"
	CommandStatusCompleted    CommandStatus = "completed"
)

// CreateAgentRequest represents the request body for creating an agent
type CreateAgentRequest struct {
	Name     string `json:"name" validate:"required"`
	Worktree string `json:"worktree" validate:"required"`
}

// UpdateAgentRequest represents the request body for updating an agent
type UpdateAgentRequest struct {
	Name            *string `json:"name,omitempty"`
	Status          *string `json:"status,omitempty"`
	CurrentTask     *string `json:"current_task,omitempty"`
	Worktree        *string `json:"worktree,omitempty"`
	FilesChanged    *int    `json:"files_changed,omitempty"`
	LinesAdded      *int    `json:"lines_added,omitempty"`
	LinesRemoved    *int    `json:"lines_removed,omitempty"`
	Progress        *int    `json:"progress,omitempty"`
	PendingQuestion *string `json:"pending_question,omitempty"`
}

// CreateEventRequest represents the request body for creating an event
type CreateEventRequest struct {
	Type     string                 `json:"type" validate:"required"`
	Message  string                 `json:"message" validate:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CreateTodoRequest represents the request body for creating a todo item
type CreateTodoRequest struct {
	Text  string `json:"text" validate:"required"`
	Order int    `json:"order,omitempty"`
}

// UpdateTodoRequest represents the request body for updating a todo item
type UpdateTodoRequest struct {
	Text      *string `json:"text,omitempty"`
	Completed *bool   `json:"completed,omitempty"`
	Current   *bool   `json:"current,omitempty"`
}

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Title           string  `json:"title" validate:"required"`
	Description     string  `json:"description" validate:"required"`
	Priority        string  `json:"priority" validate:"required"`
	AssignedAgentID *string `json:"assigned_agent_id,omitempty"`
}

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	Title           *string `json:"title,omitempty"`
	Description     *string `json:"description,omitempty"`
	Status          *string `json:"status,omitempty"`
	Priority        *string `json:"priority,omitempty"`
	AssignedAgentID *string `json:"assigned_agent_id,omitempty"`
}

// CreateCommandRequest represents the request body for creating a command
type CreateCommandRequest struct {
	Content string `json:"content" validate:"required"`
	Type    string `json:"type" validate:"required"`
}

// UpdateCommandRequest represents the request body for updating a command
type UpdateCommandRequest struct {
	Response *string `json:"response,omitempty"`
	Status   *string `json:"status,omitempty"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// AgentsListResponse represents the response for listing agents
type AgentsListResponse struct {
	Agents []Agent `json:"agents"`
	ListResponse
}

// EventsListResponse represents the response for listing events
type EventsListResponse struct {
	Events []Event `json:"events"`
	ListResponse
}

// TodosListResponse represents the response for listing todos
type TodosListResponse struct {
	Todos []TodoItem `json:"todos"`
}

// TasksListResponse represents the response for listing tasks
type TasksListResponse struct {
	Tasks []Task `json:"tasks"`
	ListResponse
}

// CommandsListResponse represents the response for listing commands
type CommandsListResponse struct {
	Commands []Command `json:"commands"`
}

// RecentUpdatesResponse represents the response for recent updates
type RecentUpdatesResponse struct {
	Updates []Event `json:"updates"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}
