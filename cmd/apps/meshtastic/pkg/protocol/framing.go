package protocol

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

// stateString returns a string representation of the frame state
func (fp *FrameParser) stateString() string {
	switch fp.state {
	case StateWaitingForStart1:
		return "WaitingForStart1"
	case StateWaitingForStart2:
		return "WaitingForStart2"
	case StateWaitingForLength:
		return "WaitingForLength"
	case StateWaitingForPayload:
		return "WaitingForPayload"
	default:
		return "Unknown"
	}
}

// hexDumpData creates a hexdump-style representation of data
func (fp *FrameParser) hexDumpData(data []byte, maxBytes int) string {
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

// Protocol constants
const (
	START1           = 0x94
	START2           = 0xC3
	HEADER_LEN       = 4
	MAX_PAYLOAD_SIZE = 512
)

// FrameState represents the state of frame parsing
type FrameState int

const (
	StateWaitingForStart1 FrameState = iota
	StateWaitingForStart2
	StateWaitingForLength
	StateWaitingForPayload
)

// Frame represents a complete protocol frame
type Frame struct {
	Header  [HEADER_LEN]byte
	Payload []byte
}

// PayloadLength returns the length of the payload
func (f *Frame) PayloadLength() int {
	return (int(f.Header[2]) << 8) | int(f.Header[3])
}

// ToRadioMessage parses the frame payload as a ToRadio message
func (f *Frame) ToRadioMessage() (*pb.ToRadio, error) {
	var toRadio pb.ToRadio
	if err := proto.Unmarshal(f.Payload, &toRadio); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal ToRadio message")
	}
	return &toRadio, nil
}

// FromRadioMessage parses the frame payload as a FromRadio message
func (f *Frame) FromRadioMessage() (*pb.FromRadio, error) {
	var fromRadio pb.FromRadio
	if err := proto.Unmarshal(f.Payload, &fromRadio); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal FromRadio message")
	}
	return &fromRadio, nil
}

// FrameParser handles binary protocol frame parsing
type FrameParser struct {
	state       FrameState
	buffer      []byte
	expectLen   int
	onFrame     func(*Frame)
	onLogByte   func(byte)
	debugOutput io.Writer

	// Debug tracking
	bytesProcessed  uint64
	framesProcessed uint64
	parseErrors     uint64
	stateResets     uint64
	lastFrameTime   time.Time
	debugSerial     bool
	hexDump         bool
}

// NewFrameParser creates a new frame parser
func NewFrameParser(onFrame func(*Frame), onLogByte func(byte)) *FrameParser {
	return &FrameParser{
		state:     StateWaitingForStart1,
		buffer:    make([]byte, 0, MAX_PAYLOAD_SIZE+HEADER_LEN),
		onFrame:   onFrame,
		onLogByte: onLogByte,
	}
}

// SetDebugOutput sets the debug output writer
func (fp *FrameParser) SetDebugOutput(w io.Writer) {
	fp.debugOutput = w
}

// SetDebugSerial enables debug serial logging
func (fp *FrameParser) SetDebugSerial(enable bool) {
	fp.debugSerial = enable
}

// SetHexDump enables hex dump logging
func (fp *FrameParser) SetHexDump(enable bool) {
	fp.hexDump = enable
}

