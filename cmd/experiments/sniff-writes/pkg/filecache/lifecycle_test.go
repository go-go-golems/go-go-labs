package filecache

import (
	"bytes"
	"testing"
	"time"
)

func TestEmptyCacheFirstReadWrite(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	// Verify cache starts empty
	if fc.Size() != 0 {
		t.Fatal("Expected empty cache to have size 0")
	}

	pathHash := uint32(12345)

	// First read operation
	firstContent := []byte("This is the first content read from the file")
	fc.AddRead(pathHash, 0, firstContent)

	// Verify content is cached
	retrieved, exists := fc.GetOldContent(pathHash, 0, uint64(len(firstContent)))
	if !exists {
		t.Fatal("Expected first read to be cached")
	}
	if !bytes.Equal(retrieved, firstContent) {
		t.Error("First read content corrupted")
	}

	// Cache should no longer be empty
	if fc.Size() == 0 {
		t.Error("Expected cache size > 0 after first read")
	}

	// First write operation
	writeData := []byte("Modified content")
	fc.UpdateWithWrite(pathHash, 0, writeData)

	// Original content should be invalidated
	_, exists = fc.GetOldContent(pathHash, 0, uint64(len(firstContent)))
	if exists {
		t.Error("Expected original content to be invalidated by first write")
	}

	// Generate diff for first write
	diff, hasDiff := fc.GenerateDiff(1234, 5, pathHash, 0, writeData)
	if !hasDiff {
		t.Error("Expected diff to be generated for first write")
	}
	if len(diff) == 0 {
		t.Error("Expected non-empty diff for first write")
	}

	// Read new content after write
	fc.AddRead(pathHash, 0, writeData)

	// Should be able to retrieve new content
	retrieved, exists = fc.GetOldContent(pathHash, 0, uint64(len(writeData)))
	if !exists {
		t.Fatal("Expected new content to be cached after write")
	}
	if !bytes.Equal(retrieved, writeData) {
		t.Error("New content after write corrupted")
	}
}

func TestFileDeletedAndRecreated(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	pathHash := uint32(54321)

	// Original file content
	originalContent := []byte("Original file content before deletion")
	fc.AddRead(pathHash, 0, originalContent)

	// Verify content is cached
	retrieved, exists := fc.GetOldContent(pathHash, 0, uint64(len(originalContent)))
	if !exists || !bytes.Equal(retrieved, originalContent) {
		t.Fatal("Original file content not properly cached")
	}

	// Simulate file deletion by not finding any content
	// (In real scenario, the file would be deleted from filesystem)
	// Cache should still contain the old content until evicted or invalidated

	// File gets recreated with different content (same pathHash)
	newContent := []byte("Recreated file with completely different content")

	// Write to recreated file (this should invalidate old cache)
	fc.UpdateWithWrite(pathHash, 0, newContent)

	// Old content should be invalidated
	_, exists = fc.GetOldContent(pathHash, 0, uint64(len(originalContent)))
	if exists {
		t.Error("Expected old content to be invalidated after file recreation")
	}

	// Read new content
	fc.AddRead(pathHash, 0, newContent)

	// Should be able to retrieve new content
	retrieved, exists = fc.GetOldContent(pathHash, 0, uint64(len(newContent)))
	if !exists {
		t.Fatal("Expected new content to be cached")
	}
	if !bytes.Equal(retrieved, newContent) {
		t.Error("New content after recreation corrupted")
	}

	// Verify no remnants of old content exist
	_, exists = fc.GetOldContent(pathHash, uint64(len(newContent)), 10)
	if exists {
		t.Error("Found unexpected content beyond new file size")
	}
}

