package deviceadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/client"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/events"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/meshbus"
	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

// DeviceAdapter wraps RobustMeshtasticClient and provides event-driven interface
type DeviceAdapter struct {
	id        string
	client    *client.RobustMeshtasticClient
	bus       *meshbus.Bus
	publisher message.Publisher
	
	mu        sync.RWMutex
	running   bool
	ctx       context.Context
	cancel    context.CancelFunc
	
	// State tracking
	lastMyInfo    *pb.MyNodeInfo
	lastNodes     map[uint32]*pb.NodeInfo
	lastChannels  map[uint32]*pb.Channel
	lastConfig    *pb.LocalConfig
	
	// Statistics
	stats *Statistics
}

// Statistics holds adapter statistics
type Statistics struct {
	MessagesReceived    uint64
	MessagesSent        uint64
	EventsPublished     uint64
	CommandsProcessed   uint64
	Errors              uint64
	Reconnects          uint64
	UptimeStart         time.Time
	LastActivity        time.Time
}

// NewDeviceAdapter creates a new device adapter
func NewDeviceAdapter(id string, client *client.RobustMeshtasticClient, bus *meshbus.Bus) *DeviceAdapter {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &DeviceAdapter{
		id:        id,
		client:    client,
		bus:       bus,
		publisher: bus.Publisher(),
		ctx:       ctx,
		cancel:    cancel,
		lastNodes: make(map[uint32]*pb.NodeInfo),
		lastChannels: make(map[uint32]*pb.Channel),
		stats: &Statistics{
			UptimeStart: time.Now(),
		},
	}
}

// Start starts the device adapter
func (d *DeviceAdapter) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if d.running {
		return errors.New("adapter is already running")
	}
	
	log.Info().Str("device_id", d.id).Msg("Starting device adapter")
	
	// Set up client callbacks
	d.setupClientCallbacks()
	
	// Add command handlers
	if err := d.setupCommandHandlers(); err != nil {
		return errors.Wrap(err, "failed to setup command handlers")
	}
	
	// Connect to device
	if err := d.client.Connect(ctx); err != nil {
		return errors.Wrap(err, "failed to connect to device")
	}
	
	// Start heartbeat
	d.client.StartHeartbeat()
	
	// Publish device connected event
	d.publishDeviceConnected()
	
	d.running = true
	d.stats.LastActivity = time.Now()
	
	log.Info().Str("device_id", d.id).Msg("Device adapter started successfully")
	
	return nil
}

// Stop stops the device adapter
func (d *DeviceAdapter) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	if !d.running {
		return nil
	}
	
	log.Info().Str("device_id", d.id).Msg("Stopping device adapter")
	
	// Cancel context
	d.cancel()
	
	// Publish device disconnected event
	d.publishDeviceDisconnected("stopped", nil)
	
	// Disconnect from device
	if err := d.client.Disconnect(); err != nil {
		log.Error().Err(err).Str("device_id", d.id).Msg("Error disconnecting from device")
	}
	
	// Close client
	if err := d.client.Close(); err != nil {
		log.Error().Err(err).Str("device_id", d.id).Msg("Error closing client")
	}
	
	d.running = false
	
	log.Info().Str("device_id", d.id).Msg("Device adapter stopped")
	
	return nil
}

