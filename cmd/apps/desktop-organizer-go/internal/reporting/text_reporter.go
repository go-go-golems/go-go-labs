package reporting

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/analysis"
	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/log"
	"github.com/pkg/errors"
)

// TextReporter generates a simple text-based report.
type TextReporter struct{}

// NewTextReporter creates a new TextReporter.
func NewTextReporter() *TextReporter {
	return &TextReporter{}
}

// इंश्योर करें कि TextReporter रिपोर्टर इंटरफ़ेस को लागू करता है।
var _ Reporter = &TextReporter{}

// FormatName returns the format name "text".
func (r *TextReporter) FormatName() string {
	return "text"
}

// GenerateReport generates a text report and writes it to the writer.
func (r *TextReporter) GenerateReport(ctx context.Context, result *analysis.AnalysisResult, writer io.Writer) error {
	logger := log.FromCtx(ctx) // Get logger from context
	logger.Debug().Msg("Generating text report")

	// Basic Summary
	_, err := fmt.Fprintf(writer, "Analysis Report for: %s\n", result.RootDir)
	if err != nil {
		return errors.Wrap(err, "failed to write header")
	}
	_, err = fmt.Fprintf(writer, "Scan Time: %s to %s\n",
		result.ScanStartTime.Format(time.RFC3339),
		result.ScanEndTime.Format(time.RFC3339))
	if err != nil {
		return errors.Wrap(err, "failed to write scan times")
	}

	scanDuration := result.ScanEndTime.Sub(result.ScanStartTime)
	_, err = fmt.Fprintf(writer, "Duration: %s\n", scanDuration.String())
	if err != nil {
		return errors.Wrap(err, "failed to write duration")
	}

	_, err = fmt.Fprintf(writer, "Totals: %d files, %d directories, %s total size\n",
		result.TotalFiles,
		result.TotalDirs,
		humanize.Bytes(uint64(result.TotalSize))) // Use humanize for size
	if err != nil {
		return errors.Wrap(err, "failed to write totals")
	}
	_, err = fmt.Fprintln(writer, "--- File Details ---")
	if err != nil {
		return errors.Wrap(err, "failed to write details separator")
	}

	// Use tabwriter for aligned columns
	tw := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0) // Min width 0, tab width 0, padding 2, pad char space, flags 0

	// Write header
	_, err = fmt.Fprintln(tw, "Path\tSize\tModified\tType (Magika)\tMIME")
	if err != nil {
		return errors.Wrap(err, "failed to write table header")
	}
	_, err = fmt.Fprintln(tw, "----\t----\t--------\t-------------\t----")
	if err != nil {
		return errors.Wrap(err, "failed to write table header separator")
	}

	// Write file entries
	for _, entry := range result.FileEntries {
		if entry.IsDir {
			continue // Skip directories for now in this basic report
		}

		// Extract Magika info safely
		magikaLabel := "(unknown)"
		if labelAny, ok := entry.TypeInfo["label"]; ok {
			if labelStr, okAssert := labelAny.(string); okAssert && labelStr != "" {
				magikaLabel = labelStr
			}
		}
		magikaMIME := "(unknown)"
		if mimeAny, ok := entry.TypeInfo["mime"]; ok {
			if mimeStr, okAssert := mimeAny.(string); okAssert && mimeStr != "" {
				magikaMIME = mimeStr
			}
		}

		_, err = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			entry.Path,
			humanize.Bytes(uint64(entry.Size)),
			entry.ModTime.Format("2006-01-02"), // Just date for brevity
			magikaLabel,
			magikaMIME,
		)
		if err != nil {
			// Log error but continue if possible
			logger.Error().Err(err).Str("file", entry.Path).Msg("Error writing file entry to report")
		}
	}

	// Flush the tabwriter buffer
	err = tw.Flush()
	if err != nil {
		return errors.Wrap(err, "failed to flush tabwriter")
	}

	logger.Debug().Msg("Text report generation complete")
	return nil
}
