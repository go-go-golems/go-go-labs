# Meshtastic Event-Driven Architecture Design

## Executive Summary

This document outlines the design for an event-driven framework built on top of the existing robust Meshtastic client implementation. The framework uses Watermill for event processing and enables multiple concurrent device connections, real-time UI updates, MQTT integration, and extensible architecture for future enhancements.

## Architecture Overview

### Core Concept

The framework introduces a **DeviceAdapter** layer that wraps the existing `RobustMeshtasticClient` and converts all device callbacks into Watermill messages. A central event bus handles all inter-component communication through pub/sub patterns.

```
┌─────────────────┐  callbacks  ┌──────────────────┐
│ RobustClient    │────────────▶│ DeviceAdapter    │
│ (per device)    │◀────────────│ (event wrapper)  │
└─────────────────┘  commands   └──────────────────┘
                                          │
                                          ▼
                                ┌──────────────────┐
                                │ Watermill Router │
                                │ (event bus)      │
                                └──────────────────┘
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    ▼                     ▼                     ▼
            ┌─────────────┐     ┌─────────────────┐     ┌─────────────┐
            │    TUI      │     │   MQTT Bridge   │     │ Event Store │
            │  (real-time │     │  (bidirectional) │     │ (optional)  │
            │   updates)  │     │                 │     │             │
            └─────────────┘     └─────────────────┘     └─────────────┘
```

## Component Design

### 1. DeviceAdapter

**Purpose**: Wraps `RobustMeshtasticClient` and provides event-driven interface

**Responsibilities**:
- Convert device callbacks to Watermill events
- Subscribe to command topics and execute on device
- Handle device lifecycle events
- Maintain event correlation for command/response

**Key Methods**:
```go
type DeviceAdapter struct {
    id       string
    client   *client.RobustMeshtasticClient
    pub      message.Publisher
    sub      message.Subscriber
    router   *message.Router
}

func NewDeviceAdapter(id string, client *client.RobustMeshtasticClient) *DeviceAdapter
func (d *DeviceAdapter) Start(ctx context.Context) error
func (d *DeviceAdapter) Stop() error
func (d *DeviceAdapter) publish(topic string, data interface{})
func (d *DeviceAdapter) handleCommands(ctx context.Context)
```

### 2. DeviceSupervisor

**Purpose**: Manages multiple device connections and their adapters

**Responsibilities**:
- Device discovery and enumeration
- Adapter lifecycle management
- Connection monitoring and recovery
- Load balancing across devices

**Key Methods**:
```go
type DeviceSupervisor struct {
    adapters map[string]*DeviceAdapter
    config   *SupervisorConfig
    pub      message.Publisher
    sub      message.Subscriber
}

func NewDeviceSupervisor(config *SupervisorConfig) *DeviceSupervisor
func (s *DeviceSupervisor) Start(ctx context.Context) error
func (s *DeviceSupervisor) AddDevice(devicePath string) error
func (s *DeviceSupervisor) RemoveDevice(deviceID string) error
func (s *DeviceSupervisor) DiscoverDevices() ([]string, error)
```

### 3. Event Bus (Watermill)

**Purpose**: Central message routing and pub/sub infrastructure

**Components**:
- **Publisher**: Sends events to topics
- **Subscriber**: Receives events from topics
- **Router**: Routes messages between publishers and subscribers
- **Middleware**: Provides cross-cutting concerns

**Configuration**:
```go
type EventBusConfig struct {
    Transport    string // "memory", "nats", "kafka"
    BufferSize   int
    Marshaler    message.Marshaler
    Middlewares  []message.HandlerMiddleware
}
```

### 4. Event Schema

**Envelope Structure**:
```go
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
```

**Event Types**:
- `device.connected` / `device.disconnected` / `device.reconnecting`
- `mesh.packet.rx` / `mesh.packet.tx`
- `mesh.nodeinfo.updated`
- `mesh.telemetry.received`
- `mesh.position.updated`
- `command.send_text` / `command.request_telemetry`
- `ui.tab_changed` / `ui.filter_applied`

## Topic Structure

### Topic Naming Convention
- `mesh.<deviceID>.*` - Device-specific events
- `mesh.broadcast.*` - Aggregate events (fan-out)
- `command.<deviceID>.*` - Device-specific commands
- `command.all.*` - Broadcast commands
- `ui.*` - UI-related events
- `mqtt.*` - MQTT bridge events

### Example Topics
```
mesh.dev_ttyACM0.packet.rx
mesh.dev_ttyACM0.nodeinfo.updated
mesh.broadcast.telemetry.received
command.dev_ttyACM0.send_text
command.all.request_position
ui.tab_changed
mqtt.message.rx
```

## Integration Patterns

### 1. TUI Integration

**Event Subscriptions**:
```go
// Real-time updates
router.AddNoPublisherHandler("tui-messages", 
    "mesh.broadcast.packet.rx", sub, tuiHandler.HandleMessage)

router.AddNoPublisherHandler("tui-nodes", 
    "mesh.broadcast.nodeinfo.updated", sub, tuiHandler.HandleNodeUpdate)

router.AddNoPublisherHandler("tui-telemetry", 
    "mesh.broadcast.telemetry.received", sub, tuiHandler.HandleTelemetry)
```

**Command Publishing**:
```go
// User actions
func (t *TUIHandler) SendMessage(deviceID, text string, dest uint32) error {
    cmd := SendTextCommand{Text: text, Destination: dest}
    return t.publishCommand(fmt.Sprintf("command.%s.send_text", deviceID), cmd)
}
```

