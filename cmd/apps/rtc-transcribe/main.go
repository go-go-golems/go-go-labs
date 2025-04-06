package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/apps/rtc-transcribe/sse"
	"github.com/go-go-golems/go-go-labs/cmd/apps/rtc-transcribe/transcribe"
	webrtc_handlers "github.com/go-go-golems/go-go-labs/cmd/apps/rtc-transcribe/webrtc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// Application constants
const (
	DefaultPort     = "8080"
	ShutdownTimeout = 5 * time.Second
)

// Command line flags
var (
	port              string
	logLevel          string
	apiKey            string
	transcriptionMode string
	useIceServers     bool
)

func main() {
	// Configure root command
	rootCmd := &cobra.Command{
		Use:   "rtc-transcribe",
		Short: "A real-time audio transcription server using WebRTC and OpenAI Whisper",
		Run:   run,
	}

	// Define flags
	rootCmd.Flags().StringVarP(&port, "port", "p", DefaultPort, "The HTTP server port")
	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "OpenAI API key (defaults to OPENAI_API_KEY env var)")
	rootCmd.Flags().StringVarP(&transcriptionMode, "mode", "m", "api", "Transcription mode (api)")
	rootCmd.Flags().BoolVar(&useIceServers, "use-ice-servers", false, "Use STUN/TURN servers for WebRTC (not needed for localhost)")

	// Add examples in usage template
	rootCmd.Example = `  # Run with default settings
  rtc-transcribe
  
  # Run with debug logging
  rtc-transcribe --log-level debug
  
  # Run on a different port
  rtc-transcribe --port 9000
  
  # Specify OpenAI API key
  rtc-transcribe --api-key sk-xxxxxxxxxxxxx
`

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

// run is the main function that starts the server
func run(cmd *cobra.Command, args []string) {
	// Configure logging
	configureLogging(logLevel)

	// Set the UseIceServers flag in the webrtc package
	webrtc_handlers.UseIceServers = useIceServers
	log.Info().Bool("useIceServers", useIceServers).Msg("WebRTC ICE servers configuration set")

	// Initialize the transcription service
	if err := initializeTranscription(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize transcription service")
	}

	// Create the HTTP server
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Define routes
	mux.HandleFunc("/offer", webrtc_handlers.HandleOffer)
	mux.HandleFunc("/transcribe", sse.HandleSSE)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"status":"ok","useIceServers":%t}`, useIceServers)))
	})
	mux.Handle("/", http.FileServer(http.Dir(getStaticDir())))

	// Start the server in a goroutine
	go func() {
		log.Info().Str("port", port).Msg("Starting server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}

// configureLogging sets up the logging configuration
func configureLogging(level string) {
	// Set a default log level
	logLevel := zerolog.InfoLevel

	// Parse the log level from the command line
	if parsedLevel, err := zerolog.ParseLevel(level); err == nil {
		logLevel = parsedLevel
	}

	// Configure zerolog
	zerolog.SetGlobalLevel(logLevel)
	// Enable caller information in logs
	log.Logger = log.Logger.With().Caller().Timestamp().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	log.Info().Str("level", logLevel.String()).Msg("Log level set")
}

// initializeTranscription sets up the transcription service
func initializeTranscription() error {
	// Determine the transcription mode
	var mode transcribe.TranscriptionMode
	switch transcriptionMode {
	case "api":
		mode = transcribe.APIMode
	default:
		return errors.Errorf("unsupported transcription mode: %s", transcriptionMode)
	}

	// Get the API key
	apiKeyToUse := apiKey
	if apiKeyToUse == "" {
		apiKeyToUse = os.Getenv("OPENAI_API_KEY")
	}

	// Create the configuration
	config := &transcribe.OpenAIWhisperConfig{
		APIKey:      apiKeyToUse,
		Model:       "whisper-1",
		Language:    "en",
		Temperature: 0.0,
		Timeout:     10 * time.Second,
	}

	// Initialize the transcription service
	return transcribe.Initialize(mode, config)
}

// getStaticDir returns the path to the static directory
func getStaticDir() string {
	// In development, look for the static directory in the current directory
	_, err := os.Stat("static")
	if err == nil {
		return "static"
	}

	// In production, look for the static directory relative to the executable
	exePath, err := os.Executable()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get executable path")
		return "static" // Fallback
	}

	staticPath := filepath.Join(filepath.Dir(exePath), "static")
	if _, err := os.Stat(staticPath); err == nil {
		return staticPath
	}

	// If all else fails, use the repository-relative path
	return filepath.Join("cmd", "apps", "rtc-transcribe", "static")
}
