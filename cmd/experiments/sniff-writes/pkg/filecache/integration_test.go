package filecache

import (
	"testing"
	"time"
)

// TestFullWorkflow tests the complete read->write->diff workflow
func TestFullWorkflow(t *testing.T) {
	cache := NewFileCache(1024, 4096, time.Hour, RealTimeProvider{})

	pathHash := uint32(1)

	// Step 1: Read some initial content
	initialContent := []byte("Hello, this is the initial file content that we will modify.")
	cache.AddRead(pathHash, 0, initialContent)

	// Step 2: Simulate a write that overlaps part of the read content
	writeOffset := uint64(7)
	writeData := []byte("world! This")
	writeEnd := writeOffset + uint64(len(writeData))

	// Get old content for diffing
	oldContent, exists := cache.GetOldContent(pathHash, writeOffset, uint64(len(writeData)))
	if !exists {
		t.Fatal("expected to find old content for diff")
	}

	expectedOldContent := "this is"
	if string(oldContent) != expectedOldContent {
		t.Errorf("expected old content %q, got %q", expectedOldContent, string(oldContent))
	}

	// Step 3: Update cache with the write
	cache.UpdateWithWrite(pathHash, writeOffset, writeData)

	// Step 4: Verify the cache now contains the updated content
	newContent, exists := cache.GetOldContent(pathHash, writeOffset, uint64(len(writeData)))
	if !exists {
		t.Fatal("expected to find new content after write")
	}

	if string(newContent) != string(writeData) {
		t.Errorf("expected new content %q, got %q", writeData, string(newContent))
	}

	// Step 5: Verify content before and after the write is preserved
	beforeContent, exists := cache.GetOldContent(pathHash, 0, writeOffset)
	if !exists {
		t.Fatal("expected content before write to be preserved")
	}

	expectedBefore := "Hello, "
	if string(beforeContent) != expectedBefore {
		t.Errorf("expected before content %q, got %q", expectedBefore, string(beforeContent))
	}

	afterOffset := writeEnd
	afterLength := uint64(len(initialContent)) - afterOffset
	afterContent, exists := cache.GetOldContent(pathHash, afterOffset, afterLength)
	if !exists {
		t.Fatal("expected content after write to be preserved")
	}

	expectedAfter := " the initial file content that we will modify."
	if string(afterContent) != expectedAfter {
		t.Errorf("expected after content %q, got %q", expectedAfter, string(afterContent))
	}
}

// TestMultipleFileWorkflow tests operations across multiple files
func TestMultipleFileWorkflow(t *testing.T) {
	cache := NewFileCache(1024, 4096, time.Hour, RealTimeProvider{})

	// Work with multiple files simultaneously
	files := map[uint32]string{
		1: "Content for file one with some text",
		2: "Content for file two with different text",
		3: "Content for file three with more text",
	}

	// Add initial content for all files
	for pathHash, content := range files {
		cache.AddRead(pathHash, 0, []byte(content))
	}

	// Modify each file
	for pathHash := range files {
		writeData := []byte("MODIFIED")
		cache.UpdateWithWrite(pathHash, 12, writeData) // Overwrite "file X with"

		// Verify modification
		result, exists := cache.GetOldContent(pathHash, 12, uint64(len(writeData)))
		if !exists {
			t.Errorf("file %d: expected modified content to exist", pathHash)
			continue
		}

		if string(result) != string(writeData) {
			t.Errorf("file %d: expected %q, got %q", pathHash, writeData, string(result))
		}
	}

	// Verify all files still exist and have correct content
	for pathHash := range files {
		prefix, exists := cache.GetOldContent(pathHash, 0, 12)
		if !exists {
			t.Errorf("file %d: expected prefix to exist", pathHash)
			continue
		}

		expectedPrefix := "Content for "
		if string(prefix) != expectedPrefix {
			t.Errorf("file %d: expected prefix %q, got %q", pathHash, expectedPrefix, string(prefix))
		}
	}
}

