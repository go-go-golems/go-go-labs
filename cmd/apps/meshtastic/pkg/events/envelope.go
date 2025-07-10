package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

// Envelope represents the standard event envelope
type Envelope struct {
	EventID       string            `json:"event_id"`
	Timestamp     time.Time         `json:"timestamp"`
	DeviceID      string            `json:"device_id"`
	Source        string            `json:"source"`
	Type          string            `json:"type"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Data          json.RawMessage   `json:"data"`
}

// NewEnvelope creates a new event envelope
func NewEnvelope(eventType, deviceID, source string, data interface{}) (*Envelope, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal event data")
	}

	return &Envelope{
		EventID:   uuid.New().String(),
		Timestamp: time.Now().UTC(),
		DeviceID:  deviceID,
		Source:    source,
		Type:      eventType,
		Metadata:  make(map[string]string),
		Data:      json.RawMessage(dataBytes),
	}, nil
}

// WithCorrelationID sets the correlation ID
func (e *Envelope) WithCorrelationID(correlationID string) *Envelope {
	e.CorrelationID = correlationID
	return e
}

// WithMetadata adds metadata to the envelope
func (e *Envelope) WithMetadata(key, value string) *Envelope {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
	return e
}

// GetData unmarshals the data into the provided interface
func (e *Envelope) GetData(v interface{}) error {
	return json.Unmarshal(e.Data, v)
}

// ToJSON converts the envelope to JSON bytes
func (e *Envelope) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON creates an envelope from JSON bytes
func FromJSON(data []byte) (*Envelope, error) {
	var envelope Envelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal envelope")
	}
	return &envelope, nil
}

// Device lifecycle events
type DeviceConnectedEvent struct {
	DevicePath string            `json:"device_path"`
	DeviceInfo *pb.MyNodeInfo    `json:"device_info,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type DeviceDisconnectedEvent struct {
	DevicePath string `json:"device_path"`
	Reason     string `json:"reason"`
	Error      string `json:"error,omitempty"`
}

type DeviceReconnectingEvent struct {
	DevicePath string `json:"device_path"`
	Attempt    int    `json:"attempt"`
	MaxAttempts int   `json:"max_attempts"`
}

type DeviceErrorEvent struct {
	DevicePath string `json:"device_path"`
	Error      string `json:"error"`
	Severity   string `json:"severity"` // "warning", "error", "critical"
}

