I'll create a Go architecture design document for rewriting the downloads folder inspector script in Go.

# Downloads Folder Inspector - Go Architecture Design

## 1. System Overview

The Downloads Folder Inspector is a tool that analyzes the contents of a downloads folder to provide insights about file types, sizes, duplicates, and organization recommendations. This document outlines the Go architecture for reimplementing the bash script.

## 2. Core Components

### 2.1 Domain Models

```go
// File represents a file in the downloads directory
type File struct {
    Path         string
    Size         int64
    ModTime      time.Time
    FileType     string
    Group        string // Media group (image, video, audio, etc.)
    ContentHash  string // For deduplication
}

// FileSet represents a collection of duplicate files
type FileSet struct {
    Files       []File
    ContentHash string
    WastedSpace int64
}

// AnalysisResult holds the complete analysis output
type AnalysisResult struct {
    BasicStats      BasicStatistics
    FileTypeStats   map[string]TypeStatistics
    DuplicateSets   []FileSet
    RecentFiles     []File
    LargeFiles      []File
    MediaFiles      []File
    MonthlyStats    map[string]MonthStatistics
    Recommendations []string
}

// Various statistics structs
type BasicStatistics struct {
    TotalItems       int
    TotalFiles       int
    TotalDirectories int
    TotalSize        int64
}

type TypeStatistics struct {
    Count     int
    TotalSize int64
}

type MonthStatistics struct {
    Count     int
    TotalSize int64
}
```

### 2.2 Core Interfaces

```go
// FileScanner scans the filesystem and returns file information
type FileScanner interface {
    // ScanFiles scans the given directory and returns file information
    ScanFiles(ctx context.Context, dir string, recursive bool) ([]File, error)
    
    // ScanTopLevel scans only the top level of the directory
    ScanTopLevel(ctx context.Context, dir string) ([]File, []string, error)
}

// FileTypeDetector determines file types
type FileTypeDetector interface {
    // DetectType analyzes a file and returns its type and group
    DetectType(ctx context.Context, path string) (fileType string, group string, err error)
    
    // BatchDetectTypes analyzes multiple files in one operation
    BatchDetectTypes(ctx context.Context, files []string) (map[string]struct{Type, Group string}, error)
    
    // SampleAndDetect samples files from directories and detects types
    SampleAndDetect(ctx context.Context, files []File, samplesPerDir int) (map[string]struct{Type, Group string}, error)
}

// MetadataExtractor extracts metadata from media files
type MetadataExtractor interface {
    // ExtractMetadata extracts metadata from media files
    ExtractMetadata(ctx context.Context, file File) (map[string]string, error)
}

// DuplicateFinder finds duplicate files
type DuplicateFinder interface {
    // FindDuplicates returns sets of duplicate files
    FindDuplicates(ctx context.Context, files []File) ([]FileSet, error)
}

// Reporter generates analysis reports
type Reporter interface {
    // GenerateReport creates a report from the analysis results
    GenerateReport(result AnalysisResult) (string, error)
}
```

### 2.3 Service Interfaces

```go
// DownloadAnalyzer runs the complete analysis process
type DownloadAnalyzer interface {
    // Analyze performs a complete analysis of the downloads directory
    Analyze(ctx context.Context, dir string, options AnalysisOptions) (AnalysisResult, error)
}

// AnalysisOptions holds configuration options for the analysis
type AnalysisOptions struct {
    Recursive      bool
    Verbose        bool
    SamplesPerDir  int
    RecentDays     int
    LargeFileSize  int64 // In bytes
}
```

## 3. Implementation Details

### 3.1 File Scanning

```go
// FileSystemScanner implements FileScanner using the standard library
type FileSystemScanner struct {
    logger Logger
}

// Each component will have a logger dependency for consistent logging
type Logger interface {
    Debug(msg string, args ...interface{})
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
}

// Implementation will use filepath.Walk or filepath.WalkDir for scanning
```

### 3.2 Type Detection

```go
// MagikaDetector implements FileTypeDetector using Magika
type MagikaDetector struct {
    logger Logger
    // Configuration options for calling Magika
}

// FallbackDetector uses file extension and mime types as fallback
type FallbackDetector struct {
    logger Logger
}

// CompositeDetector tries different detectors in sequence
type CompositeDetector struct {
    primary   FileTypeDetector
    secondary FileTypeDetector
    logger    Logger
}
```

### 3.3 External Tool Integration

```go
// ExternalToolExecutor executes external tools like Magika, ExifTool
type ExternalToolExecutor interface {
    // Execute runs an external command and returns the output
    Execute(ctx context.Context, command string, args ...string) ([]byte, error)
    
    // CheckAvailability checks if a tool is available
    CheckAvailability(tool string) bool
}

// Implementation for each external tool
type MagikaExecutor struct {
    executor ExternalToolExecutor
    logger   Logger
}

type ExifToolExecutor struct {
    executor ExternalToolExecutor
    logger   Logger
}

type JDupesExecutor struct {
    executor ExternalToolExecutor
    logger   Logger
}
```

