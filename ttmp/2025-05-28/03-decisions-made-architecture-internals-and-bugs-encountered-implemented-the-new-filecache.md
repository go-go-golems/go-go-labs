# Implementation Report: Sparse File Cache for sniff-writes

## Overview

This document details the implementation of a new sparse file cache system for the sniff-writes project, based on the proposal in `02-proposal-for-sparse-file-filecache.md`. The goal was to replace the simple single-offset cache with a sophisticated sparse representation that can track multiple read segments per file and provide better diff generation.

## Architecture Design

### Core Data Structures

```go
// Segment represents a contiguous range of file content
type Segment struct {
    Start   uint64    // inclusive offset
    End     uint64    // exclusive offset (Start+len(Data))
    Data    []byte    // actual content data
    AddedAt time.Time // for TTL expiration
}

// SparseFile represents a file as a collection of non-overlapping segments
type SparseFile struct {
    mu       sync.RWMutex
    Segments []*Segment // sorted by Start offset, non-overlapping
    Size     uint64     // total bytes stored (for memory limits)
    LastUsed time.Time  // for LRU eviction
}

// FileCache is the main cache with both new sparse and legacy storage
type FileCache struct {
    files map[uint32]*SparseFile    // pathHash -> sparse file representation
    cache map[string]*FileContent   // legacy single-offset cache
    
    mu           sync.RWMutex
    maxAge       time.Duration
    perFileLimit uint64           // 512KB per file default
    globalLimit  uint64           // 64MB total default
    totalSize    uint64           // current total size
    timeProvider TimeProvider     // for testing with mock time
}
```

### Key Design Decisions

1. **Two-Level Indexing**: Files indexed by `pathHash`, segments within files ordered by start offset
2. **Automatic Merging**: Adjacent and overlapping segments are automatically merged during insertion
3. **Memory Limits**: Both per-file and global limits with LRU eviction
4. **Lock Hierarchy**: FileCache.mu > SparseFile.mu to prevent deadlocks
5. **Time Abstraction**: TimeProvider interface for reliable testing

## Implementation Approach

### 1. Segment Merging Logic

The `insertSegment` method implements smart merging:

```go
func (sf *SparseFile) insertSegment(newSeg *Segment, timeProvider TimeProvider) {
    // Find all segments that overlap or are adjacent
    var toMerge []*Segment
    var toKeep []*Segment
    
    for _, seg := range sf.Segments {
        if seg.End >= newSeg.Start && seg.Start <= newSeg.End {
            toMerge = append(toMerge, seg)  // Overlaps
        } else {
            toKeep = append(toKeep, seg)    // Keep separate
        }
    }
    
    // Merge all overlapping segments into one
    // ... (boundary calculation and data merging)
}
```

**Key Insight**: This approach ensures segments remain non-overlapping and sorted, which simplifies all other operations.

### 2. Write Invalidation Strategy

When a write occurs, we need to update cached segments:

```go
func (sf *SparseFile) UpdateWithWrite(offset uint64, data []byte, timeProvider TimeProvider) {
    writeEnd := offset + uint64(len(data))
    
    for each overlapping segment {
        if segment.Start < writeOffset {
            // Keep prefix [segment.Start, writeOffset)
        }
        if segment.End > writeEnd {
            // Keep suffix [writeEnd, segment.End)  
        }
        // Remove the overlapping middle part
    }
    
    // Add new segment with write data
}
```

**Key Insight**: This properly handles all overlap cases - complete replacement, splitting, and partial overlaps.

### 3. Content Reconstruction

For diff generation, we reconstruct content across multiple segments:

```go
func (sf *SparseFile) GetContentRange(offset, length uint64, gapByte byte) ([]byte, bool) {
    result := make([]byte, length)
    // Fill with gap bytes initially
    
    for each overlapping segment {
        // Copy segment data to appropriate position in result
        // Track which bytes were filled vs gaps
    }
}
```

## Bugs Encountered and Solutions

### 1. Critical Deadlock in AddRead Method

**Problem**: The `AddRead` method was holding both `fc.mu` and `sf.mu` simultaneously, then calling `enforcePerFileLimit(sf)` which tried to lock `sf.mu` again.

```go
// BUGGY CODE:
func (fc *FileCache) AddRead(...) {
    fc.mu.Lock()
    defer fc.mu.Unlock()
    
    sf.mu.Lock()
    defer sf.mu.Unlock()
    
    // ... insert segment ...
    
    fc.enforcePerFileLimit(sf)  // Tries to lock sf.mu again! DEADLOCK
}
```

**Solution**: Restructured locking to release locks before calling methods that need them:

```go
// FIXED CODE:
func (fc *FileCache) AddRead(...) {
    fc.mu.Lock()
    // Get or create sparse file
    fc.mu.Unlock()
    
    // Insert segment (handles its own locking)
    sf.insertSegment(newSeg, fc.timeProvider)
    
    fc.mu.Lock()
    fc.totalSize += segmentSize
    fc.mu.Unlock()
    
    // Enforce limits (handle their own locking)
    fc.enforcePerFileLimit(sf)
    fc.enforceGlobalLimit()
}
```

### 2. Time Dependency in Tests

**Problem**: Tests using real time were slow and unreliable, especially cache expiration tests.

**Solution**: Implemented `TimeProvider` interface:

```go
type TimeProvider interface {
    Now() time.Time
}

type MockTimeProvider struct {
    mu   sync.Mutex
    time time.Time
}

func (m *MockTimeProvider) Advance(d time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.time = m.time.Add(d)
}
```

