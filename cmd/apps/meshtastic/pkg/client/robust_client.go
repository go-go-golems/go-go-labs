package client

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/serial/discovery"
)

// RobustMeshtasticClient represents a robust Meshtastic client
type RobustMeshtasticClient struct {
	SerialInterface

	// Configuration
	config *Config

	// State management
	stateHandler *DefaultStateHandler

	// Connection management
	connectionManager *ConnectionManager

	// Heartbeat
	heartbeatManager *HeartbeatManager
}

// NewRobustMeshtasticClient creates a new robust Meshtastic client
func NewRobustMeshtasticClient(config *Config) (*RobustMeshtasticClient, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return nil, errors.Wrap(err, "invalid configuration")
	}

	// Create serial configuration
	serialConfig := &SerialConfig{
		DevicePath:   config.DevicePath,
		BaudRate:     115200,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 1 * time.Second,
		StopBits:     1,
		DataBits:     8,
		Parity:       "N",
		FlowControl:  false,
		DisableHUPCL: true,
	}

	// Create serial client
	serialClient, err := NewSerialClient(serialConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create serial client")
	}

	// Create robust client
	client := &RobustMeshtasticClient{
		SerialInterface:   serialClient,
		config:            config,
		stateHandler:      NewDefaultStateHandler(),
		connectionManager: NewConnectionManager(serialClient),
		heartbeatManager:  NewHeartbeatManager(serialClient),
	}

	// Set up state handler
	serialClient.StreamClient.stateHandler = client.stateHandler

	// Set up connection manager
	client.connectionManager.SetOnStateChange(client.stateHandler.OnStateChange)

	return client, nil
}

// Connect connects to the device with robust error handling
func (rmc *RobustMeshtasticClient) Connect(ctx context.Context) error {
	log.Info().Str("device", rmc.config.DevicePath).Msg("Connecting to Meshtastic device")

	// Use connection manager for robust connection
	return rmc.connectionManager.Connect(ctx)
}

// Disconnect disconnects from the device
func (rmc *RobustMeshtasticClient) Disconnect() error {
	log.Info().Str("device", rmc.config.DevicePath).Msg("Disconnecting from Meshtastic device")

	// Stop heartbeat
	rmc.heartbeatManager.Stop()

	// Disconnect using connection manager
	return rmc.connectionManager.Disconnect()
}

// SendMessage sends a message with retry logic
func (rmc *RobustMeshtasticClient) SendMessage(packet *pb.MeshPacket) error {
	return rmc.connectionManager.SendMessageWithRetry(packet)
}

// SendText sends a text message with retry logic
func (rmc *RobustMeshtasticClient) SendText(text string, destination uint32) error {
	return rmc.connectionManager.SendTextWithRetry(text, destination)
}

// StartHeartbeat starts the heartbeat mechanism
func (rmc *RobustMeshtasticClient) StartHeartbeat() {
	rmc.heartbeatManager.Start()
}

// GetConnectionStatus returns detailed connection status
func (rmc *RobustMeshtasticClient) GetConnectionStatus() ConnectionStatus {
	return rmc.connectionManager.GetStatus()
}

// GetStatistics returns comprehensive statistics
func (rmc *RobustMeshtasticClient) GetStatistics() ClientStatistics {
	return ClientStatistics{
		Connection: rmc.SerialInterface.GetStatistics(),
		State:      rmc.stateHandler.GetStatistics(),
		Heartbeat:  rmc.heartbeatManager.GetStatistics(),
	}
}

// Close closes the robust client
func (rmc *RobustMeshtasticClient) Close() error {
	log.Info().Str("device", rmc.config.DevicePath).Msg("Closing robust Meshtastic client")

	// Stop heartbeat
	rmc.heartbeatManager.Stop()

	// Disconnect
	if err := rmc.Disconnect(); err != nil {
		log.Error().Err(err).Msg("Error during disconnect")
	}

	// Close underlying serial interface
	if rmc.SerialInterface != nil {
		return rmc.SerialInterface.Close()
	}

	return nil
}

// DefaultStateHandler handles state transitions
type DefaultStateHandler struct {
	mu               sync.RWMutex
	currentState     DeviceState
	previousState    DeviceState
	stateTransitions uint64
	stateHistory     []StateTransition
	maxHistorySize   int
	onStateChange    func(DeviceState, DeviceState)
}

// NewDefaultStateHandler creates a new state handler
func NewDefaultStateHandler() *DefaultStateHandler {
	return &DefaultStateHandler{
		currentState:   StateDisconnected,
		previousState:  StateDisconnected,
		maxHistorySize: 100,
		stateHistory:   make([]StateTransition, 0, 100),
	}
}

// OnStateChange handles state changes
func (sh *DefaultStateHandler) OnStateChange(oldState, newState DeviceState) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	sh.previousState = sh.currentState
	sh.currentState = newState
	sh.stateTransitions++

	// Add to history
	transition := StateTransition{
		From:      oldState,
		To:        newState,
		Timestamp: time.Now(),
	}

	sh.stateHistory = append(sh.stateHistory, transition)

	// Trim history if too long
	if len(sh.stateHistory) > sh.maxHistorySize {
		sh.stateHistory = sh.stateHistory[1:]
	}

	log.Info().
		Str("from", oldState.String()).
		Str("to", newState.String()).
		Uint64("transitions", sh.stateTransitions).
		Msg("State transition")

	// Call external handler
	if sh.onStateChange != nil {
		sh.onStateChange(oldState, newState)
	}
}

