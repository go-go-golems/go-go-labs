I'll create a Go architecture for rewriting the Downloads folder inspection script. Here's the markdown design:

# Downloads Folder Analyzer - Go Architecture

## Overview

This document outlines the architecture for rewriting the bash script `01-inspect-downloads-folder.sh` in Go, utilizing Go's strengths in concurrent processing, strong typing, and modular design.

## Core Components

### 1. Domain Models

```go
// File represents a file in the Downloads directory
type File struct {
    Path          string
    Size          int64
    ModTime       time.Time
    Type          string
    Group         string  // Media type group (image, video, audio, etc.)
    MetadataCache map[string]interface{} // For storing extracted metadata
}

// DuplicateSet represents a set of duplicate files
type DuplicateSet struct {
    Files      []File
    Size       int64  // Size of each file in the set
    WastedSize int64  // (n-1) * Size
}

// AnalysisResult contains all analysis data
type AnalysisResult struct {
    TotalItems         int
    TotalSize          int64
    FilesByType        map[string][]File
    LargeFiles         []File
    RecentFiles        []File
    FilesByMonth       map[string][]File
    DuplicateSets      []DuplicateSet
    MediaFiles         []File
    AnalysisTimestamp  time.Time
}
```

### 2. Core Interfaces

```go
// FileCollector gathers files from the target directory
type FileCollector interface {
    CollectFiles(ctx context.Context, dir string, options CollectorOptions) ([]File, error)
}

// FileTypeDetector determines file types
type FileTypeDetector interface {
    DetectType(ctx context.Context, file File) (string, string, error) // Returns type, group, error
}

// FileMetadataExtractor extracts metadata from files
type FileMetadataExtractor interface {
    ExtractMetadata(ctx context.Context, file File) (map[string]interface{}, error)
}

// DuplicateFinder finds duplicate files
type DuplicateFinder interface {
    FindDuplicates(ctx context.Context, files []File) ([]DuplicateSet, error)
}

// Analyzer performs specific analysis on files
type Analyzer interface {
    Analyze(ctx context.Context, files []File) (interface{}, error)
}

// Reporter generates reports from analysis results
type Reporter interface {
    GenerateReport(ctx context.Context, result AnalysisResult) (string, error)
}
```

### 3. Configuration

```go
// Config holds application configuration
type Config struct {
    DownloadsDir      string
    OutputFile        string
    LogFile           string
    Verbose           bool
    SamplePerDir      int
    LargeFileThreshold int64  // Bytes
    RecentDaysThreshold int   // Days
}
```

## Implementation Components

### 1. File Collectors

```go
// DefaultFileCollector implements FileCollector
type DefaultFileCollector struct {
    Logger Logger
}

// FilteredFileCollector wraps another collector and applies filters
type FilteredFileCollector struct {
    BaseCollector FileCollector
    ExcludePatterns []string  // e.g., "*/__MACOSX/*"
    SamplePerDir   int        // Number of files to sample per directory
}
```

### 2. File Type Detectors

```go
// MagikaFileTypeDetector uses Google's Magika for AI-powered type detection
type MagikaFileTypeDetector struct {
    Logger Logger
}

// FallbackFileTypeDetector uses mimetype or other Go libraries
type FallbackFileTypeDetector struct {
    Logger Logger
}

// ChainedFileTypeDetector tries multiple detectors in sequence
type ChainedFileTypeDetector struct {
    Detectors []FileTypeDetector
}
```

### 3. Metadata Extractors

```go
// ExifToolMetadataExtractor uses ExifTool command-line wrapper
type ExifToolMetadataExtractor struct {
    Logger Logger
}

// MediaInfoExtractor extracts media-specific metadata
type MediaInfoExtractor struct {
    Logger Logger
}
```

### 4. Duplicate Finders

```go
// HashBasedDuplicateFinder uses file hashing to find duplicates
type HashBasedDuplicateFinder struct {
    Logger Logger
    HashAlgorithm string  // "md5", "sha256", etc.
}

// JDupesDuplicateFinder wraps the jdupes command-line tool
type JDupesDuplicateFinder struct {
    Logger Logger
}
```

