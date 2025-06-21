package main

import (
	"context"
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

// FieldsCommand lists project fields
type FieldsCommand struct {
	*cmds.CommandDescription
}

// FieldsSettings holds the command settings
type FieldsSettings struct {
	Owner  string `glazed.parameter:"owner"`
	Number int    `glazed.parameter:"number"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &FieldsCommand{}

// RunIntoGlazeProcessor implements the GlazeCommand interface
func (c *FieldsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	start := time.Now()

	// Parse settings
	s := &FieldsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	log.Debug().
		Str("function", "RunIntoGlazeProcessor").
		Str("owner", s.Owner).
		Int("number", s.Number).
		Dur("initialization_duration", time.Since(start)).
		Msg("function entry - fields command starting")

	defer func() {
		log.Debug().
			Str("function", "RunIntoGlazeProcessor").
			Dur("total_duration", time.Since(start)).
			Msg("function exit - fields command completed")
	}()

	// Create GitHub client
	clientStart := time.Now()
	log.Debug().Msg("creating GitHub client")
	client, err := github.NewClient()
	if err != nil {
		log.Error().
			Err(err).
			Dur("client_creation_duration", time.Since(clientStart)).
			Msg("failed to create GitHub client")
		return errors.Wrap(err, "failed to create GitHub client")
	}
	log.Debug().
		Dur("client_creation_duration", time.Since(clientStart)).
		Msg("GitHub client created successfully")

	// Get project
	projectStart := time.Now()
	log.Debug().
		Str("owner", s.Owner).
		Int("number", s.Number).
		Msg("fetching project from GitHub API")
	project, err := client.GetProject(ctx, s.Owner, s.Number)
	if err != nil {
		log.Error().
			Err(err).
			Str("owner", s.Owner).
			Int("number", s.Number).
			Dur("project_fetch_duration", time.Since(projectStart)).
			Msg("failed to get project from GitHub API")
		return errors.Wrap(err, "failed to get project")
	}
	log.Debug().
		Str("project_id", project.ID).
		Str("project_title", project.Title).
		Bool("project_public", project.Public).
		Bool("project_closed", project.Closed).
		Int("project_items_count", project.Items.TotalCount).
		Dur("project_fetch_duration", time.Since(projectStart)).
		Msg("project fetched successfully from GitHub API")

	// Get project fields
	fieldsStart := time.Now()
	log.Debug().
		Str("project_id", project.ID).
		Msg("fetching project fields from GitHub API")
	fields, err := client.GetProjectFields(ctx, project.ID)
	if err != nil {
		log.Error().
			Err(err).
			Str("project_id", project.ID).
			Dur("fields_fetch_duration", time.Since(fieldsStart)).
			Msg("failed to get project fields from GitHub API")
		return errors.Wrap(err, "failed to get project fields")
	}
	log.Debug().
		Int("field_count", len(fields)).
		Str("project_id", project.ID).
		Dur("fields_fetch_duration", time.Since(fieldsStart)).
		Msg("project fields fetched successfully from GitHub API")

	// Create rows for each field
	processingStart := time.Now()
	log.Debug().
		Int("field_count", len(fields)).
		Msg("starting field processing loop")

	for i, field := range fields {
		fieldStart := time.Now()
		log.Debug().
			Int("field_index", i).
			Str("field_id", field.ID).
			Str("field_name", field.Name).
			Str("field_type", field.Typename).
			Msg("processing field")

		row := types.NewRow(
			types.MRP("id", field.ID),
			types.MRP("name", field.Name),
			types.MRP("type", field.Typename),
		)

		// Add options if it's a single-select field
		if field.Typename == "ProjectV2SingleSelectField" && len(field.Options) > 0 {
			log.Debug().
				Str("field_id", field.ID).
				Str("field_name", field.Name).
				Int("option_count", len(field.Options)).
				Msg("processing single-select field options")

			var optionNames []string
			var optionIDs []string
			for j, option := range field.Options {
				log.Debug().
					Int("option_index", j).
					Str("option_id", option.ID).
					Str("option_name", option.Name).
					Msg("processing field option")
				optionNames = append(optionNames, option.Name)
				optionIDs = append(optionIDs, option.ID)
			}
			row.Set("option_names", optionNames)
			row.Set("option_ids", optionIDs)

			log.Debug().
				Str("field_id", field.ID).
				Str("field_name", field.Name).
				Int("option_count", len(field.Options)).
				Interface("option_names", optionNames).
				Interface("option_ids", optionIDs).
				Msg("single-select field options processed")
		}

		// Add iterations if it's an iteration field
		if field.Typename == "ProjectV2IterationField" && field.Configuration != nil && len(field.Configuration.Iterations) > 0 {
			log.Debug().
				Str("field_id", field.ID).
				Str("field_name", field.Name).
				Int("iteration_count", len(field.Configuration.Iterations)).
				Msg("processing iteration field iterations")

			var iterationTitles []string
			var iterationIDs []string
			var iterationDates []string
			for j, iteration := range field.Configuration.Iterations {
				log.Debug().
					Int("iteration_index", j).
					Str("iteration_id", iteration.ID).
					Str("iteration_title", iteration.Title).
					Str("iteration_start_date", iteration.StartDate).
					Msg("processing iteration")
				iterationTitles = append(iterationTitles, iteration.Title)
				iterationIDs = append(iterationIDs, iteration.ID)
				iterationDates = append(iterationDates, iteration.StartDate)
			}
			row.Set("iteration_titles", iterationTitles)
			row.Set("iteration_ids", iterationIDs)
			row.Set("iteration_dates", iterationDates)

			log.Debug().
				Str("field_id", field.ID).
				Str("field_name", field.Name).
				Int("iteration_count", len(field.Configuration.Iterations)).
				Interface("iteration_titles", iterationTitles).
				Interface("iteration_ids", iterationIDs).
				Interface("iteration_dates", iterationDates).
				Msg("iteration field iterations processed")
		}

		// Add row to processor
		if err := gp.AddRow(ctx, row); err != nil {
			log.Error().
				Err(err).
				Int("field_index", i).
				Str("field_id", field.ID).
				Str("field_name", field.Name).
				Str("field_type", field.Typename).
				Dur("field_processing_duration", time.Since(fieldStart)).
				Msg("failed to add row to processor")
			return err
		}

		log.Debug().
			Int("field_index", i).
			Str("field_id", field.ID).
			Str("field_name", field.Name).
			Str("field_type", field.Typename).
			Dur("field_processing_duration", time.Since(fieldStart)).
			Msg("field processed and row added successfully")
	}

	log.Debug().
		Int("field_count", len(fields)).
		Dur("processing_duration", time.Since(processingStart)).
		Msg("field processing loop completed")

	return nil
}

// NewFieldsCommand creates a new fields command
func NewFieldsCommand() (*FieldsCommand, error) {
	start := time.Now()

	log.Debug().
		Str("function", "NewFieldsCommand").
		Msg("function entry - creating new fields command")

	defer func() {
		log.Debug().
			Str("function", "NewFieldsCommand").
			Dur("total_duration", time.Since(start)).
			Msg("function exit - fields command creation completed")
	}()

	// Create Glazed layer for output formatting
	glazedStart := time.Now()
	log.Debug().Msg("creating glazed parameter layers")
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		log.Error().
			Err(err).
			Dur("glazed_creation_duration", time.Since(glazedStart)).
			Msg("failed to create glazed parameter layers")
		return nil, err
	}
	log.Debug().
		Dur("glazed_creation_duration", time.Since(glazedStart)).
		Msg("glazed parameter layers created successfully")

	// Create command description
	cmdStart := time.Now()
	log.Debug().Msg("creating command description")
	cmdDesc := cmds.NewCommandDescription(
		"fields",
		cmds.WithShort("List project fields"),
		cmds.WithLong(`
List all fields for a GitHub Project v2, including their types and options.

Examples:
  github-graphql-cli fields --owner=myorg --number=5
  github-graphql-cli fields --owner=myorg --number=5 --output=json
  github-graphql-cli fields --owner=myorg --number=5 --fields=name,type
		`),
		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"owner",
				parameters.ParameterTypeString,
				parameters.WithHelp("Organization or user name that owns the project"),
				parameters.WithDefault(githubConfig.Owner),
			),
			parameters.NewParameterDefinition(
				"number",
				parameters.ParameterTypeInteger,
				parameters.WithHelp("Project number"),
				parameters.WithDefault(githubConfig.ProjectNumber),
			),
		),
		// Add parameter layers
		cmds.WithLayersList(
			glazedLayer,
		),
	)
	log.Debug().
		Dur("command_description_duration", time.Since(cmdStart)).
		Msg("command description created successfully")

	log.Debug().
		Str("command_name", "fields").
		Str("command_short", "List project fields").
		Msg("fields command initialized with parameters")

	return &FieldsCommand{
		CommandDescription: cmdDesc,
	}, nil
}
