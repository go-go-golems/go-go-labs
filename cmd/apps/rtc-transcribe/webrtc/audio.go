package webrtc

import (
	"io"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/apps/rtc-transcribe/transcribe"
	"github.com/pion/webrtc/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	// AudioSampleRate is the sample rate for audio processing
	AudioSampleRate = 48000 // WebRTC default is 48kHz
	// AudioChannels is the number of audio channels
	AudioChannels = 1 // Mono audio for speech recognition
	// OpusFrameSize is the size of an Opus frame in samples
	OpusFrameSize = 960 // 20ms at 48kHz
	// BufferDuration is the duration in seconds to buffer audio before transcription
	BufferDuration = 3 // Buffer 3 seconds of audio for transcription
)

// SetupAudioTrackHandler configures the handlers for incoming audio tracks
func SetupAudioTrackHandler(pc *webrtc.PeerConnection) {
	logger := log.With().
		Str("component", "AudioTrackHandler").
		Str("peerID", time.Now().Format("20060102-150405.000000")).
		Logger()

	logger.Info().Msg("Setting up audio track handler")

	// Improved error logging and track setup diagnostics
	logger.Debug().
		Bool("peerConnectionInitialized", pc != nil).
		Str("peerConnectionState", pc.ConnectionState().String()).
		Str("iceConnectionState", pc.ICEConnectionState().String()).
		Str("signalingState", pc.SignalingState().String()).
		Msg("Peer connection details when setting up track handler")

	// Handle incoming audio tracks
	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		trackID := track.ID()
		kind := track.Kind()
		codec := track.Codec()

		trackLogger := logger.With().
			Str("trackID", trackID).
			Str("kind", string(kind)).
			Str("codecMimeType", codec.MimeType).
			Str("codecName", codec.SDPFmtpLine).
			Int("payloadType", int(codec.PayloadType)).
			Uint32("clockRate", codec.ClockRate).
			Uint16("channels", codec.Channels).
			Logger()

		trackLogger.Info().Msg("Remote track received")

		// Only process audio tracks
		if kind != webrtc.RTPCodecTypeAudio {
			trackLogger.Warn().Msg("Ignoring non-audio track")
			return
		}

		// For Opus codec, decode the audio and process it
		if codec.MimeType == "audio/opus" {
			trackLogger.Info().Msg("Starting Opus audio processing")
			go processOpusTrack(track)
		} else {
			trackLogger.Warn().Msg("Unsupported audio codec, only Opus is supported")
		}
	})

	// Add handlers for data channel and other events
	pc.OnDataChannel(func(d *webrtc.DataChannel) {
		logger.Info().
			Str("dataChannelID", d.Label()).
			Uint16("id", *d.ID()).
			Msg("Data channel received")

		d.OnOpen(func() {
			logger.Info().
				Str("dataChannelID", d.Label()).
				Msg("Data channel opened")
		})

		d.OnClose(func() {
			logger.Info().
				Str("dataChannelID", d.Label()).
				Msg("Data channel closed")
		})

		d.OnError(func(err error) {
			logger.Error().
				Err(err).
				Str("dataChannelID", d.Label()).
				Msg("Data channel error")
		})
	})

	logger.Info().Msg("Audio track handler configured successfully")
}

// OpusDecoder interface defines the requirements for an Opus decoder
type OpusDecoder interface {
	Decode(data []byte, pcm []int16) (int, error)
	Close() error
}

// Implementation of RealOpusDecoder will be provided in separate files
// based on build tags (real_opus.go for real implementation, mock_opus.go for mock)

// Declare the interface for OpusDecoder so it can be used in audio.go
// The real implementation will be in real_opus.go (with build tags)

// NewOpusDecoder creates a new Opus decoder instance
// The real implementation is in real_opus.go, but the function is declared here
// so it can be called from the rest of the code
func NewOpusDecoder(sampleRate int, channels int) (OpusDecoder, error) {
	// We'll try to use the real implementation if it's available (based on build tags)
	// If not, we'll fall back to the mock implementation
	logger := log.With().
		Str("component", "OpusDecoderFactory").
		Int("sampleRate", sampleRate).
		Int("channels", channels).
		Logger()

	// Try to use the real implementation first
	decoder, err := newRealOpusDecoder(sampleRate, channels)
	if err != nil {
		logger.Warn().
			Err(err).
			Msg("Failed to create real Opus decoder, falling back to mock")
		return NewMockOpusDecoder(), nil
	}

	logger.Info().Msg("Successfully created real Opus decoder")
	return decoder, nil
}

// MockOpusDecoder provides a fallback implementation for development/testing
type MockOpusDecoder struct {
	logger zerolog.Logger
}

// NewMockOpusDecoder creates a new mock Opus decoder
func NewMockOpusDecoder() OpusDecoder {
	logger := log.With().
		Str("component", "MockOpusDecoder").
		Bool("mock", true).
		Logger()

	logger.Warn().Msg("Using mock Opus decoder - audio quality will be poor")

	return &MockOpusDecoder{
		logger: logger,
	}
}

