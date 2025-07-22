package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "redis-monitor",
	Short: "Redis Streams Monitor - CLI and TUI tools for monitoring Redis streams",
	Long: `Redis Streams Monitor provides both command-line tools and a terminal UI
for monitoring Redis streams, consumer groups, and performance metrics.

The CLI commands provide structured output suitable for scripting and automation,
while the TUI provides a top-like interface for interactive monitoring.`,
}

func init() {
	// Add subcommands
	addStreamCommands()
	addGroupCommands()
	addMetricsCommands()
	addTUICommands()
	addDemoCommand()

	helpSystem := help.NewHelpSystem()
	helpCmd := help.NewCobraHelpCommand(helpSystem)
	rootCmd.AddCommand(helpCmd)
}

func addStreamCommands() {
	streamCmd := &cobra.Command{
		Use:   "streams",
		Short: "Commands for managing and monitoring Redis streams",
	}

	// Add stream subcommands
	listStreamsCmd, err := NewListStreamsCommand()
	if err != nil {
		log.Fatalf("Failed to create list-streams command: %v", err)
	}
	cobraListStreamsCmd, err := cli.BuildCobraCommandFromCommand(listStreamsCmd)
	if err != nil {
		log.Fatalf("Failed to build Cobra command for list-streams: %v", err)
	}
	streamCmd.AddCommand(cobraListStreamsCmd)

	streamInfoCmd, err := NewStreamInfoCommand()
	if err != nil {
		log.Fatalf("Failed to create stream-info command: %v", err)
	}
	cobraStreamInfoCmd, err := cli.BuildCobraCommandFromCommand(streamInfoCmd)
	if err != nil {
		log.Fatalf("Failed to build Cobra command for stream-info: %v", err)
	}
	streamCmd.AddCommand(cobraStreamInfoCmd)

	rootCmd.AddCommand(streamCmd)
}

func addGroupCommands() {
	groupCmd := &cobra.Command{
		Use:   "groups",
		Short: "Commands for monitoring Redis stream consumer groups",
	}

	listGroupsCmd, err := NewListGroupsCommand()
	if err != nil {
		log.Fatalf("Failed to create list-groups command: %v", err)
	}
	cobraListGroupsCmd, err := cli.BuildCobraCommandFromCommand(listGroupsCmd)
	if err != nil {
		log.Fatalf("Failed to build Cobra command for list-groups: %v", err)
	}
	groupCmd.AddCommand(cobraListGroupsCmd)

	groupInfoCmd, err := NewGroupInfoCommand()
	if err != nil {
		log.Fatalf("Failed to create group-info command: %v", err)
	}
	cobraGroupInfoCmd, err := cli.BuildCobraCommandFromCommand(groupInfoCmd)
	if err != nil {
		log.Fatalf("Failed to build Cobra command for group-info: %v", err)
	}
	groupCmd.AddCommand(cobraGroupInfoCmd)

	rootCmd.AddCommand(groupCmd)
}

func addMetricsCommands() {
	metricsCmd := &cobra.Command{
		Use:   "metrics",
		Short: "Commands for monitoring Redis performance metrics",
	}

	memoryCmd, err := NewMemoryCommand()
	if err != nil {
		log.Fatalf("Failed to create memory command: %v", err)
	}
	cobraMemoryCmd, err := cli.BuildCobraCommandFromCommand(memoryCmd)
	if err != nil {
		log.Fatalf("Failed to build Cobra command for memory: %v", err)
	}
	metricsCmd.AddCommand(cobraMemoryCmd)

	throughputCmd, err := NewThroughputCommand()
	if err != nil {
		log.Fatalf("Failed to create throughput command: %v", err)
	}
	cobraThroughputCmd, err := cli.BuildCobraCommandFromCommand(throughputCmd)
	if err != nil {
		log.Fatalf("Failed to build Cobra command for throughput: %v", err)
	}
	metricsCmd.AddCommand(cobraThroughputCmd)

	rootCmd.AddCommand(metricsCmd)
}

func addTUICommands() {
	tuiCmd, err := NewTUICommand()
	if err != nil {
		log.Fatalf("Failed to create TUI command: %v", err)
	}
	cobraTUICmd, err := cli.BuildCobraCommandFromCommand(tuiCmd)
	if err != nil {
		log.Fatalf("Failed to build Cobra command for TUI: %v", err)
	}
	rootCmd.AddCommand(cobraTUICmd)
}

func addDemoCommand() {
	demoCmd, err := NewDemoCommand()
	if err != nil {
		log.Fatalf("Failed to create demo command: %v", err)
	}
	cobraDemoCmd, err := cli.BuildCobraCommandFromCommand(demoCmd)
	if err != nil {
		log.Fatalf("Failed to build Cobra command for demo: %v", err)
	}
	rootCmd.AddCommand(cobraDemoCmd)
}

func main() {
	ctx := context.Background()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