// GetCurrentState returns the current state
func (sh *DefaultStateHandler) GetCurrentState() DeviceState {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.currentState
}

// GetStatistics returns state handler statistics
func (sh *DefaultStateHandler) GetStatistics() StateStatistics {
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	return StateStatistics{
		CurrentState:  sh.currentState,
		PreviousState: sh.previousState,
		Transitions:   sh.stateTransitions,
		HistorySize:   len(sh.stateHistory),
		StateHistory:  append([]StateTransition(nil), sh.stateHistory...),
	}
}

// SetOnStateChange sets the external state change handler
func (sh *DefaultStateHandler) SetOnStateChange(handler func(DeviceState, DeviceState)) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	sh.onStateChange = handler
}

// StateTransition represents a state transition
type StateTransition struct {
	From      DeviceState
	To        DeviceState
	Timestamp time.Time
}

// StateStatistics holds state statistics
type StateStatistics struct {
	CurrentState  DeviceState
	PreviousState DeviceState
	Transitions   uint64
	HistorySize   int
	StateHistory  []StateTransition
}

// ConnectionManager manages connection lifecycle
type ConnectionManager struct {
	client             SerialInterface
	mu                 sync.RWMutex
	connectionAttempts uint64
	lastConnectionTime time.Time
	onStateChange      func(DeviceState, DeviceState)

	// Retry configuration
	maxRetries      int
	retryDelay      time.Duration
	maxRetryDelay   time.Duration
	retryMultiplier float64

	// Connection status
	status ConnectionStatus
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(client SerialInterface) *ConnectionManager {
	return &ConnectionManager{
		client:          client,
		maxRetries:      5,
		retryDelay:      1 * time.Second,
		maxRetryDelay:   30 * time.Second,
		retryMultiplier: 2.0,
		status: ConnectionStatus{
			State:       StateDisconnected,
			LastAttempt: time.Now(),
		},
	}
}

// Connect connects with retry logic
func (cm *ConnectionManager) Connect(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.connectionAttempts++
	cm.lastConnectionTime = time.Now()
	cm.status.LastAttempt = time.Now()

	delay := cm.retryDelay

	for attempt := 1; attempt <= cm.maxRetries; attempt++ {
		cm.status.Attempts = attempt

		log.Info().
			Int("attempt", attempt).
			Int("max", cm.maxRetries).
			Dur("delay", delay).
			Msg("Connection attempt")

		err := cm.client.Connect(ctx)
		if err == nil {
			cm.status.State = StateConnected
			cm.status.ConnectedAt = time.Now()
			cm.status.LastError = nil

			log.Info().
				Int("attempt", attempt).
				Msg("Successfully connected")

			return nil
		}

		cm.status.LastError = err

		if attempt < cm.maxRetries {
			log.Warn().
				Err(err).
				Int("attempt", attempt).
				Dur("delay", delay).
				Msg("Connection failed, retrying")

			// Wait with exponential backoff
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				delay = time.Duration(float64(delay) * cm.retryMultiplier)
				if delay > cm.maxRetryDelay {
					delay = cm.maxRetryDelay
				}
			}
		}
	}

	cm.status.State = StateError
	return errors.Wrapf(cm.status.LastError, "failed to connect after %d attempts", cm.maxRetries)
}

// Disconnect disconnects from the device
func (cm *ConnectionManager) Disconnect() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.status.State = StateDisconnected
	cm.status.DisconnectedAt = time.Now()

	return cm.client.Disconnect()
}

// SendMessageWithRetry sends a message with retry logic
func (cm *ConnectionManager) SendMessageWithRetry(packet *pb.MeshPacket) error {
	for attempt := 1; attempt <= cm.maxRetries; attempt++ {
		err := cm.client.SendMessage(packet)
		if err == nil {
			return nil
		}

		if attempt < cm.maxRetries {
			log.Warn().
				Err(err).
				Int("attempt", attempt).
				Msg("Message send failed, retrying")

			time.Sleep(cm.retryDelay)
		}
	}

	return errors.Errorf("failed to send message after %d attempts", cm.maxRetries)
}

// SendTextWithRetry sends a text message with retry logic
func (cm *ConnectionManager) SendTextWithRetry(text string, destination uint32) error {
	for attempt := 1; attempt <= cm.maxRetries; attempt++ {
		err := cm.client.SendText(text, destination)
		if err == nil {
			return nil
		}

		if attempt < cm.maxRetries {
			log.Warn().
				Err(err).
				Int("attempt", attempt).
				Msg("Text send failed, retrying")

			time.Sleep(cm.retryDelay)
		}
	}

	return errors.Errorf("failed to send text after %d attempts", cm.maxRetries)
}

// GetStatus returns connection status
func (cm *ConnectionManager) GetStatus() ConnectionStatus {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.status
}

