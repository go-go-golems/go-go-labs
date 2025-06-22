package mcp

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// ensureGitHubService ensures GitHub service is initialized
func ensureGitHubService(ctx context.Context, owner string, projectNumber int, repository string) error {
	if githubService != nil {
		return nil
	}

	// Initialize GitHub service
	var err error
	githubService, err = NewGitHubProjectService()
	if err != nil {
		return fmt.Errorf("failed to initialize GitHub service: %w", err)
	}

	// Initialize project
	if err := githubService.InitProject(ctx, owner, projectNumber, repository); err != nil {
		return fmt.Errorf("failed to initialize project: %w", err)
	}

	return nil
}

// ToolHandler is a function type for MCP tool handlers
type ToolHandler = embeddable.EnhancedToolHandler

// ToolHandlers holds all the tool handlers for the MCP server
type ToolHandlers struct {
	ReadProjectItems      ToolHandler
	AddProjectItem        ToolHandler
	UpdateProjectItem     ToolHandler
	AddProjectItemComment ToolHandler
	GetProjectInfo        ToolHandler
}

// AddMCPCommand adds MCP server capability to the root command
func AddMCPCommand(rootCmd *cobra.Command, handlers *ToolHandlers) error {
	log.Info().Msg("adding MCP command to root command")

	return embeddable.AddMCPCommand(rootCmd,
		embeddable.WithName("GitHub Projects Item Management"),
		embeddable.WithVersion("1.0.0"),
		embeddable.WithServerDescription("GitHub Projects v2 item management. Manage project items as tasks through MCP."),

		// Read project items tool
		embeddable.WithEnhancedTool("read_project_items", handlers.ReadProjectItems,
			embeddable.WithEnhancedDescription("Get all current project items (tasks)"),
			embeddable.WithReadOnlyHint(true),
			embeddable.WithIdempotentHint(true),
		),

		// Add project item tool
		embeddable.WithEnhancedTool("add_project_item", handlers.AddProjectItem,
			embeddable.WithEnhancedDescription("Add a new project item. By default creates a real issue in the configured repository. Can also add existing issues or pull requests by providing their ID."),
			embeddable.WithStringProperty("content",
				embeddable.PropertyDescription("Title/description for new issue, or leave empty when adding existing issue/PR"),
				embeddable.MinLength(1),
			),
			embeddable.WithStringProperty("content_id",
				embeddable.PropertyDescription("ID of existing issue or pull request to add to project (alternative to creating new issue)"),
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
				embeddable.PropertyDescription("Comma-separated labels to set on the issue/PR (labels must exist in the configured repository)"),
			),
		),

		// Update project item tool
		embeddable.WithEnhancedTool("update_project_item", handlers.UpdateProjectItem,
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
		embeddable.WithEnhancedTool("add_project_item_comment", handlers.AddProjectItemComment,
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
		embeddable.WithEnhancedTool("get_project_info", handlers.GetProjectInfo,
			embeddable.WithEnhancedDescription("Get detailed project information including all fields, field types, and repository labels"),
			embeddable.WithReadOnlyHint(true),
			embeddable.WithIdempotentHint(true),
		),
	)
}

// GetService returns the global GitHub service instance
func GetService() *GitHubProjectService {
	return githubService
}

// EnsureService initializes the global service with lazy config loading
func EnsureService(ctx context.Context) error {
	if githubService != nil {
		return nil
	}

	// Load config using the main package's helper (we'll need to access this somehow)
	// For now, let's use environment variables directly
	owner := os.Getenv("GITHUB_OWNER")
	if owner == "" {
		return fmt.Errorf("GITHUB_OWNER environment variable is required")
	}

	projectNumberStr := os.Getenv("GITHUB_PROJECT_NUMBER")
	if projectNumberStr == "" {
		return fmt.Errorf("GITHUB_PROJECT_NUMBER environment variable is required")
	}

	projectNumber, err := strconv.Atoi(projectNumberStr)
	if err != nil {
		return fmt.Errorf("invalid GITHUB_PROJECT_NUMBER: %v", err)
	}

	repository := os.Getenv("GITHUB_REPOSITORY")
	if repository == "" {
		return fmt.Errorf("GITHUB_REPOSITORY environment variable is required")
	}

	return ensureGitHubService(ctx, owner, projectNumber, repository)
}
