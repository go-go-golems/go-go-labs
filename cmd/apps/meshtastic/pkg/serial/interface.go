package serial

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tarm/serial"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/protocol"
)

// SerialInterface handles communication with a Meshtastic device over serial
type SerialInterface struct {
	port    *serial.Port
	parser  *protocol.FrameParser
	builder *protocol.FrameBuilder
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.RWMutex

	// Handlers
	onFromRadio  func(*pb.FromRadio)
	onLogOutput  func(string)
	onDisconnect func(error)

	// Configuration
	devicePath  string
	debugOutput io.Writer

	// State
	connected bool
	logBuffer []byte
}

// Config represents configuration for the serial interface
type Config struct {
	DevicePath  string
	Baud        int
	ReadTimeout time.Duration
	DebugOutput io.Writer
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Baud:        115200,
		ReadTimeout: 500 * time.Millisecond,
	}
}

// NewSerialInterface creates a new serial interface
func NewSerialInterface(config *Config) (*SerialInterface, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Auto-discover device if not specified
	if config.DevicePath == "" {
		devicePath, err := FindBestMeshtasticPort()
		if err != nil {
			return nil, errors.Wrap(err, "failed to find Meshtastic device")
		}
		config.DevicePath = devicePath
	}

	ctx, cancel := context.WithCancel(context.Background())

	si := &SerialInterface{
		ctx:         ctx,
		cancel:      cancel,
		devicePath:  config.DevicePath,
		debugOutput: config.DebugOutput,
		logBuffer:   make([]byte, 0, 1024),
	}

	// Create frame parser
	si.parser = protocol.NewFrameParser(si.handleFrame, si.handleLogByte)
	if config.DebugOutput != nil {
		si.parser.SetDebugOutput(config.DebugOutput)
	}

	// Create frame builder
	si.builder = protocol.NewFrameBuilder()
	if config.DebugOutput != nil {
		si.builder.SetDebugOutput(config.DebugOutput)
	}

	// Open serial port
	if err := si.connect(config); err != nil {
		return nil, err
	}

	return si, nil
}

// connect establishes the serial connection
func (si *SerialInterface) connect(config *Config) error {
	si.mu.Lock()
	defer si.mu.Unlock()

	log.Info().Str("device", config.DevicePath).Msg("Connecting to Meshtastic device")

	// Configure serial port
	serialConfig := &serial.Config{
		Name:        config.DevicePath,
		Baud:        config.Baud,
		ReadTimeout: config.ReadTimeout,
	}

	port, err := serial.OpenPort(serialConfig)
	if err != nil {
		return errors.Wrapf(err, "failed to open serial port %s", config.DevicePath)
	}

	si.port = port
	si.connected = true

	// Start reader goroutine
	si.wg.Add(1)
	go si.readerLoop()

	log.Info().Str("device", config.DevicePath).Msg("Connected to Meshtastic device")
	return nil
}

// readerLoop reads data from the serial port
func (si *SerialInterface) readerLoop() {
	defer si.wg.Done()

	log.Debug().Msg("Starting serial reader loop")

	buffer := make([]byte, 256)

	for {
		select {
		case <-si.ctx.Done():
			log.Debug().Msg("Serial reader loop cancelled")
			return
		default:
			// Read from serial port
			n, err := si.port.Read(buffer)
			if err != nil {
				if si.ctx.Err() == nil {
					log.Error().Err(err).Msg("Serial read error")
					si.handleDisconnect(err)
				}
				return
			}

			if n > 0 {
				// Process received bytes
				si.parser.ProcessBytes(buffer[:n])
			}
		}
	}
}

// handleFrame handles a complete frame received from the device
func (si *SerialInterface) handleFrame(frame *protocol.Frame) {
	// Parse as FromRadio message
	fromRadio, err := frame.FromRadioMessage()
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse FromRadio message")
		return
	}

	// Call handler if set
	if si.onFromRadio != nil {
		si.onFromRadio(fromRadio)
	}
}

// handleLogByte handles a single byte that might be part of debug output
func (si *SerialInterface) handleLogByte(b byte) {
	si.logBuffer = append(si.logBuffer, b)

	// Check for complete lines
	if b == '\n' || b == '\r' {
		if len(si.logBuffer) > 1 {
			// Convert to string and remove trailing newline
			line := string(si.logBuffer[:len(si.logBuffer)-1])
			if si.onLogOutput != nil {
				si.onLogOutput(line)
			}
		}
		si.logBuffer = si.logBuffer[:0]
	}

	// Prevent buffer overflow
	if len(si.logBuffer) > 1024 {
		si.logBuffer = si.logBuffer[:0]
	}
}

// handleDisconnect handles device disconnection
func (si *SerialInterface) handleDisconnect(err error) {
	si.mu.Lock()
	si.connected = false
	si.mu.Unlock()

	if si.onDisconnect != nil {
		si.onDisconnect(err)
	}
}

// SendToRadio sends a ToRadio message to the device
func (si *SerialInterface) SendToRadio(toRadio *pb.ToRadio) error {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if !si.connected {
		return errors.New("device not connected")
	}

	// Build frame
	frameData, err := si.builder.BuildFrame(toRadio)
	if err != nil {
		return errors.Wrap(err, "failed to build frame")
	}

	// Write to serial port
	n, err := si.port.Write(frameData)
	if err != nil {
		return errors.Wrap(err, "failed to write to serial port")
	}

	if n != len(frameData) {
		return errors.Errorf("partial write: %d of %d bytes", n, len(frameData))
	}

	return nil
}

// IsConnected returns true if the device is connected
func (si *SerialInterface) IsConnected() bool {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.connected
}

// DevicePath returns the device path
func (si *SerialInterface) DevicePath() string {
	return si.devicePath
}

// SetOnFromRadio sets the handler for FromRadio messages
func (si *SerialInterface) SetOnFromRadio(handler func(*pb.FromRadio)) {
	si.onFromRadio = handler
}

// SetOnLogOutput sets the handler for device log output
func (si *SerialInterface) SetOnLogOutput(handler func(string)) {
	si.onLogOutput = handler
}

// SetOnDisconnect sets the handler for device disconnection
func (si *SerialInterface) SetOnDisconnect(handler func(error)) {
	si.onDisconnect = handler
}

// Close closes the serial interface
func (si *SerialInterface) Close() error {
	log.Info().Msg("Closing serial interface")

	// Cancel context to stop goroutines
	si.cancel()

	// Wait for goroutines to finish
	si.wg.Wait()

	// Close serial port
	si.mu.Lock()
	defer si.mu.Unlock()

	if si.port != nil {
		err := si.port.Close()
		si.port = nil
		si.connected = false
		return err
	}

	return nil
}

// Flush flushes the serial port
func (si *SerialInterface) Flush() error {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if !si.connected {
		return errors.New("device not connected")
	}

	return si.port.Flush()
}
