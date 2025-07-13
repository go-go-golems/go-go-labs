package commands

import (
	"context"
	"fmt"
	"os"

	agentredis "github.com/go-go-golems/go-go-labs/cmd/apps/agentbus/pkg/redis"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
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

// getRedisClient creates a Redis client from configuration
func getRedisClient() (*agentredis.Client, error) {
	redisURL := viper.GetString("redis-url")
	return agentredis.NewClient(redisURL)
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
	streamKey := client.ChannelKey()
	values := map[string]interface{}{
		"agent_id": agentID,
		"message":  message,
	}
	if topic != "" {
		values["topic"] = topic
	}
	
	_, err := client.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: values,
	}).Result()
	
	return errors.Wrap(err, "failed to publish to channel")
}
