package filecache

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// TimeProvider interface allows mocking time for testing
type TimeProvider interface {
	Now() time.Time
}

// RealTimeProvider implements TimeProvider using real time
type RealTimeProvider struct{}

func (r RealTimeProvider) Now() time.Time {
	return time.Now()
}

// Segment represents a contiguous range of file content
type Segment struct {
	Start   uint64 // inclusive
	End     uint64 // exclusive (Start+len(Data))
	Data    []byte // len == End-Start
	AddedAt time.Time
}

// SparseFile represents a file as a collection of segments
type SparseFile struct {
	mu       sync.RWMutex
	Segments []*Segment // sorted, non-overlapping, non-adjacent (merged)
	Size     uint64     // bytes stored (for limit accounting)
	LastUsed time.Time  // for LRU eviction
}

// FileContent represents cached file content (legacy compatibility)
type FileContent struct {
	Content   []byte
	Hash      string
	Timestamp time.Time
	Offset    uint64
	Size      uint64
}

// FileCache caches file contents for diff generation
type FileCache struct {
	// New sparse file representation
	files map[uint32]*SparseFile // pathHash -> sparse file representation

	// Legacy single-offset cache for backward compatibility
	cache map[string]*FileContent // key: "pid:fd:pathHash:offset"

	mu           sync.RWMutex
	maxAge       time.Duration
	perFileLimit uint64       // bytes per file
	globalLimit  uint64       // total bytes across all files
	totalSize    uint64       // current total bytes stored
	timeProvider TimeProvider // for mocking time in tests
}

// New creates a new file cache
func New() *FileCache {
	return &FileCache{
		files:        make(map[uint32]*SparseFile),
		cache:        make(map[string]*FileContent),
		maxAge:       10 * time.Minute, // Cache for 10 minutes
		perFileLimit: 512 * 1024,       // 512 KB per file
		globalLimit:  64 * 1024 * 1024, // 64 MB total
		timeProvider: RealTimeProvider{},
	}
}

// NewFileCache creates a new file cache with specified limits
func NewFileCache(perFileLimit, globalLimit uint64, maxAge time.Duration, timeProvider TimeProvider) *FileCache {
	return &FileCache{
		files:        make(map[uint32]*SparseFile),
		cache:        make(map[string]*FileContent),
		maxAge:       maxAge,
		perFileLimit: perFileLimit,
		globalLimit:  globalLimit,
		timeProvider: timeProvider,
	}
}

// NewWithTimeProvider creates a new file cache with custom time provider (for testing)
func NewWithTimeProvider(tp TimeProvider) *FileCache {
	fc := New()
	fc.timeProvider = tp
	return fc
}

// StoreReadContent stores content from a read operation
func (fc *FileCache) StoreReadContent(pid uint32, fd int32, pathHash uint32, content []byte, offset uint64) {
	// Store in new sparse cache
	fc.AddRead(pathHash, offset, content)

	// Also store in legacy cache for backward compatibility
	fc.mu.Lock()
	defer fc.mu.Unlock()

	key := fmt.Sprintf("%d:%d:%d:%d", pid, fd, pathHash, offset)

	// Calculate hash of content
	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	fc.cache[key] = &FileContent{
		Content:   make([]byte, len(content)),
		Hash:      hash,
		Timestamp: fc.timeProvider.Now(),
		Offset:    offset,
		Size:      uint64(len(content)),
	}
	copy(fc.cache[key].Content, content)
}

// GetContentForDiff retrieves cached content for diff comparison
func (fc *FileCache) GetContentForDiff(pid uint32, fd int32, pathHash uint32, offset uint64) (*FileContent, bool) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	key := fmt.Sprintf("%d:%d:%d:%d", pid, fd, pathHash, offset)
	content, exists := fc.cache[key]
	if !exists {
		return nil, false
	}

	// Check if content is too old
	if time.Since(content.Timestamp) > fc.maxAge {
		return nil, false
	}

	return content, true
}

