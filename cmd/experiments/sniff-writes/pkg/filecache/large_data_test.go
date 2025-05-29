package filecache

import (
	"bytes"
	"math"
	"testing"
	"time"
)

func TestLargeDataHandling(t *testing.T) {
	// Test with 128KB segments - maximum expected size
	const maxSegmentSize = 128 * 1024

	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(
		200*1024,     // 200KB per file limit
		1024*1024,    // 1MB global limit
		time.Hour,
		mockTime,
	)

	pathHash := uint32(12345)

	// Test 1: Store and retrieve 128KB segment
	largeData := make([]byte, maxSegmentSize)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	fc.AddRead(pathHash, 0, largeData)

	retrieved, exists := fc.GetOldContent(pathHash, 0, maxSegmentSize)
	if !exists {
		t.Fatal("Expected large data to be stored")
	}

	if !bytes.Equal(largeData, retrieved) {
		t.Fatal("Large data corruption detected")
	}

	// Test 2: Multiple large segments that exceed per-file limit
	// This should trigger eviction
	secondLargeData := make([]byte, maxSegmentSize)
	for i := range secondLargeData {
		secondLargeData[i] = byte((i + 128) % 256)
	}

	fc.AddRead(pathHash, maxSegmentSize, secondLargeData)

	// Should have evicted first segment due to per-file limit (200KB < 256KB)
	_, exists1 := fc.GetOldContent(pathHash, 0, maxSegmentSize)
	retrieved2, exists2 := fc.GetOldContent(pathHash, maxSegmentSize, maxSegmentSize)

	if exists1 {
		t.Error("Expected first large segment to be evicted due to per-file limit")
	}
	if !exists2 {
		t.Fatal("Expected second large segment to be retained")
	}
	if !bytes.Equal(secondLargeData, retrieved2) {
		t.Fatal("Second large data corrupted")
	}
}

func TestMaxOffsetHandling(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(1024*1024, 10*1024*1024, time.Hour, mockTime)

	pathHash := uint32(67890)

	// Test with maximum possible offset (near uint64 limit)
	maxOffset := uint64(math.MaxUint64 - 1000) // Leave some room to avoid overflow
	testData := []byte("data at max offset")

	fc.AddRead(pathHash, maxOffset, testData)

	retrieved, exists := fc.GetOldContent(pathHash, maxOffset, uint64(len(testData)))
	if !exists {
		t.Fatal("Expected data at max offset to be stored")
	}

	if !bytes.Equal(testData, retrieved) {
		t.Fatal("Data at max offset corrupted")
	}

	// Test write invalidation at max offset
	writeData := []byte("overwrite at max")
	fc.UpdateWithWrite(pathHash, maxOffset, writeData)

	_, exists = fc.GetOldContent(pathHash, maxOffset, uint64(len(testData)))
	if exists {
		t.Error("Expected data at max offset to be invalidated by write")
	}
}

func TestLargeOffsetRanges(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(512*1024, 2*1024*1024, time.Hour, mockTime)

	pathHash := uint32(11111)

	// Test segments at various large offsets
	offsets := []uint64{
		1 << 20,                              // 1MB
		1 << 30,                              // 1GB
		1 << 40,                              // 1TB
		uint64(math.MaxUint32),               // 4GB boundary
		uint64(math.MaxUint32) + 1000,        // Just over 4GB
		uint64(math.MaxUint64/2),             // Half of max
		uint64(math.MaxUint64 - 10000),       // Near max
	}

	dataMap := make(map[uint64][]byte)

	// Store data at each offset
	for _, offset := range offsets {
		data := make([]byte, 100)
		for i := range data {
			data[i] = byte(offset % 256) // Unique pattern based on offset
		}
		dataMap[offset] = data
		fc.AddRead(pathHash, offset, data)
	}

	// Verify all data can be retrieved correctly
	for _, offset := range offsets {
		expected := dataMap[offset]
		retrieved, exists := fc.GetOldContent(pathHash, offset, uint64(len(expected)))
		
		if !exists {
			t.Errorf("Expected data at offset %d to exist", offset)
			continue
		}

		if !bytes.Equal(expected, retrieved) {
			t.Errorf("Data corruption at offset %d", offset)
		}
	}
}

func TestLargeContentReconstruction(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(1024*1024, 10*1024*1024, time.Hour, mockTime)

	pathHash := uint32(22222)

	// Create large sparse file with gaps
	segmentSize := 32 * 1024 // 32KB segments
	gapSize := 16 * 1024     // 16KB gaps

	segments := []struct {
		offset uint64
		data   []byte
	}{
		{0, make([]byte, segmentSize)},
		{uint64(segmentSize + gapSize), make([]byte, segmentSize)},
		{uint64(2*(segmentSize+gapSize)), make([]byte, segmentSize)},
		{uint64(3*(segmentSize+gapSize)), make([]byte, segmentSize)},
	}

	// Fill each segment with unique data
	for i, seg := range segments {
		for j := range seg.data {
			seg.data[j] = byte(i*100 + j%256)
		}
		fc.AddRead(pathHash, seg.offset, seg.data)
	}

	// Reconstruct large range spanning all segments and gaps
	totalRange := uint64(4 * (segmentSize + gapSize))
	reconstructed, exists := fc.GetOldContent(pathHash, 0, totalRange)
	
	if !exists {
		t.Fatal("Expected reconstruction to succeed")
	}

	if len(reconstructed) != int(totalRange) {
		t.Fatalf("Expected reconstructed length %d, got %d", totalRange, len(reconstructed))
	}

	// Verify segment data and gaps
	for i, seg := range segments {
		segmentStart := int(seg.offset)
		segmentEnd := segmentStart + len(seg.data)
		
		if !bytes.Equal(reconstructed[segmentStart:segmentEnd], seg.data) {
			t.Errorf("Segment %d data corrupted during reconstruction", i)
		}

		// Check gap after segment (except last)
		if i < len(segments)-1 {
			gapStart := segmentEnd
			nextSegmentStart := int(segments[i+1].offset)
			gap := reconstructed[gapStart:nextSegmentStart]
			
			// All gap bytes should be zero
			for j, b := range gap {
				if b != 0 {
					t.Errorf("Gap byte at position %d should be 0, got %d", gapStart+j, b)
					break
				}
			}
		}
	}
}

func TestSegmentSizeLimits(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime) // Small limits

	pathHash := uint32(33333)

	// Test segment exactly at per-file limit
	limitData := make([]byte, 64*1024)
	for i := range limitData {
		limitData[i] = byte(i % 256)
	}

	fc.AddRead(pathHash, 0, limitData)

	retrieved, exists := fc.GetOldContent(pathHash, 0, uint64(len(limitData)))
	if !exists {
		t.Fatal("Expected data at limit size to be stored")
	}

	if !bytes.Equal(limitData, retrieved) {
		t.Fatal("Data at limit size corrupted")
	}

	// Test segment slightly over per-file limit should trigger eviction
	overLimitData := make([]byte, 65*1024)
	for i := range overLimitData {
		overLimitData[i] = byte((i + 50) % 256)
	}

	fc.AddRead(pathHash, 64*1024, overLimitData)

	// First segment should be evicted
	_, exists1 := fc.GetOldContent(pathHash, 0, uint64(len(limitData)))
	_, exists2 := fc.GetOldContent(pathHash, 64*1024, uint64(len(overLimitData)))

	if exists1 {
		t.Error("Expected first segment to be evicted when adding over-limit segment")
	}
	if !exists2 {
		t.Error("Expected over-limit segment to be stored (after eviction)")
	}
}
