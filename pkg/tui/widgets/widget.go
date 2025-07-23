// Package widgets provides modular bubbletea components for the Redis monitor TUI
package widgets

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Widget interface that all TUI widgets must implement
type Widget interface {
	tea.Model
	SetSize(width, height int)
	SetFocused(focused bool)
	MinHeight() int
	MaxHeight() int
}

// Common message types used across widgets
type (
	// DataUpdateMsg contains fresh data from Redis
	DataUpdateMsg struct {
		ServerData  ServerData
		StreamsData []StreamData
		Timestamp   time.Time
	}

	// FocusChangeMsg changes widget focus
	FocusChangeMsg struct {
		Widget string
	}

	// RefreshRateChangeMsg changes the refresh rate
	RefreshRateChangeMsg struct {
		NewRate time.Duration
	}
)

// Data structures for widget communication
type ServerData struct {
	Uptime      time.Duration
	MemoryUsed  int64
	MemoryTotal int64
	Version     string
	Throughput  float64
}

type StreamData struct {
	Name           string
	Length         int64
	MemoryUsage    int64
	Groups         int64
	LastID         string
	ConsumerGroups []GroupData
	MessageRates   []float64
}

type GroupData struct {
	Name      string
	Stream    string
	Consumers []ConsumerData
	Pending   int64
}

type ConsumerData struct {
	Name    string
	Pending int64
	Idle    time.Duration
}
