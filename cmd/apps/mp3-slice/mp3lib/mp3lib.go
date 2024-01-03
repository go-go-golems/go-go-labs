// Package mp3lib provides utilities for working with MP3 files.
// It depends on the external tools 'mp3cut' and 'mp3length'.
//
// mp3length provides the length of an MP3 file.
// Usage: mp3length mp3file
// Output Format: Length of mp3file: hh:mm:ss+ms
//
// mp3cut is used to extract segments from an MP3 file.
// Usage: mp3cut [-o outputfile] [-T title] [-A artist] [-N album-name] [-t [hh:]mm:ss[+ms]-[hh:]mm:ss[+ms]] mp3 [-t ...] mp3
// Parameters:
//
//	-o output: Output file, default mp3file.out.mp3
//	-T title: Set title metadata for the output mp3
//	-A artist: Set artist metadata for the output mp3
//	-N album-name: Set album name metadata for the output mp3
//	-t: Define the segment to extract in the format [hh:]mm:ss[+ms]-[hh:]mm:ss[+ms]
package mp3lib

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
)

var (
	// regexPattern is used to parse the output of mp3length
	regexPattern = regexp.MustCompile(`Length of .*: (\d{2}):(\d{2}):(\d{2})\+(\d{1,3})`)
)

// GetLengthSeconds returns the length of the provided MP3 file in seconds.
// It utilizes the mp3length tool to obtain the length.
//
// Returns:
//
//	Length of the MP3 file in seconds.
//	Error, if any (e.g., if mp3length tool encounters an issue).
func GetLengthSeconds(mp3Path string) (int, error) {
	cmd := exec.Command("mp3length", mp3Path)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	matches := regexPattern.FindStringSubmatch(string(output))
	if len(matches) < 5 {
		return 0, errors.New("invalid output format from mp3length")
	}

	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.Atoi(matches[3])

	return hours*3600 + minutes*60 + seconds, nil
}

// ExtractSectionToFile extracts a section of the MP3 file and saves it to a given path.
// It utilizes the mp3cut tool to extract the segment.
func ExtractSectionToFile(mp3Path, outputPath string, startSec, endSec int) error {
	startTime := formatTime(startSec)
	endTime := formatTime(endSec)

	cmd := exec.Command("mp3cut", "-o", outputPath, "-t", fmt.Sprintf("%s-%s", startTime, endTime), mp3Path)
	return cmd.Run()
}

// ExtractSectionToWriter extracts a section of the MP3 file and writes it to the provided io.Writer.
//
// It utilizes the mp3cut tool and a named FIFO pipe to communicate between the tool and the writer.
// This function blocks until the mp3cut process has finished and the FIFO has been read out in its entirety.
func ExtractSectionToWriter(ctx context.Context, mp3Path string, w io.Writer, startSec, endSec int) error {
	g, ctx := errgroup.WithContext(ctx)

	fifoPath := filepath.Join(os.TempDir(), "mp3fifo")
	err := syscall.Mkfifo(fifoPath, 0666)
	if err != nil {
		return err
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(fifoPath)

	startTime := formatTime(startSec)
	endTime := formatTime(endSec)

	r, pipeW := io.Pipe()

	// Start the mp3cut process in a separate goroutine.
	g.Go(func() error {
		cmd := exec.CommandContext(ctx, "mp3cut", "-o", fifoPath, "-t", fmt.Sprintf("%s-%s", startTime, endTime), mp3Path)
		cmd.Stdout = pipeW
		cmd.Stderr = pipeW
		if err := cmd.Run(); err != nil {
			return err
		}
		return pipeW.Close()
	})

	// Copy data from the pipe reader to the writer.
	g.Go(func() error {
		_, err := CopyWithCancel(ctx, w, r)
		return err
	})

	// Wait for all tasks to complete or return on the first error.
	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

// formatTime is a utility function that converts seconds to the format hh:mm:ss.
//
// It's primarily for use with the mp3cut command.
func formatTime(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60

	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
