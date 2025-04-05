# Technical Design: desktop-organizer-go

## 1. Introduction

### 1.1. Purpose
This document details the technical design for `desktop-organizer-go`, a Go application intended to replace the existing Bash script (`cmd/apps/desktop-organizer/01-inspect-downloads-folder.sh`). The primary goal is to analyze the contents of a user-specified directory (typically a Downloads folder) to gather rich metadata about the files within it.

### 1.2. Goals
*   **Refactor Bash Script:** Port the core functionality of the Bash script to Go for improved maintainability, testability, and error handling.
*   **Concurrency:** Leverage Go's concurrency features to speed up file system scanning and analysis, especially for large directories.
*   **Modularity & Extensibility:** Design the application with clear interfaces and components to easily add new analysis types (Analyzers) or output formats (Reporters).
*   **Structured Output:** Produce well-defined, structured output (primarily JSON) suitable for consumption by other programs, particularly Large Language Models (LLMs), to facilitate automated directory cleanup and organization tasks.
*   **Robustness:** Implement proper error handling and reporting.
*   **Tool Integration:** Provide a flexible way to integrate external command-line tools (like `magika`, `exiftool`, `jdupes`) where reimplementation is impractical.

### 1.3. Target Audience
This document is intended for developers involved in building or maintaining the `desktop-organizer-go` application. It assumes familiarity with Go programming concepts.

## 2. High-Level Architecture

The application follows a pipeline architecture orchestrated by a central `Runner`.

1.  **Initialization:** Parse command-line arguments (Cobra), load configuration (Viper), and set up logging (zerolog). Check for the availability of required external tools.
2.  **Phase 1: File Discovery & Basic Metadata:** Recursively scan the target directory using `filepath.WalkDir`. For each file, gather basic metadata (`os.FileInfo`) and create an initial `analysis.FileEntry` object. This phase is performed concurrently where possible (e.g., multiple `stat` calls).
3.  **Phase 2: Concurrent File-Level Analysis:** Utilize a worker pool (`errgroup`) to process `FileEntry` objects concurrently. Apply registered `Analyzer` components that operate on individual files (e.g., file type detection via `magika`, hashing via `crypto/md5`). Analyzers update the `FileEntry` objects with their results (e.g., adding type info, hashes).
4.  **Phase 3: Aggregate Analysis:** Once all individual file analysis is complete, run `Analyzer` components that require the full dataset (e.g., duplicate detection based on hashes, generating type summaries, monthly summaries). These update the main `analysis.AnalysisResult` object.
5.  **Phase 4: Reporting:** Use a configured `Reporter` component (e.g., `JsonReporter`, `TextReporter`) to format the final `AnalysisResult` and write it to the specified output (stdout or a file).

```mermaid
flowchart TD
    subgraph Initialization
        A[Start: main.Execute] --> B(Parse Flags & Config - Cobra/Viper)
        B --> C(Setup Logging - zerolog)
        C --> D(Check External Tools)
    end

    subgraph Analysis Pipeline (Runner.Run)
        D --> E[Phase 1: Walk Directory - filepath.WalkDir]
        E --> F(Create FileEntry objects)
        F --> G[Phase 2: Concurrent File Analysis - Worker Pool / errgroup]
        G -- Applies --> H(File-Level Analyzers: Magika, Hasher, etc.)
        H -- Updates --> F
        G --> I[Phase 3: Aggregate Analysis]
        I -- Applies --> J(Aggregate Analyzers: DuplicateFinder, Summarizer)
        J -- Updates --> K(AnalysisResult)
    end

    subgraph Reporting
        K --> L[Phase 4: Select Reporter - Text/JSON]
        L --> M(Format AnalysisResult)
        M --> N(Write to Output - stdout/file)
        N --> O[End]
    end

    Initialization --> Analysis Pipeline
    Analysis Pipeline --> Reporting
```

## 3. Command Line Interface (CLI)

