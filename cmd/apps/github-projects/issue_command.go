package main

import (
	"context"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

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
	Labels        string `glazed.parameter:"labels"`
	ProjectOwner  string `glazed.parameter:"project-owner"`
	ProjectNumber int    `glazed.parameter:"project-number"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &IssueCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *IssueCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	start := time.Now()

	// Parse settings
	s := &IssueSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Create contextual logger
	logger := log.With().
		Str("function", "RunIntoGlazeProcessor").
		Str("repo_owner", s.RepoOwner).
		Str("repo_name", s.RepoName).
		Str("title", s.Title).
		Logger()

	logger.Debug().Msg("starting issue creation")

	// Create GitHub client
	clientStart := time.Now()
	client, err := github.NewClient()
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(clientStart)).
			Msg("failed to create GitHub client")
		return errors.Wrap(err, "failed to create GitHub client")
	}
	logger.Trace().
		Dur("duration", time.Since(clientStart)).
		Msg("GitHub client created")

	// Get repository
	repoStart := time.Now()
	logger.Debug().Msg("fetching repository information")
	repo, err := client.GetRepository(ctx, s.RepoOwner, s.RepoName)
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(repoStart)).
			Msg("failed to get repository")
		return errors.Wrap(err, "failed to get repository")
	}

	repoLogger := logger.With().
		Str("repo_id", repo.ID).
		Logger()

	repoLogger.Debug().
		Dur("duration", time.Since(repoStart)).
		Msg("repository fetched")

	// Create issue with or without labels
	var issue *github.Issue
	if s.Labels != "" {
		repoLogger.Debug().
			Str("labels_raw", s.Labels).
			Msg("processing labels for issue creation")

		labelNames := strings.Split(s.Labels, ",")
		for i, label := range labelNames {
			labelNames[i] = strings.TrimSpace(label)
		}

		repoLogger.Trace().
			Strs("label_names", labelNames).
			Msg("label names processed")

		// Get label IDs from names
		labelStart := time.Now()
		labelIDs, err := client.GetLabelIDsByNames(ctx, s.RepoOwner, s.RepoName, labelNames)
		if err != nil {
			repoLogger.Error().
				Err(err).
				Strs("label_names", labelNames).
				Dur("duration", time.Since(labelStart)).
				Msg("failed to get label IDs")
			return errors.Wrap(err, "failed to get label IDs")
		}
		repoLogger.Trace().
			Strs("label_ids", labelIDs).
			Dur("duration", time.Since(labelStart)).
			Msg("label IDs fetched")

		repoLogger.Debug().Msg("creating issue with labels")
		issueStart := time.Now()
		issue, err = client.CreateIssueWithLabels(ctx, repo.ID, s.Title, s.Body, labelIDs)
		if err != nil {
			repoLogger.Error().
				Err(err).
				Strs("label_ids", labelIDs).
				Dur("duration", time.Since(issueStart)).
				Msg("failed to create issue with labels")
			return errors.Wrap(err, "failed to create issue with labels")
		}
		repoLogger.Debug().
			Str("issue_id", issue.ID).
			Int("issue_number", issue.Number).
			Dur("duration", time.Since(issueStart)).
			Msg("issue with labels created")
	} else {
		repoLogger.Debug().Msg("creating issue without labels")
		issueStart := time.Now()
		issue, err = client.CreateIssue(ctx, repo.ID, s.Title, s.Body)
		if err != nil {
			repoLogger.Error().
				Err(err).
				Dur("duration", time.Since(issueStart)).
				Msg("failed to create issue")
			return errors.Wrap(err, "failed to create issue")
		}
		repoLogger.Debug().
			Str("issue_id", issue.ID).
			Int("issue_number", issue.Number).
			Dur("duration", time.Since(issueStart)).
			Msg("issue created")
	}

	issueLogger := repoLogger.With().
		Str("issue_id", issue.ID).
		Int("issue_number", issue.Number).
		Logger()

	issueLogger.Info().
		Str("issue_url", issue.URL).
		Msg("issue created successfully")

	// Extract label names for display
	var labelNames []string
	for _, label := range issue.Labels {
		labelNames = append(labelNames, label.Name)
	}

	row := types.NewRow(
		types.MRP("issue_id", issue.ID),
		types.MRP("issue_number", issue.Number),
		types.MRP("issue_url", issue.URL),
		types.MRP("title", issue.Title),
		types.MRP("body", issue.Body),
		types.MRP("labels", labelNames),
	)

	// If project details are provided, add issue to project
	if s.ProjectOwner != "" && s.ProjectNumber > 0 {
		issueLogger.Debug().
			Str("project_owner", s.ProjectOwner).
			Int("project_number", s.ProjectNumber).
			Msg("adding issue to project")

		projectStart := time.Now()
		project, err := client.GetProject(ctx, s.ProjectOwner, s.ProjectNumber)
		if err != nil {
			issueLogger.Warn().
				Err(err).
				Str("project_owner", s.ProjectOwner).
				Int("project_number", s.ProjectNumber).
				Dur("duration", time.Since(projectStart)).
				Msg("failed to get project, issue created but not added to project")
		} else {
			projectLogger := issueLogger.With().
				Str("project_id", project.ID).
				Str("project_title", project.Title).
				Logger()

			projectLogger.Debug().
				Dur("duration", time.Since(projectStart)).
				Msg("project fetched, adding issue to project")

			addItemStart := time.Now()
			itemID, err := client.AddItemToProject(ctx, project.ID, issue.ID)
			if err != nil {
				projectLogger.Warn().
					Err(err).
					Dur("duration", time.Since(addItemStart)).
					Msg("failed to add issue to project")
			} else {
				projectLogger.Info().
					Str("item_id", itemID).
					Dur("duration", time.Since(addItemStart)).
					Msg("issue added to project successfully")

				row.Set("project_item_id", itemID)
				row.Set("project_title", project.Title)
			}
		}
	}

	logger.Debug().
		Dur("total_duration", time.Since(start)).
		Msg("issue creation completed")

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
				"labels",
				parameters.ParameterTypeString,
				parameters.WithHelp("Comma-separated list of label names to apply to the issue"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"project-owner",
				parameters.ParameterTypeString,
				parameters.WithHelp("Project owner (optional, to add issue to project)"),
				parameters.WithDefault(GetDefaultOwner()),
			),
			parameters.NewParameterDefinition(
				"project-number",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Project number (optional, to add issue to project)"),
				parameters.WithDefault(GetDefaultProjectNumber()),
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
