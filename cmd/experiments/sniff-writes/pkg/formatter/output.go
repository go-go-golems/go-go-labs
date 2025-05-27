package formatter

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/models"
)

func CreateEventOutput(event *models.Event, resolvedPath string, config *models.Config) models.EventOutput {
	comm := cString(event.Comm[:])

	// Format filename to be relative and user-friendly
	displayFilename := formatFilename(resolvedPath)

	eventOutput := models.EventOutput{
		Timestamp:   time.Now().Format(time.RFC3339),
		Pid:         event.Pid,
		Process:     comm,
		Filename:    displayFilename,
		FileOffset:  event.FileOffset,
		NewOffset:   event.NewOffset,
		ChunkSeq:    event.ChunkSeq,
		TotalChunks: event.TotalChunks,
	}

	// Add write/read-specific information if available
	if (event.Type == 2 || event.Type == 1) && event.WriteSize > 0 { // write or read event
		eventOutput.WriteSize = event.WriteSize

		if config.CaptureContent && event.ContentLen > 0 {
			// Respect user's content size limit
			contentLen := event.ContentLen
			if int(contentLen) > config.ContentSize {
				contentLen = uint32(config.ContentSize)
			}
			content := cString(event.Content[:contentLen])
			eventOutput.Content = content
			// For chunked events, check if this chunk is truncated or if there are more chunks
			eventOutput.Truncated = (event.WriteSize > uint64(contentLen)) || (event.TotalChunks > 1)
		}
	}

	switch event.Type {
	case 0:
		eventOutput.Operation = "open"
	case 1:
		eventOutput.Operation = "read"
		if config.ShowFd {
			eventOutput.Fd = event.Fd
		}
	case 2:
		eventOutput.Operation = "write"
		if config.ShowFd {
			eventOutput.Fd = event.Fd
		}
	case 3:
		eventOutput.Operation = "close"
		if config.ShowFd {
			eventOutput.Fd = event.Fd
		}
	case 4:
		eventOutput.Operation = "lseek"
		if config.ShowFd {
			eventOutput.Fd = event.Fd
		}
		// Add human-readable whence value
		switch event.Whence {
		case 0:
			eventOutput.Whence = "SEEK_SET"
		case 1:
			eventOutput.Whence = "SEEK_CUR"
		case 2:
			eventOutput.Whence = "SEEK_END"
		default:
			eventOutput.Whence = fmt.Sprintf("unknown(%d)", event.Whence)
		}
	}

	return eventOutput
}

func OutputPlain(event models.EventOutput, writer *os.File, config *models.Config) {
	fdInfo := ""
	if config.ShowFd && event.Fd != 0 {
		fdInfo = fmt.Sprintf(" (fd: %d)", event.Fd)
	}

	switch event.Operation {
	case "open":
		fmt.Fprintf(writer, "[%s] Process %s (PID %d) opening file: %s\n",
			event.Timestamp, event.Process, event.Pid, event.Filename)
	case "read":
		sizeInfo := ""
		if event.WriteSize > 0 {
			sizeInfo = fmt.Sprintf(" (%d bytes)", event.WriteSize)
		}
		offsetInfo := ""
		if event.FileOffset > 0 {
			offsetInfo = fmt.Sprintf(" at offset %d", event.FileOffset)
		}
		chunkInfo := ""
		if event.TotalChunks > 1 {
			chunkInfo = fmt.Sprintf(" [chunk %d/%d]", event.ChunkSeq+1, event.TotalChunks)
		}
		fmt.Fprintf(writer, "[%s] Process %s (PID %d) reading from file: %s%s%s%s%s\n",
			event.Timestamp, event.Process, event.Pid, event.Filename, fdInfo, sizeInfo, offsetInfo, chunkInfo)

		if config.CaptureContent && event.Content != "" {
			truncated := ""
			if event.Truncated {
				truncated = " [TRUNCATED]"
			}
			fmt.Fprintf(writer, "    Content: %q%s\n", event.Content, truncated)
		}
	case "write":
		sizeInfo := ""
		if event.WriteSize > 0 {
			sizeInfo = fmt.Sprintf(" (%d bytes)", event.WriteSize)
		}
		offsetInfo := ""
		if event.FileOffset > 0 {
			offsetInfo = fmt.Sprintf(" at offset %d", event.FileOffset)
		}
		chunkInfo := ""
		if event.TotalChunks > 1 {
			chunkInfo = fmt.Sprintf(" [chunk %d/%d]", event.ChunkSeq+1, event.TotalChunks)
		}
		fmt.Fprintf(writer, "[%s] Process %s (PID %d) writing to file: %s%s%s%s%s\n",
			event.Timestamp, event.Process, event.Pid, event.Filename, fdInfo, sizeInfo, offsetInfo, chunkInfo)

		if config.CaptureContent && event.Content != "" {
			truncated := ""
			if event.Truncated {
				truncated = " [TRUNCATED]"
			}
			fmt.Fprintf(writer, "    Content: %q%s\n", event.Content, truncated)
		}

		if config.ShowDiffs && event.Diff != "" {
			if config.NoColor {
				fmt.Fprintf(writer, "    Diff:\n%s\n", event.Diff)
			} else {
				diffFormatter := NewColoredDiffFormatter()
				coloredDiff := diffFormatter.FormatDiff(event.Diff)
				fmt.Fprintf(writer, "    Diff:\n%s\n", coloredDiff)
			}
		}
	case "lseek":
		whenceInfo := ""
		if event.Whence != "" {
			whenceInfo = fmt.Sprintf(" (%s)", event.Whence)
		}
		offsetInfo := fmt.Sprintf(" to offset %d", event.FileOffset)
		fmt.Fprintf(writer, "[%s] Process %s (PID %d) seeking in file: %s%s%s%s\n",
			event.Timestamp, event.Process, event.Pid, event.Filename, fdInfo, offsetInfo, whenceInfo)
	case "close":
		fmt.Fprintf(writer, "[%s] Process %s (PID %d) closing file descriptor%s\n",
			event.Timestamp, event.Process, event.Pid, fdInfo)
	}
}