*   **Framework:** `github.com/spf13/cobra`
*   **Main Command:** `desktop-organizer`
*   **Flags:**
    *   `--config, -c`: (String) Path to config file.
    *   `--verbose, -v`: (Bool) Enable verbose/debug logging.
    *   `--debug-log`: (String) Path to write detailed debug logs to a file.
    *   `--downloads-dir, -d`: (String, **Required**) The target directory to analyze.
    *   `--output-file, -o`: (String) File path for the report output (default: stdout).
    *   `--sample-per-dir, -s`: (Int) Max files per directory for type analysis (0=disabled).
    *   `--max-workers`: (Int) Number of concurrent workers for file analysis (default: 4 or based on CPU).
    *   `--output-format`: (String) Report format: `text`, `json`, `markdown` (default: `text`).
    *   `--exclude-path`: (String Slice) Glob patterns for paths to exclude during scanning.
    *   `--large-file-mb`: (Int) Threshold in MB to tag files as 'large' (default: 100).
    *   `--recent-days`: (Int) Threshold in days to tag files as 'recent' (default: 30).
    *   `--tool-path`: (String Slice) Override path for external tools (e.g., `--tool-path magika=/usr/local/bin/magika`).
    *   `--enable-analyzer`: (String Slice) Explicitly enable specific analyzers (if needed for finer control).
    *   `--disable-analyzer`: (String Slice) Explicitly disable specific analyzers.

## 4. Configuration Management

*   **Framework:** `github.com/spf13/viper`
*   **Loading Priority:**
    1.  Command-line flags.
    2.  Environment variables (prefixed with `DESKTOP_ORGANIZER_`, e.g., `DESKTOP_ORGANIZER_TARGET_DIR`).
    3.  Configuration file.
    4.  Default values.
*   **Config File:** YAML format preferred (`.desktop-organizer.yaml`). Located in `$HOME` or current directory (`.`).
*   **`Config` Struct (`internal/config/config.go`):** Defines the application's configuration settings. Mirrors CLI flags where applicable but allows for more complex structures (e.g., maps for tool paths, analyzer settings).

```go
package config

// Config holds all configuration settings for the application.
type Config struct {
	TargetDir          string                 `mapstructure:"targetDir"`
	OutputFile         string                 `mapstructure:"outputFile"`
	DebugLog           string                 `mapstructure:"debugLog"`
	Verbose            bool                   `mapstructure:"verbose"`
	SamplingPerDir     int                    `mapstructure:"samplingPerDir"`
	MaxWorkers         int                    `mapstructure:"maxWorkers"`
	OutputFormat       string                 `mapstructure:"outputFormat"`
	ExcludePaths       []string               `mapstructure:"excludePaths"` // Glob patterns
	LargeFileThreshold int64                  `mapstructure:"largeFileThreshold"` // Bytes
	RecentFileDays     int                    `mapstructure:"recentFileDays"`
	ToolPaths          map[string]string      `mapstructure:"toolPaths"`      // e.g., "magika": "/usr/bin/magika"
	EnabledAnalyzers   []string               `mapstructure:"enabledAnalyzers"` // If empty, use default set
	DisabledAnalyzers  []string               `mapstructure:"disabledAnalyzers"`
	AnalyzerConfigs    map[string]map[string]any `mapstructure:"analyzerConfigs"` // Specific settings per analyzer
}

// LoadConfig initializes Viper and unmarshals the configuration.
func LoadConfig(cmd *cobra.Command, cfgFile string) (*Config, error) {
    // ... Viper initialization (set defaults, bind flags, read env, read config file) ...
    
    // Unmarshal into Config struct
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("unable to decode config: %w", err)
    }

    // Post-processing/validation (e.g., convert MB to Bytes for threshold)
    config.LargeFileThreshold = config.LargeFileThreshold * 1024 * 1024 // Assuming input is MB

    return &config, nil
}

```

## 5. Logging

*   **Framework:** `github.com/rs/zerolog`
*   **Default Output:** Console (stderr) with human-friendly formatting (`zerolog.ConsoleWriter`).
*   **Log Level:** Controlled by `--verbose` flag (`InfoLevel` default, `DebugLevel` if verbose).
*   **File Logging:** If `--debug-log` is specified, create a secondary logger writing detailed (Debug level) JSON logs to the specified file.
*   **Setup:** Initialize logging early in `rootCmd.PersistentPreRunE`. A dedicated package `internal/log` might handle setting up the multi-writer (console + optional file).

