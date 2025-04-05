# Go Architecture Design: Desktop Organizer

## 1. Introduction

This document outlines the proposed Go architecture for the `desktop-organizer` application. The initial goal is to replicate and enhance the functionality of the `01-inspect-downloads-folder.sh` script, providing analysis of a target directory (e.g., Downloads folder). The architecture prioritizes modularity, testability, and extensibility.

## 2. Core Components & Packages

The application will be structured into several packages:

*   **`cmd/desktop-organizer`**: The main application entry point, using `cobra` for CLI parsing and command definition. It initializes configuration, logging, and the main `analyzer`.
*   **`pkg/analyzer`**: Contains the core analysis orchestration logic.
    *   `Analyzer` struct: Holds configuration, logger, and dependencies (interfaces).
    *   `Analyze()` method: Executes the different analysis steps.
    *   Defines core interfaces (`FileSystem`, `ToolRunner`, `Reporter`).
*   **`pkg/config`**: Defines the `Config` struct and handles loading configuration from flags and potentially files.
*   **`pkg/fs`**: Provides an implementation of the `analyzer.FileSystem` interface for interacting with the actual file system. Contains functions for listing files, getting stats, etc.
*   **`pkg/tools`**: Provides implementations of the `analyzer.ToolRunner` interface for executing external tools (`magika`, `exiftool`, `jdupes`). Includes logic for detecting tool availability.
*   **`pkg/report`**: Provides implementations of the `analyzer.Reporter` interface for generating the output (e.g., text file, JSON).
*   **`pkg/analysis`**: Contains the specific analysis modules/logic.
    *   **`stats`**: Basic directory statistics.
    *   **`types`**: File type detection (using `magika` or `file`).
    *   **`metadata`**: Media metadata extraction (using `exiftool`).
    *   **`duplicates`**: Duplicate file detection (using `jdupes` or hashing).
    *   **`temporal`**: Recent files and by-month analysis.
    *   **`size`**: Large file analysis.
    *   **`recommend`**: Recommendation generation.
*   **`pkg/log`**: Configures and provides access to the application logger (e.g., using `log/slog`).

## 3. Key Interfaces

Interfaces are crucial for decoupling components and enabling testing.

```go
package analyzer

import (
    "context"
    "io/fs"
    "time"
)

// FileInfo holds basic information about a file.
type FileInfo struct {
    Path       string
    Size       int64
    Mode       fs.FileMode
    ModTime    time.Time
    SysStat    any // To hold platform-specific stat info if needed
    IsDir      bool
    TypeLabel  string    // e.g., "JPEG image data" or "pdf"
    TypeGroup  string    // e.g., "image", "document", "archive" (from Magika)
    MIMEType   string    // Best guess MIME type
    // Potential future fields: Hash, Metadata, etc.
}

// FileSystem defines operations for interacting with the file system.
type FileSystem interface {
    // ListFiles recursively lists files in a directory, applying filters.
    ListFiles(ctx context.Context, rootDir string, excludePatterns []string) (<-chan FileInfo, error)
    // Stat returns basic info for a single path.
    Stat(ctx context.Context, path string) (FileInfo, error)
    // GetTotalSize calculates the total size of a directory.
    GetTotalSize(ctx context.Context, dir string) (int64, error)
    // ReadFile reads content (used for hashing if needed).
    ReadFile(ctx context.Context, path string) ([]byte, error)
}

// ToolRunner defines how to execute external tools.
type ToolRunner interface {
    // Run executes a command and returns its stdout, stderr, and error.
    Run(ctx context.Context, name string, args ...string) (stdout []byte, stderr []byte, err error)
    // CheckAvailable checks if a tool exists in PATH.
    CheckAvailable(ctx context.Context, name string) bool
    // RunWithInput runs a command, piping input and capturing output.
    RunWithInput(ctx context.Context, stdin []byte, name string, args ...string) (stdout []byte, stderr []byte, err error)
     // RunStreaming runs a command processing input/output via channels/pipes (e.g., for magika --jsonl)
    RunStreaming(ctx context.Context, inputPaths []string, name string, args ...string) (<-chan string, <-chan error, error) // Simplified example
}

// Reporter defines how analysis results are outputted.
type Reporter interface {
    // GenerateReport writes the complete analysis results.
    GenerateReport(ctx context.Context, reportData AnalysisReport, outputFile string) error
}

// --- Interfaces for specific analysis steps (could reside in pkg/analysis/*) ---

package analysis

// FileTypeDetector defines how to get type information for a file.
type FileTypeDetector interface {
    // DetectTypes takes a list of file paths and returns type info, potentially streaming results.
    DetectTypes(ctx context.Context, filePaths []string, samplePerDir int) (map[string]FileInfo, error) // Or return a channel
    // IsAvailable checks if the detector (e.g., Magika tool) is ready.
    IsAvailable(ctx context.Context) bool
}

// MetadataExtractor defines how to get specific metadata (e.g., media info).
type MetadataExtractor interface {
    ExtractMediaMetadata(ctx context.Context, filePaths []string) (map[string]MediaMetadata, error)
    IsAvailable(ctx context.Context) bool
}
type MediaMetadata struct {
    Resolution string
    Duration   string
    CreateDate time.Time
    FileType   string
    // ... other relevant fields
}


// DuplicateFinder defines how to find duplicate files within a directory.
type DuplicateFinder interface {
    FindDuplicates(ctx context.Context, rootDir string) (DuplicateInfo, error)
    IsAvailable(ctx context.Context) bool
}
type DuplicateInfo struct {
    Sets         [][]string // List of sets, each set is a list of paths
    WastedSpace  int64
    TotalFiles   int // Total number of files that are duplicates (excluding originals)
}

```
*Note: The `var _ Interface = (*Implementation)(nil)` pattern should be used in implementation files to ensure interfaces are correctly implemented.*

