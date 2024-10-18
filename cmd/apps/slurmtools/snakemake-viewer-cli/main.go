package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/go-go-labs/pkg/snakemake"
	"github.com/go-go-golems/go-go-labs/pkg/snakemake/doc"
	"github.com/spf13/cobra"
)

type SnakemakeViewerCommand struct {
	*cmds.CommandDescription
}

type SnakemakeViewerSettings struct {
	LogFiles []string `glazed.parameter:"logfiles"`
	Verbose  bool     `glazed.parameter:"verbose"`
	Debug    bool     `glazed.parameter:"debug"`
	Legacy   bool     `glazed.parameter:"legacy"`
	Data     string   `glazed.parameter:"data"`
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

	for _, logFile := range s.LogFiles {
		logData, err := snakemake.ParseLog(logFile, s.Debug)
		if err != nil {
			return fmt.Errorf("error parsing log file %s: %w", logFile, err)
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
			fmt.Println()
		}
	}

	return nil
}

func (c *SnakemakeViewerCommand) structuredOutput(ctx context.Context, gp middlewares.Processor, logData snakemake.LogData, verbose bool, dataType string, filename string) error {
	return snakemake.OutputLogDataToGlazedProcessor(ctx, gp, logData, verbose, dataType, filename)
}

func main() {
	cmd, err := NewSnakemakeViewerCommand()
	if err != nil {
		fmt.Printf("Error creating command: %v\n", err)
		os.Exit(1)
	}

	// Create a new help system
	helpSystem := help.NewHelpSystem()

	// Add documentation to the help system
	err = doc.AddDocToHelpSystem(helpSystem)
	if err != nil {
		fmt.Printf("Error adding documentation to help system: %v\n", err)
		os.Exit(1)
	}

	glazedCmd, err := cli.BuildCobraCommandFromGlazeCommand(cmd)
	if err != nil {
		fmt.Printf("Error creating Glazed command: %v\n", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{Use: "snakemake-viewer-cli"}
	rootCmd.AddCommand(glazedCmd)

	helpSystem.SetupCobraRootCommand(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
