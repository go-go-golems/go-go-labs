package config

import (
	"fmt"
	"time"
)

// Config holds all configuration for the video splitting operations
type Config struct {
	// Input settings
	InputFile string

	// Output settings
	OutputDir string
	
	// Splitting settings
	Segments        int           // Number of segments for equal split
	Overlap         time.Duration // Overlap between segments
	SegmentDuration time.Duration // Duration of each segment
	Intervals       []string      // Time intervals for splitting
	
	// Audio settings
	ExtractAudio bool
	AudioFormat  string
	
	// Logging settings
	Verbose  bool
	LogLevel string
}

// SplitMode represents different ways to split video
type SplitMode int

const (
	SplitModeEqual SplitMode = iota
	SplitModeTime
	SplitModeDuration
)

// AudioFormat represents supported audio formats
type AudioFormat string

const (
	AudioFormatMP3  AudioFormat = "mp3"
	AudioFormatWAV  AudioFormat = "wav"
	AudioFormatAAC  AudioFormat = "aac"
	AudioFormatFLAC AudioFormat = "flac"
)

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.InputFile == "" {
		return fmt.Errorf("input file is required")
	}
	
	if c.OutputDir == "" {
		c.OutputDir = "."
	}
	
	if c.AudioFormat == "" {
		c.AudioFormat = "mp3"
	}
	
	return nil
}