### 5. Analyzers

```go
// FileTypeAnalyzer analyzes file type distribution
type FileTypeAnalyzer struct{}

// LargeFileAnalyzer finds large files
type LargeFileAnalyzer struct {
    Threshold int64  // Size in bytes
}

// RecentFileAnalyzer finds recently modified files
type RecentFileAnalyzer struct {
    Days int  // Number of days to consider "recent"
}

// FilesByMonthAnalyzer groups files by modification month
type FilesByMonthAnalyzer struct{}

// MediaFileAnalyzer analyzes media files
type MediaFileAnalyzer struct{
    MetadataExtractor FileMetadataExtractor
}
```

### 6. Reporters

```go
// TextReporter generates plain text reports
type TextReporter struct{}

// JSONReporter generates JSON reports
type JSONReporter struct{}

// MarkdownReporter generates Markdown reports
type MarkdownReporter struct{}
```

### 7. Logging and Progress

```go
// Logger provides logging functionality
type Logger interface {
    Debug(msg string, args ...interface{})
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
}

// ProgressReporter reports progress during long operations
type ProgressReporter interface {
    Start(total int, operation string)
    Update(current int)
    Complete()
}
```

## Execution Flow

### 1. Main Application

```go
// App is the main application
type App struct {
    Config            Config
    Logger            Logger
    FileCollector     FileCollector
    TypeDetector      FileTypeDetector
    DuplicateFinder   DuplicateFinder
    Analyzers         map[string]Analyzer
    Reporter          Reporter
    ProgressReporter  ProgressReporter
}

// Run executes the full analysis process
func (a *App) Run(ctx context.Context) error {
    // 1. Collect files
    // 2. Detect types
    // 3. Run analyzers
    // 4. Find duplicates
    // 5. Generate report
    // 6. Return results
}
```

### 2. Concurrency Management

```go
// Worker processes files in parallel
type Worker struct {
    WorkerID   int
    InputChan  <-chan File
    OutputChan chan<- File
    ErrChan    chan<- error
    Processor  func(File) (File, error)
}

// WorkerPool manages multiple workers
type WorkerPool struct {
    Workers    []*Worker
    NumWorkers int
}
```

## Command Line Interface

```go
// CLI represents the command-line interface
type CLI struct {
    App    *App
    Config Config
}

// rootCmd is the main cobra command
var rootCmd = &cobra.Command{
    Use:   "downloads-analyzer",
    Short: "Analyze the contents of the Downloads folder",
    Long:  `A comprehensive tool to analyze Downloads folder content, find duplicates, and provide organization recommendations.`,
    Run: func(cmd *cobra.Command, args []string) {
        // Initialize and run the application
    },
}
```

## Implementation Strategy

1. **Initial Setup**:
   - Define core interfaces and models
   - Implement file collection with parallelism
   - Set up CLI with cobra

2. **Core Features**:
   - Implement file type detection (integrate Magika if available)
   - Implement basic file analysis (size, date, etc.)
   - Implement duplicate detection

3. **Advanced Features**:
   - Implement media file metadata extraction
   - Implement reporting with multiple formats
   - Add visualization capabilities

4. **Performance Optimizations**:
   - Use worker pools for parallel processing
   - Implement sampling for large directories
   - Add caching mechanisms for file metadata

## External Dependencies

- **cobra**: Command-line interface
- **magika-go**: Go bindings for Google's Magika (if available)
- **go-exiftool**: Go wrapper for ExifTool
- **errgroup**: Error handling for parallel operations
- **zap**: Structured logging
- **progressbar**: Terminal progress bar

## Testing Strategy

1. **Unit Tests**:
   - Test individual analyzers with mock file data
   - Test file collectors with test directories
   - Test duplicate detection with known duplicates

2. **Integration Tests**:
   - Test the full analysis pipeline with sample data
   - Test report generation with known results

3. **Benchmark Tests**:
   - Test performance with large directories
   - Compare performance with the original bash script
