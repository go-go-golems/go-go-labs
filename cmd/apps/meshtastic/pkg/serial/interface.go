package serial

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tarm/serial"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/protocol"
)

// hexDump creates a hexdump-style representation of data
func hexDump(data []byte, maxBytes int) string {
	if len(data) == 0 {
		return ""
	}

	// Limit data to maxBytes for display
	displayData := data
	if len(displayData) > maxBytes {
		displayData = data[:maxBytes]
	}

	var result strings.Builder
	for i := 0; i < len(displayData); i += 16 {
		// Hex bytes
		hexPart := make([]string, 0, 16)
		asciiPart := make([]byte, 0, 16)

		for j := 0; j < 16 && i+j < len(displayData); j++ {
			b := displayData[i+j]
			hexPart = append(hexPart, fmt.Sprintf("%02x", b))
			if b >= 32 && b < 127 {
				asciiPart = append(asciiPart, b)
			} else {
				asciiPart = append(asciiPart, '.')
			}
		}

		// Format line
		if i > 0 {
			result.WriteString(" ")
		}
		result.WriteString(fmt.Sprintf("%04x: %-48s |%s|", i, strings.Join(hexPart, " "), string(asciiPart)))
		if i+16 < len(displayData) {
			result.WriteString("\n")
		}
	}

	if len(data) > maxBytes {
		result.WriteString(fmt.Sprintf("\n... (%d more bytes)", len(data)-maxBytes))
	}

	return result.String()
}

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
	debugSerial bool
	hexDump     bool

	// State
	connected            bool
	logBuffer            []byte
	reconnectAttempts    int
	maxReconnectAttempts int
	reconnectDelay       time.Duration

	// Debug tracking
	bytesRead    uint64
	bytesWritten uint64
	readErrors   uint64
	writeErrors  uint64
	reconnects   uint64
}

// Config represents configuration for the serial interface
type Config struct {
	DevicePath  string
	Baud        int
	ReadTimeout time.Duration
	DebugOutput io.Writer
	DebugSerial bool
	HexDump     bool
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
		ctx:                  ctx,
		cancel:               cancel,
		devicePath:           config.DevicePath,
		debugOutput:          config.DebugOutput,
		debugSerial:          config.DebugSerial,
		hexDump:              config.HexDump,
		logBuffer:            make([]byte, 0, 1024),
		maxReconnectAttempts: 5,
		reconnectDelay:       2 * time.Second,
	}

	// Create frame parser
	si.parser = protocol.NewFrameParser(si.handleFrame, si.handleLogByte)
	if config.DebugOutput != nil {
		si.parser.SetDebugOutput(config.DebugOutput)
	}
	si.parser.SetDebugSerial(config.DebugSerial)
	si.parser.SetHexDump(config.HexDump)

	// Create frame builder
	si.builder = protocol.NewFrameBuilder()
	if config.DebugOutput != nil {
		si.builder.SetDebugOutput(config.DebugOutput)
	}
	si.builder.SetDebugSerial(config.DebugSerial)
	si.builder.SetHexDump(config.HexDump)

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

	if si.debugSerial {
		log.Info().
			Str("device", config.DevicePath).
			Int("baud", config.Baud).
			Dur("timeout", config.ReadTimeout).
			Msg("Attempting to connect to Meshtastic device")
	} else {
		log.Info().Str("device", config.DevicePath).Msg("Connecting to Meshtastic device")
	}

	// Configure serial port
	serialConfig := &serial.Config{
		Name:        config.DevicePath,
		Baud:        config.Baud,
		ReadTimeout: config.ReadTimeout,
	}

	startTime := time.Now()
	port, err := serial.OpenPort(serialConfig)
	connectDuration := time.Since(startTime)

	if err != nil {
		if si.debugSerial {
			log.Error().
				Err(err).
				Str("device", config.DevicePath).
				Dur("duration", connectDuration).
				Msg("Failed to open serial port")
		}
		return errors.Wrapf(err, "failed to open serial port %s", config.DevicePath)
	}

	si.port = port
	si.connected = true
	si.reconnectAttempts = 0

	// Start reader goroutine
	si.wg.Add(1)
	go si.readerLoop()

	if si.debugSerial {
		log.Info().
			Str("device", config.DevicePath).
			Dur("duration", connectDuration).
			Msg("Successfully connected to Meshtastic device")
	} else {
		log.Info().Str("device", config.DevicePath).Msg("Connected to Meshtastic device")
	}
	return nil
}

