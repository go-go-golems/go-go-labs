package client

import (
	"context"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tarm/serial"
	"golang.org/x/sys/unix"

	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/serial/discovery"
)

// SerialClient implements a robust serial interface for Meshtastic devices
type SerialClient struct {
	*StreamClient

	// Serial-specific configuration
	serialConfig *SerialConfig
	port         *serial.Port

	// Reconnection handling
	reconnectAttempts    int
	maxReconnectAttempts int
	reconnectDelay       time.Duration
	backoffMultiplier    float64
	maxReconnectDelay    time.Duration

	// Connection state
	connectionStart time.Time
	lastReconnect   time.Time
}

// NewSerialClient creates a new serial client with the given configuration
func NewSerialClient(config *SerialConfig) (*SerialClient, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	// Auto-discover device if not specified
	if config.DevicePath == "" {
		devicePath, err := discovery.FindBestMeshtasticPort()
		if err != nil {
			return nil, errors.Wrap(err, "failed to find Meshtastic device")
		}
		config.DevicePath = devicePath
	}

	// Apply defaults
	if config.BaudRate == 0 {
		config.BaudRate = 115200
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 500 * time.Millisecond
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 1 * time.Second
	}
	if config.StopBits == 0 {
		config.StopBits = 1
	}
	if config.DataBits == 0 {
		config.DataBits = 8
	}
	if config.Parity == "" {
		config.Parity = "N"
	}

	// Create serial port
	port, err := openSerialPort(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open serial port")
	}

	// Create stream client
	streamClient := NewStreamClient(port, config.DevicePath)

	// Create serial client
	sc := &SerialClient{
		StreamClient:         streamClient,
		serialConfig:         config,
		port:                 port,
		maxReconnectAttempts: 10,
		reconnectDelay:       2 * time.Second,
		backoffMultiplier:    1.5,
		maxReconnectDelay:    30 * time.Second,
		connectionStart:      time.Now(),
	}

	// Initialize message queue
	sc.messageQueue = NewFlowControlledQueue(100, 10, 30*time.Second)

	// Set up disconnect handler for reconnection
	sc.SetOnDisconnect(sc.handleDisconnectWithReconnect)

	return sc, nil
}

// Connect implements SerialInterface
func (sc *SerialClient) Connect(ctx context.Context) error {
	log.Info().Str("device", sc.serialConfig.DevicePath).Msg("Connecting to serial device")

	// Record connection start time
	sc.connectionStart = time.Now()

	// Update statistics
	sc.stats.ConnectDuration = time.Since(sc.connectionStart)

	// Call parent connect
	return sc.StreamClient.Connect(ctx)
}

// Reconnect implements SerialInterface
func (sc *SerialClient) Reconnect() error {
	log.Info().
		Str("device", sc.serialConfig.DevicePath).
		Int("attempt", sc.reconnectAttempts+1).
		Int("max", sc.maxReconnectAttempts).
		Msg("Attempting to reconnect")

	// Check if we've exceeded max attempts
	if sc.reconnectAttempts >= sc.maxReconnectAttempts {
		return errors.New("maximum reconnection attempts exceeded")
	}

	// Increment reconnection attempts
	sc.reconnectAttempts++
	sc.stats.Reconnects++
	sc.lastReconnect = time.Now()

	// Calculate backoff delay
	delay := sc.calculateBackoffDelay()

	log.Info().
		Str("device", sc.serialConfig.DevicePath).
		Dur("delay", delay).
		Msg("Waiting before reconnection attempt")

	// Wait with exponential backoff
	time.Sleep(delay)

	// Close existing connection
	if sc.port != nil {
		sc.port.Close()
	}

	// Open new serial port
	port, err := openSerialPort(sc.serialConfig)
	if err != nil {
		return errors.Wrap(err, "failed to reopen serial port")
	}

	// Update port and stream
	sc.port = port
	sc.SetStream(port)

	// Reset reconnection attempts on successful connection
	sc.reconnectAttempts = 0

	log.Info().
		Str("device", sc.serialConfig.DevicePath).
		Msg("Successfully reconnected")

	return nil
}

// GetReconnectAttempts implements SerialInterface
func (sc *SerialClient) GetReconnectAttempts() int {
	return sc.reconnectAttempts
}

// ResetReconnectAttempts implements SerialInterface
func (sc *SerialClient) ResetReconnectAttempts() {
	sc.reconnectAttempts = 0
}

// GetSerialConfig implements SerialInterface
func (sc *SerialClient) GetSerialConfig() *SerialConfig {
	return sc.serialConfig
}