// ProcessByte processes a single byte from the serial stream
func (fp *FrameParser) ProcessByte(b byte) {
	switch fp.state {
	case StateWaitingForStart1:
		if b == START1 {
			fp.buffer = append(fp.buffer[:0], b)
			fp.state = StateWaitingForStart2
		} else {
			// This might be debug output from the device
			if fp.onLogByte != nil {
				fp.onLogByte(b)
			}
		}

	case StateWaitingForStart2:
		if b == START2 {
			fp.buffer = append(fp.buffer, b)
			fp.state = StateWaitingForLength
		} else {
			// Not a valid frame, reset
			fp.resetParser()
			if fp.onLogByte != nil {
				fp.onLogByte(b)
			}
		}

	case StateWaitingForLength:
		fp.buffer = append(fp.buffer, b)
		if len(fp.buffer) >= HEADER_LEN {
			// We have the complete header, calculate payload length
			payloadLen := (int(fp.buffer[2]) << 8) | int(fp.buffer[3])

			if payloadLen > MAX_PAYLOAD_SIZE {
				log.Warn().Int("length", payloadLen).Msg("Invalid payload length, resetting parser")
				fp.resetParser()
				return
			}

			if payloadLen < 0 {
				log.Warn().Int("length", payloadLen).Msg("Negative payload length, resetting parser")
				fp.resetParser()
				return
			}

			fp.expectLen = payloadLen
			if payloadLen == 0 {
				// Empty payload, frame is complete
				fp.handleCompleteFrame()
			} else {
				fp.state = StateWaitingForPayload
			}
		}

	case StateWaitingForPayload:
		fp.buffer = append(fp.buffer, b)
		if len(fp.buffer) >= HEADER_LEN+fp.expectLen {
			// Frame is complete
			fp.handleCompleteFrame()
		}
	}
}

// ProcessBytes processes multiple bytes from the serial stream
func (fp *FrameParser) ProcessBytes(data []byte) {
	defer func() {
		if r := recover(); r != nil {
			fp.parseErrors++
			if fp.debugSerial {
				log.Error().
					Interface("panic", r).
					Uint64("bytesProcessed", fp.bytesProcessed).
					Uint64("framesProcessed", fp.framesProcessed).
					Uint64("parseErrors", fp.parseErrors).
					Msg("Panic in frame parser, resetting state")
			} else {
				log.Error().
					Interface("panic", r).
					Msg("Panic in frame parser, resetting state")
			}
			fp.resetParser()
		}
	}()

	fp.bytesProcessed += uint64(len(data))

	if fp.debugSerial {
		log.Debug().
			Int("dataLength", len(data)).
			Uint64("totalBytesProcessed", fp.bytesProcessed).
			Str("currentState", fp.stateString()).
			Msg("Processing bytes in frame parser")
	}

	if fp.hexDump {
		log.Debug().
			Str("hexDump", fp.hexDumpData(data, 64)).
			Msg("Raw bytes to process")
	}

	for i, b := range data {
		if fp.debugSerial {
			log.Debug().
				Int("byteIndex", i).
				Int("byteValue", int(b)).
				Str("state", fp.stateString()).
				Int("bufferLen", len(fp.buffer)).
				Msg("Processing byte")
		}
		fp.ProcessByte(b)
	}
}

// resetParser resets the parser state
func (fp *FrameParser) resetParser() {
	fp.stateResets++

	if fp.debugSerial {
		log.Debug().
			Str("oldState", fp.stateString()).
			Int("bufferLen", len(fp.buffer)).
			Uint64("stateResets", fp.stateResets).
			Msg("Resetting frame parser state")
	}

	fp.state = StateWaitingForStart1
	fp.buffer = fp.buffer[:0]
	fp.expectLen = 0
}

// handleCompleteFrame handles a complete frame
func (fp *FrameParser) handleCompleteFrame() {
	fp.framesProcessed++
	frameProcessStart := time.Now()

	frame := &Frame{
		Payload: make([]byte, fp.expectLen),
	}

	// Copy header
	copy(frame.Header[:], fp.buffer[:HEADER_LEN])

	// Copy payload
	if fp.expectLen > 0 {
		copy(frame.Payload, fp.buffer[HEADER_LEN:HEADER_LEN+fp.expectLen])
	}

	if fp.debugSerial {
		timeSinceLastFrame := time.Duration(0)
		if !fp.lastFrameTime.IsZero() {
			timeSinceLastFrame = frameProcessStart.Sub(fp.lastFrameTime)
		}

		log.Debug().
			Int("payloadLen", fp.expectLen).
			Int("totalFrameLen", len(fp.buffer)).
			Uint64("framesProcessed", fp.framesProcessed).
			Dur("timeSinceLastFrame", timeSinceLastFrame).
			Msg("Processing complete frame")
	}

	if fp.debugOutput != nil {
		fmt.Fprintf(fp.debugOutput, "RX: %s\n", formatFrame(frame))
	}

	if fp.hexDump {
		log.Debug().
			Str("frameHexDump", fp.hexDumpData(fp.buffer[:HEADER_LEN+fp.expectLen], 64)).
			Msg("Complete frame hex dump")
	}

	// Call the frame handler
	if fp.onFrame != nil {
		fp.onFrame(frame)
	}

	frameProcessDuration := time.Since(frameProcessStart)
	fp.lastFrameTime = time.Now()

	if fp.debugSerial {
		log.Debug().
			Dur("frameProcessDuration", frameProcessDuration).
			Msg("Frame processing complete")
	}

	// Reset parser for next frame
	fp.resetParser()
}

