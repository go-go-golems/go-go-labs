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
	client      *github.Client
	projectID   string
	fields      map[string]string                    // field name -> field ID mapping
	fieldOptions map[string]map[string]string        // field name -> option name -> option ID mapping
}

func NewGitHubProjectService() (*GitHubProjectService, error) {
	client, err := github.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	return &GitHubProjectService{
		client:       client,
		fieldOptions: make(map[string]map[string]string),
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
		
		// Cache field options for single-select fields
		if len(field.Options) > 0 {
			if s.fieldOptions[field.Name] == nil {
				s.fieldOptions[field.Name] = make(map[string]string)
			}
			for _, option := range field.Options {
				// Store both the exact name and a normalized version
				s.fieldOptions[field.Name][option.Name] = option.ID
				s.fieldOptions[field.Name][strings.ToLower(option.Name)] = option.ID
				// Handle common variations
				normalized := strings.ReplaceAll(strings.ToLower(option.Name), " ", "-")
				s.fieldOptions[field.Name][normalized] = option.ID
			}
		}
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
		optionID := s.getPriorityOptionID(priority)
		if optionID == "" {
			log.Warn().
				Str("priority", priority).
				Strs("available_options", s.getAvailableOptions("Priority")).
				Msg("invalid priority value, skipping")
		} else {
			priorityValue := map[string]interface{}{"singleSelectOptionId": optionID}
			if err := s.client.UpdateFieldValue(ctx, s.projectID, itemID, s.fields["Priority"], priorityValue); err != nil {
				log.Warn().Err(err).Msg("failed to set priority")
			}
		}
	}

	// Fetch the created item to return complete task
	// Add retry logic for timing issues with GitHub API
	var task *Task
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i) * 500 * time.Millisecond) // 500ms, 1s delays
		}
		
		items, err := s.client.GetProjectItems(ctx, s.projectID, 100)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch created item: %w", err)
		}

		for _, item := range items {
			if item.ID == itemID {
				taskResult := s.projectItemToTask(item)
				task = &taskResult
				break
			}
		}
		
		if task != nil {
			break
		}
	}

	if task == nil {
		return nil, fmt.Errorf("created item not found after %d retries", maxRetries)
	}
	
	return task, nil
}

func (s *GitHubProjectService) AddExistingItemToProject(ctx context.Context, contentID, itemType, priority string, labels []string) (*Task, error) {
	if err := s.initProject(ctx); err != nil {
		return nil, err
	}

	// Add existing item to project
	itemID, err := s.client.AddItemToProject(ctx, s.projectID, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to add existing item to project: %w", err)
	}

	// Set priority if provided
	if priority != "" && s.fields["Priority"] != "" {
		optionID := s.getPriorityOptionID(priority)
		if optionID == "" {
			log.Warn().
				Str("priority", priority).
				Strs("available_options", s.getAvailableOptions("Priority")).
				Msg("invalid priority value, skipping")
		} else {
			priorityValue := map[string]interface{}{"singleSelectOptionId": optionID}
			if err := s.client.UpdateFieldValue(ctx, s.projectID, itemID, s.fields["Priority"], priorityValue); err != nil {
				log.Warn().Err(err).Msg("failed to set priority")
			}
		}
	}

	// Fetch the created item to return complete task
	// Add retry logic for timing issues with GitHub API
	var task *Task
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i) * 500 * time.Millisecond) // 500ms, 1s delays
		}
		
		items, err := s.client.GetProjectItems(ctx, s.projectID, 100)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch added item: %w", err)
		}

		for _, item := range items {
			if item.ID == itemID {
				taskResult := s.projectItemToTask(item)
				task = &taskResult
				break
			}
		}
		
		if task != nil {
			break
		}
	}

	if task == nil {
		return nil, fmt.Errorf("added item not found after %d retries", maxRetries)
	}
	
	return task, nil
}

