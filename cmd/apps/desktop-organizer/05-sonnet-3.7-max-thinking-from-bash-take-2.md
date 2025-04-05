
# Go Architecture for Downloads Folder Analyzer

## Overview

This document outlines the architecture for rewriting the `01-inspect-downloads-folder.sh` bash script in Go, providing a structured, concurrent, and maintainable design focused on analyzing the contents of a Downloads folder.

## Core Components

```
downloads-analyzer/
├── cmd/
│   └── analyzer/                 # Main entry point
├── internal/
│   ├── config/                   # Configuration handling
│   ├── analyzer/                 # Core analyzer logic
│   │   ├── collector/            # File data collection
│   │   ├── processor/            # Data processing
│   │   └── reporter/             # Output generation
│   ├── domain/                   # Domain models
│   ├── integrations/             # External tool integrations
│   └── utils/                    # Utility functions
└── pkg/                          # Reusable packages
```

## Domain Models

```go
// FileInfo represents core information about a file
type FileInfo struct {
    Path         string
    Size         int64
    ModTime      time.Time
    Type         string
    Group        string
    Hash         string
    MetadataMap  map[string]interface{}
}

// AnalysisResult contains aggregated analysis data
type AnalysisResult struct {
    BasicStats       *BasicStats
    FileTypes        map[string]*TypeInfo
    DuplicateSets    []*DuplicateSet
    RecentFiles      []*FileInfo
    LargeFiles       []*FileInfo
    FilesByMonth     map[string]*MonthStats
    MediaFiles       []*MediaFileInfo
    Recommendations  []string
}
```

## Key Interfaces

```go
// FileCollector gathers file information
type FileCollector interface {
    CollectFiles(ctx context.Context, dir string, options *CollectionOptions) (<-chan *FileInfo, <-chan error)
}

// FileProcessor processes collected file data
type FileProcessor interface {
    ProcessFiles(ctx context.Context, files <-chan *FileInfo) (*AnalysisResult, error)
}

// TypeDetector detects file types
type TypeDetector interface {
    DetectType(ctx context.Context, file *FileInfo) error
}

// DuplicateDetector finds duplicate files
type DuplicateDetector interface {
    FindDuplicates(ctx context.Context, files []*FileInfo) ([]*DuplicateSet, error)
}

// MetadataExtractor extracts metadata from files
type MetadataExtractor interface {
    ExtractMetadata(ctx context.Context, file *FileInfo) error
}

// Reporter generates reports from analysis results
type Reporter interface {
    GenerateReport(ctx context.Context, result *AnalysisResult, options *ReportOptions) error
}
```

## Implementations

### File Collectors

```go
// DefaultFileCollector implements FileCollector with concurrent file scanning
type DefaultFileCollector struct {
    maxConcurrency int
    logger         Logger
}

// SamplingFileCollector implements FileCollector with per-directory sampling
type SamplingFileCollector struct {
    maxPerDir int
    collector FileCollector
}
```

### Type Detectors

```go
// MagikaTypeDetector implements TypeDetector using Magika
type MagikaTypeDetector struct {
    binaryPath string
    timeout    time.Duration
}

// FallbackTypeDetector implements TypeDetector using "file" command
type FallbackTypeDetector struct {
    timeout time.Duration
}

// CompositeTypeDetector implements TypeDetector with fallback capability
type CompositeTypeDetector struct {
    primary   TypeDetector
    secondary TypeDetector
}
```

### Duplicate Detectors

```go
// JDupesDetector implements DuplicateDetector using jdupes
type JDupesDetector struct {
    binaryPath string
    timeout    time.Duration
}

// MD5Detector implements DuplicateDetector using MD5 hashing
type MD5Detector struct {
    maxConcurrency int
}

// CompositeDuplicateDetector implements DuplicateDetector with fallback
type CompositeDuplicateDetector struct {
    primary   DuplicateDetector
    secondary DuplicateDetector
}
```

### Metadata Extractors

