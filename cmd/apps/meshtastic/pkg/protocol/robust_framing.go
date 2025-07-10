package protocol

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
)

// RobustFrameParser implements a robust frame parser with better error handling
type RobustFrameParser struct {
	reader    *bufio.Reader
	state     FrameState
	buffer    *bytes.Buffer
	expectLen int
	onFrame   func(*Frame)
	onLogByte func(byte)
	onError   func(error)

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Statistics
	mu              sync.RWMutex
	bytesProcessed  uint64
	framesProcessed uint64
	parseErrors     uint64
	stateResets     uint64
	lastFrameTime   time.Time

	// Configuration
	debugSerial  bool
	hexDump      bool
	maxFrameSize int
	readTimeout  time.Duration

	// Error recovery
	consecutiveErrors    int
	maxConsecutiveErrors int
	errorBackoff         time.Duration
	lastErrorTime        time.Time

	// Buffer management
	maxBufferSize int
	bufferPool    *sync.Pool
}

// NewRobustFrameParser creates a new robust frame parser
func NewRobustFrameParser(reader io.Reader, onFrame func(*Frame), onLogByte func(byte)) *RobustFrameParser {
	ctx, cancel := context.WithCancel(context.Background())

	rfp := &RobustFrameParser{
		reader:               bufio.NewReaderSize(reader, 4096),
		state:                StateWaitingForStart1,
		buffer:               bytes.NewBuffer(make([]byte, 0, MAX_PAYLOAD_SIZE+HEADER_LEN)),
		onFrame:              onFrame,
		onLogByte:            onLogByte,
		ctx:                  ctx,
		cancel:               cancel,
		maxFrameSize:         MAX_PAYLOAD_SIZE + HEADER_LEN,
		readTimeout:          1 * time.Second,
		maxConsecutiveErrors: 10,
		errorBackoff:         100 * time.Millisecond,
		maxBufferSize:        8192,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 1024)
			},
		},
	}

	return rfp
}

// SetContext sets the context for cancellation
func (rfp *RobustFrameParser) SetContext(ctx context.Context) {
	rfp.cancel() // Cancel old context
	rfp.ctx = ctx
}

// SetOnError sets the error handler
func (rfp *RobustFrameParser) SetOnError(onError func(error)) {
	rfp.onError = onError
}

// SetDebugSerial enables debug serial logging
func (rfp *RobustFrameParser) SetDebugSerial(enable bool) {
	rfp.debugSerial = enable
}

// SetHexDump enables hex dump logging
func (rfp *RobustFrameParser) SetHexDump(enable bool) {
	rfp.hexDump = enable
}

// SetReadTimeout sets the read timeout
func (rfp *RobustFrameParser) SetReadTimeout(timeout time.Duration) {
	rfp.readTimeout = timeout
}

// Start starts the frame parsing loop
func (rfp *RobustFrameParser) Start() error {
	log.Debug().Msg("Starting robust frame parser")

	for {
		select {
		case <-rfp.ctx.Done():
			log.Debug().Msg("Frame parser cancelled")
			return rfp.ctx.Err()
		default:
			if err := rfp.processNextFrame(); err != nil {
				if err == io.EOF {
					log.Debug().Msg("EOF reached, stopping frame parser")
					return err
				}

				// Handle recoverable errors
				if rfp.handleError(err) {
					continue
				}

				return err
			}
		}
	}
}

// processNextFrame processes the next frame from the stream
func (rfp *RobustFrameParser) processNextFrame() error {
	// Read byte from buffered reader
	b, err := rfp.reader.ReadByte()
	if err != nil {
		return err
	}

	rfp.mu.Lock()
	rfp.bytesProcessed++
	rfp.mu.Unlock()

	if rfp.debugSerial {
		log.Debug().
			Int("byte", int(b)).
			Str("state", rfp.stateString()).
			Msg("Processing byte")
	}

	return rfp.processByte(b)
}

