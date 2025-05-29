# Proposal: Sparse FileRepresentation in FileCache

## 1 Motivation & Problem Statement

The current `FileCache` implementation only retains **one read chunk per exact offset** (`pid:fd:pathHash:offset`).
When a write occurs we look for **a single cached read at the same offset**. This is insufficient in many real-world
workloads:

1. **Partial reads** – An application may read the same region in smaller pieces (e.g. 512 B pages).
2. **Overlapping reads** – Buffered I/O frequently re-reads overlapping blocks.
3. **Multi-chunk writes** – A single write can cover a range that spans **several previously cached reads**.

To improve diff quality we need a **sparse in-memory representation** of the file so we can reconstruct all content that
**overlaps** the incoming write.

## 2 Goals & Non-Goals

| Goal | Description                                                                                |
| ---- | ------------------------------------------------------------------------------------------ |
| G1   | Accurately diff writes against **all previously read data** that overlaps the write range. |
| G2   | Maintain bounded memory usage (configurable).                                              |
| G3   | Remain lock-safe and performant under concurrent events.                                   |
| NG1  | Full byte-perfect reconstruction of entire files across program restarts.                  |
| NG2  | Persistent on-disk cache (can be future work).                                             |

## 3 High-Level Approach

1. **Segment the file into ranges** (`[start,end)`). Every read produces one or more segments.
2. **Merge overlapping segments** to avoid duplicate storage.
3. **Index segments per file** (identified by `pathHash` or resolved path) in an ordered structure.
4. On **write event**:
   - Determine write range `[offset, offset+len(write))`.
   - Collect all cached segments that overlap this range.
   - Assemble a **contiguous old-content buffer** (sparse fill with UNKNOWN bytes where data is missing).
   - Run diff algorithms against new write content.
5. **Evict** segments by LRU/TTL and per-file memory limits.

## 4 Data Model

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

## 5 Algorithms

### 5.1 InsertSegment (on read)

```
1. Locate SparseFile for pathHash (create if absent)
2. Binary-search `Segments` to find insertion point
3. Merge with left/right neighbours if they overlap or are adjacent
4. Update Size & LastUsed
5. Enforce per-file & global byte limits (drop oldest segments first)
```

### 5.2 CollectSegments (on write - for diff generation)

```
Input: offset, length
1. Binary-search first segment with End > offset
2. Iterate while Start < offset+length
3. Copy overlapping ranges into buffer
4. If gaps exist, fill with 0x00 or '?' marker (configurable) to keep offsets aligned
```

### 5.3 UpdateSegments (on write - invalidate/update cached data)

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

### 5.4 Cleanup

Periodic goroutine:

- Remove segments older than `maxAge`
- Evict least-recently-used files until total bytes < limit

## 6 Write Handling Strategy

### 6.1 Problem: Stale Cached Data

When a write occurs, any cached segments that overlap the write range become **stale** and must be updated:

```
Before write:
Segments: [10-20: "old data"], [30-40: "more old"]

Write at offset 15, length 10, data "new content"
Range: [15-25)

After write:
Segments: [10-15: "old d"], [15-25: "new content"], [30-40: "more old"]
```

### 6.2 Segment Invalidation Cases

| Case                        | Description                                  | Action                                    |
| --------------------------- | -------------------------------------------- | ----------------------------------------- |
| **Complete overlap**        | Write completely covers a cached segment     | Remove the segment                        |
| **Partial overlap (left)**  | Write overlaps the end of a cached segment   | Truncate segment, keep prefix             |
| **Partial overlap (right)** | Write overlaps the start of a cached segment | Truncate segment, keep suffix             |
| **Split overlap**           | Write is contained within a cached segment   | Split into prefix + suffix, remove middle |

### 6.3 Implementation Details

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

## 7 API Sketch

```go
// Called from read handler
func (fc *FileCache) AddRead(pathHash uint32, offset uint64, data []byte)

// Called from write handler – returns oldContent with same length as newContent
func (fc *FileCache) GetOldContent(pathHash uint32, offset uint64, size uint64) ([]byte, bool)

// Called from write handler after diff generation – updates cached segments
func (fc *FileCache) UpdateWithWrite(pathHash uint32, offset uint64, data []byte)
```

