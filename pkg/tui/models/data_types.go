// Package models contains data structures used throughout the Redis monitor TUI
package models

import "time"

// StreamData represents data for a single Redis stream
type StreamData struct {
	Name           string
	Length         int64
	MemoryUsage    int64
	Groups         int64
	LastID         string
	ConsumerGroups []GroupData
	MessageRates   []float64 // for sparkline visualization
}

// GroupData represents data for a Redis consumer group
type GroupData struct {
	Name      string
	Stream    string
	Consumers []ConsumerData
	Pending   int64
}

// ConsumerData represents data for a single Redis consumer
type ConsumerData struct {
	Name    string
	Pending int64
	Idle    time.Duration
}

// ServerData represents Redis server information and metrics
type ServerData struct {
	Uptime      time.Duration
	MemoryUsed  int64
	MemoryTotal int64
	Version     string
	Throughput  float64
}
