package github

import (
	"context"

	"github.com/pkg/errors"
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

// Repository represents a GitHub repository
type Repository struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetRepository retrieves a repository by owner and name
func (c *Client) GetRepository(ctx context.Context, owner, name string) (*Repository, error) {
	query := `
		query($owner: String!, $name: String!) {
			repository(owner: $owner, name: $name) {
				id
				name
			}
		}
	`

	variables := map[string]interface{}{
		"owner": owner,
		"name":  name,
	}

	var resp struct {
		Repository Repository `json:"repository"`
	}

	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		return nil, errors.Wrap(err, "failed to get repository")
	}

	return &resp.Repository, nil
}

// GetIssue retrieves an issue by repository and number
func (c *Client) GetIssue(ctx context.Context, owner, name string, number int) (*Issue, error) {
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

	variables := map[string]interface{}{
		"owner":  owner,
		"name":   name,
		"number": number,
	}

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

	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		return nil, errors.Wrap(err, "failed to get issue")
	}

	return &Issue{
		ID:     resp.Repository.Issue.ID,
		Number: resp.Repository.Issue.Number,
		URL:    resp.Repository.Issue.URL,
		Title:  resp.Repository.Issue.Title,
		Body:   resp.Repository.Issue.Body,
		Labels: resp.Repository.Issue.Labels.Nodes,
	}, nil
}

// CreateIssue creates a new issue in a repository
func (c *Client) CreateIssue(ctx context.Context, repositoryID, title, body string) (*Issue, error) {
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

	variables := map[string]interface{}{
		"repositoryId": repositoryID,
		"title":        title,
		"body":         body,
	}

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

	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		return nil, errors.Wrap(err, "failed to create issue")
	}

	return &Issue{
		ID:     resp.CreateIssue.Issue.ID,
		Number: resp.CreateIssue.Issue.Number,
		URL:    resp.CreateIssue.Issue.URL,
		Title:  resp.CreateIssue.Issue.Title,
		Body:   resp.CreateIssue.Issue.Body,
		Labels: resp.CreateIssue.Issue.Labels.Nodes,
	}, nil
}

// CreateIssueWithLabels creates a new issue in a repository with labels
func (c *Client) CreateIssueWithLabels(ctx context.Context, repositoryID, title, body string, labelIDs []string) (*Issue, error) {
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

	variables := map[string]interface{}{
		"repositoryId": repositoryID,
		"title":        title,
		"body":         body,
		"labelIds":     labelIDs,
	}

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

	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		return nil, errors.Wrap(err, "failed to create issue with labels")
	}

	return &Issue{
		ID:     resp.CreateIssue.Issue.ID,
		Number: resp.CreateIssue.Issue.Number,
		URL:    resp.CreateIssue.Issue.URL,
		Title:  resp.CreateIssue.Issue.Title,
		Body:   resp.CreateIssue.Issue.Body,
		Labels: resp.CreateIssue.Issue.Labels.Nodes,
	}, nil
}

// UpdateIssue updates an existing issue
func (c *Client) UpdateIssue(ctx context.Context, issueID, title, body string) (*Issue, error) {
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

	variables := map[string]interface{}{
		"id": issueID,
	}

	if title != "" {
		variables["title"] = title
	}
	if body != "" {
		variables["body"] = body
	}

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

	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		return nil, errors.Wrap(err, "failed to update issue")
	}

	return &Issue{
		ID:     resp.UpdateIssue.Issue.ID,
		Number: resp.UpdateIssue.Issue.Number,
		URL:    resp.UpdateIssue.Issue.URL,
		Title:  resp.UpdateIssue.Issue.Title,
		Body:   resp.UpdateIssue.Issue.Body,
		Labels: resp.UpdateIssue.Issue.Labels.Nodes,
	}, nil
}

// GetRepositoryLabels retrieves labels from a repository
func (c *Client) GetRepositoryLabels(ctx context.Context, owner, name string) ([]Label, error) {
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

	variables := map[string]interface{}{
		"owner": owner,
		"name":  name,
	}

	var resp struct {
		Repository struct {
			Labels struct {
				Nodes []Label `json:"nodes"`
			} `json:"labels"`
		} `json:"repository"`
	}

	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		return nil, errors.Wrap(err, "failed to get repository labels")
	}

	return resp.Repository.Labels.Nodes, nil
}

// GetLabelIDsByNames retrieves label IDs by their names from a repository
func (c *Client) GetLabelIDsByNames(ctx context.Context, owner, name string, labelNames []string) ([]string, error) {
	labels, err := c.GetRepositoryLabels(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	labelMap := make(map[string]string)
	for _, label := range labels {
		labelMap[label.Name] = label.ID
	}

	var labelIDs []string
	var missingLabels []string

	for _, labelName := range labelNames {
		if labelID, exists := labelMap[labelName]; exists {
			labelIDs = append(labelIDs, labelID)
		} else {
			missingLabels = append(missingLabels, labelName)
		}
	}

	if len(missingLabels) > 0 {
		return nil, errors.Errorf("labels not found in repository: %v", missingLabels)
	}

	return labelIDs, nil
}

// AddLabelsToLabelable adds labels to an issue or pull request
func (c *Client) AddLabelsToLabelable(ctx context.Context, labelableID string, labelIDs []string) error {
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

	variables := map[string]interface{}{
		"labelableId": labelableID,
		"labelIds":    labelIDs,
	}

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

	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		return errors.Wrap(err, "failed to add labels to labelable")
	}

	return nil
}

// RemoveLabelsFromLabelable removes labels from an issue or pull request
func (c *Client) RemoveLabelsFromLabelable(ctx context.Context, labelableID string, labelIDs []string) error {
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

	variables := map[string]interface{}{
		"labelableId": labelableID,
		"labelIds":    labelIDs,
	}

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

	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		return errors.Wrap(err, "failed to remove labels from labelable")
	}

	return nil
}
