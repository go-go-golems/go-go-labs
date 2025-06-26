package video

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"split-video/pkg/config"
)

// VideoInfo holds metadata about a video file
type VideoInfo struct {
	Duration time.Duration
	Width    int
	Height   int
	Bitrate  int
	Format   string
}

// GetVideoInfo extracts metadata from a video file using ffprobe
func GetVideoInfo(filename string) (*VideoInfo, error) {
	log.Debug().Str("file", filename).Msg("Getting video info")
	
	cmd := exec.Command("ffprobe", 
		"-v", "quiet",
		"-show_entries", "format=duration:stream=width,height,bit_rate",
		"-of", "csv=p=0",
		filename)
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}
	
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("unexpected ffprobe output format")
	}
	
	// Parse video stream info (first line)
	streamInfo := strings.Split(lines[0], ",")
	if len(streamInfo) < 3 {
		return nil, fmt.Errorf("unexpected stream info format")
	}
	
	width, _ := strconv.Atoi(streamInfo[0])
	height, _ := strconv.Atoi(streamInfo[1])
	bitrate, _ := strconv.Atoi(streamInfo[2])
	
	// Parse format info (last line)
	duration, err := strconv.ParseFloat(lines[len(lines)-1], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration: %w", err)
	}
	
	return &VideoInfo{
		Duration: time.Duration(duration * float64(time.Second)),
		Width:    width,
		Height:   height,
		Bitrate:  bitrate,
		Format:   filepath.Ext(filename),
	}, nil
}

// SplitEqual splits a video into equal segments
func SplitEqual(cfg *config.Config) error {
	log.Info().Msg("Starting equal split operation")
	
	videoInfo, err := GetVideoInfo(cfg.InputFile)
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}
	
	log.Info().
		Dur("duration", videoInfo.Duration).
		Int("segments", cfg.Segments).
		Msg("Video info obtained")
	
	// Calculate segment duration
	segmentDuration := videoInfo.Duration / time.Duration(cfg.Segments)
	
	// Adjust for overlap
	overlapDuration := cfg.Overlap
	adjustedSegmentDuration := segmentDuration + overlapDuration
	
	log.Debug().
		Dur("segment_duration", segmentDuration).
		Dur("overlap", overlapDuration).
		Dur("adjusted_segment_duration", adjustedSegmentDuration).
		Msg("Calculated segment parameters")
	
	// Create output directory
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	baseName := getBaseName(cfg.InputFile)
	
	for i := 0; i < cfg.Segments; i++ {
		startTime := time.Duration(i) * (segmentDuration - overlapDuration)
		
		// Don't go beyond the video duration
		if startTime >= videoInfo.Duration {
			break
		}
		
		// Adjust duration for the last segment
		duration := adjustedSegmentDuration
		if startTime+duration > videoInfo.Duration {
			duration = videoInfo.Duration - startTime
		}
		
		outputFile := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_segment_%d.mp4", baseName, i+1))
		
		log.Info().
			Int("segment", i+1).
			Dur("start", startTime).
			Dur("duration", duration).
			Str("output", outputFile).
			Msg("Processing segment")
		
		if err := extractSegment(cfg.InputFile, outputFile, startTime, duration); err != nil {
			return fmt.Errorf("failed to extract segment %d: %w", i+1, err)
		}
		
		// Extract audio if requested
		if cfg.ExtractAudio {
			audioFile := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_segment_%d.%s", baseName, i+1, cfg.AudioFormat))
			if err := ExtractAudio(outputFile, audioFile, cfg.AudioFormat); err != nil {
				log.Warn().Err(err).Int("segment", i+1).Msg("Failed to extract audio")
			}
		}
	}
	
	return nil
}

