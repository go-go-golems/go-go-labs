package audio

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

// NotificationSound represents a notification sound generator
type NotificationSound struct {
	SampleRate int
	Duration   time.Duration
	Frequency  float64
	Volume     float64
}

// NewNotificationSound creates a new notification sound generator with default values
func NewNotificationSound() *NotificationSound {
	return &NotificationSound{
		SampleRate: 44100,
		Duration:   500 * time.Millisecond,
		Frequency:  800.0, // Pleasant notification frequency
		Volume:     0.3,   // 30% volume to avoid being too loud
	}
}

// GenerateBeep generates a simple sine wave beep as WAV data
func (ns *NotificationSound) GenerateBeep() ([]byte, error) {
	samples := int(float64(ns.SampleRate) * ns.Duration.Seconds())

	// WAV header (44 bytes)
	header := make([]byte, 44)

	// RIFF header
	copy(header[0:4], "RIFF")
	// File size (will be filled later)
	copy(header[8:12], "WAVE")

	// Format chunk
	copy(header[12:16], "fmt ")
	// Format chunk size (16 for PCM)
	header[16] = 16
	// Audio format (1 = PCM)
	header[20] = 1
	// Number of channels (1 = mono)
	header[22] = 1
	// Sample rate
	sampleRate := uint32(ns.SampleRate)
	header[24] = byte(sampleRate)
	header[25] = byte(sampleRate >> 8)
	header[26] = byte(sampleRate >> 16)
	header[27] = byte(sampleRate >> 24)
	// Byte rate (sample rate * channels * bits per sample / 8)
	byteRate := sampleRate * 2 // 16-bit mono
	header[28] = byte(byteRate)
	header[29] = byte(byteRate >> 8)
	header[30] = byte(byteRate >> 16)
	header[31] = byte(byteRate >> 24)
	// Block align (channels * bits per sample / 8)
	header[32] = 2
	// Bits per sample
	header[34] = 16

	// Data chunk
	copy(header[36:40], "data")
	// Data size
	dataSize := uint32(samples * 2) // 16-bit samples
	header[40] = byte(dataSize)
	header[41] = byte(dataSize >> 8)
	header[42] = byte(dataSize >> 16)
	header[43] = byte(dataSize >> 24)

	// Update file size in header
	fileSize := uint32(36 + dataSize)
	header[4] = byte(fileSize)
	header[5] = byte(fileSize >> 8)
	header[6] = byte(fileSize >> 16)
	header[7] = byte(fileSize >> 24)

	// Generate audio data
	audioData := make([]byte, samples*2)
	for i := 0; i < samples; i++ {
		// Generate sine wave with fade in/out to avoid clicks
		t := float64(i) / float64(ns.SampleRate)

		// Fade in/out envelope to prevent clicks
		envelope := 1.0
		fadeTime := 0.01 // 10ms fade
		if t < fadeTime {
			envelope = t / fadeTime
		} else if t > ns.Duration.Seconds()-fadeTime {
			envelope = (ns.Duration.Seconds() - t) / fadeTime
		}

		// Generate sine wave
		sample := math.Sin(2*math.Pi*ns.Frequency*t) * ns.Volume * envelope

		// Convert to 16-bit signed integer
		sampleInt := int16(sample * 32767)

		// Store as little-endian
		audioData[i*2] = byte(sampleInt)
		audioData[i*2+1] = byte(sampleInt >> 8)
	}

	// Combine header and audio data
	result := make([]byte, len(header)+len(audioData))
	copy(result, header)
	copy(result[len(header):], audioData)

	return result, nil
}

// SaveToFile saves the generated beep to a WAV file
func (ns *NotificationSound) SaveToFile(filename string) error {
	data, err := ns.GenerateBeep()
	if err != nil {
		return errors.Wrap(err, "failed to generate beep")
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, "failed to create directory")
	}

	return os.WriteFile(filename, data, 0644)
}

// DownloadNotificationSound downloads a free notification sound from the internet
func DownloadNotificationSound(ctx context.Context, url, filename string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to download sound")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download sound: status %d", resp.StatusCode)
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap(err, "failed to create directory")
	}

	file, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to write file")
	}

	return nil
}

// GetFreeNotificationSounds returns a list of URLs for free notification sounds
func GetFreeNotificationSounds() map[string]string {
	return map[string]string{
		"gentle_bell": "https://www.soundjay.com/misc/sounds/bell-ringing-05.wav",
		"soft_chime":  "https://www.zapsplat.com/wp-content/uploads/2015/sound-effects-one/notification_simple-01.wav",
		// Note: These are example URLs. For production use, you should:
		// 1. Use royalty-free sound libraries like freesound.org
		// 2. Host your own sound files
		// 3. Use sounds with proper licensing
	}
}