## 4. Concurrency Model

The Go implementation will leverage concurrency for better performance:

```go
// Worker pool pattern for file processing
type FileProcessor struct {
    workerCount int
    logger      Logger
}

func (p *FileProcessor) ProcessFiles(ctx context.Context, files []File, processor func(File) error) error {
    // Implementation using goroutines, channels, and error groups
}
```

## 5. Progress Reporting

```go
// ProgressReporter provides feedback during long-running operations
type ProgressReporter interface {
    // Start begins a new progress tracking operation
    Start(total int, description string)
    
    // Update updates the progress
    Update(current int)
    
    // Finish completes the progress tracking
    Finish()
}

// Console implementation
type ConsoleProgressReporter struct {
    logger Logger
}
```

## 6. Complete System Architecture

```
┌─────────────────────────┐
│   DownloadAnalyzerApp   │
└───────────┬─────────────┘
            │
            ▼
┌─────────────────────────┐     ┌─────────────────────┐
│   DownloadAnalyzer      │────▶│   FileScanner       │
└───────────┬─────────────┘     └─────────────────────┘
            │
            ├──────────────┐    ┌─────────────────────┐
            │              ├───▶│   FileTypeDetector  │
            │              │    └─────────────────────┘
            │              │
            │              │    ┌─────────────────────┐
            │              ├───▶│   MetadataExtractor │
            │              │    └─────────────────────┘
            │              │
            │              │    ┌─────────────────────┐
            │              ├───▶│   DuplicateFinder   │
            ▼              │    └─────────────────────┘
┌─────────────────────────┐│    
│   AnalysisAggregator    ││    ┌─────────────────────┐
└───────────┬─────────────┘├───▶│   ProgressReporter  │
            │               │   └─────────────────────┘
            │               │
            │               │   ┌─────────────────────┐
            │               └──▶│   ExternalToolInteg.│
            ▼                   └─────────────────────┘
┌─────────────────────────┐
│       Reporter          │
└─────────────────────────┘
```

## 7. Package Structure

```
downloads-inspector/
├── cmd/                      # Command-line application
│   └── inspector/            # Main application entry point
├── internal/                 # Private application code
│   ├── analyzer/             # Core analysis logic
│   ├── domain/               # Domain models
│   ├── detection/            # File type detection
│   │   ├── magika/           # Magika integration
│   │   └── fallback/         # Fallback detection
│   ├── scanner/              # File scanning
│   ├── metadata/             # Metadata extraction
│   │   └── exiftool/         # ExifTool integration
│   ├── duplicates/           # Duplicate finding
│   │   ├── jdupes/           # JDupes integration
│   │   └── hash/             # Hash-based deduplication
│   ├── progress/             # Progress reporting
│   ├── external/             # External tool integration
│   └── report/               # Report generation
└── pkg/                      # Public packages for reuse
    ├── logger/               # Logging utilities
    └── tools/                # Tool helpers
```

## 8. Configuration Management

```go
// Config holds the application configuration
type Config struct {
    DownloadsDir    string
    OutputFile      string
    DebugLog        string
    Verbose         bool
    SamplesPerDir   int
    RecentDaysLimit int
    LargeFileSizeThreshold int64
}

// ConfigProvider loads configuration from various sources
type ConfigProvider interface {
    // LoadConfig loads the configuration from various sources
    LoadConfig() (Config, error)
}

// Implementations for file-based, environment and command-line
```

## 9. Error Handling Strategy

```go
// Custom error types for domain-specific errors
type FileAccessError struct {
    Path string
    Err  error
}

type ExternalToolError struct {
    Tool string
    Args []string
    Err  error
}

// Error handling middleware for graceful recovery
func withRecovery(logger Logger, fn func() error) func() error {
    return func() error {
        // Implementation with defer/recover
    }
}
```

## 10. Main Application Flow

```go
func main() {
    // 1. Parse command-line flags
    // 2. Initialize logger
    // 3. Load configuration
    // 4. Create dependencies
    // 5. Create and configure analyzer
    // 6. Run analysis
    // 7. Generate and save report
    // 8. Display summary
}
```

## 11. Testing Strategy

- Unit tests for each component
- Integration tests for external tool integration
- Mock implementations of interfaces for testing
- Test fixtures with sample download directories
- Benchmarks for performance-critical operations

## 12. Future Extensions

- Web UI for interactive exploration of results
- Scheduled analysis and comparison over time
- File organization suggestions with automation
- Integration with cloud storage analysis
- Machine learning for smarter file categorization
