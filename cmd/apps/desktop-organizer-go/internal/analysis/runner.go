package analysis

import (
	"context"
	"io/fs"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/config"
	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/log"
	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/tools"
)

// PathMatcher interface for matching paths against patterns
type PathMatcher interface {
	Match(path string) bool
}

// GlobMatcher implements the PathMatcher interface using filepath.Match
type GlobMatcher struct {
	pattern string
}

// Match checks if a path matches the glob pattern
func (g *GlobMatcher) Match(path string) bool {
	match, err := filepath.Match(g.pattern, path)
	return err == nil && match
}

// Runner orchestrates the analysis process.
type Runner struct {
	cfg             *config.Config
	analyzers       []Analyzer // Ordered list of enabled analyzers
	fileAnalyzers   []Analyzer // File-level analyzers only
	aggrAnalyzers   []Analyzer // Aggregate analyzers only
	toolRunners     map[string]tools.Runner
	excludeMatchers []PathMatcher // Path matchers for exclusion
}

// Verify Runner implements ToolProvider interface
var _ ToolProvider = (*Runner)(nil)

// NewRunner creates a new analysis runner.
func NewRunner(ctx context.Context, cfg *config.Config, registry *Registry) (*Runner, error) {
	runner := &Runner{
		cfg:         cfg,
		toolRunners: make(map[string]tools.Runner),
	}

	// Compile exclude patterns
	if err := runner.compileExcludePatterns(cfg.ExcludePaths); err != nil {
		return nil, errors.Wrap(err, "failed to compile exclude patterns")
	}

	// Initialize analyzers
	if err := runner.setupAnalyzers(ctx, cfg, registry); err != nil {
		return nil, errors.Wrap(err, "failed to setup analyzers")
	}

	return runner, nil
}

// compileExcludePatterns compiles the glob patterns for path exclusion.
func (r *Runner) compileExcludePatterns(patterns []string) error {
	r.excludeMatchers = make([]PathMatcher, 0, len(patterns))
	for _, pattern := range patterns {
		// Validate the pattern by trying to compile it first
		_, err := filepath.Match(pattern, "")
		if err != nil {
			return errors.Wrapf(err, "invalid glob pattern: %s", pattern)
		}
		r.excludeMatchers = append(r.excludeMatchers, &GlobMatcher{pattern: pattern})
	}
	return nil
}

// setupAnalyzers initializes the list of analyzers to run.
func (r *Runner) setupAnalyzers(ctx context.Context, cfg *config.Config, registry *Registry) error {
	// Filter analyzers based on config
	filteredAnalyzers := registry.FilterAnalyzers(cfg.EnabledAnalyzers, cfg.DisabledAnalyzers)

	// Sort analyzers based on dependencies (topological sort)
	// TODO: Implement actual topological sort
	r.analyzers = filteredAnalyzers

	// Separate analyzers by type
	r.fileAnalyzers = make([]Analyzer, 0)
	r.aggrAnalyzers = make([]Analyzer, 0)

	for _, analyzer := range r.analyzers {
		switch analyzer.Type() {
		case FileAnalyzer:
			r.fileAnalyzers = append(r.fileAnalyzers, analyzer)
		case AggregateAnalyzer:
			r.aggrAnalyzers = append(r.aggrAnalyzers, analyzer)
		}
	}

	// Initialize tool runners for external tools
	if err := r.setupToolRunners(ctx, cfg); err != nil {
		return errors.Wrap(err, "failed to setup tool runners")
	}

	return nil
}

