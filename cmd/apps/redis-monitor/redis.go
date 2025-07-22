package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient wraps the Redis client with convenience methods for stream monitoring
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client with the given options
func NewRedisClient(addr, password string, db int) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisClient{client: rdb}
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Ping tests the connection to Redis
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// StreamInfo contains information about a Redis stream
type StreamInfo struct {
	Name             string
	Length           int64
	RadixTreeKeys    int64
	RadixTreeNodes   int64
	Groups           int64
	FirstEntryID     string
	LastGeneratedID  string
	MemoryUsage      int64
}

// GroupInfo contains information about a consumer group
type GroupInfo struct {
	Name      string
	Stream    string
	Consumers int64
	Pending   int64
	LastDeliveredID string
}

// ConsumerInfo contains information about a consumer
type ConsumerInfo struct {
	Name     string
	Group    string
	Stream   string
	Pending  int64
	Idle     time.Duration
}

// ServerInfo contains Redis server information
type ServerInfo struct {
	UptimeInSeconds   int64
	UsedMemory        int64
	UsedMemoryRSS     int64
	TotalSystemMemory int64
	Version           string
}

// ThroughputInfo contains throughput metrics
type ThroughputInfo struct {
	XAddCalls       int64
	XReadGroupCalls int64
	Timestamp       time.Time
}

// DiscoverStreams finds all Redis streams using SCAN
func (r *RedisClient) DiscoverStreams(ctx context.Context) ([]string, error) {
	var streams []string
	iter := r.client.Scan(ctx, 0, "*", 0).Iterator()
	
	for iter.Next(ctx) {
		key := iter.Val()
		keyType := r.client.Type(ctx, key).Val()
		if keyType == "stream" {
			streams = append(streams, key)
		}
	}
	
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("error scanning for streams: %w", err)
	}
	
	return streams, nil
}

// GetStreamInfo retrieves detailed information about a stream
func (r *RedisClient) GetStreamInfo(ctx context.Context, stream string) (*StreamInfo, error) {
	// Get basic stream info
	streamInfo := r.client.XInfoStream(ctx, stream).Val()
	
	// Get memory usage
	memUsage := r.client.MemoryUsage(ctx, stream).Val()
	
	info := &StreamInfo{
		Name:             stream,
		Length:           streamInfo.Length,
		RadixTreeKeys:    streamInfo.RadixTreeKeys,
		RadixTreeNodes:   streamInfo.RadixTreeNodes,
		Groups:           streamInfo.Groups,
		LastGeneratedID:  streamInfo.LastGeneratedID,
		MemoryUsage:      memUsage,
	}
	
	if len(streamInfo.FirstEntry.Values) > 0 {
		info.FirstEntryID = streamInfo.FirstEntry.ID
	}
	
	return info, nil
}

// GetStreamGroups lists all consumer groups for a stream
func (r *RedisClient) GetStreamGroups(ctx context.Context, stream string) ([]GroupInfo, error) {
	groups := r.client.XInfoGroups(ctx, stream).Val()
	
	var result []GroupInfo
	for _, group := range groups {
		result = append(result, GroupInfo{
			Name:            group.Name,
			Stream:          stream,
			Consumers:       group.Consumers,
			Pending:         group.Pending,
			LastDeliveredID: group.LastDeliveredID,
		})
	}
	
	return result, nil
}

// GetGroupConsumers lists all consumers in a group
func (r *RedisClient) GetGroupConsumers(ctx context.Context, stream, group string) ([]ConsumerInfo, error) {
	consumers := r.client.XInfoConsumers(ctx, stream, group).Val()
	
	var result []ConsumerInfo
	for _, consumer := range consumers {
		result = append(result, ConsumerInfo{
			Name:    consumer.Name,
			Group:   group,
			Stream:  stream,
			Pending: consumer.Pending,
			Idle:    time.Duration(consumer.Idle) * time.Millisecond,
		})
	}
	
	return result, nil
}

// GetServerInfo retrieves Redis server information
func (r *RedisClient) GetServerInfo(ctx context.Context) (*ServerInfo, error) {
	info := r.client.Info(ctx, "server", "memory").Val()
	
	serverInfo := &ServerInfo{}
	
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			key, value := parts[0], parts[1]
			
			switch key {
			case "uptime_in_seconds":
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					serverInfo.UptimeInSeconds = val
				}
			case "used_memory":
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					serverInfo.UsedMemory = val
				}
			case "used_memory_rss":
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					serverInfo.UsedMemoryRSS = val
				}
			case "total_system_memory":
				if val, err := strconv.ParseInt(value, 10, 64); err == nil {
					serverInfo.TotalSystemMemory = val
				}
			case "redis_version":
				serverInfo.Version = value
			}
		}
	}
	
	return serverInfo, nil
}

// GetThroughputInfo retrieves command statistics for throughput calculation
func (r *RedisClient) GetThroughputInfo(ctx context.Context) (*ThroughputInfo, error) {
	info := r.client.Info(ctx, "commandstats").Val()
	
	throughputInfo := &ThroughputInfo{
		Timestamp: time.Now(),
	}
	
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, "cmdstat_xadd:") {
			// Parse: cmdstat_xadd:calls=123,usec=456,usec_per_call=3.70
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				stats := strings.Split(parts[1], ",")
				for _, stat := range stats {
					if strings.HasPrefix(stat, "calls=") {
						if val, err := strconv.ParseInt(stat[6:], 10, 64); err == nil {
							throughputInfo.XAddCalls = val
						}
					}
				}
			}
		} else if strings.Contains(line, "cmdstat_xreadgroup:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				stats := strings.Split(parts[1], ",")
				for _, stat := range stats {
					if strings.HasPrefix(stat, "calls=") {
						if val, err := strconv.ParseInt(stat[6:], 10, 64); err == nil {
							throughputInfo.XReadGroupCalls = val
						}
					}
				}
			}
		}
	}
	
	return throughputInfo, nil
}

// FormatBytes formats bytes into human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats duration into human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%02d:%02d", int(d.Minutes()), int(d.Seconds())%60)
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dd %02d:%02d", days, hours, minutes)
}
