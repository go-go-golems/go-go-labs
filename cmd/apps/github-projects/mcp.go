package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/session"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// Task represents a task with all its properties
type Task struct {
	ID       string    `json:"id"`
	Content  string    `json:"content"`
	Status   string    `json:"status"`   // "todo", "in-progress", "completed"
	Priority string    `json:"priority"` // "low", "medium", "high"
	Labels   []string  `json:"labels"`   // Optional labels for categorization
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

// TaskStore manages tasks in memory with session isolation
type TaskStore struct {
	mu    sync.RWMutex
	tasks map[string][]Task // key: session ID
}

func NewTaskStore() *TaskStore {
	log.Debug().Msg("creating new task store")
	return &TaskStore{
		tasks: make(map[string][]Task),
	}
}

func (ts *TaskStore) GetTasks(sessionID string) []Task {
	start := time.Now()
	log.Debug().Str("sessionID", sessionID).Msg("getting tasks")

	ts.mu.RLock()
	defer ts.mu.RUnlock()

	tasks, exists := ts.tasks[sessionID]
	if !exists {
		log.Debug().Str("sessionID", sessionID).Msg("no tasks found for session")
		return []Task{}
	}

	log.Debug().
		Str("sessionID", sessionID).
		Int("count", len(tasks)).
		Dur("duration", time.Since(start)).
		Msg("retrieved tasks")

	return tasks
}

func (ts *TaskStore) SetTasks(sessionID string, tasks []Task) {
	start := time.Now()
	log.Debug().
		Str("sessionID", sessionID).
		Int("count", len(tasks)).
		Msg("setting tasks")

	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.tasks[sessionID] = tasks

	log.Debug().
		Str("sessionID", sessionID).
		Dur("duration", time.Since(start)).
		Msg("tasks set successfully")
}

func (ts *TaskStore) AddTask(sessionID string, task Task) {
	start := time.Now()
	log.Debug().
		Str("sessionID", sessionID).
		Str("taskID", task.ID).
		Str("content", task.Content).
		Str("priority", task.Priority).
		Interface("labels", task.Labels).
		Msg("adding task")

	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.tasks[sessionID] == nil {
		log.Debug().Str("sessionID", sessionID).Msg("initializing task list for new session")
		ts.tasks[sessionID] = []Task{}
	}

	ts.tasks[sessionID] = append(ts.tasks[sessionID], task)

	log.Debug().
		Str("sessionID", sessionID).
		Str("taskID", task.ID).
		Int("totalTasks", len(ts.tasks[sessionID])).
		Dur("duration", time.Since(start)).
		Msg("task added successfully")
}

func (ts *TaskStore) UpdateTask(sessionID, taskID string, updater func(*Task)) error {
	start := time.Now()
	log.Debug().
		Str("sessionID", sessionID).
		Str("taskID", taskID).
		Msg("updating task")

	ts.mu.Lock()
	defer ts.mu.Unlock()

	tasks := ts.tasks[sessionID]
	log.Debug().
		Str("sessionID", sessionID).
		Int("totalTasks", len(tasks)).
		Msg("searching for task to update")

	for i := range tasks {
		if tasks[i].ID == taskID {
			oldTask := tasks[i]
			updater(&tasks[i])
			tasks[i].Updated = time.Now()

			log.Debug().
				Str("sessionID", sessionID).
				Str("taskID", taskID).
				Interface("oldTask", oldTask).
				Interface("newTask", tasks[i]).
				Dur("duration", time.Since(start)).
				Msg("task updated successfully")

			return nil
		}
	}

	log.Warn().
		Str("sessionID", sessionID).
		Str("taskID", taskID).
		Dur("duration", time.Since(start)).
		Msg("task not found for update")

	return fmt.Errorf("task with ID %s not found", taskID)
}

func (ts *TaskStore) RemoveTask(sessionID, taskID string) error {
	start := time.Now()
	log.Debug().
		Str("sessionID", sessionID).
		Str("taskID", taskID).
		Msg("removing task")

	ts.mu.Lock()
	defer ts.mu.Unlock()

	tasks := ts.tasks[sessionID]
	log.Debug().
		Str("sessionID", sessionID).
		Int("totalTasks", len(tasks)).
		Msg("searching for task to remove")

	for i, task := range tasks {
		if task.ID == taskID {
			removedTask := task
			ts.tasks[sessionID] = append(tasks[:i], tasks[i+1:]...)

			log.Debug().
				Str("sessionID", sessionID).
				Str("taskID", taskID).
				Interface("removedTask", removedTask).
				Int("remainingTasks", len(ts.tasks[sessionID])).
				Dur("duration", time.Since(start)).
				Msg("task removed successfully")

			return nil
		}
	}

	log.Warn().
		Str("sessionID", sessionID).
		Str("taskID", taskID).
		Dur("duration", time.Since(start)).
		Msg("task not found for removal")

	return fmt.Errorf("task with ID %s not found", taskID)
}

// Global task store instance
var taskStore = NewTaskStore()

// addMCPCommand adds MCP server capability to the root command
func addMCPCommand(rootCmd *cobra.Command) error {
	log.Info().Msg("adding MCP command to root command")

	return embeddable.AddMCPCommand(rootCmd,
		embeddable.WithName("GitHub GraphQL CLI with Task Management"),
		embeddable.WithVersion("1.0.0"),
		embeddable.WithServerDescription(fmt.Sprintf("GitHub GraphQL CLI enhanced with task management capabilities for project coordination. Connected to %s project #%d", githubConfig.Owner, githubConfig.ProjectNumber)),

		// Read tasks tool
		embeddable.WithEnhancedTool("read_tasks", readTasksHandler,
			embeddable.WithEnhancedDescription("Get all current tasks for the agent session"),
			embeddable.WithReadOnlyHint(true),
			embeddable.WithIdempotentHint(true),
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
			embeddable.WithStringProperty("labels",
				embeddable.PropertyDescription("Comma-separated labels for categorization"),
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
			embeddable.WithStringProperty("labels",
				embeddable.PropertyDescription("Comma-separated labels for categorization"),
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
	log.Debug().Msg("getting session ID from context")

	sess, ok := session.GetSessionFromContext(ctx)
	if !ok {
		log.Error().Msg("no session found in context")
		return "", fmt.Errorf("no session found in context")
	}

	sessionID := string(sess.ID)
	log.Debug().Str("sessionID", sessionID).Msg("session ID retrieved successfully")

	return sessionID, nil
}

// Task management tool handlers
func readTasksHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().Msg("entering readTasksHandler")

	sessionID, err := getSessionID(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get session ID in readTasksHandler")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Session error: " + err.Error())), nil
	}

	log.Debug().Str("sessionID", sessionID).Msg("reading tasks for session")
	tasks := taskStore.GetTasks(sessionID)

	log.Debug().
		Str("sessionID", sessionID).
		Int("taskCount", len(tasks)).
		Msg("marshaling tasks to JSON")

	tasksJSON, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		log.Error().
			Err(err).
			Str("sessionID", sessionID).
			Msg("failed to marshal tasks")
		return nil, fmt.Errorf("failed to marshal tasks: %w", err)
	}

	result := fmt.Sprintf("Current tasks (%d total) for GitHub project %s/%d:\n%s",
		len(tasks), githubConfig.Owner, githubConfig.ProjectNumber, string(tasksJSON))

	log.Debug().
		Str("sessionID", sessionID).
		Int("taskCount", len(tasks)).
		Dur("duration", time.Since(start)).
		Msg("readTasksHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(result),
	), nil
}

func addTaskHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().Msg("entering addTaskHandler")

	sessionID, err := getSessionID(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get session ID in addTaskHandler")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Session error: " + err.Error())), nil
	}

	log.Debug().Str("sessionID", sessionID).Msg("processing add task parameters")

	content, err := args.RequireString("content")
	if err != nil {
		log.Error().
			Err(err).
			Str("sessionID", sessionID).
			Msg("content parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("content is required")), nil
	}

	priority := args.GetString("priority", "medium")
	log.Debug().
		Str("sessionID", sessionID).
		Str("priority", priority).
		Msg("validating priority")

	if !isValidPriority(priority) {
		log.Error().
			Str("sessionID", sessionID).
			Str("priority", priority).
			Msg("invalid priority provided")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Invalid priority: " + priority)), nil
	}

	// Parse labels from comma-separated string
	var labels []string
	labelsStr := args.GetString("labels", "")
	if labelsStr != "" {
		log.Debug().
			Str("sessionID", sessionID).
			Str("labelsStr", labelsStr).
			Msg("parsing labels")
		labels = parseLabels(labelsStr)
	}

	// Generate a simple ID based on timestamp
	taskID := fmt.Sprintf("task_%d", time.Now().UnixNano())

	log.Debug().
		Str("sessionID", sessionID).
		Str("taskID", taskID).
		Str("content", content).
		Str("priority", priority).
		Interface("labels", labels).
		Msg("creating new task")

	task := Task{
		ID:       taskID,
		Content:  content,
		Status:   "todo",
		Priority: priority,
		Labels:   labels,
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	taskStore.AddTask(sessionID, task)

	taskJSON, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		log.Error().
			Err(err).
			Str("sessionID", sessionID).
			Str("taskID", taskID).
			Msg("failed to marshal task")
		return nil, fmt.Errorf("failed to marshal task: %w", err)
	}

	log.Debug().
		Str("sessionID", sessionID).
		Str("taskID", taskID).
		Dur("duration", time.Since(start)).
		Msg("addTaskHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Task added successfully:\n%s", string(taskJSON))),
	), nil
}

func updateTaskHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().Msg("entering updateTaskHandler")

	sessionID, err := getSessionID(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get session ID in updateTaskHandler")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Session error: " + err.Error())), nil
	}

	taskID, err := args.RequireString("id")
	if err != nil {
		log.Error().
			Err(err).
			Str("sessionID", sessionID).
			Msg("id parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("id is required")), nil
	}

	log.Debug().
		Str("sessionID", sessionID).
		Str("taskID", taskID).
		Msg("processing update task parameters")

	content := args.GetString("content", "")
	status := args.GetString("status", "")
	priority := args.GetString("priority", "")
	labelsStr := args.GetString("labels", "")

	log.Debug().
		Str("sessionID", sessionID).
		Str("taskID", taskID).
		Str("content", content).
		Str("status", status).
		Str("priority", priority).
		Str("labelsStr", labelsStr).
		Msg("update parameters parsed")

	// Validate provided values
	if status != "" && !isValidStatus(status) {
		log.Error().
			Str("sessionID", sessionID).
			Str("taskID", taskID).
			Str("status", status).
			Msg("invalid status provided")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Invalid status: " + status)), nil
	}
	if priority != "" && !isValidPriority(priority) {
		log.Error().
			Str("sessionID", sessionID).
			Str("taskID", taskID).
			Str("priority", priority).
			Msg("invalid priority provided")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Invalid priority: " + priority)), nil
	}

	log.Debug().
		Str("sessionID", sessionID).
		Str("taskID", taskID).
		Msg("updating task")

	err = taskStore.UpdateTask(sessionID, taskID, func(task *Task) {
		if content != "" {
			log.Debug().
				Str("sessionID", sessionID).
				Str("taskID", taskID).
				Str("newContent", content).
				Msg("updating task content")
			task.Content = content
		}
		if status != "" {
			log.Debug().
				Str("sessionID", sessionID).
				Str("taskID", taskID).
				Str("newStatus", status).
				Msg("updating task status")
			task.Status = status
		}
		if priority != "" {
			log.Debug().
				Str("sessionID", sessionID).
				Str("taskID", taskID).
				Str("newPriority", priority).
				Msg("updating task priority")
			task.Priority = priority
		}
		if labelsStr != "" {
			log.Debug().
				Str("sessionID", sessionID).
				Str("taskID", taskID).
				Str("labelsStr", labelsStr).
				Msg("updating task labels")
			task.Labels = parseLabels(labelsStr)
		}
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("sessionID", sessionID).
			Str("taskID", taskID).
			Msg("failed to update task")
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	log.Debug().
		Str("sessionID", sessionID).
		Str("taskID", taskID).
		Dur("duration", time.Since(start)).
		Msg("updateTaskHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Task %s updated successfully", taskID)),
	), nil
}

func removeTaskHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().Msg("entering removeTaskHandler")

	sessionID, err := getSessionID(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get session ID in removeTaskHandler")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Session error: " + err.Error())), nil
	}

	taskID, err := args.RequireString("id")
	if err != nil {
		log.Error().
			Err(err).
			Str("sessionID", sessionID).
			Msg("id parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("id is required")), nil
	}

	log.Debug().
		Str("sessionID", sessionID).
		Str("taskID", taskID).
		Msg("removing task")

	err = taskStore.RemoveTask(sessionID, taskID)
	if err != nil {
		log.Error().
			Err(err).
			Str("sessionID", sessionID).
			Str("taskID", taskID).
			Msg("failed to remove task")
		return protocol.NewErrorToolResult(protocol.NewTextContent(err.Error())), nil
	}

	log.Debug().
		Str("sessionID", sessionID).
		Str("taskID", taskID).
		Dur("duration", time.Since(start)).
		Msg("removeTaskHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Task %s removed successfully", taskID)),
	), nil
}

// Validation helpers
func isValidStatus(status string) bool {
	log.Debug().Str("status", status).Msg("validating status")
	valid := status == "todo" || status == "in-progress" || status == "completed"
	log.Debug().Str("status", status).Bool("valid", valid).Msg("status validation result")
	return valid
}

func isValidPriority(priority string) bool {
	log.Debug().Str("priority", priority).Msg("validating priority")
	valid := priority == "low" || priority == "medium" || priority == "high"
	log.Debug().Str("priority", priority).Bool("valid", valid).Msg("priority validation result")
	return valid
}

// parseLabels converts a comma-separated string into a slice of trimmed labels
func parseLabels(labelsStr string) []string {
	log.Debug().Str("labelsStr", labelsStr).Msg("parsing labels")

	if labelsStr == "" {
		log.Debug().Msg("empty labels string, returning empty slice")
		return []string{}
	}

	var labels []string
	splitLabels := strings.Split(labelsStr, ",")
	log.Debug().Interface("splitLabels", splitLabels).Msg("split labels by comma")

	for _, label := range splitLabels {
		trimmed := strings.TrimSpace(label)
		if trimmed != "" {
			labels = append(labels, trimmed)
		}
	}

	log.Debug().Interface("parsedLabels", labels).Msg("labels parsed successfully")
	return labels
}