This allows tests to control time progression deterministically.

### 3. Concurrent Access Deadlock

**Problem**: The `TestConcurrentAccess` test hangs, indicating another deadlock scenario.

**Current Status**: Still investigating. Likely related to lock ordering between the file-level and global-level operations.

**Debug Output**: Shows that some goroutines complete but others hang, suggesting a race condition in the locking hierarchy.

### 4. Test Case Logic Error

**Problem**: The `write_spans_multiple_segments` test expects incorrect behavior.

**Debug Output**:
```
Initial segments:
  [0] [10,20) "first" (5 bytes)
  [1] [30,40) "second" (6 bytes)  
  [2] [50,60) "third" (5 bytes)

Write: offset=15, data="REPLACEMENT" (11 bytes)
Write range: [15,26)

Expected: only 3 segments (missing [30,40))
Actual: 4 segments (includes [30,40))
```

**Analysis**: The write range [15,26) does NOT overlap with segment [30,40), so the segment should remain. The test expectation is incorrect.

## Performance Characteristics

### Memory Complexity
- **Per-file overhead**: O(number of segments per file)
- **Typical case**: 1-3 segments per file (most reads are sequential)
- **Worst case**: O(number of individual reads) if all reads are non-adjacent

### Time Complexity
- **Insert segment**: O(n) where n = segments in file
- **Get content**: O(n) for overlapping segments  
- **Write invalidation**: O(n) for overlapping segments
- **Binary search optimization**: Could improve to O(log n) but current O(n) is fine for typical segment counts

### Memory Limits
- **Per-file limit**: 512KB (configurable)
- **Global limit**: 64MB (configurable)
- **Eviction**: LRU at file level, oldest segments first within files

## Integration with Existing Code

### Backward Compatibility
The implementation maintains the existing API:
- `StoreReadContent()` - now calls `AddRead()` internally
- `GenerateDiff()` - tries sparse cache first, falls back to legacy
- `GetContentForDiff()` - legacy method still works

### New API Methods
- `AddRead(pathHash, offset, data)` - modern segment storage
- `GetOldContent(pathHash, offset, size)` - retrieve for diff  
- `UpdateWithWrite(pathHash, offset, data)` - invalidate stale data

## Testing Strategy

### Comprehensive Test Coverage

1. **Segment Merging Tests**
   - Adjacent segments: [0,5) + [5,10) → [0,10)
   - Overlapping segments: [0,11) + [6,17) → [0,17) 
   - Contained segments: [10,30) + [15,20) → [10,30)
   - Multiple merging: [0,5) + [25,30) + [5,25) → [0,30)

2. **Write Invalidation Tests**
   - Complete replacement: write exactly matches segment
   - Segment splitting: write in middle of segment
   - Partial overlaps: write overlaps start/end of segments
   - Multiple segment spanning: write covers multiple segments

3. **Content Reconstruction Tests**  
   - Exact matches, partial matches, gap filling
   - Multiple segment assembly
   - Edge cases with no coverage

4. **Cache Management Tests**
   - TTL expiration with mock time
   - Memory limit enforcement  
   - LRU eviction behavior
   - Concurrent access (currently problematic)

### Current Test Status

✅ **Passing Tests (38/40)**:
- All diff utility tests
- Segment merging logic  
- Content reconstruction
- Cache expiration (with mock time)
- Memory limits
- API compatibility

❌ **Failing Tests (2/40)**:
- `write_spans_multiple_segments` - test case logic error
- `TestConcurrentAccess` - deadlock in concurrent operations

## Future Improvements

### Short-term Fixes
1. **Fix concurrent access deadlock** - requires careful lock ordering analysis
2. **Correct test case expectations** - fix the spanning segments test
3. **Add more debug logging** - for production troubleshooting

### Medium-term Enhancements  
1. **Binary search optimization** - for files with many segments
2. **Compression** - for stored segment data
3. **Metrics collection** - cache hit rates, segment counts, memory usage

### Long-term Possibilities
1. **Persistent storage** - survive process restarts
2. **Cross-process sharing** - shared memory cache
3. **Content-based deduplication** - identical content across files

## Lessons Learned

### Lock Design
- **Always define lock hierarchy explicitly** - document the order locks must be acquired
- **Minimize lock scope** - release locks before calling other methods that need locks
- **Use read locks when possible** - many operations only need read access

### Time Dependencies
- **Always abstract time in tests** - makes tests fast and deterministic  
- **Mock external dependencies early** - don't let real time creep into test logic

### Test Design
- **Write tests first when possible** - helps clarify expected behavior
- **Add debug logging proactively** - makes debugging much easier
- **Test concurrent scenarios carefully** - they often reveal subtle bugs

### Memory Management
- **Track resource usage explicitly** - don't rely on garbage collection alone
- **Implement limits from the start** - easier than adding them later
- **Design for observability** - expose metrics for monitoring

## Conclusion

The sparse file cache implementation successfully achieves the primary goals:

1. ✅ **Multiple segments per file** - can track overlapping reads
2. ✅ **Automatic merging** - keeps memory usage reasonable  
3. ✅ **Better diff generation** - reconstructs content across segments
4. ✅ **Memory management** - enforces per-file and global limits
5. ✅ **Backward compatibility** - existing code still works

The main remaining work is resolving the concurrent access deadlock and correcting the test case expectations. Once these are fixed, the cache should be ready for production use.

The implementation demonstrates the complexity of building thread-safe, memory-managed caching systems. Key insights include the importance of careful lock design, time abstraction for testing, and comprehensive debug logging for troubleshooting.
