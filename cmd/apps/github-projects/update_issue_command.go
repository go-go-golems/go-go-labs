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
	startTime := time.Now()

	// Parse settings
	s := &UpdateIssueSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	log.Debug().
		Str("function", "RunIntoGlazeProcessor").
		Str("repo_owner", s.RepoOwner).
		Str("repo_name", s.RepoName).
		Int("issue_number", s.IssueNumber).
		Str("title", s.Title).
		Str("body", s.Body).
		Str("add_labels", s.AddLabels).
		Str("remove_labels", s.RemoveLabels).
		Msg("Function entry - starting issue update")

	defer func() {
		duration := time.Since(startTime)
		log.Debug().
			Str("function", "RunIntoGlazeProcessor").
			Dur("duration_ms", duration).
			Msg("Function exit - completed issue update")
	}()

	// Create GitHub client
	log.Debug().Msg("Creating GitHub client")
	clientStartTime := time.Now()
	client, err := github.NewClient()
	if err != nil {
		log.Error().
			Err(err).
			Dur("duration_ms", time.Since(clientStartTime)).
			Msg("Failed to create GitHub client")
		return errors.Wrap(err, "failed to create GitHub client")
	}
	log.Debug().
		Dur("duration_ms", time.Since(clientStartTime)).
		Msg("GitHub client created successfully")

	// Get the existing issue to verify it exists and get its ID
	log.Debug().
		Str("repo_owner", s.RepoOwner).
		Str("repo_name", s.RepoName).
		Int("issue_number", s.IssueNumber).
		Msg("Fetching existing issue")

	getIssueStartTime := time.Now()
	issue, err := client.GetIssue(ctx, s.RepoOwner, s.RepoName, s.IssueNumber)
	if err != nil {
		log.Error().
			Err(err).
			Str("repo_owner", s.RepoOwner).
			Str("repo_name", s.RepoName).
			Int("issue_number", s.IssueNumber).
			Dur("duration_ms", time.Since(getIssueStartTime)).
			Msg("Failed to get existing issue")
		return errors.Wrap(err, "failed to get issue")
	}

	log.Debug().
		Str("issue_id", issue.ID).
		Int("issue_number", issue.Number).
		Str("current_title", issue.Title).
		Str("current_url", issue.URL).
		Int("current_label_count", len(issue.Labels)).
		Dur("duration_ms", time.Since(getIssueStartTime)).
		Msg("Found existing issue")

	var currentLabelNames []string
	for _, label := range issue.Labels {
		currentLabelNames = append(currentLabelNames, label.Name)
	}
	log.Debug().
		Strs("current_labels", currentLabelNames).
		Msg("Current issue labels")

	// Update issue title and/or body if provided
	if s.Title != "" || s.Body != "" {
		log.Debug().
			Bool("updating_title", s.Title != "").
			Bool("updating_body", s.Body != "").
			Str("new_title", s.Title).
			Int("new_body_length", len(s.Body)).
			Msg("Starting issue field update")

		updateStartTime := time.Now()
		updatedIssue, err := client.UpdateIssue(ctx, issue.ID, s.Title, s.Body)
		if err != nil {
			log.Error().
				Err(err).
				Str("issue_id", issue.ID).
				Str("new_title", s.Title).
				Int("new_body_length", len(s.Body)).
				Dur("duration_ms", time.Since(updateStartTime)).
				Msg("Failed to update issue fields")
			return errors.Wrap(err, "failed to update issue")
		}
		issue = updatedIssue

		log.Debug().
			Str("issue_id", issue.ID).
			Int("issue_number", issue.Number).
			Str("updated_title", issue.Title).
			Int("updated_body_length", len(issue.Body)).
			Dur("duration_ms", time.Since(updateStartTime)).
			Msg("Issue fields updated successfully")

		log.Info().
			Str("issue_id", issue.ID).
			Int("issue_number", issue.Number).
			Msg("Issue updated successfully")
	} else {
		log.Debug().Msg("No field updates requested - skipping issue update")
	}

	// Add labels if provided
	if s.AddLabels != "" {
		log.Debug().
			Str("raw_add_labels", s.AddLabels).
			Msg("Starting label addition process")

		addLabelNames := strings.Split(s.AddLabels, ",")
		for i, label := range addLabelNames {
			addLabelNames[i] = strings.TrimSpace(label)
		}

		log.Debug().
			Strs("parsed_add_labels", addLabelNames).
			Int("label_count", len(addLabelNames)).
			Msg("Parsed labels to add")

		getLabelIDsStartTime := time.Now()
		addLabelIDs, err := client.GetLabelIDsByNames(ctx, s.RepoOwner, s.RepoName, addLabelNames)
		if err != nil {
			log.Error().
				Err(err).
				Str("repo_owner", s.RepoOwner).
				Str("repo_name", s.RepoName).
				Strs("label_names", addLabelNames).
				Dur("duration_ms", time.Since(getLabelIDsStartTime)).
				Msg("Failed to get add label IDs")
			return errors.Wrap(err, "failed to get add label IDs")
		}

		log.Debug().
			Strs("add_label_names", addLabelNames).
			Strs("add_label_ids", addLabelIDs).
			Dur("duration_ms", time.Since(getLabelIDsStartTime)).
			Msg("Retrieved label IDs for addition")

		addLabelsStartTime := time.Now()
		if err := client.AddLabelsToLabelable(ctx, issue.ID, addLabelIDs); err != nil {
			log.Error().
				Err(err).
				Str("issue_id", issue.ID).
				Strs("label_ids", addLabelIDs).
				Strs("label_names", addLabelNames).
				Dur("duration_ms", time.Since(addLabelsStartTime)).
				Msg("Failed to add labels to issue")
			return errors.Wrap(err, "failed to add labels to issue")
		}

		log.Debug().
			Str("issue_id", issue.ID).
			Strs("added_label_ids", addLabelIDs).
			Strs("added_label_names", addLabelNames).
			Dur("duration_ms", time.Since(addLabelsStartTime)).
			Msg("Labels added to issue successfully")

		log.Info().
			Strs("labels", addLabelNames).
			Msg("Labels added to issue")
	} else {
		log.Debug().Msg("No labels to add - skipping label addition")
	}

	// Remove labels if provided
	if s.RemoveLabels != "" {
		log.Debug().
			Str("raw_remove_labels", s.RemoveLabels).
			Msg("Starting label removal process")

		removeLabelNames := strings.Split(s.RemoveLabels, ",")
		for i, label := range removeLabelNames {
			removeLabelNames[i] = strings.TrimSpace(label)
		}

		log.Debug().
			Strs("parsed_remove_labels", removeLabelNames).
			Int("label_count", len(removeLabelNames)).
			Msg("Parsed labels to remove")

		getLabelIDsStartTime := time.Now()
		removeLabelIDs, err := client.GetLabelIDsByNames(ctx, s.RepoOwner, s.RepoName, removeLabelNames)
		if err != nil {
			log.Error().
				Err(err).
				Str("repo_owner", s.RepoOwner).
				Str("repo_name", s.RepoName).
				Strs("label_names", removeLabelNames).
				Dur("duration_ms", time.Since(getLabelIDsStartTime)).
				Msg("Failed to get remove label IDs")
			return errors.Wrap(err, "failed to get remove label IDs")
		}

		log.Debug().
			Strs("remove_label_names", removeLabelNames).
			Strs("remove_label_ids", removeLabelIDs).
			Dur("duration_ms", time.Since(getLabelIDsStartTime)).
			Msg("Retrieved label IDs for removal")

		removeLabelsStartTime := time.Now()
		if err := client.RemoveLabelsFromLabelable(ctx, issue.ID, removeLabelIDs); err != nil {
			log.Error().
				Err(err).
				Str("issue_id", issue.ID).
				Strs("label_ids", removeLabelIDs).
				Strs("label_names", removeLabelNames).
				Dur("duration_ms", time.Since(removeLabelsStartTime)).
				Msg("Failed to remove labels from issue")
			return errors.Wrap(err, "failed to remove labels from issue")
		}

		log.Debug().
			Str("issue_id", issue.ID).
			Strs("removed_label_ids", removeLabelIDs).
			Strs("removed_label_names", removeLabelNames).
			Dur("duration_ms", time.Since(removeLabelsStartTime)).
			Msg("Labels removed from issue successfully")

		log.Info().
			Strs("labels", removeLabelNames).
			Msg("Labels removed from issue")
	} else {
		log.Debug().Msg("No labels to remove - skipping label removal")
	}

	// Get updated issue to reflect label changes
	if s.AddLabels != "" || s.RemoveLabels != "" {
		log.Debug().
			Bool("had_label_additions", s.AddLabels != "").
			Bool("had_label_removals", s.RemoveLabels != "").
			Msg("Fetching updated issue to reflect label changes")

		getUpdatedIssueStartTime := time.Now()
		updatedIssue, err := client.GetIssue(ctx, s.RepoOwner, s.RepoName, s.IssueNumber)
		if err != nil {
			log.Error().
				Err(err).
				Str("repo_owner", s.RepoOwner).
				Str("repo_name", s.RepoName).
				Int("issue_number", s.IssueNumber).
				Dur("duration_ms", time.Since(getUpdatedIssueStartTime)).
				Msg("Failed to get updated issue after label changes")
			return errors.Wrap(err, "failed to get updated issue")
		}
		issue = updatedIssue

		log.Debug().
			Str("issue_id", issue.ID).
			Int("updated_label_count", len(issue.Labels)).
			Dur("duration_ms", time.Since(getUpdatedIssueStartTime)).
			Msg("Retrieved updated issue after label changes")
	} else {
		log.Debug().Msg("No label changes made - skipping updated issue fetch")
	}

	// Extract label names for display
	log.Debug().Msg("Extracting final label names for output")
	var labelNames []string
	for _, label := range issue.Labels {
		labelNames = append(labelNames, label.Name)
	}

	log.Debug().
		Strs("final_labels", labelNames).
		Int("final_label_count", len(labelNames)).
		Msg("Final issue labels extracted")

	log.Debug().
		Str("issue_id", issue.ID).
		Int("issue_number", issue.Number).
		Str("final_title", issue.Title).
		Str("final_url", issue.URL).
		Int("final_body_length", len(issue.Body)).
		Strs("final_labels", labelNames).
		Msg("Creating output row")

	row := types.NewRow(
		types.MRP("issue_id", issue.ID),
		types.MRP("issue_number", issue.Number),
		types.MRP("issue_url", issue.URL),
		types.MRP("title", issue.Title),
		types.MRP("body", issue.Body),
		types.MRP("labels", labelNames),
	)

	log.Debug().Msg("Adding row to processor")
	if err := gp.AddRow(ctx, row); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to add row to processor")
		return errors.Wrap(err, "failed to add row to processor")
	}

	log.Debug().
		Int("issue_number", issue.Number).
		Dur("total_duration_ms", time.Since(startTime)).
		Msg("Issue update completed successfully")

	return nil
}

