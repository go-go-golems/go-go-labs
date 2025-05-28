package filecache

import (
	"bytes"
	"testing"
	"time"
)

func TestAppendOnlyWrites(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	pathHash := uint32(12345)

	// Simulate log file append-only pattern
	baseTime := time.Now()
	
	// Initial log content
	log1 := []byte("2025-05-28 10:00:00 INFO Starting application\n")
	fc.AddRead(pathHash, 0, log1)
	
	// Advance time and append more log entries
	mockTime.time = baseTime.Add(1 * time.Minute)
	log2 := []byte("2025-05-28 10:01:00 INFO User login: alice\n")
	fc.AddRead(pathHash, uint64(len(log1)), log2)
	
	mockTime.time = baseTime.Add(2 * time.Minute)
	log3 := []byte("2025-05-28 10:02:00 ERROR Database connection failed\n")
	fc.AddRead(pathHash, uint64(len(log1)+len(log2)), log3)
	
	// Verify we can reconstruct the entire log
	totalLen := uint64(len(log1) + len(log2) + len(log3))
	reconstructed, exists := fc.GetOldContent(pathHash, 0, totalLen)
	
	if !exists {
		t.Fatal("Expected to be able to reconstruct append-only log")
	}
	
	expectedLog := append(append(log1, log2...), log3...)
	if !bytes.Equal(reconstructed, expectedLog) {
		t.Errorf("Log reconstruction failed:\nexpected: %s\ngot: %s", 
			string(expectedLog), string(reconstructed))
	}
	
	// Simulate writing to middle of log (should invalidate subsequent content)
	insertOffset := uint64(len(log1) + 10) // Middle of log2
	insertData := []byte("INSERTED ")
	fc.UpdateWithWrite(pathHash, insertOffset, insertData)
	
	// Content after write might be invalidated (implementation dependent)
	_, exists = fc.GetOldContent(pathHash, insertOffset+uint64(len(insertData)), 10)
	// This is implementation dependent - some implementations might invalidate, others might not
	
	// Content before write should remain
	beforeWrite, exists := fc.GetOldContent(pathHash, 0, insertOffset)
	if !exists {
		t.Error("Expected content before write to remain cached")
	}
	
	expected := append(log1, log2[:10]...)
	if !bytes.Equal(beforeWrite, expected) {
		t.Error("Content before write was corrupted")
	}
}

func TestFileTruncationAndNew(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	pathHash := uint32(54321)

	// Original file content
	originalContent := []byte("This is the original file content that will be truncated")
	fc.AddRead(pathHash, 0, originalContent)
	
	// Verify content is cached
	retrieved, exists := fc.GetOldContent(pathHash, 0, uint64(len(originalContent)))
	if !exists || !bytes.Equal(retrieved, originalContent) {
		t.Fatal("Original content not properly cached")
	}
	
	// Simulate file truncation (write at offset 0 with smaller content)
	newContent := []byte("Truncated file")
	fc.UpdateWithWrite(pathHash, 0, newContent)
	
	// Old content should be invalidated
	_, exists = fc.GetOldContent(pathHash, 0, uint64(len(originalContent)))
	if exists {
		t.Error("Expected original content to be invalidated after truncation")
	}
	
	// Read new content after truncation
	fc.AddRead(pathHash, 0, newContent)
	
	// Should be able to retrieve new truncated content
	retrieved, exists = fc.GetOldContent(pathHash, 0, uint64(len(newContent)))
	if !exists {
		t.Fatal("Expected new content to be cached after truncation")
	}
	
	if !bytes.Equal(retrieved, newContent) {
		t.Errorf("New content corrupted: expected %s, got %s", 
			string(newContent), string(retrieved))
	}
	
	// Content beyond new size should not exist
	_, exists = fc.GetOldContent(pathHash, uint64(len(newContent)), 10)
	if exists {
		t.Error("Expected no content beyond truncated size")
	}
}

