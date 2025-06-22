package github

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Issue represents a GitHub issue
type Issue struct {
	ID     string  `json:"id"`
	Number int     `json:"number"`
	URL    string  `json:"url"`
	Title  string  `json:"title"`
	Body   string  `json:"body"`
	Labels []Label `json:"labels"`
}

// Label represents a GitHub label
type Label struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Comment represents a GitHub issue comment
type Comment struct {
	ID        string    `json:"id"`
	Body      string    `json:"body"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Author    Author    `json:"author"`
}

// Author represents a GitHub comment author
type Author struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatarUrl"`
}

// Repository represents a GitHub repository
type Repository struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetRepository retrieves a repository by owner and name
func (c *Client) GetRepository(ctx context.Context, owner, name string) (*Repository, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "GetRepository").
		Str("owner", owner).
		Str("name", name).
		Logger()

	logger.Debug().Msg("entering GetRepository")
	defer func() {
		logger.Debug().
			Dur("duration", time.Since(start)).
			Msg("exiting GetRepository")
	}()

	query := `
		query($owner: String!, $name: String!) {
			repository(owner: $owner, name: $name) {
				id
				name
			}
		}
	`

	logger.Debug().
		Str("query", query).
		Msg("constructed GraphQL query for repository retrieval")

	variables := map[string]interface{}{
		"owner": owner,
		"name":  name,
	}

	logger.Debug().
		Interface("variables", variables).
		Msg("prepared GraphQL variables")

	var resp struct {
		Repository Repository `json:"repository"`
	}

	logger.Debug().Msg("executing GraphQL query for repository")
	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("failed to execute GraphQL query for repository")
		return nil, errors.Wrap(err, "failed to get repository")
	}

	logger.Debug().
		Str("repository_id", resp.Repository.ID).
		Str("repository_name", resp.Repository.Name).
		Msg("successfully retrieved repository")

	return &resp.Repository, nil
}

// GetIssue retrieves an issue by repository and number
func (c *Client) GetIssue(ctx context.Context, owner, name string, number int) (*Issue, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "GetIssue").
		Str("owner", owner).
		Str("name", name).
		Int("number", number).
		Logger()

	logger.Debug().Msg("entering GetIssue")
	defer func() {
		logger.Debug().
			Dur("duration", time.Since(start)).
			Msg("exiting GetIssue")
	}()

	query := `
		query($owner: String!, $name: String!, $number: Int!) {
			repository(owner: $owner, name: $name) {
				issue(number: $number) {
					id
					number
					url
					title
					body
					labels(first: 100) {
						nodes {
							id
							name
						}
					}
				}
			}
		}
	`

	logger.Debug().
		Str("query", query).
		Msg("constructed GraphQL query for issue retrieval")

	variables := map[string]interface{}{
		"owner":  owner,
		"name":   name,
		"number": number,
	}

	logger.Debug().
		Interface("variables", variables).
		Msg("prepared GraphQL variables for issue query")

	var resp struct {
		Repository struct {
			Issue struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				URL    string `json:"url"`
				Title  string `json:"title"`
				Body   string `json:"body"`
				Labels struct {
					Nodes []Label `json:"nodes"`
				} `json:"labels"`
			} `json:"issue"`
		} `json:"repository"`
	}

	logger.Debug().Msg("executing GraphQL query for issue")
	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("failed to execute GraphQL query for issue")
		return nil, errors.Wrap(err, "failed to get issue")
	}

	issue := &Issue{
		ID:     resp.Repository.Issue.ID,
		Number: resp.Repository.Issue.Number,
		URL:    resp.Repository.Issue.URL,
		Title:  resp.Repository.Issue.Title,
		Body:   resp.Repository.Issue.Body,
		Labels: resp.Repository.Issue.Labels.Nodes,
	}

	logger.Debug().
		Str("issue_id", issue.ID).
		Int("issue_number", issue.Number).
		Str("issue_url", issue.URL).
		Str("issue_title", issue.Title).
		Int("labels_count", len(issue.Labels)).
		Msg("successfully retrieved issue")

	return issue, nil
}

