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
	"github.com/go-go-golems/glazed/pkg/settings"
	agentredis "github.com/go-go-golems/go-go-labs/cmd/apps/agentbus/pkg/redis"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type MonitorCommand struct {
	*cmds.CommandDescription
}

type MonitorSettings struct {
	Interval int  `glazed.parameter:"interval"`
	Follow   bool `glazed.parameter:"follow"`
}

var _ cmds.WriterCommand = (*MonitorCommand)(nil)

func NewMonitorCommand() (*MonitorCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create glazed parameter layer")
	}

	return &MonitorCommand{
		CommandDescription: cmds.NewCommandDescription(
			"monitor",
			cmds.WithShort("Monitor all AgentBus activity in real-time"),
			cmds.WithLong(`Monitor all AgentBus activity including messages, knowledge updates, and coordination flags in real-time.

This command provides a live view of all agent coordination activity:
- 💬 New messages in the communication stream
- 📝 Knowledge snippets being added/updated  
- 🚩 Coordination flags being announced
- ✅ Coordination flags being satisfied
- 📊 System statistics and health

The monitor runs continuously, showing timestamped events as they happen.
Use Ctrl+C to stop monitoring.

This is ideal for:
- Debugging agent coordination issues
- Understanding system-wide activity
- Monitoring deployment progress
- Observing agent communication patterns

Example usage:
  agentbus monitor
  agentbus monitor --interval 500 --follow`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"interval",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Polling interval in milliseconds"),
					parameters.WithDefault(1000),
				),
				parameters.NewParameterDefinition(
					"follow",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Follow mode - block and wait for new events"),
					parameters.WithDefault(true),
				),
			),
			cmds.WithLayersList(glazedParameterLayer),
		),
	}, nil
}

func (c *MonitorCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	s := &MonitorSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return err
	}

	client, err := getRedisClient()
	if err != nil {
		return err
	}
	defer client.Close()

	agentID, err := getAgentID()
	if err != nil {
		return err
	}

	log.Info().Str("agent_id", agentID).Msg("Starting AgentBus monitor")
	fmt.Fprintf(w, "🔍 AgentBus Monitor Started (Agent: %s)\n", agentID)
	fmt.Fprintln(w, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Initialize monitoring state
	streamKey := client.ChannelKey()
	lastStreamID := "$" // Start from newest messages
	knownFlags := make(map[string]string)

	// Get initial flag state
	flagPattern := client.FlagKey("*")
	initialFlags, err := client.Keys(ctx, flagPattern).Result()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get initial flags")
	} else {
		for _, flagKey := range initialFlags {
			flagName := strings.TrimPrefix(flagKey, client.FlagKey(""))
			flagValue, err := client.Get(ctx, flagKey).Result()
			if err == nil {
				knownFlags[flagName] = flagValue
				fmt.Fprintf(w, "🚩 EXISTING FLAG: %s = %s\n", flagName, flagValue)
			}
		}
	}

	ticker := time.NewTicker(time.Duration(s.Interval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Monitor stopped by context")
			return ctx.Err()
		case <-ticker.C:
			// Check for new messages
			err := c.checkNewMessages(ctx, client, streamKey, &lastStreamID, w)
			if err != nil {
				log.Error().Err(err).Msg("Error checking messages")
				fmt.Fprintf(w, "❌ Error checking messages: %v\n", err)
			}

			// Check for flag changes
			err = c.checkFlagChanges(ctx, client, knownFlags, w)
			if err != nil {
				log.Error().Err(err).Msg("Error checking flags")
				fmt.Fprintf(w, "❌ Error checking flags: %v\n", err)
			}

			// Print statistics periodically (every 30 seconds)
			if time.Now().Unix()%30 == 0 {
				c.printStatistics(ctx, client, w)
			}
		}
	}
}

