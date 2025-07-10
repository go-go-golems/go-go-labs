package meshbus

import (
	"context"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Bus represents the main event bus using Watermill
type Bus struct {
	router       *message.Router
	pubSub       *gochannel.GoChannel
	logger       watermill.LoggerAdapter
	middlewares  []message.HandlerMiddleware
	mu           sync.RWMutex
	running      bool
	ctx          context.Context
	cancel       context.CancelFunc
}

// Config holds configuration for the event bus
type Config struct {
	Logger             watermill.LoggerAdapter
	Middlewares        []message.HandlerMiddleware
	PublishTimeout     time.Duration
	SubscribeTimeout   time.Duration
	RouterCloseTimeout time.Duration
	BufferSize         int
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Logger:             NewZerologAdapter(),
		Middlewares:        []message.HandlerMiddleware{},
		PublishTimeout:     10 * time.Second,
		SubscribeTimeout:   30 * time.Second,
		RouterCloseTimeout: 30 * time.Second,
		BufferSize:         1000,
	}
}

// NewBus creates a new event bus
func NewBus(config *Config) (*Bus, error) {
	if config == nil {
		config = DefaultConfig()
	}

	pubSub := gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: int64(config.BufferSize),
		},
		config.Logger,
	)

	router, err := message.NewRouter(message.RouterConfig{
		CloseTimeout: config.RouterCloseTimeout,
	}, config.Logger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create router")
	}

	// Add default middlewares
	defaultMiddlewares := []message.HandlerMiddleware{
		NewCorrelationMiddleware(),
		NewLoggingMiddleware(),
		NewRecoveryMiddleware(),
	}

	// Combine with custom middlewares
	allMiddlewares := append(defaultMiddlewares, config.Middlewares...)
	router.AddMiddleware(allMiddlewares...)

	ctx, cancel := context.WithCancel(context.Background())

	return &Bus{
		router:      router,
		pubSub:      pubSub,
		logger:      config.Logger,
		middlewares: allMiddlewares,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Start starts the event bus
func (b *Bus) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		return errors.New("bus is already running")
	}

	log.Info().Msg("Starting event bus")

	// Start the router
	go func() {
		if err := b.router.Run(b.ctx); err != nil {
			log.Error().Err(err).Msg("Router stopped with error")
		}
	}()

	// Wait for router to be ready
	<-b.router.Running()

	b.running = true
	log.Info().Msg("Event bus started successfully")

	return nil
}

// Stop stops the event bus
func (b *Bus) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return nil
	}

	log.Info().Msg("Stopping event bus")

	// Cancel context to stop router
	b.cancel()

	// Close pubsub
	if err := b.pubSub.Close(); err != nil {
		log.Error().Err(err).Msg("Error closing pubsub")
	}

	b.running = false
	log.Info().Msg("Event bus stopped")

	return nil
}

// Publisher returns the publisher
func (b *Bus) Publisher() message.Publisher {
	return b.pubSub
}

// Subscriber returns the subscriber
func (b *Bus) Subscriber() message.Subscriber {
	return b.pubSub
}

// AddHandler adds a handler to the router
func (b *Bus) AddHandler(handlerName, subscribeTopic string, handler message.NoPublishHandlerFunc) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.running {
		return errors.New("cannot add handler to running bus")
	}

	b.router.AddNoPublisherHandler(
		handlerName,
		subscribeTopic,
		b.pubSub,
		handler,
	)

	return nil
}

// AddRouterHandler adds a handler that can publish to other topics
func (b *Bus) AddRouterHandler(handlerName, subscribeTopic, publishTopic string, handler message.HandlerFunc) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.running {
		return errors.New("cannot add handler to running bus")
	}

	b.router.AddHandler(
		handlerName,
		subscribeTopic,
		b.pubSub,
		publishTopic,
		b.pubSub,
		handler,
	)

	return nil
}

// IsRunning returns true if the bus is running
func (b *Bus) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.running
}

// WaitForReady waits for the bus to be ready
func (b *Bus) WaitForReady(timeout time.Duration) error {
	if !b.IsRunning() {
		return errors.New("bus is not running")
	}

	select {
	case <-b.router.Running():
		return nil
	case <-time.After(timeout):
		return errors.New("timeout waiting for bus to be ready")
	}
}

// Topic constants
const (
	// Device lifecycle events
	TopicDeviceConnected     = "device.connected"
	TopicDeviceDisconnected  = "device.disconnected"
	TopicDeviceReconnecting  = "device.reconnecting"
	TopicDeviceError         = "device.error"

	// Mesh packet events
	TopicMeshPacketRx       = "mesh.packet.rx"
	TopicMeshPacketTx       = "mesh.packet.tx"
	TopicMeshPacketAck      = "mesh.packet.ack"
	TopicMeshPacketTimeout  = "mesh.packet.timeout"

	// Node events
	TopicNodeInfoUpdated    = "mesh.nodeinfo.updated"
	TopicNodePresence       = "mesh.node.presence"
	TopicNodeBattery        = "mesh.node.battery"

	// Telemetry events
	TopicTelemetryReceived  = "mesh.telemetry.received"
	TopicPositionUpdated    = "mesh.position.updated"
	TopicEnvironmentUpdated = "mesh.environment.updated"

	// Command events
	TopicCommandSendText       = "command.send_text"
	TopicCommandRequestInfo    = "command.request_info"
	TopicCommandRequestTelemetry = "command.request_telemetry"
	TopicCommandRequestPosition  = "command.request_position"

	// Response events
	TopicResponseSuccess = "response.success"
	TopicResponseError   = "response.error"
	TopicResponseTimeout = "response.timeout"

	// System events
	TopicSystemStartup   = "system.startup"
	TopicSystemShutdown  = "system.shutdown"
	TopicSystemError     = "system.error"
)

// BuildTopicName builds a topic name with device ID
func BuildTopicName(baseTopic, deviceID string) string {
	if deviceID == "" {
		return baseTopic
	}
	return baseTopic + "." + deviceID
}

// BuildBroadcastTopic builds a broadcast topic name
func BuildBroadcastTopic(baseTopic string) string {
	return "broadcast." + baseTopic
}