// FrameBuilder helps build protocol frames
type FrameBuilder struct {
	debugOutput io.Writer
	debugSerial bool
	hexDump     bool
}

// NewFrameBuilder creates a new frame builder
func NewFrameBuilder() *FrameBuilder {
	return &FrameBuilder{}
}

// SetDebugOutput sets the debug output writer
func (fb *FrameBuilder) SetDebugOutput(w io.Writer) {
	fb.debugOutput = w
}

// SetDebugSerial enables debug serial logging
func (fb *FrameBuilder) SetDebugSerial(enable bool) {
	fb.debugSerial = enable
}

// SetHexDump enables hex dump logging
func (fb *FrameBuilder) SetHexDump(enable bool) {
	fb.hexDump = enable
}

// BuildFrame builds a frame from a ToRadio message
func (fb *FrameBuilder) BuildFrame(toRadio *pb.ToRadio) ([]byte, error) {
	// Marshal the protobuf message
	payload, err := proto.Marshal(toRadio)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal ToRadio message")
	}

	if len(payload) > MAX_PAYLOAD_SIZE {
		return nil, errors.Errorf("payload too large: %d bytes (max %d)", len(payload), MAX_PAYLOAD_SIZE)
	}

	// Build the frame
	frame := &Frame{
		Header:  [HEADER_LEN]byte{START1, START2, byte(len(payload) >> 8), byte(len(payload) & 0xFF)},
		Payload: payload,
	}

	if fb.debugOutput != nil {
		fmt.Fprintf(fb.debugOutput, "TX: %s\n", formatFrame(frame))
	}

	// Create the complete frame bytes
	result := make([]byte, HEADER_LEN+len(payload))
	copy(result[:HEADER_LEN], frame.Header[:])
	copy(result[HEADER_LEN:], payload)

	return result, nil
}

// formatFrame formats a frame for debug output
func formatFrame(frame *Frame) string {
	var buf bytes.Buffer

	// Write header
	for i, b := range frame.Header {
		if i > 0 {
			buf.WriteByte(' ')
		}
		fmt.Fprintf(&buf, "%02x", b)
	}

	// Write payload (first 32 bytes)
	if len(frame.Payload) > 0 {
		buf.WriteString(" | ")
		maxLen := len(frame.Payload)
		if maxLen > 32 {
			maxLen = 32
		}
		for i := 0; i < maxLen; i++ {
			if i > 0 {
				buf.WriteByte(' ')
			}
			fmt.Fprintf(&buf, "%02x", frame.Payload[i])
		}
		if len(frame.Payload) > 32 {
			buf.WriteString("...")
		}
	}

	return buf.String()
}

// ValidateFrame validates a frame structure
func ValidateFrame(frame *Frame) error {
	if frame.Header[0] != START1 {
		return errors.Errorf("invalid START1 byte: 0x%02x", frame.Header[0])
	}

	if frame.Header[1] != START2 {
		return errors.Errorf("invalid START2 byte: 0x%02x", frame.Header[1])
	}

	expectedLen := frame.PayloadLength()
	if len(frame.Payload) != expectedLen {
		return errors.Errorf("payload length mismatch: expected %d, got %d", expectedLen, len(frame.Payload))
	}

	if expectedLen > MAX_PAYLOAD_SIZE {
		return errors.Errorf("payload too large: %d bytes (max %d)", expectedLen, MAX_PAYLOAD_SIZE)
	}

	return nil
}
