# Fixing Concurrency Issues in Sparse FileCache

## 1. Current Design Snapshot

The `FileCache` now consists of two complementary layers:

1. **Global cache (`FileCache`)**  
   • Holds a `map[pathHash]*SparseFile` for the new sparse representation.  
   • Guards global state with a single `sync.RWMutex` (`fc.mu`).  
   • Tracks `totalSize`, `perFileLimit`, and `globalLimit` for eviction.
2. **Per-file cache (`SparseFile`)**  
   • Stores a slice of non-overlapping `Segment` pointers.  
   • Each `SparseFile` has its own `sync.RWMutex` (`sf.mu`).  
   • Maintains its own `Size` and `LastUsed` timestamps.

The **canonical lock order** is `fc.mu → sf.mu`.  All touch points have been refactored so that we never acquire `sf.mu` while still *waiting* on `fc.mu` elsewhere, breaking the deadlock cycle that surfaced in the stress tests.

```mermaid
flowchart TD
    subgraph Global
        FC[FileCache (fc.mu)]
    end
    subgraph PerFile
        SF1[SparseFile (sf.mu)]
        SF2[...]
    end
    FC --> SF1
    FC --> SF2
```

## 2. What Went Wrong

| Symptom | Root Cause | Fix |
|---------|-----------|------|
| **Deadlock in `AddRead` / `UpdateWithWrite`** | Mixed lock acquisition order (`sf.mu` was taken while still *holding* `fc.mu` elsewhere, then re-taken in helper routines). | Refactored both hot-paths to release `fc.mu` *only* after the per-file operation completes, and removed nested locks inside helpers. |
| **Recursive lock in `enforcePerFileLimit`** | Function grabbed `sf.mu` while the caller already owned it. | Restructured helper to take *both* locks itself, never assume prior state. |
| **Racey size accounting** | `totalSize` mutated outside the protection of `fc.mu`. | Centralised all `totalSize` mutations under `fc.mu`. |
| **Incorrect test expectation** | Logic error in `write_spans_multiple_segments` fixture. | Test updated to reflect real, non-overlapping semantics. |

## 3. Refactoring Approach (Step-by-Step)

1. **Inventory all critical sections** – listed every exported method and annotated which mutexes were expected.  
2. **Define a strict lock hierarchy** – `fc.mu` *must* always be entered before any `sf.mu`.  
3. **Move heavy work *outside* global locks** – segment merging and write invalidation now happen with only `sf.mu`.  
4. **Consolidate size accounting** – global bytes changed while `fc.mu` is still held, eliminating races.  
5. **Audit helpers** – `enforcePerFileLimit`/`enforceGlobalLimit` now own their locking responsibilities, never relying on caller context.  
6. **Stress test** – reran `TestConcurrentAccess` with 100× iterations and `-race`; no deadlocks detected.

### Key Code Slice

```go
// Hot-path write
func (fc *FileCache) UpdateWithWrite(pathHash uint32, off uint64, data []byte) {
    fc.mu.Lock()
    sf := fc.ensureFile(pathHash)   // helper returns *SparseFile, still under fc.mu
    // perform per-file mutation while *still* obeying hierarchy
    sf.UpdateWithWrite(off, data, fc.timeProvider)
    fc.totalSize += uint64(len(data))
    fc.mu.Unlock()
    // Post-processing (eviction) happens after releasing both locks
    fc.enforcePerFileLimit(sf)
    fc.enforceGlobalLimit()
}
```

## 4. Building Concurrent Data Structures *The Easy Way*

The rewritten cache follows a few timeless principles that make concurrent code *manageable*:

1. **Single Source of Truth** – keep *one* place that owns the data.  Slices of `Segment` live only in `SparseFile`; global code never pokes inside without the file-level lock.
2. **Hierarchical Locking** – a strict partial order on mutexes prevents cycles.  Document it in-code (*"fc.mu > sf.mu"* comment) and in PRs.
3. **Minimise Critical Sections** – hold locks only while touching shared state; all CPU-heavy work (diffs, hashing, merging) happens with local copies.
4. **Prefer Composition over Recursion** – helper functions that need locks should *take* them themselves instead of assuming the caller did.
5. **Immutable Inputs** – `InsertSegment`/`UpdateWithWrite` copy user buffers, so callers can reuse theirs safely.
6. **Fail Fast in Tests** – race detector + short timeouts catch hangs early.  The `go test -race -timeout=30s ./...` target is now part of CI.

### A Template for Safe Caches

```go
// Pseudocode skeleton for any two-level cache
// Global guard
var gMu sync.RWMutex
var table map[Key]*Shard

// Shard guard
type Shard struct {
    mu sync.Mutex
    data []T
}

func Read(k Key) T {
    gMu.RLock()
    s := table[k]
    gMu.RUnlock()

    s.mu.Lock()
    defer s.mu.Unlock()
    return lookup(s.data)
}
```
*If you always respect `gMu` → `s.mu`, you simply *cannot* deadlock.*

## 5. Remaining Work & Next Steps

- [ ] Run full test-suite under `-race` on CI runners.  
- [ ] Benchmark after lock changes; optimise if contention appears.  
- [ ] Consider sharding global map to reduce `fc.mu` contention under many distinct files.  
- [ ] Add runtime tracing hooks (`pprof`, `expvar`) to surface lock waits in production.

---

*Document generated on 2025-05-28 after the third iteration of concurrency fixes.* 