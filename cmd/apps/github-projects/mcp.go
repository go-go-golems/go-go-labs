package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/session"
	"github.com/spf13/cobra"
)

// Task represents a task with all its properties
type Task struct {
	ID       string    `json:"id"`
	Content  string    `json:"content"`
	Status   string    `json:"status"`   // "todo", "in-progress", "completed"
	Priority string    `json:"priority"` // "low", "medium", "high"
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

// TaskStore manages tasks in memory with session isolation
type TaskStore struct {
	mu    sync.RWMutex
	tasks map[string][]Task // key: session ID
}

func NewTaskStore() *TaskStore {
	return &TaskStore{
		tasks: make(map[string][]Task),
	}
}

func (ts *TaskStore) GetTasks(sessionID string) []Task {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	tasks, exists := ts.tasks[sessionID]
	if !exists {
		return []Task{}
	}
	return tasks
}

func (ts *TaskStore) SetTasks(sessionID string, tasks []Task) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.tasks[sessionID] = tasks
}

func (ts *TaskStore) AddTask(sessionID string, task Task) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.tasks[sessionID] == nil {
		ts.tasks[sessionID] = []Task{}
	}
	ts.tasks[sessionID] = append(ts.tasks[sessionID], task)
}

func (ts *TaskStore) UpdateTask(sessionID, taskID string, updater func(*Task)) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	tasks := ts.tasks[sessionID]
	for i := range tasks {
		if tasks[i].ID == taskID {
			updater(&tasks[i])
			tasks[i].Updated = time.Now()
			return nil
		}
	}
	return fmt.Errorf("task with ID %s not found", taskID)
}

func (ts *TaskStore) RemoveTask(sessionID, taskID string) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	tasks := ts.tasks[sessionID]
	for i, task := range tasks {
		if task.ID == taskID {
			ts.tasks[sessionID] = append(tasks[:i], tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task with ID %s not found", taskID)
}

// Global task store instance
var taskStore = NewTaskStore()

// addMCPCommand adds MCP server capability to the root command
func addMCPCommand(rootCmd *cobra.Command) error {
	return embeddable.AddMCPCommand(rootCmd,
		embeddable.WithName("GitHub GraphQL CLI with Task Management"),
		embeddable.WithVersion("1.0.0"),
		embeddable.WithServerDescription("GitHub GraphQL CLI enhanced with task management capabilities for project coordination"),
		
		// Read tasks tool
		embeddable.WithEnhancedTool("read_tasks", readTasksHandler,
			embeddable.WithEnhancedDescription("Get all current tasks for the agent session"),
			embeddable.WithReadOnlyHint(true),
			embeddable.WithIdempotentHint(true),
		),
		
		// Write tasks tool (replace all)
		embeddable.WithEnhancedTool("write_tasks", writeTasksHandler,
			embeddable.WithEnhancedDescription("Replace all tasks for the agent session with provided tasks"),
			embeddable.WithDestructiveHint(true),
			embeddable.WithStringProperty("tasks_json",
				embeddable.PropertyDescription("JSON array of tasks to set"),
				embeddable.PropertyRequired(),
				embeddable.MinLength(1),
			),
		),
		
		// Add task tool
		embeddable.WithEnhancedTool("add_task", addTaskHandler,
			embeddable.WithEnhancedDescription("Add a single new task to track work items"),
			embeddable.WithStringProperty("content",
				embeddable.PropertyDescription("Description of the task"),
				embeddable.PropertyRequired(),
				embeddable.MinLength(1),
			),
			embeddable.WithStringProperty("priority",
				embeddable.PropertyDescription("Priority level"),
				embeddable.StringEnum("low", "medium", "high"),
				embeddable.DefaultString("medium"),
			),
		),
		
		// Update task tool
		embeddable.WithEnhancedTool("update_task", updateTaskHandler,
			embeddable.WithEnhancedDescription("Update a specific task's status, priority, or content"),
			embeddable.WithStringProperty("id",
				embeddable.PropertyDescription("Task ID to update"),
				embeddable.PropertyRequired(),
			),
			embeddable.WithStringProperty("content",
				embeddable.PropertyDescription("New task content"),
			),
			embeddable.WithStringProperty("status",
				embeddable.PropertyDescription("New task status"),
				embeddable.StringEnum("todo", "in-progress", "completed"),
			),
			embeddable.WithStringProperty("priority",
				embeddable.PropertyDescription("New task priority"),
				embeddable.StringEnum("low", "medium", "high"),
			),
		),
		
		// Remove task tool
		embeddable.WithEnhancedTool("remove_task", removeTaskHandler,
			embeddable.WithEnhancedDescription("Remove a specific task by ID"),
			embeddable.WithDestructiveHint(true),
			embeddable.WithStringProperty("id",
				embeddable.PropertyDescription("Task ID to remove"),
				embeddable.PropertyRequired(),
			),
		),
	)
}

// Helper function to get session ID from context
func getSessionID(ctx context.Context) (string, error) {
	sess, ok := session.GetSessionFromContext(ctx)
	if !ok {
		return "", fmt.Errorf("no session found in context")
	}
	return string(sess.ID), nil
}

// Task management tool handlers
func readTasksHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	sessionID, err := getSessionID(ctx)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent("Session error: " + err.Error())), nil
	}

	tasks := taskStore.GetTasks(sessionID)
	
	tasksJSON, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tasks: %w", err)
	}

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Current tasks (%d total):\n%s", len(tasks), string(tasksJSON))),
	), nil
}

func writeTasksHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	sessionID, err := getSessionID(ctx)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent("Session error: " + err.Error())), nil
	}

	tasksJSON, err := args.RequireString("tasks_json")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent("tasks_json is required")), nil
	}

	var tasks []Task
	if err := json.Unmarshal([]byte(tasksJSON), &tasks); err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent("Invalid JSON format: " + err.Error())), nil
	}

	// Validate tasks
	for i, task := range tasks {
		if task.ID == "" {
			return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Task %d missing ID", i))), nil
		}
		if task.Content == "" {
			return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Task %s missing content", task.ID))), nil
		}
		if !isValidStatus(task.Status) {
			return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Task %s has invalid status: %s", task.ID, task.Status))), nil
		}
		if !isValidPriority(task.Priority) {
			return protocol.NewErrorToolResult(protocol.NewTextContent(fmt.Sprintf("Task %s has invalid priority: %s", task.ID, task.Priority))), nil
		}
		// Set timestamps if not provided
		if tasks[i].Created.IsZero() {
			tasks[i].Created = time.Now()
		}
		if tasks[i].Updated.IsZero() {
			tasks[i].Updated = time.Now()
		}
	}

	taskStore.SetTasks(sessionID, tasks)

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Successfully replaced all tasks. New task count: %d", len(tasks))),
	), nil
}

func addTaskHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	sessionID, err := getSessionID(ctx)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent("Session error: " + err.Error())), nil
	}

	content, err := args.RequireString("content")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent("content is required")), nil
	}

	priority := args.GetString("priority", "medium")
	if !isValidPriority(priority) {
		return protocol.NewErrorToolResult(protocol.NewTextContent("Invalid priority: " + priority)), nil
	}

	// Generate a simple ID based on timestamp
	taskID := fmt.Sprintf("task_%d", time.Now().UnixNano())
	
	task := Task{
		ID:       taskID,
		Content:  content,
		Status:   "todo",
		Priority: priority,
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	taskStore.AddTask(sessionID, task)

	taskJSON, _ := json.MarshalIndent(task, "", "  ")
	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Task added successfully:\n%s", string(taskJSON))),
	), nil
}

func updateTaskHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	sessionID, err := getSessionID(ctx)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent("Session error: " + err.Error())), nil
	}

	taskID, err := args.RequireString("id")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent("id is required")), nil
	}

	content := args.GetString("content", "")
	status := args.GetString("status", "")
	priority := args.GetString("priority", "")

	// Validate provided values
	if status != "" && !isValidStatus(status) {
		return protocol.NewErrorToolResult(protocol.NewTextContent("Invalid status: " + status)), nil
	}
	if priority != "" && !isValidPriority(priority) {
		return protocol.NewErrorToolResult(protocol.NewTextContent("Invalid priority: " + priority)), nil
	}

	err = taskStore.UpdateTask(sessionID, taskID, func(task *Task) {
		if content != "" {
			task.Content = content
		}
		if status != "" {
			task.Status = status
		}
		if priority != "" {
			task.Priority = priority
		}
	})

	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Task %s updated successfully", taskID)),
	), nil
}

func removeTaskHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	sessionID, err := getSessionID(ctx)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent("Session error: " + err.Error())), nil
	}

	taskID, err := args.RequireString("id")
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent("id is required")), nil
	}

	err = taskStore.RemoveTask(sessionID, taskID)
	if err != nil {
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Task %s removed successfully", taskID)),
	), nil
}

// Validation helpers
func isValidStatus(status string) bool {
	return status == "todo" || status == "in-progress" || status == "completed"
}

func isValidPriority(priority string) bool {
	return priority == "low" || priority == "medium" || priority == "high"
}
