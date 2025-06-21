package github

import (
	"context"
	"time"

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
	start := time.Now()
	log.Debug().
		Str("function", "GetProject").
		Str("owner", owner).
		Int("number", number).
		Msg("entering function")

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

	log.Debug().
		Str("function", "GetProject").
		Str("query", query).
		Msg("constructed GraphQL query")

	variables := map[string]interface{}{
		"owner":  owner,
		"number": number,
	}

	log.Debug().
		Str("function", "GetProject").
		Interface("variables", variables).
		Msg("constructed query variables")

	var resp struct {
		Organization struct {
			ProjectV2 Project `json:"projectV2"`
		} `json:"organization"`
	}

	log.Debug().
		Str("function", "GetProject").
		Msg("executing GraphQL query")

	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		log.Error().
			Str("function", "GetProject").
			Err(err).
			Str("owner", owner).
			Int("number", number).
			Dur("duration", time.Since(start)).
			Msg("query execution failed")
		return nil, errors.Wrap(err, "failed to get project")
	}

	log.Debug().
		Str("function", "GetProject").
		Interface("response", resp).
		Msg("received GraphQL response")

	log.Debug().
		Str("function", "GetProject").
		Str("projectID", resp.Organization.ProjectV2.ID).
		Str("projectTitle", resp.Organization.ProjectV2.Title).
		Bool("public", resp.Organization.ProjectV2.Public).
		Bool("closed", resp.Organization.ProjectV2.Closed).
		Int("itemCount", resp.Organization.ProjectV2.Items.TotalCount).
		Dur("duration", time.Since(start)).
		Msg("project data processed successfully")

	log.Debug().
		Str("function", "GetProject").
		Dur("duration", time.Since(start)).
		Msg("exiting function")

	return &resp.Organization.ProjectV2, nil
}

// GetProjectFields retrieves fields for a project
func (c *Client) GetProjectFields(ctx context.Context, projectID string) ([]ProjectField, error) {
	start := time.Now()
	log.Debug().
		Str("function", "GetProjectFields").
		Str("projectID", projectID).
		Msg("entering function")

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

	log.Debug().
		Str("function", "GetProjectFields").
		Str("query", query).
		Msg("constructed GraphQL query")

	variables := map[string]interface{}{
		"projectId": projectID,
	}

	log.Debug().
		Str("function", "GetProjectFields").
		Interface("variables", variables).
		Msg("constructed query variables")

	var resp struct {
		Node struct {
			Fields struct {
				Nodes []ProjectField `json:"nodes"`
			} `json:"fields"`
		} `json:"node"`
	}

	log.Debug().
		Str("function", "GetProjectFields").
		Msg("executing GraphQL query")

	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		log.Error().
			Str("function", "GetProjectFields").
			Err(err).
			Str("projectID", projectID).
			Dur("duration", time.Since(start)).
			Msg("query execution failed")
		return nil, errors.Wrap(err, "failed to get project fields")
	}

	log.Debug().
		Str("function", "GetProjectFields").
		Interface("response", resp).
		Msg("received GraphQL response")

	fields := resp.Node.Fields.Nodes
	log.Debug().
		Str("function", "GetProjectFields").
		Int("fieldCount", len(fields)).
		Msg("processing field data")

	// Log individual field processing
	for i, field := range fields {
		log.Debug().
			Str("function", "GetProjectFields").
			Int("fieldIndex", i).
			Str("fieldID", field.ID).
			Str("fieldName", field.Name).
			Str("fieldType", field.Typename).
			Int("optionsCount", len(field.Options)).
			Msg("processing field")

		if field.Configuration != nil && len(field.Configuration.Iterations) > 0 {
			log.Debug().
				Str("function", "GetProjectFields").
				Str("fieldID", field.ID).
				Int("iterationsCount", len(field.Configuration.Iterations)).
				Msg("field has iteration configuration")
		}
	}

	log.Debug().
		Str("function", "GetProjectFields").
		Int("fieldCount", len(fields)).
		Dur("duration", time.Since(start)).
		Msg("field data processed successfully")

	log.Debug().
		Str("function", "GetProjectFields").
		Dur("duration", time.Since(start)).
		Msg("exiting function")

	return fields, nil
}

