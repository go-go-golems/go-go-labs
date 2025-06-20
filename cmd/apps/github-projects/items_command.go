package main

import (
	"context"
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/go-go-golems/go-go-labs/cmd/apps/github-projects/pkg/github"
)

// ItemsCommand lists project items
type ItemsCommand struct {
	*cmds.CommandDescription
}

// ItemsSettings holds the command settings
type ItemsSettings struct {
	Owner    string `glazed.parameter:"owner"`
	Number   int    `glazed.parameter:"number"`
	Limit    int    `glazed.parameter:"limit"`
	LogLevel string `glazed.parameter:"log-level"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &ItemsCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *ItemsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings
	s := &ItemsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Set up logger
	level, err := zerolog.ParseLevel(s.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	logger := zerolog.New(os.Stderr).Level(level).With().Timestamp().Logger()

	// Create GitHub client
	client, err := github.NewClient(logger)
	if err != nil {
		return errors.Wrap(err, "failed to create GitHub client")
	}

	// Get project
	project, err := client.GetProject(ctx, s.Owner, s.Number)
	if err != nil {
		return errors.Wrap(err, "failed to get project")
	}

	// Get project items
	items, err := client.GetProjectItems(ctx, project.ID, s.Limit)
	if err != nil {
		return errors.Wrap(err, "failed to get project items")
	}

	// Create rows for each item
	for _, item := range items {
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
			row.Set("assignees", strings.Join(assignees, ", "))
		}

		// Add body for draft issues
		if item.Content.Typename == "DraftIssue" {
			row.Set("body", item.Content.Body)
		}

		// Add field values
		fieldValues := make(map[string]interface{})
		for _, fieldValue := range item.FieldValues.Nodes {
			fieldName := fieldValue.Field.Name
			switch fieldValue.Typename {
			case "ProjectV2ItemFieldTextValue":
				if fieldValue.Text != nil {
					fieldValues[fieldName] = *fieldValue.Text
				}
			case "ProjectV2ItemFieldNumberValue":
				if fieldValue.Number != nil {
					fieldValues[fieldName] = *fieldValue.Number
				}
			case "ProjectV2ItemFieldDateValue":
				if fieldValue.Date != nil {
					fieldValues[fieldName] = *fieldValue.Date
				}
			case "ProjectV2ItemFieldSingleSelectValue":
				if fieldValue.Name != nil {
					fieldValues[fieldName] = *fieldValue.Name
				}
			case "ProjectV2ItemFieldIterationValue":
				if fieldValue.Title != nil {
					fieldValues[fieldName] = *fieldValue.Title
				}
			}
		}

		// Add field values to row
		for fieldName, value := range fieldValues {
			row.Set("field_"+strings.ToLower(strings.ReplaceAll(fieldName, " ", "_")), value)
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// NewItemsCommand creates a new items command
func NewItemsCommand() (*ItemsCommand, error) {
	// Create Glazed layer for output formatting
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

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
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"number",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Project number"),
				parameters.WithRequired(true),
			),
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Maximum number of items to return"),
				parameters.WithDefault(20),
			),
			parameters.NewParameterDefinition(
				"log-level",
				parameters.ParameterTypeChoice,
				parameters.WithHelp("Log level"),
				parameters.WithDefault("info"),
				parameters.WithChoices("trace", "debug", "info", "warn", "error"),
			),
		),
		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)

	return &ItemsCommand{
		CommandDescription: cmdDesc,
	}, nil
}
