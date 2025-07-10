package client

import (
	"context"
	"fmt"
	"io"
	"time"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

// MeshInterface defines the core interface for Meshtastic device communication
type MeshInterface interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect() error
	IsConnected() bool

	// Device information
	GetMyInfo() *pb.MyNodeInfo
	GetNodes() map[uint32]*pb.NodeInfo
	GetChannels() map[uint32]*pb.Channel
	GetConfig() *pb.LocalConfig
	GetModuleConfig() *pb.LocalModuleConfig

	// Message handling
	SendMessage(packet *pb.MeshPacket) error
	SendText(text string, destination uint32) error

	// Event handlers
	SetOnMessage(handler func(*pb.MeshPacket))
	SetOnNodeInfo(handler func(*pb.NodeInfo))
	SetOnPosition(handler func(*pb.Position))
	SetOnTelemetry(handler func(*pb.Telemetry))
	SetOnLogLine(handler func(string))
	SetOnDisconnect(handler func(error))

	// Device path
	DevicePath() string

	// Cleanup
	Close() error
}

// StreamInterface defines the interface for stream-based communication
type StreamInterface interface {
	MeshInterface

	// Stream-specific methods
	GetStream() io.ReadWriteCloser
	SetStream(stream io.ReadWriteCloser)

	// Flow control
	WaitForConfig(timeout time.Duration) error
	SendWantConfig() error

	// Advanced features
	SendAdminMessage(msg *pb.AdminMessage) (*pb.AdminMessage, error)
	GetQueueStatus() (int, int) // queued, capacity
}

// SerialInterface defines the interface for serial-specific communication
type SerialInterface interface {
	StreamInterface

	// Serial-specific methods
	GetSerialConfig() *SerialConfig
	SetSerialConfig(config *SerialConfig) error

	// Connection management
	Reconnect() error
	GetReconnectAttempts() int
	ResetReconnectAttempts()

	// Advanced features
	Flush() error
	GetStatistics() ConnectionStatistics
}

// ConnectionStatistics holds connection statistics
type ConnectionStatistics struct {
	BytesRead        uint64
	BytesWritten     uint64
	ReadErrors       uint64
	WriteErrors      uint64
	Reconnects       uint64
	ConnectDuration  time.Duration
	LastReconnect    time.Time
	FramesReceived   uint64
	FramesSent       uint64
	MessagesReceived uint64
	MessagesSent     uint64
}

// SerialConfig represents serial port configuration
type SerialConfig struct {
	DevicePath   string
	BaudRate     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	StopBits     int
	DataBits     int
	Parity       string
	FlowControl  bool
	DisableHUPCL bool
}

// ConnectionError represents connection-related errors
type ConnectionError struct {
	Op          string
	Err         error
	Recoverable bool
	RetryAfter  time.Duration
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error during %s: %v", e.Op, e.Err)
}

func (e *ConnectionError) Temporary() bool {
	return e.Recoverable
}

func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// DeviceState represents the current state of device communication
type DeviceState int

const (
	StateDisconnected DeviceState = iota
	StateConnecting
	StateConfiguring
	StateConnected
	StateReconnecting
	StateError
)

func (s DeviceState) String() string {
	switch s {
	case StateDisconnected:
		return "Disconnected"
	case StateConnecting:
		return "Connecting"
	case StateConfiguring:
		return "Configuring"
	case StateConnected:
		return "Connected"
	case StateReconnecting:
		return "Reconnecting"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

// StateHandler handles device state transitions
type StateHandler interface {
	OnStateChange(oldState, newState DeviceState)
}

// MessageQueue defines interface for message queuing
type MessageQueue interface {
	// Queue management
	Enqueue(packet *pb.MeshPacket) error
	Dequeue() (*pb.MeshPacket, error)
	Peek() (*pb.MeshPacket, error)
	Size() int
	IsEmpty() bool
	Clear()

	// Flow control
	HasSpace() bool
	WaitForSpace(ctx context.Context) error

	// Priority handling
	EnqueuePriority(packet *pb.MeshPacket) error

	// Close
	Close() error
}

// Timeout represents a timeout handler
type Timeout struct {
	Duration time.Duration
	started  time.Time
}

func NewTimeout(duration time.Duration) *Timeout {
	return &Timeout{
		Duration: duration,
		started:  time.Now(),
	}
}

func (t *Timeout) Reset() {
	t.started = time.Now()
}

func (t *Timeout) Expired() bool {
	return time.Since(t.started) >= t.Duration
}

func (t *Timeout) Remaining() time.Duration {
	elapsed := time.Since(t.started)
	if elapsed >= t.Duration {
		return 0
	}
	return t.Duration - elapsed
}

func (t *Timeout) WaitForCondition(ctx context.Context, condition func() bool) bool {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			if condition() {
				return true
			}
			if t.Expired() {
				return false
			}
		}
	}
}