// GetProjectItems retrieves items for a project
func (c *Client) GetProjectItems(ctx context.Context, projectID string, first int) ([]ProjectItem, error) {
	start := time.Now()
	log.Debug().
		Str("function", "GetProjectItems").
		Str("projectID", projectID).
		Int("first", first).
		Msg("entering function")

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

	log.Debug().
		Str("function", "GetProjectItems").
		Str("query", query).
		Msg("constructed GraphQL query")

	variables := map[string]interface{}{
		"projectId": projectID,
		"first":     first,
	}

	log.Debug().
		Str("function", "GetProjectItems").
		Interface("variables", variables).
		Msg("constructed query variables")

	var resp struct {
		Node struct {
			Items struct {
				Nodes []ProjectItem `json:"nodes"`
			} `json:"items"`
		} `json:"node"`
	}

	log.Debug().
		Str("function", "GetProjectItems").
		Msg("executing GraphQL query")

	if err := c.ExecuteQuery(ctx, query, variables, &resp); err != nil {
		log.Error().
			Str("function", "GetProjectItems").
			Err(err).
			Str("projectID", projectID).
			Int("first", first).
			Dur("duration", time.Since(start)).
			Msg("query execution failed")
		return nil, errors.Wrap(err, "failed to get project items")
	}

	log.Debug().
		Str("function", "GetProjectItems").
		Interface("response", resp).
		Msg("received GraphQL response")

	items := resp.Node.Items.Nodes
	log.Debug().
		Str("function", "GetProjectItems").
		Int("itemCount", len(items)).
		Msg("processing item data")

	// Log individual item processing
	for i, item := range items {
		log.Debug().
			Str("function", "GetProjectItems").
			Int("itemIndex", i).
			Str("itemID", item.ID).
			Str("itemType", item.Type).
			Str("contentType", item.Content.Typename).
			Str("contentTitle", item.Content.Title).
			Int("contentNumber", item.Content.Number).
			Int("fieldValueCount", len(item.FieldValues.Nodes)).
			Msg("processing item")

		// Log assignees if present
		if len(item.Content.Assignees.Nodes) > 0 {
			assignees := make([]string, len(item.Content.Assignees.Nodes))
			for j, assignee := range item.Content.Assignees.Nodes {
				assignees[j] = assignee.Login
			}
			log.Debug().
				Str("function", "GetProjectItems").
				Str("itemID", item.ID).
				Strs("assignees", assignees).
				Msg("item has assignees")
		}

		// Log field values
		for j, fieldValue := range item.FieldValues.Nodes {
			log.Debug().
				Str("function", "GetProjectItems").
				Str("itemID", item.ID).
				Int("fieldValueIndex", j).
				Str("fieldValueType", fieldValue.Typename).
				Str("fieldName", fieldValue.Field.Name).
				Interface("fieldValue", fieldValue).
				Msg("processing field value")
		}
	}

	log.Debug().
		Str("function", "GetProjectItems").
		Int("itemCount", len(items)).
		Dur("duration", time.Since(start)).
		Msg("item data processed successfully")

	log.Debug().
		Str("function", "GetProjectItems").
		Dur("duration", time.Since(start)).
		Msg("exiting function")

	return items, nil
}

// AddItemToProject adds an existing issue or PR to a project
func (c *Client) AddItemToProject(ctx context.Context, projectID, contentID string) (string, error) {
	start := time.Now()
	log.Debug().
		Str("function", "AddItemToProject").
		Str("projectID", projectID).
		Str("contentID", contentID).
		Msg("entering function")

	mutation := `
		mutation($projectId: ID!, $contentId: ID!) {
			addProjectV2ItemById(input: { projectId: $projectId, contentId: $contentId }) {
				item {
					id
				}
			}
		}
	`

	log.Debug().
		Str("function", "AddItemToProject").
		Str("mutation", mutation).
		Msg("constructed GraphQL mutation")

	variables := map[string]interface{}{
		"projectId": projectID,
		"contentId": contentID,
	}

	log.Debug().
		Str("function", "AddItemToProject").
		Interface("variables", variables).
		Msg("constructed mutation variables")

	var resp struct {
		AddProjectV2ItemById struct {
			Item struct {
				ID string `json:"id"`
			} `json:"item"`
		} `json:"addProjectV2ItemById"`
	}

	log.Debug().
		Str("function", "AddItemToProject").
		Msg("executing GraphQL mutation")

	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		log.Error().
			Str("function", "AddItemToProject").
			Err(err).
			Str("projectID", projectID).
			Str("contentID", contentID).
			Dur("duration", time.Since(start)).
			Msg("mutation execution failed")
		return "", errors.Wrap(err, "failed to add item to project")
	}

	log.Debug().
		Str("function", "AddItemToProject").
		Interface("response", resp).
		Msg("received GraphQL response")

	itemID := resp.AddProjectV2ItemById.Item.ID
	log.Debug().
		Str("function", "AddItemToProject").
		Str("newItemID", itemID).
		Dur("duration", time.Since(start)).
		Msg("item added to project successfully")

	log.Debug().
		Str("function", "AddItemToProject").
		Dur("duration", time.Since(start)).
		Msg("exiting function")

	return itemID, nil
}