// GenerateDiff compares cached content with new write content
func (fc *FileCache) GenerateDiff(pid uint32, fd int32, pathHash uint32, offset uint64, newContent []byte) (string, bool) {
	// Try new sparse cache first
	oldContent, exists := fc.GetOldContent(pathHash, offset, uint64(len(newContent)))
	if !exists {
		// Fall back to legacy cache
		cachedContent, exists := fc.GetContentForDiff(pid, fd, pathHash, offset)
		if !exists {
			return "", false
		}
		oldContent = cachedContent.Content
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(oldContent), string(newContent), false)

	// Only return diff if there are actual changes
	hasChanges := false
	for _, diff := range diffs {
		if diff.Type != diffmatchpatch.DiffEqual {
			hasChanges = true
			break
		}
	}

	if !hasChanges {
		return "", false
	}

	// Update cache with new write content
	fc.UpdateWithWrite(pathHash, offset, newContent)

	return dmp.DiffPrettyText(diffs), true
}

// GenerateUnifiedDiff generates a unified diff format with proper line numbers
func (fc *FileCache) GenerateUnifiedDiff(pid uint32, fd int32, pathHash uint32, offset uint64, newContent []byte, filename string) (string, bool) {
	// Try new sparse cache first
	oldContent, exists := fc.GetOldContent(pathHash, offset, uint64(len(newContent)))
	if !exists {
		// Fall back to legacy cache
		cachedContent, exists := fc.GetContentForDiff(pid, fd, pathHash, offset)
		if !exists {
			return "", false
		}
		oldContent = cachedContent.Content
	}

	dmp := diffmatchpatch.New()
	a, b, c := dmp.DiffLinesToChars(string(oldContent), string(newContent))
	diffs := dmp.DiffMain(a, b, false)
	result := dmp.DiffCharsToLines(diffs, c)

	if len(result) == 1 && result[0].Type == diffmatchpatch.DiffEqual {
		return "", false
	}

	var diff strings.Builder
	diff.WriteString(fmt.Sprintf("--- %s (cached)\n", filename))
	diff.WriteString(fmt.Sprintf("+++ %s (new write)\n", filename))

	oldLineNum := 1
	newLineNum := 1

	for _, d := range result {
		lines := strings.Split(d.Text, "\n")
		for i, line := range lines {
			if i == len(lines)-1 && line == "" {
				continue
			}
			switch d.Type {
			case diffmatchpatch.DiffDelete:
				diff.WriteString(fmt.Sprintf("-%d:%s\n", oldLineNum, line))
				oldLineNum++
			case diffmatchpatch.DiffInsert:
				diff.WriteString(fmt.Sprintf("+%d:%s\n", newLineNum, line))
				newLineNum++
			case diffmatchpatch.DiffEqual:
				diff.WriteString(fmt.Sprintf(" %d:%s\n", newLineNum, line))
				oldLineNum++
				newLineNum++
			}
		}
	}

	// Update cache with new write content
	fc.UpdateWithWrite(pathHash, offset, newContent)

	return diff.String(), true
}

// GenerateElidedUnifiedDiff generates a unified diff with context line limiting
func (fc *FileCache) GenerateElidedUnifiedDiff(pid uint32, fd int32, pathHash uint32, offset uint64, newContent []byte, filename string, contextLines int) (string, bool) {
	// Generate the full diff first
	fullDiff, hasDiff := fc.GenerateUnifiedDiff(pid, fd, pathHash, offset, newContent, filename)
	if !hasDiff {
		return "", false
	}

	// Apply elision if context lines is specified and > 0
	if contextLines > 0 {
		elidedDiff := ElideUnifiedDiff(fullDiff, contextLines)
		return elidedDiff, true
	}

	return fullDiff, true
}

// UpdateWriteContent updates the cache with new write content
func (fc *FileCache) UpdateWriteContent(pid uint32, fd int32, pathHash uint32, content []byte, offset uint64) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	key := fmt.Sprintf("%d:%d:%d:%d", pid, fd, pathHash, offset)

	// Update or create new entry with write content
	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	fc.cache[key] = &FileContent{
		Content:   make([]byte, len(content)),
		Hash:      hash,
		Timestamp: time.Now(),
		Offset:    offset,
		Size:      uint64(len(content)),
	}
	copy(fc.cache[key].Content, content)
}