```go
// ExifToolExtractor implements MetadataExtractor using ExifTool
type ExifToolExtractor struct {
    binaryPath string
    timeout    time.Duration
}

// BasicExtractor implements MetadataExtractor with basic Go stdlib capabilities
type BasicExtractor struct{}
```

### Reporters

```go
// TextReporter implements Reporter to generate text reports
type TextReporter struct {
    writer io.Writer
}

// JSONReporter implements Reporter to generate JSON reports
type JSONReporter struct {
    writer io.Writer
}

// CSVReporter implements Reporter to generate CSV reports
type CSVReporter struct {
    writer io.Writer
}
```

## Concurrency Model

### Worker Pool Pattern

```go
// WorkerPool manages a pool of worker goroutines
type WorkerPool struct {
    size      int
    taskChan  chan func()
    wg        sync.WaitGroup
    ctx       context.Context
    cancelFn  func()
}

// FileProcessor uses worker pool for concurrent processing
func (p *DefaultFileProcessor) ProcessFiles(ctx context.Context, files <-chan *FileInfo) (*AnalysisResult, error) {
    // Create worker pool
    pool := NewWorkerPool(p.maxConcurrency)
    
    // Process files concurrently
    for file := range files {
        file := file // Capture for closure
        pool.Submit(func() {
            p.typeDetector.DetectType(ctx, file)
            p.metadataExtractor.ExtractMetadata(ctx, file)
            // More processing...
        })
    }
    
    // Wait for completion
    if err := pool.Wait(); err != nil {
        return nil, err
    }
    
    // Aggregate results
    // ...
}
```

### Rate Limiting

```go
// RateLimiter limits the rate of operations
type RateLimiter struct {
    limiter *rate.Limiter
}

// For external tool calls to prevent overloading
func (d *JDupesDetector) FindDuplicates(ctx context.Context, files []*FileInfo) ([]*DuplicateSet, error) {
    // Rate limit external process creation
    if err := d.rateLimiter.Wait(ctx); err != nil {
        return nil, err
    }
    
    // Execute jdupes
    // ...
}
```

## Error Handling

```go
// ErrCollectionFailed represents a collection failure
type ErrCollectionFailed struct {
    Path string
    Err  error
}

func (e *ErrCollectionFailed) Error() string {
    return fmt.Sprintf("failed to collect files from %s: %v", e.Path, e.Err)
}

func (e *ErrCollectionFailed) Unwrap() error {
    return e.Err
}

// ErrorGroup for managing multiple concurrent errors
func (c *DefaultFileCollector) CollectFiles(ctx context.Context, dir string, options *CollectionOptions) (<-chan *FileInfo, <-chan error) {
    filesChan := make(chan *FileInfo)
    errChan := make(chan error, 1)
    
    g, ctx := errgroup.WithContext(ctx)
    
    // Start collection
    g.Go(func() error {
        defer close(filesChan)
        // Collection logic...
        return nil
    })
    
    // Monitor errors
    go func() {
        defer close(errChan)
        if err := g.Wait(); err != nil {
            errChan <- err
        }
    }()
    
    return filesChan, errChan
}
```

## Configuration Handling

```go
// Config represents the application configuration
type Config struct {
    DownloadsDir     string
    OutputFile       string
    DebugLogFile     string
    Verbose          bool
    SamplePerDir     int
    UseExifTool      bool
    UseMagika        bool
    UseJDupes        bool
    LargeFileThreshold int64
    RecentDaysThreshold int
    MaxConcurrency   int
}

// ConfigFromFlags creates Config from command-line flags
func ConfigFromFlags() (*Config, error) {
    // Flag parsing logic
    // ...
    return config, nil
}

// ConfigFromFile creates Config from a config file
func ConfigFromFile(path string) (*Config, error) {
    // File parsing logic
    // ...
    return config, nil
}
```

## Dependency Injection

