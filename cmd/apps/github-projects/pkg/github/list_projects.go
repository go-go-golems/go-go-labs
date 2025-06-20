package github

import (
	"context"

	"github.com/pkg/errors"
)

// ProjectSummary represents a project summary for listing
type ProjectSummary struct {
	ID               string `json:"id"`
	Number           int    `json:"number"`
	Title            string `json:"title"`
	Public           bool   `json:"public"`
	Closed           bool   `json:"closed"`
	ShortDescription string `json:"shortDescription"`
	URL              string `json:"url"`
}

// ListProjectsResponse represents the response from listing projects
type ListProjectsResponse struct {
	Projects    []ProjectSummary `json:"projects"`
	HasNextPage bool             `json:"hasNextPage"`
	EndCursor   *string          `json:"endCursor"`
}

// ListUserProjects lists projects for the authenticated user
func (c *Client) ListUserProjects(ctx context.Context, first int, after *string) (*ListProjectsResponse, error) {
	query := `
		query($first: Int!, $after: String) {
			viewer {
				projectsV2(first: $first, after: $after) {
					nodes {
						id
						number
						title
						public
						closed
						shortDescription
						url
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"first": first,
	}
	if after != nil {
		variables["after"] = *after
	}

	var resp struct {
		Viewer struct {
			ProjectsV2 struct {
				Nodes    []ProjectSummary `json:"nodes"`
				PageInfo struct {
					HasNextPage bool    `json:"hasNextPage"`
					EndCursor   *string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"projectsV2"`
		} `json:"viewer"`
	}

	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		return nil, errors.Wrap(err, "failed to list user projects")
	}

	return &ListProjectsResponse{
		Projects:    resp.Viewer.ProjectsV2.Nodes,
		HasNextPage: resp.Viewer.ProjectsV2.PageInfo.HasNextPage,
		EndCursor:   resp.Viewer.ProjectsV2.PageInfo.EndCursor,
	}, nil
}

// ListOrganizationProjects lists projects for an organization
func (c *Client) ListOrganizationProjects(ctx context.Context, owner string, first int, after *string) (*ListProjectsResponse, error) {
	query := `
		query($owner: String!, $first: Int!, $after: String) {
			organization(login: $owner) {
				projectsV2(first: $first, after: $after) {
					nodes {
						id
						number
						title
						public
						closed
						shortDescription
						url
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"owner": owner,
		"first": first,
	}
	if after != nil {
		variables["after"] = *after
	}

	var resp struct {
		Organization struct {
			ProjectsV2 struct {
				Nodes    []ProjectSummary `json:"nodes"`
				PageInfo struct {
					HasNextPage bool    `json:"hasNextPage"`
					EndCursor   *string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"projectsV2"`
		} `json:"organization"`
	}

	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		return nil, errors.Wrap(err, "failed to list organization projects")
	}

	return &ListProjectsResponse{
		Projects:    resp.Organization.ProjectsV2.Nodes,
		HasNextPage: resp.Organization.ProjectsV2.PageInfo.HasNextPage,
		EndCursor:   resp.Organization.ProjectsV2.PageInfo.EndCursor,
	}, nil
}