## 8 Concurrency

- Per-file locks (`SparseFile.mu`) avoid global bottlenecks.
- `FileCache.mu` only guards the `files` map (insert/delete), not every segment access.
- Read-mostly workloads benefit from `RLock` on `SparseFile`.

## 9 Memory Management

| Limit          | Default | Notes                                   |
| -------------- | ------- | --------------------------------------- |
| Per-file bytes | 512 KB  | Configurable `--filecache-file-limit`   |
| Global bytes   | 64 MB   | Configurable `--filecache-global-limit` |
| Segment TTL    | 10 min  | Same as existing `maxAge`               |

Eviction order: (1) expired → (2) LRU segments → (3) oldest files.

## 10 Diff Generation Workflow

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

## 11 Integration Plan

1. **Extend FileCache** with `files` map & new APIs (keep old API for backward compatibility).
2. **Redirect StoreReadContent** to `AddRead` plus existing single-offset cache (transitional phase).
3. **Modify Generate\*Diff** helpers to:
   - Call `GetOldContent` before diffing
   - Call `UpdateWithWrite` after diffing
4. **Add config flags** for memory limits & gap-fill character.
5. **Add unit tests** for:
   - Segment merging logic
   - Overlap retrieval
   - Write invalidation logic
   - Eviction policy
6. **Benchmark** memory & CPU with simulated heavy workloads.

## 12 Implementation Status & Testing

### 12.1 Implemented Test Suites

**Test File Organization:**

- [`api_test.go`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache/api_test.go) - API compatibility and parameter validation
- [`segment_test.go`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache/segment_test.go) - Segment merging and insertion logic
- [`write_invalidation_test.go`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache/write_invalidation_test.go) - Write operations and segment invalidation
- [`content_reconstruction_test.go`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache/content_reconstruction_test.go) - Content reconstruction with gaps
- [`cache_management_test.go`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache/cache_management_test.go) - Memory limits, TTL expiration, LRU eviction
- [`concurrency_test.go`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache/concurrency_test.go) - Race condition detection and concurrent access
- [`integration_test.go`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache/integration_test.go) - End-to-end workflow testing
- [`diff_utils_test.go`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache/diff_utils_test.go) - Unified diff generation and parsing