## 6. Core Data Structures (`internal/analysis/types.go`)

These structs define how analysis data is stored and structured. JSON tags (`json:"..."`) are crucial for the structured output goal.

```go
package analysis

import (
	"io/fs"
	"time"
)

// FileEntry holds information about a scanned file system entry.
type FileEntry struct {
	Path     string      `json:"path"` // Relative to TargetDir
	FullPath string      `json:"-"`    // Absolute path
	IsDir    bool        `json:"is_dir"`
	Size     int64       `json:"size"`
	ModTime  time.Time   `json:"mod_time"`
	Mode     fs.FileMode `json:"-"` 

	TypeInfo map[string]any `json:"type_info,omitempty"` // Source ("magika", "file"), Label, Group, MIME
	Metadata map[string]any `json:"metadata,omitempty"`  // Source ("exiftool"), Key-Value pairs
	Hashes   map[string]string `json:"hashes,omitempty"`    // Algorithm ("md5", "sha256") -> Hash string
	Tags     []string       `json:"tags,omitempty"`      // "LargeFile", "RecentFile", "DuplicateSetID:xyz"
	Error    string         `json:"error,omitempty"`     // Record file-specific processing errors
}

// AnalysisResult contains the overall results of a directory analysis run.
type AnalysisResult struct {
	RootDir         string                `json:"root_dir"`
	ScanStartTime   time.Time             `json:"scan_start_time"`
	ScanEndTime     time.Time             `json:"scan_end_time"`
	TotalFiles      int                   `json:"total_files"`
	TotalDirs       int                   `json:"total_dirs"`
	TotalSize       int64                 `json:"total_size"`
	FileEntries     []*FileEntry          `json:"file_entries"` // Consider omitting if very large, provide summary only?
	TypeSummary     map[string]*TypeStats `json:"type_summary"` // Keyed by type label
	DuplicateSets   []*DuplicateSet       `json:"duplicate_sets,omitempty"`
	MonthlySummary  map[string]*MonthStats `json:"monthly_summary"` // Keyed by "YYYY-MM"
	ToolStatus      map[string]*ToolInfo  `json:"tool_status"`
	OverallErrors   []string              `json:"overall_errors,omitempty"`
	Config          *config.Config        `json:"config_used"` // Include the config used for this run
}

type TypeStats struct {
	Label string   `json:"-"`
	Count int      `json:"count"`
	Size  int64    `json:"size"`
	Paths []string `json:"-"` // Temporary storage during aggregation
}

type DuplicateSet struct {
	ID          string   `json:"id"` // Hash or external tool ID
	FilePaths   []string `json:"file_paths"` // Relative paths
	Size        int64    `json:"size"` // Size of one file
	Count       int      `json:"count"`
	WastedSpace int64    `json:"wasted_space"`
}

type MonthStats struct {
	YearMonth string `json:"-"` // YYYY-MM
	Count     int    `json:"count"`
	Size      int64  `json:"size"`
}

type ToolInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path,omitempty"`
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}
```

## 7. Key Interfaces

### 7.1. `Analyzer` (`internal/analysis/analyzer.go`)
Defines components responsible for specific analysis tasks.

```go
package analysis

import "context"

// Analyzer performs a specific analysis task.
type Analyzer interface {
	// Name returns a unique identifier for the analyzer (e.g., "MagikaTypeAnalyzer").
	Name() string
	// DependsOn returns names of analyzers that must run before this one.
	DependsOn() []string
	// Analyze performs the analysis.
	// For file-level analyzers, it typically updates the passed *FileEntry.
	// For aggregate analyzers, it updates the *AnalysisResult.
	Analyze(ctx context.Context, cfg *config.Config, result *AnalysisResult, entry *FileEntry) error
}
```

### 7.2. `ExternalToolRunner` (`internal/tools/runner.go`)
Abstracts the execution of external command-line tools.

```go
package tools

import "context"

