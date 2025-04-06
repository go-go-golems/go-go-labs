package transcribe

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// TranscriptionMode represents the mode of transcription (local or API)
type TranscriptionMode string

const (
	// LocalMode uses the local Whisper model
	LocalMode TranscriptionMode = "local"
	// APIMode uses the OpenAI Whisper API
	APIMode TranscriptionMode = "api"
)

var (
	// currentMode is the current transcription mode
	currentMode = APIMode

	// whisperAPIClient is the client for the OpenAI Whisper API
	whisperAPIClient *OpenAIWhisperClient
)

// Initialize initializes the transcription service
func Initialize(mode TranscriptionMode, config *OpenAIWhisperConfig) error {
	logger := log.With().
		Str("component", "TranscriptionService").
		Str("mode", string(mode)).
		Logger()
	
	logger.Info().Msg("Initializing transcription service")
	
	currentMode = mode

	switch mode {
	case APIMode:
		if config == nil {
			logger.Warn().Msg("No config provided, using default configuration")
			config = DefaultOpenAIConfig()
		}
		
		logger.Info().
			Str("model", config.Model).
			Str("language", config.Language).
			Float64("temperature", config.Temperature).
			Str("timeout", config.Timeout.String()).
			Bool("hasApiKey", config.APIKey != "").
			Msg("Creating Whisper API client")
			
		whisperAPIClient = NewOpenAIWhisperClient(config)
		logger.Info().Msg("Successfully initialized API transcription mode")
		return nil
		
	case LocalMode:
		logger.Error().Msg("Local transcription mode is not implemented")
		return errors.New("local mode is not implemented")
		
	default:
		logger.Error().Str("mode", string(mode)).Msg("Unknown transcription mode")
		return errors.Errorf("unknown transcription mode: %s", mode)
	}
}

// TranscribePCM transcribes PCM audio data
func TranscribePCM(samples []int16) error {
	logger := log.With().
		Str("component", "TranscriptionService").
		Str("mode", string(currentMode)).
		Int("sampleCount", len(samples)).
		Logger()
	
	logger.Debug().Msg("Transcribing PCM audio data")
	
	switch currentMode {
	case APIMode:
		if whisperAPIClient == nil {
			logger.Error().Msg("Whisper API client is not initialized")
			return errors.New("whisper API client is not initialized")
		}
		
		logger.Debug().Msg("Using API mode for transcription")
		return whisperAPIClient.TranscribePCMWithAPI(samples)
		
	case LocalMode:
		logger.Error().Msg("Local transcription mode is not implemented")
		return errors.New("local mode is not implemented")
		
	default:
		logger.Error().Str("mode", string(currentMode)).Msg("Unknown transcription mode")
		return errors.Errorf("unknown transcription mode: %s", currentMode)
	}
}