// CreateDraftIssue creates a draft issue in a project
func (c *Client) CreateDraftIssue(ctx context.Context, projectID, title, body string) (string, error) {
	start := time.Now()
	log.Debug().
		Str("function", "CreateDraftIssue").
		Str("projectID", projectID).
		Str("title", title).
		Int("bodyLength", len(body)).
		Msg("entering function")

	mutation := `
		mutation($projectId: ID!, $title: String!, $body: String) {
			addProjectV2DraftIssue(input: { projectId: $projectId, title: $title, body: $body }) {
				projectItem {
					id
				}
			}
		}
	`

	log.Debug().
		Str("function", "CreateDraftIssue").
		Str("mutation", mutation).
		Msg("constructed GraphQL mutation")

	variables := map[string]interface{}{
		"projectId": projectID,
		"title":     title,
		"body":      body,
	}

	log.Debug().
		Str("function", "CreateDraftIssue").
		Interface("variables", variables).
		Msg("constructed mutation variables")

	var resp struct {
		AddProjectV2DraftIssue struct {
			ProjectItem struct {
				ID string `json:"id"`
			} `json:"projectItem"`
		} `json:"addProjectV2DraftIssue"`
	}

	log.Debug().
		Str("function", "CreateDraftIssue").
		Msg("executing GraphQL mutation")

	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		log.Error().
			Str("function", "CreateDraftIssue").
			Err(err).
			Str("projectID", projectID).
			Str("title", title).
			Dur("duration", time.Since(start)).
			Msg("mutation execution failed")
		return "", errors.Wrap(err, "failed to create draft issue")
	}

	log.Debug().
		Str("function", "CreateDraftIssue").
		Interface("response", resp).
		Msg("received GraphQL response")

	itemID := resp.AddProjectV2DraftIssue.ProjectItem.ID
	log.Debug().
		Str("function", "CreateDraftIssue").
		Str("draftIssueID", itemID).
		Dur("duration", time.Since(start)).
		Msg("draft issue created successfully")

	log.Debug().
		Str("function", "CreateDraftIssue").
		Dur("duration", time.Since(start)).
		Msg("exiting function")

	return itemID, nil
}

// UpdateFieldValue updates a field value for a project item
func (c *Client) UpdateFieldValue(ctx context.Context, projectID, itemID, fieldID string, value interface{}) error {
	start := time.Now()
	log.Debug().
		Str("function", "UpdateFieldValue").
		Str("projectID", projectID).
		Str("itemID", itemID).
		Str("fieldID", fieldID).
		Interface("value", value).
		Msg("entering function")

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

	log.Debug().
		Str("function", "UpdateFieldValue").
		Str("mutation", mutation).
		Msg("constructed GraphQL mutation")

	variables := map[string]interface{}{
		"projectId": projectID,
		"itemId":    itemID,
		"fieldId":   fieldID,
		"value":     value,
	}

	log.Debug().
		Str("function", "UpdateFieldValue").
		Interface("variables", variables).
		Msg("constructed mutation variables")

	var resp struct {
		UpdateProjectV2ItemFieldValue struct {
			ProjectV2Item struct {
				ID string `json:"id"`
			} `json:"projectV2Item"`
		} `json:"updateProjectV2ItemFieldValue"`
	}

	log.Debug().
		Str("function", "UpdateFieldValue").
		Msg("executing GraphQL mutation")

	if err := c.ExecuteQuery(ctx, mutation, variables, &resp); err != nil {
		log.Error().
			Str("function", "UpdateFieldValue").
			Err(err).
			Str("projectID", projectID).
			Str("itemID", itemID).
			Str("fieldID", fieldID).
			Interface("value", value).
			Dur("duration", time.Since(start)).
			Msg("mutation execution failed")
		return errors.Wrap(err, "failed to update field value")
	}

	log.Debug().
		Str("function", "UpdateFieldValue").
		Interface("response", resp).
		Msg("received GraphQL response")

	log.Debug().
		Str("function", "UpdateFieldValue").
		Str("updatedItemID", resp.UpdateProjectV2ItemFieldValue.ProjectV2Item.ID).
		Dur("duration", time.Since(start)).
		Msg("field value updated successfully")

	log.Debug().
		Str("function", "UpdateFieldValue").
		Dur("duration", time.Since(start)).
		Msg("exiting function")

	return nil
}