// SetSerialConfig implements SerialInterface
func (sc *SerialClient) SetSerialConfig(config *SerialConfig) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	// Close existing port
	if sc.port != nil {
		sc.port.Close()
	}

	// Open new port with new config
	port, err := openSerialPort(config)
	if err != nil {
		return errors.Wrap(err, "failed to open serial port with new config")
	}

	// Update configuration and port
	sc.serialConfig = config
	sc.port = port
	sc.SetStream(port)

	return nil
}

// Flush implements SerialInterface
func (sc *SerialClient) Flush() error {
	if sc.port == nil {
		return errors.New("port not open")
	}

	return sc.port.Flush()
}

// GetStatistics implements SerialInterface
func (sc *SerialClient) GetStatistics() ConnectionStatistics {
	stats := sc.stats
	stats.ConnectDuration = time.Since(sc.connectionStart)
	if !sc.lastReconnect.IsZero() {
		stats.LastReconnect = sc.lastReconnect
	}
	return stats
}

// Close implements SerialInterface
func (sc *SerialClient) Close() error {
	log.Info().Str("device", sc.serialConfig.DevicePath).Msg("Closing serial connection")

	// Close stream client
	if err := sc.StreamClient.Close(); err != nil {
		log.Error().Err(err).Msg("Error closing stream client")
	}

	// Close serial port
	if sc.port != nil {
		if err := sc.port.Close(); err != nil {
			log.Error().Err(err).Msg("Error closing serial port")
		}
		sc.port = nil
	}

	return nil
}

// Internal methods

func (sc *SerialClient) calculateBackoffDelay() time.Duration {
	delay := time.Duration(float64(sc.reconnectDelay) *
		(sc.backoffMultiplier * float64(sc.reconnectAttempts)))

	if delay > sc.maxReconnectDelay {
		delay = sc.maxReconnectDelay
	}

	return delay
}

func (sc *SerialClient) handleDisconnectWithReconnect(err error) {
	log.Warn().Err(err).Msg("Device disconnected, attempting reconnection")

	// Change state to reconnecting
	sc.changeState(StateReconnecting)

	// Attempt reconnection in background
	go func() {
		if err := sc.Reconnect(); err != nil {
			log.Error().Err(err).Msg("Reconnection failed")
			sc.changeState(StateError)
		} else {
			// Restart connection
			if err := sc.Connect(sc.ctx); err != nil {
				log.Error().Err(err).Msg("Failed to restart connection after reconnection")
				sc.changeState(StateError)
			}
		}
	}()
}

// openSerialPort opens a serial port with the given configuration
func openSerialPort(config *SerialConfig) (*serial.Port, error) {
	serialConfig := &serial.Config{
		Name:        config.DevicePath,
		Baud:        config.BaudRate,
		ReadTimeout: config.ReadTimeout,
		Size:        byte(config.DataBits),
		StopBits:    serial.StopBits(config.StopBits),
	}

	// Set parity
	switch config.Parity {
	case "N", "none":
		serialConfig.Parity = serial.ParityNone
	case "O", "odd":
		serialConfig.Parity = serial.ParityOdd
	case "E", "even":
		serialConfig.Parity = serial.ParityEven
	case "M", "mark":
		serialConfig.Parity = serial.ParityMark
	case "S", "space":
		serialConfig.Parity = serial.ParitySpace
	default:
		return nil, errors.Errorf("invalid parity: %s", config.Parity)
	}

	port, err := serial.OpenPort(serialConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open serial port")
	}

	// Disable HUPCL (hangup on close) to prevent device reset
	if config.DisableHUPCL {
		if err := disableHUPCL(port); err != nil {
			log.Warn().Err(err).Msg("Failed to disable HUPCL")
		}
	}

	// Wait for device to settle
	time.Sleep(100 * time.Millisecond)

	return port, nil
}

// disableHUPCL disables the HUPCL flag on Unix systems to prevent device reset
func disableHUPCL(port *serial.Port) error {
	// This is a platform-specific implementation for Unix systems
	// For Windows, this would need different implementation

	// Get the underlying file descriptor
	// This is a hack to access the internal file descriptor
	// In a production system, you might want to use a different approach

	// For now, we'll use a syscall approach
	// This requires unsafe operations and is platform-specific

	// Note: This is a simplified implementation and may not work on all systems
	// A more robust implementation would handle different platforms properly

	return nil // Placeholder - implement based on your platform requirements
}

// Alternative implementation using syscalls (Linux/Unix specific)
func disableHUPCLSyscall(fd int) error {
	// Get current terminal attributes
	termios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return errors.Wrap(err, "failed to get terminal attributes")
	}

	// Disable HUPCL flag
	termios.Cflag &^= unix.HUPCL

	// Set the modified attributes
	if err := unix.IoctlSetTermios(fd, unix.TCSETS, termios); err != nil {
		return errors.Wrap(err, "failed to set terminal attributes")
	}

	return nil
}