func OutputJSON(event models.EventOutput, writer *os.File, config *models.Config) {
	data, err := json.Marshal(event)
	if err != nil {
		if config.Verbose || config.Debug {
			log.Printf("failed to marshal JSON: %v", err)
		}
		return
	}
	fmt.Fprintf(writer, "%s\n", data)
}

func PrintTableHeader(writer *os.File, config *models.Config) {
	if config.ShowFd {
		fmt.Fprintf(writer, "%-8s %-12s %-8s %-8s %-8s %s\n",
			"TIME", "PROCESS", "PID", "OPERATION", "FD", "FILENAME")
		fmt.Fprintf(writer, "%-8s %-12s %-8s %-8s %-8s %s\n",
			"--------", "------------", "--------", "--------", "--------", "--------")
	} else {
		fmt.Fprintf(writer, "%-8s %-12s %-8s %-8s %s\n",
			"TIME", "PROCESS", "PID", "OPERATION", "FILENAME")
		fmt.Fprintf(writer, "%-8s %-12s %-8s %-8s %s\n",
			"--------", "------------", "--------", "--------", "--------")
	}
}

func OutputTable(event models.EventOutput, writer *os.File, config *models.Config) {
	// Truncate timestamp to show only time part for better readability
	timestamp := event.Timestamp
	if len(timestamp) > 19 {
		timestamp = timestamp[11:19] // Extract HH:MM:SS part
	}

	fdCol := ""
	if config.ShowFd && event.Fd != 0 {
		fdCol = fmt.Sprintf("%d", event.Fd)
	}

	// Truncate filename if too long for better table formatting
	filename := event.Filename
	if len(filename) > 50 {
		filename = "..." + filename[len(filename)-47:]
	}

	if config.ShowFd {
		fmt.Fprintf(writer, "%-8s %-12s %-8d %-8s %-8s %s\n",
			timestamp, event.Process, event.Pid, event.Operation, fdCol, filename)
	} else {
		fmt.Fprintf(writer, "%-8s %-12s %-8d %-8s %s\n",
			timestamp, event.Process, event.Pid, event.Operation, filename)
	}
}

func formatFilename(filename string) string {
	if filename == "" {
		return ""
	}

	cwd, err := os.Getwd()
	if err != nil {
		return filename
	}

	// Try to make path relative to current directory
	if rel, err := filepath.Rel(cwd, filename); err == nil && !filepath.IsAbs(rel) {
		// Only use relative path if it doesn't start with ../ (i.e., it's actually under cwd)
		if !filepath.IsAbs(rel) && len(rel) > 0 && rel[0] != '.' {
			return rel
		}
		if !filepath.IsAbs(rel) && len(rel) > 2 && rel[:2] != ".." {
			return rel
		}
	}

	return filename
}

func cString(b []int8) string {
	n := -1
	for i, v := range b {
		if v == 0 {
			n = i
			break
		}
	}
	if n == -1 {
		n = len(b)
	}
	// Convert []int8 to []byte
	bytes := make([]byte, n)
	for i := 0; i < n; i++ {
		bytes[i] = byte(b[i])
	}
	return string(bytes)
}
