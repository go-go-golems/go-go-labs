package transcribe

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/apps/rtc-transcribe/sse"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// OpenAIWhisperConfig holds the configuration for the OpenAI Whisper API
type OpenAIWhisperConfig struct {
	APIKey      string
	Model       string
	Language    string
	Temperature float64
	Timeout     time.Duration
}

// DefaultOpenAIConfig returns a default configuration for the OpenAI Whisper API
func DefaultOpenAIConfig() *OpenAIWhisperConfig {
	return &OpenAIWhisperConfig{
		APIKey:      os.Getenv("OPENAI_API_KEY"),
		Model:       "whisper-1",
		Language:    "en",
		Temperature: 0.0, // Lower temperature for more accurate transcription
		Timeout:     10 * time.Second,
	}
}

// OpenAIWhisperClient is a client for the OpenAI Whisper API
type OpenAIWhisperClient struct {
	Config *OpenAIWhisperConfig
	Client *http.Client
}

// NewOpenAIWhisperClient creates a new client for the OpenAI Whisper API
func NewOpenAIWhisperClient(config *OpenAIWhisperConfig) *OpenAIWhisperClient {
	logger := log.With().
		Str("component", "WhisperAPIClient").
		Logger()

	if config == nil {
		logger.Warn().Msg("No config provided, using default configuration")
		config = DefaultOpenAIConfig()
	}

	// Ensure we have an API key
	if config.APIKey == "" {
		logger.Warn().Msg("No API key provided, trying to use OPENAI_API_KEY environment variable")
		config.APIKey = os.Getenv("OPENAI_API_KEY")
		if config.APIKey == "" {
			logger.Error().Msg("No API key available from config or environment")
		}
	}

	// Mask the API key for logging (show last 4 chars)
	var maskedKey string
	if len(config.APIKey) > 4 {
		maskedKey = "..." + config.APIKey[len(config.APIKey)-4:]
	} else if config.APIKey != "" {
		maskedKey = "***"
	} else {
		maskedKey = "MISSING"
	}

	// Create HTTP client with appropriate timeout
	client := &http.Client{
		Timeout: config.Timeout,
	}

	logger.Info().
		Str("model", config.Model).
		Str("language", config.Language).
		Float64("temperature", config.Temperature).
		Str("timeout", config.Timeout.String()).
		Str("apiKey", maskedKey).
		Msg("Created Whisper API client")

	return &OpenAIWhisperClient{
		Config: config,
		Client: client,
	}
}

// WhisperResponse is the response from the OpenAI Whisper API
type WhisperResponse struct {
	Text string `json:"text"`
}

// TranscribePCMWithAPI transcribes PCM audio using the OpenAI Whisper API
func (c *OpenAIWhisperClient) TranscribePCMWithAPI(samples []int16) error {
	// Generate a unique ID for this transcription request
	requestID := fmt.Sprintf("req-%s", time.Now().Format("20060102-150405.000000"))
	audioDuration := float64(len(samples)) / float64(48000) // Assuming 48kHz sample rate

	// Calculate audio fingerprint for logging/debugging
	audioHash := sha256.Sum256(binary.LittleEndian.AppendUint32(nil, uint32(samples[0])))
	audioFingerprint := hex.EncodeToString(audioHash[:8]) // Use first 8 bytes for brevity

	logger := log.With().
		Str("component", "WhisperAPIClient").
		Str("requestID", requestID).
		Int("sampleCount", len(samples)).
		Float64("audioDuration", audioDuration).
		Str("audioFingerprint", audioFingerprint).
		Logger()

	startTime := time.Now()
	logger.Info().Msg("Starting transcription request")

	// Validate configuration
	if c.Config.APIKey == "" {
		logger.Error().Msg("OpenAI API key is not set")
		return errors.New("OpenAI API key is not set")
	}

	// Convert PCM samples to a WAV file in memory
	logger.Debug().Msg("Converting PCM to WAV")
	conversionStartTime := time.Now()
	wavData, err := convertPCMToWAV(samples)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to convert PCM to WAV")
		return errors.Wrap(err, "failed to convert PCM to WAV")
	}
	conversionTime := time.Since(conversionStartTime)

	logger.Debug().
		Int("wavBytes", len(wavData)).
		Dur("conversionTime", conversionTime).
		Msg("Converted PCM to WAV")

	// Create a new HTTP request with the WAV file
	logger.Debug().Msg("Creating multipart request for Whisper API")
	url := "https://api.openai.com/v1/audio/transcriptions"
	body := &bytes.Buffer{}
	writer := NewMultipartWriterWithFile(body, "file", "audio.wav", wavData)
	writer.WriteField("model", c.Config.Model)
	writer.WriteField("language", c.Config.Language)
	writer.WriteField("temperature", fmt.Sprintf("%f", c.Config.Temperature))
	writer.Close()

	requestStartTime := time.Now()
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		logger.Error().
			Err(err).
			Str("url", url).
			Msg("Failed to create HTTP request")
		return errors.Wrap(err, "failed to create HTTP request")
	}

	req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType)
	req.Header.Set("X-Request-ID", requestID)

	// Send the request
	logger.Debug().
		Str("url", url).
		Str("model", c.Config.Model).
		Str("language", c.Config.Language).
		Float64("temperature", c.Config.Temperature).
		Int("contentLength", body.Len()).
		Msg("Sending request to Whisper API")

	resp, err := c.Client.Do(req)
	if err != nil {
		logger.Error().
			Err(err).
			Dur("requestAttemptDuration", time.Since(requestStartTime)).
			Msg("Failed to send HTTP request")
		return errors.Wrap(err, "failed to send HTTP request")
	}
	defer resp.Body.Close()

	requestDuration := time.Since(requestStartTime)

	// Check the response status
	logger.Debug().
		Int("statusCode", resp.StatusCode).
		Str("status", resp.Status).
		Dur("requestDuration", requestDuration).
		Msg("Received response from Whisper API")

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		logger.Error().
			Int("statusCode", resp.StatusCode).
			Str("responseBody", string(respBody)).
			Dur("requestDuration", requestDuration).
			Msg("API request failed")
		return errors.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the response
	var result WhisperResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Error().
			Err(err).
			Dur("requestDuration", requestDuration).
			Msg("Failed to decode API response")
		return errors.Wrap(err, "failed to decode API response")
	}

	// Statistics about the transcription
	totalDuration := time.Since(startTime)
	transcriptionLength := len(result.Text)
	wordsCount := len(bytes.Fields([]byte(result.Text)))

	// Send the transcription to the client
	logger.Info().
		Str("text", result.Text).
		Int("textLength", transcriptionLength).
		Int("wordCount", wordsCount).
		Dur("apiRequestDuration", requestDuration).
		Dur("totalDuration", totalDuration).
		Float64("processingRatio", totalDuration.Seconds()/audioDuration).
		Msg("Received transcription from API")

	// Send the transcription to clients via SSE
	sse.SendTranscription(result.Text)

	return nil
}

