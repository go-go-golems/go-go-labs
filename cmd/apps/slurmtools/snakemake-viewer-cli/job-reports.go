package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/go-go-labs/pkg/jobreports"
)

type JobReportsCommand struct {
	*cmds.CommandDescription
}

type JobReportsSettings struct {
	ReportFiles []string `glazed.parameter:"reportfiles"`
	Verbose     bool     `glazed.parameter:"verbose"`
	Legacy      bool     `glazed.parameter:"legacy"`
	Data        string   `glazed.parameter:"data"`
}

func NewJobReportsCommand() (*JobReportsCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &JobReportsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"job-reports",
			cmds.WithShort("View parsed job report information"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"verbose",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Display verbose output"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"legacy",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Use legacy text output"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"data",
					parameters.ParameterTypeChoice,
					parameters.WithHelp("Specify data to output: summary, jobs, or all"),
					parameters.WithDefault("jobs"),
					parameters.WithChoices("summary", "jobs", "all"),
				),
			),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"reportfiles",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Paths to the job report files"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (c *JobReportsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &JobReportsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	for _, reportFile := range s.ReportFiles {
		reportData, err := jobreports.ParseJobReportFile(reportFile)
		if err != nil {
			return fmt.Errorf("error parsing report file %s: %w", reportFile, err)
		}

		if s.Legacy {
			if err := c.legacyOutput(reportData, s.Verbose, s.Data, reportFile); err != nil {
				return err
			}
		} else {
			if err := c.structuredOutput(ctx, gp, reportData, s.Verbose, s.Data, reportFile); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *JobReportsCommand) legacyOutput(reportData *jobreports.ReportData, verbose bool, dataType string, filename string) error {
	if dataType == "summary" || dataType == "all" {
		fmt.Printf("File: %s\n", filename)
		fmt.Printf("Total jobs: %d\n", reportData.TotalJobs)
		fmt.Printf("Last updated: %s\n", reportData.LastUpdated.Format(time.RFC3339))
		fmt.Println()
	}

	if dataType == "jobs" || dataType == "all" {
		fmt.Println("Job Information:")
		for _, job := range reportData.Jobs {
			fmt.Printf("Job ID: %s\n", job.ID)
			fmt.Printf("  User: %s\n", job.User)
			fmt.Printf("  Account: %s\n", job.Account)
			fmt.Printf("  Partition: %s\n", job.Partition)
			fmt.Printf("  Status: %s\n", job.Status)
			fmt.Printf("  Start Time: %s\n", job.StartTime.Format(time.RFC3339))
			fmt.Printf("  Wall Time: %s\n", job.WallTime)
			fmt.Printf("  Run Time: %s\n", job.RunTime)
			fmt.Printf("  CPUs: %d\n", job.CPUs)
			fmt.Printf("  RAM: %.2f GB\n", job.RAM)
			fmt.Printf("  GPUs: %d\n", job.GPUs)
			fmt.Printf("  Pending Time: %s\n", job.PendingTime)
			fmt.Printf("  CPU Efficiency: %.2f%%\n", job.CPUEfficiency)
			fmt.Printf("  RAM Efficiency: %.2f%%\n", job.RAMEfficiency)
			fmt.Printf("  Wall Time Efficiency: %.2f%%\n", job.WallTimeEfficiency)
			fmt.Println()
		}
	}

	return nil
}

func (c *JobReportsCommand) structuredOutput(ctx context.Context, gp middlewares.Processor, reportData *jobreports.ReportData, verbose bool, dataType string, filename string) error {
	if dataType == "summary" || dataType == "all" {
		err := gp.AddRow(ctx, types.NewRowFromMap(map[string]interface{}{
			"file":         filename,
			"total_jobs":   reportData.TotalJobs,
			"last_updated": reportData.LastUpdated,
		}))
		if err != nil {
			return err
		}
	}

	if dataType == "jobs" || dataType == "all" {
		for _, job := range reportData.Jobs {
			err := gp.AddRow(ctx, types.NewRowFromMap(map[string]interface{}{
				"id":                   job.ID,
				"user":                 job.User,
				"account":              job.Account,
				"partition":            job.Partition,
				"status":               job.Status,
				"start_time":           job.StartTime,
				"wall_time":            job.WallTime,
				"run_time":             job.RunTime,
				"cpus":                 job.CPUs,
				"ram":                  job.RAM,
				"gpus":                 job.GPUs,
				"pending_time":         job.PendingTime,
				"cpu_efficiency":       job.CPUEfficiency,
				"ram_efficiency":       job.RAMEfficiency,
				"wall_time_efficiency": job.WallTimeEfficiency,
			}))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
