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
)

// DemoCommand provides sample data for testing the Redis monitor without a real Redis instance
type DemoCommand struct {
	*cmds.CommandDescription
}

type DemoSettings struct {
	DataType string `glazed.parameter:"data-type"`
}

func NewDemoCommand() (*DemoCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &DemoCommand{
		CommandDescription: cmds.NewCommandDescription(
			"demo",
			cmds.WithShort("Generate demo data for testing Redis monitor features"),
			cmds.WithLong(`Generate sample Redis streams data for testing and demonstration.

This command creates realistic-looking Redis streams data without requiring
a Redis connection. Useful for:
- Testing the CLI output formats
- Demonstrating the monitoring capabilities
- Development and debugging

Data types available:
- streams: Sample stream information
- groups: Sample consumer group data
- consumers: Sample consumer details
- memory: Sample memory usage data
- throughput: Sample throughput metrics`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"data-type",
					parameters.ParameterTypeChoice,
					parameters.WithChoices("streams", "groups", "consumers", "memory", "throughput"),
					parameters.WithDefault("streams"),
					parameters.WithHelp("Type of demo data to generate"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

func (c *DemoCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &DemoSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	switch s.DataType {
	case "streams":
		return c.generateStreamsData(ctx, gp)
	case "groups":
		return c.generateGroupsData(ctx, gp)
	case "consumers":
		return c.generateConsumersData(ctx, gp)
	case "memory":
		return c.generateMemoryData(ctx, gp)
	case "throughput":
		return c.generateThroughputData(ctx, gp)
	default:
		return c.generateStreamsData(ctx, gp)
	}
}

func (c *DemoCommand) generateStreamsData(ctx context.Context, gp middlewares.Processor) error {
	streams := []struct {
		name   string
		length int64
		memory int64
		groups int64
		lastID string
	}{
		{"orders", 1243592, 126177280, 3, "160123-7"},
		{"events", 98234, 9663488, 5, "160123-3"},
		{"logs", 5432100, 536870912, 1, "160122-9"},
		{"notifications", 45672, 4194304, 2, "160121-4"},
		{"analytics", 234567, 25165824, 4, "160120-1"},
	}

	for _, stream := range streams {
		row := types.NewRow(
			types.MRP("name", stream.name),
			types.MRP("length", stream.length),
			types.MRP("memory_bytes", stream.memory),
			types.MRP("memory_formatted", FormatBytes(stream.memory)),
			types.MRP("groups", stream.groups),
			types.MRP("last_id", stream.lastID),
			types.MRP("first_id", "160000-0"),
			types.MRP("radix_keys", stream.length/100),
			types.MRP("radix_nodes", stream.length/50),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func (c *DemoCommand) generateGroupsData(ctx context.Context, gp middlewares.Processor) error {
	groups := []struct {
		name     string
		stream   string
		consumers int64
		pending   int64
	}{
		{"cg-1", "orders", 2, 12},
		{"cg-2", "orders", 1, 0},
		{"cg-3", "orders", 1, 28},
		{"cg-A", "events", 2, 3},
		{"cg-B", "events", 1, 47},
		{"cg-logs", "logs", 1, 0},
		{"cg-notif", "notifications", 3, 5},
		{"cg-analytics", "analytics", 2, 15},
	}

	for _, group := range groups {
		row := types.NewRow(
			types.MRP("group_name", group.name),
			types.MRP("stream", group.stream),
			types.MRP("consumers", group.consumers),
			types.MRP("pending", group.pending),
			types.MRP("last_delivered_id", "160123-1"),
			types.MRP("timestamp", time.Now().Format(time.RFC3339)),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func (c *DemoCommand) generateConsumersData(ctx context.Context, gp middlewares.Processor) error {
	consumers := []struct {
		name    string
		group   string
		stream  string
		pending int64
		idle    time.Duration
	}{
		{"Alice", "cg-1", "orders", 3, 5 * time.Second},
		{"Bob", "cg-1", "orders", 2, 1 * time.Second},
		{"Charlie", "cg-2", "orders", 5, 10 * time.Second},
		{"Dave", "cg-3", "orders", 1, 2*time.Minute + 12*time.Second},
		{"Eve", "cg-A", "events", 4, 0},
		{"Frank", "cg-A", "events", 4, 30 * time.Second},
		{"Grace", "cg-B", "events", 2, 1*time.Minute + 23*time.Second},
		{"Heidi", "cg-logs", "logs", 10, 0},
	}

	for _, consumer := range consumers {
		row := types.NewRow(
			types.MRP("consumer_name", consumer.name),
			types.MRP("group", consumer.group),
			types.MRP("stream", consumer.stream),
			types.MRP("pending", consumer.pending),
			types.MRP("idle_ms", consumer.idle.Milliseconds()),
			types.MRP("idle_formatted", FormatDuration(consumer.idle)),
			types.MRP("timestamp", time.Now().Format(time.RFC3339)),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

func (c *DemoCommand) generateMemoryData(ctx context.Context, gp middlewares.Processor) error {
	// Server memory info
	serverRow := types.NewRow(
		types.MRP("type", "server"),
		types.MRP("name", "Redis Server"),
		types.MRP("memory_bytes", int64(348127232)),
		types.MRP("memory_formatted", FormatBytes(348127232)),
		types.MRP("memory_rss_bytes", int64(385875968)),
		types.MRP("memory_rss_formatted", FormatBytes(385875968)),
		types.MRP("total_system_memory", FormatBytes(1073741824)), // 1GB
		types.MRP("memory_utilization_percent", "33.2%"),
		types.MRP("uptime_seconds", int64(1098276)), // ~12.7 days
		types.MRP("uptime_formatted", FormatDuration(1098276*time.Second)),
		types.MRP("redis_version", "7.0.8"),
		types.MRP("timestamp", time.Now().Format(time.RFC3339)),
	)

	if err := gp.AddRow(ctx, serverRow); err != nil {
		return err
	}

	return nil
}

func (c *DemoCommand) generateThroughputData(ctx context.Context, gp middlewares.Processor) error {
	// Generate 5 measurements over time
	baseTime := time.Now().Add(-4 * time.Minute)
	
	measurements := []struct {
		xaddRate      float64
		xreadRate     float64
		xaddTotal     int64
		xreadTotal    int64
	}{
		{145.2, 156.8, 12456789, 13567890},
		{152.1, 164.3, 12457245, 13568384},
		{139.7, 148.2, 12457664, 13568976},
		{168.4, 175.9, 12458169, 13569735},
		{158.8, 167.1, 12458643, 13570442},
	}

	for i, m := range measurements {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		
		row := types.NewRow(
			types.MRP("timestamp", timestamp.Format(time.RFC3339)),
			types.MRP("xadd_calls_total", m.xaddTotal),
			types.MRP("xreadgroup_calls_total", m.xreadTotal),
			types.MRP("xadd_calls_per_sec", m.xaddRate),
			types.MRP("xreadgroup_calls_per_sec", m.xreadRate),
			types.MRP("total_calls_per_sec", m.xaddRate+m.xreadRate),
			types.MRP("interval_seconds", 60.0),
			types.MRP("measurement", i+1),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &DemoCommand{}