func TestSparseFileOperations(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(128*1024, 512*1024, time.Hour, mockTime)

	pathHash := uint32(99999)

	// Create sparse file pattern with large gaps
	segments := []struct {
		offset uint64
		data   []byte
	}{
		{0, []byte("File header")},
		{1024, []byte("Data block 1")},          // 1KB gap
		{1024*10, []byte("Data block 2")},       // 9KB gap
		{1024*100, []byte("Data block 3")},      // 90KB gap
		{1024*1000, []byte("Data block 4")},     // 900KB gap
	}

	// Add all segments
	for _, seg := range segments {
		fc.AddRead(pathHash, seg.offset, seg.data)
	}

	// Verify each segment can be retrieved individually
	for i, seg := range segments {
		retrieved, exists := fc.GetOldContent(pathHash, seg.offset, uint64(len(seg.data)))
		if !exists {
			t.Errorf("Segment %d not found at offset %d", i, seg.offset)
			continue
		}
		if !bytes.Equal(retrieved, seg.data) {
			t.Errorf("Segment %d corrupted at offset %d", i, seg.offset)
		}
	}

	// Test reconstruction spanning multiple segments with gaps
	// Request range that includes first two segments plus gaps
	spanStart := uint64(0)
	spanEnd := segments[2].offset + uint64(len(segments[2].data))
	spanLength := spanEnd - spanStart

	reconstructed, exists := fc.GetOldContent(pathHash, spanStart, spanLength)
	if !exists {
		t.Fatal("Expected to reconstruct sparse range")
	}

	if uint64(len(reconstructed)) != spanLength {
		t.Fatalf("Reconstructed length mismatch: expected %d, got %d", 
			spanLength, len(reconstructed))
	}

	// Verify data segments are correct and gaps are zero-filled
	// Check first segment
	if !bytes.Equal(reconstructed[0:len(segments[0].data)], segments[0].data) {
		t.Error("First segment corrupted in reconstruction")
	}

	// Check gap between first and second segment (should be zeros)
	gapStart := len(segments[0].data)
	gapEnd := int(segments[1].offset)
	for i := gapStart; i < gapEnd; i++ {
		if reconstructed[i] != 0 {
			t.Errorf("Gap not zero-filled at position %d", i)
			break
		}
	}

	// Write to a gap and verify it invalidates nothing
	gapWriteOffset := uint64(500) // In gap between header and first data block
	gapWriteData := []byte("gap data")
	fc.UpdateWithWrite(pathHash, gapWriteOffset, gapWriteData)

	// All original segments should still exist
	for i, seg := range segments[:3] { // Check first 3 segments
		retrieved, exists := fc.GetOldContent(pathHash, seg.offset, uint64(len(seg.data)))
		if !exists {
			t.Errorf("Segment %d was invalidated by gap write", i)
		}
		if !bytes.Equal(retrieved, seg.data) {
			t.Errorf("Segment %d corrupted by gap write", i)
		}
	}
}

