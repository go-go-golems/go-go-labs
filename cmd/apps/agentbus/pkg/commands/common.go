package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	agentredis "github.com/go-go-golems/go-go-labs/cmd/apps/agentbus/pkg/redis"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// getAgentID retrieves the agent ID from flag or environment
func getAgentID() (string, error) {
	agentID := viper.GetString("agent")
	if agentID == "" {
		return "", errors.New("AGENT_ID is required (use --agent flag or AGENT_ID env var)")
	}
	return agentID, nil
}

// getProjectPrefix retrieves the project prefix from environment
func getProjectPrefix() (string, error) {
	projectPrefix := viper.GetString("project-prefix")
	if projectPrefix == "" {
		return "", errors.New("PROJECT_PREFIX is required (use --project-prefix flag or PROJECT_PREFIX env var)")
	}
	return projectPrefix, nil
}

// getRedisClient creates a Redis client from configuration
func getRedisClient() (*agentredis.Client, error) {
	redisURL := viper.GetString("redis-url")
	projectPrefix, err := getProjectPrefix()
	if err != nil {
		return nil, err
	}

	log.Debug().Str("redis_url", redisURL).Str("project_prefix", projectPrefix).Msg("REDIS: Creating Redis client")

	start := time.Now()
	client, err := agentredis.NewClient(redisURL, projectPrefix)
	if err != nil {
		log.Error().Err(err).Str("redis_url", redisURL).Str("project_prefix", projectPrefix).Dur("duration", time.Since(start)).Msg("REDIS: Failed to create Redis client")
		return nil, err
	}

	log.Debug().Str("redis_url", redisURL).Str("project_prefix", projectPrefix).Dur("duration", time.Since(start)).Msg("REDIS: Successfully created Redis client")
	return client, nil
}

// getOutputFormat returns the output format
func getOutputFormat() string {
	format := viper.GetString("format")
	if format != "json" && format != "text" {
		fmt.Fprintf(os.Stderr, "Warning: invalid format '%s', using 'json'\n", format)
		return "json"
	}
	return format
}

// publishToChannel publishes a message to the shared communication channel
func publishToChannel(ctx context.Context, client *agentredis.Client, agentID, message, topic string) error {
	log.Debug().Str("agent_id", agentID).Str("message", message).Str("topic", topic).Msg("PUBLISH: Starting publish to channel")

	// Add timeout to the context
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	streamKey := client.ChannelKey()
	values := map[string]interface{}{
		"agent_id": agentID,
		"message":  message,
	}
	if topic != "" {
		values["topic"] = topic
	}

	log.Debug().Str("stream_key", streamKey).Interface("values", values).Msg("PUBLISH: Adding to Redis stream")

	start := time.Now()
	result, err := client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: values,
	}).Result()

	if err != nil {
		log.Error().Err(err).Str("stream_key", streamKey).Dur("duration", time.Since(start)).Msg("PUBLISH: Failed to publish to channel")
		return errors.Wrap(err, "failed to publish to channel")
	}

	log.Debug().Str("stream_id", result).Dur("duration", time.Since(start)).Msg("PUBLISH: Successfully published to channel")
	return nil
}

// showLatestMessages retrieves and displays the latest N messages from the communication stream
func showLatestMessages(ctx context.Context, client *agentredis.Client, w io.Writer, agentID string, numMessages int) error {
	log.Debug().Str("agent_id", agentID).Int("num_messages", numMessages).Msg("MESSAGES: Starting to show latest messages")

	// Add timeout to the context to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	streamKey := client.ChannelKey()
	log.Debug().Str("stream_key", streamKey).Msg("MESSAGES: Getting messages from Redis stream")

	// Get the latest messages using XREVRANGE (reverse order, newest first)
	start := time.Now()
	messages, err := client.XRevRange(ctx, streamKey, "+", "-").Result()
	if err != nil {
		log.Error().Err(err).Str("stream_key", streamKey).Dur("duration", time.Since(start)).Msg("MESSAGES: Failed to retrieve latest messages")
		return errors.Wrap(err, "failed to retrieve latest messages")
	}
	log.Debug().Int("message_count", len(messages)).Dur("duration", time.Since(start)).Msg("MESSAGES: Retrieved messages from Redis")

	// Limit to requested number of messages
	if len(messages) > numMessages {
		messages = messages[:numMessages]
		log.Debug().Int("limited_count", numMessages).Msg("MESSAGES: Limited message count")
	}

	// Show separator and header
	fmt.Fprintf(w, "\n--- Recent Messages ---\n")

	if len(messages) == 0 {
		log.Debug().Msg("MESSAGES: No messages in stream")
		fmt.Fprintf(w, "(No messages in the communication stream)\n")
		return nil
	}

	// Display messages (newest first)
	log.Debug().Int("displaying_count", len(messages)).Msg("MESSAGES: Displaying messages")
	for i, msg := range messages {
		log.Debug().Int("message_index", i).Str("message_id", msg.ID).Msg("MESSAGES: Processing message")

		// Parse timestamp from stream ID
		timestampStr := msg.ID[:strings.Index(msg.ID, "-")]
		timestamp, parseErr := strconv.ParseInt(timestampStr, 10, 64)
		if parseErr != nil {
			log.Warn().Err(parseErr).Str("timestamp_str", timestampStr).Msg("MESSAGES: Failed to parse timestamp")
			timestamp = 0
		}
		timeStr := time.Unix(timestamp/1000, 0).Format("15:04:05")

		// Extract message fields
		msgAgentID, _ := msg.Values["agent_id"].(string)
		message, _ := msg.Values["message"].(string)
		topic, _ := msg.Values["topic"].(string)

		// Format the message display
		if topic != "" {
			fmt.Fprintf(w, "👂 [%s] %s in #%s: %s\n", timeStr, msgAgentID, topic, message)
		} else {
			fmt.Fprintf(w, "👂 [%s] %s: %s\n", timeStr, msgAgentID, message)
		}
	}

	log.Debug().Int("displayed_count", len(messages)).Msg("MESSAGES: Completed showing latest messages")
	return nil
}
