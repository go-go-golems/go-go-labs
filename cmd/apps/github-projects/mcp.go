package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/cmd/apps/github-projects/pkg/github"
)

// Task represents a task with all its properties mapped to GitHub Project items
type Task struct {
	ID       string    `json:"id"`        // GitHub Project Item ID
	Content  string    `json:"content"`   // Title of draft issue or issue/PR
	Status   string    `json:"status"`    // Status field from project
	Priority string    `json:"priority"`  // Priority field from project
	Labels   []string  `json:"labels"`    // Labels from issue/PR
	Created  time.Time `json:"created"`   // Computed from project item
	Updated  time.Time `json:"updated"`   // Computed from project item
	ItemType string    `json:"item_type"` // "DRAFT_ISSUE", "ISSUE", "PULL_REQUEST"
}

// GitHubProjectService manages tasks via GitHub Projects API
type GitHubProjectService struct {
	client    *github.Client
	projectID string
	fields    map[string]string // field name -> field ID mapping
}

func NewGitHubProjectService() (*GitHubProjectService, error) {
	client, err := github.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	return &GitHubProjectService{
		client: client,
	}, nil
}

func (s *GitHubProjectService) initProject(ctx context.Context) error {
	if s.projectID != "" {
		return nil // already initialized
	}

	project, err := s.client.GetProject(ctx, githubConfig.Owner, githubConfig.ProjectNumber)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	s.projectID = project.ID

	// Get field mappings
	fields, err := s.client.GetProjectFields(ctx, s.projectID)
	if err != nil {
		return fmt.Errorf("failed to get project fields: %w", err)
	}

	s.fields = make(map[string]string)
	for _, field := range fields {
		s.fields[field.Name] = field.ID
	}

	return nil
}

func (s *GitHubProjectService) GetTasks(ctx context.Context) ([]Task, error) {
	if err := s.initProject(ctx); err != nil {
		return nil, err
	}

	items, err := s.client.GetProjectItems(ctx, s.projectID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get project items: %w", err)
	}

	tasks := make([]Task, 0, len(items))
	for _, item := range items {
		task := s.projectItemToTask(item)
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *GitHubProjectService) AddTask(ctx context.Context, content, priority string, labels []string) (*Task, error) {
	if err := s.initProject(ctx); err != nil {
		return nil, err
	}

	// Create draft issue
	itemID, err := s.client.CreateDraftIssue(ctx, s.projectID, content, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create draft issue: %w", err)
	}

	// Set priority if provided
	if priority != "" && s.fields["Priority"] != "" {
		priorityValue := map[string]interface{}{"singleSelectOptionId": s.getPriorityOptionID(priority)}
		if err := s.client.UpdateFieldValue(ctx, s.projectID, itemID, s.fields["Priority"], priorityValue); err != nil {
			log.Warn().Err(err).Msg("failed to set priority")
		}
	}

	// Fetch the created item to return complete task
	items, err := s.client.GetProjectItems(ctx, s.projectID, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created item: %w", err)
	}

	for _, item := range items {
		if item.ID == itemID {
			task := s.projectItemToTask(item)
			return &task, nil
		}
	}

	return nil, fmt.Errorf("created item not found")
}

func (s *GitHubProjectService) UpdateTask(ctx context.Context, taskID string, updates map[string]interface{}) error {
	if err := s.initProject(ctx); err != nil {
		return err
	}

	for field, value := range updates {
		fieldID, exists := s.fields[field]
		if !exists {
			continue // skip unknown fields
		}

		var fieldValue interface{}
		switch field {
		case "Status":
			fieldValue = map[string]interface{}{"singleSelectOptionId": s.getStatusOptionID(value.(string))}
		case "Priority":
			fieldValue = map[string]interface{}{"singleSelectOptionId": s.getPriorityOptionID(value.(string))}
		default:
			fieldValue = map[string]interface{}{"text": value}
		}

		if err := s.client.UpdateFieldValue(ctx, s.projectID, taskID, fieldID, fieldValue); err != nil {
			return fmt.Errorf("failed to update field %s: %w", field, err)
		}
	}

	return nil
}

func (s *GitHubProjectService) projectItemToTask(item github.ProjectItem) Task {
	task := Task{
		ID:       item.ID,
		Content:  item.Content.Title,
		ItemType: item.Type,
		Created:  time.Now(), // We don't have creation time from API
		Updated:  time.Now(), // We don't have update time from API
	}

	// Extract field values
	for _, fieldValue := range item.FieldValues.Nodes {
		switch fieldValue.Field.Name {
		case "Status":
			if fieldValue.Name != nil {
				task.Status = strings.ToLower(*fieldValue.Name)
			}
		case "Priority":
			if fieldValue.Name != nil {
				task.Priority = strings.ToLower(*fieldValue.Name)
			}
		}
	}

	// Set default values if not found
	if task.Status == "" {
		task.Status = "todo"
	}
	if task.Priority == "" {
		task.Priority = "medium"
	}

	return task
}

func (s *GitHubProjectService) getStatusOptionID(status string) string {
	// This would need to be populated from project field options
	// For now, return a placeholder
	return "status_" + status
}

func (s *GitHubProjectService) getPriorityOptionID(priority string) string {
	// This would need to be populated from project field options
	// For now, return a placeholder
	return "priority_" + priority
}

// Global GitHub service instance
var githubService *GitHubProjectService

// maskToken censors the GitHub token for logging purposes
func maskToken(token string) string {
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:8] + strings.Repeat("*", len(token)-8)
}

