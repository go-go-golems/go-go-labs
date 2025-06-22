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

	"github.com/go-go-golems/go-go-labs/pkg/github"
	"github.com/go-go-golems/go-go-labs/pkg/github/mcp"
)

// Response structures for MCP tools
type AddProjectItemResponse struct {
	Message       string    `json:"message"`
	ProjectItemID string    `json:"project_item_id"`
	IssueID       string    `json:"issue_id,omitempty"`   // Only for new issues
	ContentID     string    `json:"content_id,omitempty"` // Only for existing content
	Task          *mcp.Task `json:"task"`
}

type AddCommentResponse struct {
	Message   string `json:"message"`
	CommentID string `json:"comment_id"`
}

type UpdateCommentResponse struct {
	Message   string `json:"message"`
	CommentID string `json:"comment_id"`
	URL       string `json:"url"`
}

type UpdateProjectItemResponse struct {
	Message       string `json:"message"`
	ProjectItemID string `json:"project_item_id"`
	NewIssueID    string `json:"new_issue_id,omitempty"` // Only when converting DRAFT_ISSUE to ISSUE
}

// ReadProjectItemsHandler handles reading all project items
func ReadProjectItemsHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().
		Str("function", "readProjectItemsHandler").
		Msg("entering readProjectItemsHandler")

	if err := mcp.EnsureService(ctx); err != nil {
		log.Error().Err(err).Msg("failed to ensure GitHub service")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to initialize GitHub service: " + err.Error())), nil
	}

	service := mcp.GetService()
	if service == nil {
		log.Error().Msg("GitHub service not initialized")
		return protocol.NewErrorToolResult(protocol.NewTextContent("GitHub service not initialized")), nil
	}

	tasks, err := service.GetTasks(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to get tasks")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to get tasks: " + err.Error())), nil
	}

	tasksJSON, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to marshal tasks")
		return nil, fmt.Errorf("failed to marshal tasks: %w", err)
	}

	log.Debug().
		Int("taskCount", len(tasks)).
		Dur("duration", time.Since(start)).
		Msg("readProjectItemsHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(string(tasksJSON)),
	), nil
}

// AddProjectItemHandler handles adding a new project item
func AddProjectItemHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().
		Str("function", "addProjectItemHandler").
		Msg("entering addProjectItemHandler")

	if err := mcp.EnsureService(ctx); err != nil {
		log.Error().Err(err).Msg("failed to ensure GitHub service")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to initialize GitHub service: " + err.Error())), nil
	}

	service := mcp.GetService()
	if service == nil {
		log.Error().Msg("GitHub service not initialized")
		return protocol.NewErrorToolResult(protocol.NewTextContent("GitHub service not initialized")), nil
	}

	content := args.GetString("content", "")
	contentID := args.GetString("content_id", "")
	itemType := args.GetString("item_type", "ISSUE")

	// Validate that either content or content_id is provided
	if content == "" && contentID == "" {
		log.Error().Msg("either content or content_id parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Either 'content' (for new issue) or 'content_id' (for existing issue/PR) is required")), nil
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

	var task *mcp.Task
	var err error

	if content != "" {
		// Create real issue in repository
		task, err = service.AddTask(ctx, content, priority, labels)
	} else {
		// Add existing content to project
		task, err = service.AddExistingItemToProject(ctx, contentID, itemType, priority, labels)
	}

	if err != nil {
		log.Error().
			Err(err).
			Str("content", content).
			Str("content_id", contentID).
			Msg("failed to create project item")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to create project item: " + err.Error())), nil
	}

	// Create detailed response
	response := AddProjectItemResponse{
		Message:       "Project item added successfully",
		ProjectItemID: task.ID,
		Task:          task,
	}

	// Add issue/content ID based on what was created
	if content != "" {
		// This was a new issue creation, get the underlying issue ID from the task
		items, err := service.GetClient().GetProjectItems(ctx, service.GetProjectID(), 100)
		if err == nil {
			for _, item := range items {
				if item.ID == task.ID && item.Content.ID != "" {
					response.IssueID = item.Content.ID
					break
				}
			}
		}
	} else {
		// This was adding existing content
		response.ContentID = contentID
	}

	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Error().
			Err(err).
			Str("taskID", task.ID).
			Msg("failed to marshal response")
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	log.Debug().
		Str("taskID", task.ID).
		Str("issueID", response.IssueID).
		Str("contentID", response.ContentID).
		Dur("duration", time.Since(start)).
		Msg("addProjectItemHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(string(responseJSON)),
	), nil
}

