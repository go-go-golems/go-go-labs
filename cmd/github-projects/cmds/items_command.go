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

	"github.com/go-go-golems/go-go-labs/cmd/github-projects/config"
	"github.com/go-go-golems/go-go-labs/pkg/github"
)

// ItemsCommand lists project items
type ItemsCommand struct {
	*cmds.CommandDescription
}

// ItemsSettings holds the command settings
type ItemsSettings struct {
	Owner  string `glazed.parameter:"owner"`
	Number int    `glazed.parameter:"number"`
	Limit  int    `glazed.parameter:"limit"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &ItemsCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *ItemsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	start := time.Now()

	// Parse settings
	s := &ItemsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Create contextual logger
	logger := log.With().
		Str("function", "RunIntoGlazeProcessor").
		Str("owner", s.Owner).
		Int("number", s.Number).
		Int("limit", s.Limit).
		Logger()

	logger.Debug().Msg("starting items command")

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

	// Get project
	projectStart := time.Now()
	logger.Debug().Msg("fetching project from GitHub API")

	project, err := client.GetProject(ctx, s.Owner, s.Number)
	if err != nil {
		logger.Error().
			Err(err).
			Dur("duration", time.Since(projectStart)).
			Msg("failed to get project")
		return errors.Wrap(err, "failed to get project")
	}

	projectLogger := logger.With().
		Str("project_id", project.ID).
		Str("project_title", project.Title).
		Logger()

	projectLogger.Debug().
		Dur("duration", time.Since(projectStart)).
		Msg("project retrieved")

	// Get project items
	itemsStart := time.Now()
	projectLogger.Debug().Msg("fetching project items from GitHub API")

	items, err := client.GetProjectItems(ctx, project.ID, s.Limit)
	if err != nil {
		projectLogger.Error().
			Err(err).
			Dur("duration", time.Since(itemsStart)).
			Msg("failed to get project items")
		return errors.Wrap(err, "failed to get project items")
	}

	projectLogger.Debug().
		Int("items_count", len(items)).
		Dur("duration", time.Since(itemsStart)).
		Msg("project items retrieved")

	// Create rows for each item
	processStart := time.Now()
	projectLogger.Debug().
		Int("items_to_process", len(items)).
		Msg("processing items")

	for i, item := range items {
		itemStart := time.Now()
		itemLogger := projectLogger.With().
			Int("item_index", i).
			Str("item_id", item.ID).
			Str("type", item.Type).
			Str("content_type", item.Content.Typename).
			Logger()

		itemLogger.Trace().
			Str("title", item.Content.Title).
			Msg("processing item")

		row := types.NewRow(
			types.MRP("item_id", item.ID),
			types.MRP("type", item.Type),
			types.MRP("content_type", item.Content.Typename),
			types.MRP("title", item.Content.Title),
			types.MRP("number", item.Content.Number),
			types.MRP("url", item.Content.URL),
		)

		// Add assignees
		if len(item.Content.Assignees.Nodes) > 0 {
			var assignees []string
			for _, assignee := range item.Content.Assignees.Nodes {
				assignees = append(assignees, assignee.Login)
			}
			assigneesStr := strings.Join(assignees, ", ")
			row.Set("assignees", assigneesStr)
			itemLogger.Trace().
				Strs("assignees", assignees).
				Msg("assignees processed")
		}

		// Add body for draft issues
		if item.Content.Typename == "DraftIssue" {
			row.Set("body", item.Content.Body)
			itemLogger.Trace().
				Int("body_length", len(item.Content.Body)).
				Msg("draft issue body added")
		}

		// Add field values
		fieldValuesStart := time.Now()
		fieldValues := make(map[string]interface{})
		itemLogger.Trace().
			Int("field_values_count", len(item.FieldValues.Nodes)).
			Msg("processing field values")

		for _, fieldValue := range item.FieldValues.Nodes {
			fieldName := fieldValue.Field.Name
			fieldLogger := itemLogger.With().
				Str("field_name", fieldName).
				Str("field_type", fieldValue.Typename).
				Logger()

			switch fieldValue.Typename {
			case "ProjectV2ItemFieldTextValue":
				if fieldValue.Text != nil {
					fieldValues[fieldName] = *fieldValue.Text
					fieldLogger.Trace().
						Str("text_value", *fieldValue.Text).
						Msg("text field value set")
				}
			case "ProjectV2ItemFieldNumberValue":
				if fieldValue.Number != nil {
					fieldValues[fieldName] = *fieldValue.Number
					fieldLogger.Trace().
						Float64("number_value", *fieldValue.Number).
						Msg("number field value set")
				}
			case "ProjectV2ItemFieldDateValue":
				if fieldValue.Date != nil {
					fieldValues[fieldName] = *fieldValue.Date
					fieldLogger.Trace().
						Str("date_value", *fieldValue.Date).
						Msg("date field value set")
				}
			case "ProjectV2ItemFieldSingleSelectValue":
				if fieldValue.Name != nil {
					fieldValues[fieldName] = *fieldValue.Name
					fieldLogger.Trace().
						Str("select_value", *fieldValue.Name).
						Msg("single select field value set")
				}
			case "ProjectV2ItemFieldIterationValue":
				if fieldValue.Title != nil {
					fieldValues[fieldName] = *fieldValue.Title
					fieldLogger.Trace().
						Str("iteration_value", *fieldValue.Title).
						Msg("iteration field value set")
				}
			default:
				fieldLogger.Trace().
					Msg("unknown field type, skipping")
			}
		}

		itemLogger.Trace().
			Int("processed_field_values", len(fieldValues)).
			Dur("duration", time.Since(fieldValuesStart)).
			Msg("field values processed")

		// Add field values to row
		for fieldName, value := range fieldValues {
			columnName := "field_" + strings.ToLower(strings.ReplaceAll(fieldName, " ", "_"))
			row.Set(columnName, value)
		}

		// Add row to processor
		if err := gp.AddRow(ctx, row); err != nil {
			itemLogger.Error().
				Err(err).
				Dur("item_duration", time.Since(itemStart)).
				Msg("failed to add row to processor")
			return err
		}

		itemLogger.Trace().
			Dur("total_item_duration", time.Since(itemStart)).
			Msg("item processed")
	}

	projectLogger.Debug().
		Int("total_items_processed", len(items)).
		Dur("processing_duration", time.Since(processStart)).
		Dur("total_duration", time.Since(start)).
		Msg("items command completed")

	return nil
}

// NewItemsCommand creates a new items command
func NewItemsCommand() (*ItemsCommand, error) {
	start := time.Now()
	logger := log.With().
		Str("function", "NewItemsCommand").
		Logger()

	logger.Trace().Msg("creating items command")

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
		"items",
		cmds.WithShort("List project items"),
		cmds.WithLong(`
List all items in a GitHub Project v2 with their field values.

Examples:
  github-graphql-cli items --owner=myorg --number=5 --limit=10
  github-graphql-cli items --owner=myorg --number=5 --limit=10 --output=json
  github-graphql-cli items --owner=myorg --number=5 --fields=title,type,assignees
		`),
		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"owner",
				parameters.ParameterTypeString,
				parameters.WithHelp("Organization or user name that owns the project"),
				parameters.WithDefault(config.GetDefaultOwner()),
			),
			parameters.NewParameterDefinition(
				"number",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Project number"),
				parameters.WithDefault(config.GetDefaultProjectNumber()),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of items to return"),
				parameters.WithDefault(20),
			),
		),
		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	command := &ItemsCommand{
		CommandDescription: cmdDesc,
	}

	logger.Trace().
		Dur("total_duration", time.Since(start)).
		Msg("items command created")

	return command, nil
}
