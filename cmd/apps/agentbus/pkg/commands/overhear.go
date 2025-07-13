package commands

import (
	"context"
	"fmt"
	"io"
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
	"github.com/rs/zerolog/log"
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
var _ cmds.WriterCommand = (*OverhearCommand)(nil)

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
		log.Error().Err(err).Msg("Failed to initialize overhear settings")
		return err
	}

	log.Info().
		Str("since", s.Since).
		Bool("follow", s.Follow).
		Int("max", s.Max).
		Str("topic", s.Topic).
		Msg("Starting overhear operation")

	agentID, err := getAgentID()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get agent ID")
		return err
	}

	log.Debug().Str("agent_id", agentID).Msg("Retrieved agent ID")

	client, err := getRedisClient()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Redis client")
		return err
	}
	defer client.Close()

	streamKey := client.ChannelKey()
	lastKey := client.LastKey(agentID)

	// Determine starting position and track what's new
	startID := s.Since
	lastReadPosition := ""
	if startID == "" {
		log.Debug().Str("agent_id", agentID).Msg("Getting last read position for agent")
		// Get last read position for this agent
		lastID, err := client.Get(ctx, lastKey).Result()
		if err == redis.Nil {
			log.Debug().Str("agent_id", agentID).Msg("No previous read position found, starting from beginning")
			startID = "0" // Start from beginning if never read before
		} else if err != nil {
			log.Error().Err(err).Str("agent_id", agentID).Msg("Failed to get last read position")
			return errors.Wrap(err, "failed to get last read position")
		} else {
			lastReadPosition = lastID
			startID = lastID
			log.Debug().
				Str("agent_id", agentID).
				Str("last_position", lastID).
				Msg("Retrieved last read position")
		}
	} else {
		log.Debug().Str("since", s.Since).Msg("Using explicit start position")
	}

	var messages []redis.XMessage
	if s.Follow {
		log.Debug().
			Int("max", s.Max).
			Str("timeout", "1 minute").
			Msg("Blocking for new messages")
		// Block for new messages
		result, err := client.XRead(ctx, &redis.XReadArgs{
			Streams: []string{streamKey, "$"}, // $ means only new messages
			Block:   time.Minute,              // 1 minute timeout
			Count:   int64(s.Max),
		}).Result()
		if err != nil && err != redis.Nil {
			log.Error().Err(err).Msg("Failed to read messages in follow mode")
			return errors.Wrap(err, "failed to read messages")
		}
		if len(result) > 0 && len(result[0].Messages) > 0 {
			messages = result[0].Messages
			log.Debug().Int("message_count", len(messages)).Msg("Retrieved messages in follow mode")
		} else {
			log.Debug().Msg("No new messages received in follow mode")
		}
	} else {
		log.Debug().
			Str("start_id", startID).
			Int("max", s.Max).
			Msg("Reading messages from stream")
		// Read messages since startID
		result, err := client.XRange(ctx, streamKey, startID, "+").Result()
		if err != nil {
			log.Error().Err(err).Str("start_id", startID).Msg("Failed to read messages from stream")
			return errors.Wrap(err, "failed to read messages")
		}

		log.Debug().Int("raw_count", len(result)).Str("start_id", startID).Msg("Retrieved raw messages")

		// Skip the first message if it's exactly the startID (already read)
		if len(result) > 0 && result[0].ID == startID {
			result = result[1:]
			log.Debug().Msg("Skipped first message as it matches start ID")
		}

		// Limit results
		if len(result) > s.Max {
			result = result[:s.Max]
			log.Debug().Int("limited_to", s.Max).Msg("Limited message results")
		}

		messages = result
		log.Debug().Int("final_count", len(messages)).Msg("Final message count after processing")
	}

	// Count total messages and new messages
	newMessageCount := 0
	filteredMessages := make([]redis.XMessage, 0, len(messages))

	log.Debug().
		Int("total_messages", len(messages)).
		Str("topic_filter", s.Topic).
		Msg("Filtering messages")

	// Pre-filter and count
	for _, msg := range messages {
		// Filter by topic if specified
		if s.Topic != "" {
			msgTopic, _ := msg.Values["topic"].(string)
			if msgTopic != s.Topic {
				log.Debug().
					Str("msg_topic", msgTopic).
					Str("filter_topic", s.Topic).
					Str("message_id", msg.ID).
					Msg("Skipping message due to topic filter")
				continue
			}
		}
		filteredMessages = append(filteredMessages, msg)

		// Count as new if after last read position
		if lastReadPosition == "" || msg.ID > lastReadPosition {
			newMessageCount++
		}
	}

	log.Debug().
		Int("filtered_count", len(filteredMessages)).
		Int("new_count", newMessageCount).
		Msg("Message filtering complete")

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
		log.Debug().
			Str("agent_id", agentID).
			Str("last_read_id", lastReadID).
			Msg("Updating last read position")
		err = client.Set(ctx, lastKey, lastReadID, 0).Err()
		if err != nil {
			log.Error().Err(err).
				Str("agent_id", agentID).
				Str("last_read_id", lastReadID).
				Msg("Failed to update last read position")
			return errors.Wrap(err, "failed to update last read position")
		}
		log.Debug().
			Str("agent_id", agentID).
			Str("last_read_id", lastReadID).
			Msg("Successfully updated last read position")
	}

	log.Info().
		Int("total_messages", len(filteredMessages)).
		Int("new_messages", newMessageCount).
		Str("topic_filter", s.Topic).
		Bool("follow_mode", s.Follow).
		Msg("Successfully completed overhear operation")

	return nil
}