// Mesh packet events
type MeshPacketRxEvent struct {
	Packet    *pb.MeshPacket `json:"packet"`
	RawBytes  []byte         `json:"raw_bytes,omitempty"`
	RSSI      int32          `json:"rssi,omitempty"`
	SNR       float32        `json:"snr,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

type MeshPacketTxEvent struct {
	Packet    *pb.MeshPacket `json:"packet"`
	RawBytes  []byte         `json:"raw_bytes,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

type MeshPacketAckEvent struct {
	OriginalPacket *pb.MeshPacket `json:"original_packet"`
	AckPacket      *pb.MeshPacket `json:"ack_packet"`
	Timestamp      time.Time      `json:"timestamp"`
}

type MeshPacketTimeoutEvent struct {
	Packet    *pb.MeshPacket `json:"packet"`
	Timeout   time.Duration  `json:"timeout"`
	Timestamp time.Time      `json:"timestamp"`
}

// Node events
type NodeInfoUpdatedEvent struct {
	NodeInfo  *pb.NodeInfo `json:"node_info"`
	Previous  *pb.NodeInfo `json:"previous,omitempty"`
	IsNew     bool         `json:"is_new"`
	Timestamp time.Time    `json:"timestamp"`
}

type NodePresenceEvent struct {
	NodeID    uint32    `json:"node_id"`
	Online    bool      `json:"online"`
	LastSeen  time.Time `json:"last_seen"`
	Timestamp time.Time `json:"timestamp"`
}

type NodeBatteryEvent struct {
	NodeID         uint32    `json:"node_id"`
	BatteryLevel   uint32    `json:"battery_level"`
	Voltage        float32   `json:"voltage,omitempty"`
	IsCharging     bool      `json:"is_charging"`
	Timestamp      time.Time `json:"timestamp"`
}

// Telemetry events
type TelemetryReceivedEvent struct {
	NodeID    uint32        `json:"node_id"`
	Telemetry *pb.Telemetry `json:"telemetry"`
	Timestamp time.Time     `json:"timestamp"`
}

type PositionUpdatedEvent struct {
	NodeID    uint32      `json:"node_id"`
	Position  *pb.Position `json:"position"`
	Previous  *pb.Position `json:"previous,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
}

type EnvironmentUpdatedEvent struct {
	NodeID          uint32                  `json:"node_id"`
	Environment     *pb.EnvironmentMetrics  `json:"environment"`
	Previous        *pb.EnvironmentMetrics  `json:"previous,omitempty"`
	Timestamp       time.Time               `json:"timestamp"`
}

// Command events
type SendTextCommandEvent struct {
	Text         string `json:"text"`
	Destination  uint32 `json:"destination"`
	Channel      uint32 `json:"channel,omitempty"`
	WantResponse bool   `json:"want_response,omitempty"`
}

type RequestInfoCommandEvent struct {
	NodeID   uint32 `json:"node_id"`
	InfoType string `json:"info_type"` // "device", "node", "channels", "config"
}

type RequestTelemetryCommandEvent struct {
	NodeID        uint32 `json:"node_id"`
	TelemetryType string `json:"telemetry_type"` // "device", "environment", "power"
}

type RequestPositionCommandEvent struct {
	NodeID uint32 `json:"node_id"`
}

// Response events
type ResponseSuccessEvent struct {
	CommandID     string      `json:"command_id"`
	CorrelationID string      `json:"correlation_id"`
	Response      interface{} `json:"response"`
	Duration      time.Duration `json:"duration"`
	Timestamp     time.Time   `json:"timestamp"`
}

type ResponseErrorEvent struct {
	CommandID     string        `json:"command_id"`
	CorrelationID string        `json:"correlation_id"`
	Error         string        `json:"error"`
	Duration      time.Duration `json:"duration"`
	Timestamp     time.Time     `json:"timestamp"`
}

type ResponseTimeoutEvent struct {
	CommandID     string        `json:"command_id"`
	CorrelationID string        `json:"correlation_id"`
	Timeout       time.Duration `json:"timeout"`
	Timestamp     time.Time     `json:"timestamp"`
}

// System events
type SystemStartupEvent struct {
	Version   string    `json:"version"`
	Config    string    `json:"config"`
	Timestamp time.Time `json:"timestamp"`
}

type SystemShutdownEvent struct {
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

type SystemErrorEvent struct {
	Error     string    `json:"error"`
	Component string    `json:"component"`
	Severity  string    `json:"severity"`
	Timestamp time.Time `json:"timestamp"`
}

// Helper functions for creating common events
func NewDeviceConnectedEvent(devicePath string, deviceInfo *pb.MyNodeInfo) *DeviceConnectedEvent {
	return &DeviceConnectedEvent{
		DevicePath: devicePath,
		DeviceInfo: deviceInfo,
		Metadata:   make(map[string]string),
	}
}

func NewDeviceDisconnectedEvent(devicePath, reason string, err error) *DeviceDisconnectedEvent {
	event := &DeviceDisconnectedEvent{
		DevicePath: devicePath,
		Reason:     reason,
	}
	if err != nil {
		event.Error = err.Error()
	}
	return event
}

func NewMeshPacketRxEvent(packet *pb.MeshPacket) *MeshPacketRxEvent {
	return &MeshPacketRxEvent{
		Packet:    packet,
		Timestamp: time.Now().UTC(),
	}
}

func NewMeshPacketTxEvent(packet *pb.MeshPacket) *MeshPacketTxEvent {
	return &MeshPacketTxEvent{
		Packet:    packet,
		Timestamp: time.Now().UTC(),
	}
}

func NewNodeInfoUpdatedEvent(nodeInfo, previous *pb.NodeInfo, isNew bool) *NodeInfoUpdatedEvent {
	return &NodeInfoUpdatedEvent{
		NodeInfo:  nodeInfo,
		Previous:  previous,
		IsNew:     isNew,
		Timestamp: time.Now().UTC(),
	}
}

func NewTelemetryReceivedEvent(nodeID uint32, telemetry *pb.Telemetry) *TelemetryReceivedEvent {
	return &TelemetryReceivedEvent{
		NodeID:    nodeID,
		Telemetry: telemetry,
		Timestamp: time.Now().UTC(),
	}
}

func NewPositionUpdatedEvent(nodeID uint32, position, previous *pb.Position) *PositionUpdatedEvent {
	return &PositionUpdatedEvent{
		NodeID:    nodeID,
		Position:  position,
		Previous:  previous,
		Timestamp: time.Now().UTC(),
	}
}