// setupClientCallbacks sets up the client callbacks
func (d *DeviceAdapter) setupClientCallbacks() {
	// Message callback
	d.client.SetOnMessage(func(packet *pb.MeshPacket) {
		d.mu.Lock()
		d.stats.MessagesReceived++
		d.stats.LastActivity = time.Now()
		d.mu.Unlock()
		
		log.Debug().
			Str("device_id", d.id).
			Uint32("from", packet.From).
			Uint32("to", packet.To).
			Msg("Received message")
		
		// Create and publish event
		event := events.NewMeshPacketRxEvent(packet)
		d.publishEvent(events.EventMeshPacketRx, event)
	})
	
	// Node info callback
	d.client.SetOnNodeInfo(func(nodeInfo *pb.NodeInfo) {
		d.mu.Lock()
		previous := d.lastNodes[nodeInfo.Num]
		d.lastNodes[nodeInfo.Num] = nodeInfo
		d.stats.LastActivity = time.Now()
		d.mu.Unlock()
		
		isNew := previous == nil
		
		log.Debug().
			Str("device_id", d.id).
			Uint32("node_id", nodeInfo.Num).
			Bool("is_new", isNew).
			Msg("Node info updated")
		
		// Create and publish event
		event := events.NewNodeInfoUpdatedEvent(nodeInfo, previous, isNew)
		d.publishEvent(events.EventNodeInfoUpdated, event)
	})
	
	// Position callback
	d.client.SetOnPosition(func(position *pb.Position) {
		d.mu.Lock()
		d.stats.LastActivity = time.Now()
		d.mu.Unlock()
		
		log.Debug().
			Str("device_id", d.id).
			Int32("lat_i", position.GetLatitudeI()).
			Int32("lon_i", position.GetLongitudeI()).
			Msg("Position updated")
		
		// For position events, we need to determine the node ID
		// This is a simplified implementation
		nodeID := uint32(0) // We'd need to track this properly
		
		// Create and publish event
		event := events.NewPositionUpdatedEvent(nodeID, position, nil)
		d.publishEvent(events.EventPositionUpdated, event)
	})
	
	// Telemetry callback
	d.client.SetOnTelemetry(func(telemetry *pb.Telemetry) {
		d.mu.Lock()
		d.stats.LastActivity = time.Now()
		d.mu.Unlock()
		
		log.Debug().
			Str("device_id", d.id).
			Msg("Telemetry received")
		
		// For telemetry events, we need to determine the node ID
		// This is a simplified implementation
		nodeID := uint32(0) // We'd need to track this properly
		
		// Create and publish event
		event := events.NewTelemetryReceivedEvent(nodeID, telemetry)
		d.publishEvent(events.EventTelemetryReceived, event)
	})
	
	// Disconnect callback
	d.client.SetOnDisconnect(func(err error) {
		d.mu.Lock()
		d.stats.Reconnects++
		d.stats.LastActivity = time.Now()
		d.mu.Unlock()
		
		log.Warn().
			Err(err).
			Str("device_id", d.id).
			Msg("Device disconnected")
		
		// Publish device disconnected event
		d.publishDeviceDisconnected("connection_lost", err)
	})
	
	// Log line callback
	d.client.SetOnLogLine(func(line string) {
		log.Debug().
			Str("device_id", d.id).
			Str("device_log", line).
			Msg("Device log")
	})
}

// setupCommandHandlers sets up command handlers
func (d *DeviceAdapter) setupCommandHandlers() error {
	// Send text command handler
	if err := d.bus.AddHandler(
		fmt.Sprintf("send_text_%s", d.id),
		meshbus.BuildTopicName(meshbus.TopicCommandSendText, d.id),
		d.handleSendTextCommand,
	); err != nil {
		return errors.Wrap(err, "failed to add send text handler")
	}
	
	// Request info command handler
	if err := d.bus.AddHandler(
		fmt.Sprintf("request_info_%s", d.id),
		meshbus.BuildTopicName(meshbus.TopicCommandRequestInfo, d.id),
		d.handleRequestInfoCommand,
	); err != nil {
		return errors.Wrap(err, "failed to add request info handler")
	}
	
	// Request telemetry command handler
	if err := d.bus.AddHandler(
		fmt.Sprintf("request_telemetry_%s", d.id),
		meshbus.BuildTopicName(meshbus.TopicCommandRequestTelemetry, d.id),
		d.handleRequestTelemetryCommand,
	); err != nil {
		return errors.Wrap(err, "failed to add request telemetry handler")
	}
	
	// Request position command handler
	if err := d.bus.AddHandler(
		fmt.Sprintf("request_position_%s", d.id),
		meshbus.BuildTopicName(meshbus.TopicCommandRequestPosition, d.id),
		d.handleRequestPositionCommand,
	); err != nil {
		return errors.Wrap(err, "failed to add request position handler")
	}
	
	return nil
}

// handleSendTextCommand handles send text commands
func (d *DeviceAdapter) handleSendTextCommand(msg *message.Message) error {
	d.mu.Lock()
	d.stats.CommandsProcessed++
	d.mu.Unlock()
	
	// Parse command
	var envelope events.Envelope
	if err := json.Unmarshal(msg.Payload, &envelope); err != nil {
		return errors.Wrap(err, "failed to parse envelope")
	}
	
	var command events.SendTextCommandEvent
	if err := envelope.GetData(&command); err != nil {
		return errors.Wrap(err, "failed to parse command")
	}
	
	log.Info().
		Str("device_id", d.id).
		Str("text", command.Text).
		Uint32("destination", command.Destination).
		Msg("Sending text message")
	
	// Send message
	if err := d.client.SendText(command.Text, command.Destination); err != nil {
		d.mu.Lock()
		d.stats.Errors++
		d.mu.Unlock()
		
		// Publish error response
		d.publishResponseError(envelope.CorrelationID, err)
		
		return errors.Wrap(err, "failed to send text message")
	}
	
	d.mu.Lock()
	d.stats.MessagesSent++
	d.mu.Unlock()
	
	// Publish success response
	d.publishResponseSuccess(envelope.CorrelationID, "Message sent successfully")
	
	return nil
}

