package cmds

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

	"github.com/go-go-golems/go-go-labs/pkg/github"
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
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &UpdateIssueCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *UpdateIssueCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	start := time.Now()

	// Parse settings
	s := &UpdateIssueSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Create contextual logger
	logger := log.With().
		Str("function", "RunIntoGlazeProcessor").
		Str("repo_owner", s.RepoOwner).
		Str("repo_name", s.RepoName).
		Int("issue_number", s.IssueNumber).
		Logger()

	logger.Debug().Msg("starting issue update")

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

	// Get the existing issue to verify it exists and get its ID
	getIssueStart := time.Now()
	logger.Debug().Msg("fetching existing issue")

	issue, err := client.GetIssue(ctx, s.RepoOwner, s.RepoName, s.IssueNumber)
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(getIssueStart)).
			Msg("failed to get existing issue")
		return errors.Wrap(err, "failed to get issue")
	}

	issueLogger := logger.With().
		Str("issue_id", issue.ID).
		Logger()

	issueLogger.Debug().
		Str("current_title", issue.Title).
		Int("current_label_count", len(issue.Labels)).
		Dur("duration", time.Since(getIssueStart)).
		Msg("existing issue found")

	var currentLabelNames []string
	for _, label := range issue.Labels {
		currentLabelNames = append(currentLabelNames, label.Name)
	}
	issueLogger.Trace().
		Strs("current_labels", currentLabelNames).
		Msg("current issue labels")

	// Update issue title and/or body if provided
	if s.Title != "" || s.Body != "" {
		issueLogger.Debug().
			Bool("updating_title", s.Title != "").
			Bool("updating_body", s.Body != "").
			Msg("updating issue fields")

		updateStart := time.Now()
		updatedIssue, err := client.UpdateIssue(ctx, issue.ID, s.Title, s.Body)
		if err != nil {
			issueLogger.Error().
				Err(err).
				Dur("duration", time.Since(updateStart)).
				Msg("failed to update issue fields")
			return errors.Wrap(err, "failed to update issue")
		}
		issue = updatedIssue

		issueLogger.Debug().
			Str("updated_title", issue.Title).
			Dur("duration", time.Since(updateStart)).
			Msg("issue fields updated")

		issueLogger.Info().Msg("issue updated successfully")
	}

	// Add labels if provided
	if s.AddLabels != "" {
		issueLogger.Debug().
			Str("raw_add_labels", s.AddLabels).
			Msg("processing labels to add")

		addLabelNames := strings.Split(s.AddLabels, ",")
		for i, label := range addLabelNames {
			addLabelNames[i] = strings.TrimSpace(label)
		}

		issueLogger.Trace().
			Strs("parsed_add_labels", addLabelNames).
			Msg("parsed labels to add")

		getLabelIDsStart := time.Now()
		addLabelIDs, err := client.GetLabelIDsByNames(ctx, s.RepoOwner, s.RepoName, addLabelNames)
		if err != nil {
			issueLogger.Error().
				Err(err).
				Strs("label_names", addLabelNames).
				Dur("duration", time.Since(getLabelIDsStart)).
				Msg("failed to get add label IDs")
			return errors.Wrap(err, "failed to get add label IDs")
		}

		issueLogger.Trace().
			Strs("add_label_ids", addLabelIDs).
			Dur("duration", time.Since(getLabelIDsStart)).
			Msg("retrieved label IDs for addition")

		addLabelsStart := time.Now()
		if err := client.AddLabelsToLabelable(ctx, issue.ID, addLabelIDs); err != nil {
			issueLogger.Error().
				Err(err).
				Strs("label_ids", addLabelIDs).
				Dur("duration", time.Since(addLabelsStart)).
				Msg("failed to add labels to issue")
			return errors.Wrap(err, "failed to add labels to issue")
		}

		issueLogger.Debug().
			Strs("added_label_names", addLabelNames).
			Dur("duration", time.Since(addLabelsStart)).
			Msg("labels added to issue")

		issueLogger.Info().
			Strs("labels", addLabelNames).
			Msg("labels added to issue")
	}

	// Remove labels if provided
	if s.RemoveLabels != "" {
		issueLogger.Debug().
			Str("raw_remove_labels", s.RemoveLabels).
			Msg("processing labels to remove")

		removeLabelNames := strings.Split(s.RemoveLabels, ",")
		for i, label := range removeLabelNames {
			removeLabelNames[i] = strings.TrimSpace(label)
		}

		issueLogger.Trace().
			Strs("parsed_remove_labels", removeLabelNames).
			Msg("parsed labels to remove")

		getLabelIDsStart := time.Now()
		removeLabelIDs, err := client.GetLabelIDsByNames(ctx, s.RepoOwner, s.RepoName, removeLabelNames)
		if err != nil {
			issueLogger.Error().
				Err(err).
				Strs("label_names", removeLabelNames).
				Dur("duration", time.Since(getLabelIDsStart)).
				Msg("failed to get remove label IDs")
			return errors.Wrap(err, "failed to get remove label IDs")
		}

		issueLogger.Trace().
			Strs("remove_label_ids", removeLabelIDs).
			Dur("duration", time.Since(getLabelIDsStart)).
			Msg("retrieved label IDs for removal")

		removeLabelsStart := time.Now()
		if err := client.RemoveLabelsFromLabelable(ctx, issue.ID, removeLabelIDs); err != nil {
			issueLogger.Error().
				Err(err).
				Strs("label_ids", removeLabelIDs).
				Dur("duration", time.Since(removeLabelsStart)).
				Msg("failed to remove labels from issue")
			return errors.Wrap(err, "failed to remove labels from issue")
		}

		issueLogger.Debug().
			Strs("removed_label_names", removeLabelNames).
			Dur("duration", time.Since(removeLabelsStart)).
			Msg("labels removed from issue")

		issueLogger.Info().
			Strs("labels", removeLabelNames).
			Msg("labels removed from issue")
	}

	// Get updated issue to reflect label changes
	if s.AddLabels != "" || s.RemoveLabels != "" {
		issueLogger.Trace().Msg("fetching updated issue to reflect label changes")

		getUpdatedIssueStart := time.Now()
		updatedIssue, err := client.GetIssue(ctx, s.RepoOwner, s.RepoName, s.IssueNumber)
		if err != nil {
			issueLogger.Error().
				Err(err).
				Dur("duration", time.Since(getUpdatedIssueStart)).
				Msg("failed to get updated issue after label changes")
			return errors.Wrap(err, "failed to get updated issue")
		}
		issue = updatedIssue

		issueLogger.Trace().
			Int("updated_label_count", len(issue.Labels)).
			Dur("duration", time.Since(getUpdatedIssueStart)).
			Msg("retrieved updated issue after label changes")
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

	if err := gp.AddRow(ctx, row); err != nil {
		issueLogger.Error().
			Err(err).
			Msg("failed to add row to processor")
		return errors.Wrap(err, "failed to add row to processor")
	}

	issueLogger.Debug().
		Dur("total_duration", time.Since(start)).
		Msg("issue update completed")

	return nil
}

// NewUpdateIssueCommand creates a new update issue command
func NewUpdateIssueCommand() (*UpdateIssueCommand, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "NewUpdateIssueCommand").
		Logger()

	logger.Trace().Msg("creating update issue command")

	// Create Glazed layer for output formatting
	glazedStart := time.Now()
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(glazedStart)).
			Msg("failed to create glazed parameter layers")
		return nil, err
	}
	logger.Trace().
		Dur("duration", time.Since(glazedStart)).
		Msg("glazed parameter layers created")

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
		),
		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	cmd := &UpdateIssueCommand{
		CommandDescription: cmdDesc,
	}

	logger.Trace().
		Dur("total_duration", time.Since(start)).
		Msg("update issue command created")

	return cmd, nil
}