```go
// Analyzer is the main application facade
type Analyzer struct {
    collector   FileCollector
    processor   FileProcessor
    reporter    Reporter
    config      *Config
    logger      Logger
}

// NewAnalyzer creates a new Analyzer with dependencies
func NewAnalyzer(config *Config) (*Analyzer, error) {
    // Create type detector based on availability
    var typeDetector TypeDetector
    if config.UseMagika && isMagikaAvailable() {
        typeDetector = NewMagikaTypeDetector(config)
    } else {
        typeDetector = NewFallbackTypeDetector(config)
    }
    
    // Create duplicate detector based on availability
    var dupDetector DuplicateDetector
    if config.UseJDupes && isJDupesAvailable() {
        dupDetector = NewJDupesDetector(config)
    } else {
        dupDetector = NewMD5Detector(config)
    }
    
    // Create metadata extractor based on availability
    var metaExtractor MetadataExtractor
    if config.UseExifTool && isExifToolAvailable() {
        metaExtractor = NewExifToolExtractor(config)
    } else {
        metaExtractor = NewBasicExtractor()
    }
    
    // Create processor with dependencies
    processor := NewFileProcessor(typeDetector, dupDetector, metaExtractor, config)
    
    // Create collector
    collector := NewFileCollector(config)
    if config.SamplePerDir > 0 {
        collector = NewSamplingFileCollector(collector, config.SamplePerDir)
    }
    
    // Create reporter
    reporter := NewTextReporter(config.OutputFile)
    
    return &Analyzer{
        collector: collector,
        processor: processor,
        reporter:  reporter,
        config:    config,
        logger:    NewLogger(config.DebugLogFile, config.Verbose),
    }, nil
}
```

## Main Execution Flow

```go
// Main execution flow
func (a *Analyzer) Run(ctx context.Context) error {
    // Setup context with cancellation
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()
    
    // Start file collection
    filesChan, errChan := a.collector.CollectFiles(ctx, a.config.DownloadsDir, &CollectionOptions{
        SkipHidden: true,
        ExcludePaths: []string{"**/__MACOSX/**"},
    })
    
    // Process collected files
    result, err := a.processor.ProcessFiles(ctx, filesChan)
    if err != nil {
        return fmt.Errorf("processing failed: %w", err)
    }
    
    // Generate report
    if err := a.reporter.GenerateReport(ctx, result, &ReportOptions{
        IncludeRecommendations: true,
        ShowLargeFilesLimit:    10,
        ShowRecentFilesLimit:   30,
    }); err != nil {
        return fmt.Errorf("report generation failed: %w", err)
    }
    
    // Check for collection errors
    for err := range errChan {
        a.logger.Error("Collection error", "error", err)
    }
    
    return nil
}
```

## External Tool Integration

```go
// ExternalCommand represents a command to an external tool
type ExternalCommand struct {
    Path      string
    Args      []string
    Timeout   time.Duration
    InputFile string
}

// ExecuteCommand executes an external command and returns its output
func ExecuteCommand(ctx context.Context, cmd *ExternalCommand) ([]byte, error) {
    // Create command with timeout
    ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
    defer cancel()
    
    // Create command
    execCmd := exec.CommandContext(ctx, cmd.Path, cmd.Args...)
    
    // Setup I/O
    if cmd.InputFile != "" {
        file, err := os.Open(cmd.InputFile)
        if err != nil {
            return nil, err
        }
        defer file.Close()
        execCmd.Stdin = file
    }
    
    // Execute and return output
    return execCmd.CombinedOutput()
}

// Example usage in MagikaTypeDetector
func (m *MagikaTypeDetector) DetectType(ctx context.Context, file *FileInfo) error {
    output, err := ExecuteCommand(ctx, &ExternalCommand{
        Path:    m.binaryPath,
        Args:    []string{"--json", file.Path},
        Timeout: m.timeout,
    })
    if err != nil {
        return fmt.Errorf("magika execution failed: %w", err)
    }
    
    // Parse JSON output
    var result magikaOutput
    if err := json.Unmarshal(output, &result); err != nil {
        return fmt.Errorf("failed to parse magika output: %w", err)
    }
    
    // Update file info
    file.Type = result.Result.Value.Output.CtLabel
    file.Group = result.Result.Value.Output.Group
    
    return nil
}
```