// processByte processes a single byte
func (rfp *RobustFrameParser) processByte(b byte) error {
	switch rfp.state {
	case StateWaitingForStart1:
		if b == START1 {
			rfp.buffer.Reset()
			rfp.buffer.WriteByte(b)
			rfp.state = StateWaitingForStart2
		} else {
			// This might be debug output
			if rfp.onLogByte != nil {
				rfp.onLogByte(b)
			}
		}

	case StateWaitingForStart2:
		if b == START2 {
			rfp.buffer.WriteByte(b)
			rfp.state = StateWaitingForLength
		} else {
			// Invalid sequence, reset
			rfp.resetParser()
			if rfp.onLogByte != nil {
				rfp.onLogByte(b)
			}
		}

	case StateWaitingForLength:
		rfp.buffer.WriteByte(b)
		if rfp.buffer.Len() >= HEADER_LEN {
			// Calculate payload length
			bufferBytes := rfp.buffer.Bytes()
			payloadLen := (int(bufferBytes[2]) << 8) | int(bufferBytes[3])

			if payloadLen < 0 || payloadLen > MAX_PAYLOAD_SIZE {
				log.Warn().Int("length", payloadLen).Msg("Invalid payload length")
				rfp.resetParser()
				return nil
			}

			rfp.expectLen = payloadLen
			if payloadLen == 0 {
				// Empty payload, frame complete
				return rfp.handleCompleteFrame()
			} else {
				rfp.state = StateWaitingForPayload
			}
		}

	case StateWaitingForPayload:
		rfp.buffer.WriteByte(b)
		if rfp.buffer.Len() >= HEADER_LEN+rfp.expectLen {
			// Frame complete
			return rfp.handleCompleteFrame()
		}

		// Check for buffer overflow
		if rfp.buffer.Len() > rfp.maxBufferSize {
			log.Warn().Int("bufferSize", rfp.buffer.Len()).Msg("Buffer overflow, resetting parser")
			rfp.resetParser()
			return nil
		}
	}

	return nil
}

// handleCompleteFrame handles a complete frame
func (rfp *RobustFrameParser) handleCompleteFrame() error {
	rfp.mu.Lock()
	rfp.framesProcessed++
	frameProcessStart := time.Now()
	timeSinceLastFrame := time.Duration(0)
	if !rfp.lastFrameTime.IsZero() {
		timeSinceLastFrame = frameProcessStart.Sub(rfp.lastFrameTime)
	}
	rfp.lastFrameTime = frameProcessStart
	rfp.mu.Unlock()

	if rfp.debugSerial {
		log.Debug().
			Int("payloadLen", rfp.expectLen).
			Int("totalFrameLen", rfp.buffer.Len()).
			Uint64("framesProcessed", rfp.framesProcessed).
			Dur("timeSinceLastFrame", timeSinceLastFrame).
			Msg("Processing complete frame")
	}

	// Create frame from buffer
	frame := &Frame{
		Payload: make([]byte, rfp.expectLen),
	}

	bufferBytes := rfp.buffer.Bytes()
	copy(frame.Header[:], bufferBytes[:HEADER_LEN])

	if rfp.expectLen > 0 {
		copy(frame.Payload, bufferBytes[HEADER_LEN:HEADER_LEN+rfp.expectLen])
	}

	if rfp.hexDump {
		log.Debug().
			Str("frameHex", rfp.hexDumpData(bufferBytes, 64)).
			Msg("Complete frame received")
	}

	// Validate frame
	if err := ValidateFrame(frame); err != nil {
		log.Warn().Err(err).Msg("Invalid frame received")
		rfp.resetParser()
		return nil
	}

	// Call frame handler
	if rfp.onFrame != nil {
		rfp.onFrame(frame)
	}

	// Reset consecutive errors on successful frame
	rfp.consecutiveErrors = 0

	// Reset for next frame
	rfp.resetParser()

	return nil
}

// resetParser resets the parser state
func (rfp *RobustFrameParser) resetParser() {
	rfp.mu.Lock()
	rfp.stateResets++
	rfp.mu.Unlock()

	if rfp.debugSerial {
		log.Debug().
			Str("oldState", rfp.stateString()).
			Int("bufferLen", rfp.buffer.Len()).
			Uint64("stateResets", rfp.stateResets).
			Msg("Resetting parser state")
	}

	rfp.state = StateWaitingForStart1
	rfp.buffer.Reset()
	rfp.expectLen = 0
}

// handleError handles errors with recovery logic
func (rfp *RobustFrameParser) handleError(err error) bool {
	rfp.mu.Lock()
	rfp.parseErrors++
	rfp.consecutiveErrors++
	rfp.lastErrorTime = time.Now()
	rfp.mu.Unlock()

	if rfp.debugSerial {
		log.Error().
			Err(err).
			Int("consecutiveErrors", rfp.consecutiveErrors).
			Uint64("totalErrors", rfp.parseErrors).
			Msg("Frame parser error")
	}

	// Call error handler
	if rfp.onError != nil {
		rfp.onError(err)
	}

	// Check if we should continue trying
	if rfp.consecutiveErrors >= rfp.maxConsecutiveErrors {
		log.Error().
			Int("consecutiveErrors", rfp.consecutiveErrors).
			Int("maxErrors", rfp.maxConsecutiveErrors).
			Msg("Too many consecutive errors, stopping parser")
		return false
	}

	// Reset parser state on error
	rfp.resetParser()

	// Backoff before continuing
	if rfp.errorBackoff > 0 {
		time.Sleep(rfp.errorBackoff)
	}

	return true
}

