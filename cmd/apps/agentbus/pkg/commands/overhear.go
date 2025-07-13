package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type OverhearCommand struct {
	*cmds.CommandDescription
}

type OverhearSettings struct {
	Since  string `glazed.parameter:"since"`
	Follow bool   `glazed.parameter:"follow"`
	Max    int    `glazed.parameter:"max"`
	Topic  string `glazed.parameter:"topic"`
}

var _ cmds.GlazeCommand = (*OverhearCommand)(nil)

func NewOverhearCommand() (*OverhearCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	return &OverhearCommand{
		CommandDescription: cmds.NewCommandDescription(
			"overhear",
			cmds.WithShort("Listen for messages from the shared communication channel"),
			cmds.WithLong(`Listen for messages from the shared Redis-backed communication stream.

This command uses a pull model where each agent tracks its own read position.
Messages are retrieved from the shared Redis Stream starting from the last position
this agent read, ensuring no agent starves another.

Optionally filter by topic to only see messages with specific topic slugs.

Modes:
- Default: Read new messages since last time (one-shot)
- --since <id>: Read messages after specific stream ID  
- --follow: Block and wait for new messages (good for polling)

This is ideal for:
- Monitoring status updates from other agents
- Waiting for specific notifications
- Checking for error reports or alerts
- Staying informed about system state changes

Example usage in agent tool calling:
  agentbus overhear --max 10
  agentbus overhear --topic "build" --follow
  agentbus overhear --topic "deploy" --since "1234567890-0"`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"since",
					parameters.ParameterTypeString,
					parameters.WithHelp("Start reading from this stream ID (default: last read position)"),
				),
				parameters.NewParameterDefinition(
					"follow",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Block until new messages arrive"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"max",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Maximum number of messages to return"),
					parameters.WithDefault(100),
				),
				parameters.NewParameterDefinition(
					"topic",
					parameters.ParameterTypeString,
					parameters.WithHelp("Filter messages by topic slug"),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *OverhearCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &OverhearSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	agentID, err := getAgentID()
	if err != nil {
		return err
	}

	client, err := getRedisClient()
	if err != nil {
		return err
	}
	defer client.Close()

	streamKey := client.ChannelKey()
	lastKey := client.LastKey(agentID)

	// Determine starting position and track what's new
	startID := s.Since
	lastReadPosition := ""
	if startID == "" {
		// Get last read position for this agent
		lastID, err := client.Get(ctx, lastKey).Result()
		if err == redis.Nil {
			startID = "0" // Start from beginning if never read before
		} else if err != nil {
			return errors.Wrap(err, "failed to get last read position")
		} else {
			lastReadPosition = lastID
			startID = lastID
		}
	}

	var messages []redis.XMessage
	if s.Follow {
		// Block for new messages
		result, err := client.XRead(ctx, &redis.XReadArgs{
			Streams: []string{streamKey, "$"}, // $ means only new messages
			Block:   time.Minute,              // 1 minute timeout
			Count:   int64(s.Max),
		}).Result()
		if err != nil && err != redis.Nil {
			return errors.Wrap(err, "failed to read messages")
		}
		if len(result) > 0 && len(result[0].Messages) > 0 {
			messages = result[0].Messages
		}
	} else {
		// Read messages since startID
		result, err := client.XRange(ctx, streamKey, startID, "+").Result()
		if err != nil {
			return errors.Wrap(err, "failed to read messages")
		}

		// Skip the first message if it's exactly the startID (already read)
		if len(result) > 0 && result[0].ID == startID {
			result = result[1:]
		}

		// Limit results
		if len(result) > s.Max {
			result = result[:s.Max]
		}

		messages = result
	}

	// Count total messages and new messages
	newMessageCount := 0
	filteredMessages := make([]redis.XMessage, 0, len(messages))

	// Pre-filter and count
	for _, msg := range messages {
		// Filter by topic if specified
		if s.Topic != "" {
			msgTopic, _ := msg.Values["topic"].(string)
			if msgTopic != s.Topic {
				continue
			}
		}
		filteredMessages = append(filteredMessages, msg)

		// Count as new if after last read position
		if lastReadPosition == "" || msg.ID > lastReadPosition {
			newMessageCount++
		}
	}

	// Add summary header as first row
	summaryRow := types.NewRow(
		types.MRP("stream_id", "SUMMARY"),
		types.MRP("agent_id", "system"),
		types.MRP("topic", "info"),
		types.MRP("message", fmt.Sprintf("Found %d total messages, %d new since last read", len(filteredMessages), newMessageCount)),
		types.MRP("timestamp", time.Now().Format(time.RFC3339)),
	)
	err = gp.AddRow(ctx, summaryRow)
	if err != nil {
		return err
	}

	// Process and output messages
	var lastReadID string
	for _, msg := range filteredMessages {
		// Parse timestamp from stream ID
		timestampStr := msg.ID[:strings.Index(msg.ID, "-")]
		timestamp, _ := strconv.ParseInt(timestampStr, 10, 64)

		topic, _ := msg.Values["topic"].(string)

		// Determine if this message is new
		isNew := lastReadPosition == "" || msg.ID > lastReadPosition
		messageText := msg.Values["message"].(string)
		if isNew {
			messageText = "NEW: " + messageText
		}

		row := types.NewRow(
			types.MRP("stream_id", msg.ID),
			types.MRP("agent_id", msg.Values["agent_id"]),
			types.MRP("topic", topic),
			types.MRP("message", messageText),
			types.MRP("timestamp", time.Unix(timestamp/1000, 0).Format(time.RFC3339)),
		)

		err = gp.AddRow(ctx, row)
		if err != nil {
			return err
		}

		lastReadID = msg.ID
	}

	// Update last read position if we read any messages
	if lastReadID != "" {
		err = client.Set(ctx, lastKey, lastReadID, 0).Err()
		if err != nil {
			return errors.Wrap(err, "failed to update last read position")
		}
	}

	return nil
}