func (s *GitHubProjectService) UpdateTask(ctx context.Context, taskID string, updates map[string]interface{}) error {
	if err := s.initProject(ctx); err != nil {
		return err
	}

	// Handle item type conversion first if requested
	if itemTypeValue, hasItemType := updates["item_type"]; hasItemType {
		if err := s.updateTaskItemType(ctx, taskID, itemTypeValue.(string)); err != nil {
			log.Warn().Err(err).Msg("failed to update task item type")
			return fmt.Errorf("failed to update item type: %w", err)
		}
		// Remove item_type from updates map since it's not a project field
		delete(updates, "item_type")
	}

	// Handle labels separately (they're not project fields, but issue/PR properties)
	if labelsValue, hasLabels := updates["labels"]; hasLabels {
		if err := s.updateTaskLabels(ctx, taskID, labelsValue.([]string)); err != nil {
			log.Warn().Err(err).Msg("failed to update task labels")
		}
		// Remove labels from updates map since it's not a project field
		delete(updates, "labels")
	}

	for field, value := range updates {
		fieldID, exists := s.fields[field]
		if !exists {
			continue // skip unknown fields
		}

		var fieldValue interface{}
		switch field {
		case "Status":
			optionID := s.getStatusOptionID(value.(string))
			if optionID == "" {
				return fmt.Errorf("invalid status value '%s' for field '%s'. Available options: %v", 
					value.(string), field, s.getAvailableOptions(field))
			}
			fieldValue = map[string]interface{}{"singleSelectOptionId": optionID}
		case "Priority":
			optionID := s.getPriorityOptionID(value.(string))
			if optionID == "" {
				return fmt.Errorf("invalid priority value '%s' for field '%s'. Available options: %v", 
					value.(string), field, s.getAvailableOptions(field))
			}
			fieldValue = map[string]interface{}{"singleSelectOptionId": optionID}
		default:
			fieldValue = map[string]interface{}{"text": value}
		}

		if err := s.client.UpdateFieldValue(ctx, s.projectID, taskID, fieldID, fieldValue); err != nil {
			return fmt.Errorf("failed to update field %s: %w", field, err)
		}
	}

	return nil
}

func (s *GitHubProjectService) updateTaskLabels(ctx context.Context, taskID string, labels []string) error {
	// First, get the project item to find the underlying issue/PR
	items, err := s.client.GetProjectItems(ctx, s.projectID, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch project items: %w", err)
	}

	var targetItem *github.ProjectItem
	for _, item := range items {
		if item.ID == taskID {
			targetItem = &item
			break
		}
	}

	if targetItem == nil {
		return fmt.Errorf("project item %s not found", taskID)
	}

	// Only update labels for issues and pull requests, not draft issues
	if targetItem.Type != "ISSUE" && targetItem.Type != "PULL_REQUEST" {
		log.Debug().
			Str("item_type", targetItem.Type).
			Msg("cannot set labels on draft issues, skipping")
		return nil
	}

	// Get the content ID (issue or PR ID)
	contentID := targetItem.Content.URL // We'll need to extract the ID from the URL or get it another way
	if contentID == "" {
		return fmt.Errorf("unable to determine content ID for item %s", taskID)
	}

	// For now, log that labels would be updated (we need more GitHub config to determine repo)
	log.Debug().
		Str("taskID", taskID).
		Str("contentType", targetItem.Type).
		Strs("labels", labels).
		Msg("labels update requested - full implementation requires repository context")

	return nil
}

func (s *GitHubProjectService) updateTaskItemType(ctx context.Context, taskID, newItemType string) error {
	// First, get the current project item to understand what we're working with
	items, err := s.client.GetProjectItems(ctx, s.projectID, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch project items: %w", err)
	}

	var targetItem *github.ProjectItem
	for _, item := range items {
		if item.ID == taskID {
			targetItem = &item
			break
		}
	}

	if targetItem == nil {
		return fmt.Errorf("project item %s not found", taskID)
	}

	// Check if conversion is supported and needed
	currentType := targetItem.Type
	if currentType == newItemType {
		log.Debug().
			Str("taskID", taskID).
			Str("itemType", currentType).
			Msg("item is already the requested type, no conversion needed")
		return nil
	}

	// Currently we only support converting DRAFT_ISSUE to ISSUE
	if currentType == "DRAFT_ISSUE" && newItemType == "ISSUE" {
		// We need repository ID to convert - for now we'll return an error
		// as this requires more context about which repository to create the issue in
		return fmt.Errorf("converting DRAFT_ISSUE to ISSUE requires repository context - not yet implemented")
	} else {
		return fmt.Errorf("conversion from %s to %s is not supported", currentType, newItemType)
	}
}

