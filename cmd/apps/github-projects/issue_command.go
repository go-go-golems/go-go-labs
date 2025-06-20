package main

import (
	"context"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/go-go-golems/go-go-labs/cmd/apps/github-projects/pkg/github"
)

// IssueCommand creates an issue and optionally adds it to a project
type IssueCommand struct {
	*cmds.CommandDescription
}

// IssueSettings holds the command settings
type IssueSettings struct {
	RepoOwner     string `glazed.parameter:"repo-owner"`
	RepoName      string `glazed.parameter:"repo-name"`
	Title         string `glazed.parameter:"title"`
	Body          string `glazed.parameter:"body"`
	ProjectOwner  string `glazed.parameter:"project-owner"`
	ProjectNumber int    `glazed.parameter:"project-number"`
	LogLevel      string `glazed.parameter:"log-level"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &IssueCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *IssueCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings
	s := &IssueSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Set up logger
	level, err := zerolog.ParseLevel(s.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	logger := zerolog.New(os.Stderr).Level(level).With().Timestamp().Logger()

	// Create GitHub client
	client, err := github.NewClient(logger)
	if err != nil {
		return errors.Wrap(err, "failed to create GitHub client")
	}

	// Get repository
	repo, err := client.GetRepository(ctx, s.RepoOwner, s.RepoName)
	if err != nil {
		return errors.Wrap(err, "failed to get repository")
	}

	// Create issue
	issue, err := client.CreateIssue(ctx, repo.ID, s.Title, s.Body)
	if err != nil {
		return errors.Wrap(err, "failed to create issue")
	}

	logger.Info().
		Str("issue_id", issue.ID).
		Int("issue_number", issue.Number).
		Str("issue_url", issue.URL).
		Msg("Issue created successfully")

	row := types.NewRow(
		types.MRP("issue_id", issue.ID),
		types.MRP("issue_number", issue.Number),
		types.MRP("issue_url", issue.URL),
		types.MRP("title", issue.Title),
		types.MRP("body", issue.Body),
	)

	// If project details are provided, add issue to project
	if s.ProjectOwner != "" && s.ProjectNumber > 0 {
		project, err := client.GetProject(ctx, s.ProjectOwner, s.ProjectNumber)
		if err != nil {
			logger.Warn().
				Err(err).
				Str("project_owner", s.ProjectOwner).
				Int("project_number", s.ProjectNumber).
				Msg("Failed to get project, issue created but not added to project")
		} else {
			itemID, err := client.AddItemToProject(ctx, project.ID, issue.ID)
			if err != nil {
				logger.Warn().
					Err(err).
					Str("project_id", project.ID).
					Str("issue_id", issue.ID).
					Msg("Failed to add issue to project")
			} else {
				logger.Info().
					Str("project_id", project.ID).
					Str("item_id", itemID).
					Msg("Issue added to project successfully")
				row.Set("project_item_id", itemID)
				row.Set("project_title", project.Title)
			}
		}
	}

	return gp.AddRow(ctx, row)
}

// NewIssueCommand creates a new issue command
func NewIssueCommand() (*IssueCommand, error) {
	// Create Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"create-issue",
		cmds.WithShort("Create an issue and optionally add it to a project"),
		cmds.WithLong(`
Create a new GitHub issue in a repository and optionally add it to a project.

Examples:
  github-graphql-cli create-issue --repo-owner=myorg --repo-name=myrepo --title="Bug fix" --body="Description"
  github-graphql-cli create-issue --repo-owner=myorg --repo-name=myrepo --title="Feature" --project-owner=myorg --project-number=5
		`),
		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"repo-owner",
				parameters.ParameterTypeString,
				parameters.WithHelp("Repository owner"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"repo-name",
				parameters.ParameterTypeString,
				parameters.WithHelp("Repository name"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"title",
				parameters.ParameterTypeString,
				parameters.WithHelp("Issue title"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"body",
				parameters.ParameterTypeString,
				parameters.WithHelp("Issue body"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"project-owner",
				parameters.ParameterTypeString,
				parameters.WithHelp("Project owner (optional, to add issue to project)"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"project-number",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Project number (optional, to add issue to project)"),
				parameters.WithDefault(0),
			),
			parameters.NewParameterDefinition(
				"log-level",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Log level"),
				parameters.WithDefault("info"),
				parameters.WithChoices("trace", "debug", "info", "warn", "error"),
			),
		),
		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &IssueCommand{
		CommandDescription: cmdDesc,
	}, nil
}