## 4. Data Structures

Key data structures will be defined to pass information between components.

```go
package analyzer

import "time"

// AnalysisReport aggregates all findings.
type AnalysisReport struct {
    Timestamp        time.Time
    RootDir          string
    BasicStats       BasicStatsInfo
    TypeSummary      map[string]TypeStats // Key: Type Label/Group
    MediaFiles       []MediaFileInfo      // Top N largest/most recent etc.
    DuplicateInfo    analysis.DuplicateInfo
    RecentFiles      []FileInfo           // Last N days
    LargeFiles       []FileInfo           // Over N MB
    FilesByMonth     map[string]MonthStats // Key: YYYY-MM
    Recommendations  []string
    ToolAvailability map[string]bool      // e.g., {"magika": true, "exiftool": false}
    Config           config.Config        // Include config used for the report
}

type BasicStatsInfo struct {
    TotalItemsTop    int
    TotalFilesTop    int
    TotalDirsTop     int
    TotalSize        int64 // Recursive size
    TotalFilesRec    int
}

type TypeStats struct {
    Count      int
    TotalSize  int64
}

type MediaFileInfo struct {
    FileInfo
    Metadata analysis.MediaMetadata
}

type MonthStats struct {
    Count     int
    TotalSize int64
}

// Other structs like config.Config, analysis.MediaMetadata, analysis.DuplicateInfo defined elsewhere.

```

## 5. Workflow

1.  **`main` (`cmd/desktop-organizer`)**:
    *   Parses command-line flags using `cobra`.
    *   Loads configuration (`pkg/config`).
    *   Sets up logging (`pkg/log`).
    *   Instantiates dependencies:
        *   `fs.DefaultFileSystem` (`analyzer.FileSystem`)
        *   `tools.DefaultToolRunner` (`analyzer.ToolRunner`)
        *   Analysis modules (`pkg/analysis/*`, potentially implementing interfaces like `FileTypeDetector`). These might take the `ToolRunner` as a dependency.
        *   `report.TextReporter` (`analyzer.Reporter`)
    *   Creates the main `analyzer.Analyzer` instance, injecting dependencies.
    *   Calls `analyzer.Analyze()`.
    *   Handles top-level errors.
2.  **`analyzer.Analyze()`**:
    *   Calls `FileSystem.GetTotalSize` and initial directory stats.
    *   Calls `FileSystem.ListFiles` to get a stream/list of all files. Collect file paths, basic stats (size, modtime) concurrently.
    *   Passes file list to `FileTypeDetector.DetectTypes` (handles sampling internally if configured). Updates `FileInfo` structs.
    *   Passes file list/info to other analysis modules (`metadata`, `duplicates`, `temporal`, `size`) potentially running some in parallel using `errgroup`.
        *   Modules use `ToolRunner` if necessary.
        *   Modules return their specific results (e.g., `DuplicateInfo`, lists of files).
    *   Aggregates results into the `AnalysisReport` struct.
    *   Calls `analysis.Recommender` to generate recommendations based on the report data.
    *   Calls `Reporter.GenerateReport` to write the final output.

## 6. Concurrency Model

*   **File Listing:** `FileSystem.ListFiles` can potentially use `filepath.WalkDir` which is sequential, but processing the resulting files (stats, hashing, initial type checks) can be parallelized using worker pools or `errgroup`.
*   **Tool Execution:** Running external tools like `magika` or `exiftool` on multiple files can be parallelized. `ToolRunner.RunStreaming` or managing multiple `ToolRunner.Run` calls within an `errgroup` would be suitable. Need to be mindful of resource limits (CPU/memory) when running many external processes.
*   **Aggregation:** Aggregating statistics (like type counts/sizes, month counts) needs safe concurrent access. Use mutexes (`sync.Mutex`) or atomic operations (`sync/atomic`) or channel-based aggregation patterns.

## 7. External Tools Integration (`pkg/tools`)

*   The `ToolRunner` interface abstracts the execution.
*   `DefaultToolRunner` uses `os/exec` to run commands.
*   Implementations for specific tools (`MagikaRunner`, `ExiftoolRunner`, `JdupesRunner`) might wrap `DefaultToolRunner` to provide more structured input/output parsing (e.g., parsing `magika --jsonl` output).
*   Use `ToolRunner.CheckAvailable` to determine if tools exist and adapt analysis accordingly (e.g., fallback from `jdupes` to MD5 hashing). Report tool availability in the final report.

## 8. Configuration (`pkg/config`)

*   `Config` struct holds settings like `DownloadsDir`, `OutputFile`, `LogFile`, `Verbose`, `SamplePerDir`, `LargeFileThresholdMB`, `RecentDays`.
*   Use `spf13/viper` or standard `flag` package to load from command-line arguments. Consider supporting a config file (e.g., `~/.config/desktop-organizer/config.yaml`) in the future.

## 9. Logging (`pkg/log`)

*   Use standard `log/slog`.
*   Configure level (Debug, Info) based on `Config.Verbose`.
*   Configure output to file (`Config.LogFile`) and/or stderr.
*   Pass `*slog.Logger` instance via context or direct injection where needed.

## 10. Testing

*   **Unit Tests:** Test individual analysis modules (`pkg/analysis/*`) by mocking dependencies (especially `FileSystem` and `ToolRunner`). Test helper functions and data structure logic.
*   **Integration Tests:** Test the interaction between components. For example, test the `analyzer.Analyze` flow with mock implementations that return predefined data. Test `ToolRunner` implementations by actually running the tools (if installed) on sample data, or mocking `os/exec`. Test `FileSystem` against a temporary test directory.