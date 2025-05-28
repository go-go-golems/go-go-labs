# FileCache API Documentation

## Overview

The FileCache provides a sparse file caching system that tracks read/write operations for diff generation. It maintains both a new sparse representation and legacy compatibility.

## Core Types

### FileCache

The main cache structure that manages file content segments.

```go
type FileCache struct {
    files map[uint32]*SparseFile // pathHash -> sparse file representation
    cache map[string]*FileContent // legacy cache: "pid:fd:pathHash:offset"
    
    mu           sync.RWMutex
    maxAge       time.Duration
    perFileLimit uint64       // bytes per file
    globalLimit  uint64       // total bytes across all files
    totalSize    uint64       // current total bytes stored
    timeProvider TimeProvider // for mocking time in tests
}
```

### SparseFile

Represents a file as a collection of non-overlapping segments.

```go
type SparseFile struct {
    mu       sync.RWMutex
    Segments []*Segment // sorted, non-overlapping, non-adjacent (merged)
    Size     uint64     // bytes stored (for limit accounting)
    LastUsed time.Time  // for LRU eviction
}
```

### Segment

Represents a contiguous range of file content.

```go
type Segment struct {
    Start   uint64 // inclusive offset
    End     uint64 // exclusive offset (Start+len(Data))
    Data    []byte // content data, len == End-Start
    AddedAt time.Time
}
```

### TimeProvider

Interface for time handling (allows mocking in tests).

```go
type TimeProvider interface {
    Now() time.Time
}

type RealTimeProvider struct{}
func (r RealTimeProvider) Now() time.Time { return time.Now() }

type MockTimeProvider struct {
    mu   sync.Mutex
    time time.Time
}
func NewMockTimeProvider(start time.Time) *MockTimeProvider
func (m *MockTimeProvider) Now() time.Time
func (m *MockTimeProvider) Advance(d time.Duration)
```

## Constructor Functions

### New()

Creates a FileCache with default settings.

```go
func New() *FileCache
```

**Returns**: FileCache with defaults:
- maxAge: 10 minutes
- perFileLimit: 512 KB
- globalLimit: 64 MB
- timeProvider: RealTimeProvider{}

### NewFileCache()

Creates a FileCache with custom settings.

```go
func NewFileCache(perFileLimit, globalLimit uint64, maxAge time.Duration, timeProvider TimeProvider) *FileCache
```

**Parameters**:
- `perFileLimit`: Maximum bytes per individual file
- `globalLimit`: Maximum total bytes across all files  
- `maxAge`: TTL for cache entries
- `timeProvider`: Time source (use `NewMockTimeProvider(time.Now())` for testing)

**Returns**: Configured FileCache

### NewWithTimeProvider()

Creates a FileCache with default settings but custom time provider.

```go
func NewWithTimeProvider(tp TimeProvider) *FileCache
```

**Parameters**:
- `tp`: Custom time provider

**Returns**: FileCache with custom time provider

## Core API Methods

### AddRead()

Stores content from a read operation in the sparse cache.

```go
func (fc *FileCache) AddRead(pathHash uint32, offset uint64, data []byte)
```

**Parameters**:
- `pathHash`: Hash identifying the file path
- `offset`: Byte offset in file where data was read
- `data`: Content that was read (can be nil or empty)

**Behavior**:
- No return value
- Automatically merges adjacent/overlapping segments
- Enforces per-file and global memory limits
- Updates LRU timestamps

### UpdateWithWrite()

Invalidates cached segments affected by a write operation.

```go
func (fc *FileCache) UpdateWithWrite(pathHash uint32, offset uint64, data []byte)
```

**Parameters**:
- `pathHash`: Hash identifying the file path
- `offset`: Byte offset where write occurred
- `data`: Data that was written

**Behavior**:
- No return value
- Splits/removes segments that overlap with write range
- Updates cache state to reflect write invalidation

### GetOldContent()

Retrieves cached content for diff comparison.

```go
func (fc *FileCache) GetOldContent(pathHash uint32, offset uint64, size uint64) ([]byte, bool)
```

**Parameters**:
- `pathHash`: Hash identifying the file
- `offset`: Start offset to retrieve
- `size`: Number of bytes to retrieve

**Returns**:
- `[]byte`: Retrieved content (reconstructed from segments)
- `bool`: true if content exists, false if not cached

