package reporting

import (
	"context"
	"encoding/json"
	"io"

	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/analysis"
	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/log"
)

// JSONReporter implements the Reporter interface for JSON output.
type JSONReporter struct {
	// Optional configuration fields could be added here
	pretty bool
}

// NewJSONReporter creates a new JSON reporter.
func NewJSONReporter(pretty bool) *JSONReporter {
	return &JSONReporter{
		pretty: pretty,
	}
}

// FormatName returns the name of the output format.
func (r *JSONReporter) FormatName() string {
	return "json"
}

// GenerateReport writes the formatted report to the writer.
func (r *JSONReporter) GenerateReport(ctx context.Context, result *analysis.AnalysisResult, writer io.Writer) error {
	logger := log.FromCtx(ctx)
	logger.Debug().Msg("Generating JSON report")

	var (
		data []byte
		err  error
	)

	if r.pretty {
		// Pretty format with indentation
		data, err = json.MarshalIndent(result, "", "  ")
	} else {
		// Compact format
		data, err = json.Marshal(result)
	}

	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}