// readerLoop reads data from the serial port with reconnection logic
func (si *SerialInterface) readerLoop() {
	defer si.wg.Done()

	if si.debugSerial {
		log.Debug().
			Uint64("totalBytesRead", si.bytesRead).
			Uint64("totalBytesWritten", si.bytesWritten).
			Uint64("totalReadErrors", si.readErrors).
			Uint64("totalWriteErrors", si.writeErrors).
			Uint64("totalReconnects", si.reconnects).
			Msg("Starting serial reader loop")
	} else {
		log.Debug().Msg("Starting serial reader loop")
	}

	// Use larger buffer to handle burst data
	buffer := make([]byte, 1024)

	for {
		select {
		case <-si.ctx.Done():
			if si.debugSerial {
				log.Debug().
					Uint64("totalBytesRead", si.bytesRead).
					Uint64("totalBytesWritten", si.bytesWritten).
					Uint64("totalReadErrors", si.readErrors).
					Uint64("totalWriteErrors", si.writeErrors).
					Uint64("totalReconnects", si.reconnects).
					Msg("Serial reader loop cancelled")
			} else {
				log.Debug().Msg("Serial reader loop cancelled")
			}
			return
		default:
			// Check if we're connected
			si.mu.RLock()
			connected := si.connected
			port := si.port
			si.mu.RUnlock()

			if !connected || port == nil {
				if si.debugSerial {
					log.Debug().
						Bool("connected", connected).
						Bool("portNil", port == nil).
						Msg("Not connected, attempting reconnection")
				}
				// Try to reconnect
				if si.ctx.Err() == nil {
					si.attemptReconnect()
				}
				time.Sleep(si.reconnectDelay)
				continue
			}

			// Read from serial port with timeout handling
			readStart := time.Now()
			n, err := port.Read(buffer)
			readDuration := time.Since(readStart)

			if err != nil {
				si.readErrors++
				if si.ctx.Err() == nil {
					if si.debugSerial {
						log.Error().
							Err(err).
							Dur("readDuration", readDuration).
							Uint64("totalReadErrors", si.readErrors).
							Uint64("totalBytesRead", si.bytesRead).
							Msg("Serial read error")
					} else {
						log.Error().Err(err).Msg("Serial read error")
					}

					// Handle EOF specifically - this is common when device disconnects
					if err == io.EOF {
						if si.debugSerial {
							log.Warn().
								Dur("readDuration", readDuration).
								Uint64("totalBytesRead", si.bytesRead).
								Msg("Device disconnected (EOF), attempting reconnection")
						} else {
							log.Warn().Msg("Device disconnected (EOF), attempting reconnection")
						}
						si.handleDisconnect(err)
						continue // Don't exit, try to reconnect
					}

					// Handle other errors
					si.handleDisconnect(err)
				}
				return
			}

			if n > 0 {
				si.bytesRead += uint64(n)

				if si.debugSerial {
					log.Debug().
						Int("bytesRead", n).
						Dur("readDuration", readDuration).
						Uint64("totalBytesRead", si.bytesRead).
						Msg("Serial read successful")
				}

				if si.hexDump {
					log.Debug().
						Str("hexDump", hexDump(buffer[:n], 64)).
						Msg("Raw serial data received")
				}

				// Process received bytes
				si.parser.ProcessBytes(buffer[:n])
			} else if si.debugSerial {
				log.Debug().
					Dur("readDuration", readDuration).
					Msg("Serial read returned 0 bytes")
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
	if si.port != nil {
		si.port.Close()
		si.port = nil
	}
	si.mu.Unlock()

	if si.onDisconnect != nil {
		si.onDisconnect(err)
	}
}

// attemptReconnect attempts to reconnect to the device
func (si *SerialInterface) attemptReconnect() {
	si.mu.Lock()
	defer si.mu.Unlock()

	if si.reconnectAttempts >= si.maxReconnectAttempts {
		if si.debugSerial {
			log.Error().
				Int("attempts", si.reconnectAttempts).
				Int("max", si.maxReconnectAttempts).
				Uint64("totalReconnects", si.reconnects).
				Msg("Maximum reconnection attempts reached")
		} else {
			log.Error().
				Int("attempts", si.reconnectAttempts).
				Int("max", si.maxReconnectAttempts).
				Msg("Maximum reconnection attempts reached")
		}
		return
	}

	si.reconnectAttempts++
	si.reconnects++

	if si.debugSerial {
		log.Info().
			Int("attempt", si.reconnectAttempts).
			Int("max", si.maxReconnectAttempts).
			Str("device", si.devicePath).
			Uint64("totalReconnects", si.reconnects).
			Dur("delay", si.reconnectDelay).
			Msg("Attempting to reconnect to device")
	} else {
		log.Info().
			Int("attempt", si.reconnectAttempts).
			Int("max", si.maxReconnectAttempts).
			Str("device", si.devicePath).
			Msg("Attempting to reconnect to device")
	}

	// Try to reopen the port
	serialConfig := &serial.Config{
		Name:        si.devicePath,
		Baud:        115200, // Use default baud rate
		ReadTimeout: 500 * time.Millisecond,
	}

	reconnectStart := time.Now()
	port, err := serial.OpenPort(serialConfig)
	reconnectDuration := time.Since(reconnectStart)

	if err != nil {
		if si.debugSerial {
			log.Error().Err(err).
				Int("attempt", si.reconnectAttempts).
				Dur("duration", reconnectDuration).
				Uint64("totalReconnects", si.reconnects).
				Msg("Failed to reconnect to device")
		} else {
			log.Error().Err(err).
				Int("attempt", si.reconnectAttempts).
				Msg("Failed to reconnect to device")
		}
		return
	}

	// Close old port if it exists
	if si.port != nil {
		si.port.Close()
	}

	si.port = port
	si.connected = true
	si.reconnectAttempts = 0

	if si.debugSerial {
		log.Info().
			Str("device", si.devicePath).
			Dur("duration", reconnectDuration).
			Uint64("totalReconnects", si.reconnects).
			Msg("Successfully reconnected to device")
	} else {
		log.Info().
			Str("device", si.devicePath).
			Msg("Successfully reconnected to device")
	}
}

// SendToRadio sends a ToRadio message to the device
func (si *SerialInterface) SendToRadio(toRadio *pb.ToRadio) error {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if !si.connected {
		if si.debugSerial {
			log.Error().
				Uint64("totalBytesWritten", si.bytesWritten).
				Uint64("totalWriteErrors", si.writeErrors).
				Msg("Cannot send ToRadio: device not connected")
		}
		return errors.New("device not connected")
	}

	// Build frame
	frameStart := time.Now()
	frameData, err := si.builder.BuildFrame(toRadio)
	frameDuration := time.Since(frameStart)

	if err != nil {
		if si.debugSerial {
			log.Error().
				Err(err).
				Dur("frameBuildDuration", frameDuration).
				Msg("Failed to build frame")
		}
		return errors.Wrap(err, "failed to build frame")
	}

	if si.debugSerial {
		log.Debug().
			Int("frameSize", len(frameData)).
			Dur("frameBuildDuration", frameDuration).
			Uint64("totalBytesWritten", si.bytesWritten).
			Msg("Frame built successfully")
	}

	if si.hexDump {
		log.Debug().
			Str("hexDump", hexDump(frameData, 64)).
			Msg("Frame data to be sent")
	}

	// Write to serial port
	writeStart := time.Now()
	n, err := si.port.Write(frameData)
	writeDuration := time.Since(writeStart)

	if err != nil {
		si.writeErrors++
		if si.debugSerial {
			log.Error().
				Err(err).
				Int("frameSize", len(frameData)).
				Dur("writeDuration", writeDuration).
				Uint64("totalWriteErrors", si.writeErrors).
				Uint64("totalBytesWritten", si.bytesWritten).
				Msg("Failed to write to serial port")
		}
		return errors.Wrap(err, "failed to write to serial port")
	}

	if n != len(frameData) {
		si.writeErrors++
		if si.debugSerial {
			log.Error().
				Int("bytesWritten", n).
				Int("frameSize", len(frameData)).
				Dur("writeDuration", writeDuration).
				Uint64("totalWriteErrors", si.writeErrors).
				Uint64("totalBytesWritten", si.bytesWritten).
				Msg("Partial write to serial port")
		}
		return errors.Errorf("partial write: %d of %d bytes", n, len(frameData))
	}

	si.bytesWritten += uint64(n)

	if si.debugSerial {
		log.Debug().
			Int("bytesWritten", n).
			Dur("writeDuration", writeDuration).
			Uint64("totalBytesWritten", si.bytesWritten).
			Msg("Successfully wrote frame to serial port")
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

// GetReconnectAttempts returns the current reconnection attempts
func (si *SerialInterface) GetReconnectAttempts() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.reconnectAttempts
}

// ResetReconnectAttempts resets the reconnection counter
func (si *SerialInterface) ResetReconnectAttempts() {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.reconnectAttempts = 0
}
