package main

import (
	"context"
	"os"
	"strings"

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

// UpdateIssueCommand updates an existing issue
type UpdateIssueCommand struct {
	*cmds.CommandDescription
}

// UpdateIssueSettings holds the command settings
type UpdateIssueSettings struct {
	RepoOwner    string `glazed.parameter:"repo-owner"`
	RepoName     string `glazed.parameter:"repo-name"`
	IssueNumber  int    `glazed.parameter:"issue-number"`
	Title        string `glazed.parameter:"title"`
	Body         string `glazed.parameter:"body"`
	AddLabels    string `glazed.parameter:"add-labels"`
	RemoveLabels string `glazed.parameter:"remove-labels"`
	LogLevel     string `glazed.parameter:"log-level"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &UpdateIssueCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *UpdateIssueCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings
	s := &UpdateIssueSettings{}
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

	// Get the existing issue to verify it exists and get its ID
	issue, err := client.GetIssue(ctx, s.RepoOwner, s.RepoName, s.IssueNumber)
	if err != nil {
		return errors.Wrap(err, "failed to get issue")
	}

	logger.Debug().
		Str("issue_id", issue.ID).
		Int("issue_number", issue.Number).
		Msg("Found existing issue")

	// Update issue title and/or body if provided
	if s.Title != "" || s.Body != "" {
		updatedIssue, err := client.UpdateIssue(ctx, issue.ID, s.Title, s.Body)
		if err != nil {
			return errors.Wrap(err, "failed to update issue")
		}
		issue = updatedIssue

		logger.Info().
			Str("issue_id", issue.ID).
			Int("issue_number", issue.Number).
			Msg("Issue updated successfully")
	}

	// Add labels if provided
	if s.AddLabels != "" {
		addLabelNames := strings.Split(s.AddLabels, ",")
		for i, label := range addLabelNames {
			addLabelNames[i] = strings.TrimSpace(label)
		}

		addLabelIDs, err := client.GetLabelIDsByNames(ctx, s.RepoOwner, s.RepoName, addLabelNames)
		if err != nil {
			return errors.Wrap(err, "failed to get add label IDs")
		}

		if err := client.AddLabelsToLabelable(ctx, issue.ID, addLabelIDs); err != nil {
			return errors.Wrap(err, "failed to add labels to issue")
		}

		logger.Info().
			Strs("labels", addLabelNames).
			Msg("Labels added to issue")
	}

	// Remove labels if provided
	if s.RemoveLabels != "" {
		removeLabelNames := strings.Split(s.RemoveLabels, ",")
		for i, label := range removeLabelNames {
			removeLabelNames[i] = strings.TrimSpace(label)
		}

		removeLabelIDs, err := client.GetLabelIDsByNames(ctx, s.RepoOwner, s.RepoName, removeLabelNames)
		if err != nil {
			return errors.Wrap(err, "failed to get remove label IDs")
		}

		if err := client.RemoveLabelsFromLabelable(ctx, issue.ID, removeLabelIDs); err != nil {
			return errors.Wrap(err, "failed to remove labels from issue")
		}

		logger.Info().
			Strs("labels", removeLabelNames).
			Msg("Labels removed from issue")
	}

	// Get updated issue to reflect label changes
	if s.AddLabels != "" || s.RemoveLabels != "" {
		updatedIssue, err := client.GetIssue(ctx, s.RepoOwner, s.RepoName, s.IssueNumber)
		if err != nil {
			return errors.Wrap(err, "failed to get updated issue")
		}
		issue = updatedIssue
	}

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

	return gp.AddRow(ctx, row)
}

// NewUpdateIssueCommand creates a new update issue command
func NewUpdateIssueCommand() (*UpdateIssueCommand, error) {
	// Create Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create command description
	cmdDesc := cmds.NewCommandDescription(
		"update-issue",
		cmds.WithShort("Update an existing issue's title, body, and labels"),
		cmds.WithLong(`
Update an existing GitHub issue in a repository. You can update the title, body, 
add labels, and remove labels. At least one of --title, --body, --add-labels, 
or --remove-labels must be provided.

Examples:
  github-graphql-cli update-issue --repo-owner=myorg --repo-name=myrepo --issue-number=123 --title="Updated title"
  github-graphql-cli update-issue --repo-owner=myorg --repo-name=myrepo --issue-number=123 --body="Updated description"
  github-graphql-cli update-issue --repo-owner=myorg --repo-name=myrepo --issue-number=123 --add-labels="bug,priority-high"
  github-graphql-cli update-issue --repo-owner=myorg --repo-name=myrepo --issue-number=123 --remove-labels="wontfix"
  github-graphql-cli update-issue --repo-owner=myorg --repo-name=myrepo --issue-number=123 --title="New title" --add-labels="enhancement"
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
				"issue-number",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Issue number to update"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"title",
				parameters.ParameterTypeString,
				parameters.WithHelp("New issue title (optional)"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"body",
				parameters.ParameterTypeString,
				parameters.WithHelp("New issue body (optional)"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"add-labels",
				parameters.ParameterTypeString,
				parameters.WithHelp("Comma-separated list of label names to add to the issue"),
				parameters.WithDefault(""),
			),
			parameters.NewParameterDefinition(
				"remove-labels",
				parameters.ParameterTypeString,
				parameters.WithHelp("Comma-separated list of label names to remove from the issue"),
				parameters.WithDefault(""),
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

	return &UpdateIssueCommand{
		CommandDescription: cmdDesc,
	}, nil
}