func (c *MonitorCommand) checkNewMessages(ctx context.Context, client *agentredis.Client, streamKey string, lastStreamID *string, w io.Writer) error {
	// Read new messages since last check
	result, err := client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{streamKey, *lastStreamID},
		Count:   100,
		Block:   100 * time.Millisecond, // Short block to not delay other checks
	}).Result()

	if err != nil && err != redis.Nil {
		return err
	}

	if len(result) == 0 || len(result[0].Messages) == 0 {
		return nil // No new messages
	}

	messages := result[0].Messages
	for _, msg := range messages {
		timestamp := c.parseTimestampFromStreamID(msg.ID)
		agentID, _ := msg.Values["agent_id"].(string)
		message, _ := msg.Values["message"].(string)
		topic, _ := msg.Values["topic"].(string)

		topicStr := ""
		if topic != "" {
			topicStr = fmt.Sprintf(" [%s]", topic)
		}

		fmt.Fprintf(w, "💬 %s | %s%s: %s\n",
			timestamp.Format("15:04:05"),
			agentID,
			topicStr,
			message)

		*lastStreamID = msg.ID
		log.Debug().
			Str("agent_id", agentID).
			Str("topic", topic).
			Str("message", message).
			Msg("New message monitored")
	}

	return nil
}

func (c *MonitorCommand) checkFlagChanges(ctx context.Context, client *agentredis.Client, knownFlags map[string]string, w io.Writer) error {
	flagPattern := client.FlagKey("*")
	currentFlags, err := client.Keys(ctx, flagPattern).Result()
	if err != nil {
		return err
	}

	// Check for new/changed flags
	currentFlagMap := make(map[string]string)
	for _, flagKey := range currentFlags {
		flagName := strings.TrimPrefix(flagKey, client.FlagKey(""))
		flagValue, err := client.Get(ctx, flagKey).Result()
		if err != nil {
			continue
		}
		currentFlagMap[flagName] = flagValue

		if oldValue, exists := knownFlags[flagName]; !exists {
			// New flag
			fmt.Fprintf(w, "🚩 %s | NEW FLAG: %s = %s\n",
				time.Now().Format("15:04:05"),
				flagName,
				flagValue)
			log.Info().Str("flag", flagName).Str("value", flagValue).Msg("New flag announced")
		} else if oldValue != flagValue {
			// Changed flag
			fmt.Fprintf(w, "🔄 %s | CHANGED FLAG: %s = %s (was: %s)\n",
				time.Now().Format("15:04:05"),
				flagName,
				flagValue,
				oldValue)
			log.Info().Str("flag", flagName).Str("old_value", oldValue).Str("new_value", flagValue).Msg("Flag changed")
		}
	}

	// Check for satisfied (deleted) flags
	for flagName, flagValue := range knownFlags {
		if _, exists := currentFlagMap[flagName]; !exists {
			fmt.Fprintf(w, "✅ %s | SATISFIED FLAG: %s (was: %s)\n",
				time.Now().Format("15:04:05"),
				flagName,
				flagValue)
			log.Info().Str("flag", flagName).Str("value", flagValue).Msg("Flag satisfied")
		}
	}

	// Update known flags
	for k := range knownFlags {
		delete(knownFlags, k)
	}
	for k, v := range currentFlagMap {
		knownFlags[k] = v
	}

	return nil
}

func (c *MonitorCommand) printStatistics(ctx context.Context, client *agentredis.Client, w io.Writer) {
	streamKey := client.ChannelKey()

	// Get stream length
	streamLen, err := client.XLen(ctx, streamKey).Result()
	if err != nil {
		streamLen = -1
	}

	// Count flags
	flagPattern := client.FlagKey("*")
	flags, err := client.Keys(ctx, flagPattern).Result()
	flagCount := len(flags)
	if err != nil {
		flagCount = -1
	}

	// Count jots
	jotPattern := client.JotKey("*")
	jots, err := client.Keys(ctx, jotPattern).Result()
	jotCount := len(jots)
	if err != nil {
		jotCount = -1
	}

	fmt.Fprintf(w, "📊 %s | STATS: %d messages, %d active flags, %d knowledge snippets\n",
		time.Now().Format("15:04:05"),
		streamLen,
		flagCount,
		jotCount)

	log.Debug().
		Int64("messages", streamLen).
		Int("flags", flagCount).
		Int("jots", jotCount).
		Msg("Statistics reported")
}

func (c *MonitorCommand) parseTimestampFromStreamID(streamID string) time.Time {
	parts := strings.Split(streamID, "-")
	if len(parts) < 1 {
		return time.Now()
	}

	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Now()
	}

	return time.Unix(timestamp/1000, 0)
}