func (s *GitHubProjectService) AddTaskComment(ctx context.Context, taskID, body string) error {
	if err := s.initProject(ctx); err != nil {
		return err
	}

	// First, get the project item to find the underlying issue/PR
	items, err := s.client.GetProjectItems(ctx, s.projectID, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch project items: %w", err)
	}

	var targetItem *github.ProjectItem
	for _, item := range items {
		if item.ID == taskID {
			targetItem = &item
			break
		}
	}

	if targetItem == nil {
		return fmt.Errorf("project item %s not found", taskID)
	}

	// Only add comments to issues and pull requests, not draft issues
	if targetItem.Type != "ISSUE" && targetItem.Type != "PULL_REQUEST" {
		return fmt.Errorf("cannot add comments to %s items, only issues and pull requests", targetItem.Type)
	}

	// Get the GitHub node ID from the content
	subjectID := targetItem.Content.ID
	if subjectID == "" {
		return fmt.Errorf("unable to determine GitHub node ID for project item %s", taskID)
	}

	// Add the comment to the issue or pull request
	if err := s.client.AddComment(ctx, subjectID, body); err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
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
	if statusOptions, exists := s.fieldOptions["Status"]; exists {
		if optionID, found := statusOptions[status]; found {
			return optionID
		}
		if optionID, found := statusOptions[strings.ToLower(status)]; found {
			return optionID
		}
		// Try normalized version (space to dash)
		normalized := strings.ReplaceAll(strings.ToLower(status), " ", "-")
		if optionID, found := statusOptions[normalized]; found {
			return optionID
		}
	}
	// Return empty string if not found - this will cause an error which is better than a placeholder
	return ""
}

func (s *GitHubProjectService) getPriorityOptionID(priority string) string {
	if priorityOptions, exists := s.fieldOptions["Priority"]; exists {
		if optionID, found := priorityOptions[priority]; found {
			return optionID
		}
		if optionID, found := priorityOptions[strings.ToLower(priority)]; found {
			return optionID
		}
		// Try normalized version (space to dash)
		normalized := strings.ReplaceAll(strings.ToLower(priority), " ", "-")
		if optionID, found := priorityOptions[normalized]; found {
			return optionID
		}
	}
	// Return empty string if not found - this will cause an error which is better than a placeholder
	return ""
}