// SplitByTime splits a video at specific time intervals
func SplitByTime(cfg *config.Config) error {
	log.Info().Msg("Starting time-based split operation")
	
	videoInfo, err := GetVideoInfo(cfg.InputFile)
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}
	
	// Parse time intervals
	intervals := make([]time.Duration, 0, len(cfg.Intervals))
	for _, interval := range cfg.Intervals {
		duration, err := time.ParseDuration(interval)
		if err != nil {
			return fmt.Errorf("failed to parse interval '%s': %w", interval, err)
		}
		intervals = append(intervals, duration)
	}
	
	// Add start and end times
	allTimes := append([]time.Duration{0}, intervals...)
	allTimes = append(allTimes, videoInfo.Duration)
	
	// Create output directory
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	baseName := getBaseName(cfg.InputFile)
	
	for i := 0; i < len(allTimes)-1; i++ {
		startTime := allTimes[i]
		endTime := allTimes[i+1]
		duration := endTime - startTime
		
		outputFile := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_part_%d.mp4", baseName, i+1))
		
		log.Info().
			Int("part", i+1).
			Dur("start", startTime).
			Dur("end", endTime).
			Dur("duration", duration).
			Str("output", outputFile).
			Msg("Processing part")
		
		if err := extractSegment(cfg.InputFile, outputFile, startTime, duration); err != nil {
			return fmt.Errorf("failed to extract part %d: %w", i+1, err)
		}
		
		// Extract audio if requested
		if cfg.ExtractAudio {
			audioFile := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_part_%d.%s", baseName, i+1, cfg.AudioFormat))
			if err := ExtractAudio(outputFile, audioFile, cfg.AudioFormat); err != nil {
				log.Warn().Err(err).Int("part", i+1).Msg("Failed to extract audio")
			}
		}
	}
	
	return nil
}

// SplitByDuration splits a video into segments of specific duration
func SplitByDuration(cfg *config.Config) error {
	log.Info().Msg("Starting duration-based split operation")
	
	videoInfo, err := GetVideoInfo(cfg.InputFile)
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}
	
	segmentDuration := cfg.SegmentDuration
	overlapDuration := cfg.Overlap
	adjustedSegmentDuration := segmentDuration + overlapDuration
	
	// Create output directory
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	baseName := getBaseName(cfg.InputFile)
	segmentCount := 0
	
	for startTime := time.Duration(0); startTime < videoInfo.Duration; {
		segmentCount++
		
		// Adjust duration for the last segment
		duration := adjustedSegmentDuration
		if startTime+duration > videoInfo.Duration {
			duration = videoInfo.Duration - startTime
		}
		
		outputFile := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_chunk_%d.mp4", baseName, segmentCount))
		
		log.Info().
			Int("chunk", segmentCount).
			Dur("start", startTime).
			Dur("duration", duration).
			Str("output", outputFile).
			Msg("Processing chunk")
		
		if err := extractSegment(cfg.InputFile, outputFile, startTime, duration); err != nil {
			return fmt.Errorf("failed to extract chunk %d: %w", segmentCount, err)
		}
		
		// Extract audio if requested
		if cfg.ExtractAudio {
			audioFile := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_chunk_%d.%s", baseName, segmentCount, cfg.AudioFormat))
			if err := ExtractAudio(outputFile, audioFile, cfg.AudioFormat); err != nil {
				log.Warn().Err(err).Int("chunk", segmentCount).Msg("Failed to extract audio")
			}
		}
		
		// Move to next segment (subtract overlap to create overlap)
		startTime += segmentDuration - overlapDuration
	}
	
	return nil
}

// ExtractAudio extracts audio from a video file
func ExtractAudio(inputFile, outputFile, format string) error {
	log.Info().
		Str("input", inputFile).
		Str("output", outputFile).
		Str("format", format).
		Msg("Extracting audio")
	
	var codec string
	switch format {
	case "mp3":
		codec = "libmp3lame"
	case "wav":
		codec = "pcm_s16le"
	case "aac":
		codec = "aac"
	case "flac":
		codec = "flac"
	default:
		return fmt.Errorf("unsupported audio format: %s", format)
	}
	
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-vn", // No video
		"-acodec", codec,
		"-y", // Overwrite output files
		outputFile)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract audio: %w", err)
	}
	
	return nil
}

// extractSegment extracts a segment from a video file
func extractSegment(inputFile, outputFile string, startTime, duration time.Duration) error {
	startSeconds := startTime.Seconds()
	durationSeconds := duration.Seconds()
	
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-ss", fmt.Sprintf("%.2f", startSeconds),
		"-t", fmt.Sprintf("%.2f", durationSeconds),
		"-c", "copy", // Copy streams without re-encoding
		"-y", // Overwrite output files
		outputFile)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract segment: %w", err)
	}
	
	return nil
}

// getBaseName returns the base name of a file without extension
func getBaseName(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}
