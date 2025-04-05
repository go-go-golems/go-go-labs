package analysis

import (
	"context"

	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/config"
	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/tools"
)

// AnalyzerType defines the type of analyzer (how it should be executed)
type AnalyzerType int

const (
	// FileAnalyzer operates on individual files and can run concurrently
	FileAnalyzer AnalyzerType = iota
	// AggregateAnalyzer operates on the full dataset and runs sequentially
	AggregateAnalyzer
)

// ToolProvider interface to give analyzers access to tool runners
type ToolProvider interface {
	GetToolRunner(name string) (tools.Runner, bool)
}

// Analyzer performs a specific analysis task.
type Analyzer interface {
	// Name returns a unique identifier for the analyzer (e.g., "MagikaTypeAnalyzer").
	Name() string

	// Type returns whether this is a file-level or aggregate analyzer.
	Type() AnalyzerType

	// DependsOn returns names of analyzers that must run before this one.
	DependsOn() []string

	// Analyze performs the analysis.
	// For file-level analyzers (Type() == FileAnalyzer), it operates on a single FileEntry.
	// For aggregate analyzers (Type() == AggregateAnalyzer), the entry parameter is ignored.
	// Both types update the AnalysisResult as needed.
	// The ToolProvider gives access to configured external tool runners.
	Analyze(ctx context.Context, cfg *config.Config, toolProvider ToolProvider, result *AnalysisResult, entry *FileEntry) error
}

// Registry manages the available analyzers.
type Registry struct {
	analyzers map[string]Analyzer
}

// NewRegistry creates a new analyzer registry.
func NewRegistry() *Registry {
	return &Registry{
		analyzers: make(map[string]Analyzer),
	}
}

// Register adds an analyzer to the registry.
func (r *Registry) Register(analyzer Analyzer) {
	r.analyzers[analyzer.Name()] = analyzer
}

// GetAnalyzer retrieves an analyzer by name.
func (r *Registry) GetAnalyzer(name string) (Analyzer, bool) {
	analyzer, exists := r.analyzers[name]
	return analyzer, exists
}

// ListAnalyzers returns all registered analyzers.
func (r *Registry) ListAnalyzers() []Analyzer {
	result := make([]Analyzer, 0, len(r.analyzers))
	for _, analyzer := range r.analyzers {
		result = append(result, analyzer)
	}
	return result
}

// FilterAnalyzers returns a filtered list of analyzers based on enabled/disabled lists.
// If enabled is empty, all analyzers except those in disabled are returned.
func (r *Registry) FilterAnalyzers(enabled, disabled []string) []Analyzer {
	result := make([]Analyzer, 0)

	// Check if we're using an explicit enable list
	useExplicitEnabled := len(enabled) > 0

	// Create a map for O(1) lookups
	disabledMap := make(map[string]bool)
	for _, name := range disabled {
		disabledMap[name] = true
	}

	enabledMap := make(map[string]bool)
	for _, name := range enabled {
		enabledMap[name] = true
	}

	// Filter analyzers
	for name, analyzer := range r.analyzers {
		// Skip if explicitly disabled
		if disabledMap[name] {
			continue
		}

		// If we have an explicit enabled list, only include analyzers on that list
		if useExplicitEnabled && !enabledMap[name] {
			continue
		}

		result = append(result, analyzer)
	}

	return result
}