// stateString returns a string representation of the current state
func (rfp *RobustFrameParser) stateString() string {
	switch rfp.state {
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

// hexDumpData creates a hex dump of data
func (rfp *RobustFrameParser) hexDumpData(data []byte, maxBytes int) string {
	if len(data) == 0 {
		return ""
	}

	displayData := data
	if len(displayData) > maxBytes {
		displayData = data[:maxBytes]
	}

	var buf bytes.Buffer
	for i, b := range displayData {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(fmt.Sprintf("%02x", b))
	}

	if len(data) > maxBytes {
		buf.WriteString("...")
	}

	return buf.String()
}

// GetStatistics returns parser statistics
func (rfp *RobustFrameParser) GetStatistics() FrameParserStatistics {
	rfp.mu.RLock()
	defer rfp.mu.RUnlock()

	return FrameParserStatistics{
		BytesProcessed:    rfp.bytesProcessed,
		FramesProcessed:   rfp.framesProcessed,
		ParseErrors:       rfp.parseErrors,
		StateResets:       rfp.stateResets,
		ConsecutiveErrors: rfp.consecutiveErrors,
		LastFrameTime:     rfp.lastFrameTime,
		LastErrorTime:     rfp.lastErrorTime,
		CurrentState:      rfp.state,
	}
}

// Close closes the parser
func (rfp *RobustFrameParser) Close() error {
	log.Debug().Msg("Closing robust frame parser")
	rfp.cancel()
	return nil
}

// FrameParserStatistics holds parser statistics
type FrameParserStatistics struct {
	BytesProcessed    uint64
	FramesProcessed   uint64
	ParseErrors       uint64
	StateResets       uint64
	ConsecutiveErrors int
	LastFrameTime     time.Time
	LastErrorTime     time.Time
	CurrentState      FrameState
}

// RobustFrameBuilder implements a robust frame builder
type RobustFrameBuilder struct {
	mu          sync.RWMutex
	debugSerial bool
	hexDump     bool

	// Statistics
	framesSent   uint64
	bytesWritten uint64
	buildErrors  uint64

	// Buffer pool for efficiency
	bufferPool *sync.Pool
}

// NewRobustFrameBuilder creates a new robust frame builder
func NewRobustFrameBuilder() *RobustFrameBuilder {
	return &RobustFrameBuilder{
		bufferPool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, MAX_PAYLOAD_SIZE+HEADER_LEN))
			},
		},
	}
}

// SetDebugSerial enables debug serial logging
func (rfb *RobustFrameBuilder) SetDebugSerial(enable bool) {
	rfb.debugSerial = enable
}

// SetHexDump enables hex dump logging
func (rfb *RobustFrameBuilder) SetHexDump(enable bool) {
	rfb.hexDump = enable
}

// BuildFrame builds a frame from a ToRadio message
func (rfb *RobustFrameBuilder) BuildFrame(toRadio *pb.ToRadio) ([]byte, error) {
	// Marshal the protobuf message
	payload, err := proto.Marshal(toRadio)
	if err != nil {
		rfb.mu.Lock()
		rfb.buildErrors++
		rfb.mu.Unlock()
		return nil, errors.Wrap(err, "failed to marshal ToRadio message")
	}

	if len(payload) > MAX_PAYLOAD_SIZE {
		rfb.mu.Lock()
		rfb.buildErrors++
		rfb.mu.Unlock()
		return nil, errors.Errorf("payload too large: %d bytes (max %d)", len(payload), MAX_PAYLOAD_SIZE)
	}

	// Get buffer from pool
	buffer := rfb.bufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer rfb.bufferPool.Put(buffer)

	// Build frame
	frame := &Frame{
		Header:  [HEADER_LEN]byte{START1, START2, byte(len(payload) >> 8), byte(len(payload) & 0xFF)},
		Payload: payload,
	}

	// Write header
	buffer.Write(frame.Header[:])

	// Write payload
	if len(payload) > 0 {
		buffer.Write(payload)
	}

	// Update statistics
	rfb.mu.Lock()
	rfb.framesSent++
	rfb.bytesWritten += uint64(buffer.Len())
	rfb.mu.Unlock()

	if rfb.debugSerial {
		log.Debug().
			Int("frameSize", buffer.Len()).
			Int("payloadSize", len(payload)).
			Uint64("framesSent", rfb.framesSent).
			Msg("Frame built")
	}

	if rfb.hexDump {
		log.Debug().
			Str("frameHex", rfb.hexDumpData(buffer.Bytes(), 64)).
			Msg("Frame data")
	}

	// Return copy of buffer data
	result := make([]byte, buffer.Len())
	copy(result, buffer.Bytes())

	return result, nil
}