### 2. MQTT Bridge

**Meshtastic → MQTT**:
```go
// Subscribe to mesh events, publish to MQTT
router.AddHandler("mqtt-tx", 
    "mesh.broadcast.packet.rx", 
    meshSub, 
    mqttPub, 
    mqttBridge.TranslateToMQTT)
```

**MQTT → Meshtastic**:
```go
// Subscribe to MQTT, publish commands
router.AddHandler("mqtt-rx", 
    "meshtastic/tx/+", 
    mqttSub, 
    meshPub, 
    mqttBridge.TranslateFromMQTT)
```

### 3. Persistence Layer

**Event Store**:
```go
// Store all events for replay/analytics
router.AddNoPublisherHandler("event-store", 
    "mesh.broadcast.*", 
    sub, 
    eventStore.StoreEvent)
```

**State Management**:
```go
// Maintain current state projections
router.AddNoPublisherHandler("state-projector", 
    "mesh.broadcast.nodeinfo.updated", 
    sub, 
    stateManager.UpdateNodeState)
```

## Implementation Plan

### Phase 1: Core Framework (Week 1-2)

**Deliverables**:
- `pkg/meshbus` - Watermill wrapper and topic definitions
- `pkg/deviceadapter` - Device adapter implementation
- `pkg/devicemgr` - Device supervisor
- `pkg/events` - Event schemas and marshaling
- `cmd/meshd` - Basic daemon

**Tasks**:
1. Implement DeviceAdapter with basic event publishing
2. Create DeviceSupervisor for multi-device management
3. Set up Watermill router with memory transport
4. Define core event types and schemas
5. Basic integration tests

### Phase 2: UI Integration (Week 3)

**Deliverables**:
- Updated TUI with event-driven updates
- Real-time message display
- Live node status updates
- Command execution via events

**Tasks**:
1. Refactor TUI models to subscribe to events
2. Replace direct client calls with command publishing
3. Implement real-time update mechanisms
4. Add event-driven state management

### Phase 3: MQTT Bridge (Week 4)

**Deliverables**:
- MQTT bridge implementation
- Bidirectional message translation
- Configuration management
- Topic mapping

**Tasks**:
1. Implement MQTT pub/sub integration
2. Create message translation layer
3. Add configuration for topic mapping
4. Test bidirectional communication

### Phase 4: Advanced Features (Week 5-6)

**Deliverables**:
- Event persistence layer
- Historical data access
- Metrics and monitoring
- Advanced routing rules

**Tasks**:
1. Add event store implementation
2. Implement state projections
3. Add Prometheus metrics
4. Create advanced routing configurations

## Performance Considerations

### Scalability
- **Memory usage**: ~1MB per device adapter
- **CPU usage**: Minimal overhead from event routing
- **Network**: Local events ~10µs, NATS ~1ms latency
- **Throughput**: 10K+ events/second per device

### Reliability
- **Back-pressure**: Configurable channel buffers
- **Circuit breakers**: Automatic failure isolation
- **Retry logic**: Exponential backoff for failed operations
- **Graceful shutdown**: Proper resource cleanup

### Resource Management
```go
type ResourceLimits struct {
    MaxDevices      int
    MaxEventsPerSec int
    BufferSize      int
    MaxMemoryMB     int
}
```

## Configuration Management

### Main Configuration
```yaml
# config/meshd.yaml
event_bus:
  transport: "memory"  # memory, nats, kafka
  buffer_size: 1000
  middlewares:
    - correlation
    - retry
    - circuit_breaker
    - metrics

device_supervisor:
  auto_discover: true
  max_devices: 10
  reconnect_interval: 5s

mqtt_bridge:
  enabled: true
  broker: "localhost:1883"
  topic_prefix: "meshtastic"
  qos: 1

persistence:
  enabled: false
  store_type: "sqlite"
  connection: "events.db"
```

## Security Considerations

### Authentication
- Device access control via port permissions
- MQTT broker authentication
- API key management for external integrations

### Data Protection
- Event encryption for sensitive data
- Audit logging for administrative actions
- Rate limiting for external connections

### Network Security
- TLS for MQTT connections
- mTLS for inter-service communication
- Network segmentation recommendations

## Testing Strategy

### Unit Tests
- Individual component testing
- Event serialization/deserialization
- Command handling logic
- Error scenarios

### Integration Tests
- End-to-end message flow
- Multi-device scenarios
- MQTT bridge functionality
- Persistence layer

### Performance Tests
- Load testing with multiple devices
- Memory usage under stress
- Event throughput benchmarks
- Network latency measurements

## Future Extensions

### Planned Features
- **Web Dashboard**: Real-time web interface
- **Mobile App**: Event-driven mobile client
- **Analytics**: Historical data analysis
- **Alerting**: Rule-based notifications
- **Clustering**: Multi-node deployments

### Integration Points
- **Kubernetes**: Cloud deployment
- **Prometheus**: Metrics collection
- **Grafana**: Visualization
- **Elasticsearch**: Log aggregation
- **Slack/Discord**: Notification integrations

## Conclusion

This event-driven architecture provides a scalable, maintainable foundation for Meshtastic device management. By leveraging Watermill's pub/sub patterns, we achieve loose coupling, high throughput, and extensibility while maintaining the robustness of the existing client implementation.

The phased implementation approach ensures incremental value delivery while building toward a comprehensive solution that can scale from single-device hobbyist use to large-scale deployments with hundreds of devices.
