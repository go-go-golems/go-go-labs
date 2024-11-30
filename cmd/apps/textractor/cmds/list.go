package cmds

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/pkg"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

type ListCommand struct {
	*cmds.CommandDescription
}

type ListSettings struct {
	Since time.Time `glazed.parameter:"since"`

	Status string `glazed.parameter:"status"`
	TfDir  string `glazed.parameter:"tf-dir"`
	Config string `glazed.parameter:"config"`
}

func NewListCommand() (*ListCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &ListCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List Textract jobs"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"since",
					parameters.ParameterTypeDate,
					parameters.WithHelp("Only show jobs submitted after this time (RFC3339 format)"),
				),
				parameters.NewParameterDefinition(
					"status",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter by job status (UPLOADING, SUBMITTED, PROCESSING, COMPLETED, FAILED, ERROR)"),
				),
				parameters.NewParameterDefinition(
					"tf-dir",
					parameters.ParameterTypeString,
					parameters.WithHelp("Terraform directory"),
				),
				parameters.NewParameterDefinition(
					"config",
					parameters.ParameterTypeString,
					parameters.WithHelp("Config file"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *ListCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ListSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	stateLoader := pkg.NewStateLoader()
	resources, err := stateLoader.LoadState(s.TfDir, s.Config)
	if err != nil {
		return fmt.Errorf("failed to load terraform state: %w", err)
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(resources.Region),
	}))

	jobClient := pkg.NewJobClient(sess, resources.JobsTable)

	var opts pkg.ListJobsOptions
	if !s.Since.IsZero() {
		opts.Since = &s.Since
	}
	opts.Status = s.Status

	jobs, err := jobClient.ListJobs(opts)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].SubmittedAt.After(jobs[j].SubmittedAt)
	})

	for _, job := range jobs {
		row := types.NewRow(
			types.MRP("job_id", job.JobID),
			types.MRP("document", job.DocumentKey),
			types.MRP("status", job.Status),
			types.MRP("submitted_at", job.SubmittedAt),
			types.MRP("textract_id", job.TextractID),
			types.MRP("result_key", job.ResultKey),
		)

		if job.CompletedAt != nil {
			row.Set("completed_at", job.CompletedAt)
		}
		if job.Error != "" {
			row.Set("error", job.Error)
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
