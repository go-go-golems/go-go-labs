// Package mp3lib provides tools to simulate slow writing processes,
// useful for testing timing issues with asynchronous IO operations.
package mp3lib

import (
	"io"
	"time"
)

// SlowWriter wraps an io.Writer to simulate slower write speeds
// based on a specified duration per kilobyte. This can be useful
// for testing how systems handle slow IO situations.
type SlowWriter struct {
	w                   io.Writer
	durationPerKilobyte time.Duration
}

// NewSlowWriter creates a new SlowWriter instance that wraps the provided io.Writer.
// It uses the provided durationPerKilobyte to determine how long to sleep between
// writing chunks of data.
//
// Parameters:
//   - w: The underlying io.Writer to which the data will be written.
//   - durationPerKilobyte: The time it takes to write each kilobyte of data.
func NewSlowWriter(w io.Writer, durationPerKilobyte time.Duration) *SlowWriter {
	return &SlowWriter{
		w:                   w,
		durationPerKilobyte: durationPerKilobyte,
	}
}

// Write writes the provided byte slice to the underlying writer in chunks,
// and sleeps between chunks to simulate a slower writing process.
func (sw *SlowWriter) Write(p []byte) (n int, err error) {
	chunkSize := 1024
	totalWritten := 0

	for len(p) > 0 {
		// Determine the size of the current chunk.
		currentChunkSize := chunkSize
		if len(p) < chunkSize {
			currentChunkSize = len(p)
		}

		// Write the chunk.
		written, err := sw.w.Write(p[:currentChunkSize])
		totalWritten += written
		if err != nil {
			return totalWritten, err
		}

		// Sleep to simulate the desired bandwidth.
		time.Sleep(sw.durationPerKilobyte * time.Duration(currentChunkSize) / 1024)

		// Move to the next chunk.
		p = p[currentChunkSize:]
	}

	return totalWritten, nil
}