// setupToolRunners initializes runners for external tools.
func (r *Runner) setupToolRunners(ctx context.Context, cfg *config.Config) error {
	logger := log.FromCtx(ctx)

	// Setup standard tools with potential custom paths from config
	toolSetups := []struct {
		name    string
		factory func(string) (tools.Runner, error)
	}{
		{"magika", tools.NewMagikaRunner},
		{"exiftool", tools.NewExiftoolRunner},
		{"jdupes", tools.NewJdupesRunner},
		{"file", tools.NewFileRunner},
	}

	for _, tool := range toolSetups {
		customPath := cfg.ToolPaths[tool.name]
		runner, err := tool.factory(customPath)
		if err != nil {
			logger.Warn().
				Err(err).
				Str("tool", tool.name).
				Msg("Tool unavailable")
			continue
		}

		// Check if the tool is actually runnable
		if err := runner.Check(ctx); err != nil {
			logger.Warn().
				Err(err).
				Str("tool", tool.name).
				Str("path", runner.ToolPath()).
				Msg("Tool check failed")
			continue
		}

		r.toolRunners[tool.name] = runner
		logger.Debug().
			Str("tool", tool.name).
			Str("path", runner.ToolPath()).
			Msg("Tool registered")
	}

	return nil
}

// GetToolRunner returns the tool runner for the specified tool, or nil if not available.
func (r *Runner) GetToolRunner(name string) (tools.Runner, bool) {
	runner, ok := r.toolRunners[name]
	return runner, ok
}

// IsExcluded checks if a path should be excluded based on exclude patterns.
func (r *Runner) IsExcluded(path string) bool {
	// Always exclude "." and ".." directories
	base := filepath.Base(path)
	if base == "." || base == ".." {
		return true
	}

	// Check against pattern matchers
	for _, matcher := range r.excludeMatchers {
		if matcher.Match(path) {
			return true
		}
	}

	return false
}

// Run executes the analysis pipeline.
func (r *Runner) Run(ctx context.Context) (*AnalysisResult, error) {
	logger := log.FromCtx(ctx)
	logger.Info().Str("dir", r.cfg.TargetDir).Msg("Starting directory analysis")

	// Create result object
	result := NewAnalysisResult(r.cfg.TargetDir, r.cfg)

	// Record tool status
	for name, runner := range r.toolRunners {
		result.AddToolStatus(name, runner.ToolPath(), true, nil)
	}

	// --- Phase 1: File Discovery and Basic Metadata ---
	logger.Info().Msg("Phase 1: File discovery")
	if err := r.discoverFiles(ctx, result); err != nil {
		return result, errors.Wrap(err, "file discovery failed")
	}

	// --- Phase 2: Concurrent File Analysis ---
	logger.Info().Int("workerCount", r.cfg.MaxWorkers).Msg("Phase 2: Concurrent file analysis")
	if err := r.analyzeFiles(ctx, result); err != nil {
		return result, errors.Wrap(err, "file analysis failed")
	}

	// --- Phase 3: Aggregate Analysis ---
	logger.Info().Msg("Phase 3: Aggregate analysis")
	if err := r.runAggregateAnalysis(ctx, result); err != nil {
		return result, errors.Wrap(err, "aggregate analysis failed")
	}

	// --- Finalization ---
	result.ScanEndTime = time.Now()
	logger.Info().
		Int("totalFiles", result.TotalFiles).
		Int("totalDirs", result.TotalDirs).
		Int64("totalSize", result.TotalSize).
		Float64("durationSec", result.ScanEndTime.Sub(result.ScanStartTime).Seconds()).
		Msg("Analysis completed")

	return result, nil
}

// discoverFiles walks the directory tree and creates FileEntry objects.
func (r *Runner) discoverFiles(ctx context.Context, result *AnalysisResult) error {
	logger := log.FromCtx(ctx)

	// Normalize root dir to absolute path
	rootDir, err := filepath.Abs(r.cfg.TargetDir)
	if err != nil {
		return errors.Wrapf(err, "failed to get absolute path for %s", r.cfg.TargetDir)
	}

	result.RootDir = rootDir

	// Walk the directory tree
	return filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}

		// Handle walk errors
		if err != nil {
			logger.Warn().Err(err).Str("path", path).Msg("Error accessing path")
			return nil // Continue walking
		}

		// Skip the root dir itself
		if path == rootDir {
			return nil
		}

		// Get relative path for reporting
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			logger.Warn().Err(err).Str("path", path).Msg("Failed to get relative path")
			relPath = path // Fallback to full path
		}

		// Check exclusion patterns
		if r.IsExcluded(relPath) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Get file info for size, mod time, etc.
		info, err := d.Info()
		if err != nil {
			logger.Warn().Err(err).Str("path", path).Msg("Failed to get file info")
			return nil // Continue walking
		}

		// Create and add FileEntry
		entry := NewFileEntry(
			relPath,
			path,
			d.IsDir(),
			info.Size(),
			info.ModTime(),
			info.Mode(),
		)

		// Add entry to result and update counters
		result.AddFileEntry(entry)

		if logger.GetLevel() <= zerolog.DebugLevel && (result.TotalFiles+result.TotalDirs)%1000 == 0 {
			logger.Debug().
				Int("files", result.TotalFiles).
				Int("dirs", result.TotalDirs).
				Int64("size", result.TotalSize).
				Msg("Directory walk progress")
		}

		return nil
	})
}

