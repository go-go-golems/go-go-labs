package github

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Project represents a GitHub Project v2
type Project struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Public           bool   `json:"public"`
	ShortDescription string `json:"shortDescription"`
	Closed           bool   `json:"closed"`
	Items            struct {
		TotalCount int `json:"totalCount"`
	} `json:"items"`
}

// ProjectField represents a project field
type ProjectField struct {
	Typename      string                  `json:"__typename"`
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	Options       []FieldOption           `json:"options,omitempty"`
	Configuration *IterationConfiguration `json:"configuration,omitempty"`
}

// FieldOption represents a single-select field option
type FieldOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// IterationConfiguration represents iteration field configuration
type IterationConfiguration struct {
	Iterations []Iteration `json:"iterations"`
}

// Iteration represents an iteration
type Iteration struct {
	ID        string `json:"id"`
	StartDate string `json:"startDate"`
	Title     string `json:"title"`
}

// ProjectItem represents a project item
type ProjectItem struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Content     ItemContent `json:"content"`
	FieldValues struct {
		Nodes []FieldValue `json:"nodes"`
	} `json:"fieldValues"`
}

// ItemContent represents the content of a project item
type ItemContent struct {
	Typename  string `json:"__typename"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Number    int    `json:"number"`
	Body      string `json:"body,omitempty"`
	Assignees struct {
		Nodes []struct {
			Login string `json:"login"`
		} `json:"nodes"`
	} `json:"assignees"`
}

// FieldValue represents a field value
type FieldValue struct {
	Typename  string   `json:"__typename"`
	Text      *string  `json:"text,omitempty"`
	Number    *float64 `json:"number,omitempty"`
	Date      *string  `json:"date,omitempty"`
	Name      *string  `json:"name,omitempty"`
	StartDate *string  `json:"startDate,omitempty"`
	Title     *string  `json:"title,omitempty"`
	Field     struct {
		Name string `json:"name"`
	} `json:"field"`
}

// GetProject retrieves a project by owner and number
func (c *Client) GetProject(ctx context.Context, owner string, number int) (*Project, error) {
	query := `
		query($owner: String!, $number: Int!) {
			organization(login: $owner) {
				projectV2(number: $number) {
					id
					title
					public
					shortDescription
					closed
					items(first: 0) { totalCount }
				}
			}
		}
	`

	variables := map[string]interface{}{
		"owner":  owner,
		"number": number,
	}

	var resp struct {
		Organization struct {
			ProjectV2 Project `json:"projectV2"`
		} `json:"organization"`
	}

	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		return nil, errors.Wrap(err, "failed to get project")
	}

	return &resp.Organization.ProjectV2, nil
}

// GetProjectFields retrieves fields for a project
func (c *Client) GetProjectFields(ctx context.Context, projectID string) ([]ProjectField, error) {
	query := `
		query($projectId: ID!) {
			node(id: $projectId) {
				... on ProjectV2 {
					fields(first: 20) {
						nodes {
							... on ProjectV2Field {
								id
								name
							}
							... on ProjectV2SingleSelectField {
								id
								name
								options {
									id
									name
								}
							}
							... on ProjectV2IterationField {
								id
								name
								configuration {
									iterations {
										id
										startDate
										title
									}
								}
							}
						}
					}
				}
			}
		}
	`

	log.Info().Str("projectID", projectID).Msg("GetProjectFields called")
	log.Info().Str("query", query).Msg("GraphQL query being executed")
	log.Info().Msg("About to call ExecuteQuery")

	variables := map[string]interface{}{
		"projectId": projectID,
	}

	var resp struct {
		Node struct {
			Fields struct {
				Nodes []ProjectField `json:"nodes"`
			} `json:"fields"`
		} `json:"node"`
	}

	log.Info().Interface("variables", variables).Msg("About to execute query with variables")
	
	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		log.Error().Err(err).Msg("ExecuteQuery failed")
		return nil, errors.Wrap(err, "failed to get project fields")
	}

	log.Info().Interface("response", resp).Msg("Query executed successfully")
	log.Info().Int("fieldCount", len(resp.Node.Fields.Nodes)).Msg("Number of fields returned")

	return resp.Node.Fields.Nodes, nil
}

// GetProjectItems retrieves items for a project
func (c *Client) GetProjectItems(ctx context.Context, projectID string, first int) ([]ProjectItem, error) {
	query := `
		query($projectId: ID!, $first: Int!) {
			node(id: $projectId) {
				... on ProjectV2 {
					items(first: $first) {
						nodes {
							id
							type
							content {
								__typename
								... on Issue {
									title
									number
									url
									assignees(first: 5) { nodes { login } }
								}
								... on PullRequest {
									title
									number
									url
									assignees(first: 5) { nodes { login } }
								}
								... on DraftIssue {
									title
									body
								}
							}
							fieldValues(first: 10) {
								nodes {
									__typename
									... on ProjectV2ItemFieldTextValue {
										text
										field { ... on ProjectV2FieldCommon { name } }
									}
									... on ProjectV2ItemFieldNumberValue {
										number
										field { ... on ProjectV2FieldCommon { name } }
									}
									... on ProjectV2ItemFieldDateValue {
										date
										field { ... on ProjectV2FieldCommon { name } }
									}
									... on ProjectV2ItemFieldSingleSelectValue {
										name
										field { ... on ProjectV2FieldCommon { name } }
									}
									... on ProjectV2ItemFieldIterationValue {
										title
										startDate
										field { ... on ProjectV2FieldCommon { name } }
									}
								}
							}
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"projectId": projectID,
		"first":     first,
	}

	var resp struct {
		Node struct {
			Items struct {
				Nodes []ProjectItem `json:"nodes"`
			} `json:"items"`
		} `json:"node"`
	}

	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		return nil, errors.Wrap(err, "failed to get project items")
	}

	return resp.Node.Items.Nodes, nil
}

// AddItemToProject adds an existing issue or PR to a project
func (c *Client) AddItemToProject(ctx context.Context, projectID, contentID string) (string, error) {
	mutation := `
		mutation($projectId: ID!, $contentId: ID!) {
			addProjectV2ItemById(input: { projectId: $projectId, contentId: $contentId }) {
				item {
					id
				}
			}
		}
	`

	variables := map[string]interface{}{
		"projectId": projectID,
		"contentId": contentID,
	}

	var resp struct {
		AddProjectV2ItemById struct {
			Item struct {
				ID string `json:"id"`
			} `json:"item"`
		} `json:"addProjectV2ItemById"`
	}

	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		return "", errors.Wrap(err, "failed to add item to project")
	}

	return resp.AddProjectV2ItemById.Item.ID, nil
}

// CreateDraftIssue creates a draft issue in a project
func (c *Client) CreateDraftIssue(ctx context.Context, projectID, title, body string) (string, error) {
	mutation := `
		mutation($projectId: ID!, $title: String!, $body: String) {
			addProjectV2DraftIssue(input: { projectId: $projectId, title: $title, body: $body }) {
				projectItem {
					id
				}
			}
		}
	`

	variables := map[string]interface{}{
		"projectId": projectID,
		"title":     title,
		"body":      body,
	}

	var resp struct {
		AddProjectV2DraftIssue struct {
			ProjectItem struct {
				ID string `json:"id"`
			} `json:"projectItem"`
		} `json:"addProjectV2DraftIssue"`
	}

	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		return "", errors.Wrap(err, "failed to create draft issue")
	}

	return resp.AddProjectV2DraftIssue.ProjectItem.ID, nil
}

// UpdateFieldValue updates a field value for a project item
func (c *Client) UpdateFieldValue(ctx context.Context, projectID, itemID, fieldID string, value interface{}) error {
	mutation := `
		mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $value: ProjectV2FieldValue!) {
			updateProjectV2ItemFieldValue(input: {
				projectId: $projectId
				itemId: $itemId
				fieldId: $fieldId
				value: $value
			}) {
				projectV2Item { id }
			}
		}
	`

	variables := map[string]interface{}{
		"projectId": projectID,
		"itemId":    itemID,
		"fieldId":   fieldID,
		"value":     value,
	}

	var resp struct {
		UpdateProjectV2ItemFieldValue struct {
			ProjectV2Item struct {
				ID string `json:"id"`
			} `json:"projectV2Item"`
		} `json:"updateProjectV2ItemFieldValue"`
	}

	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		return errors.Wrap(err, "failed to update field value")
	}

	return nil
}
