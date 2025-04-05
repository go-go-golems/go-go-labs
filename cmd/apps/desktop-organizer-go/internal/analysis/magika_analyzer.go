package analysis

import (
	"context"
	"encoding/json"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"

	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/config"
	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/log"
)

// MagikaTypeAnalyzer uses Magika to analyze file types.
type MagikaTypeAnalyzer struct {
	// Maps directories to count of files processed for sampling
	dirCount map[string]int
	mu       sync.Mutex
}

// Magika output structure for JSON parsing
type magikaJSONOutput struct {
	Path   string `json:"path"`
	Result struct {
		Status string `json:"status"`
		Value  struct {
			DL struct {
				Description string   `json:"description"`
				Extensions  []string `json:"extensions"`
				Group       string   `json:"group"`
				IsText      bool     `json:"is_text"`
				Label       string   `json:"label"`
				MimeType    string   `json:"mime_type"`
			} `json:"dl"`
			Output struct {
				Description string   `json:"description"`
				Extensions  []string `json:"extensions"`
				Label       string   `json:"label"`
				Group       string   `json:"group"`
				MIME        string   `json:"mime_type"`
				IsText      bool     `json:"is_text"`
			} `json:"output"`
			Score float64 `json:"score"`
		} `json:"value"`
	} `json:"result"`
}

// NewMagikaTypeAnalyzer creates a new Magika file type analyzer.
func NewMagikaTypeAnalyzer() *MagikaTypeAnalyzer {
	return &MagikaTypeAnalyzer{
		dirCount: make(map[string]int),
	}
}

// Name returns the analyzer name.
func (a *MagikaTypeAnalyzer) Name() string {
	return "MagikaTypeAnalyzer"
}

// Type returns the analyzer type.
func (a *MagikaTypeAnalyzer) Type() AnalyzerType {
	return FileAnalyzer
}

// DependsOn returns analyzer dependencies.
func (a *MagikaTypeAnalyzer) DependsOn() []string {
	return []string{} // No dependencies
}

// Analyze performs file type analysis using Magika.
func (a *MagikaTypeAnalyzer) Analyze(ctx context.Context, cfg *config.Config, toolProvider ToolProvider, result *AnalysisResult, entry *FileEntry) error {
	if entry == nil || entry.IsDir {
		return nil // Nothing to do for directories or nil entries
	}

	logger := log.FromCtx(ctx)

	// Get tool runner for Magika
	runner, ok := toolProvider.GetToolRunner("magika")
	if !ok {
		// If Magika isn't available, do nothing
		// (in a real implementation, we might fallback to FileTypeAnalyzer)
		logger.Debug().
			Str("file", entry.Path).
			Msg("Skipping Magika analysis (tool not available)")
		return nil
	}

	// Check sampling limits if configured
	if cfg.SamplingPerDir > 0 {
		// Get the parent directory
		dir := filepath.Dir(entry.Path)

		// Check if we've already processed too many files in this directory
		a.mu.Lock()
		count := a.dirCount[dir]
		if count >= cfg.SamplingPerDir {
			a.mu.Unlock()
			logger.Debug().
				Str("file", entry.Path).
				Str("dir", dir).
				Int("count", count).
				Int("limit", cfg.SamplingPerDir).
				Msg("Skipping Magika analysis (sampling limit reached)")
			return nil
		}
		// Increment the count
		a.dirCount[dir]++
		a.mu.Unlock()
	}

	// Run Magika with --json flag
	logger.Debug().
		Str("file", entry.Path).
		Msg("Running Magika analysis")

	stdout, stderr, err := runner.Run(ctx, "--json", entry.FullPath)
	if err != nil {
		return errors.Wrapf(err, "magika failed: %s", stderr)
	}

	// Log raw output if trace level is enabled
	logger.Trace().RawJSON("raw_magika_output", stdout).Msg("Received raw Magika output")

	// Parse the JSON output - Magika outputs an array even for a single file
	var magikaOutputSlice []magikaJSONOutput
	if err := json.Unmarshal(stdout, &magikaOutputSlice); err != nil {
		return errors.Wrap(err, "failed to parse magika JSON output")
	}

	// Check if we got exactly one result as expected
	if len(magikaOutputSlice) != 1 {
		return errors.Errorf("unexpected number of results (%d) in magika JSON output for single file %s", len(magikaOutputSlice), entry.Path)
	}

	magikaOutput := magikaOutputSlice[0]

	// Update entry with type information
	entry.TypeInfo["source"] = "magika"
	entry.TypeInfo["label"] = magikaOutput.Result.Value.Output.Label
	entry.TypeInfo["group"] = magikaOutput.Result.Value.Output.Group
	entry.TypeInfo["mime"] = magikaOutput.Result.Value.Output.MIME

	logger.Debug().
		Str("file", entry.Path).
		Str("type", magikaOutput.Result.Value.Output.Label).
		Str("group", magikaOutput.Result.Value.Output.Group).
		Msg("Magika analysis complete")

	return nil
}