// CreateIssue creates a new issue in a repository
func (c *Client) CreateIssue(ctx context.Context, repositoryID, title, body string) (*Issue, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "CreateIssue").
		Str("repository_id", repositoryID).
		Str("title", title).
		Int("body_length", len(body)).
		Logger()

	logger.Debug().Msg("entering CreateIssue")
	defer func() {
		logger.Debug().
			Dur("duration", time.Since(start)).
			Msg("exiting CreateIssue")
	}()

	mutation := `
		mutation($repositoryId: ID!, $title: String!, $body: String) {
			createIssue(input: { repositoryId: $repositoryId, title: $title, body: $body }) {
				issue {
					id
					number
					url
					title
					body
					labels(first: 100) {
						nodes {
							id
							name
						}
					}
				}
			}
		}
	`

	logger.Debug().
		Str("mutation", mutation).
		Msg("constructed GraphQL mutation for issue creation")

	variables := map[string]interface{}{
		"repositoryId": repositoryID,
		"title":        title,
		"body":         body,
	}

	logger.Debug().
		Interface("variables", variables).
		Msg("prepared GraphQL variables for issue creation")

	var resp struct {
		CreateIssue struct {
			Issue struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				URL    string `json:"url"`
				Title  string `json:"title"`
				Body   string `json:"body"`
				Labels struct {
					Nodes []Label `json:"nodes"`
				} `json:"labels"`
			} `json:"issue"`
		} `json:"createIssue"`
	}

	logger.Debug().Msg("executing GraphQL mutation for issue creation")
	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("failed to execute GraphQL mutation for issue creation")
		return nil, errors.Wrap(err, "failed to create issue")
	}

	issue := &Issue{
		ID:     resp.CreateIssue.Issue.ID,
		Number: resp.CreateIssue.Issue.Number,
		URL:    resp.CreateIssue.Issue.URL,
		Title:  resp.CreateIssue.Issue.Title,
		Body:   resp.CreateIssue.Issue.Body,
		Labels: resp.CreateIssue.Issue.Labels.Nodes,
	}

	logger.Debug().
		Str("created_issue_id", issue.ID).
		Int("created_issue_number", issue.Number).
		Str("created_issue_url", issue.URL).
		Int("initial_labels_count", len(issue.Labels)).
		Msg("successfully created issue")

	return issue, nil
}

// CreateIssueWithLabels creates a new issue in a repository with labels
func (c *Client) CreateIssueWithLabels(ctx context.Context, repositoryID, title, body string, labelIDs []string) (*Issue, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "CreateIssueWithLabels").
		Str("repository_id", repositoryID).
		Str("title", title).
		Int("body_length", len(body)).
		Int("label_ids_count", len(labelIDs)).
		Strs("label_ids", labelIDs).
		Logger()

	logger.Debug().Msg("entering CreateIssueWithLabels")
	defer func() {
		logger.Debug().
			Dur("duration", time.Since(start)).
			Msg("exiting CreateIssueWithLabels")
	}()

	mutation := `
		mutation($repositoryId: ID!, $title: String!, $body: String, $labelIds: [ID!]) {
			createIssue(input: { repositoryId: $repositoryId, title: $title, body: $body, labelIds: $labelIds }) {
				issue {
					id
					number
					url
					title
					body
					labels(first: 100) {
						nodes {
							id
							name
						}
					}
				}
			}
		}
	`

	logger.Debug().
		Str("mutation", mutation).
		Msg("constructed GraphQL mutation for issue creation with labels")

	variables := map[string]interface{}{
		"repositoryId": repositoryID,
		"title":        title,
		"body":         body,
		"labelIds":     labelIDs,
	}

	logger.Debug().
		Interface("variables", variables).
		Msg("prepared GraphQL variables for issue creation with labels")

	var resp struct {
		CreateIssue struct {
			Issue struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				URL    string `json:"url"`
				Title  string `json:"title"`
				Body   string `json:"body"`
				Labels struct {
					Nodes []Label `json:"nodes"`
				} `json:"labels"`
			} `json:"issue"`
		} `json:"createIssue"`
	}

	logger.Debug().Msg("executing GraphQL mutation for issue creation with labels")
	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("failed to execute GraphQL mutation for issue creation with labels")
		return nil, errors.Wrap(err, "failed to create issue with labels")
	}

	issue := &Issue{
		ID:     resp.CreateIssue.Issue.ID,
		Number: resp.CreateIssue.Issue.Number,
		URL:    resp.CreateIssue.Issue.URL,
		Title:  resp.CreateIssue.Issue.Title,
		Body:   resp.CreateIssue.Issue.Body,
		Labels: resp.CreateIssue.Issue.Labels.Nodes,
	}

	logger.Debug().
		Str("created_issue_id", issue.ID).
		Int("created_issue_number", issue.Number).
		Str("created_issue_url", issue.URL).
		Int("applied_labels_count", len(issue.Labels)).
		Msg("successfully created issue with labels")

	return issue, nil
}

