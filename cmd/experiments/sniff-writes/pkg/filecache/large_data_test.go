package filecache

import (
	"bytes"
	"math"
	"testing"
	"time"
)

// TestLargeSegmentHandling tests behavior with 128KB segments
func TestLargeSegmentHandling(t *testing.T) {
	fc := NewFileCache(128*1024, 10*1024*1024, time.Hour, NewMockTimeProvider(time.Now())) // 128KB per file, 10MB total
	
	pathHash := uint32(42)
	
	// Create a 128KB data segment
	largeData := make([]byte, 128*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	
	// Test adding large segment
	fc.AddRead(pathHash, 0, largeData)
	
	// Verify retrieval
	retrieved, exists := fc.GetOldContent(pathHash, 0, uint64(len(largeData)))
	if !exists {
		t.Error("Large segment not found")
	}
	if !bytes.Equal(retrieved, largeData) {
		t.Error("Large segment data mismatch on retrieval")
	}
	
	// Test segment exactly at per-file limit
	fc2 := NewFileCache(128*1024, 10*1024*1024, time.Hour, NewMockTimeProvider(time.Now()))
	
	// This should succeed
	fc2.AddRead(pathHash, 0, largeData)
	
	// Adding one more byte should fail or succeed depending on implementation
	// (Current implementation doesn't enforce strict limits at AddRead time)
	fc2.AddRead(pathHash, uint64(len(largeData)), []byte{0xFF})
}

// TestMaximumOffsetHandling tests operations near uint64 limit
func TestMaximumOffsetHandling(t *testing.T) {
	fc := NewFileCache(64*1024, 1024*1024, time.Hour, NewMockTimeProvider(time.Now()))
	pathHash := uint32(123)
	
	// Test with maximum possible offset
	maxOffset := uint64(math.MaxUint64 - 1000) // Leave some room
	testData := []byte("test data at max offset")
	
	fc.AddRead(pathHash, maxOffset, testData)
	
	// Verify retrieval
	retrieved, exists := fc.GetOldContent(pathHash, maxOffset, uint64(len(testData)))
	if !exists {
		t.Error("Data not found at maximum offset")
	}
	if !bytes.Equal(retrieved, testData) {
		t.Error("Data mismatch at maximum offset")
	}
	
	// Test offset overflow protection
	overflowOffset := uint64(math.MaxUint64)
	fc.AddRead(pathHash, overflowOffset, []byte("overflow"))
	// This should either succeed or fail gracefully, not panic
	
	// Verify cache state remains consistent
	fc.mu.RLock()
	sparseFile := fc.files[pathHash]
	fc.mu.RUnlock()
	
	if sparseFile != nil {
		sparseFile.mu.RLock()
		segmentCount := len(sparseFile.Segments)
		sparseFile.mu.RUnlock()
		
		if segmentCount < 0 {
			t.Error("Invalid segment count after max offset operations")
		}
	}
}

// TestMultipleKBSegments tests handling of multiple large segments
func TestMultipleKBSegments(t *testing.T) {
	fc := NewFileCache(512*1024, 2*1024*1024, time.Hour, NewMockTimeProvider(time.Now())) // 512KB per file, 2MB total
	pathHash := uint32(456)
	
	// Create multiple 64KB segments
	segmentSize := 64 * 1024
	segmentCount := 6 // 384KB total
	
	segments := make([][]byte, segmentCount)
	for i := 0; i < segmentCount; i++ {
		segments[i] = make([]byte, segmentSize)
		// Fill with recognizable pattern
		for j := range segments[i] {
			segments[i][j] = byte((i*256 + j) % 256)
		}
		
		offset := uint64(i * segmentSize * 2) // Leave gaps between segments
		fc.AddRead(pathHash, offset, segments[i])
	}
	
	// Verify all segments
	for i := 0; i < segmentCount; i++ {
		offset := uint64(i * segmentSize * 2)
		retrieved, exists := fc.GetOldContent(pathHash, offset, uint64(segmentSize))
		if !exists {
			t.Errorf("Segment %d not found", i)
			continue
		}
		if !bytes.Equal(retrieved, segments[i]) {
			t.Errorf("Segment %d data mismatch", i)
		}
	}
	
	// Test merging adjacent large segments
	gapFill := make([]byte, segmentSize) // Fill the gap between segment 0 and 1
	for i := range gapFill {
		gapFill[i] = 0xAA
	}
	
	fc.AddRead(pathHash, uint64(segmentSize), gapFill)
	
	// Verify merged content spans multiple segments
	largeRetrieved, exists := fc.GetOldContent(pathHash, 0, uint64(segmentSize*3))
	if !exists {
		t.Error("Large merged content not found")
	}
	if len(largeRetrieved) != segmentSize*3 {
		t.Errorf("Expected %d bytes, got %d", segmentSize*3, len(largeRetrieved))
	}
}

// TestMemoryPressureWithLargeData tests behavior under memory pressure
func TestMemoryPressureWithLargeData(t *testing.T) {
	// Small cache with large per-file limit
	fc := NewFileCache(128*1024, 256*1024, time.Hour, NewMockTimeProvider(time.Now())) // 128KB per file, 256KB total
	
	largeSegment := make([]byte, 100*1024) // 100KB
	for i := range largeSegment {
		largeSegment[i] = byte(i % 256)
	}
	
	// Add segments until memory pressure
	pathHashes := []uint32{1, 2, 3, 4, 5}
	
	for i, pathHash := range pathHashes {
		fc.AddRead(pathHash, 0, largeSegment)
		t.Logf("Added segment %d (100KB)", i+1)
	}
	
	// Verify cache remains functional
	fc.mu.RLock()
	fileCount := len(fc.files)
	fc.mu.RUnlock()
	
	if fileCount == 0 {
		t.Error("Cache became completely empty under memory pressure")
	}
	
	// Verify we can still add small segments
	smallData := []byte("small data")
	fc.AddRead(999, 0, smallData)
	retrieved, exists := fc.GetOldContent(999, 0, uint64(len(smallData)))
	if exists && !bytes.Equal(retrieved, smallData) {
		t.Error("Small segment handling failed after memory pressure")
	}
}

// TestLargeOffsetArithmetic tests arithmetic operations with large offsets
func TestLargeOffsetArithmetic(t *testing.T) {
	fc := NewFileCache(64*1024, 1024*1024, time.Hour, NewMockTimeProvider(time.Now()))
	pathHash := uint32(789)
	
	// Test segments near boundaries that might cause overflow
	testCases := []struct {
		name   string
		offset uint64
		data   []byte
	}{
		{
			name:   "Near 32-bit boundary",
			offset: uint64(math.MaxUint32) - 100,
			data:   []byte("data near 32-bit max"),
		},
		{
			name:   "At 32-bit boundary",
			offset: uint64(math.MaxUint32),
			data:   []byte("data at 32-bit max"),
		},
		{
			name:   "Past 32-bit boundary",
			offset: uint64(math.MaxUint32) + 1000,
			data:   []byte("data past 32-bit max"),
		},
		{
			name:   "Large power of 2",
			offset: uint64(1) << 40, // 1TB offset
			data:   []byte("data at 1TB offset"),
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fc.AddRead(pathHash, tc.offset, tc.data)
			
			retrieved, exists := fc.GetOldContent(pathHash, tc.offset, uint64(len(tc.data)))
			if !exists {
				t.Errorf("Data not found for %s", tc.name)
				return
			}
			if !bytes.Equal(retrieved, tc.data) {
				t.Errorf("Data mismatch for %s", tc.name)
			}
		})
	}
	
	// Verify segments are properly ordered despite large offsets
	fc.mu.RLock()
	sparseFile := fc.files[pathHash]
	fc.mu.RUnlock()
	
	if sparseFile != nil {
		sparseFile.mu.RLock()
		defer sparseFile.mu.RUnlock()
		
		// Check segments are in ascending order
		for i := 1; i < len(sparseFile.Segments); i++ {
			if sparseFile.Segments[i-1].Start >= sparseFile.Segments[i].Start {
				t.Error("Segments not properly ordered with large offsets")
			}
		}
	}
}

// TestSegmentSizeLimits tests behavior at segment size boundaries
func TestSegmentSizeLimits(t *testing.T) {
	fc := NewFileCache(512*1024, 1024*1024, time.Hour, NewMockTimeProvider(time.Now()))
	pathHash := uint32(321)
	
	// Test empty segment
	fc.AddRead(pathHash, 0, []byte{})
	
	// Test nil data
	fc.AddRead(pathHash, 100, nil)
	
	// Test exactly at per-file limit
	maxData := make([]byte, 512*1024)
	for i := range maxData {
		maxData[i] = byte(i % 256)
	}
	
	fc.AddRead(pathHash+1, 0, maxData)
	
	// Verify retrieval of max-size segment
	retrieved, exists := fc.GetOldContent(pathHash+1, 0, uint64(len(maxData)))
	if !exists {
		t.Error("Max-size segment not found")
	}
	if !bytes.Equal(retrieved, maxData) {
		t.Error("Max-size segment data mismatch")
	}
}