// Decode simulates decoding Opus packets to PCM
func (d *MockOpusDecoder) Decode(data []byte, pcm []int16) (int, error) {
	d.logger.Debug().Int("opusBytes", len(data)).Int("pcmSize", len(pcm)).Msg("Mock-decoding Opus packet")

	// Just fill the PCM buffer with dummy data for testing
	for i := range pcm {
		if i < len(data) {
			pcm[i] = int16(data[i]) * 100 // Dummy conversion
		} else {
			pcm[i] = 0
		}
	}
	return len(pcm), nil
}

// Close is a no-op for the mock decoder
func (d *MockOpusDecoder) Close() error {
	d.logger.Debug().Msg("Closing mock Opus decoder")
	return nil
}

// processOpusTrack handles the decoding and processing of an Opus audio track
func processOpusTrack(track *webrtc.TrackRemote) {
	trackLogger := log.With().
		Str("component", "OpusProcessor").
		Str("trackID", track.ID()).
		Str("codec", track.Codec().MimeType).
		Uint32("ssrc", uint32(track.SSRC())).
		Logger()

	trackLogger.Info().Msg("Starting to process Opus track")

	// Create Opus decoder
	decoder, err := NewOpusDecoder(AudioSampleRate, AudioChannels)
	if err != nil {
		trackLogger.Error().Err(err).Msg("Failed to create Opus decoder")
		return
	}
	defer func() {
		if err := decoder.Close(); err != nil {
			trackLogger.Error().Err(err).Msg("Failed to close Opus decoder")
		}
	}()

	// Buffer for incoming Opus packets
	opusPacket := make([]byte, 1500) // MTU size buffer
	// Buffer for decoded PCM samples
	pcmBuffer := make([]int16, OpusFrameSize*AudioChannels) // 20ms frame buffer
	// Buffer to accumulate PCM data for transcription
	transcriptionBuffer := make([]int16, 0, AudioSampleRate*BufferDuration*AudioChannels)

	packetCount := 0
	bytesReceived := 0
	// lastLogTime := time.Now()
	decodeSuccessCount := 0
	decodeFailCount := 0

	trackLogger.Info().Msg("Entering RTP packet reading loop")
	for {
		// Read incoming RTP packet
		n, _, err := track.Read(opusPacket)
		if err != nil {
			if errors.Is(err, io.EOF) {
				trackLogger.Info().Msg("Opus track ended (EOF)")
				return
			}
			trackLogger.Error().Err(err).Msg("Error reading from Opus track")
			return
		}

		packetCount++
		bytesReceived += n

		// Debug log every packet read
		trackLogger.Debug().
			Int("packetNumber", packetCount).
			Int("bytesRead", n).
			Msg("Successfully read RTP packet from track")

		// Decode Opus packet to PCM
		pcmSamplesDecoded, err := decoder.Decode(opusPacket[:n], pcmBuffer)
		if err != nil {
			decodeFailCount++
			trackLogger.Error().Err(err).Int("packetSize", n).Msg("Failed to decode Opus packet")
			continue // Skip this packet
		}

		decodeSuccessCount++
		// Debug log every successful decode
		trackLogger.Debug().
			Int("packetNumber", packetCount).
			Int("pcmSamplesDecoded", pcmSamplesDecoded).
			Int("channels", AudioChannels).
			Int("expectedPcmBufferSize", len(pcmBuffer)).
			Msg("Successfully decoded Opus packet to PCM")

		// Append decoded PCM samples to the transcription buffer
		if pcmSamplesDecoded > 0 {
			transcriptionBuffer = append(transcriptionBuffer, pcmBuffer[:pcmSamplesDecoded*AudioChannels]...)
		}

		// Check if buffer has enough audio for transcription
		if len(transcriptionBuffer) >= AudioSampleRate*BufferDuration*AudioChannels {
			trackLogger.Info().
				Int("bufferSizeSamples", len(transcriptionBuffer)).
				Float64("bufferDurationSec", float64(len(transcriptionBuffer))/float64(AudioSampleRate*AudioChannels)).
				Msg("Buffer full, sending for transcription")

			// Debug log decode statistics
			trackLogger.Debug().
				Int("successfulDecodes", decodeSuccessCount).
				Int("failedDecodes", decodeFailCount).
				Msg("Decode statistics for this buffer")
			decodeSuccessCount = 0 // Reset counters for next buffer
			decodeFailCount = 0

			// Send buffer for transcription (copy to avoid race conditions)
			bufferToSend := make([]int16, len(transcriptionBuffer))
			copy(bufferToSend, transcriptionBuffer)

			go func(buffer []int16) {
				if err := transcribe.TranscribePCM(buffer); err != nil {
					log.Error().Err(err).Msg("Transcription failed")
				}
			}(bufferToSend)

			// Reset buffer (keep potential overlap? For now, just clear)
			transcriptionBuffer = transcriptionBuffer[:0]
		}
	}
}