// UpdateIssue updates an existing issue
func (c *Client) UpdateIssue(ctx context.Context, issueID, title, body string) (*Issue, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "UpdateIssue").
		Str("issue_id", issueID).
		Str("new_title", title).
		Int("new_body_length", len(body)).
		Bool("updating_title", title != "").
		Bool("updating_body", body != "").
		Logger()

	logger.Debug().Msg("entering UpdateIssue")
	defer func() {
		logger.Debug().
			Dur("duration", time.Since(start)).
			Msg("exiting UpdateIssue")
	}()

	mutation := `
		mutation($id: ID!, $title: String, $body: String) {
			updateIssue(input: { id: $id, title: $title, body: $body }) {
				issue {
					id
					number
					url
					title
					body
					labels(first: 100) {
						nodes {
							id
							name
						}
					}
				}
			}
		}
	`

	logger.Debug().
		Str("mutation", mutation).
		Msg("constructed GraphQL mutation for issue update")

	variables := map[string]interface{}{
		"id": issueID,
	}

	if title != "" {
		variables["title"] = title
		logger.Debug().Str("title", title).Msg("adding title to update variables")
	}
	if body != "" {
		variables["body"] = body
		logger.Debug().Int("body_length", len(body)).Msg("adding body to update variables")
	}

	logger.Debug().
		Interface("variables", variables).
		Msg("prepared GraphQL variables for issue update")

	var resp struct {
		UpdateIssue struct {
			Issue struct {
				ID     string `json:"id"`
				Number int    `json:"number"`
				URL    string `json:"url"`
				Title  string `json:"title"`
				Body   string `json:"body"`
				Labels struct {
					Nodes []Label `json:"nodes"`
				} `json:"labels"`
			} `json:"issue"`
		} `json:"updateIssue"`
	}

	logger.Debug().Msg("executing GraphQL mutation for issue update")
	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("failed to execute GraphQL mutation for issue update")
		return nil, errors.Wrap(err, "failed to update issue")
	}

	issue := &Issue{
		ID:     resp.UpdateIssue.Issue.ID,
		Number: resp.UpdateIssue.Issue.Number,
		URL:    resp.UpdateIssue.Issue.URL,
		Title:  resp.UpdateIssue.Issue.Title,
		Body:   resp.UpdateIssue.Issue.Body,
		Labels: resp.UpdateIssue.Issue.Labels.Nodes,
	}

	logger.Debug().
		Str("updated_issue_id", issue.ID).
		Int("updated_issue_number", issue.Number).
		Str("updated_issue_url", issue.URL).
		Str("updated_issue_title", issue.Title).
		Int("labels_count", len(issue.Labels)).
		Msg("successfully updated issue")

	return issue, nil
}

// GetRepositoryLabels retrieves labels from a repository
func (c *Client) GetRepositoryLabels(ctx context.Context, owner, name string) ([]Label, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "GetRepositoryLabels").
		Str("owner", owner).
		Str("name", name).
		Logger()

	logger.Debug().Msg("entering GetRepositoryLabels")
	defer func() {
		logger.Debug().
			Dur("duration", time.Since(start)).
			Msg("exiting GetRepositoryLabels")
	}()

	query := `
		query($owner: String!, $name: String!) {
			repository(owner: $owner, name: $name) {
				labels(first: 100) {
					nodes {
						id
						name
					}
				}
			}
		}
	`

	logger.Debug().
		Str("query", query).
		Msg("constructed GraphQL query for repository labels")

	variables := map[string]interface{}{
		"owner": owner,
		"name":  name,
	}

	logger.Debug().
		Interface("variables", variables).
		Msg("prepared GraphQL variables for labels query")

	var resp struct {
		Repository struct {
			Labels struct {
				Nodes []Label `json:"nodes"`
			} `json:"labels"`
		} `json:"repository"`
	}

	logger.Debug().Msg("executing GraphQL query for repository labels")
	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("failed to execute GraphQL query for repository labels")
		return nil, errors.Wrap(err, "failed to get repository labels")
	}

	labels := resp.Repository.Labels.Nodes
	logger.Debug().
		Int("labels_count", len(labels)).
		Interface("labels", labels).
		Msg("successfully retrieved repository labels")

	return labels, nil
}