func TestFileGrowthAndShrinkage(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	pathHash := uint32(77777)

	// Start with small file
	phase1 := []byte("Small initial content")
	fc.AddRead(pathHash, 0, phase1)

	// Grow file by appending
	phase2 := []byte(" - appended content that makes the file longer")
	fc.AddRead(pathHash, uint64(len(phase1)), phase2)

	// Further growth
	phase3 := []byte(" - even more content added to the growing file")
	fc.AddRead(pathHash, uint64(len(phase1)+len(phase2)), phase3)

	// Verify entire grown file can be reconstructed
	totalLength := uint64(len(phase1) + len(phase2) + len(phase3))
	grown, exists := fc.GetOldContent(pathHash, 0, totalLength)
	if !exists {
		t.Fatal("Expected to reconstruct grown file")
	}

	expected := append(append(phase1, phase2...), phase3...)
	if !bytes.Equal(grown, expected) {
		t.Error("File growth reconstruction failed")
	}

	// Simulate file shrinkage by overwriting with smaller content
	shrunkContent := []byte("Shrunk")
	fc.UpdateWithWrite(pathHash, 0, shrunkContent)

	// Content beyond shrunk size should be invalidated
	_, exists = fc.GetOldContent(pathHash, uint64(len(shrunkContent)), 10)
	if exists {
		t.Error("Expected content beyond shrunk size to be invalidated")
	}

	// Add new content for shrunk file
	fc.AddRead(pathHash, 0, shrunkContent)

	// Should be able to retrieve shrunk content
	retrieved, exists := fc.GetOldContent(pathHash, 0, uint64(len(shrunkContent)))
	if !exists {
		t.Fatal("Expected shrunk content to be cached")
	}

	if !bytes.Equal(retrieved, shrunkContent) {
		t.Error("Shrunk content corrupted")
	}

	// Test pattern of grow-shrink-grow
	newGrowth := []byte(" - growing again after shrinkage")
	fc.AddRead(pathHash, uint64(len(shrunkContent)), newGrowth)

	finalLength := uint64(len(shrunkContent) + len(newGrowth))
	final, exists := fc.GetOldContent(pathHash, 0, finalLength)
	if !exists {
		t.Fatal("Expected to reconstruct re-grown file")
	}

	expectedFinal := append(shrunkContent, newGrowth...)
	if !bytes.Equal(final, expectedFinal) {
		t.Error("Re-growth pattern failed")
	}
}

func TestRandomAccessPatterns(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(128*1024, 512*1024, time.Hour, mockTime)

	pathHash := uint32(88888)

	// Simulate random access pattern (database file, etc.)
	accessPattern := []struct {
		offset uint64
		data   []byte
		desc   string
	}{
		{4096, []byte("Page 1 content"), "page 1"},
		{8192, []byte("Page 2 content"), "page 2"},
		{1024, []byte("Metadata block"), "metadata"},
		{16384, []byte("Page 3 content"), "page 3"},
		{0, []byte("File header"), "header"},
		{12288, []byte("Index block"), "index"},
	}

	// Add all blocks in random order
	for _, access := range accessPattern {
		fc.AddRead(pathHash, access.offset, access.data)
	}

	// Verify each block independently
	for _, access := range accessPattern {
		retrieved, exists := fc.GetOldContent(pathHash, access.offset, uint64(len(access.data)))
		if !exists {
			t.Errorf("Random access block %s not found", access.desc)
			continue
		}
		if !bytes.Equal(retrieved, access.data) {
			t.Errorf("Random access block %s corrupted", access.desc)
		}
	}

	// Simulate overwrite of one block
	newPage2 := []byte("Updated page 2")
	fc.UpdateWithWrite(pathHash, 8192, newPage2)

	// Other blocks should remain unchanged
	unchanged := []int{0, 2, 3, 4, 5} // All except page 2 (index 1)
	for _, i := range unchanged {
		access := accessPattern[i]
		retrieved, exists := fc.GetOldContent(pathHash, access.offset, uint64(len(access.data)))
		if !exists {
			t.Errorf("Block %s was invalidated by unrelated write", access.desc)
		}
		if !bytes.Equal(retrieved, access.data) {
			t.Errorf("Block %s corrupted by unrelated write", access.desc)
		}
	}

	// Updated block should be invalidated
	_, exists := fc.GetOldContent(pathHash, 8192, uint64(len(accessPattern[1].data)))
	if exists {
		t.Error("Updated block should be invalidated")
	}

	// Add updated content
	fc.AddRead(pathHash, 8192, newPage2)

	// Should be able to retrieve updated content
	retrieved, exists := fc.GetOldContent(pathHash, 8192, uint64(len(newPage2)))
	if !exists {
		t.Fatal("Expected updated block to be cached")
	}

	if !bytes.Equal(retrieved, newPage2) {
		t.Error("Updated block content corrupted")
	}
}
