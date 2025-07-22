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
)

// MemoryCommand shows Redis memory usage information
type MemoryCommand struct {
	*cmds.CommandDescription
}

type MemorySettings struct {
	RedisAddr     string `glazed.parameter:"redis-addr"`
	RedisPassword string `glazed.parameter:"redis-password"`
	RedisDB       int    `glazed.parameter:"redis-db"`
	PerStream     bool   `glazed.parameter:"per-stream"`
}

func NewMemoryCommand() (*MemoryCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create glazed parameter layer: %w", err)
	}

	return &MemoryCommand{
		CommandDescription: cmds.NewCommandDescription(
			"memory",
			cmds.WithShort("Show Redis memory usage information"),
			cmds.WithLong(`Show comprehensive Redis memory usage information.

This command provides memory metrics including:
- Total memory used by Redis
- RSS (Resident Set Size) memory
- System memory information
- Per-stream memory usage (with --per-stream flag)

Use --per-stream to get detailed memory usage for each stream.`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"redis-addr",
					parameters.ParameterTypeString,
					parameters.WithDefault("localhost:6379"),
					parameters.WithHelp("Redis server address"),
				),
				parameters.NewParameterDefinition(
					"redis-password",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Redis password"),
				),
				parameters.NewParameterDefinition(
					"redis-db",
					parameters.ParameterTypeInteger,
					parameters.WithDefault(0),
					parameters.WithHelp("Redis database number"),
				),
				parameters.NewParameterDefinition(
					"per-stream",
					parameters.ParameterTypeBool,
					parameters.WithDefault(false),
					parameters.WithHelp("Show memory usage per stream"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

func (c *MemoryCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &MemorySettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to initialize settings: %w", err)
	}

	client := NewRedisClient(s.RedisAddr, s.RedisPassword, s.RedisDB)
	defer client.Close()

	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	serverInfo, err := client.GetServerInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get server info: %w", err)
	}

	if s.PerStream {
		// Get per-stream memory usage
		streams, err := client.DiscoverStreams(ctx)
		if err != nil {
			return fmt.Errorf("failed to discover streams: %w", err)
		}

		var totalStreamMemory int64
		for _, streamName := range streams {
			info, err := client.GetStreamInfo(ctx, streamName)
			if err != nil {
				continue
			}

			totalStreamMemory += info.MemoryUsage
			memoryPercent := float64(info.MemoryUsage) / float64(serverInfo.UsedMemory) * 100

			row := types.NewRow(
				types.MRP("type", "stream"),
				types.MRP("name", info.Name),
				types.MRP("memory_bytes", info.MemoryUsage),
				types.MRP("memory_formatted", FormatBytes(info.MemoryUsage)),
				types.MRP("memory_percent", fmt.Sprintf("%.2f%%", memoryPercent)),
				types.MRP("entries", info.Length),
				types.MRP("timestamp", time.Now().Format(time.RFC3339)),
			)

			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to add row: %w", err)
			}
		}

		// Add total streams memory row
		streamPercent := float64(totalStreamMemory) / float64(serverInfo.UsedMemory) * 100
		row := types.NewRow(
			types.MRP("type", "total_streams"),
			types.MRP("name", "All Streams"),
			types.MRP("memory_bytes", totalStreamMemory),
			types.MRP("memory_formatted", FormatBytes(totalStreamMemory)),
			types.MRP("memory_percent", fmt.Sprintf("%.2f%%", streamPercent)),
			types.MRP("entries", len(streams)),
			types.MRP("timestamp", time.Now().Format(time.RFC3339)),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	// Add overall server memory information
	var memoryUtilization float64
	if serverInfo.TotalSystemMemory > 0 {
		memoryUtilization = float64(serverInfo.UsedMemoryRSS) / float64(serverInfo.TotalSystemMemory) * 100
	}

	serverRow := types.NewRow(
		types.MRP("type", "server"),
		types.MRP("name", "Redis Server"),
		types.MRP("memory_bytes", serverInfo.UsedMemory),
		types.MRP("memory_formatted", FormatBytes(serverInfo.UsedMemory)),
		types.MRP("memory_rss_bytes", serverInfo.UsedMemoryRSS),
		types.MRP("memory_rss_formatted", FormatBytes(serverInfo.UsedMemoryRSS)),
		types.MRP("total_system_memory", FormatBytes(serverInfo.TotalSystemMemory)),
		types.MRP("memory_utilization_percent", fmt.Sprintf("%.2f%%", memoryUtilization)),
		types.MRP("uptime_seconds", serverInfo.UptimeInSeconds),
		types.MRP("uptime_formatted", FormatDuration(time.Duration(serverInfo.UptimeInSeconds)*time.Second)),
		types.MRP("redis_version", serverInfo.Version),
		types.MRP("timestamp", time.Now().Format(time.RFC3339)),
	)

	if err := gp.AddRow(ctx, serverRow); err != nil {
		return fmt.Errorf("failed to add server row: %w", err)
	}

	return nil
}

var _ cmds.GlazeCommand = &MemoryCommand{}

// ThroughputCommand shows Redis throughput metrics
type ThroughputCommand struct {
	*cmds.CommandDescription
}

type ThroughputSettings struct {
	RedisAddr     string `glazed.parameter:"redis-addr"`
	RedisPassword string `glazed.parameter:"redis-password"`
	RedisDB       int    `glazed.parameter:"redis-db"`
	Interval      string `glazed.parameter:"interval"`
	Count         int    `glazed.parameter:"count"`
}

func NewThroughputCommand() (*ThroughputCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create glazed parameter layer: %w", err)
	}

	return &ThroughputCommand{
		CommandDescription: cmds.NewCommandDescription(
			"throughput",
			cmds.WithShort("Show Redis throughput metrics"),
			cmds.WithLong(`Show Redis throughput metrics by monitoring command statistics.

This command tracks Redis command execution rates including:
- XADD calls per second (stream writes)
- XREADGROUP calls per second (stream reads)
- Total message throughput

Use --interval to specify measurement intervals and --count to limit samples.`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"redis-addr",
					parameters.ParameterTypeString,
					parameters.WithDefault("localhost:6379"),
					parameters.WithHelp("Redis server address"),
				),
				parameters.NewParameterDefinition(
					"redis-password",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Redis password"),
				),
				parameters.NewParameterDefinition(
					"redis-db",
					parameters.ParameterTypeInteger,
					parameters.WithDefault(0),
					parameters.WithHelp("Redis database number"),
				),
				parameters.NewParameterDefinition(
					"interval",
					parameters.ParameterTypeString, // Duration will be parsed from string
					parameters.WithDefault("5s"),
					parameters.WithHelp("Measurement interval (e.g., 5s, 1m)"),
				),
				parameters.NewParameterDefinition(
					"count",
					parameters.ParameterTypeInteger,
					parameters.WithDefault(5),
					parameters.WithHelp("Number of measurements to take (0 for continuous)"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

func (c *ThroughputCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ThroughputSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to initialize settings: %w", err)
	}

	// Parse interval from string
	interval, err := time.ParseDuration(s.Interval)
	if err != nil {
		return fmt.Errorf("invalid interval format: %w", err)
	}

	client := NewRedisClient(s.RedisAddr, s.RedisPassword, s.RedisDB)
	defer client.Close()

	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Get initial throughput info
	prevInfo, err := client.GetThroughputInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get initial throughput info: %w", err)
	}

	measurements := 0
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			measurements++

			currentInfo, err := client.GetThroughputInfo(ctx)
			if err != nil {
				return fmt.Errorf("failed to get throughput info: %w", err)
			}

			// Calculate rates
			timeDiff := currentInfo.Timestamp.Sub(prevInfo.Timestamp).Seconds()
			if timeDiff <= 0 {
				continue
			}

			xaddRate := float64(currentInfo.XAddCalls-prevInfo.XAddCalls) / timeDiff
			xreadRate := float64(currentInfo.XReadGroupCalls-prevInfo.XReadGroupCalls) / timeDiff
			totalRate := xaddRate + xreadRate

			row := types.NewRow(
				types.MRP("timestamp", currentInfo.Timestamp.Format(time.RFC3339)),
				types.MRP("xadd_calls_total", currentInfo.XAddCalls),
				types.MRP("xreadgroup_calls_total", currentInfo.XReadGroupCalls),
				types.MRP("xadd_calls_per_sec", fmt.Sprintf("%.2f", xaddRate)),
				types.MRP("xreadgroup_calls_per_sec", fmt.Sprintf("%.2f", xreadRate)),
				types.MRP("total_calls_per_sec", fmt.Sprintf("%.2f", totalRate)),
				types.MRP("interval_seconds", fmt.Sprintf("%.2f", timeDiff)),
				types.MRP("measurement", measurements),
			)

			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to add row: %w", err)
			}

			prevInfo = currentInfo

			// Check if we should stop
			if s.Count > 0 && measurements >= s.Count {
				return nil
			}
		}
	}
}

var _ cmds.GlazeCommand = &ThroughputCommand{}
