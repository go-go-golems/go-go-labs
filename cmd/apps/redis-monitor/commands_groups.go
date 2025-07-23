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

// ListGroupsCommand lists all consumer groups across all streams
type ListGroupsCommand struct {
	*cmds.CommandDescription
}

type ListGroupsSettings struct {
	RedisAddr     string `glazed.parameter:"redis-addr"`
	RedisPassword string `glazed.parameter:"redis-password"`
	RedisDB       int    `glazed.parameter:"redis-db"`
	StreamFilter  string `glazed.parameter:"stream-filter"`
}

func NewListGroupsCommand() (*ListGroupsCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create glazed parameter layer: %w", err)
	}

	return &ListGroupsCommand{
		CommandDescription: cmds.NewCommandDescription(
			"list",
			cmds.WithShort("List all consumer groups across all streams"),
			cmds.WithLong(`List all consumer groups across all streams in the Redis database.

This command discovers all streams and their consumer groups, providing information about:
- Group name and associated stream
- Number of consumers in the group
- Number of pending messages
- Last delivered message ID

Use --stream-filter to limit results to a specific stream.`),
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
					"stream-filter",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Filter groups by stream name"),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

func (c *ListGroupsCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ListGroupsSettings{}
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
		// Apply stream filter if specified
		if s.StreamFilter != "" && streamName != s.StreamFilter {
			continue
		}

		groups, err := client.GetStreamGroups(ctx, streamName)
		if err != nil {
			// Continue with other streams if one fails
			continue
		}

		for _, group := range groups {
			row := types.NewRow(
				types.MRP("group_name", group.Name),
				types.MRP("stream", group.Stream),
				types.MRP("consumers", group.Consumers),
				types.MRP("pending", group.Pending),
				types.MRP("last_delivered_id", group.LastDeliveredID),
				types.MRP("timestamp", time.Now().Format(time.RFC3339)),
			)

			if err := gp.AddRow(ctx, row); err != nil {
				return fmt.Errorf("failed to add row: %w", err)
			}
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &ListGroupsCommand{}

// GroupInfoCommand shows detailed information about consumers in a specific group
type GroupInfoCommand struct {
	*cmds.CommandDescription
}

type GroupInfoSettings struct {
	RedisAddr     string `glazed.parameter:"redis-addr"`
	RedisPassword string `glazed.parameter:"redis-password"`
	RedisDB       int    `glazed.parameter:"redis-db"`
	StreamName    string `glazed.parameter:"stream"`
	GroupName     string `glazed.parameter:"group"`
}

func NewGroupInfoCommand() (*GroupInfoCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, fmt.Errorf("could not create glazed parameter layer: %w", err)
	}

	return &GroupInfoCommand{
		CommandDescription: cmds.NewCommandDescription(
			"info",
			cmds.WithShort("Show detailed information about consumers in a group"),
			cmds.WithLong(`Show detailed information about all consumers in a specific consumer group.

This command provides comprehensive information about consumers including:
- Consumer name and status
- Number of pending messages per consumer
- Idle time (how long since last activity)
- Group and stream association

Use this command to identify slow or stuck consumers in your streaming application.`),
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
					parameters.WithHelp("Name of the stream"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"group",
					parameters.ParameterTypeString,
					parameters.WithHelp("Name of the consumer group"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayersList(glazedLayer),
		),
	}, nil
}

func (c *GroupInfoCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &GroupInfoSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to initialize settings: %w", err)
	}

	client := NewRedisClient(s.RedisAddr, s.RedisPassword, s.RedisDB)
	defer client.Close()

	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	consumers, err := client.GetGroupConsumers(ctx, s.StreamName, s.GroupName)
	if err != nil {
		return fmt.Errorf("failed to get group consumers: %w", err)
	}

	for _, consumer := range consumers {
		row := types.NewRow(
			types.MRP("consumer_name", consumer.Name),
			types.MRP("group", consumer.Group),
			types.MRP("stream", consumer.Stream),
			types.MRP("pending", consumer.Pending),
			types.MRP("idle_ms", consumer.Idle.Milliseconds()),
			types.MRP("idle_formatted", FormatDuration(consumer.Idle)),
			types.MRP("timestamp", time.Now().Format(time.RFC3339)),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return fmt.Errorf("failed to add row: %w", err)
		}
	}

	return nil
}

var _ cmds.GlazeCommand = &GroupInfoCommand{}