// UpdateProjectItemHandler handles updating a project item
func UpdateProjectItemHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().
		Str("function", "updateProjectItemHandler").
		Msg("entering updateProjectItemHandler")

	if err := mcp.EnsureService(ctx); err != nil {
		log.Error().Err(err).Msg("failed to ensure GitHub service")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to initialize GitHub service: " + err.Error())), nil
	}

	service := mcp.GetService()
	if service == nil {
		log.Error().Msg("GitHub service not initialized")
		return protocol.NewErrorToolResult(protocol.NewTextContent("GitHub service not initialized")), nil
	}

	id := args.GetString("id", "")
	if id == "" {
		log.Error().Msg("id parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Project item ID is required")), nil
	}

	status := args.GetString("status", "")
	priority := args.GetString("priority", "")
	itemType := args.GetString("item_type", "")
	labelsStr := args.GetString("labels", "")

	// Validate priority if provided
	if priority != "" && !isValidPriority(priority) {
		log.Error().
			Str("priority", priority).
			Msg("invalid priority provided")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Invalid priority: " + priority)), nil
	}

	// Parse labels from comma-separated string
	var labels []string
	if labelsStr != "" {
		log.Debug().
			Str("labelsStr", labelsStr).
			Msg("parsing labels")
		labels = parseLabels(labelsStr)
	}

	// Build updates map
	updates := make(map[string]interface{})
	if status != "" {
		updates["status"] = status
	}
	if priority != "" {
		updates["priority"] = priority
	}
	if itemType != "" {
		updates["item_type"] = itemType
	}
	if len(labels) > 0 {
		updates["labels"] = labels
	}

	log.Debug().
		Str("id", id).
		Interface("updates", updates).
		Msg("updating project item")

	err := service.UpdateTask(ctx, id, updates)
	if err != nil {
		log.Error().
			Err(err).
			Str("id", id).
			Msg("failed to update project item")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to update project item: " + err.Error())), nil
	}

	// Create detailed response
	response := UpdateProjectItemResponse{
		Message:       "Project item updated successfully",
		ProjectItemID: id,
	}

	// Check if we converted a DRAFT_ISSUE to ISSUE and get the new issue ID
	if itemType == "ISSUE" {
		items, err := service.GetClient().GetProjectItems(ctx, service.GetProjectID(), 100)
		if err == nil {
			for _, item := range items {
				if item.ID == id && item.Content.Typename == "Issue" && item.Content.ID != "" {
					response.NewIssueID = item.Content.ID
					break
				}
			}
		}
	}

	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Error().
			Err(err).
			Str("id", id).
			Msg("failed to marshal response")
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	log.Debug().
		Str("id", id).
		Str("newIssueID", response.NewIssueID).
		Dur("duration", time.Since(start)).
		Msg("updateProjectItemHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(string(responseJSON)),
	), nil
}

// AddProjectItemCommentHandler handles adding comments to project items
func AddProjectItemCommentHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().
		Str("function", "addProjectItemCommentHandler").
		Msg("entering addProjectItemCommentHandler")

	if err := mcp.EnsureService(ctx); err != nil {
		log.Error().Err(err).Msg("failed to ensure GitHub service")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to initialize GitHub service: " + err.Error())), nil
	}

	service := mcp.GetService()
	if service == nil {
		log.Error().Msg("GitHub service not initialized")
		return protocol.NewErrorToolResult(protocol.NewTextContent("GitHub service not initialized")), nil
	}

	id := args.GetString("id", "")
	if id == "" {
		log.Error().Msg("id parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Project item ID is required")), nil
	}

	body := args.GetString("body", "")
	if body == "" {
		log.Error().Msg("body parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Comment body is required")), nil
	}

	log.Debug().
		Str("id", id).
		Str("body", body).
		Msg("adding comment to project item")

	comment, err := service.AddTaskComment(ctx, id, body)
	if err != nil {
		log.Error().
			Err(err).
			Str("id", id).
			Msg("failed to add comment to project item")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to add comment: " + err.Error())), nil
	}

	// Create detailed response
	response := AddCommentResponse{
		Message:   "Comment added successfully",
		CommentID: comment.ID,
	}

	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Error().
			Err(err).
			Str("id", id).
			Str("commentID", comment.ID).
			Msg("failed to marshal response")
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	log.Debug().
		Str("id", id).
		Str("commentID", comment.ID).
		Str("commentURL", comment.URL).
		Dur("duration", time.Since(start)).
		Msg("addProjectItemCommentHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(string(responseJSON)),
	), nil
}

// UpdateIssueCommentHandler handles updating existing comments on issues or pull requests
func UpdateIssueCommentHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().
		Str("function", "updateIssueCommentHandler").
		Msg("entering updateIssueCommentHandler")

	if err := mcp.EnsureService(ctx); err != nil {
		log.Error().Err(err).Msg("failed to ensure GitHub service")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to initialize GitHub service: " + err.Error())), nil
	}

	service := mcp.GetService()
	if service == nil {
		log.Error().Msg("GitHub service not initialized")
		return protocol.NewErrorToolResult(protocol.NewTextContent("GitHub service not initialized")), nil
	}

	commentID := args.GetString("comment_id", "")
	if commentID == "" {
		log.Error().Msg("comment_id parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Comment ID is required")), nil
	}

	body := args.GetString("body", "")
	if body == "" {
		log.Error().Msg("body parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Comment body is required")), nil
	}

	log.Debug().
		Str("comment_id", commentID).
		Int("body_length", len(body)).
		Msg("updating issue comment")

	comment, err := service.GetClient().UpdateIssueComment(ctx, commentID, body)
	if err != nil {
		log.Error().
			Err(err).
			Str("comment_id", commentID).
			Msg("failed to update issue comment")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to update comment: " + err.Error())), nil
	}

	// Create detailed response
	response := UpdateCommentResponse{
		Message:   "Comment updated successfully",
		CommentID: comment.ID,
		URL:       comment.URL,
	}

	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		log.Error().
			Err(err).
			Str("comment_id", commentID).
			Msg("failed to marshal response")
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	log.Debug().
		Str("comment_id", commentID).
		Str("updated_comment_url", comment.URL).
		Dur("duration", time.Since(start)).
		Msg("updateIssueCommentHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(string(responseJSON)),
	), nil
}

