# Proposal: Sparse FileRepresentation in FileCache

## 1  Motivation & Problem Statement

The current `FileCache` implementation only retains **one read chunk per exact offset** (`pid:fd:pathHash:offset`).
When a write occurs we look for **a single cached read at the same offset**. This is insufficient in many real-world
workloads:

1. **Partial reads** – An application may read the same region in smaller pieces (e.g. 512 B pages).
2. **Overlapping reads** – Buffered I/O frequently re-reads overlapping blocks.
3. **Multi-chunk writes** – A single write can cover a range that spans **several previously cached reads**.

To improve diff quality we need a **sparse in-memory representation** of the file so we can reconstruct all content that
**overlaps** the incoming write.

## 2  Goals & Non-Goals

| Goal | Description |
|------|-------------|
| G1  | Accurately diff writes against **all previously read data** that overlaps the write range. |
| G2  | Maintain bounded memory usage (configurable). |
| G3  | Remain lock-safe and performant under concurrent events. |
| NG1 | Full byte-perfect reconstruction of entire files across program restarts. |
| NG2 | Persistent on-disk cache (can be future work). |

## 3  High-Level Approach

1. **Segment the file into ranges** (`[start,end)`). Every read produces one or more segments.
2. **Merge overlapping segments** to avoid duplicate storage.
3. **Index segments per file** (identified by `pathHash` or resolved path) in an ordered structure.
4. On **write event**:
   - Determine write range `[offset, offset+len(write))`.
   - Collect all cached segments that overlap this range.
   - Assemble a **contiguous old-content buffer** (sparse fill with UNKNOWN bytes where data is missing).
   - Run diff algorithms against new write content.
5. **Evict** segments by LRU/TTL and per-file memory limits.

## 4  Data Model

```go
// Keyed by pathHash (uint32)  -> *SparseFile
// Inside SparseFile we track segments sorted by start offset.

type Segment struct {
    Start   uint64   // inclusive
    End     uint64   // exclusive (Start+len(Data))
    Data    []byte   // len == End-Start
    AddedAt time.Time
}

type SparseFile struct {
    mu       sync.RWMutex
    Segments []*Segment       // sorted, non-overlapping, non-adjacent (merged)
    Size     uint64           // bytes stored (for limit accounting)
    LastUsed time.Time        // for LRU eviction
}
```

### Global Index in FileCache

```go
// inside FileCache
files map[uint32]*SparseFile // pathHash -> sparse file representation
```

## 5  Algorithms

### 5.1  InsertSegment (on read)

```
1. Locate SparseFile for pathHash (create if absent)
2. Binary-search `Segments` to find insertion point
3. Merge with left/right neighbours if they overlap or are adjacent
4. Update Size & LastUsed
5. Enforce per-file & global byte limits (drop oldest segments first)
```

### 5.2  CollectSegments (on write - for diff generation)

```
Input: offset, length
1. Binary-search first segment with End > offset
2. Iterate while Start < offset+length
3. Copy overlapping ranges into buffer
4. If gaps exist, fill with 0x00 or '?' marker (configurable) to keep offsets aligned
```

### 5.3  UpdateSegments (on write - invalidate/update cached data)

```
Input: offset, length, newData
1. Binary-search segments that overlap [offset, offset+length)
2. For each overlapping segment:
   a. If segment is completely covered by write: remove it
   b. If segment partially overlaps:
      - Split segment at write boundaries
      - Keep non-overlapping parts
      - Remove overlapping parts
3. Insert new segment with write data at [offset, offset+length)
4. Merge adjacent segments if possible
5. Update Size & LastUsed
```

### 5.4  Cleanup

Periodic goroutine:
- Remove segments older than `maxAge`
- Evict least-recently-used files until total bytes < limit

## 6  Write Handling Strategy

### 6.1  Problem: Stale Cached Data

When a write occurs, any cached segments that overlap the write range become **stale** and must be updated:

```
Before write:
Segments: [10-20: "old data"], [30-40: "more old"]

Write at offset 15, length 10, data "new content"
Range: [15-25)

After write:
Segments: [10-15: "old d"], [15-25: "new content"], [30-40: "more old"]
```

### 6.2  Segment Invalidation Cases