## Logger Interface

```go
// Logger provides structured logging
type Logger interface {
    Debug(msg string, keysAndValues ...interface{})
    Info(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
}

// DefaultLogger implements Logger
type DefaultLogger struct {
    logger *zap.SugaredLogger
}

func NewLogger(debugFile string, verbose bool) (Logger, error) {
    // Logger configuration
    config := zap.NewProductionConfig()
    if verbose {
        config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
    } else {
        config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
    }
    
    // Output paths
    config.OutputPaths = []string{"stderr"}
    if debugFile != "" {
        config.OutputPaths = append(config.OutputPaths, debugFile)
    }
    
    // Create logger
    logger, err := config.Build()
    if err != nil {
        return nil, err
    }
    
    return &DefaultLogger{
        logger: logger.Sugar(),
    }, nil
}
```

## Testing Strategy

```go
// Mock implementations for testing
type MockFileCollector struct {
    files []*FileInfo
    err   error
}

func (m *MockFileCollector) CollectFiles(ctx context.Context, dir string, options *CollectionOptions) (<-chan *FileInfo, <-chan error) {
    filesChan := make(chan *FileInfo)
    errChan := make(chan error, 1)
    
    go func() {
        defer close(filesChan)
        for _, file := range m.files {
            select {
            case <-ctx.Done():
                return
            case filesChan <- file:
            }
        }
    }()
    
    go func() {
        defer close(errChan)
        if m.err != nil {
            errChan <- m.err
        }
    }()
    
    return filesChan, errChan
}

// Example test for Analyzer
func TestAnalyzerRun(t *testing.T) {
    // Setup mock dependencies
    mockCollector := &MockFileCollector{
        files: []*FileInfo{
            {Path: "/downloads/file1.txt", Size: 1024, ModTime: time.Now()},
            {Path: "/downloads/file2.jpg", Size: 2048, ModTime: time.Now().Add(-24 * time.Hour)},
        },
    }
    
    mockProcessor := &MockFileProcessor{
        result: &AnalysisResult{
            // Test data...
        },
    }
    
    mockReporter := &MockReporter{}
    
    // Create analyzer with mocks
    analyzer := &Analyzer{
        collector: mockCollector,
        processor: mockProcessor,
        reporter:  mockReporter,
        config:    &Config{DownloadsDir: "/downloads"},
        logger:    &MockLogger{},
    }
    
    // Run analyzer
    err := analyzer.Run(context.Background())
    
    // Assert expectations
    require.NoError(t, err)
    require.True(t, mockProcessor.called)
    require.True(t, mockReporter.called)
    // More assertions...
}
```

## Command-Line Application

```go
// Example main function using Cobra
func main() {
    // Root command
    rootCmd := &cobra.Command{
        Use:   "downloads-analyzer",
        Short: "Analyze the contents of the Downloads folder",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Load config from flags
            config, err := ConfigFromFlags()
            if err != nil {
                return err
            }
            
            // Create analyzer
            analyzer, err := NewAnalyzer(config)
            if err != nil {
                return err
            }
            
            // Run analyzer
            return analyzer.Run(cmd.Context())
        },
    }
    
    // Add flags
    rootCmd.Flags().StringP("dir", "d", defaultDownloadsDir(), "Downloads directory path")
    rootCmd.Flags().StringP("output", "o", "downloads_analysis.txt", "Output file path")
    rootCmd.Flags().StringP("log", "l", "downloads_analysis_debug.log", "Debug log file path")
    rootCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
    rootCmd.Flags().IntP("sample", "s", 0, "Maximum files to sample per directory (0 for all)")
    rootCmd.Flags().Int("concurrency", runtime.NumCPU(), "Maximum concurrent operations")
    // More flags...
    
    // Execute
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```