func (c *OverhearCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
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

	// Determine starting position
	startPosition := "0"
	if s.Since != "" {
		startPosition = s.Since
	} else {
		lastPos, err := client.Get(ctx, fmt.Sprintf("last_read:%s", agentID)).Result()
		if err == nil && lastPos != "" {
			startPosition = lastPos
		}
	}

	log.Info().
		Str("agent_id", agentID).
		Str("start_position", startPosition).
		Int("max_messages", s.Max).
		Str("topic_filter", s.Topic).
		Bool("follow_mode", s.Follow).
		Msg("Starting to overhear messages")

	newMessageCount := 0
	var lastMessageID string

	for {
		args := &redis.XReadArgs{
			Streams: []string{streamKey, startPosition},
			Count:   int64(s.Max),
			Block:   time.Duration(0),
		}

		if s.Follow {
			args.Block = 5 * time.Second
		}

		result, err := client.XRead(ctx, args).Result()
		if err != nil {
			if err == redis.Nil && s.Follow {
				continue
			}
			return errors.Wrap(err, "failed to read from stream")
		}

		if len(result) == 0 || len(result[0].Messages) == 0 {
			if s.Follow {
				continue
			}
			break
		}

		messages := result[0].Messages
		filteredMessages := make([]redis.XMessage, 0, len(messages))

		for _, msg := range messages {
			if s.Topic != "" {
				if topic, ok := msg.Values["topic"].(string); !ok || topic != s.Topic {
					continue
				}
			}
			filteredMessages = append(filteredMessages, msg)
		}

		// Output messages in human-readable format
		for _, msg := range filteredMessages {
			agentIDValue := ""
			if v, ok := msg.Values["agent_id"]; ok {
				agentIDValue = fmt.Sprintf("%v", v)
			}

			message := ""
			if v, ok := msg.Values["message"]; ok {
				message = fmt.Sprintf("%v", v)
			}

			topic := ""
			if v, ok := msg.Values["topic"]; ok {
				topic = fmt.Sprintf("%v", v)
			}

			// Parse timestamp from stream ID
			parts := strings.Split(msg.ID, "-")
			if len(parts) >= 1 {
				if timestampMs, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
					timestamp := time.Unix(timestampMs/1000, (timestampMs%1000)*1000000)
					timeStr := timestamp.Format("15:04:05")

					if topic != "" {
						fmt.Fprintf(w, "👂 [%s] %s in #%s: %s\n", timeStr, agentIDValue, topic, message)
					} else {
						fmt.Fprintf(w, "👂 [%s] %s: %s\n", timeStr, agentIDValue, message)
					}
				}
			}

			lastMessageID = msg.ID
			newMessageCount++
		}

		if len(filteredMessages) > 0 {
			startPosition = lastMessageID
		}

		if !s.Follow {
			break
		}
	}

	// Update last read position if we read any new messages
	if newMessageCount > 0 && lastMessageID != "" {
		err = client.Set(ctx, fmt.Sprintf("last_read:%s", agentID), lastMessageID, 0).Err()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to update last read position")
		}
	}

	if newMessageCount == 0 && !s.Follow {
		fmt.Fprintf(w, "🔇 No new messages to hear\n")
	}

	return nil
}