// convertPCMToWAV converts PCM audio samples to a WAV file in memory
func convertPCMToWAV(samples []int16) ([]byte, error) {
	// WAV header constants
	const (
		headerSize    = 44
		formatPCM     = 1
		numChannels   = 1
		sampleRate    = 48000
		bitsPerSample = 16
	)

	// Calculate data size and total file size
	dataSize := len(samples) * 2 // 16-bit = 2 bytes per sample
	fileSize := headerSize + dataSize

	// Create a buffer for the WAV file
	buffer := bytes.NewBuffer(make([]byte, 0, fileSize))

	// RIFF header
	buffer.WriteString("RIFF")
	binary.Write(buffer, binary.LittleEndian, uint32(fileSize-8))
	buffer.WriteString("WAVE")

	// Format chunk
	buffer.WriteString("fmt ")
	binary.Write(buffer, binary.LittleEndian, uint32(16)) // Size of format chunk
	binary.Write(buffer, binary.LittleEndian, uint16(formatPCM))
	binary.Write(buffer, binary.LittleEndian, uint16(numChannels))
	binary.Write(buffer, binary.LittleEndian, uint32(sampleRate))
	binary.Write(buffer, binary.LittleEndian, uint32(sampleRate*numChannels*bitsPerSample/8)) // Bytes per second
	binary.Write(buffer, binary.LittleEndian, uint16(numChannels*bitsPerSample/8))            // Block align
	binary.Write(buffer, binary.LittleEndian, uint16(bitsPerSample))

	// Data chunk
	buffer.WriteString("data")
	binary.Write(buffer, binary.LittleEndian, uint32(dataSize))

	// Write PCM samples
	for _, sample := range samples {
		binary.Write(buffer, binary.LittleEndian, sample)
	}

	return buffer.Bytes(), nil
}

// MultipartWriter is a helper to create multipart form data
type MultipartWriter struct {
	*bytes.Buffer
	FormDataContentType string
}

// NewMultipartWriterWithFile creates a new multipart writer with a file
func NewMultipartWriterWithFile(buffer *bytes.Buffer, fieldName, fileName string, fileData []byte) *MultipartWriter {
	boundary := fmt.Sprintf("------------------------%d", time.Now().UnixNano())
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)

	buffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buffer.WriteString(fmt.Sprintf(`Content-Disposition: form-data; name="%s"; filename="%s"`+"\r\n", fieldName, fileName))
	buffer.WriteString("Content-Type: audio/wav\r\n\r\n")
	buffer.Write(fileData)
	buffer.WriteString("\r\n")

	return &MultipartWriter{
		Buffer:              buffer,
		FormDataContentType: contentType,
	}
}

// WriteField adds a form field to the multipart writer
func (m *MultipartWriter) WriteField(fieldName, value string) {
	boundary := m.FormDataContentType[len("multipart/form-data; boundary="):]
	m.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	m.WriteString(fmt.Sprintf(`Content-Disposition: form-data; name="%s"`+"\r\n\r\n", fieldName))
	m.WriteString(value)
	m.WriteString("\r\n")
}

// Close finishes the multipart form
func (m *MultipartWriter) Close() {
	boundary := m.FormDataContentType[len("multipart/form-data; boundary="):]
	m.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
}
