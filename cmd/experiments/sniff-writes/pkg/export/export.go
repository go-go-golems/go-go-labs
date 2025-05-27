package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/models"
)

// ExportFormat represents the supported export formats
type ExportFormat string

const (
	FormatJSON     ExportFormat = "json"
	FormatCSV      ExportFormat = "csv"
	FormatMarkdown ExportFormat = "markdown"
)

// Exporter handles exporting events in different formats
type Exporter struct {
	writer io.Writer
	format ExportFormat
}

// New creates a new Exporter
func New(writer io.Writer, format ExportFormat) *Exporter {
	return &Exporter{
		writer: writer,
		format: format,
	}
}

// Export exports the events in the specified format
func (e *Exporter) Export(events []models.EventOutput) error {
	switch e.format {
	case FormatJSON:
		return e.exportJSON(events)
	case FormatCSV:
		return e.exportCSV(events)
	case FormatMarkdown:
		return e.exportMarkdown(events)
	default:
		return fmt.Errorf("unsupported export format: %s", e.format)
	}
}

func (e *Exporter) exportJSON(events []models.EventOutput) error {
	encoder := json.NewEncoder(e.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(events)
}

func (e *Exporter) exportCSV(events []models.EventOutput) error {
	writer := csv.NewWriter(e.writer)
	defer writer.Flush()

	// Write header
	header := []string{"timestamp", "pid", "process", "operation", "filename", "fd", "write_size", "content", "truncated"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, event := range events {
		record := []string{
			event.Timestamp,
			strconv.FormatUint(uint64(event.Pid), 10),
			event.Process,
			event.Operation,
			event.Filename,
			strconv.FormatInt(int64(event.Fd), 10),
			strconv.FormatUint(event.WriteSize, 10),
			event.Content,
			strconv.FormatBool(event.Truncated),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

func (e *Exporter) exportMarkdown(events []models.EventOutput) error {
	// Write markdown table header
	if _, err := e.writer.Write([]byte("# File Events Report\n\n")); err != nil {
		return err
	}

	if len(events) == 0 {
		if _, err := e.writer.Write([]byte("No events found.\n")); err != nil {
			return err
		}
		return nil
	}

	if _, err := e.writer.Write([]byte("| Timestamp | PID | Process | Operation | Filename | FD | Write Size | Content | Truncated |\n")); err != nil {
		return err
	}
	if _, err := e.writer.Write([]byte("|-----------|-----|---------|-----------|----------|----|-----------|---------|-----------|\n")); err != nil {
		return err
	}

	// Write data rows
	for _, event := range events {
		content := event.Content
		if len(content) > 50 {
			content = content[:47] + "..."
		}
		// Escape markdown characters
		content = strings.ReplaceAll(content, "|", "\\|")
		
		row := fmt.Sprintf("| %s | %d | %s | %s | %s | %d | %d | %s | %t |\n",
			event.Timestamp,
			event.Pid,
			event.Process,
			event.Operation,
			event.Filename,
			event.Fd,
			event.WriteSize,
			content,
			event.Truncated,
		)
		if _, err := e.writer.Write([]byte(row)); err != nil {
			return err
		}
	}

	return nil
}