// GetLabelIDsByNames retrieves label IDs by their names from a repository
func (c *Client) GetLabelIDsByNames(ctx context.Context, owner, name string, labelNames []string) ([]string, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "GetLabelIDsByNames").
		Str("owner", owner).
		Str("name", name).
		Strs("label_names", labelNames).
		Int("requested_labels_count", len(labelNames)).
		Logger()

	logger.Debug().Msg("entering GetLabelIDsByNames")
	defer func() {
		logger.Debug().
			Dur("duration", time.Since(start)).
			Msg("exiting GetLabelIDsByNames")
	}()

	logger.Debug().Msg("retrieving repository labels for mapping")
	labels, err := c.GetRepositoryLabels(ctx, owner, name)
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("failed to retrieve repository labels")
		return nil, err
	}

	logger.Debug().
		Int("available_labels_count", len(labels)).
		Msg("retrieved repository labels, building label map")

	labelMap := make(map[string]string)
	for _, label := range labels {
		labelMap[label.Name] = label.ID
	}

	logger.Debug().
		Int("label_map_size", len(labelMap)).
		Msg("built label name to ID mapping")

	var labelIDs []string
	var missingLabels []string

	for _, labelName := range labelNames {
		if labelID, exists := labelMap[labelName]; exists {
			labelIDs = append(labelIDs, labelID)
			logger.Debug().
				Str("label_name", labelName).
				Str("label_id", labelID).
				Msg("found label ID for name")
		} else {
			missingLabels = append(missingLabels, labelName)
			logger.Debug().
				Str("missing_label_name", labelName).
				Msg("label name not found in repository")
		}
	}

	if len(missingLabels) > 0 {
		logger.Error().
			Strs("missing_labels", missingLabels).
			Dur("duration", time.Since(start)).
			Msg("some requested labels were not found in repository")
		return nil, errors.Errorf("labels not found in repository: %v", missingLabels)
	}

	logger.Debug().
		Int("resolved_label_ids_count", len(labelIDs)).
		Strs("label_ids", labelIDs).
		Msg("successfully resolved all label names to IDs")

	return labelIDs, nil
}

// UpdateIssueComment updates an existing issue comment
func (c *Client) UpdateIssueComment(ctx context.Context, commentID, body string) (*Comment, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "UpdateIssueComment").
		Str("comment_id", commentID).
		Int("new_body_length", len(body)).
		Logger()

	logger.Debug().Msg("entering UpdateIssueComment")
	defer func() {
		logger.Debug().
			Dur("duration", time.Since(start)).
			Msg("exiting UpdateIssueComment")
	}()

	mutation := `
		mutation($id: ID!, $body: String!) {
			updateIssueComment(input: { id: $id, body: $body }) {
				issueComment {
					id
					body
					url
					createdAt
					updatedAt
					author {
						login
						avatarUrl
					}
				}
			}
		}
	`

	logger.Debug().
		Str("mutation", mutation).
		Msg("constructed GraphQL mutation for comment update")

	variables := map[string]interface{}{
		"id":   commentID,
		"body": body,
	}

	logger.Debug().
		Interface("variables", variables).
		Msg("prepared GraphQL variables for comment update")

	var resp struct {
		UpdateIssueComment struct {
			IssueComment struct {
				ID        string `json:"id"`
				Body      string `json:"body"`
				URL       string `json:"url"`
				CreatedAt string `json:"createdAt"`
				UpdatedAt string `json:"updatedAt"`
				Author    struct {
					Login     string `json:"login"`
					AvatarURL string `json:"avatarUrl"`
				} `json:"author"`
			} `json:"issueComment"`
		} `json:"updateIssueComment"`
	}

	logger.Debug().Msg("executing GraphQL mutation for comment update")
	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("failed to execute GraphQL mutation for comment update")
		return nil, errors.Wrap(err, "failed to update issue comment")
	}

	// Parse timestamps
	createdAt, err := time.Parse(time.RFC3339, resp.UpdateIssueComment.IssueComment.CreatedAt)
	if err != nil {
		logger.Error().
			Err(err).
			Str("createdAt", resp.UpdateIssueComment.IssueComment.CreatedAt).
			Msg("failed to parse createdAt timestamp")
		createdAt = time.Time{}
	}

	updatedAt, err := time.Parse(time.RFC3339, resp.UpdateIssueComment.IssueComment.UpdatedAt)
	if err != nil {
		logger.Error().
			Err(err).
			Str("updatedAt", resp.UpdateIssueComment.IssueComment.UpdatedAt).
			Msg("failed to parse updatedAt timestamp")
		updatedAt = time.Time{}
	}

	comment := &Comment{
		ID:        resp.UpdateIssueComment.IssueComment.ID,
		Body:      resp.UpdateIssueComment.IssueComment.Body,
		URL:       resp.UpdateIssueComment.IssueComment.URL,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Author: Author{
			Login:     resp.UpdateIssueComment.IssueComment.Author.Login,
			AvatarURL: resp.UpdateIssueComment.IssueComment.Author.AvatarURL,
		},
	}

	logger.Debug().
		Str("updated_comment_id", comment.ID).
		Str("updated_comment_url", comment.URL).
		Int("updated_body_length", len(comment.Body)).
		Msg("successfully updated issue comment")

	return comment, nil
}