| Case | Description | Action |
|------|-------------|---------|
| **Complete overlap** | Write completely covers a cached segment | Remove the segment |
| **Partial overlap (left)** | Write overlaps the end of a cached segment | Truncate segment, keep prefix |
| **Partial overlap (right)** | Write overlaps the start of a cached segment | Truncate segment, keep suffix |
| **Split overlap** | Write is contained within a cached segment | Split into prefix + suffix, remove middle |

### 6.3  Implementation Details

```go
func (sf *SparseFile) UpdateWithWrite(offset uint64, data []byte) {
    writeEnd := offset + uint64(len(data))
    
    // Find all segments that overlap [offset, writeEnd)
    var toRemove []int
    var toAdd []*Segment
    
    for i, seg := range sf.Segments {
        if seg.End <= offset || seg.Start >= writeEnd {
            continue // No overlap
        }
        
        // Mark for removal
        toRemove = append(toRemove, i)
        
        // Keep non-overlapping parts
        if seg.Start < offset {
            // Keep prefix [seg.Start, offset)
            prefixLen := offset - seg.Start
            toAdd = append(toAdd, &Segment{
                Start:   seg.Start,
                End:     offset,
                Data:    seg.Data[:prefixLen],
                AddedAt: time.Now(),
            })
        }
        
        if seg.End > writeEnd {
            // Keep suffix [writeEnd, seg.End)
            suffixStart := writeEnd - seg.Start
            toAdd = append(toAdd, &Segment{
                Start:   writeEnd,
                End:     seg.End,
                Data:    seg.Data[suffixStart:],
                AddedAt: time.Now(),
            })
        }
    }
    
    // Remove overlapping segments (in reverse order to maintain indices)
    for i := len(toRemove) - 1; i >= 0; i-- {
        sf.removeSegment(toRemove[i])
    }
    
    // Add the new write segment
    newSeg := &Segment{
        Start:   offset,
        End:     writeEnd,
        Data:    make([]byte, len(data)),
        AddedAt: time.Now(),
    }
    copy(newSeg.Data, data)
    toAdd = append(toAdd, newSeg)
    
    // Insert all new segments
    for _, seg := range toAdd {
        sf.insertSegment(seg)
    }
}
```

## 7  API Sketch

```go
// Called from read handler
func (fc *FileCache) AddRead(pathHash uint32, offset uint64, data []byte)

// Called from write handler – returns oldContent with same length as newContent
func (fc *FileCache) GetOldContent(pathHash uint32, offset uint64, size uint64) ([]byte, bool)

// Called from write handler after diff generation – updates cached segments
func (fc *FileCache) UpdateWithWrite(pathHash uint32, offset uint64, data []byte)
```

## 8  Concurrency

* Per-file locks (`SparseFile.mu`) avoid global bottlenecks.
* `FileCache.mu` only guards the `files` map (insert/delete), not every segment access.
* Read-mostly workloads benefit from `RLock` on `SparseFile`.

## 9  Memory Management

| Limit | Default | Notes |
|-------|---------|-------|
| Per-file bytes  | 512 KB | Configurable `--filecache-file-limit` |
| Global bytes    | 64 MB  | Configurable `--filecache-global-limit` |
| Segment TTL     | 10 min | Same as existing `maxAge` |

Eviction order: (1) expired → (2) LRU segments → (3) oldest files.

## 10  Diff Generation Workflow

```
Write Event (offset O, len L, data D)
    │
    ├─▶ cache.GetOldContent(pathHash,O,L)  ⟶  oldBuf []byte (len L)
    │        └─ collects/merges overlapping segments
    │
    ├─▶ diff(oldBuf, newWriteBuf)
    │        ├─ quick hash equality check (optional)
    │        └─ diffmatchpatch / unified diff utils
    │
    └─▶ cache.UpdateWithWrite(pathHash,O,D)  ⟶  invalidate/update segments
             └─ removes stale segments, adds new write segment
```

**Key insight**: We generate the diff **before** updating the cache, so we can compare against the old state.

If `GetOldContent` returns `false` (no overlap), we still call `UpdateWithWrite` to cache the new content.

## 11  Integration Plan

1. **Extend FileCache** with `files` map & new APIs (keep old API for backward compatibility).
2. **Redirect StoreReadContent** to `AddRead` plus existing single-offset cache (transitional phase).
3. **Modify Generate*Diff** helpers to:
   - Call `GetOldContent` before diffing
   - Call `UpdateWithWrite` after diffing
