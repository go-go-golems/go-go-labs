package redis

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// Client wraps redis.Client with agentbus-specific functionality
type Client struct {
	*redis.Client
	prefix string
}

// NewClient creates a new Redis client for agentbus
func NewClient(redisURL, projectPrefix string) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse redis URL")
	}

	// Configure timeouts to prevent hanging
	if opts.DialTimeout == 0 {
		opts.DialTimeout = 5 * time.Second
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = 10 * time.Second
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = 10 * time.Second
	}

	client := redis.NewClient(opts)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to redis")
	}

	return &Client{
		Client: client,
		prefix: "agentbus:" + projectPrefix + ":",
	}, nil
}

// Key returns a prefixed key for agentbus
func (c *Client) Key(parts ...string) string {
	key := c.prefix
	for _, part := range parts {
		key += part
	}
	return key
}

// ChannelKey returns a key for the shared communication channel
func (c *Client) ChannelKey() string {
	return c.Key("ch:main")
}

// LastKey returns a key for tracking last read position for the shared channel
func (c *Client) LastKey(agentID string) string {
	return c.Key("last:", agentID)
}

// JotKey returns a key for a knowledge snippet
func (c *Client) JotKey(title string) string {
	return c.Key("jot:", title)
}

// JotsByTagKey returns a key for jots by tag
func (c *Client) JotsByTagKey(tag string) string {
	return c.Key("jots_by_tag:", tag)
}

// FlagKey returns a key for a coordination flag
func (c *Client) FlagKey(name string) string {
	return c.Key("flag:", name)
}

// ChatMessage represents a message in the communication stream
type ChatMessage struct {
	ID        string    `json:"id"`
	AgentID   string    `json:"agent_id"`
	Topic     string    `json:"topic,omitempty"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Jot represents a knowledge snippet
type Jot struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Author    string    `json:"author"`
	Tags      []string  `json:"tags"`
	Timestamp time.Time `json:"timestamp"`
}

// Flag represents a coordination flag
type Flag struct {
	Name      string    `json:"name"`
	AgentID   string    `json:"agent_id"`
	Timestamp time.Time `json:"timestamp"`
}
