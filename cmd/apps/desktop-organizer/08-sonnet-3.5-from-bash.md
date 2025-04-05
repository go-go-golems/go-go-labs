I'll help you design a Go architecture to rewrite the downloads folder inspection script. Let me analyze the script and create a comprehensive design document.

# Downloads Folder Inspector - Go Architecture Design

## 1. Core Domain Models

### 1.1 File Information Model
```go
type FileInfo struct {
    Path          string
    RelativePath  string
    Size          int64
    ModTime       time.Time
    Type          FileType
    Group         FileGroup
    Metadata      map[string]interface{}
    Hash          string // For deduplication
}

type FileType string
type FileGroup string // image, video, audio, document, etc.
```

### 1.2 Analysis Results Models
```go
type AnalysisResults struct {
    BasicStats      BasicStatistics
    TypeStats       TypeStatistics
    DuplicateStats  DuplicateStatistics
    MediaStats      MediaStatistics
    TimelineStats   TimelineStatistics
    LargeFiles      []FileInfo
    RecentFiles     []FileInfo
    Recommendations []string
}

type BasicStatistics struct {
    TotalItems      int
    TotalFiles      int
    TotalDirs       int
    TotalSize       int64
    TopLevelStats   BasicStatistics // For top-level only stats
}

type TypeStatistics struct {
    TypeCounts map[FileType]int
    TypeSizes  map[FileType]int64
    Groups     map[FileGroup][]FileInfo
}

type DuplicateStatistics struct {
    DuplicateSets     int
    TotalDuplicates   int
    WastedSpace       int64
    DuplicateGroups   [][]FileInfo
}

type MediaStatistics struct {
    TopMediaFiles []MediaFileInfo
}

type TimelineStatistics struct {
    FilesByMonth map[string]MonthlyStats
}

type MonthlyStats struct {
    Count int
    Size  int64
    Files []FileInfo
}
```

## 2. Core Interfaces

### 2.1 File Type Detection Interface
```go
type FileTypeDetector interface {
    Detect(ctx context.Context, path string) (FileType, FileGroup, error)
    DetectBatch(ctx context.Context, paths []string) (map[string]FileType, error)
}

// Implementations:
// - MagikaDetector: Uses Google's Magika
// - FallbackDetector: Uses Go's built-in mime package
```

### 2.2 Metadata Extractor Interface
```go
type MetadataExtractor interface {
    Extract(ctx context.Context, path string) (map[string]interface{}, error)
    ExtractMedia(ctx context.Context, path string) (*MediaFileInfo, error)
}

// Implementations:
// - ExifToolExtractor: Uses ExifTool
// - BasicExtractor: Uses built-in Go image/metadata packages
```

### 2.3 Deduplication Interface
```go
type Deduplicator interface {
    FindDuplicates(ctx context.Context, paths []string) (DuplicateStatistics, error)
    GenerateFileHash(path string) (string, error)
}

// Implementations:
// - JdupesDeduplicator: Wraps jdupes command
// - MD5Deduplicator: Uses MD5 hashing
```

### 2.4 Progress Reporter Interface
```go
type ProgressReporter interface {
    Start(total int)
    Update(current int, message string)
    Complete()
    Debug(message string)
    VerboseDebug(message string)
}
```

## 3. Services Layer

### 3.1 Scanner Service
```go
type ScannerService interface {
    ScanDirectory(ctx context.Context, path string, options ScanOptions) (*AnalysisResults, error)
}

type ScanOptions struct {
    MaxDepth        int
    SamplesPerDir   int
    IncludeHidden   bool
    MinFileSize     int64
    MaxFileSize     int64
    FileTypes       []FileType
    Since           time.Time
    Until           time.Time
}
```

### 3.2 Analysis Service
```go
type AnalyzerService interface {
    AnalyzeTypes(ctx context.Context, files []FileInfo) TypeStatistics
    AnalyzeMedia(ctx context.Context, files []FileInfo) MediaStatistics
    AnalyzeDuplicates(ctx context.Context, files []FileInfo) DuplicateStatistics
    AnalyzeTimeline(files []FileInfo) TimelineStatistics
    GenerateRecommendations(results *AnalysisResults) []string
}
```

### 3.3 Report Generation Service
```go
type ReportGenerator interface {
    Generate(results *AnalysisResults) (string, error)
    GenerateJSON(results *AnalysisResults) ([]byte, error)
    GenerateYAML(results *AnalysisResults) ([]byte, error)
}
```

## 4. Application Layer

### 4.1 Command Line Interface
```go
type CLI struct {
    scanner    ScannerService
    analyzer   AnalyzerService
    reporter   ReportGenerator
    progress   ProgressReporter
    options    ScanOptions
}
```

### 4.2 Configuration
```go
type Config struct {
    DefaultScanOptions ScanOptions
    OutputFormat      string
    DebugLog         bool
    Verbose          bool
    Workers          int
}
```

## 5. Implementation Patterns

1. **Worker Pool Pattern**
   - For parallel processing of file analysis
   - Controlled concurrency for I/O operations

2. **Pipeline Pattern**
   - Staged processing: Scan → Analyze → Report
   - Each stage can be parallelized independently

3. **Options Pattern**
   - For configurable behavior
   - Allows extension without breaking existing code

4. **Factory Pattern**
   - For creating appropriate implementations based on available tools
   - Example: FileTypeDetector factory choosing between Magika and fallback

5. **Observer Pattern**
   - For progress reporting and logging
   - Allows multiple subscribers to progress updates

## 6. Error Handling Strategy

1. **Custom Error Types**
```go
type ScanError struct {
    Path string
    Err  error
}

type AnalysisError struct {
    Stage string
    Err   error
}
```

2. **Error Aggregation**
- Collect non-fatal errors during scanning/analysis
- Continue processing when possible
- Report all errors in final results

## 7. Concurrency Strategy

1. **Context-based cancellation**
   - All operations accept context.Context
   - Proper cleanup on cancellation

2. **Worker Pools**
   - Configurable number of workers
   - Rate limiting for external tool calls

3. **Synchronization**
   - Mutex protection for shared data structures
   - Channel-based communication between components

## 8. Extension Points

1. **Plugin System**
   - Custom FileTypeDetector implementations
   - Custom MetadataExtractor implementations
   - Custom Deduplicator implementations

2. **Custom Reporters**
   - Additional output formats
   - Integration with external systems

3. **Analysis Extensions**
   - Custom analysis modules
   - Custom recommendation generators

Would you like me to elaborate on any part of this architecture or proceed with implementing specific components?