// AddLabelsToLabelable adds labels to an issue or pull request
func (c *Client) AddLabelsToLabelable(ctx context.Context, labelableID string, labelIDs []string) error {
	start := time.Now()
	logger := log.With().
		Str("function", "AddLabelsToLabelable").
		Str("labelable_id", labelableID).
		Strs("label_ids", labelIDs).
		Int("labels_to_add_count", len(labelIDs)).
		Logger()

	logger.Debug().Msg("entering AddLabelsToLabelable")
	defer func() {
		logger.Debug().
			Dur("duration", time.Since(start)).
			Msg("exiting AddLabelsToLabelable")
	}()

	mutation := `
		mutation($labelableId: ID!, $labelIds: [ID!]!) {
			addLabelsToLabelable(input: { labelableId: $labelableId, labelIds: $labelIds }) {
				labelable {
					... on Issue {
						id
						labels(first: 100) {
							nodes {
								id
								name
							}
						}
					}
				}
			}
		}
	`

	logger.Debug().
		Str("mutation", mutation).
		Msg("constructed GraphQL mutation for adding labels to labelable")

	variables := map[string]interface{}{
		"labelableId": labelableID,
		"labelIds":    labelIDs,
	}

	logger.Debug().
		Interface("variables", variables).
		Msg("prepared GraphQL variables for adding labels")

	var resp struct {
		AddLabelsToLabelable struct {
			Labelable struct {
				ID     string `json:"id"`
				Labels struct {
					Nodes []Label `json:"nodes"`
				} `json:"labels"`
			} `json:"labelable"`
		} `json:"addLabelsToLabelable"`
	}

	logger.Debug().Msg("executing GraphQL mutation to add labels to labelable")
	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("failed to execute GraphQL mutation for adding labels")
		return errors.Wrap(err, "failed to add labels to labelable")
	}

	logger.Debug().
		Str("updated_labelable_id", resp.AddLabelsToLabelable.Labelable.ID).
		Int("total_labels_count", len(resp.AddLabelsToLabelable.Labelable.Labels.Nodes)).
		Interface("current_labels", resp.AddLabelsToLabelable.Labelable.Labels.Nodes).
		Msg("successfully added labels to labelable")

	return nil
}

