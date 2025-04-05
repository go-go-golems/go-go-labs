package reporting

import (
	"context"
	"fmt"
	"io"

	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/analysis"
)

// Reporter defines an interface for formatting and writing analysis results.
type Reporter interface {
	// FormatName returns the name of the output format (e.g., "json", "text", "markdown").
	FormatName() string

	// GenerateReport writes the formatted report to the writer.
	GenerateReport(ctx context.Context, result *analysis.AnalysisResult, writer io.Writer) error
}

// Registry manages available reporters.
type Registry struct {
	reporters map[string]Reporter
}

// NewRegistry creates a new reporter registry.
func NewRegistry() *Registry {
	return &Registry{
		reporters: make(map[string]Reporter),
	}
}

// Register adds a reporter to the registry.
func (r *Registry) Register(reporter Reporter) {
	r.reporters[reporter.FormatName()] = reporter
}

// GetReporter retrieves a reporter by format name.
func (r *Registry) GetReporter(format string) (Reporter, error) {
	reporter, exists := r.reporters[format]
	if !exists {
		return nil, fmt.Errorf("no reporter found for format %q, available formats: %v", format, r.AvailableFormats())
	}
	return reporter, nil
}

// AvailableFormats returns a list of available report formats.
func (r *Registry) AvailableFormats() []string {
	formats := make([]string, 0, len(r.reporters))
	for format := range r.reporters {
		formats = append(formats, format)
	}
	return formats
}
