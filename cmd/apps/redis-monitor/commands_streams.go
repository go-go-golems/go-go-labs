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

// ListStreamsCommand lists all Redis streams
type ListStreamsCommand struct {
	*cmds.CommandDescription
}

type ListStreamsSettings struct {
	RedisAddr     string `glazed.parameter:"redis-addr"`
	RedisPassword string `glazed.parameter:"redis-password"`
	RedisDB       int    `glazed.parameter:"redis-db"`
}

func NewListStreamsCommand() (*ListStreamsCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create glazed parameter layer: %w", err)
	}

	return &ListStreamsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List all Redis streams with basic information"),
			cmds.WithLong(`List all Redis streams in the database with key metrics.

This command discovers all streams using Redis SCAN and provides information about:
- Stream name
- Number of entries
- Memory usage
- Number of consumer groups
- Last generated ID

The output can be sorted by different metrics and formatted in various ways.`),
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

			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

func (c *ListStreamsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ListStreamsSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to initialize settings: %w", err)
	}

	client := NewRedisClient(s.RedisAddr, s.RedisPassword, s.RedisDB)
	defer client.Close()

	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	streams, err := client.DiscoverStreams(ctx)
	if err != nil {
		return fmt.Errorf("failed to discover streams: %w", err)
	}

	for _, streamName := range streams {
		info, err := client.GetStreamInfo(ctx, streamName)
		if err != nil {
			// Continue with other streams if one fails
			continue
		}

		row := types.NewRow(
			types.MRP("name", info.Name),
			types.MRP("length", info.Length),
			types.MRP("memory_bytes", info.MemoryUsage),
			types.MRP("memory_formatted", FormatBytes(info.MemoryUsage)),
			types.MRP("groups", info.Groups),
			types.MRP("last_id", info.LastGeneratedID),
			types.MRP("first_id", info.FirstEntryID),
			types.MRP("radix_keys", info.RadixTreeKeys),
			types.MRP("radix_nodes", info.RadixTreeNodes),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &ListStreamsCommand{}

// StreamInfoCommand shows detailed information about a specific stream
type StreamInfoCommand struct {
	*cmds.CommandDescription
}

type StreamInfoSettings struct {
	RedisAddr     string `glazed.parameter:"redis-addr"`
	RedisPassword string `glazed.parameter:"redis-password"`
	RedisDB       int    `glazed.parameter:"redis-db"`
	StreamName    string `glazed.parameter:"stream"`
}

func NewStreamInfoCommand() (*StreamInfoCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create glazed parameter layer: %w", err)
	}

	return &StreamInfoCommand{
		CommandDescription: cmds.NewCommandDescription(
			"info",
			cmds.WithShort("Show detailed information about a Redis stream"),
			cmds.WithLong(`Show comprehensive information about a specific Redis stream.

This command provides detailed metrics for a single stream including:
- Stream metadata (length, memory usage)
- Radix tree structure details
- Consumer group information
- First and last entry IDs

Use this command to get in-depth information about stream performance and structure.`),
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
			),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"stream",
					parameters.ParameterTypeString,
					parameters.WithHelp("Name of the stream to inspect"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

func (c *StreamInfoCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &StreamInfoSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to initialize settings: %w", err)
	}

	client := NewRedisClient(s.RedisAddr, s.RedisPassword, s.RedisDB)
	defer client.Close()

	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	info, err := client.GetStreamInfo(ctx, s.StreamName)
	if err != nil {
		return fmt.Errorf("failed to get stream info: %w", err)
	}

	// Get groups for this stream
	groups, err := client.GetStreamGroups(ctx, s.StreamName)
	if err != nil {
		return fmt.Errorf("failed to get stream groups: %w", err)
	}

	var groupNames []string
	for _, group := range groups {
		groupNames = append(groupNames, group.Name)
	}

	row := types.NewRow(
		types.MRP("name", info.Name),
		types.MRP("length", info.Length),
		types.MRP("memory_bytes", info.MemoryUsage),
		types.MRP("memory_formatted", FormatBytes(info.MemoryUsage)),
		types.MRP("groups_count", info.Groups),
		types.MRP("groups_names", groupNames),
		types.MRP("last_generated_id", info.LastGeneratedID),
		types.MRP("first_entry_id", info.FirstEntryID),
		types.MRP("radix_tree_keys", info.RadixTreeKeys),
		types.MRP("radix_tree_nodes", info.RadixTreeNodes),
		types.MRP("timestamp", time.Now().Format(time.RFC3339)),
	)

	if err := gp.AddRow(ctx, row); err != nil {
		return fmt.Errorf("failed to add row: %w", err)
	}

	return nil
}

var _ cmds.GlazeCommand = &StreamInfoCommand{}