// SetOnStateChange sets the state change handler
func (cm *ConnectionManager) SetOnStateChange(handler func(DeviceState, DeviceState)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onStateChange = handler
}

// ConnectionStatus represents connection status
type ConnectionStatus struct {
	State          DeviceState
	Attempts       int
	LastAttempt    time.Time
	ConnectedAt    time.Time
	DisconnectedAt time.Time
	LastError      error
}

// HeartbeatManager manages heartbeat functionality
type HeartbeatManager struct {
	client             SerialInterface
	mu                 sync.RWMutex
	active             bool
	interval           time.Duration
	timeout            time.Duration
	lastSent           time.Time
	lastReceived       time.Time
	heartbeatsSent     uint64
	heartbeatsReceived uint64
	missedHeartbeats   uint64

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc
}

// NewHeartbeatManager creates a new heartbeat manager
func NewHeartbeatManager(client SerialInterface) *HeartbeatManager {
	return &HeartbeatManager{
		client:   client,
		interval: 300 * time.Second, // 5 minutes
		timeout:  30 * time.Second,
	}
}

// Start starts the heartbeat mechanism
func (hm *HeartbeatManager) Start() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.active {
		return
	}

	hm.active = true
	hm.ctx, hm.cancel = context.WithCancel(context.Background())

	go hm.heartbeatLoop()

	log.Info().Dur("interval", hm.interval).Msg("Heartbeat started")
}

// Stop stops the heartbeat mechanism
func (hm *HeartbeatManager) Stop() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if !hm.active {
		return
	}

	hm.active = false
	hm.cancel()

	log.Info().Msg("Heartbeat stopped")
}

// heartbeatLoop runs the heartbeat loop
func (hm *HeartbeatManager) heartbeatLoop() {
	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-ticker.C:
			if hm.client.IsConnected() {
				hm.sendHeartbeat()
			}
		}
	}
}

// sendHeartbeat sends a heartbeat message
func (hm *HeartbeatManager) sendHeartbeat() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// Create heartbeat packet
	packet := &pb.MeshPacket{
		To:      BROADCAST_NUM,
		Id:      uint32(time.Now().Unix()),
		Channel: 0,
		PayloadVariant: &pb.MeshPacket_Decoded{
			Decoded: &pb.Data{
				Portnum: pb.PortNum_TEXT_MESSAGE_APP, // Use TEXT_MESSAGE_APP for ping
				Payload: []byte("ping"),
			},
		},
	}

	err := hm.client.SendMessage(packet)
	if err != nil {
		hm.missedHeartbeats++
		log.Warn().Err(err).Msg("Failed to send heartbeat")
	} else {
		hm.heartbeatsSent++
		hm.lastSent = time.Now()
		log.Debug().Msg("Heartbeat sent")
	}
}

// GetStatistics returns heartbeat statistics
func (hm *HeartbeatManager) GetStatistics() HeartbeatStatistics {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	return HeartbeatStatistics{
		Active:             hm.active,
		Interval:           hm.interval,
		Timeout:            hm.timeout,
		LastSent:           hm.lastSent,
		LastReceived:       hm.lastReceived,
		HeartbeatsSent:     hm.heartbeatsSent,
		HeartbeatsReceived: hm.heartbeatsReceived,
		MissedHeartbeats:   hm.missedHeartbeats,
	}
}

// HeartbeatStatistics holds heartbeat statistics
type HeartbeatStatistics struct {
	Active             bool
	Interval           time.Duration
	Timeout            time.Duration
	LastSent           time.Time
	LastReceived       time.Time
	HeartbeatsSent     uint64
	HeartbeatsReceived uint64
	MissedHeartbeats   uint64
}

// ClientStatistics holds comprehensive client statistics
type ClientStatistics struct {
	Connection ConnectionStatistics
	State      StateStatistics
	Heartbeat  HeartbeatStatistics
}

// ValidateConfig validates client configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	if config.DevicePath == "" {
		// Try to auto-discover
		devicePath, err := discovery.FindBestMeshtasticPort()
		if err != nil {
			return errors.Wrap(err, "device path not specified and auto-discovery failed")
		}
		config.DevicePath = devicePath
	}

	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	return nil
}

// CreateRobustClient creates a robust client with sensible defaults
func CreateRobustClient(devicePath string) (*RobustMeshtasticClient, error) {
	config := &Config{
		DevicePath:  devicePath,
		Timeout:     30 * time.Second,
		DebugSerial: false,
		HexDump:     false,
	}

	return NewRobustMeshtasticClient(config)
}

// AutoDiscoverAndConnect automatically discovers a device and connects
func AutoDiscoverAndConnect(ctx context.Context) (*RobustMeshtasticClient, error) {
	// Discover device
	devicePath, err := discovery.FindBestMeshtasticPort()
	if err != nil {
		return nil, errors.Wrap(err, "failed to discover Meshtastic device")
	}

	// Create client
	client, err := CreateRobustClient(devicePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	// Connect
	if err := client.Connect(ctx); err != nil {
		client.Close()
		return nil, errors.Wrap(err, "failed to connect")
	}

	return client, nil
}
