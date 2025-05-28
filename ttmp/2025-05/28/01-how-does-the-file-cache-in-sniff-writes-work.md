# How the FileCache Works in sniff-writes

## Overview

The `filecache` package in the sniff-writes project is a critical component that enables **diff generation** between file read and write operations. It caches file content from read operations and compares it with subsequent write operations to show what changed in the file.

## Purpose and Context

The sniff-writes tool is an eBPF-based file monitoring system that tracks file operations (open, read, write, close) in real-time. The filecache component adds intelligent diff capabilities by:

1. **Storing content from read operations** for future comparison
2. **Generating diffs when write operations occur** to show what changed
3. **Managing cache lifecycle** with automatic cleanup and expiration

## Architecture

### Core Data Structures

#### FileContent
```go
type FileContent struct {
    Content   []byte    // The actual file content
    Hash      string    // SHA256 hash of the content
    Timestamp time.Time // When this content was cached
    Offset    uint64    // File offset where this content was read
    Size      uint64    // Size of the content in bytes
}
```

#### FileCache
```go
type FileCache struct {
    cache  map[string]*FileContent // key: "pid:fd:pathHash:offset"
    mu     sync.RWMutex           // Thread-safe access
    maxAge time.Duration          // Cache expiration (10 minutes)
}
```

### Cache Key Format

The cache uses a composite key format: `"pid:fd:pathHash:offset"`

- **pid**: Process ID that performed the operation
- **fd**: File descriptor number
- **pathHash**: 32-bit hash of the file path (for efficient lookup)
- **offset**: File offset where the operation occurred

This allows tracking content at specific file positions across different processes and file descriptors.

## What is Stored in the Cache

### Content Storage Strategy

The filecache stores **raw byte content** from file read operations along with critical metadata for reconstruction and comparison:

#### 1. Raw Content Data
```go
Content: []byte  // Actual bytes read from the file
```
- **Direct byte storage**: The exact bytes read from the file operation
- **No interpretation**: Content is stored as-is, preserving binary data, text encoding, line endings
- **Size limits**: Respects eBPF content size limits (default 4KB per chunk, max 128KB total)
- **Deep copy**: Content is copied to prevent memory corruption from concurrent access

#### 2. Positional Metadata
```go
Offset: uint64  // File offset where this content was read
Size:   uint64  // Number of bytes in this content chunk
```
- **File position tracking**: Records exactly where in the file this content came from
- **Chunk boundaries**: Enables reconstruction of larger files from multiple read operations
- **Offset-based lookup**: Allows precise matching of read/write operations at the same file position

#### 3. Integrity and Lifecycle Data
```go
Hash:      string    // SHA256 hash for content verification
Timestamp: time.Time // When this content was cached
```
- **Content verification**: SHA256 hash ensures content integrity over time
- **Cache expiration**: Timestamp enables TTL-based cleanup (10-minute default)
- **Change detection**: Hash comparison provides fast "content changed" detection

### File Reconstruction for Diffing

The filecache doesn't attempt to reconstruct complete files. Instead, it uses a **position-based matching strategy** for targeted diffing:

#### Position-Based Content Matching

When a write operation occurs, the cache lookup uses the exact same key components:
```go
key := fmt.Sprintf("%d:%d:%d:%d", pid, fd, pathHash, offset)
```

**Matching Logic:**
1. **Exact position match**: Write at offset X is compared with cached read at offset X
2. **Same file context**: Must be same process (pid), file descriptor (fd), and file (pathHash)
3. **Recent content**: Cached content must be within TTL window (not expired)

#### Content Comparison Process

```go
func (fc *FileCache) GenerateDiff(pid uint32, fd int32, pathHash uint32, offset uint64, newContent []byte) (string, bool) {
    // 1. Lookup cached content at exact same position
    cachedContent, exists := fc.GetContentForDiff(pid, fd, pathHash, offset)
    if !exists {
        return "", false  // No cached content to compare against
    }

    // 2. Direct byte-to-byte comparison
    dmp := diffmatchpatch.New()
    diffs := dmp.DiffMain(string(cachedContent.Content), string(newContent), false)

    // 3. Check if there are actual changes
    hasChanges := false
    for _, diff := range diffs {
        if diff.Type != diffmatchpatch.DiffEqual {
            hasChanges = true
            break
        }
    }

    if !hasChanges {
        return "", false  // No differences found
    }

    return dmp.DiffPrettyText(diffs), true
}
```

#### Why No Full File Reconstruction?

The filecache deliberately **does not reconstruct complete files** for several important reasons:

1. **Memory Efficiency**: Storing complete files would consume excessive memory
2. **eBPF Limitations**: eBPF programs can only capture limited content per operation
3. **Real-time Performance**: File reconstruction would add significant latency
4. **Partial Updates**: Most file operations modify specific regions, not entire files
5. **Cache Complexity**: Managing file assembly across multiple operations would be error-prone

#### Chunked Content Handling

For operations that span multiple eBPF events (large reads/writes), the system handles chunks:

