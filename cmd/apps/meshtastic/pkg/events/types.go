package events

// Event type constants
const (
	// Device lifecycle events
	EventDeviceConnected    = "device.connected"
	EventDeviceDisconnected = "device.disconnected"
	EventDeviceReconnecting = "device.reconnecting"
	EventDeviceError        = "device.error"

	// Mesh packet events
	EventMeshPacketRx      = "mesh.packet.rx"
	EventMeshPacketTx      = "mesh.packet.tx"
	EventMeshPacketAck     = "mesh.packet.ack"
	EventMeshPacketTimeout = "mesh.packet.timeout"

	// Node events
	EventNodeInfoUpdated = "mesh.nodeinfo.updated"
	EventNodePresence    = "mesh.node.presence"
	EventNodeBattery     = "mesh.node.battery"

	// Telemetry events
	EventTelemetryReceived  = "mesh.telemetry.received"
	EventPositionUpdated    = "mesh.position.updated"
	EventEnvironmentUpdated = "mesh.environment.updated"

	// Command events
	EventCommandSendText         = "command.send_text"
	EventCommandRequestInfo      = "command.request_info"
	EventCommandRequestTelemetry = "command.request_telemetry"
	EventCommandRequestPosition  = "command.request_position"

	// Response events
	EventResponseSuccess = "response.success"
	EventResponseError   = "response.error"
	EventResponseTimeout = "response.timeout"

	// System events
	EventSystemStartup  = "system.startup"
	EventSystemShutdown = "system.shutdown"
	EventSystemError    = "system.error"
)

// Event source constants
const (
	SourceDeviceAdapter = "device_adapter"
	SourceMeshBus       = "mesh_bus"
	SourceREPL          = "repl"
	SourceTUI           = "tui"
	SourceMQTT          = "mqtt"
	SourceSystem        = "system"
)

// Event priority constants
const (
	PriorityLow      = "low"
	PriorityNormal   = "normal"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

// Event severity constants
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"
)

// Command status constants
const (
	StatusPending   = "pending"
	StatusExecuting = "executing"
	StatusSuccess   = "success"
	StatusError     = "error"
	StatusTimeout   = "timeout"
)

// Node status constants
const (
	NodeStatusOnline  = "online"
	NodeStatusOffline = "offline"
	NodeStatusUnknown = "unknown"
)

// Device state constants
const (
	DeviceStateDisconnected = "disconnected"
	DeviceStateConnecting   = "connecting"
	DeviceStateConnected    = "connected"
	DeviceStateError        = "error"
)

// Message type constants
const (
	MessageTypeText      = "text"
	MessageTypePosition  = "position"
	MessageTypeTelemetry = "telemetry"
	MessageTypeNodeInfo  = "nodeinfo"
	MessageTypeAdmin     = "admin"
	MessageTypeUnknown   = "unknown"
)

// Telemetry type constants
const (
	TelemetryTypeDevice      = "device"
	TelemetryTypeEnvironment = "environment"
	TelemetryTypePower       = "power"
	TelemetryTypeUnknown     = "unknown"
)

// Channel type constants
const (
	ChannelPrimary   = "primary"
	ChannelSecondary = "secondary"
	ChannelUnknown   = "unknown"
)

// Direction constants
const (
	DirectionInbound  = "inbound"
	DirectionOutbound = "outbound"
)

// Predefined metadata keys
const (
	MetadataCorrelationID = "correlation_id"
	MetadataDeviceID      = "device_id"
	MetadataNodeID        = "node_id"
	MetadataChannelID     = "channel_id"
	MetadataTimestamp     = "timestamp"
	MetadataSource        = "source"
	MetadataDirection     = "direction"
	MetadataPriority      = "priority"
	MetadataSeverity      = "severity"
	MetadataStatus        = "status"
	MetadataRetryCount    = "retry_count"
	MetadataRetryAttempt  = "retry_attempt"
	MetadataProcessedAt   = "processed_at"
	MetadataRSSI          = "rssi"
	MetadataSNR           = "snr"
	MetadataHopLimit      = "hop_limit"
	MetadataWantAck       = "want_ack"
	MetadataViaRepeater   = "via_repeater"
	MetadataEncrypted     = "encrypted"
	MetadataMessageType   = "message_type"
	MetadataPayloadSize   = "payload_size"
	MetadataCommand       = "command"
	MetadataResponse      = "response"
	MetadataError         = "error"
	MetadataTimeout       = "timeout"
	MetadataDuration      = "duration"
	MetadataAttempt       = "attempt"
	MetadataMaxAttempts   = "max_attempts"
	MetadataReason        = "reason"
	MetadataComponent     = "component"
	MetadataVersion       = "version"
	MetadataConfig        = "config"
)