// Cleanup removes old entries from the cache
func (fc *FileCache) Cleanup() {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	now := fc.timeProvider.Now()

	// Clean up legacy cache
	for key, content := range fc.cache {
		if now.Sub(content.Timestamp) > fc.maxAge {
			delete(fc.cache, key)
		}
	}

	// Clean up sparse files
	for pathHash, sf := range fc.files {
		sf.mu.Lock()

		// Remove expired segments
		validSegments := make([]*Segment, 0, len(sf.Segments))
		for _, seg := range sf.Segments {
			if now.Sub(seg.AddedAt) <= fc.maxAge {
				validSegments = append(validSegments, seg)
			} else {
				sf.Size -= uint64(len(seg.Data))
				fc.totalSize -= uint64(len(seg.Data))
			}
		}
		sf.Segments = validSegments

		// Remove empty files
		if len(sf.Segments) == 0 {
			sf.mu.Unlock()
			delete(fc.files, pathHash)
		} else {
			sf.mu.Unlock()
		}
	}
}

// insertSegment adds a segment to the sparse file, merging with adjacent/overlapping segments
func (sf *SparseFile) insertSegment(newSeg *Segment, timeProvider TimeProvider) {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	if newSeg.Start >= newSeg.End {
		return // Invalid segment
	}

	// Simple case: no existing segments
	if len(sf.Segments) == 0 {
		sf.Segments = []*Segment{newSeg}
		sf.Size += uint64(len(newSeg.Data))
		sf.LastUsed = time.Now()
		return
	}

	// Find all segments that overlap or are adjacent to the new segment
	var toMerge []*Segment
	var toKeep []*Segment

	for _, seg := range sf.Segments {
		// Check if segments overlap or are adjacent
		if seg.End >= newSeg.Start && seg.Start <= newSeg.End {
			toMerge = append(toMerge, seg)
			sf.Size -= uint64(len(seg.Data)) // Remove from size accounting
		} else {
			toKeep = append(toKeep, seg)
		}
	}

	// Add the new segment to the merge list
	toMerge = append(toMerge, newSeg)

	// Find the overall bounds
	start := newSeg.Start
	end := newSeg.End

	for _, seg := range toMerge {
		if seg.Start < start {
			start = seg.Start
		}
		if seg.End > end {
			end = seg.End
		}
	}

	// Create merged data
	mergedData := make([]byte, end-start)

	// Fill with data from all segments, with later segments overwriting earlier ones
	for _, seg := range toMerge {
		relStart := seg.Start - start
		copy(mergedData[relStart:relStart+uint64(len(seg.Data))], seg.Data)
	}

	// Create the final merged segment
	finalSeg := &Segment{
		Start:   start,
		End:     end,
		Data:    mergedData,
		AddedAt: timeProvider.Now(),
	}

	// Update size accounting
	sf.Size += uint64(len(finalSeg.Data))

	// Add the merged segment to the keep list
	toKeep = append(toKeep, finalSeg)

	// Sort segments by start position
	sort.Slice(toKeep, func(i, j int) bool {
		return toKeep[i].Start < toKeep[j].Start
	})

	sf.Segments = toKeep
	sf.LastUsed = timeProvider.Now()
}

