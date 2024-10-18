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
	"github.com/go-go-golems/go-go-labs/pkg/snakemake"
	"github.com/go-go-golems/go-go-labs/pkg/jobreports"
)

type SnakemakeViewerCommand struct {
	*cmds.CommandDescription
}

type SnakemakeViewerSettings struct {
	LogFiles  []string `glazed.parameter:"logfiles"`
	Verbose   bool     `glazed.parameter:"verbose"`
	Debug     bool     `glazed.parameter:"debug"`
	Legacy    bool     `glazed.parameter:"legacy"`
	Data      string   `glazed.parameter:"data"`
	JobReport string   `glazed.parameter:"job-report"`
}

func NewSnakemakeViewerCommand() (*SnakemakeViewerCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create Glazed parameter layer: %w", err)
	}

	return &SnakemakeViewerCommand{
		CommandDescription: cmds.NewCommandDescription(
			"view",
			cmds.WithShort("View parsed Snakemake log information"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"verbose",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Display verbose output"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"debug",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Enable debug logging"),
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
					parameters.WithHelp("Specify data to output: rules, summary, jobs, or all"),
					parameters.WithDefault("jobs"),
					parameters.WithChoices("rules", "summary", "jobs", "all"),
				),
				parameters.NewParameterDefinition(
					"job-report",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to the job report file"),
				),
			),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"logfiles",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Paths to the Snakemake log files"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (c *SnakemakeViewerCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &SnakemakeViewerSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	var jobReportData *jobreports.ReportData
	if s.JobReport != "" {
		var err error
		jobReportData, err = jobreports.ParseJobReportFile(s.JobReport)
		if err != nil {
			return fmt.Errorf("error parsing job report file %s: %w", s.JobReport, err)
		}
	}

	for _, logFile := range s.LogFiles {
		logData, err := snakemake.ParseLog(logFile, s.Debug)
		if err != nil {
			return fmt.Errorf("error parsing log file %s: %w", logFile, err)
		}

		if jobReportData != nil {
			logData = c.joinJobReportData(logData, jobReportData)
		}

		if s.Legacy {
			if err := c.legacyOutput(logData, s.Verbose, s.Data, logFile); err != nil {
				return err
			}
		} else {
			if err := c.structuredOutput(ctx, gp, logData, s.Verbose, s.Data, logFile); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *SnakemakeViewerCommand) joinJobReportData(logData snakemake.LogData, jobReportData *jobreports.ReportData) snakemake.LogData {
	jobMap := make(map[string]*jobreports.Job)
	for _, job := range jobReportData.Jobs {
		jobMap[job.ID] = job
	}

	for i, job := range logData.Jobs {
		if reportJob, ok := jobMap[job.ID]; ok {
			logData.Jobs[i].JobReport = reportJob
		}
	}

	return logData
}

func (c *SnakemakeViewerCommand) legacyOutput(logData snakemake.LogData, verbose bool, dataType string, filename string) error {
	// Update the summary output
	if dataType == "summary" || dataType == "all" {
		fmt.Printf("Total jobs: %d\n", logData.TotalJobs)
		fmt.Printf("Completed jobs: %d\n", logData.Completed)
		fmt.Printf("In-progress jobs: %d\n", logData.InProgress)
		fmt.Println()
	}

	// Output rule summary
	if dataType == "rules" || dataType == "all" {
		fmt.Println("Rule Summary:")
		for ruleName, rule := range logData.Rules {
			completedJobs := 0
			var totalDuration time.Duration
			for _, job := range rule.Jobs {
				if job.Status == snakemake.StatusCompleted {
					completedJobs++
					totalDuration += job.Duration
				}
			}

			avgDuration := time.Duration(0)
			if completedJobs > 0 {
				avgDuration = totalDuration / time.Duration(completedJobs)
			}

			fmt.Printf("Rule: %s\n", ruleName)
			fmt.Printf("  Total jobs: %d\n", len(rule.Jobs))
			fmt.Printf("  Completed jobs: %d\n", completedJobs)
			fmt.Printf("  Average duration: %s\n", avgDuration)
			for _, resource := range rule.Resources {
				fmt.Printf("  Resource %s: %v\n", resource.Name, resource.Value)
			}
			fmt.Println()
		}
	}

	// Update the detailed job information output
	if (dataType == "jobs" || dataType == "all") && verbose {
		fmt.Println("Detailed Job Information:")
		for _, job := range logData.Jobs {
			fmt.Printf("Job ID: %s\n", job.ID)
			fmt.Printf("  Rule: %s\n", job.Rule)
			fmt.Printf("  Start Time: %s\n", job.StartTime.Format(time.RFC3339))
			fmt.Printf("  End Time: %s\n", job.EndTime.Format(time.RFC3339))
			fmt.Printf("  Duration: %s\n", job.Duration)
			fmt.Printf("  Status: %s\n", job.Status)
			fmt.Printf("  Details: %s\n", job.Details)
			for _, resource := range job.Resources {
				fmt.Printf("  Resource %s: %v\n", resource.Name, resource.Value)
			}
			if job.ScannerError != "" {
				fmt.Printf("  Scanner Error: %s\n", job.ScannerError)
			}
			// Add job report data if available
			if job.JobReport != nil {
				fmt.Printf("  User: %s\n", job.JobReport.User)
				fmt.Printf("  Account: %s\n", job.JobReport.Account)
				fmt.Printf("  Partition: %s\n", job.JobReport.Partition)
				fmt.Printf("  CPUs: %d\n", job.JobReport.CPUs)
				fmt.Printf("  RAM: %.2f GB\n", job.JobReport.RAM)
				fmt.Printf("  GPUs: %d\n", job.JobReport.GPUs)
				fmt.Printf("  Pending Time: %s\n", job.JobReport.PendingTime)
				fmt.Printf("  CPU Efficiency: %.2f%%\n", job.JobReport.CPUEfficiency)
				fmt.Printf("  RAM Efficiency: %.2f%%\n", job.JobReport.RAMEfficiency)
				fmt.Printf("  Wall Time Efficiency: %.2f%%\n", job.JobReport.WallTimeEfficiency)
			}
			fmt.Println()
		}
	}

	return nil
}

func (c *SnakemakeViewerCommand) structuredOutput(ctx context.Context, gp middlewares.Processor, logData snakemake.LogData, verbose bool, dataType string, filename string) error {
	if dataType == "summary" || dataType == "all" {
		err := gp.AddRow(ctx, types.NewRowFromMap(map[string]interface{}{
			"total_jobs":      logData.TotalJobs,
			"completed_jobs":  logData.Completed,
			"in_progress_jobs": logData.InProgress,
			"filename":        filename,
		}))
		if err != nil {
			return err
		}
	}

	if dataType == "rules" || dataType == "all" {
		for ruleName, rule := range logData.Rules {
			completedJobs := 0
			var totalDuration time.Duration
			for _, job := range rule.Jobs {
				if job.Status == snakemake.StatusCompleted {
					completedJobs++
					totalDuration += job.Duration
				}
			}

			avgDuration := time.Duration(0)
			if completedJobs > 0 {
				avgDuration = totalDuration / time.Duration(completedJobs)
			}

			ruleRow := types.NewRowFromMap(map[string]interface{}{
				"rule_name":       ruleName,
				"total_jobs":      len(rule.Jobs),
				"completed_jobs":  completedJobs,
				"average_duration": avgDuration.String(),
				"filename":        filename,
			})

			for _, resource := range rule.Resources {
				ruleRow.Set(fmt.Sprintf("resource_%s", resource.Name), resource.Value)
			}

			if err := gp.AddRow(ctx, ruleRow); err != nil {
				return err
			}
		}
	}

	if dataType == "jobs" || dataType == "all" {
		for _, job := range logData.Jobs {
			jobRow := types.NewRowFromMap(map[string]interface{}{
				"job_id":     job.ID,
				"rule":       job.Rule,
				"start_time": job.StartTime,
				"end_time":   job.EndTime,
				"duration":   job.Duration.String(),
				"status":     string(job.Status),
				"filename":   filename,
			})

			// Add job report data if available
			if job.JobReport != nil {
				jobRow.Set("jobreport_id", job.JobReport.ID)
				jobRow.Set("jobreport_user", job.JobReport.User)
				jobRow.Set("jobreport_account", job.JobReport.Account)
				jobRow.Set("jobreport_partition", job.JobReport.Partition)
				jobRow.Set("jobreport_status", string(job.JobReport.Status))
				jobRow.Set("jobreport_start_time", job.JobReport.StartTime)
				jobRow.Set("jobreport_wall_time", job.JobReport.WallTime.String())
				jobRow.Set("jobreport_run_time", job.JobReport.RunTime.String())
				jobRow.Set("jobreport_cpus", job.JobReport.CPUs)
				jobRow.Set("jobreport_ram", job.JobReport.RAM)
				jobRow.Set("jobreport_gpus", job.JobReport.GPUs)
				jobRow.Set("jobreport_pending_time", job.JobReport.PendingTime.String())
				jobRow.Set("jobreport_cpu_efficiency", job.JobReport.CPUEfficiency)
				jobRow.Set("jobreport_ram_efficiency", job.JobReport.RAMEfficiency)
				jobRow.Set("jobreport_wall_time_efficiency", job.JobReport.WallTimeEfficiency)
			}

			for _, resource := range job.Resources {
				jobRow.Set(fmt.Sprintf("resource_%s", resource.Name), resource.Value)
			}

			if verbose {
				jobRow.Set("input", job.Input)
				jobRow.Set("output", job.Output)
				jobRow.Set("reason", job.Reason)
				for field, value := range job.Details {
					jobRow.Set(fmt.Sprintf("detail_%s", field), value)
				}
			}

			if job.ScannerError != "" {
				jobRow.Set("scanner_error", job.ScannerError)
			}

			if err := gp.AddRow(ctx, jobRow); err != nil {
				return err
			}
		}
	}

	return nil
}
