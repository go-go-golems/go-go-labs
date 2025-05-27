package filecache

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// FileContent represents cached file content
type FileContent struct {
	Content   []byte
	Hash      string
	Timestamp time.Time
	Offset    uint64
	Size      uint64
}

// FileCache caches file contents for diff generation
type FileCache struct {
	cache  map[string]*FileContent // key: "pid:fd" or "hash:offset"
	mu     sync.RWMutex
	maxAge time.Duration
}

// New creates a new file cache
func New() *FileCache {
	return &FileCache{
		cache:  make(map[string]*FileContent),
		maxAge: 10 * time.Minute, // Cache for 10 minutes
	}
}

// StoreReadContent stores content from a read operation
func (fc *FileCache) StoreReadContent(pid uint32, fd int32, pathHash uint32, content []byte, offset uint64) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	key := fmt.Sprintf("%d:%d:%d:%d", pid, fd, pathHash, offset)

	// Calculate hash of content
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
	cachedContent, exists := fc.GetContentForDiff(pid, fd, pathHash, offset)
	if !exists {
		return "", false
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(cachedContent.Content), string(newContent), false)

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

	return dmp.DiffPrettyText(diffs), true
}

// GenerateUnifiedDiff generates a unified diff format with proper line numbers
func (fc *FileCache) GenerateUnifiedDiff(pid uint32, fd int32, pathHash uint32, offset uint64, newContent []byte, filename string) (string, bool) {
	cachedContent, exists := fc.GetContentForDiff(pid, fd, pathHash, offset)
	if !exists {
		return "", false
	}

	dmp := diffmatchpatch.New()
	a, b, c := dmp.DiffLinesToChars(string(cachedContent.Content), string(newContent))
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

	now := time.Now()
	for key, content := range fc.cache {
		if now.Sub(content.Timestamp) > fc.maxAge {
			delete(fc.cache, key)
		}
	}
}

// Size returns the number of cached entries
func (fc *FileCache) Size() int {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return len(fc.cache)
}
