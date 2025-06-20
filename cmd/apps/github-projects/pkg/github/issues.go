package github

import (
	"context"

	"github.com/pkg/errors"
)

// Issue represents a GitHub issue
type Issue struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	Body   string `json:"body"`
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
			Issue Issue `json:"issue"`
		} `json:"repository"`
	}

	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		return nil, errors.Wrap(err, "failed to get issue")
	}

	return &resp.Repository.Issue, nil
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
			Issue Issue `json:"issue"`
		} `json:"createIssue"`
	}

	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		return nil, errors.Wrap(err, "failed to create issue")
	}

	return &resp.CreateIssue.Issue, nil
}
