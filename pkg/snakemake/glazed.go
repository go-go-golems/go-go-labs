package snakemake

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// OutputLogDataToGlazedProcessor outputs the LogData to a Glazed Processor
func OutputLogDataToGlazedProcessor(ctx context.Context, gp middlewares.Processor, logData LogData, verbose bool, dataType string, filename string) error {
	// Output summary
	if dataType == "summary" || dataType == "all" {
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("total_jobs", logData.TotalJobs),
			types.MRP("completed_jobs", logData.Completed),
			types.MRP("in_progress_jobs", logData.InProgress),
			types.MRP("filename", filename),
		)); err != nil {
			return err
		}
	}

	// Output rule summary
	if dataType == "rules" || dataType == "all" {
		if err := outputRuleSummary(ctx, gp, logData, filename); err != nil {
			return err
		}
	}

	// Output detailed job information
	if dataType == "jobs" || dataType == "all" {
		if err := outputJobDetails(ctx, gp, logData, verbose, filename); err != nil {
			return err
		}
	}

	return nil
}

func outputRuleSummary(ctx context.Context, gp middlewares.Processor, logData LogData, filename string) error {
	for ruleName, rule := range logData.Rules {
		completedJobs := 0
		var totalDuration time.Duration
		for _, job := range rule.Jobs {
			if job.Status == StatusCompleted {
				completedJobs++
				totalDuration += job.Duration
			}
		}

		avgDuration := time.Duration(0)
		if completedJobs > 0 {
			avgDuration = totalDuration / time.Duration(completedJobs)
		}

		ruleRow := types.NewRow(
			types.MRP("rule_name", ruleName),
			types.MRP("total_jobs", len(rule.Jobs)),
			types.MRP("completed_jobs", completedJobs),
			types.MRP("average_duration", avgDuration.String()),
			types.MRP("filename", filename),
		)

		for _, resource := range rule.Resources {
			ruleRow.Set(fmt.Sprintf("resource_%s", resource.Name), resource.Value)
		}

		if err := gp.AddRow(ctx, ruleRow); err != nil {
			return err
		}
	}
	return nil
}

func outputJobDetails(ctx context.Context, gp middlewares.Processor, logData LogData, verbose bool, filename string) error {
	for _, job := range logData.Jobs {
		jobRow := types.NewRow(
			types.MRP("job_id", job.ID),
			types.MRP("rule", job.Rule),
			types.MRP("start_time", job.StartTime.Format(time.RFC3339)),
			types.MRP("end_time", job.EndTime.Format(time.RFC3339)),
			types.MRP("duration", job.Duration.String()),
			types.MRP("duration_s", job.Duration.Seconds()),
			types.MRP("status", string(job.Status)),
			types.MRP("threads", job.Threads),
			types.MRP("filename", filename),
		)

		for _, resource := range job.Resources {
			jobRow.Set(fmt.Sprintf("resource_%s", resource.Name), resource.Value)
		}

		if verbose {
			jobRow.Set("input", strings.Join(job.Input, ", "))
			jobRow.Set("output", strings.Join(job.Output, ", "))
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
	return nil
}