// TestReadModifyWritePattern tests common read-modify-write patterns
func TestReadModifyWritePattern(t *testing.T) {
	cache := NewFileCache(1024, 4096, time.Hour, RealTimeProvider{})

	pathHash := uint32(1)
	
	// Initial state: empty file
	// Read 1: Add content at offset 0
	cache.AddRead(pathHash, 0, []byte("Line 1\n"))
	
	// Read 2: Add content at offset 7
	cache.AddRead(pathHash, 7, []byte("Line 2\n"))
	
	// Read 3: Add content at offset 14
	cache.AddRead(pathHash, 14, []byte("Line 3\n"))

	// Now we have a file with three lines: "Line 1\nLine 2\nLine 3\n"
	
	// Simulate editing line 2 (offset 7-13)
	cache.UpdateWithWrite(pathHash, 7, []byte("EDITED\n"))

	// Verify the result
	fullContent, exists := cache.GetOldContent(pathHash, 0, 21) // Full content length
	if !exists {
		t.Fatal("expected full content to exist")
	}

	expected := "Line 1\nEDITED\nLine 3\n"
	if string(fullContent) != expected {
		t.Errorf("expected %q, got %q", expected, string(fullContent))
	}
}

// TestGapFilling tests content reconstruction with gaps
func TestGapFilling(t *testing.T) {
	cache := NewFileCache(1024, 4096, time.Hour, RealTimeProvider{})

	pathHash := uint32(1)

	// Create sparse content with gaps
	cache.AddRead(pathHash, 0, []byte("START"))     // [0-5)
	cache.AddRead(pathHash, 20, []byte("MIDDLE"))   // [20-26)
	cache.AddRead(pathHash, 50, []byte("END"))      // [50-53)

	// Request content that spans gaps
	content, exists := cache.GetOldContent(pathHash, 0, 53)
	if !exists {
		t.Fatal("expected sparse content to exist")
	}

	// Should fill gaps with 0x00
	expected := make([]byte, 53)
	copy(expected[0:5], "START")
	copy(expected[20:26], "MIDDLE")
	copy(expected[50:53], "END")
	// Gaps at [5-20) and [26-50) should be zeros

	if len(content) != len(expected) {
		t.Fatalf("expected length %d, got %d", len(expected), len(content))
	}

	for i, expectedByte := range expected {
		if content[i] != expectedByte {
			t.Errorf("byte %d: expected 0x%02X, got 0x%02X", i, expectedByte, content[i])
		}
	}
}

// TestCacheStateConsistency tests that cache maintains consistent state
func TestCacheStateConsistency(t *testing.T) {
	cache := NewFileCache(1024, 4096, time.Hour, RealTimeProvider{})

	pathHash := uint32(1)

	// Perform a series of operations and verify consistency
	operations := []struct {
		op     string
		offset uint64
		data   []byte
	}{
		{"read", 0, []byte("abcdefghij")},
		{"read", 10, []byte("klmnopqrst")},
		{"write", 5, []byte("12345")},
		{"read", 25, []byte("uvwxyz")},
		{"write", 15, []byte("67890")},
		{"write", 8, []byte("XYZ")},
	}

	for i, op := range operations {
		switch op.op {
		case "read":
			cache.AddRead(pathHash, op.offset, op.data)
		case "write":
			cache.UpdateWithWrite(pathHash, op.offset, op.data)
		}

		// After each operation, verify cache state consistency
		sf, exists := cache.files[pathHash]
		if !exists {
			t.Fatalf("operation %d: file should exist", i)
		}

		// Verify segments are sorted
		for j := 1; j < len(sf.Segments); j++ {
			if sf.Segments[j-1].Start >= sf.Segments[j].Start {
				t.Errorf("operation %d: segments not sorted at index %d", i, j)
			}
		}

		// Verify no overlapping segments
		for j := 1; j < len(sf.Segments); j++ {
			if sf.Segments[j-1].End > sf.Segments[j].Start {
				t.Errorf("operation %d: overlapping segments at index %d", i, j)
			}
		}

		// Verify segment data length matches range
		for j, seg := range sf.Segments {
			expectedLen := seg.End - seg.Start
			if uint64(len(seg.Data)) != expectedLen {
				t.Errorf("operation %d, segment %d: data length %d doesn't match range length %d", 
					i, j, len(seg.Data), expectedLen)
			}
		}
	}
}
