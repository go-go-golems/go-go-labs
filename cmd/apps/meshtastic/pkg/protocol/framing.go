package protocol

import (
	"bytes"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

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
	for _, b := range data {
		fp.ProcessByte(b)
	}
}

// resetParser resets the parser state
func (fp *FrameParser) resetParser() {
	fp.state = StateWaitingForStart1
	fp.buffer = fp.buffer[:0]
	fp.expectLen = 0
}

// handleCompleteFrame handles a complete frame
func (fp *FrameParser) handleCompleteFrame() {
	frame := &Frame{
		Payload: make([]byte, fp.expectLen),
	}

	// Copy header
	copy(frame.Header[:], fp.buffer[:HEADER_LEN])

	// Copy payload
	if fp.expectLen > 0 {
		copy(frame.Payload, fp.buffer[HEADER_LEN:HEADER_LEN+fp.expectLen])
	}

	if fp.debugOutput != nil {
		fmt.Fprintf(fp.debugOutput, "RX: %s\n", formatFrame(frame))
	}

	// Call the frame handler
	if fp.onFrame != nil {
		fp.onFrame(frame)
	}

	// Reset parser for next frame
	fp.resetParser()
}

// FrameBuilder helps build protocol frames
type FrameBuilder struct {
	debugOutput io.Writer
}

// NewFrameBuilder creates a new frame builder
func NewFrameBuilder() *FrameBuilder {
	return &FrameBuilder{}
}

// SetDebugOutput sets the debug output writer
func (fb *FrameBuilder) SetDebugOutput(w io.Writer) {
	fb.debugOutput = w
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
