I'll help you design a Go architecture to rewrite the downloads folder inspection script. Let me analyze the script and create a comprehensive design document.

First, let me outline the key functionalities of the script:
1. File system traversal and analysis
2. Multiple tool integrations (magika, exiftool, jdupes)
3. File type detection and categorization
4. Size analysis and statistics
5. Duplicate detection
6. Reporting and output generation

Let me create a detailed architecture document:

```markdown
# Downloads Folder Inspector - Go Architecture Design

## Overview
A Go-based system for analyzing and reporting on the contents of a downloads folder, providing insights about file types, sizes, duplicates, and organization recommendations.

## Core Design Principles
- Clean separation of concerns
- Extensible tool integration
- Concurrent processing where possible
- Strong typing for file metadata
- Interface-based design for flexibility
- Context-aware operations for cancellation

## Domain Models

### File Information
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

type FileType struct {
    Label       string
    MimeType    string
    Confidence  float64
    Source      string // "magika", "mime", "extension"
}

type FileGroup string // "image", "video", "document", etc.
```

### Statistics
```go
type FolderStats struct {
    TotalFiles      int64
    TotalSize       int64
    FileTypes       map[FileType]int
    FileGroups      map[FileGroup]int
    MonthlyStats    map[YearMonth]MonthStats
    DuplicateSets   []DuplicateSet
    LargeFiles      []FileInfo
    RecentFiles     []FileInfo
}

type MonthStats struct {
    Count int64
    Size  int64
}

type DuplicateSet struct {
    Hash      string
    Files     []FileInfo
    WastedSpace int64
}
```

## Core Interfaces

### Scanner Interface
```go
type Scanner interface {
    Scan(ctx context.Context, path string, opts ScanOptions) (<-chan FileInfo, <-chan error)
}

type ScanOptions struct {
    MaxDepth      int
    SamplesPerDir int
    IncludeHidden bool
    FilePatterns  []string
}
```

### Type Detector Interface
```go
type TypeDetector interface {
    DetectType(ctx context.Context, file FileInfo) (FileType, error)
}

// Implementations:
// - MagikaDetector
// - MimeDetector
// - ExtensionDetector
```

### Metadata Extractor Interface
```go
type MetadataExtractor interface {
    Extract(ctx context.Context, file FileInfo) (map[string]interface{}, error)
}

// Implementations:
// - ExifToolExtractor
// - MediaInfoExtractor
```

### Deduplicator Interface
```go
type Deduplicator interface {
    FindDuplicates(ctx context.Context, files []FileInfo) ([]DuplicateSet, error)
}

// Implementations:
// - JdupesDeduplicator
// - HashBasedDeduplicator
```

### Reporter Interface
```go
type Reporter interface {
    GenerateReport(ctx context.Context, stats FolderStats) error
    AddSection(name string, content interface{}) error
}

// Implementations:
// - TextReporter
// - JSONReporter
// - HTMLReporter
```

## Services

### Analysis Service
```go
type AnalysisService struct {
    scanner     Scanner
    detector    TypeDetector
    extractor   MetadataExtractor
    deduplicator Deduplicator
    reporter    Reporter
    logger      Logger
}

type AnalysisOptions struct {
    IncludeMetadata   bool
    FindDuplicates    bool
    GenerateReport    bool
    ConcurrencyLimit  int
}
```

### Progress Tracking
```go
type ProgressTracker interface {
    UpdateProgress(stage string, current, total int64)
    SetStatus(status string)
    Error(err error)
}
```

## Concurrent Processing Design

1. **Pipeline Architecture**
   - File Discovery → Type Detection → Metadata Extraction → Analysis
   - Each stage processes files concurrently using worker pools
   - Uses channels for communication between stages

2. **Worker Pool Pattern**
   ```go
   type Worker struct {
       ID      int
       Input   <-chan FileInfo
       Output  chan<- ProcessedFile
       Errors  chan<- error
   }
   ```

3. **Fan-out/Fan-in Pattern** for parallel processing
   - File scanning fans out to multiple type detectors
   - Results fan in to central analysis

## Error Handling

1. **Error Types**
   ```go
   type ScanError struct {
       Path string
       Err  error
   }

   type ProcessError struct {
       Stage string
       File  string
       Err   error
   }
   ```

2. **Error Aggregation**
   - Collect errors during processing
   - Continue on non-critical errors
   - Provide error summary in report

## Configuration

```go
type Config struct {
    ScannerConfig struct {
        MaxDepth        int
        ExcludePatterns []string
        IncludeHidden   bool
    }
    
    DetectorConfig struct {
        PreferredDetector string
        Fallbacks        []string
    }
    
    ExtractorConfig struct {
        EnableExifTool   bool
        EnableMediaInfo  bool
        TimeoutSeconds   int
    }
    
    ReporterConfig struct {
        Format          string
        OutputPath      string
        IncludeSections []string
    }
}
```

## CLI Integration

```go
type CLI struct {
    rootCmd    *cobra.Command
    analyzeCmd *cobra.Command
    config     *Config
    service    *AnalysisService
}
```

## Dependency Injection

Use a DI container (e.g., uber/dig or wire) to manage service instantiation and dependencies.

## Extension Points

1. **Plugin System**
   - Custom type detectors
   - Custom metadata extractors
   - Custom report formats

2. **Event System**
   ```go
   type Event struct {
       Type    EventType
       File    FileInfo
       Data    interface{}
       Time    time.Time
   }
   ```

## Testing Strategy

1. **Unit Tests**
   - Mock interfaces for isolated testing
   - Table-driven tests for type detection
   - Benchmark tests for performance

2. **Integration Tests**
   - Test with real filesystem
   - Test tool integration
   - Test concurrent processing

3. **Test Fixtures**
   - Sample directory structures
   - Various file types
   - Known duplicates

## Performance Considerations

1. **Memory Management**
   - Stream large files
   - Use pools for frequent allocations
   - Garbage collection tuning

2. **Concurrency Control**
   - Configurable worker pools
   - Resource limiting
   - Context cancellation

3. **Caching**
   - Cache type detection results
   - Cache file hashes
   - Cache metadata extraction

## Logging and Monitoring

1. **Structured Logging**
   ```go
   type Logger interface {
       With(fields map[string]interface{}) Logger
       Debug(msg string, fields ...Field)
       Info(msg string, fields ...Field)
       Error(msg string, err error, fields ...Field)
   }
   ```

2. **Metrics**
   ```go
   type Metrics interface {
       IncCounter(name string, value int64)
       ObserveValue(name string, value float64)
       StartTimer(name string) Timer
   }
   ```

This architecture provides a solid foundation for rewriting the shell script in Go, with proper separation of concerns, extensibility, and robust error handling. The interface-based design allows for easy testing and future enhancements.

Would you like me to elaborate on any particular aspect of this design?