// ensureGitHubService ensures GitHub service is initialized
func ensureGitHubService(ctx context.Context) error {
	if githubService != nil {
		return nil
	}

	// Load config first
	if err := EnsureGitHubConfig(); err != nil {
		return err
	}

	// Initialize GitHub service
	var err error
	githubService, err = NewGitHubProjectService()
	if err != nil {
		return fmt.Errorf("failed to initialize GitHub service: %w", err)
	}

	return nil
}

// addMCPCommand adds MCP server capability to the root command
func addMCPCommand(rootCmd *cobra.Command) error {
	log.Info().Msg("adding MCP command to root command")

	return embeddable.AddMCPCommand(rootCmd,
		embeddable.WithName("GitHub Projects Item Management"),
		embeddable.WithVersion("1.0.0"),
		embeddable.WithServerDescription("GitHub Projects v2 item management. Manage project items as tasks through MCP."),

		// Read project items tool
		embeddable.WithEnhancedTool("read_project_items", readProjectItemsHandler,
			embeddable.WithEnhancedDescription("Get all current project items (tasks)"),
			embeddable.WithReadOnlyHint(true),
			embeddable.WithIdempotentHint(true),
		),

		// Add project item tool
		embeddable.WithEnhancedTool("add_project_item", addProjectItemHandler,
			embeddable.WithEnhancedDescription("Add a new draft issue as a project item"),
			embeddable.WithStringProperty("content",
				embeddable.PropertyDescription("Title/description of the draft issue"),
				embeddable.PropertyRequired(),
				embeddable.MinLength(1),
			),
			embeddable.WithStringProperty("priority",
				embeddable.PropertyDescription("Priority level"),
				embeddable.StringEnum("low", "medium", "high"),
				embeddable.DefaultString("medium"),
			),
			embeddable.WithStringProperty("labels",
				embeddable.PropertyDescription("Comma-separated labels for categorization (not yet implemented)"),
			),
		),

		// Update project item tool
		embeddable.WithEnhancedTool("update_project_item", updateProjectItemHandler,
			embeddable.WithEnhancedDescription("Update a project item's field values"),
			embeddable.WithStringProperty("id",
				embeddable.PropertyDescription("Project item ID to update"),
				embeddable.PropertyRequired(),
			),
			embeddable.WithStringProperty("status",
				embeddable.PropertyDescription("New status (field must exist in project)"),
				embeddable.StringEnum("todo", "in-progress", "completed", "done", "backlog"),
			),
			embeddable.WithStringProperty("priority",
				embeddable.PropertyDescription("New priority (field must exist in project)"),
				embeddable.StringEnum("low", "medium", "high"),
			),
		),
	)
}

// Project item management tool handlers
func readProjectItemsHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().
		Str("function", "readProjectItemsHandler").
		Msg("entering readProjectItemsHandler")

	if err := ensureGitHubService(ctx); err != nil {
		log.Error().Err(err).Msg("failed to initialize GitHub service")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Configuration error: " + err.Error())), nil
	}

	log.Debug().
		Str("github_owner", githubConfig.Owner).
		Int("github_project_number", githubConfig.ProjectNumber).
		Str("github_token_masked", maskToken(githubConfig.Token)).
		Msg("using GitHub config")

	tasks, err := githubService.GetTasks(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to get project items")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to get project items: " + err.Error())), nil
	}

	log.Debug().
		Int("taskCount", len(tasks)).
		Msg("marshaling tasks to JSON")

	tasksJSON, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to marshal tasks")
		return nil, fmt.Errorf("failed to marshal tasks: %w", err)
	}

	result := fmt.Sprintf("Current project items (%d total) for GitHub project %s/%d:\n%s",
		len(tasks), githubConfig.Owner, githubConfig.ProjectNumber, string(tasksJSON))

	log.Debug().
		Int("taskCount", len(tasks)).
		Dur("duration", time.Since(start)).
		Msg("readProjectItemsHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(result),
	), nil
}

func addProjectItemHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().
		Str("function", "addProjectItemHandler").
		Msg("entering addProjectItemHandler")

	if err := ensureGitHubService(ctx); err != nil {
		log.Error().Err(err).Msg("failed to initialize GitHub service")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Configuration error: " + err.Error())), nil
	}

	content, err := args.RequireString("content")
	if err != nil {
		log.Error().
			Err(err).
			Msg("content parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("content is required")), nil
	}

	priority := args.GetString("priority", "medium")
	log.Debug().
		Str("priority", priority).
		Msg("validating priority")

	if !isValidPriority(priority) {
		log.Error().
			Str("priority", priority).
			Msg("invalid priority provided")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Invalid priority: " + priority)), nil
	}

	// Parse labels from comma-separated string
	var labels []string
	labelsStr := args.GetString("labels", "")
	if labelsStr != "" {
		log.Debug().
			Str("labelsStr", labelsStr).
			Msg("parsing labels")
		labels = parseLabels(labelsStr)
	}

	log.Debug().
		Str("content", content).
		Str("priority", priority).
		Interface("labels", labels).
		Msg("creating new project item")

	task, err := githubService.AddTask(ctx, content, priority, labels)
	if err != nil {
		log.Error().
			Err(err).
			Str("content", content).
			Msg("failed to create project item")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to create project item: " + err.Error())), nil
	}

	taskJSON, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		log.Error().
			Err(err).
			Str("taskID", task.ID).
			Msg("failed to marshal task")
		return nil, fmt.Errorf("failed to marshal task: %w", err)
	}

	log.Debug().
		Str("taskID", task.ID).
		Dur("duration", time.Since(start)).
		Msg("addProjectItemHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Project item added successfully:\n%s", string(taskJSON))),
	), nil
}

func updateProjectItemHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().
		Str("function", "updateProjectItemHandler").
		Msg("entering updateProjectItemHandler")

	if err := ensureGitHubService(ctx); err != nil {
		log.Error().Err(err).Msg("failed to initialize GitHub service")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Configuration error: " + err.Error())), nil
	}

	taskID, err := args.RequireString("id")
	if err != nil {
		log.Error().
			Err(err).
			Msg("id parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("id is required")), nil
	}

	log.Debug().
		Str("taskID", taskID).
		Msg("processing update project item parameters")

	status := args.GetString("status", "")
	priority := args.GetString("priority", "")

	log.Debug().
		Str("taskID", taskID).
		Str("status", status).
		Str("priority", priority).
		Msg("update parameters parsed")

	// Validate provided values
	if status != "" && !isValidStatus(status) {
		log.Error().
			Str("taskID", taskID).
			Str("status", status).
			Msg("invalid status provided")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Invalid status: " + status)), nil
	}
	if priority != "" && !isValidPriority(priority) {
		log.Error().
			Str("taskID", taskID).
			Str("priority", priority).
			Msg("invalid priority provided")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Invalid priority: " + priority)), nil
	}

	// Build updates map
	updates := make(map[string]interface{})
	if status != "" {
		updates["Status"] = status
	}
	if priority != "" {
		updates["Priority"] = priority
	}

	if len(updates) == 0 {
		return protocol.NewErrorToolResult(protocol.NewTextContent("No valid fields to update")), nil
	}

	log.Debug().
		Str("taskID", taskID).
		Interface("updates", updates).
		Msg("updating project item")

	err = githubService.UpdateTask(ctx, taskID, updates)
	if err != nil {
		log.Error().
			Err(err).
			Str("taskID", taskID).
			Msg("failed to update project item")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to update project item: " + err.Error())), nil
	}

	log.Debug().
		Str("taskID", taskID).
		Dur("duration", time.Since(start)).
		Msg("updateProjectItemHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Project item %s updated successfully", taskID)),
	), nil
}

// Validation helpers
func isValidStatus(status string) bool {
	log.Debug().Str("status", status).Msg("validating status")
	valid := status == "todo" || status == "in-progress" || status == "completed" || status == "done" || status == "backlog"
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