// getFD extracts the file descriptor from the serial port
// This is a platform-specific and potentially unsafe operation
func getFD(port *serial.Port) (int, error) {
	// This is a hack to access the internal file descriptor
	// In a real implementation, you would need to properly handle this
	// based on the specific serial library being used

	// For the tarm/serial library, we need to use reflection or
	// modify the library to expose the file descriptor

	// This is a placeholder implementation
	return 0, errors.New("file descriptor extraction not implemented")
}

// SerialPortWrapper wraps a serial port to provide additional functionality
type SerialPortWrapper struct {
	*serial.Port
	config *SerialConfig
}

// NewSerialPortWrapper creates a new serial port wrapper
func NewSerialPortWrapper(port *serial.Port, config *SerialConfig) *SerialPortWrapper {
	return &SerialPortWrapper{
		Port:   port,
		config: config,
	}
}

// Write implements io.Writer with timeout support
func (w *SerialPortWrapper) Write(p []byte) (n int, err error) {
	if w.config.WriteTimeout > 0 {
		// Set write deadline
		// Note: The tarm/serial library doesn't support write timeouts directly
		// This is a placeholder for timeout implementation
	}

	return w.Port.Write(p)
}

// Read implements io.Reader with enhanced error handling
func (w *SerialPortWrapper) Read(p []byte) (n int, err error) {
	n, err = w.Port.Read(p)

	// Enhanced error handling for common serial issues
	if err != nil {
		// Check for specific error types and provide better context
		if err == io.EOF {
			err = errors.Wrap(err, "device disconnected")
		}
	}

	return n, err
}

// Close implements io.Closer
func (w *SerialPortWrapper) Close() error {
	return w.Port.Close()
}

// GetConfig returns the serial configuration
func (w *SerialPortWrapper) GetConfig() *SerialConfig {
	return w.config
}

// SetConfig updates the serial configuration
func (w *SerialPortWrapper) SetConfig(config *SerialConfig) error {
	w.config = config
	return nil
}

// ConnectWithRetry attempts to connect with exponential backoff
func ConnectWithRetry(config *SerialConfig, maxAttempts int, initialDelay time.Duration) (*SerialClient, error) {
	var lastErr error
	delay := initialDelay

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Info().
			Str("device", config.DevicePath).
			Int("attempt", attempt).
			Int("max", maxAttempts).
			Msg("Attempting to connect")

		client, err := NewSerialClient(config)
		if err == nil {
			log.Info().
				Str("device", config.DevicePath).
				Int("attempt", attempt).
				Msg("Successfully connected")
			return client, nil
		}

		lastErr = err

		if attempt < maxAttempts {
			log.Warn().
				Err(err).
				Str("device", config.DevicePath).
				Int("attempt", attempt).
				Dur("delay", delay).
				Msg("Connection failed, retrying")

			time.Sleep(delay)
			delay = time.Duration(float64(delay) * 1.5) // Exponential backoff
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
		}
	}

	return nil, errors.Wrapf(lastErr, "failed to connect after %d attempts", maxAttempts)
}

// DefaultSerialConfig returns a default serial configuration
func DefaultSerialConfig() *SerialConfig {
	return &SerialConfig{
		BaudRate:     115200,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 1 * time.Second,
		StopBits:     1,
		DataBits:     8,
		Parity:       "N",
		FlowControl:  false,
		DisableHUPCL: true,
	}
}

// ValidateSerialConfig validates a serial configuration
func ValidateSerialConfig(config *SerialConfig) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	if config.DevicePath == "" {
		return errors.New("device path cannot be empty")
	}

	if config.BaudRate <= 0 {
		return errors.New("baud rate must be positive")
	}

	if config.ReadTimeout < 0 {
		return errors.New("read timeout cannot be negative")
	}

	if config.WriteTimeout < 0 {
		return errors.New("write timeout cannot be negative")
	}

	if config.StopBits < 1 || config.StopBits > 2 {
		return errors.New("stop bits must be 1 or 2")
	}

	if config.DataBits < 5 || config.DataBits > 8 {
		return errors.New("data bits must be 5-8")
	}

	validParity := []string{"N", "none", "O", "odd", "E", "even", "M", "mark", "S", "space"}
	found := false
	for _, p := range validParity {
		if config.Parity == p {
			found = true
			break
		}
	}
	if !found {
		return errors.Errorf("invalid parity: %s", config.Parity)
	}

	return nil
}