// RemoveLabelsFromLabelable removes labels from an issue or pull request
func (c *Client) RemoveLabelsFromLabelable(ctx context.Context, labelableID string, labelIDs []string) error {
	start := time.Now()
	logger := log.With().
		Str("function", "RemoveLabelsFromLabelable").
		Str("labelable_id", labelableID).
		Strs("label_ids", labelIDs).
		Int("labels_to_remove_count", len(labelIDs)).
		Logger()

	logger.Debug().Msg("entering RemoveLabelsFromLabelable")
	defer func() {
		logger.Debug().
			Dur("duration", time.Since(start)).
			Msg("exiting RemoveLabelsFromLabelable")
	}()

	mutation := `
		mutation($labelableId: ID!, $labelIds: [ID!]!) {
			removeLabelsFromLabelable(input: { labelableId: $labelableId, labelIds: $labelIds }) {
				labelable {
					... on Issue {
						id
						labels(first: 100) {
							nodes {
								id
								name
							}
						}
					}
				}
			}
		}
	`

	logger.Debug().
		Str("mutation", mutation).
		Msg("constructed GraphQL mutation for removing labels from labelable")

	variables := map[string]interface{}{
		"labelableId": labelableID,
		"labelIds":    labelIDs,
	}

	logger.Debug().
		Interface("variables", variables).
		Msg("prepared GraphQL variables for removing labels")

	var resp struct {
		RemoveLabelsFromLabelable struct {
			Labelable struct {
				ID     string `json:"id"`
				Labels struct {
					Nodes []Label `json:"nodes"`
				} `json:"labels"`
			} `json:"labelable"`
		} `json:"removeLabelsFromLabelable"`
	}

	logger.Debug().Msg("executing GraphQL mutation to remove labels from labelable")
	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("failed to execute GraphQL mutation for removing labels")
		return errors.Wrap(err, "failed to remove labels from labelable")
	}

	logger.Debug().
		Str("updated_labelable_id", resp.RemoveLabelsFromLabelable.Labelable.ID).
		Int("remaining_labels_count", len(resp.RemoveLabelsFromLabelable.Labelable.Labels.Nodes)).
		Interface("remaining_labels", resp.RemoveLabelsFromLabelable.Labelable.Labels.Nodes).
		Msg("successfully removed labels from labelable")

	return nil
}

// GetIssueComments retrieves all comments for an issue or pull request by node ID
func (c *Client) GetIssueComments(ctx context.Context, issueID string) ([]Comment, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "GetIssueComments").
		Str("issue_id", issueID).
		Logger()

	logger.Debug().Msg("entering GetIssueComments")
	defer func() {
		logger.Debug().
			Dur("duration", time.Since(start)).
			Msg("exiting GetIssueComments")
	}()

	query := `
		query($id: ID!) {
			node(id: $id) {
				... on Issue {
					comments(first: 100) {
						nodes {
							id
							body
							url
							createdAt
							updatedAt
							author {
								login
								avatarUrl
							}
						}
					}
				}
				... on PullRequest {
					comments(first: 100) {
						nodes {
							id
							body
							url
							createdAt
							updatedAt
							author {
								login
								avatarUrl
							}
						}
					}
				}
			}
		}
	`

	logger.Debug().
		Str("query", query).
		Msg("constructed GraphQL query for issue comments")

	variables := map[string]interface{}{
		"id": issueID,
	}

	logger.Debug().
		Interface("variables", variables).
		Msg("prepared GraphQL variables for comments query")

	var resp struct {
		Node struct {
			Comments struct {
				Nodes []struct {
					ID        string `json:"id"`
					Body      string `json:"body"`
					URL       string `json:"url"`
					CreatedAt string `json:"createdAt"`
					UpdatedAt string `json:"updatedAt"`
					Author    struct {
						Login     string `json:"login"`
						AvatarURL string `json:"avatarUrl"`
					} `json:"author"`
				} `json:"nodes"`
			} `json:"comments"`
		} `json:"node"`
	}

	logger.Debug().Msg("executing GraphQL query for issue comments")
	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(start)).
			Msg("failed to execute GraphQL query for issue comments")
		return nil, errors.Wrap(err, "failed to get issue comments")
	}

	comments := make([]Comment, 0, len(resp.Node.Comments.Nodes))
	for _, comment := range resp.Node.Comments.Nodes {
		// Parse timestamps
		createdAt, err := time.Parse(time.RFC3339, comment.CreatedAt)
		if err != nil {
			logger.Error().
				Err(err).
				Str("createdAt", comment.CreatedAt).
				Msg("failed to parse createdAt timestamp")
			createdAt = time.Time{}
		}

		updatedAt, err := time.Parse(time.RFC3339, comment.UpdatedAt)
		if err != nil {
			logger.Error().
				Err(err).
				Str("updatedAt", comment.UpdatedAt).
				Msg("failed to parse updatedAt timestamp")
			updatedAt = time.Time{}
		}

		comments = append(comments, Comment{
			ID:        comment.ID,
			Body:      comment.Body,
			URL:       comment.URL,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			Author: Author{
				Login:     comment.Author.Login,
				AvatarURL: comment.Author.AvatarURL,
			},
		})
	}

	logger.Debug().
		Int("comments_count", len(comments)).
		Msg("successfully retrieved issue comments")

	return comments, nil
}