// GetProjectInfoHandler handles getting project information
func GetProjectInfoHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().
		Str("function", "getProjectInfoHandler").
		Msg("entering getProjectInfoHandler")

	if err := mcp.EnsureService(ctx); err != nil {
		log.Error().Err(err).Msg("failed to ensure GitHub service")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to initialize GitHub service: " + err.Error())), nil
	}

	service := mcp.GetService()
	if service == nil {
		log.Error().Msg("GitHub service not initialized")
		return protocol.NewErrorToolResult(protocol.NewTextContent("GitHub service not initialized")), nil
	}

	log.Debug().Msg("getting project information")

	project, fields, labels, err := service.GetProjectWithCurrentConfig(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to get project information")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to get project info: " + err.Error())), nil
	}

	// Combine project, fields, and labels information
	info := map[string]interface{}{
		"project": project,
		"fields":  fields,
		"labels":  labels,
	}

	infoJSON, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to marshal project info")
		return nil, fmt.Errorf("failed to marshal project info: %w", err)
	}

	log.Debug().
		Dur("duration", time.Since(start)).
		Msg("getProjectInfoHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(string(infoJSON)),
	), nil
}

// GetIssueCommentsHandler handles getting all comments from an issue or pull request
func GetIssueCommentsHandler(ctx context.Context, args embeddable.Arguments) (*protocol.ToolResult, error) {
	start := time.Now()
	log.Debug().
		Str("function", "getIssueCommentsHandler").
		Msg("entering getIssueCommentsHandler")

	if err := mcp.EnsureService(ctx); err != nil {
		log.Error().Err(err).Msg("failed to ensure GitHub service")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to initialize GitHub service: " + err.Error())), nil
	}

	service := mcp.GetService()
	if service == nil {
		log.Error().Msg("GitHub service not initialized")
		return protocol.NewErrorToolResult(protocol.NewTextContent("GitHub service not initialized")), nil
	}

	issueID := args.GetString("issue_id", "")
	projectItemID := args.GetString("project_item_id", "")

	// Validate that either issue_id or project_item_id is provided
	if issueID == "" && projectItemID == "" {
		log.Error().Msg("either issue_id or project_item_id parameter is required")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Either 'issue_id' (GitHub node ID) or 'project_item_id' (project item ID) is required")), nil
	}

	if issueID != "" && projectItemID != "" {
		log.Error().Msg("cannot specify both issue_id and project_item_id")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Cannot specify both 'issue_id' and 'project_item_id' - use one or the other")), nil
	}

	log.Debug().
		Str("issue_id", issueID).
		Str("project_item_id", projectItemID).
		Msg("getting issue comments")

	var comments []github.Comment
	var err error

	if issueID != "" {
		comments, err = service.GetIssueCommentsByIssueID(ctx, issueID)
	} else {
		comments, err = service.GetIssueCommentsByProjectItemID(ctx, projectItemID)
	}

	if err != nil {
		log.Error().
			Err(err).
			Str("issue_id", issueID).
			Str("project_item_id", projectItemID).
			Msg("failed to get issue comments")
		return protocol.NewErrorToolResult(protocol.NewTextContent("Failed to get issue comments: " + err.Error())), nil
	}

	commentsJSON, err := json.MarshalIndent(comments, "", "  ")
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to marshal comments")
		return nil, fmt.Errorf("failed to marshal comments: %w", err)
	}

	log.Debug().
		Int("comments_count", len(comments)).
		Dur("duration", time.Since(start)).
		Msg("getIssueCommentsHandler completed successfully")

	return protocol.NewToolResult(
		protocol.WithText(string(commentsJSON)),
	), nil
}

// Helper functions
func isValidPriority(priority string) bool {
	return priority == "low" || priority == "medium" || priority == "high"
}

func parseLabels(labelsStr string) []string {
	if labelsStr == "" {
		return []string{}
	}

	var labels []string
	splitLabels := strings.Split(labelsStr, ",")
	for _, label := range splitLabels {
		trimmed := strings.TrimSpace(label)
		if trimmed != "" {
			labels = append(labels, trimmed)
		}
	}
	return labels
}