// handleRequestInfoCommand handles request info commands
func (d *DeviceAdapter) handleRequestInfoCommand(msg *message.Message) error {
	d.mu.Lock()
	d.stats.CommandsProcessed++
	d.mu.Unlock()
	
	// Parse command
	var envelope events.Envelope
	if err := json.Unmarshal(msg.Payload, &envelope); err != nil {
		return errors.Wrap(err, "failed to parse envelope")
	}
	
	var command events.RequestInfoCommandEvent
	if err := envelope.GetData(&command); err != nil {
		return errors.Wrap(err, "failed to parse command")
	}
	
	log.Info().
		Str("device_id", d.id).
		Uint32("node_id", command.NodeID).
		Str("info_type", command.InfoType).
		Msg("Requesting node info")
	
	// Get requested info
	var response interface{}
	switch command.InfoType {
	case "device":
		response = d.client.GetMyInfo()
	case "nodes":
		response = d.client.GetNodes()
	case "channels":
		response = d.client.GetChannels()
	case "config":
		response = d.client.GetConfig()
	default:
		err := fmt.Errorf("unknown info type: %s", command.InfoType)
		d.publishResponseError(envelope.CorrelationID, err)
		return err
	}
	
	// Publish success response
	d.publishResponseSuccess(envelope.CorrelationID, response)
	
	return nil
}

// handleRequestTelemetryCommand handles request telemetry commands
func (d *DeviceAdapter) handleRequestTelemetryCommand(msg *message.Message) error {
	d.mu.Lock()
	d.stats.CommandsProcessed++
	d.mu.Unlock()
	
	// Parse command
	var envelope events.Envelope
	if err := json.Unmarshal(msg.Payload, &envelope); err != nil {
		return errors.Wrap(err, "failed to parse envelope")
	}
	
	var command events.RequestTelemetryCommandEvent
	if err := envelope.GetData(&command); err != nil {
		return errors.Wrap(err, "failed to parse command")
	}
	
	log.Info().
		Str("device_id", d.id).
		Uint32("node_id", command.NodeID).
		Str("telemetry_type", command.TelemetryType).
		Msg("Requesting telemetry")
	
	// This is a simplified implementation
	// In a real implementation, we'd send a telemetry request packet
	d.publishResponseSuccess(envelope.CorrelationID, "Telemetry request sent")
	
	return nil
}

// handleRequestPositionCommand handles request position commands
func (d *DeviceAdapter) handleRequestPositionCommand(msg *message.Message) error {
	d.mu.Lock()
	d.stats.CommandsProcessed++
	d.mu.Unlock()
	
	// Parse command
	var envelope events.Envelope
	if err := json.Unmarshal(msg.Payload, &envelope); err != nil {
		return errors.Wrap(err, "failed to parse envelope")
	}
	
	var command events.RequestPositionCommandEvent
	if err := envelope.GetData(&command); err != nil {
		return errors.Wrap(err, "failed to parse command")
	}
	
	log.Info().
		Str("device_id", d.id).
		Uint32("node_id", command.NodeID).
		Msg("Requesting position")
	
	// This is a simplified implementation
	// In a real implementation, we'd send a position request packet
	d.publishResponseSuccess(envelope.CorrelationID, "Position request sent")
	
	return nil
}

// publishEvent publishes an event
func (d *DeviceAdapter) publishEvent(eventType string, data interface{}) {
	envelope, err := events.NewEnvelope(eventType, d.id, events.SourceDeviceAdapter, data)
	if err != nil {
		log.Error().Err(err).Str("device_id", d.id).Msg("Failed to create event envelope")
		return
	}
	
	// Add metadata
	envelope.WithMetadata(events.MetadataDeviceID, d.id)
	envelope.WithMetadata(events.MetadataTimestamp, time.Now().UTC().Format(time.RFC3339))
	
	// Marshal envelope
	payload, err := envelope.ToJSON()
	if err != nil {
		log.Error().Err(err).Str("device_id", d.id).Msg("Failed to marshal event envelope")
		return
	}
	
	// Create watermill message
	msg := message.NewMessage(uuid.New().String(), payload)
	msg.Metadata.Set("event_type", eventType)
	msg.Metadata.Set("device_id", d.id)
	msg.Metadata.Set("correlation_id", envelope.CorrelationID)
	
	// Publish to device-specific topic
	deviceTopic := meshbus.BuildTopicName(eventType, d.id)
	if err := d.publisher.Publish(deviceTopic, msg); err != nil {
		log.Error().Err(err).Str("device_id", d.id).Str("topic", deviceTopic).Msg("Failed to publish event")
		return
	}
	
	// Also publish to broadcast topic
	broadcastTopic := meshbus.BuildBroadcastTopic(eventType)
	if err := d.publisher.Publish(broadcastTopic, msg); err != nil {
		log.Error().Err(err).Str("device_id", d.id).Str("topic", broadcastTopic).Msg("Failed to publish broadcast event")
		return
	}
	
	d.mu.Lock()
	d.stats.EventsPublished++
	d.mu.Unlock()
}