```go
// eBPF emits multiple events for large operations
struct event {
    // ... other fields ...
    __u32 chunk_seq;     // Sequence number (0-based)
    __u32 total_chunks;  // Total chunks for this operation
    char content[MAX_CONTENT_LEN]; // Content for this chunk
};
```

**Chunk Processing:**
- Each chunk is stored separately with its specific offset
- Diffs are generated per-chunk, not across reconstructed content
- This maintains real-time performance while handling large operations

#### Content Lifecycle in Cache

```
Read Operation → Store Content → Wait for Write → Generate Diff → Update Cache
     ↓              ↓              ↓               ↓              ↓
[File Read]    [Cache Entry]   [Write Detected] [Diff Output]  [New Content]
  offset=100     key="1:5:hash:100"  offset=100    "- old line"   updated entry
  content="old"  content="old"       content="new"  "+ new line"  content="new"
```

#### Diff Generation Strategies

The cache supports multiple diff strategies based on content type and size:

1. **Byte-level diffing**: For binary content or small changes
2. **Line-level diffing**: For text content with line structure
3. **Character-level diffing**: For detailed text analysis
4. **Elided diffing**: For large content with limited context

#### Cache Miss Scenarios

When cached content is not available for diffing:

```go
// Common cache miss scenarios:
// 1. No prior read operation at this offset
// 2. Cached content has expired (> 10 minutes old)
// 3. Different process/fd accessing same file
// 4. Write operation at new file offset
// 5. Cache was cleared or process restarted
```

**Graceful Handling:**
- No diff is generated (returns `"", false`)
- Write operation is still logged and processed
- New content is stored for future comparisons
- No errors or failures in the monitoring pipeline

#### Memory Management

The cache implements several strategies to manage memory usage:

```go
// Content storage optimization:
func (fc *FileCache) StoreReadContent(pid uint32, fd int32, pathHash uint32, content []byte, offset uint64) {
    // Deep copy to prevent memory sharing issues
    contentCopy := make([]byte, len(content))
    copy(contentCopy, content)
    
    // Store with metadata for efficient lookup and cleanup
    fc.cache[key] = &FileContent{
        Content:   contentCopy,           // Isolated copy
        Hash:      calculateHash(content), // For integrity checking
        Timestamp: time.Now(),            // For TTL expiration
        Offset:    offset,                // For position matching
        Size:      uint64(len(content)),  // For size tracking
    }
}
```

**Memory Efficiency Features:**
- **TTL-based expiration**: Automatic cleanup of old entries
- **Size limits**: Respects eBPF and configuration limits
- **Deep copying**: Prevents memory corruption but uses more space
- **Hash-based deduplication**: Could be added for identical content (future enhancement)

## Usage Flow in sniff-writes

### 1. Initialization
```go
// In main.go
fileCache = filecache.New()
```

### 2. Read Operation Handling
When the eBPF program captures a read operation:

```go
case 1: // read event
    // Store read content for future diff comparison
    fileCache.StoreReadContent(event.Pid, event.Fd, event.PathHash, contentBytes, event.FileOffset)
```

**What happens:**
- Content from the read operation is stored in the cache
- SHA256 hash is calculated for integrity
- Timestamp is recorded for expiration tracking
- Content is deep-copied to prevent memory issues

### 3. Write Operation Handling
When the eBPF program captures a write operation:

```go
case 2: // write event
    if config.ShowDiffs {
        var diffText string
        var hasDiff bool

        if config.DiffFormat == "pretty" {
            diffText, hasDiff = fileCache.GenerateDiff(event.Pid, event.Fd, event.PathHash, event.FileOffset, contentBytes)
        } else {
            // Use elided diff if context lines is configured
            if config.DiffContextLines >= 0 {
                diffText, hasDiff = fileCache.GenerateElidedUnifiedDiff(event.Pid, event.Fd, event.PathHash, event.FileOffset, contentBytes, resolvedPath, config.DiffContextLines)
            } else {
                diffText, hasDiff = fileCache.GenerateUnifiedDiff(event.Pid, event.Fd, event.PathHash, event.FileOffset, contentBytes, resolvedPath)
            }
        }

        if hasDiff {
            eventOutput.Diff = diffText
        }
    }

    // Update cache with new written content
    fileCache.UpdateWriteContent(event.Pid, event.Fd, event.PathHash, contentBytes, event.FileOffset)
```

**What happens:**
1. **Diff Generation**: Compare cached read content with new write content
2. **Format Selection**: Choose between pretty, unified, or elided diff formats
3. **Cache Update**: Store the new written content for future comparisons

## Diff Generation Methods

### 1. GenerateDiff (Pretty Format)
- Uses `github.com/sergi/go-diff/diffmatchpatch` library
- Produces human-readable diff with inline changes
- Good for small content changes

### 2. GenerateUnifiedDiff (Standard Unified Format)
- Produces traditional unified diff format with line numbers
- Format: `--- filename (cached)` / `+++ filename (new write)`
- Shows line-by-line changes with `+`, `-`, and ` ` prefixes

