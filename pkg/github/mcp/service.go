package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/pkg/github"
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
	client        *github.Client
	projectID     string
	owner         string // Store owner for reference
	projectNumber int    // Store project number for reference
	fields        map[string]string            // field name -> field ID mapping
	fieldOptions  map[string]map[string]string // field name -> option name -> option ID mapping
}

// Global GitHub service instance
var githubService *GitHubProjectService

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

func (s *GitHubProjectService) InitProject(ctx context.Context, owner string, projectNumber int) error {
	if s.projectID != "" {
		return nil // already initialized
	}

	// Store config for reference
	s.owner = owner
	s.projectNumber = projectNumber

	project, err := s.client.GetProject(ctx, owner, projectNumber)
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

func (s *GitHubProjectService) GetProjectWithCurrentConfig(ctx context.Context) (*github.Project, []github.ProjectField, error) {
	project, err := s.client.GetProject(ctx, s.owner, s.projectNumber)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get project: %w", err)
	}

	fields, err := s.client.GetProjectFields(ctx, s.projectID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get project fields: %w", err)
	}

	return project, fields, nil
}

func (s *GitHubProjectService) AddTask(ctx context.Context, content, priority string, labels []string) (*Task, error) {
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

func (s *GitHubProjectService) AddTaskComment(ctx context.Context, taskID, comment string) error {
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

	// Check if the item type supports comments
	if targetItem.Content.Typename != "Issue" && targetItem.Content.Typename != "PullRequest" {
		return fmt.Errorf("comments can only be added to issues and pull requests, not %s", targetItem.Content.Typename)
	}

	// Use the GitHub node ID to add the comment
	if targetItem.Content.ID == "" {
		return fmt.Errorf("unable to find GitHub node ID for project item %s", taskID)
	}

	return s.client.AddComment(ctx, targetItem.Content.ID, comment)
}

// Helper methods
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

	// Only issues and pull requests can have labels
	if targetItem.Content.Typename != "Issue" && targetItem.Content.Typename != "PullRequest" {
		return fmt.Errorf("labels can only be set on issues and pull requests, not %s", targetItem.Content.Typename)
	}

	// Use the GitHub node ID to update labels
	if targetItem.Content.ID == "" {
		return fmt.Errorf("unable to find GitHub node ID for project item %s", taskID)
	}

	// Convert label names to IDs
	if len(labels) == 0 {
		// TODO: If we want to remove all labels, we'd need to implement that
		return nil
	}

	// Get repository info from URL to resolve labels
	// For now, just add the labels (this is a simplified approach)
	labelIDs, err := s.getLabelIDs(ctx, targetItem, labels)
	if err != nil {
		return fmt.Errorf("failed to resolve labels: %w", err)
	}

	return s.client.AddLabelsToLabelable(ctx, targetItem.Content.ID, labelIDs)
}

func (s *GitHubProjectService) updateTaskItemType(ctx context.Context, taskID, newItemType string) error {
	if !isValidItemType(newItemType) {
		return fmt.Errorf("invalid item type: %s", newItemType)
	}

	// First, get the project item to understand current state
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

	// Handle conversion from draft issue to issue
	if targetItem.Content.Typename == "DraftIssue" && newItemType == "ISSUE" {
		// For converting draft issue to issue, we need repository ID
		// For now, we'll extract it from the project URL or return an error
		return fmt.Errorf("draft issue to issue conversion requires repository setup - not implemented yet")
	}

	// Other conversions may need different handling
	return fmt.Errorf("conversion from %s to %s not yet supported", targetItem.Content.Typename, newItemType)
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
	log.Debug().Str("itemType", itemType).Msg("validating item type")
	valid := itemType == "DRAFT_ISSUE" || itemType == "ISSUE" || itemType == "PULL_REQUEST"
	log.Debug().Str("itemType", itemType).Bool("valid", valid).Msg("item type validation result")
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

// getLabelIDs resolves label names to IDs for a given repository
func (s *GitHubProjectService) getLabelIDs(ctx context.Context, item *github.ProjectItem, labelNames []string) ([]string, error) {
	// Extract repository owner and name from URL
	// URL format: https://github.com/owner/repo/issues/123
	url := item.Content.URL
	parts := strings.Split(url, "/")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid URL format: %s", url)
	}
	
	owner := parts[3]
	repo := parts[4]
	
	return s.client.GetLabelIDsByNames(ctx, owner, repo, labelNames)
}

// maskToken censors the GitHub token for logging purposes
func maskToken(token string) string {
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:8] + strings.Repeat("*", len(token)-8)
}
