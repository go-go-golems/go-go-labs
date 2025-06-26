package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/cmd/apps/split-video/pkg/config"
	"github.com/go-go-golems/go-go-labs/cmd/apps/split-video/pkg/tui"
	"github.com/go-go-golems/go-go-labs/cmd/apps/split-video/pkg/video"
)

var (
	cfg = &config.Config{}
)

func main() {
	// Setup zerolog
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Set log level based on configuration
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		switch cfg.LogLevel {
		case "debug":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "info":
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case "warn":
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case "error":
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		default:
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		}
		
		if cfg.Verbose {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

var rootCmd = &cobra.Command{
	Use:   "split-video [video-file]",
	Short: "A powerful video splitting tool with TUI interface",
	Long: `split-video is a command-line tool that allows you to split video files
in various ways with an interactive terminal user interface.

If a video file is provided as an argument, the TUI will launch with that file pre-loaded.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Set input file if provided as argument
		if len(args) > 0 {
			cfg.InputFile = args[0]
		}
		
		// Launch TUI
		app := tui.NewApp(cfg)
		if err := app.Run(); err != nil {
			log.Fatal().Err(err).Msg("TUI application failed")
		}
	},
}

var equalCmd = &cobra.Command{
	Use:   "equal [video-file]",
	Short: "Split video into equal segments",
	Long:  `Split a video file into equal-length segments with optional overlap.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFile := args[0]
		
		segments, _ := cmd.Flags().GetInt("segments")
		overlap, _ := cmd.Flags().GetDuration("overlap")
		outputDir, _ := cmd.Flags().GetString("output")
		extractAudio, _ := cmd.Flags().GetBool("extract-audio")
		audioFormat, _ := cmd.Flags().GetString("audio-format")
		
		cfg := &config.Config{
			InputFile:    inputFile,
			Segments:     segments,
			Overlap:      overlap,
			OutputDir:    outputDir,
			ExtractAudio: extractAudio,
			AudioFormat:  audioFormat,
		}
		
		log.Info().
			Str("input", inputFile).
			Int("segments", segments).
			Dur("overlap", overlap).
			Msg("Starting equal split")
		
		if err := video.SplitEqual(cfg); err != nil {
			log.Fatal().Err(err).Msg("Failed to split video")
		}
		
		log.Info().Msg("Video split completed successfully")
	},
}

var timeCmd = &cobra.Command{
	Use:   "time [video-file]",
	Short: "Split video at specific time intervals",
	Long:  `Split a video file at specified time intervals.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFile := args[0]
		
		intervals, _ := cmd.Flags().GetStringSlice("intervals")
		outputDir, _ := cmd.Flags().GetString("output")
		extractAudio, _ := cmd.Flags().GetBool("extract-audio")
		audioFormat, _ := cmd.Flags().GetString("audio-format")
		
		cfg := &config.Config{
			InputFile:    inputFile,
			Intervals:    intervals,
			OutputDir:    outputDir,
			ExtractAudio: extractAudio,
			AudioFormat:  audioFormat,
		}
		
		log.Info().
			Str("input", inputFile).
			Strs("intervals", intervals).
			Msg("Starting time-based split")
		
		if err := video.SplitByTime(cfg); err != nil {
			log.Fatal().Err(err).Msg("Failed to split video")
		}
		
		log.Info().Msg("Video split completed successfully")
	},
}

var durationCmd = &cobra.Command{
	Use:   "duration [video-file]",
	Short: "Split video into segments of specific duration",
	Long:  `Split a video file into segments of a specified duration.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFile := args[0]
		
		segmentDuration, _ := cmd.Flags().GetDuration("duration")
		overlap, _ := cmd.Flags().GetDuration("overlap")
		outputDir, _ := cmd.Flags().GetString("output")
		extractAudio, _ := cmd.Flags().GetBool("extract-audio")
		audioFormat, _ := cmd.Flags().GetString("audio-format")
		
		cfg := &config.Config{
			InputFile:       inputFile,
			SegmentDuration: segmentDuration,
			Overlap:         overlap,
			OutputDir:       outputDir,
			ExtractAudio:    extractAudio,
			AudioFormat:     audioFormat,
		}
		
		log.Info().
			Str("input", inputFile).
			Dur("segment_duration", segmentDuration).
			Dur("overlap", overlap).
			Msg("Starting duration-based split")
		
		if err := video.SplitByDuration(cfg); err != nil {
			log.Fatal().Err(err).Msg("Failed to split video")
		}
		
		log.Info().Msg("Video split completed successfully")
	},
}

var audioCmd = &cobra.Command{
	Use:   "audio [video-file]",
	Short: "Extract audio from video file",
	Long:  `Extract audio track from a video file in various formats.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFile := args[0]
		
		outputFile, _ := cmd.Flags().GetString("output")
		audioFormat, _ := cmd.Flags().GetString("format")
		
		if outputFile == "" {
			outputFile = fmt.Sprintf("%s.%s", inputFile[:len(inputFile)-4], audioFormat)
		}
		
		log.Info().
			Str("input", inputFile).
			Str("output", outputFile).
			Str("format", audioFormat).
			Msg("Extracting audio")
		
		if err := video.ExtractAudio(inputFile, outputFile, audioFormat); err != nil {
			log.Fatal().Err(err).Msg("Failed to extract audio")
		}
		
		log.Info().Msg("Audio extraction completed successfully")
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", "warn", "Log level (debug, info, warn, error)")
	
	// Equal split command flags
	equalCmd.Flags().IntP("segments", "s", 5, "Number of segments to create")
	equalCmd.Flags().DurationP("overlap", "o", 0, "Overlap duration between segments (e.g., 5m, 30s)")
	equalCmd.Flags().StringP("output", "d", ".", "Output directory")
	equalCmd.Flags().BoolP("extract-audio", "a", false, "Also extract audio from each segment")
	equalCmd.Flags().StringP("audio-format", "f", "mp3", "Audio format (mp3, wav, aac, flac)")
	
	// Time-based split command flags
	timeCmd.Flags().StringSliceP("intervals", "i", []string{}, "Time intervals to split at (e.g., 10m,20m,30m)")
	timeCmd.Flags().StringP("output", "d", ".", "Output directory")
	timeCmd.Flags().BoolP("extract-audio", "a", false, "Also extract audio from each segment")
	timeCmd.Flags().StringP("audio-format", "f", "mp3", "Audio format (mp3, wav, aac, flac)")
	
	// Duration-based split command flags
	durationCmd.Flags().DurationP("duration", "t", 10*time.Minute, "Duration of each segment")
	durationCmd.Flags().DurationP("overlap", "o", 0, "Overlap duration between segments")
	durationCmd.Flags().StringP("output", "d", ".", "Output directory")
	durationCmd.Flags().BoolP("extract-audio", "a", false, "Also extract audio from each segment")
	durationCmd.Flags().StringP("audio-format", "f", "mp3", "Audio format (mp3, wav, aac, flac)")
	
	// Audio extraction command flags
	audioCmd.Flags().StringP("output", "o", "", "Output audio file (default: input_file.format)")
	audioCmd.Flags().StringP("format", "f", "mp3", "Audio format (mp3, wav, aac, flac)")
	
	rootCmd.AddCommand(equalCmd)
	rootCmd.AddCommand(timeCmd)
	rootCmd.AddCommand(durationCmd)
	rootCmd.AddCommand(audioCmd)
}