// Well-known device IDs
const (
	DeviceIDBroadcast = "broadcast"
	DeviceIDUnknown   = "unknown"
	DeviceIDAll       = "all"
)

// Well-known node IDs
const (
	NodeIDBroadcast = 0xFFFFFFFF
	NodeIDUnknown   = 0x00000000
)

// IsValidEventType checks if an event type is valid
func IsValidEventType(eventType string) bool {
	validTypes := map[string]bool{
		EventDeviceConnected:         true,
		EventDeviceDisconnected:      true,
		EventDeviceReconnecting:      true,
		EventDeviceError:             true,
		EventMeshPacketRx:            true,
		EventMeshPacketTx:            true,
		EventMeshPacketAck:           true,
		EventMeshPacketTimeout:       true,
		EventNodeInfoUpdated:         true,
		EventNodePresence:            true,
		EventNodeBattery:             true,
		EventTelemetryReceived:       true,
		EventPositionUpdated:         true,
		EventEnvironmentUpdated:      true,
		EventCommandSendText:         true,
		EventCommandRequestInfo:      true,
		EventCommandRequestTelemetry: true,
		EventCommandRequestPosition:  true,
		EventResponseSuccess:         true,
		EventResponseError:           true,
		EventResponseTimeout:         true,
		EventSystemStartup:           true,
		EventSystemShutdown:          true,
		EventSystemError:             true,
	}
	return validTypes[eventType]
}

// IsValidSource checks if a source is valid
func IsValidSource(source string) bool {
	validSources := map[string]bool{
		SourceDeviceAdapter: true,
		SourceMeshBus:       true,
		SourceREPL:          true,
		SourceTUI:           true,
		SourceMQTT:          true,
		SourceSystem:        true,
	}
	return validSources[source]
}

// IsValidPriority checks if a priority is valid
func IsValidPriority(priority string) bool {
	validPriorities := map[string]bool{
		PriorityLow:      true,
		PriorityNormal:   true,
		PriorityHigh:     true,
		PriorityCritical: true,
	}
	return validPriorities[priority]
}

// IsValidSeverity checks if a severity is valid
func IsValidSeverity(severity string) bool {
	validSeverities := map[string]bool{
		SeverityInfo:     true,
		SeverityWarning:  true,
		SeverityError:    true,
		SeverityCritical: true,
	}
	return validSeverities[severity]
}

// GetEventCategory returns the category of an event type
func GetEventCategory(eventType string) string {
	switch {
	case eventType == EventDeviceConnected || eventType == EventDeviceDisconnected ||
		eventType == EventDeviceReconnecting || eventType == EventDeviceError:
		return "device"
	case eventType == EventMeshPacketRx || eventType == EventMeshPacketTx ||
		eventType == EventMeshPacketAck || eventType == EventMeshPacketTimeout:
		return "mesh"
	case eventType == EventNodeInfoUpdated || eventType == EventNodePresence ||
		eventType == EventNodeBattery:
		return "node"
	case eventType == EventTelemetryReceived || eventType == EventPositionUpdated ||
		eventType == EventEnvironmentUpdated:
		return "telemetry"
	case eventType == EventCommandSendText || eventType == EventCommandRequestInfo ||
		eventType == EventCommandRequestTelemetry || eventType == EventCommandRequestPosition:
		return "command"
	case eventType == EventResponseSuccess || eventType == EventResponseError ||
		eventType == EventResponseTimeout:
		return "response"
	case eventType == EventSystemStartup || eventType == EventSystemShutdown ||
		eventType == EventSystemError:
		return "system"
	default:
		return "unknown"
	}
}

// GetEventPriority returns the default priority for an event type
func GetEventPriority(eventType string) string {
	switch eventType {
	case EventDeviceError, EventSystemError:
		return PriorityHigh
	case EventDeviceDisconnected, EventResponseError, EventResponseTimeout:
		return PriorityHigh
	case EventDeviceConnected, EventDeviceReconnecting:
		return PriorityNormal
	case EventSystemStartup, EventSystemShutdown:
		return PriorityNormal
	case EventMeshPacketRx, EventMeshPacketTx:
		return PriorityNormal
	case EventNodeInfoUpdated, EventTelemetryReceived, EventPositionUpdated:
		return PriorityLow
	default:
		return PriorityNormal
	}
}