// NewUpdateIssueCommand creates a new update issue command
func NewUpdateIssueCommand() (*UpdateIssueCommand, error) {
	startTime := time.Now()

	log.Debug().
		Str("function", "NewUpdateIssueCommand").
		Msg("Function entry - creating update issue command")

	defer func() {
		duration := time.Since(startTime)
		log.Debug().
			Str("function", "NewUpdateIssueCommand").
			Dur("duration_ms", duration).
			Msg("Function exit - update issue command created")
	}()

	// Create Glazed layer for output formatting
	log.Debug().Msg("Creating glazed parameter layers")
	glazedLayerStartTime := time.Now()
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		log.Error().
			Err(err).
			Dur("duration_ms", time.Since(glazedLayerStartTime)).
			Msg("Failed to create glazed parameter layers")
		return nil, err
	}
	log.Debug().
		Dur("duration_ms", time.Since(glazedLayerStartTime)).
		Msg("Glazed parameter layers created successfully")

	// Create command description
	log.Debug().Msg("Creating command description with parameters")
	cmdDescStartTime := time.Now()
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

	log.Debug().
		Dur("duration_ms", time.Since(cmdDescStartTime)).
		Int("parameter_count", 7).
		Msg("Command description created with all parameters")

	log.Debug().Msg("Creating UpdateIssueCommand instance")
	cmd := &UpdateIssueCommand{
		CommandDescription: cmdDesc,
	}

	log.Debug().
		Str("command_name", "update-issue").
		Msg("UpdateIssueCommand instance created successfully")

	return cmd, nil
}