4. **Add config flags** for memory limits & gap-fill character.
5. **Add unit tests** for:
   - Segment merging logic
   - Overlap retrieval
   - Write invalidation logic
   - Eviction policy
6. **Benchmark** memory & CPU with simulated heavy workloads.

## 12  Comprehensive Testing Scenarios

### 12.1  Basic Read Operations

**Happy Cases:**
• Single read at offset 0 creates first segment
• Multiple non-overlapping reads create separate segments  
• Sequential reads (0-100, 100-200, 200-300) merge into single segment
• Adjacent reads (0-50, 50-100) merge correctly
• Read at arbitrary offset (e.g., 1000-1100) works correctly

**Edge Cases:**
• Zero-length read (should be ignored or handled gracefully)
• Single-byte read creates minimal segment
• Maximum-size read (128KB) stores correctly
• Read at maximum file offset (near uint64 limit)
• Read with empty data array

### 12.2  Segment Merging Logic

**Happy Cases:**
• Two adjacent segments [0-50] + [50-100] → [0-100]
• Two overlapping segments [0-60] + [40-100] → [0-100] with merged data
• Three segments merge: [0-30] + [20-50] + [40-80] → [0-80]
• Segments merge in any insertion order (left-to-right, right-to-left, middle-first)

**Edge Cases:**
• Identical segments [50-100] + [50-100] → single [50-100]
• Fully contained segment [20-80] + [30-40] → [20-80] (inner absorbed)
• Single-byte overlap [0-50] + [49-100] → [0-100]
• Multiple tiny segments (1-byte each) merge into larger segment
• Segments with identical start but different end: [10-50] + [10-80] → [10-80]
• Segments with identical end but different start: [10-50] + [30-50] → [10-50]

**Complex Cases:**
• Chain merging: insert [40-60], then [20-40], then [60-80] → final [20-80]
• Merge cascade: inserting one segment triggers multiple merges
• Interleaved segments: [0-10], [20-30], [40-50], then [5-45] merges all
• Reverse-order insertion: insert [80-90], [60-70], [40-50], [20-30], [0-100]

### 12.3  Write Operations and Invalidation

**Happy Cases:**
• Write completely replaces existing segment: [10-50] write [10-50] → new content
• Write creates new segment in empty cache
• Write at end of file extends cached representation
• Write between existing segments fills gap

**Edge Cases - Complete Overlap:**
• Write exactly matches segment boundaries [20-40] over [20-40]
• Write covers multiple complete segments: write [10-80] over [20-30], [40-50], [60-70]
• Write covers single segment plus gaps: write [15-85] over [20-30], [60-70]

**Edge Cases - Partial Overlap:**
• Write overlaps segment start: write [15-35] over [20-50] → keep [35-50]
• Write overlaps segment end: write [30-60] over [20-40] → keep [20-30]
• Write splits segment: write [25-35] over [20-50] → keep [20-25] + [35-50]
• Write extends beyond segment: write [30-70] over [20-40] → keep [20-30]

**Complex Overlap Cases:**
• Write spans multiple segments with gaps: write [15-85] over [20-30], [40-50], [70-80]
• Write partially overlaps multiple segments: write [25-65] over [20-40], [60-80]
• Write creates holes: write [30-40] over [20-60] → [20-30] + [40-60]
• Cascading splits: write affects segment that was result of previous merge
• Write at segment boundaries (exactly at start/end of existing segments)

### 12.4  Content Reconstruction for Diffing

**Happy Cases:**
• Request range exactly matches single segment
• Request range spans multiple adjacent segments
• Request range covered by overlapping segments (use latest data)

**Edge Cases:**
• Request range with gaps → fill with placeholder bytes
• Request range partially covered → mix of real data and placeholders
• Request range completely uncovered → all placeholder bytes
• Request zero-length range
• Request range larger than any cached data

**Gap Handling:**
• Single gap in middle: segments [0-20], [40-60], request [10-50]
• Multiple gaps: segments [0-10], [30-40], [70-80], request [0-80]
• Gap at start: segments [20-40], request [0-50]
• Gap at end: segments [0-20], request [0-50]
• Interleaved gaps and data: complex pattern reconstruction

**Complex Reconstruction:**
• Overlapping segments with different timestamps (use newest)
• Request spans segments added in different order
• Partial segment coverage with multiple gap sizes
• Request range extends beyond all cached data
• Segments with different data at same offset (conflict resolution)