// analyzeFiles processes each non-directory entry with file-level analyzers.
func (r *Runner) analyzeFiles(ctx context.Context, result *AnalysisResult) error {
	if len(r.fileAnalyzers) == 0 || len(result.FileEntries) == 0 {
		return nil
	}

	logger := log.FromCtx(ctx)

	// Create worker pool
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(r.cfg.MaxWorkers)

	// Create a channel to feed file entries to workers
	entries := make(chan *FileEntry)

	// Track progress
	var processedCount int32
	totalToProcess := 0

	// Count non-directory entries for progress reporting
	for _, entry := range result.FileEntries {
		if !entry.IsDir {
			totalToProcess++
		}
	}

	// Start workers
	for i := 0; i < r.cfg.MaxWorkers; i++ {
		g.Go(func() error {
			for entry := range entries {
				// Skip directories
				if entry.IsDir {
					continue
				}

				// Process entry
				if err := r.processFileEntry(gctx, result, entry); err != nil {
					entry.AddError(err)
					logger.Error().
						Err(err).
						Str("path", entry.Path).
						Msg("Error processing file")
					// Continue with other files
				}

				// Update and log progress periodically
				processed := atomic.AddInt32(&processedCount, 1)
				if logger.GetLevel() <= zerolog.DebugLevel && processed%100 == 0 {
					logger.Debug().
						Int32("processed", processed).
						Int("total", totalToProcess).
						Float64("percent", float64(processed)/float64(totalToProcess)*100).
						Msg("File analysis progress")
				}
			}
			return nil
		})
	}

	// Feed entries to workers
	go func() {
		defer close(entries)
		for _, entry := range result.FileEntries {
			select {
			case <-gctx.Done():
				return
			case entries <- entry:
				// Entry sent to worker
			}
		}
	}()

	// Wait for all workers to finish
	return g.Wait()
}

// processFileEntry applies all file-level analyzers to a single entry.
func (r *Runner) processFileEntry(ctx context.Context, result *AnalysisResult, entry *FileEntry) error {
	for _, analyzer := range r.fileAnalyzers {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}

		// Run analyzer, passing the runner itself as the ToolProvider
		if err := analyzer.Analyze(ctx, r.cfg, r, result, entry); err != nil {
			return errors.Wrapf(err, "analyzer %s failed", analyzer.Name())
		}
	}
	return nil
}

// runAggregateAnalysis runs all aggregate analyzers on the complete dataset.
func (r *Runner) runAggregateAnalysis(ctx context.Context, result *AnalysisResult) error {
	logger := log.FromCtx(ctx)

	// Run aggregate analyzers sequentially
	for _, analyzer := range r.aggrAnalyzers {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue processing
		}

		logger.Debug().Str("analyzer", analyzer.Name()).Msg("Running aggregate analyzer")

		// Run analyzer (pass runner as ToolProvider, nil for entry since this is aggregate analysis)
		if err := analyzer.Analyze(ctx, r.cfg, r, result, nil); err != nil {
			result.AddError(err)
			logger.Error().
				Err(err).
				Str("analyzer", analyzer.Name()).
				Msg("Aggregate analyzer failed")
			// Continue with other analyzers
		}
	}

	return nil
}