**Behavior**:
- Reconstructs content from sparse segments
- Fills gaps with zero bytes (0x00)
- Updates LRU timestamps

### GenerateDiff()

Generates a diff between cached content and new content.

```go
func (fc *FileCache) GenerateDiff(pid uint32, fd int32, pathHash uint32, offset uint64, newContent []byte) (string, bool)
```

**Parameters**:
- `pid`: Process ID (for legacy API compatibility)
- `fd`: File descriptor (for legacy API compatibility)
- `pathHash`: Hash identifying the file
- `offset`: Offset where comparison starts
- `newContent`: New content to compare against

**Returns**:
- `string`: Generated diff text
- `bool`: true if diff was generated, false if no cached content

### StoreReadContent()

Legacy API method that stores content in both new and old cache formats.

```go
func (fc *FileCache) StoreReadContent(pid uint32, fd int32, pathHash uint32, content []byte, offset uint64)
```

**Parameters**:
- `pid`: Process ID
- `fd`: File descriptor
- `pathHash`: Hash identifying the file path
- `content`: Content that was read
- `offset`: Byte offset in file

**Note**: Calls `AddRead()` internally and maintains legacy cache for backward compatibility.

## Utility Methods

### Size()

Returns the total number of cached entries across both cache formats.

```go
func (fc *FileCache) Size() int
```

### CleanExpired()

Removes expired cache entries based on TTL.

```go
func (fc *FileCache) CleanExpired()
```

## SparseFile Methods (Internal)

### GetContentRange()

Reconstructs content for a specific range from segments.

```go
func (sf *SparseFile) GetContentRange(offset, length uint64, gapByte byte) ([]byte, bool)
```

**Parameters**:
- `offset`: Start offset
- `length`: Number of bytes to retrieve
- `gapByte`: Byte value to fill gaps (typically 0x00)

**Returns**:
- `[]byte`: Reconstructed content
- `bool`: true if any relevant segments found

### UpdateWithWrite()

Invalidates segments affected by a write operation.

```go
func (sf *SparseFile) UpdateWithWrite(offset uint64, data []byte, timeProvider TimeProvider)
```

## Usage Examples

### Basic Usage

```go
// Create cache
fc := NewFileCache(512*1024, 64*1024*1024, time.Hour, RealTimeProvider{})

// Store read data
pathHash := uint32(12345)
data := []byte("file content")
fc.AddRead(pathHash, 0, data)

// Retrieve cached content
retrieved, exists := fc.GetOldContent(pathHash, 0, uint64(len(data)))
if exists {
    fmt.Printf("Cached: %s\n", retrieved)
}

// Update with write
newData := []byte("modified content")
fc.UpdateWithWrite(pathHash, 0, newData)

// Generate diff
diff, hasDiff := fc.GenerateDiff(1234, 5, pathHash, 0, newData)
if hasDiff {
    fmt.Printf("Diff: %s\n", diff)
}
```

### Testing Usage

```go
// Create cache with mock time for testing
mockTime := NewMockTimeProvider(time.Now())
fc := NewFileCache(1024, 4096, time.Hour, mockTime)

// Add test data
fc.AddRead(42, 0, []byte("test data"))

// Advance time for TTL testing
mockTime.Advance(2 * time.Hour)
fc.CleanExpired()

// Check if data was evicted
_, exists := fc.GetOldContent(42, 0, 9)
// exists should be false due to TTL expiration
```

## Important Notes

### Thread Safety
- All public methods are thread-safe
- Uses RWMutex for concurrent read access
- Internal segment operations are protected by per-file mutexes

### Memory Management
- Enforces per-file limits (`perFileLimit`)
- Enforces global cache limits (`globalLimit`)
- Uses LRU eviction when limits exceeded
- Automatic cleanup of expired entries

### Segment Merging
- Adjacent segments are automatically merged
- Overlapping segments are consolidated
- Zero-length segments are handled gracefully

### Legacy Compatibility
- Maintains dual cache formats during transition
- `StoreReadContent()` populates both caches
- Legacy cache key format: `"pid:fd:pathHash:offset"`

### Error Handling
- Methods don't return errors for invalid inputs
- Gracefully handles nil data, zero offsets, etc.
- Invalid operations are logged but don't panic

### Hash Collisions
- Different files with same `pathHash` are treated as same file
- Implementation assumes good hash distribution
- Consider using file inode + device for better uniqueness