### 12.5  Cache Lifecycle and Expiration

**TTL Expiration:**
• Segments expire after maxAge (10 minutes default)
• Mixed expired/fresh segments in same file
• All segments expired for a file
• Expired segments don't participate in diff generation
• Cleanup removes only expired segments, keeps fresh ones

**Memory Limits:**
• Per-file limit exceeded → remove oldest segments first
• Global limit exceeded → remove least-recently-used files
• Eviction during active read/write operations
• Eviction preserves most recently used data
• Eviction handles edge case where single segment exceeds per-file limit

**File Lifecycle:**
• File created, used, then abandoned (segments eventually expire)
• File deleted but segments remain until TTL
• Same pathHash reused for different files (hash collision)
• File accessed across multiple processes (different PIDs)

### 12.6  Data Integrity and Consistency

**Content Verification:**
• SHA256 hashes match stored content
• Content retrieved exactly matches content stored
• Binary data preserved correctly (no encoding issues)
• Large content blocks (multi-KB) stored/retrieved correctly
• Content with null bytes, special characters, unicode

**Boundary Conditions:**
• Segment at offset 0
• Segment at maximum offset (near uint64 limit)
• Maximum segment size (128KB)
• Minimum segment size (1 byte)
• Empty file operations

**Data Corruption Scenarios:**
• Detect if segment data gets corrupted in memory
• Handle gracefully if segment metadata becomes inconsistent
• Verify segment ordering remains correct after operations
• Ensure segment boundaries are always valid (Start < End)

### 12.7  File Path and Identification

**PathHash Handling:**
• Same pathHash used across multiple operations
• Different pathHashes for different files
• Hash collisions (different files, same hash)
• PathHash changes for same file (symlinks, renames)

**Multi-Process Scenarios:**
• Same file accessed by different PIDs
• Different files with same name in different directories
• Process dies, new process reuses PID
• File descriptor reuse across processes

### 12.8  Diff Generation Integration

**Diff Input Validation:**
• Old content exactly matches write range
• Old content shorter than write range (partial coverage)
• Old content longer than write range (over-coverage)
• No old content available (cache miss)

**Diff Content Types:**
• Text files with line endings (Unix, Windows, Mac)
• Binary files with null bytes
• Mixed text/binary content
• Very large diffs (multi-KB changes)
• Identical content (no diff needed)

**Diff Edge Cases:**
• Write identical content (no actual change)
• Write completely different content (full replacement)
• Write that only changes whitespace
• Write that adds/removes content at boundaries
• Write with encoding changes (UTF-8, ASCII, etc.)

### 12.9  API Contract Validation

**AddRead Validation:**
• Handles nil/empty data gracefully
• Validates offset ranges
• Processes duplicate reads correctly
• Handles reads in any order (not necessarily sequential)

**GetOldContent Validation:**
• Returns correct length buffer (matches requested size)
• Handles requests beyond cached data
• Returns false when no overlap exists
• Fills gaps consistently with configured placeholder

**UpdateWithWrite Validation:**
• Correctly invalidates all overlapping segments
• Preserves non-overlapping segment parts
• Handles writes that create new segments
• Maintains segment ordering after updates
• Updates file metadata (size, last-used) correctly

### 12.10  Complex Integration Scenarios

**Read-Write Patterns:**
• Read file, modify small portion, write back
• Read multiple chunks, write overlapping region
• Interleaved reads and writes to same file
• Read-modify-write cycles with growing file
• Random access patterns (non-sequential reads/writes)

**File Modification Patterns:**
• Append-only writes (log files)
• In-place edits (configuration files)
• Truncation followed by new content
• Sparse file operations (writes with large gaps)
• File growth and shrinkage patterns

**Cache State Transitions:**
• Empty cache → first read → first write → subsequent operations
• Cache with data → file deleted → new file with same path
• Cache eviction during active file operations
• Cache cleanup during heavy read/write activity
• Recovery from various cache corruption scenarios

## 13  Future Enhancements

* **Persistent backing store** (bolt/LMDB) for long-running monitors.
* **Compression** of stored segments.
* **Hash-only segments** – keep SHA256 for large blocks, lazily fetch bytes if needed.
* **Cross-process correlation** – unify segments for identical path across PIDs.

---

*Author*: <your-name>
*Date*: 2025-05-28 