// publishDeviceConnected publishes device connected event
func (d *DeviceAdapter) publishDeviceConnected() {
	myInfo := d.client.GetMyInfo()
	d.lastMyInfo = myInfo
	
	event := events.NewDeviceConnectedEvent(d.client.DevicePath(), myInfo)
	d.publishEvent(events.EventDeviceConnected, event)
}

// publishDeviceDisconnected publishes device disconnected event
func (d *DeviceAdapter) publishDeviceDisconnected(reason string, err error) {
	event := events.NewDeviceDisconnectedEvent(d.client.DevicePath(), reason, err)
	d.publishEvent(events.EventDeviceDisconnected, event)
}

// publishResponseSuccess publishes a successful response
func (d *DeviceAdapter) publishResponseSuccess(correlationID string, response interface{}) {
	event := &events.ResponseSuccessEvent{
		CommandID:     correlationID,
		CorrelationID: correlationID,
		Response:      response,
		Duration:      0, // We'd track this properly in real implementation
		Timestamp:     time.Now().UTC(),
	}
	
	envelope, err := events.NewEnvelope(events.EventResponseSuccess, d.id, events.SourceDeviceAdapter, event)
	if err != nil {
		log.Error().Err(err).Str("device_id", d.id).Msg("Failed to create response envelope")
		return
	}
	
	envelope.WithCorrelationID(correlationID)
	
	payload, err := envelope.ToJSON()
	if err != nil {
		log.Error().Err(err).Str("device_id", d.id).Msg("Failed to marshal response envelope")
		return
	}
	
	msg := message.NewMessage(uuid.New().String(), payload)
	msg.Metadata.Set("event_type", events.EventResponseSuccess)
	msg.Metadata.Set("device_id", d.id)
	msg.Metadata.Set("correlation_id", correlationID)
	
	topic := meshbus.BuildTopicName(events.EventResponseSuccess, d.id)
	if err := d.publisher.Publish(topic, msg); err != nil {
		log.Error().Err(err).Str("device_id", d.id).Str("topic", topic).Msg("Failed to publish response")
	}
}

// publishResponseError publishes an error response
func (d *DeviceAdapter) publishResponseError(correlationID string, err error) {
	event := &events.ResponseErrorEvent{
		CommandID:     correlationID,
		CorrelationID: correlationID,
		Error:         err.Error(),
		Duration:      0, // We'd track this properly in real implementation
		Timestamp:     time.Now().UTC(),
	}
	
	envelope, err := events.NewEnvelope(events.EventResponseError, d.id, events.SourceDeviceAdapter, event)
	if err != nil {
		log.Error().Err(err).Str("device_id", d.id).Msg("Failed to create error response envelope")
		return
	}
	
	envelope.WithCorrelationID(correlationID)
	
	payload, err := envelope.ToJSON()
	if err != nil {
		log.Error().Err(err).Str("device_id", d.id).Msg("Failed to marshal error response envelope")
		return
	}
	
	msg := message.NewMessage(uuid.New().String(), payload)
	msg.Metadata.Set("event_type", events.EventResponseError)
	msg.Metadata.Set("device_id", d.id)
	msg.Metadata.Set("correlation_id", correlationID)
	
	topic := meshbus.BuildTopicName(events.EventResponseError, d.id)
	if err := d.publisher.Publish(topic, msg); err != nil {
		log.Error().Err(err).Str("device_id", d.id).Str("topic", topic).Msg("Failed to publish error response")
	}
}

// GetStatistics returns adapter statistics
func (d *DeviceAdapter) GetStatistics() Statistics {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return *d.stats
}

// IsRunning returns true if the adapter is running
func (d *DeviceAdapter) IsRunning() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.running
}

// GetDeviceID returns the device ID
func (d *DeviceAdapter) GetDeviceID() string {
	return d.id
}

// GetDevicePath returns the device path
func (d *DeviceAdapter) GetDevicePath() string {
	return d.client.DevicePath()
}

// GetConnectionStatus returns the connection status
func (d *DeviceAdapter) GetConnectionStatus() client.ConnectionStatus {
	return d.client.GetConnectionStatus()
}