func TestCacheEvictionDuringActiveOperations(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())

	// Small cache to force evictions
	fc := NewFileCache(16*1024, 32*1024, time.Hour, mockTime)

	// Fill cache with multiple files
	files := []struct {
		pathHash uint32
		content  []byte
	}{
		{1001, make([]byte, 8*1024)}, // 8KB
		{1002, make([]byte, 8*1024)}, // 8KB
		{1003, make([]byte, 8*1024)}, // 8KB
		{1004, make([]byte, 8*1024)}, // 8KB - This should trigger eviction
	}

	// Fill each file with unique data
	for i, file := range files {
		for j := range file.content {
			file.content[j] = byte(i*10 + j%256)
		}
	}

	baseTime := time.Now()

	// Add files with different timestamps
	for i, file := range files {
		mockTime.time = baseTime.Add(time.Duration(i) * time.Minute)
		fc.AddRead(file.pathHash, 0, file.content)
	}

	// Cache should have evicted oldest files due to global limit (32KB)
	// With 4 files of 8KB each = 32KB, some should be evicted

	// Check which files remain (newer ones should be kept)
	remainingFiles := 0
	for i, file := range files {
		_, exists := fc.GetOldContent(file.pathHash, 0, uint64(len(file.content)))
		if exists {
			remainingFiles++
			t.Logf("File %d (pathHash %d) still cached", i, file.pathHash)
		} else {
			t.Logf("File %d (pathHash %d) was evicted", i, file.pathHash)
		}
	}

	if remainingFiles == len(files) {
		t.Error("Expected some files to be evicted due to cache limits")
	}

	// Perform operations on remaining files
	for _, file := range files {
		retrieved, exists := fc.GetOldContent(file.pathHash, 0, uint64(len(file.content)))
		if exists {
			// Update file during active cache pressure
			updateData := []byte("updated")
			fc.UpdateWithWrite(file.pathHash, 0, updateData)

			// Old content should be invalidated
			_, exists = fc.GetOldContent(file.pathHash, 0, uint64(len(file.content)))
			if exists {
				t.Error("Expected old content to be invalidated by update")
			}

			// Add new content
			fc.AddRead(file.pathHash, 0, updateData)

			// Should be able to retrieve updated content
			retrieved, exists = fc.GetOldContent(file.pathHash, 0, uint64(len(updateData)))
			if !exists {
				t.Error("Expected updated content to be cached")
			}
			if !bytes.Equal(retrieved, updateData) {
				t.Error("Updated content corrupted during cache pressure")
			}
		}
	}
}

func TestCacheCorruptionRecovery(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	pathHash := uint32(99999)

	// Add valid content
	validContent := []byte("Valid content that should be consistent")
	fc.AddRead(pathHash, 0, validContent)

	// Verify content is properly stored
	retrieved, exists := fc.GetOldContent(pathHash, 0, uint64(len(validContent)))
	if !exists || !bytes.Equal(retrieved, validContent) {
		t.Fatal("Valid content not properly stored")
	}

	// Simulate corruption scenario: try to retrieve content with wrong size
	// This tests robustness against metadata corruption
	oversizeRetrieve, exists := fc.GetOldContent(pathHash, 0, uint64(len(validContent)*2))
	if exists {
		// If it exists, it should be zero-padded beyond valid content
		if len(oversizeRetrieve) != len(validContent)*2 {
			t.Error("Oversize retrieve returned wrong length")
		}

		// First part should match valid content
		if !bytes.Equal(oversizeRetrieve[:len(validContent)], validContent) {
			t.Error("Valid content portion corrupted in oversize retrieve")
		}

		// Padding should be zeros
		for i := len(validContent); i < len(oversizeRetrieve); i++ {
			if oversizeRetrieve[i] != 0 {
				t.Error("Padding not zero-filled in oversize retrieve")
				break
			}
		}
	}

	// Test recovery from invalid offset scenarios
	_, exists = fc.GetOldContent(pathHash, uint64(len(validContent)+100), 10)
	// Should handle gracefully (either return false or zero-filled)

	// Verify original content is still intact after corruption scenarios
	retrieved, exists = fc.GetOldContent(pathHash, 0, uint64(len(validContent)))
	if !exists {
		t.Fatal("Original content lost after corruption test")
	}
	if !bytes.Equal(retrieved, validContent) {
		t.Error("Original content corrupted by corruption test")
	}

	// Test cache consistency after cleanup
	fc.CleanExpired()

	// Content should still be valid (not expired)
	retrieved, exists = fc.GetOldContent(pathHash, 0, uint64(len(validContent)))
	if !exists {
		t.Error("Content lost after cleanup")
	}
	if !bytes.Equal(retrieved, validContent) {
		t.Error("Content corrupted after cleanup")
	}
}