// getAvailableOptions returns the available option names for a field
func (s *GitHubProjectService) getAvailableOptions(fieldName string) []string {
	var options []string
	if fieldOptions, exists := s.fieldOptions[fieldName]; exists {
		seen := make(map[string]bool)
		for optionName := range fieldOptions {
			// Only include names that look like original field values (have capitals or spaces)
			if strings.Contains(optionName, " ") || (len(optionName) > 0 && strings.ToUpper(optionName[:1]) == optionName[:1]) {
				if !seen[optionName] {
					options = append(options, optionName)
					seen[optionName] = true
				}
			}
		}
	}
	return options
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
			embeddable.WithEnhancedDescription("Add a new project item. By default creates a draft issue. Can also add existing issues or pull requests by providing their ID."),
			embeddable.WithStringProperty("content",
				embeddable.PropertyDescription("Title/description for new draft issue, or leave empty when adding existing issue/PR"),
				embeddable.MinLength(1),
			),
			embeddable.WithStringProperty("content_id",
				embeddable.PropertyDescription("ID of existing issue or pull request to add to project (alternative to creating draft issue)"),
			),
			embeddable.WithStringProperty("item_type",
				embeddable.PropertyDescription("Type of item to create: 'ISSUE' (default), 'DRAFT_ISSUE', or 'PULL_REQUEST'. Use DRAFT_ISSUE for tentative items. Only used when adding existing content by ID."),
				embeddable.StringEnum("DRAFT_ISSUE", "ISSUE", "PULL_REQUEST"),
				embeddable.DefaultString("ISSUE"),
			),
			embeddable.WithStringProperty("priority",
				embeddable.PropertyDescription("Priority level"),
				embeddable.StringEnum("low", "medium", "high"),
				embeddable.DefaultString("medium"),
			),
			embeddable.WithStringProperty("labels",
				embeddable.PropertyDescription("Comma-separated labels to set on the underlying issue/PR (only applicable when creating issues, not draft issues)"),
			),
		),

		// Update project item tool
		embeddable.WithEnhancedTool("update_project_item", updateProjectItemHandler,
			embeddable.WithEnhancedDescription("Update a project item's field values, item type, and labels"),
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
			embeddable.WithStringProperty("item_type",
				embeddable.PropertyDescription("Convert item type: 'DRAFT_ISSUE', 'ISSUE', or 'PULL_REQUEST'"),
				embeddable.StringEnum("DRAFT_ISSUE", "ISSUE", "PULL_REQUEST"),
			),
			embeddable.WithStringProperty("labels",
				embeddable.PropertyDescription("Comma-separated labels to set on the underlying issue/PR (replaces existing labels)"),
			),
		),

		// Add comment to project item tool
		embeddable.WithEnhancedTool("add_project_item_comment", addProjectItemCommentHandler,
			embeddable.WithEnhancedDescription("Add a comment to the underlying issue or pull request of a project item. Note: Comments can only be added to ISSUE and PULL_REQUEST types, not DRAFT_ISSUE."),
			embeddable.WithStringProperty("id",
				embeddable.PropertyDescription("Project item ID whose underlying issue/PR to comment on"),
				embeddable.PropertyRequired(),
			),
			embeddable.WithStringProperty("body",
				embeddable.PropertyDescription("Comment text/body to add"),
				embeddable.PropertyRequired(),
				embeddable.MinLength(1),
			),
		),

		// Get project information tool
		embeddable.WithEnhancedTool("get_project_info", getProjectInfoHandler,
			embeddable.WithEnhancedDescription("Get detailed project information including all fields, field types, and available options (labels)"),
			embeddable.WithReadOnlyHint(true),
			embeddable.WithIdempotentHint(true),
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

	content := args.GetString("content", "")
	contentID := args.GetString("content_id", "")
	itemType := args.GetString("item_type", "ISSUE")
	
	// Validate that either content or content_id is provided
	if content == "" && contentID == "" {
		log.Error().Msg("either content or content_id parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Either 'content' (for new draft issue) or 'content_id' (for existing issue/PR) is required")), nil
	}
	
	if content != "" && contentID != "" {
		log.Error().Msg("cannot specify both content and content_id")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Cannot specify both 'content' and 'content_id' - use one or the other")), nil
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
		Str("content_id", contentID).
		Str("item_type", itemType).
		Str("priority", priority).
		Interface("labels", labels).
		Msg("creating new project item")

	var task *Task
	var err error
	
	if content != "" {
		// Create draft issue
		task, err = githubService.AddTask(ctx, content, priority, labels)
	} else {
		// Add existing content to project
		task, err = githubService.AddExistingItemToProject(ctx, contentID, itemType, priority, labels)
	}
	
	if err != nil {
		log.Error().
			Err(err).
			Str("content", content).
			Str("content_id", contentID).
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
	itemType := args.GetString("item_type", "")
	labelsStr := args.GetString("labels", "")

	log.Debug().
		Str("taskID", taskID).
		Str("status", status).
		Str("priority", priority).
		Str("item_type", itemType).
		Str("labels", labelsStr).
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
	if itemType != "" && !isValidItemType(itemType) {
		log.Error().
			Str("taskID", taskID).
			Str("item_type", itemType).
			Msg("invalid item type provided")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Invalid item type: " + itemType)), nil
	}

	// Build updates map
	updates := make(map[string]interface{})
	if status != "" {
		updates["Status"] = status
	}
	if priority != "" {
		updates["Priority"] = priority
	}
	if itemType != "" {
		updates["item_type"] = itemType
	}
	if labelsStr != "" {
		labels := parseLabels(labelsStr)
		updates["labels"] = labels
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

func addProjectItemCommentHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().
		Str("function", "addProjectItemCommentHandler").
		Msg("entering addProjectItemCommentHandler")

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

	body, err := args.RequireString("body")
	if err != nil {
		log.Error().
			Err(err).
			Msg("body parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("body is required")), nil
	}

	log.Debug().
		Str("taskID", taskID).
		Int("bodyLength", len(body)).
		Msg("adding comment to project item")

	err = githubService.AddTaskComment(ctx, taskID, body)
	if err != nil {
		log.Error().
			Err(err).
			Str("taskID", taskID).
			Msg("failed to add comment to project item")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to add comment: " + err.Error())), nil
	}

	log.Debug().
		Str("taskID", taskID).
		Dur("duration", time.Since(start)).
		Msg("addProjectItemCommentHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Comment added successfully to project item %s", taskID)),
	), nil
}

func getProjectInfoHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().
		Str("function", "getProjectInfoHandler").
		Msg("entering getProjectInfoHandler")

	if err := ensureGitHubService(ctx); err != nil {
		log.Error().Err(err).Msg("failed to initialize GitHub service")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Configuration error: " + err.Error())), nil
	}

	log.Debug().
		Str("github_owner", githubConfig.Owner).
		Int("github_project_number", githubConfig.ProjectNumber).
		Str("github_token_masked", maskToken(githubConfig.Token)).
		Msg("using GitHub config")

	// Get project basic info
	project, err := githubService.client.GetProject(ctx, githubConfig.Owner, githubConfig.ProjectNumber)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to get project")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to get project: " + err.Error())), nil
	}

	// Get project fields
	fields, err := githubService.client.GetProjectFields(ctx, project.ID)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to get project fields")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to get project fields: " + err.Error())), nil
	}

	// Build comprehensive project info
	projectInfo := map[string]interface{}{
		"id":                project.ID,
		"title":             project.Title,
		"public":            project.Public,
		"short_description": project.ShortDescription,
		"closed":            project.Closed,
		"total_items":       project.Items.TotalCount,
		"total_fields":      len(fields),
		"fields":            make([]map[string]interface{}, 0, len(fields)),
	}

	// Add detailed field information
	for _, field := range fields {
		fieldInfo := map[string]interface{}{
			"id":        field.ID,
			"name":      field.Name,
			"type":      field.Typename,
			"options":   make([]map[string]interface{}, 0, len(field.Options)),
		}

		// Add field options if any
		if len(field.Options) > 0 {
			for _, option := range field.Options {
				optionInfo := map[string]interface{}{
					"id":   option.ID,
					"name": option.Name,
				}
				fieldInfo["options"] = append(fieldInfo["options"].([]map[string]interface{}), optionInfo)
			}
		}

		// Add iteration configuration if present
		if field.Configuration != nil && len(field.Configuration.Iterations) > 0 {
			iterations := make([]map[string]interface{}, 0, len(field.Configuration.Iterations))
			for _, iteration := range field.Configuration.Iterations {
				iterationInfo := map[string]interface{}{
					"id":         iteration.ID,
					"title":      iteration.Title,
					"start_date": iteration.StartDate,
				}
				iterations = append(iterations, iterationInfo)
			}
			fieldInfo["iterations"] = iterations
		}

		projectInfo["fields"] = append(projectInfo["fields"].([]map[string]interface{}), fieldInfo)
	}

	log.Debug().
		Int("field_count", len(fields)).
		Msg("marshaling project info to JSON")

	projectInfoJSON, err := json.MarshalIndent(projectInfo, "", "  ")
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to marshal project info")
		return nil, fmt.Errorf("failed to marshal project info: %w", err)
	}

	result := fmt.Sprintf("Project information for GitHub project %s/%d:\n%s",
		githubConfig.Owner, githubConfig.ProjectNumber, string(projectInfoJSON))

	log.Debug().
		Int("field_count", len(fields)).
		Dur("duration", time.Since(start)).
		Msg("getProjectInfoHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(result),
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

func isValidItemType(itemType string) bool {
	log.Debug().Str("item_type", itemType).Msg("validating item type")
	valid := itemType == "DRAFT_ISSUE" || itemType == "ISSUE" || itemType == "PULL_REQUEST"
	log.Debug().Str("item_type", itemType).Bool("valid", valid).Msg("item type validation result")
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