// UpdateWithWrite invalidates segments that overlap with the write and adds the new write segment
func (sf *SparseFile) UpdateWithWrite(offset uint64, data []byte, timeProvider TimeProvider) {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	writeEnd := offset + uint64(len(data))
	// Debug logging would go here but we need a way to conditionally enable it

	// Find all segments that overlap [offset, writeEnd)
	var toRemove []int
	var toAdd []*Segment

	for i, seg := range sf.Segments {
		if seg.End <= offset || seg.Start >= writeEnd {
			continue // No overlap
		}

		// Mark for removal
		toRemove = append(toRemove, i)

		// Keep non-overlapping parts only if they exist
		if seg.Start < offset {
			// Keep prefix [seg.Start, offset)
			prefixLen := offset - seg.Start
			toAdd = append(toAdd, &Segment{
				Start:   seg.Start,
				End:     offset,
				Data:    make([]byte, prefixLen),
				AddedAt: timeProvider.Now(),
			})
			copy(toAdd[len(toAdd)-1].Data, seg.Data[:prefixLen])
		}

		if seg.End > writeEnd {
			// Keep suffix [writeEnd, seg.End) only if there's actual content
			suffixStart := writeEnd - seg.Start
			suffixLen := seg.End - writeEnd
			if suffixLen > 0 && suffixStart < uint64(len(seg.Data)) {
				toAdd = append(toAdd, &Segment{
					Start:   writeEnd,
					End:     seg.End,
					Data:    make([]byte, suffixLen),
					AddedAt: timeProvider.Now(),
				})
				copy(toAdd[len(toAdd)-1].Data, seg.Data[suffixStart:])
			}
		}

		// Update size accounting for removed segment
		sf.Size -= uint64(len(seg.Data))
	}

	// Remove overlapping segments (in reverse order to maintain indices)
	for i := len(toRemove) - 1; i >= 0; i-- {
		idx := toRemove[i]
		sf.Segments = append(sf.Segments[:idx], sf.Segments[idx+1:]...)
	}

	// Add the new write segment
	newSeg := &Segment{
		Start:   offset,
		End:     writeEnd,
		Data:    make([]byte, len(data)),
		AddedAt: timeProvider.Now(),
	}
	copy(newSeg.Data, data)
	toAdd = append(toAdd, newSeg)

	// Add all new segments to the segments list
	sf.Segments = append(sf.Segments, toAdd...)

	// Sort segments by start position
	sort.Slice(sf.Segments, func(i, j int) bool {
		return sf.Segments[i].Start < sf.Segments[j].Start
	})
}

// GetContentRange reconstructs content for the given range, filling gaps with gapByte
func (sf *SparseFile) GetContentRange(offset, length uint64, gapByte byte) ([]byte, bool) {
	sf.mu.RLock()
	defer sf.mu.RUnlock()

	if length == 0 {
		return []byte{}, false
	}

	endOffset := offset + length
	result := make([]byte, length)
	filled := make([]bool, length)

	// Fill with gap bytes initially
	for i := range result {
		result[i] = gapByte
	}

	hasData := false

	// Find segments that overlap with the requested range
	for _, seg := range sf.Segments {
		if seg.End <= offset || seg.Start >= endOffset {
			continue // No overlap
		}

		hasData = true

		// Calculate overlap
		overlapStart := maxUint64(seg.Start, offset)
		overlapEnd := minUint64(seg.End, endOffset)

		// Copy data from segment
		segDataStart := overlapStart - seg.Start
		resultStart := overlapStart - offset
		copyLen := overlapEnd - overlapStart

		copy(result[resultStart:resultStart+copyLen], seg.Data[segDataStart:segDataStart+copyLen])
		for i := resultStart; i < resultStart+copyLen; i++ {
			filled[i] = true
		}
	}

	// Note: sf.LastUsed should be updated by caller
	return result, hasData
}

// AddRead adds a read segment to the sparse file representation
func (fc *FileCache) AddRead(pathHash uint32, offset uint64, data []byte) {
	// Ensure we always follow the fc.mu -> sf.mu lock ordering to avoid
	// circular wait with global eviction routines.
	fc.mu.Lock()

	sf, exists := fc.files[pathHash]
	if !exists {
		sf = &SparseFile{
			Segments: make([]*Segment, 0),
			LastUsed: fc.timeProvider.Now(),
		}
		fc.files[pathHash] = sf
	}

	// Create new segment while still holding fc.mu so the subsequent sf lock
	// respects ordering.
	newSeg := &Segment{
		Start:   offset,
		End:     offset + uint64(len(data)),
		Data:    make([]byte, len(data)),
		AddedAt: fc.timeProvider.Now(),
	}
	copy(newSeg.Data, data)

	// Perform insertion (insertSegment obtains sf.mu internally); since we
	// still hold fc.mu, the lock acquisition order remains fc.mu -> sf.mu.
	sf.insertSegment(newSeg, fc.timeProvider)

	// Update global size accounting while fc.mu is still held.
	fc.totalSize += uint64(len(data))

	fc.mu.Unlock()

	// Enforce limits (these functions manage their own locking).
	fc.enforcePerFileLimit(sf)
	fc.enforceGlobalLimit()
}