type Runner interface {
	Check(ctx context.Context) error
	Run(ctx context.Context, args ...string) (stdout []byte, stderr []byte, err error)
	RunWithInput(ctx context.Context, input []byte, args ...string) (stdout []byte, stderr []byte, err error)
	ToolName() string
	ToolPath() string
}

// NewRunner creates a concrete implementation using os/exec.
// func NewRunner(name, path string) Runner
```

### 7.3. `Reporter` (`internal/reporting/reporter.go`)
Defines components responsible for formatting the final output.

```go
package reporting

import (
	"context"
	"io"
	"github.com/your_org/project/internal/analysis" // Use correct import path
)

type Reporter interface {
	FormatName() string // e.g., "json", "text"
	GenerateReport(ctx context.Context, result *analysis.AnalysisResult, writer io.Writer) error
}
```

## 8. Orchestration: `Runner` (`internal/analysis/runner.go`)

The `Runner` struct manages the entire analysis process.

```go
package analysis

import (
    "context"
    "fmt"
    "path/filepath"
    "sync"
    // other imports: config, log, tools, errgroup, os, path/filepath
)

type Runner struct {
	cfg        *config.Config
	analyzers  []Analyzer          // Ordered list of enabled analyzers
	toolRunners map[string]tools.Runner
}

func NewRunner(cfg *config.Config) (*Runner, error) {
    // 1. Discover available analyzers (e.g., via a registry)
    // 2. Filter based on cfg.EnabledAnalyzers / cfg.DisabledAnalyzers
    // 3. Perform topological sort based on DependsOn()
    // 4. Check availability of required external tools via tools.NewRunner & Check()
    // 5. Store tool runners and ordered analyzers
}

func (r *Runner) Run(ctx context.Context) (*AnalysisResult, error) {
	result := &AnalysisResult{ /* ... initialize ... */ }
	result.Config = r.cfg // Store config used

	// --- Phase 1: File Discovery ---
	log.Ctx(ctx).Info().Str("dir", r.cfg.TargetDir).Msg("Starting directory scan")
	fileChan := make(chan string) // Channel for discovered file paths
	var wgWalk sync.WaitGroup
	var walkErr error
	
	wgWalk.Add(1)
	go func() {
		defer wgWalk.Done()
		defer close(fileChan)
		walkErr = filepath.WalkDir(r.cfg.TargetDir, func(path string, d fs.DirEntry, err error) error {
			// Handle entry errors, check exclusions (r.cfg.ExcludePaths)
            // Stat the file to get initial info
            // Create initial FileEntry (relative path, basic info)
            // Send *FileEntry (or just path) to the next stage channel
            // Increment result.TotalFiles/TotalDirs/TotalSize (atomically or use mutex)
		})
	}()

    // --- Phase 2: Concurrent File Analysis ---
	log.Ctx(ctx).Info().Msg("Starting concurrent file analysis")
    entryChan := make(chan *FileEntry) // Channel for entries needing analysis
    var wgProcess sync.WaitGroup
    g, workerCtx := errgroup.WithContext(ctx)
    g.SetLimit(r.cfg.MaxWorkers)

    // Launch worker goroutines
    for i := 0; i < r.cfg.MaxWorkers; i++ {
        wgProcess.Add(1)
        go func() {
            defer wgProcess.Done()
            for entry := range entryChan {
                processEntry(workerCtx, r.cfg, r.analyzers, entry) // Apply file-level analyzers
                // Send processed entry to the next stage (aggregation)
            }
        }()
    }

    // Feed entries from file discovery to workers
    // Need to collect processed entries for Phase 3
    processedEntries := make([]*FileEntry, 0)
    var mu sync.Mutex // Mutex to protect processedEntries slice

    go func() {
        // Read from fileChan (or intermediate channel with initial FileEntry)
        // Send to entryChan for workers
        // Collect results from workers (via another channel?)
        // mu.Lock(); processedEntries = append(processedEntries, processedEntry); mu.Unlock()
    }()

	wgWalk.Wait() // Ensure walking is done before closing entryChan feeder
    // Ensure feeder finishes before closing entryChan
    close(entryChan)
	wgProcess.Wait() // Wait for all workers to finish

	// Check for errors from WalkDir and workers (g.Wait())
    result.FileEntries = processedEntries // Assign collected entries

    // --- Phase 3: Aggregate Analysis ---
	log.Ctx(ctx).Info().Msg("Starting aggregate analysis")
    for _, analyzer := range r.analyzers {
        // Identify aggregate analyzers (e.g., by type assertion or a method)
        // Run aggregate analyzers sequentially (or concurrently if safe)
        // analyzer.Analyze(ctx, r.cfg, result, nil) // Pass nil for entry
    }

	// --- Finalization ---
	result.ScanEndTime = time.Now()
	log.Ctx(ctx).Info().Msg("Analysis finished")
    // Aggregate overall errors
	return result, combinedError
}