**Basic Read Operations - ✅ IMPLEMENTED:**
• Single read at offset 0 creates first segment ([`TestSegmentMerging/single_segment_to_empty`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache/segment_test.go))
• Multiple non-overlapping reads create separate segments ([`TestMultipleFileWorkflow`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache/integration_test.go))
• Sequential reads merge into single segment ([`TestSegmentMerging/adjacent_segments_merge`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache/segment_test.go))
• Adjacent reads merge correctly ([`TestSegmentMerging/adjacent_segments_merge`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache/segment_test.go))
• Read at arbitrary offset works correctly ([`TestSegmentMerging/*`](file:///home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache/segment_test.go))

**Edge Cases - ✅ IMPLEMENTED:**
• Zero-length read handling (`TestSegmentMergingEdgeCases/zero-length_segment_ignored`)
• Single-byte read creates minimal segment (`TestSegmentMergingEdgeCases/single-byte_segment`)
• Empty and nil data handling (`TestEmptyAndNilData/*`)
• Binary data integrity (`TestSegmentDataIntegrity`)

### 12.2 Segment Merging Logic - ✅ IMPLEMENTED

**Happy Cases - ✅ IMPLEMENTED:**
• Two adjacent segments [0-50] + [50-100] → [0-100] (`TestSegmentMerging/adjacent_segments_merge`)
• Two overlapping segments [0-60] + [40-100] → [0-100] with merged data (`TestSegmentMerging/overlapping_segments_merge`)
• Three segments merge: multiple cascade (`TestSegmentMerging/multiple_segments_merge_cascade`)
• Segments merge in any insertion order (`TestInsertionOrder/*`)

**Edge Cases - ✅ IMPLEMENTED:**
• Identical segments [50-100] + [50-100] → single [50-100] (`TestSegmentMerging/identical_segments_merge`)
• Fully contained segment [20-80] + [30-40] → [20-80] (`TestSegmentMerging/fully_contained_segment_absorbed`)
• Single-byte overlap and edge merging (`TestSegmentMergingEdgeCases/merge_at_boundaries`)
• Zero-length segments ignored (`TestSegmentMergingEdgeCases/zero-length_segment_ignored`)

**Complex Cases - ✅ IMPLEMENTED:**
• Chain merging scenarios (`TestInsertionOrder/chain_merging`)
• Merge cascade detection (`TestInsertionOrder/reverse_order_insertion_merges_correctly`)
• Reverse-order insertion (`TestInsertionOrder/*`)

### 12.3 Write Operations and Invalidation - ✅ IMPLEMENTED

**Happy Cases - ✅ IMPLEMENTED:**
• Write completely replaces existing segment (`TestWriteInvalidation/write_completely_replaces_segment`)
• Write creates new segment in empty cache (covered in various integration tests)
• Write between existing segments fills gap (`TestWriteInvalidation/write_spans_multiple_segments`)

**Edge Cases - Complete Overlap - ✅ IMPLEMENTED:**
• Write exactly matches segment boundaries (`TestWriteInvalidationComplexCases/write_at_exact_segment_boundaries`)
• Write covers multiple complete segments (`TestWriteInvalidationComplexCases/write_covers_multiple_complete_segments`)

**Edge Cases - Partial Overlap - ✅ IMPLEMENTED:**
• Write overlaps segment start (`TestWriteInvalidation/write_overlaps_segment_start`)
• Write overlaps segment end (`TestWriteInvalidation/write_overlaps_segment_end`)
• Write splits segment (`TestWriteInvalidation/write_splits_segment`)
• Write creates holes in existing segments (`TestWriteInvalidationComplexCases/write_creates_holes_in_existing_segments`)

**Complex Overlap Cases - ✅ IMPLEMENTED:**
• Write spans multiple segments with gaps (`TestWriteInvalidation/write_spans_multiple_segments`)
• Write at exact segment boundaries (`TestWriteInvalidationComplexCases/write_at_exact_segment_boundaries`)

### 12.4 Content Reconstruction for Diffing - ✅ IMPLEMENTED

**Happy Cases - ✅ IMPLEMENTED:**
• Request range exactly matches single segment (`TestContentReconstruction/exact_single_segment_match`)
• Request range spans multiple adjacent segments (`TestContentReconstruction/multiple_adjacent_segments`)

**Edge Cases - ✅ IMPLEMENTED:**
• Request range with gaps → fill with placeholder bytes (`TestContentReconstruction/segments_with_gap`)
• Request range partially covered → mix of real data and placeholders (`TestContentReconstruction/partial_coverage_at_*`)
• Request range completely uncovered → all placeholder bytes (`TestContentReconstruction/request_with_no_coverage`)
• Request zero-length range (`TestContentReconstructionEdgeCases/zero_length_request`)
• Request range larger than any cached data (`TestContentReconstructionEdgeCases/request_beyond_all_segments`)

**Gap Handling - ✅ IMPLEMENTED:**
• Single gap in middle (`TestContentReconstruction/segments_with_gap`)
• Multiple gaps (`TestContentReconstructionComplexGaps/multiple_gaps_pattern`)
• Interleaved gaps and data (`TestContentReconstructionComplexGaps/interleaved_gaps_and_data`)
• Complex pattern reconstruction (`TestContentReconstructionComplexGaps/overlapping_reconstruction_window`)

**Complex Reconstruction - ✅ IMPLEMENTED:**
• Gap filling with configurable placeholder bytes (`TestGapFilling`)
• Request spans segments added in different order (covered in various tests)
• Partial segment coverage detection (`TestContentReconstruction/partial_segment_match`)

### 12.5 Cache Lifecycle and Expiration - ✅ IMPLEMENTED

**TTL Expiration - ✅ IMPLEMENTED:**
• Segments expire after maxAge (`TestCacheExpiration`)
• Mixed expired/fresh segments in same file (`TestCacheCleanupEdgeCases/cleanup_with_mixed_expired_and_fresh_data`)
• Cleanup removes only expired segments, keeps fresh ones (`TestCacheCleanupEdgeCases/*`)
• Empty cache cleanup handling (`TestCacheCleanupEdgeCases/cleanup_empty_cache`)

**Memory Limits - ✅ IMPLEMENTED:**
• Per-file limit exceeded → remove oldest segments first (`TestMemoryLimits/per-file_limit`)
• Global limit exceeded → remove least-recently-used files (`TestMemoryLimits/global_limit`)
• LRU eviction policy testing (`TestLRUEviction`)
• Per-file memory management (`TestPerFileMemoryManagement`)

**File Lifecycle - ✅ IMPLEMENTED:**
• Multiple file management (`TestFileHashingAndIdentification`)
• Same pathHash handling for different files (implicit in tests)

### 12.6 Data Integrity and Consistency - ✅ IMPLEMENTED

**Content Verification - ✅ IMPLEMENTED:**
• Content retrieved exactly matches content stored (`TestSegmentDataIntegrity`)
• Binary data preserved correctly (`TestSegmentDataIntegrity`)
• Content with null bytes, special characters (`TestSegmentDataIntegrity`)
• Empty and nil data handling (`TestEmptyAndNilData/*`)

**Boundary Conditions - ✅ IMPLEMENTED:**
• Segment at offset 0 (covered in multiple tests)
• Minimum segment size (1 byte) (`TestSegmentMergingEdgeCases/single-byte_segment`)
• Empty data operations (`TestEmptyAndNilData/*`)

**Data Corruption Scenarios - ✅ IMPLEMENTED:**
• Verify segment ordering remains correct after operations (`TestCacheStateConsistency`)
• Ensure segment boundaries are always valid (`TestCacheStateConsistency`)
• Segment metadata consistency (`TestCacheStateConsistency`)

### 12.7 Concurrency and Thread Safety - ✅ IMPLEMENTED

**Race Condition Prevention - ✅ IMPLEMENTED:**
• Concurrent reads, writes, and retrievals on same file (`TestConcurrentAccess`)
• Concurrent operations on different files (`TestConcurrentFileAccess`)
• Concurrent eviction under memory pressure (`TestConcurrentEviction`)
• Race conditions in segment operations (`TestRaceConditionsInSegmentOperations`)
• Concurrent cleanup operations (`TestConcurrentCleanup`)

### 12.8 Integration Workflows - ✅ IMPLEMENTED

**End-to-End Workflows - ✅ IMPLEMENTED:**
• Complete read→write→diff workflow (`TestFullWorkflow`)
• Multiple file operations (`TestMultipleFileWorkflow`)
• Read-modify-write patterns (`TestReadModifyWritePattern`)
• Gap filling in sparse files (`TestGapFilling`)
• Cache state consistency across operations (`TestCacheStateConsistency`)

### 12.9 API and Compatibility - ✅ IMPLEMENTED

**API Validation - ✅ IMPLEMENTED:**
• API compatibility and parameter validation (`TestAPICompatibility`)
• Constructor parameter validation (`TestNewFileCacheParameters`)
• Time provider interface (`TestRealTimeProvider`)
• File identification and hashing (`TestFileHashingAndIdentification`)
• Empty and nil data handling (`TestEmptyAndNilData`)

### 12.10 Comprehensive Test Coverage Summary

**✅ FULLY IMPLEMENTED (38+ test functions across 8 test files):**

- Segment merging and insertion logic
- Write invalidation and segment splitting
- Content reconstruction with gap filling
- Memory management (per-file and global limits)
- TTL expiration and LRU eviction
- Concurrent access and race condition prevention
- End-to-end integration workflows
- API compatibility and edge cases
- Binary data integrity and boundary conditions
- Cache state consistency validation

**🔧 IMPLEMENTATION STATUS:**

- All core functionality implemented and tested
- Race detector passes all tests
- Comprehensive edge case coverage
- Memory-safe concurrent operations
- Production-ready sparse file cache

## 13 Future Enhancements

- **Persistent backing store** (bolt/LMDB) for long-running monitors.
- **Compression** of stored segments.
- **Hash-only segments** – keep SHA256 for large blocks, lazily fetch bytes if needed.
- **Cross-process correlation** – unify segments for identical path across PIDs.

---

_Author_: <your-name>
_Date_: 2025-05-28