// GetOldContent retrieves cached content for diff comparison
func (fc *FileCache) GetOldContent(pathHash uint32, offset uint64, size uint64) ([]byte, bool) {
	fc.mu.RLock()
	sf, exists := fc.files[pathHash]
	fc.mu.RUnlock()

	if !exists {
		return nil, false
	}

	result, exists := sf.GetContentRange(offset, size, 0x00)
	if exists {
		sf.mu.Lock()
		sf.LastUsed = fc.timeProvider.Now()
		sf.mu.Unlock()
	}
	return result, exists
}

// UpdateWithWrite updates the cache with new write content
func (fc *FileCache) UpdateWithWrite(pathHash uint32, offset uint64, data []byte) {
	// Acquire cache-level lock only for map access / size update.
	fc.mu.Lock()

	sf, exists := fc.files[pathHash]
	if !exists {
		sf = &SparseFile{
			Segments: make([]*Segment, 0),
			LastUsed: fc.timeProvider.Now(),
		}
		fc.files[pathHash] = sf
	}
	// fc.mu.Unlock()  // removed to maintain lock ordering

	// NOTE: maintain lock ordering by keeping fc.mu locked until size update

	// Perform the write while holding fc.mu so that lock acquisition order
	sf.UpdateWithWrite(offset, data, fc.timeProvider)

	// Update global size accounting before releasing cache mutex.
	fc.totalSize += uint64(len(data))

	fc.mu.Unlock()

	// Enforce limits (these functions acquire the required locks themselves).
	fc.enforcePerFileLimit(sf)
	fc.enforceGlobalLimit()
}

// Helper functions
func maxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func minUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// enforcePerFileLimit removes oldest segments if file exceeds per-file limit
func (fc *FileCache) enforcePerFileLimit(sf *SparseFile) {
	// Lock order: acquire cache lock first, then the sparse file lock to
	// maintain the global fc.mu -> sf.mu hierarchy and prevent deadlocks.
	fc.mu.Lock()
	sf.mu.Lock()

	removedBytes := uint64(0)

	for sf.Size > fc.perFileLimit && len(sf.Segments) > 0 {
		// Find oldest segment within the file.
		oldestIdx := 0
		for i := 1; i < len(sf.Segments); i++ {
			if sf.Segments[i].AddedAt.Before(sf.Segments[oldestIdx].AddedAt) {
				oldestIdx = i
			}
		}

		// Remove oldest segment and update accounting.
		removed := sf.Segments[oldestIdx]
		segSize := uint64(len(removed.Data))
		sf.Size -= segSize
		removedBytes += segSize

		sf.Segments = append(sf.Segments[:oldestIdx], sf.Segments[oldestIdx+1:]...)
	}

	sf.mu.Unlock()
	// Update global size outside of sf.mu but still within fc.mu.
	if removedBytes > 0 {
		fc.totalSize -= removedBytes
	}
	fc.mu.Unlock()
}

// enforceGlobalLimit removes least recently used files if global limit exceeded
func (fc *FileCache) enforceGlobalLimit() {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	for fc.totalSize > fc.globalLimit && len(fc.files) > 0 {
		// Find least recently used file
		var oldestPathHash uint32
		var oldestTime time.Time = fc.timeProvider.Now()

		for pathHash, sf := range fc.files {
			sf.mu.RLock()
			if sf.LastUsed.Before(oldestTime) {
				oldestTime = sf.LastUsed
				oldestPathHash = pathHash
			}
			sf.mu.RUnlock()
		}

		// Remove entire file
		if sf, exists := fc.files[oldestPathHash]; exists {
			sf.mu.Lock()
			for _, seg := range sf.Segments {
				fc.totalSize -= uint64(len(seg.Data))
			}
			sf.mu.Unlock()
			delete(fc.files, oldestPathHash)
		}
	}
}

// Size returns the number of cached entries
func (fc *FileCache) Size() int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	size := len(fc.cache) // Legacy cache entries
	for _, sf := range fc.files {
		sf.mu.RLock()
		size += len(sf.Segments)
		sf.mu.RUnlock()
	}
	return size
}
