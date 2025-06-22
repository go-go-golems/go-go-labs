package github

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
	startTime := time.Now()
	logger := log.With().
		Str("function", "ListUserProjects").
		Int("first", first).
		Interface("after", after).
		Logger()

	logger.Debug().Msg("entering ListUserProjects")
	defer func() {
		logger.Debug().
			Dur("duration", time.Since(startTime)).
			Msg("exiting ListUserProjects")
	}()

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

	logger.Debug().
		Str("query", query).
		Msg("constructed GraphQL query for user projects")

	variables := map[string]interface{}{
		"first": first,
	}
	if after != nil {
		variables["after"] = *after
		logger.Debug().
			Str("after_cursor", *after).
			Msg("pagination cursor provided")
	} else {
		logger.Debug().Msg("no pagination cursor - fetching first page")
	}

	logger.Debug().
		Interface("variables", variables).
		Msg("prepared GraphQL variables")

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

	logger.Debug().Msg("executing GraphQL query for user projects")
	queryStartTime := time.Now()

	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		logger.Error().
			Err(err).
			Dur("query_duration", time.Since(queryStartTime)).
			Msg("GraphQL query execution failed")
		return nil, errors.Wrap(err, "failed to list user projects")
	}

	queryDuration := time.Since(queryStartTime)
	logger.Debug().
		Dur("query_duration", queryDuration).
		Int("projects_count", len(resp.Viewer.ProjectsV2.Nodes)).
		Bool("has_next_page", resp.Viewer.ProjectsV2.PageInfo.HasNextPage).
		Interface("end_cursor", resp.Viewer.ProjectsV2.PageInfo.EndCursor).
		Msg("GraphQL query executed successfully")

	result := &ListProjectsResponse{
		Projects:    resp.Viewer.ProjectsV2.Nodes,
		HasNextPage: resp.Viewer.ProjectsV2.PageInfo.HasNextPage,
		EndCursor:   resp.Viewer.ProjectsV2.PageInfo.EndCursor,
	}

	logger.Debug().
		Int("result_projects_count", len(result.Projects)).
		Bool("result_has_next_page", result.HasNextPage).
		Msg("constructed response")

	for i, project := range result.Projects {
		logger.Debug().
			Int("index", i).
			Str("project_id", project.ID).
			Int("project_number", project.Number).
			Str("project_title", project.Title).
			Bool("project_public", project.Public).
			Bool("project_closed", project.Closed).
			Str("project_url", project.URL).
			Msg("project details")
	}

	return result, nil
}

// ListOrganizationProjects lists projects for an organization
func (c *Client) ListOrganizationProjects(ctx context.Context, owner string, first int, after *string) (*ListProjectsResponse, error) {
	startTime := time.Now()
	logger := log.With().
		Str("function", "ListOrganizationProjects").
		Str("owner", owner).
		Int("first", first).
		Interface("after", after).
		Logger()

	logger.Debug().Msg("entering ListOrganizationProjects")
	defer func() {
		logger.Debug().
			Dur("duration", time.Since(startTime)).
			Msg("exiting ListOrganizationProjects")
	}()

	logger.Debug().
		Str("organization_login", owner).
		Msg("processing organization parameter")

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

	logger.Debug().
		Str("query", query).
		Msg("constructed GraphQL query for organization projects")

	variables := map[string]interface{}{
		"owner": owner,
		"first": first,
	}
	if after != nil {
		variables["after"] = *after
		logger.Debug().
			Str("after_cursor", *after).
			Msg("pagination cursor provided")
	} else {
		logger.Debug().Msg("no pagination cursor - fetching first page")
	}

	logger.Debug().
		Interface("variables", variables).
		Msg("prepared GraphQL variables")

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

	logger.Debug().Msg("executing GraphQL query for organization projects")
	queryStartTime := time.Now()

	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		logger.Error().
			Err(err).
			Dur("query_duration", time.Since(queryStartTime)).
			Str("organization", owner).
			Msg("GraphQL query execution failed")
		return nil, errors.Wrap(err, "failed to list organization projects")
	}

	queryDuration := time.Since(queryStartTime)
	logger.Debug().
		Dur("query_duration", queryDuration).
		Int("projects_count", len(resp.Organization.ProjectsV2.Nodes)).
		Bool("has_next_page", resp.Organization.ProjectsV2.PageInfo.HasNextPage).
		Interface("end_cursor", resp.Organization.ProjectsV2.PageInfo.EndCursor).
		Str("organization", owner).
		Msg("GraphQL query executed successfully")

	result := &ListProjectsResponse{
		Projects:    resp.Organization.ProjectsV2.Nodes,
		HasNextPage: resp.Organization.ProjectsV2.PageInfo.HasNextPage,
		EndCursor:   resp.Organization.ProjectsV2.PageInfo.EndCursor,
	}

	logger.Debug().
		Int("result_projects_count", len(result.Projects)).
		Bool("result_has_next_page", result.HasNextPage).
		Str("organization", owner).
		Msg("constructed response")

	for i, project := range result.Projects {
		logger.Debug().
			Int("index", i).
			Str("project_id", project.ID).
			Int("project_number", project.Number).
			Str("project_title", project.Title).
			Bool("project_public", project.Public).
			Bool("project_closed", project.Closed).
			Str("project_url", project.URL).
			Str("organization", owner).
			Msg("project details")
	}

	return result, nil
}