// processEntry applies file-level analyzers to a single entry
func processEntry(ctx context.Context, cfg *config.Config, analyzers []Analyzer, entry *FileEntry) {
    for _, analyzer := range analyzers {
        // Skip aggregate analyzers
        // Check dependencies satisfied (maybe pre-filtered)
        if err := analyzer.Analyze(ctx, cfg, nil, entry); err != nil {
            log.Ctx(ctx).Error().Err(err).Str("analyzer", analyzer.Name()).Str("file", entry.Path).Msg("Analyzer error")
            // Record error in entry.Error
        }
    }
}

```
*(Note: The Runner implementation above is conceptual and needs refinement regarding channel management, error aggregation, and analyzer scheduling.)*

## 9. Analyzer Implementations (`internal/analysis/*_analyzer.go`)

Planned analyzers:

*   **`BasicInfoAnalyzer`:** (Implicit in Phase 1) Gets `fs.FileInfo`.
*   **`MagikaTypeAnalyzer`:**
    *   Depends On: None.
    *   Action: Runs `magika --json <filepath>` via `ExternalToolRunner`. Parses JSON output. Updates `FileEntry.TypeInfo`. Handles sampling (`cfg.SamplingPerDir`).
*   **`FileTypeAnalyzer`:** (Fallback if Magika unavailable)
    *   Depends On: None.
    *   Action: Runs `file -b <filepath>` via `ExternalToolRunner`. Updates `FileEntry.TypeInfo`.
*   **`HashingAnalyzer`:**
    *   Depends On: None.
    *   Action: Reads file content, calculates MD5/SHA256 hash using `crypto/*`. Updates `FileEntry.Hashes`.
*   **`ExiftoolAnalyzer`:**
    *   Depends On: `MagikaTypeAnalyzer` (or `FileTypeAnalyzer`) to identify media types.
    *   Action: For files identified as media (image, video, audio), run `exiftool -s3 -Tag1 ... <filepath>` via `ExternalToolRunner`. Parses output. Updates `FileEntry.Metadata`. Might only run on top N largest media files as an optimization.
*   **`JdupesDuplicateAnalyzer`:**
    *   Depends On: None (runs externally on the whole directory).
    *   Action: Runs `jdupes -r -q -o name <dir>` and `jdupes -r -q -S <dir>` via `ExternalToolRunner`. Parses output. Populates `AnalysisResult.DuplicateSets`. Optionally adds tags (`DuplicateSetID:jdupes_set_1`) to `FileEntry.Tags`. This is more of an "aggregate" analyzer run early.
*   **`HashDuplicateAnalyzer`:** (Fallback if jdupes unavailable)
    *   Depends On: `HashingAnalyzer`.
    *   Action: (Aggregate Phase) Groups `FileEntry` objects by hash. Identifies sets with >1 entry. Populates `AnalysisResult.DuplicateSets`. Adds tags (`DuplicateSetID:<hash>`) to `FileEntry.Tags`.
*   **`SizeThresholdTagger`:**
    *   Depends On: None (uses basic info).
    *   Action: Checks `FileEntry.Size` against `cfg.LargeFileThreshold`. Adds tag "LargeFile" to `FileEntry.Tags`.
*   **`RecencyTagger`:**
    *   Depends On: None (uses basic info).
    *   Action: Checks `FileEntry.ModTime` against `time.Now() - days(cfg.RecentFileDays)`. Adds tag "RecentFile" to `FileEntry.Tags`.
*   **`SummaryGenerator`:**
    *   Depends On: All analyzers that contribute data (Type, Basic Info).
    *   Action: (Aggregate Phase) Iterates through all `FileEntry` objects. Populates `AnalysisResult.TypeSummary` and `AnalysisResult.MonthlySummary`.

## 10. External Tool Runner (`internal/tools/runner.go`)

*   Provides a consistent interface (`Runner`) for executing external commands.
*   The concrete implementation uses `os/exec`.
*   Handles context cancellation, capturing stdout/stderr, and error checking.
*   Factory functions (`NewMagikaRunner`, `NewExiftoolRunner`, etc.) can encapsulate default paths or specific argument parsing logic if needed, taking paths from `cfg.ToolPaths`.

## 11. Reporter Implementations (`internal/reporting/*_reporter.go`)

*   **`JsonReporter`:**
    *   Action: Marshals the `AnalysisResult` struct directly into JSON using `encoding/json`. Provides the most structured data.
*   **`TextReporter`:**
    *   Action: Iterates through `AnalysisResult`. Formats output similarly to the original Bash script using `fmt.Printf` and potentially `text/tabwriter` for alignment. Prioritizes human readability over machine parsing.
*   **`MarkdownReporter`:**
    *   Action: Formats output using Markdown tables and sections. Good balance between human readability and structure.

## 12. Concurrency Strategy

*   **File Discovery:** `filepath.WalkDir` is sequential by nature, but the processing of each found entry (stat call, initial `FileEntry` creation) *could* be parallelized slightly if I/O becomes a bottleneck, though WalkDir itself is often fast enough.
*   **File Analysis (Phase 2):** A fixed-size worker pool (`errgroup` with `SetLimit`) processes `FileEntry` objects from a channel. Each worker applies relevant file-level `Analyzer`s to one entry at a time. This parallelizes CPU-bound tasks (hashing) and I/O-bound tasks (running external tools like `magika`). `cfg.MaxWorkers` controls pool size.
*   **Aggregate Analysis (Phase 3):** Typically runs sequentially after Phase 2, as these analyzers often need the complete, processed dataset. Some aggregation tasks (like building summaries) might be parallelizable if broken down carefully.
*   **Synchronization:** Mutexes or atomic operations will be needed to safely update shared fields in `AnalysisResult` (like `TotalFiles`, `TotalSize`) from concurrent operations. Collecting processed `FileEntry` objects requires careful channel management or mutex-protected slices.

## 13. Project Structure

```
desktop-organizer-go/
├── cmd/
│   └── desktop-organizer/
│       └── main.go          # Cobra entrypoint
├── internal/
│   ├── analysis/            # Core analysis logic
│   │   ├── types.go         # FileEntry, AnalysisResult, etc.
│   │   ├── analyzer.go      # Analyzer interface
│   │   ├── runner.go        # Orchestrator (Runner struct, pipeline logic)
│   │   ├── magika_analyzer.go
│   │   ├── hash_analyzer.go
│   │   ├── duplicate_analyzer.go
│   │   └── summary_analyzer.go 
│   │   └── ... (other analyzers) ...
│   ├── config/
│   │   └── config.go        # Config struct, loading logic (Viper)
│   ├── log/
│   │   └── log.go           # Logging setup (zerolog, console/file writers)
│   ├── reporting/
│   │   ├── reporter.go      # Reporter interface
│   │   ├── json_reporter.go
│   │   ├── text_reporter.go
│   │   └── markdown_reporter.go
│   └── tools/
│       └── runner.go        # ExternalToolRunner interface and implementation
├── go.mod
├── go.sum
└── config.yaml.example      # Example configuration file
```

## 14. Future Considerations

*   **Caching:** Implement caching for expensive operations (e.g., file hashing, Magika results) based on file path and modification time.
*   **Advanced Exclusions:** Support more complex exclusion rules beyond simple glob patterns.
*   **Interactive Mode:** Add an interactive mode for duplicate deletion (similar to `jdupes -dN`).
*   **Plugin System:** Design a more formal plugin system for loading third-party Analyzers or Reporters.
*   **Web UI:** Create a simple web interface (perhaps using Go templates or a framework) to display results visually. 