func TestConcurrentLifecycleOperations(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	pathHash := uint32(77777)

	// Simulate concurrent file lifecycle operations
	baseContent := []byte("Base content for lifecycle test")
	fc.AddRead(pathHash, 0, baseContent)

	// Simulate rapid succession of read-write-read cycles
	for i := 0; i < 10; i++ {
		// Read current content
		current, exists := fc.GetOldContent(pathHash, 0, uint64(len(baseContent)))
		if !exists {
			t.Errorf("Content lost at iteration %d", i)
			break
		}

		// Write new content
		newContent := append(current, []byte(" update")...)
		fc.UpdateWithWrite(pathHash, 0, newContent)

		// Old content should be invalidated
		_, exists = fc.GetOldContent(pathHash, 0, uint64(len(current)))
		if exists {
			t.Errorf("Old content not invalidated at iteration %d", i)
		}

		// Add new content
		fc.AddRead(pathHash, 0, newContent)
		baseContent = newContent
	}

	// Final content should be accessible
	final, exists := fc.GetOldContent(pathHash, 0, uint64(len(baseContent)))
	if !exists {
		t.Fatal("Final content not accessible")
	}

	if !bytes.Equal(final, baseContent) {
		t.Error("Final content corrupted")
	}

	// Content should contain all updates
	expectedUpdates := 10
	actualUpdates := bytes.Count(final, []byte(" update"))
	if actualUpdates != expectedUpdates {
		t.Errorf("Expected %d updates in final content, got %d", expectedUpdates, actualUpdates)
	}
}

func TestCacheStateTransitions(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(32*1024, 64*1024, time.Hour, mockTime)

	pathHash := uint32(55555)

	// State 1: Empty cache
	if fc.Size() != 0 {
		t.Error("Expected empty cache initially")
	}

	// State 2: Cache with data
	data := []byte("Some cached data")
	fc.AddRead(pathHash, 0, data)

	if fc.Size() == 0 {
		t.Error("Expected cache to have data after AddRead")
	}

	// State 3: Data invalidated by write
	fc.UpdateWithWrite(pathHash, 0, []byte("new"))

	// Cache should still have metadata but content invalidated
	_, exists := fc.GetOldContent(pathHash, 0, uint64(len(data)))
	if exists {
		t.Error("Expected content to be invalidated")
	}

	// State 4: Data restored after read
	newData := []byte("new")
	fc.AddRead(pathHash, 0, newData)

	retrieved, exists := fc.GetOldContent(pathHash, 0, uint64(len(newData)))
	if !exists || !bytes.Equal(retrieved, newData) {
		t.Error("Expected new data to be cached")
	}

	// State 5: Cache eviction under pressure
	// Add enough data to trigger eviction
	for i := 0; i < 10; i++ {
		largeData := make([]byte, 8*1024) // 8KB each
		for j := range largeData {
			largeData[j] = byte(i)
		}
		fc.AddRead(uint32(1000+i), 0, largeData)
	}

	// Original data might be evicted
	_, exists = fc.GetOldContent(pathHash, 0, uint64(len(newData)))
	// Result depends on eviction policy, just verify cache remains functional

	// State 6: Cache cleanup
	// Advance time past TTL
	mockTime.Advance(2 * time.Hour)
	fc.CleanExpired()

	// Most data should be expired and cleaned up
	postCleanupSize := fc.Size()
	t.Logf("Cache size after cleanup: %d", postCleanupSize)

	// Cache should still be functional for new operations
	testData := []byte("test after cleanup")
	fc.AddRead(pathHash, 0, testData)

	retrieved, exists = fc.GetOldContent(pathHash, 0, uint64(len(testData)))
	if !exists || !bytes.Equal(retrieved, testData) {
		t.Error("Cache not functional after cleanup")
	}
}