// GetStatistics returns builder statistics
func (rfb *RobustFrameBuilder) GetStatistics() FrameBuilderStatistics {
	rfb.mu.RLock()
	defer rfb.mu.RUnlock()

	return FrameBuilderStatistics{
		FramesSent:   rfb.framesSent,
		BytesWritten: rfb.bytesWritten,
		BuildErrors:  rfb.buildErrors,
	}
}

// hexDumpData creates a hex dump of data
func (rfb *RobustFrameBuilder) hexDumpData(data []byte, maxBytes int) string {
	if len(data) == 0 {
		return ""
	}

	displayData := data
	if len(displayData) > maxBytes {
		displayData = data[:maxBytes]
	}

	var buf bytes.Buffer
	for i, b := range displayData {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(fmt.Sprintf("%02x", b))
	}

	if len(data) > maxBytes {
		buf.WriteString("...")
	}

	return buf.String()
}

// FrameBuilderStatistics holds builder statistics
type FrameBuilderStatistics struct {
	FramesSent   uint64
	BytesWritten uint64
	BuildErrors  uint64
}

// StreamFrameProcessor combines robust parsing and building
type StreamFrameProcessor struct {
	parser  *RobustFrameParser
	builder *RobustFrameBuilder

	// Stream handling
	reader io.Reader
	writer io.Writer

	// Context
	ctx    context.Context
	cancel context.CancelFunc

	// Configuration
	debugSerial bool
	hexDump     bool
}

// NewStreamFrameProcessor creates a new stream frame processor
func NewStreamFrameProcessor(reader io.Reader, writer io.Writer, onFrame func(*Frame), onLogByte func(byte)) *StreamFrameProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	sfp := &StreamFrameProcessor{
		reader:  reader,
		writer:  writer,
		ctx:     ctx,
		cancel:  cancel,
		parser:  NewRobustFrameParser(reader, onFrame, onLogByte),
		builder: NewRobustFrameBuilder(),
	}

	sfp.parser.SetContext(ctx)

	return sfp
}

// SetDebugSerial enables debug serial logging
func (sfp *StreamFrameProcessor) SetDebugSerial(enable bool) {
	sfp.debugSerial = enable
	sfp.parser.SetDebugSerial(enable)
	sfp.builder.SetDebugSerial(enable)
}

// SetHexDump enables hex dump logging
func (sfp *StreamFrameProcessor) SetHexDump(enable bool) {
	sfp.hexDump = enable
	sfp.parser.SetHexDump(enable)
	sfp.builder.SetHexDump(enable)
}

// SetOnError sets the error handler
func (sfp *StreamFrameProcessor) SetOnError(onError func(error)) {
	sfp.parser.SetOnError(onError)
}

// Start starts the frame processor
func (sfp *StreamFrameProcessor) Start() error {
	log.Debug().Msg("Starting stream frame processor")
	return sfp.parser.Start()
}

// SendFrame sends a frame to the stream
func (sfp *StreamFrameProcessor) SendFrame(toRadio *pb.ToRadio) error {
	frameData, err := sfp.builder.BuildFrame(toRadio)
	if err != nil {
		return errors.Wrap(err, "failed to build frame")
	}

	_, err = sfp.writer.Write(frameData)
	if err != nil {
		return errors.Wrap(err, "failed to write frame")
	}

	return nil
}

// GetStatistics returns combined statistics
func (sfp *StreamFrameProcessor) GetStatistics() (FrameParserStatistics, FrameBuilderStatistics) {
	return sfp.parser.GetStatistics(), sfp.builder.GetStatistics()
}

// Close closes the processor
func (sfp *StreamFrameProcessor) Close() error {
	log.Debug().Msg("Closing stream frame processor")
	sfp.cancel()
	return sfp.parser.Close()
}
