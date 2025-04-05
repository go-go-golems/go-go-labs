<think>
Okay, the user wants me to design a Go architecture for rewriting their existing Bash script into Go. Let me start by understanding what the Bash script does.

Looking at the provided script, it's a Downloads folder analyzer. It collects various statistics like file types, duplicates, large files, recent files, and generates a report. It uses tools like Magika, ExifTool, and jdupes, and has options for verbose logging and sampling.

Now, the user wants a Go design with proper patterns and interfaces. Let me think about how to structure this in Go.

First, I'll need to break down the functionalities into components. The main parts are directory traversal, file analysis, metadata extraction, duplicate detection, and report generation. Each of these can be separate modules or packages.

For design patterns, the Strategy pattern makes sense for interchangeable algorithms, like different duplicate detectors. The Factory pattern could create analysis modules based on configuration. The Observer pattern can handle progress updates, which is useful for long-running tasks.

Interfaces will be key here. I'll define interfaces for FileAnalyzer, DuplicateDetector, MetadataExtractor, and ReportGenerator. This allows swapping implementations without changing the core code. For example, using Magika or a fallback method for file type detection.

Concurrency is important in Go. Using goroutines and channels for parallel file processing would speed things up. An errgroup can manage goroutines and handle errors gracefully.

Error handling should be robust. Wrapping errors with context using github.com/pkg/errors will help in debugging. Config validation ensures that the application starts with correct parameters.

Logging needs to be structured, with different levels (debug, info, warn). The user mentioned not removing debug logs, so a configurable logger that includes debug info when verbose is enabled is necessary.

For the CLI, using Cobra is a good choice as it's a popular Go library for building command-line applications. It supports flags, subcommands, and help generation.

I should also consider the existing Bash script's options, like --verbose and --sample-per-dir, and map them to Go flags. The report structure should mirror the current text output but could be extended to support different formats (HTML, JSON) in the future.

Testing is another aspect. Each component should be testable, with mock implementations for dependencies. For example, a mock file walker for testing analysis without touching the filesystem.

Finally, the architecture should be modular, allowing the user to add new analysis modules or output formats without major rewrites. Using interfaces and dependency injection will facilitate this.

I need to outline all these points in a structured markdown document, covering design patterns, core interfaces, package structure, concurrency, error handling, logging, CLI, and potential extensions. Make sure to highlight how each part of the Bash script is translated into the Go design, ensuring all functionalities are covered and improved upon where possible.
</think>

Here's a proposed Go architecture for rewriting the Downloads folder analyzer:

# Go Architecture Design Document

## Core Design Patterns
1. **Strategy Pattern**  
- For interchangeable analysis modules (file type detection, duplicates, metadata)
- For different output formats (text, JSON, HTML reports)

2. **Factory Pattern**  
- Create analysis modules based on configuration
- Create different metadata extractors (Exif, Magika, fallback)

3. **Observer Pattern**  
- Progress reporting system
- Real-time analysis updates

4. **Pipeline Pattern**  
- File processing pipeline:  
  `Walk → Analyze → Enrich → Aggregate → Report`

## Core Interfaces

```go
// File Analysis
type FileAnalyzer interface {
    Analyze(path string, info os.FileInfo) (FileMetadata, error)
    Name() string
}

// Duplicate Detection
type DuplicateDetector interface {
    AddFile(path string, hash string) ([]string, error) // returns duplicate paths
    GetDuplicates() map[string][]string
}

// Metadata Extraction
type MetadataExtractor interface {
    Extract(path string) (FileMetadata, error)
    SupportsType(mimeType string) bool
}

// Report Generation
type ReportGenerator interface {
    AddSection(name string, data interface{}) error
    Generate(outputPath string) error
}
```

## Package Structure

```
pkg/downloads-analyzer/
├── core             // Main analysis logic
├── metadata         // Metadata extraction implementations
│   ├── exif
│   ├── magika
│   └── file
├── duplicates       // Duplicate detection strategies
│   ├── jdupes
│   └── hash
├── reporting        // Report generators
│   ├── text
│   └── html
├── walker           // Directory traversal
└── types            // Common data types
```

## Key Components

### 1. Concurrent File Walker
- Uses `filepath.WalkDir` with worker pool
- Implements sampling logic per directory
- Emits files through channel for parallel processing

### 2. Analysis Pipeline
```go
type AnalysisPipeline struct {
    steps []AnalysisStep
}

type AnalysisStep interface {
    Process(file *FileMetadata) error
    Finalize() error
}
```

### 3. Metadata Enrichers
- Chain of responsibility pattern for metadata extraction:
  1. Magika-based type detection
  2. ExifTool metadata
  3. Fallback to file command

### 4. Duplicate Detection
- Pluggable strategies:
  - jdupes (if available)
  - Content hash comparison
  - Fast hash (xxHash) for quick comparisons

### 5. Report System
- Template-based reporting
- Supports multiple output formats
- Progressive rendering for large datasets

## Concurrency Model
- Worker pool pattern for parallel file processing
- `errgroup` for coordinated goroutine management
- Buffered channels for file processing pipeline
- Atomic counters for statistics collection

## Error Handling
- Wrapped errors with context using `github.com/pkg/errors`
- Error classification:
  ```go
  type ErrorSeverity int
  const (
      SeverityWarning ErrorSeverity = iota
      SeverityPartial
      SeverityFatal
  )
  ```
- Retry mechanism for transient errors

## Configuration Management
- Hierarchical config using Viper:
  ```go
  type Config struct {
      MaxWorkers       int
      SamplePerDir     int
      AnalysisModules  []string
      ReportFormat     string
      DuplicateMethod  string
  }
  ```

## Logging System
- Structured logging with levels
- Progress bar integration
- Debug log file writer
- Contextual logging with fields:
  ```go
  logger.WithFields(logrus.Fields{
      "path": filePath,
      "size": fileSize,
  }).Debug("Processing file")
  ```

## CLI Architecture (Cobra)
```text
downloads-analyzer/
  ├── analyze
  │   ├── --output-format
  │   ├── --sample-per-dir
  │   └── --parallel
  ├── report
  │   ├── generate
  │   └── compare
  └── clean
      ├── duplicates
      └── old-files
```

## Key Differences from Bash Version
1. **Performance**  
- Parallel processing of files
- Memory-mapped file operations
- Cache-friendly data structures

2. **Extensibility**  
- Plugin system for new analysis modules
- Support multiple output formats

3. **Safety**  
- Type-safe error handling
- Memory safety
- Graceful shutdown handling

4. **Maintainability**  
- Strong typing for analysis results
- Unit testable components
- Versioned configuration

## Security Considerations
- Sandboxed execution environment
- File path sanitization
- Memory limits for large files
- Secure temp file handling

Would you like me to elaborate on any particular component or discuss implementation details for specific subsystems?