### 3. GenerateElidedUnifiedDiff (Context-Limited)
- Same as unified diff but limits context lines around changes
- Reduces noise in large files by showing only relevant sections
- Uses `...` markers to indicate elided content

## Diff Elision Feature

The filecache includes sophisticated diff elision capabilities in `diff_utils.go`:

### Key Components

#### DiffLine Structure
```go
type DiffLine struct {
    Type    DiffLineType // Type of line (add, remove, context)
    Content string       // Content of the line
    OldLine int          // Line number in old file
    NewLine int          // Line number in new file
}
```

#### ElideUnifiedDiff Function
- Parses unified diff into structured format
- Identifies change positions
- Keeps only specified number of context lines around changes
- Adds `...` markers for elided sections

## Cache Management

### Automatic Cleanup
```go
func (fc *FileCache) Cleanup() {
    fc.mu.Lock()
    defer fc.mu.Unlock()

    now := time.Now()
    for key, content := range fc.cache {
        if now.Sub(content.Timestamp) > fc.maxAge {
            delete(fc.cache, key)
        }
    }
}
```

### Expiration Policy
- **Default TTL**: 10 minutes
- **Automatic expiration**: Content older than maxAge is ignored
- **Memory management**: Prevents unbounded cache growth

## Thread Safety

The FileCache is designed for concurrent access:
- **RWMutex**: Allows multiple concurrent readers, exclusive writers
- **Read operations**: Use `RLock()` for better performance
- **Write operations**: Use `Lock()` for exclusive access
- **Deep copying**: Content is copied to prevent race conditions

## Integration with eBPF Events

### Event Flow
1. **eBPF Program** captures syscalls (read/write) with content
2. **Ring Buffer** transfers events to userspace
3. **Event Processing** in main.go handles each event
4. **FileCache** stores/compares content based on event type
5. **Output Formatting** includes diff information if available

### Content Capture Control
- Content capture is controlled by eBPF map `content_capture_enabled`
- Can be toggled via command-line flags
- Affects both read and write content storage

## Configuration Options

### Command-line Flags
- `--capture-content`: Enable content capture (required for diffs)
- `--show-diffs`: Enable diff generation
- `--diff-format`: Choose between "pretty" and "unified"
- `--diff-context`: Number of context lines for elided diffs

### Example Usage
```bash
# Enable content capture and diffs with 3 context lines
sudo ./sniff-writes monitor --capture-content --show-diffs --diff-context 3

# Use pretty diff format
sudo ./sniff-writes monitor --capture-content --show-diffs --diff-format pretty
```

## Performance Considerations

### Memory Usage
- Content is stored in memory with configurable size limits
- Default content size limit: 4KB per chunk (configurable up to 128KB)
- Automatic cleanup prevents memory leaks

### CPU Impact
- Diff generation uses efficient algorithms from diffmatchpatch library
- Elision reduces processing time for large files
- SHA256 hashing provides fast content comparison

### Cache Efficiency
- Composite keys enable precise content matching
- RWMutex allows concurrent read access
- TTL-based expiration balances memory and functionality

## Error Handling

### Graceful Degradation
- Missing cached content: No diff generated, operation continues
- Expired content: Treated as cache miss
- Memory allocation failures: Events logged without content

### Logging
- Failed cache operations are logged but don't stop monitoring
- SQLite logging errors are reported but don't affect real-time output

## Use Cases

### Development Monitoring
- Track configuration file changes
- Monitor log file updates
- Debug application file I/O patterns

### Security Analysis
- Detect unauthorized file modifications
- Monitor sensitive file access patterns
- Audit file content changes

### Performance Analysis
- Identify frequent file operations
- Analyze I/O patterns
- Monitor application behavior

## Limitations

### Current Constraints
- **Memory-based**: Cache is not persistent across restarts
- **Single-machine**: No distributed caching
- **Content size**: Limited by eBPF and memory constraints
- **File position**: Assumes sequential or known offset operations

### Future Enhancements
- Persistent cache storage
- Configurable cache policies
- Enhanced diff algorithms
- Cross-process content correlation

## Related Components

### PathCache (`pkg/cache/pathcache.go`)
- Resolves file paths from file descriptors
- Provides path information for diff headers
- Complements filecache functionality

### Database Integration
- SQLite storage includes diff information
- Web UI displays diffs with HTML formatting
- Query interface supports diff-based filtering

### Web Interface
- Real-time diff display via WebSocket
- HTML-formatted diff output
- Interactive file monitoring dashboard

## Testing

The filecache includes comprehensive tests in `diff_utils_test.go`:
- Unit tests for diff parsing and elision
- Benchmark tests for performance validation
- Edge case handling (empty files, large diffs)
- Format validation tests

## Conclusion

The filecache component is a sophisticated caching and diff system that transforms basic file monitoring into intelligent change tracking. By caching read content and comparing it with write operations, it provides valuable insights into how files are being modified in real-time, making it an essential tool for development, security, and performance analysis. 