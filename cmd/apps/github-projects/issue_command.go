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
	startTime := time.Now()

	// Parse settings
	s := &IssueSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	log.Debug().
		Str("function", "RunIntoGlazeProcessor").
		Str("repo_owner", s.RepoOwner).
		Str("repo_name", s.RepoName).
		Str("title", s.Title).
		Int("body_length", len(s.Body)).
		Str("labels", s.Labels).
		Str("project_owner", s.ProjectOwner).
		Int("project_number", s.ProjectNumber).
		Msg("Function entry - starting issue creation process")

	defer func() {
		duration := time.Since(startTime)
		log.Debug().
			Str("function", "RunIntoGlazeProcessor").
			Dur("duration", duration).
			Msg("Function exit - issue creation process completed")
	}()

	// Create GitHub client
	log.Debug().Msg("Creating GitHub client")
	clientStartTime := time.Now()
	client, err := github.NewClient()
	if err != nil {
		log.Error().
			Err(err).
			Dur("duration", time.Since(clientStartTime)).
			Msg("Failed to create GitHub client")
		return errors.Wrap(err, "failed to create GitHub client")
	}
	log.Debug().
		Dur("duration", time.Since(clientStartTime)).
		Msg("GitHub client created successfully")

	// Get repository
	log.Debug().
		Str("repo_owner", s.RepoOwner).
		Str("repo_name", s.RepoName).
		Msg("Fetching repository information")
	repoStartTime := time.Now()
	repo, err := client.GetRepository(ctx, s.RepoOwner, s.RepoName)
	if err != nil {
		log.Error().
			Err(err).
			Str("repo_owner", s.RepoOwner).
			Str("repo_name", s.RepoName).
			Dur("duration", time.Since(repoStartTime)).
			Msg("Failed to get repository")
		return errors.Wrap(err, "failed to get repository")
	}
	log.Debug().
		Str("repo_id", repo.ID).
		Str("repo_owner", s.RepoOwner).
		Str("repo_name", s.RepoName).
		Dur("duration", time.Since(repoStartTime)).
		Msg("Repository fetched successfully")

	// Create issue with or without labels
	var issue *github.Issue
	if s.Labels != "" {
		log.Debug().
			Str("labels_raw", s.Labels).
			Msg("Processing labels for issue creation")

		labelNames := strings.Split(s.Labels, ",")
		for i, label := range labelNames {
			labelNames[i] = strings.TrimSpace(label)
		}

		log.Debug().
			Strs("label_names", labelNames).
			Int("label_count", len(labelNames)).
			Msg("Label names processed")

		// Get label IDs from names
		log.Debug().
			Strs("label_names", labelNames).
			Msg("Fetching label IDs from names")
		labelStartTime := time.Now()
		labelIDs, err := client.GetLabelIDsByNames(ctx, s.RepoOwner, s.RepoName, labelNames)
		if err != nil {
			log.Error().
				Err(err).
				Strs("label_names", labelNames).
				Str("repo_owner", s.RepoOwner).
				Str("repo_name", s.RepoName).
				Dur("duration", time.Since(labelStartTime)).
				Msg("Failed to get label IDs")
			return errors.Wrap(err, "failed to get label IDs")
		}
		log.Debug().
			Strs("label_ids", labelIDs).
			Strs("label_names", labelNames).
			Dur("duration", time.Since(labelStartTime)).
			Msg("Label IDs fetched successfully")

		log.Debug().
			Str("repo_id", repo.ID).
			Str("title", s.Title).
			Int("body_length", len(s.Body)).
			Strs("label_ids", labelIDs).
			Msg("Creating issue with labels")
		issueStartTime := time.Now()
		issue, err = client.CreateIssueWithLabels(ctx, repo.ID, s.Title, s.Body, labelIDs)
		if err != nil {
			log.Error().
				Err(err).
				Str("repo_id", repo.ID).
				Str("title", s.Title).
				Strs("label_ids", labelIDs).
				Dur("duration", time.Since(issueStartTime)).
				Msg("Failed to create issue with labels")
			return errors.Wrap(err, "failed to create issue with labels")
		}
		log.Debug().
			Str("issue_id", issue.ID).
			Int("issue_number", issue.Number).
			Strs("label_ids", labelIDs).
			Dur("duration", time.Since(issueStartTime)).
			Msg("Issue with labels created successfully")
	} else {
		log.Debug().
			Str("repo_id", repo.ID).
			Str("title", s.Title).
			Int("body_length", len(s.Body)).
			Msg("Creating issue without labels")
		issueStartTime := time.Now()
		issue, err = client.CreateIssue(ctx, repo.ID, s.Title, s.Body)
		if err != nil {
			log.Error().
				Err(err).
				Str("repo_id", repo.ID).
				Str("title", s.Title).
				Dur("duration", time.Since(issueStartTime)).
				Msg("Failed to create issue")
			return errors.Wrap(err, "failed to create issue")
		}
		log.Debug().
			Str("issue_id", issue.ID).
			Int("issue_number", issue.Number).
			Dur("duration", time.Since(issueStartTime)).
			Msg("Issue created successfully")
	}

	log.Info().
		Str("issue_id", issue.ID).
		Int("issue_number", issue.Number).
		Str("issue_url", issue.URL).
		Msg("Issue created successfully")

	// Extract label names for display
	log.Debug().
		Str("issue_id", issue.ID).
		Int("label_count", len(issue.Labels)).
		Msg("Extracting label names for display")
	var labelNames []string
	for _, label := range issue.Labels {
		labelNames = append(labelNames, label.Name)
	}
	log.Debug().
		Strs("display_labels", labelNames).
		Msg("Label names extracted for display")

	log.Debug().
		Str("issue_id", issue.ID).
		Int("issue_number", issue.Number).
		Str("issue_url", issue.URL).
		Msg("Building output row")
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
		log.Debug().
			Str("project_owner", s.ProjectOwner).
			Int("project_number", s.ProjectNumber).
			Str("issue_id", issue.ID).
			Msg("Project details provided, attempting to add issue to project")

		projectStartTime := time.Now()
		project, err := client.GetProject(ctx, s.ProjectOwner, s.ProjectNumber)
		if err != nil {
			log.Warn().
				Err(err).
				Str("project_owner", s.ProjectOwner).
				Int("project_number", s.ProjectNumber).
				Dur("duration", time.Since(projectStartTime)).
				Msg("Failed to get project, issue created but not added to project")
		} else {
			log.Debug().
				Str("project_id", project.ID).
				Str("project_title", project.Title).
				Str("project_owner", s.ProjectOwner).
				Int("project_number", s.ProjectNumber).
				Dur("get_project_duration", time.Since(projectStartTime)).
				Msg("Project fetched successfully, adding issue to project")

			addItemStartTime := time.Now()
			itemID, err := client.AddItemToProject(ctx, project.ID, issue.ID)
			if err != nil {
				log.Warn().
					Err(err).
					Str("project_id", project.ID).
					Str("project_title", project.Title).
					Str("issue_id", issue.ID).
					Dur("duration", time.Since(addItemStartTime)).
					Msg("Failed to add issue to project")
			} else {
				log.Info().
					Str("project_id", project.ID).
					Str("project_title", project.Title).
					Str("item_id", itemID).
					Str("issue_id", issue.ID).
					Dur("duration", time.Since(addItemStartTime)).
					Msg("Issue added to project successfully")

				log.Debug().
					Str("item_id", itemID).
					Str("project_title", project.Title).
					Msg("Updating output row with project information")
				row.Set("project_item_id", itemID)
				row.Set("project_title", project.Title)
			}
		}
	} else {
		log.Debug().
			Str("project_owner", s.ProjectOwner).
			Int("project_number", s.ProjectNumber).
			Msg("No project details provided or incomplete, skipping project integration")
	}

	log.Debug().
		Str("issue_id", issue.ID).
		Int("issue_number", issue.Number).
		Msg("Adding row to processor")
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
				parameters.WithDefault(githubConfig.Owner),
			),
			parameters.NewParameterDefinition(
				"project-number",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Project number (optional, to add issue to project)"),
				parameters.WithDefault(githubConfig.ProjectNumber),
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
