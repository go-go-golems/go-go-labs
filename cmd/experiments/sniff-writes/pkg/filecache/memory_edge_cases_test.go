package filecache

import (
	"testing"
	"time"
)

func TestSingleSegmentExceedsPerFileLimit(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	
	// Set per-file limit smaller than the segment we'll add
	perFileLimit := uint64(32 * 1024)  // 32KB
	globalLimit := uint64(256 * 1024)  // 256KB
	
	fc := NewFileCache(perFileLimit, globalLimit, time.Hour, mockTime)
	
	pathHash := uint32(12345)
	
	// Create segment larger than per-file limit
	largeSegment := make([]byte, 64*1024) // 64KB > 32KB limit
	for i := range largeSegment {
		largeSegment[i] = byte(i % 256)
	}
	
	// This should still be stored (individual segments can exceed per-file limit)
	fc.AddRead(pathHash, 0, largeSegment)
	
	retrieved, exists := fc.GetOldContent(pathHash, 0, uint64(len(largeSegment)))
	if !exists {
		t.Fatal("Expected large segment to be stored even when exceeding per-file limit")
	}
	
	if len(retrieved) != len(largeSegment) {
		t.Fatalf("Expected retrieved length %d, got %d", len(largeSegment), len(retrieved))
	}
	
	// Try to add another segment - this should trigger eviction of the first
	secondSegment := make([]byte, 16*1024)
	for i := range secondSegment {
		secondSegment[i] = byte((i + 100) % 256)
	}
	
	fc.AddRead(pathHash, uint64(len(largeSegment)), secondSegment)
	
	// First segment should be evicted due to total size exceeding per-file limit
	_, exists1 := fc.GetOldContent(pathHash, 0, uint64(len(largeSegment)))
	retrieved2, exists2 := fc.GetOldContent(pathHash, uint64(len(largeSegment)), uint64(len(secondSegment)))
	
	if exists1 {
		t.Error("Expected first large segment to be evicted when adding second segment")
	}
	
	if !exists2 {
		t.Fatal("Expected second segment to be retained")
	}
	
	if len(retrieved2) != len(secondSegment) {
		t.Errorf("Second segment length mismatch: expected %d, got %d", len(secondSegment), len(retrieved2))
	}
}

func TestGlobalLimitWithMultipleFiles(t *testing.T) {
	t.Log("Starting TestGlobalLimitWithMultipleFiles")
	mockTime := NewMockTimeProvider(time.Now())
	
	perFileLimit := uint64(64 * 1024)  // 64KB per file
	globalLimit := uint64(128 * 1024)  // 128KB total (only 2 files max)
	
	fc := NewFileCache(perFileLimit, globalLimit, time.Hour, mockTime)
	t.Log("Created cache with limits: per-file=64KB, global=128KB")
	
	// Fill first file to per-file limit
	pathHash1 := uint32(1111)
	data1 := make([]byte, 60*1024) // 60KB
	for i := range data1 {
		data1[i] = byte(1)
	}
	t.Log("Adding first file (60KB)")
	fc.AddRead(pathHash1, 0, data1)
	t.Log("First file added successfully")
	
	// Fill second file to per-file limit
	pathHash2 := uint32(2222)
	data2 := make([]byte, 60*1024) // 60KB
	for i := range data2 {
		data2[i] = byte(2)
	}
	t.Log("Adding second file (60KB)")
	fc.AddRead(pathHash2, 0, data2)
	t.Log("Second file added successfully")
	
	// Both should fit within global limit (120KB < 128KB)
	t.Log("Checking both files exist")
	_, exists1 := fc.GetOldContent(pathHash1, 0, uint64(len(data1)))
	_, exists2 := fc.GetOldContent(pathHash2, 0, uint64(len(data2)))
	t.Logf("File 1 exists: %v, File 2 exists: %v", exists1, exists2)
	
	if !exists1 || !exists2 {
		t.Fatal("Both files should fit within global limit")
	}
	
	// Add third file - should trigger global eviction
	pathHash3 := uint32(3333)
	data3 := make([]byte, 20*1024) // 20KB
	for i := range data3 {
		data3[i] = byte(3)
	}
	t.Log("Adding third file (20KB) - should trigger eviction")
	fc.AddRead(pathHash3, 0, data3)
	t.Log("Third file added successfully")
	
	// Should evict oldest file (file 1) to make room
	t.Log("Checking final state after eviction")
	_, exists1After := fc.GetOldContent(pathHash1, 0, uint64(len(data1)))
	_, exists2After := fc.GetOldContent(pathHash2, 0, uint64(len(data2)))
	_, exists3After := fc.GetOldContent(pathHash3, 0, uint64(len(data3)))
	t.Logf("After eviction - File 1: %v, File 2: %v, File 3: %v", exists1After, exists2After, exists3After)
	
	if exists1After {
		t.Error("File 1 should be evicted due to global limit")
	}
	if !exists2After {
		t.Error("File 2 should be retained")
	}
	if !exists3After {
		t.Error("File 3 should be retained")
	}
	t.Log("TestGlobalLimitWithMultipleFiles completed")
}

func TestMemoryAccountingAccuracy(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(128*1024, 512*1024, time.Hour, mockTime)
	
	pathHash := uint32(5555)
	
	// Add segments and track expected memory usage
	segments := []struct {
		offset uint64
		size   int
	}{
		{0, 1000},
		{2000, 1500},    // Gap between 1000-2000
		{4000, 2000},    // Gap between 3500-4000
		{7000, 500},     // Gap between 6000-7000
	}
	
	expectedTotal := uint64(0)
	
	for _, seg := range segments {
		data := make([]byte, seg.size)
		for i := range data {
			data[i] = byte(seg.offset % 256)
		}
		
		fc.AddRead(pathHash, seg.offset, data)
		expectedTotal += uint64(seg.size)
		
		// Check memory accounting after each addition
		sf := fc.files[pathHash]
		if sf == nil {
			t.Fatal("SparseFile should exist")
		}
		
		if sf.Size != expectedTotal {
			t.Errorf("Memory accounting mismatch after adding segment at %d: expected %d, got %d",
				seg.offset, expectedTotal, sf.Size)
		}
	}
	
	// Test memory accounting after write invalidation
	// Write that splits a segment
	fc.UpdateWithWrite(pathHash, 2500, []byte("split"))
	
	// Should invalidate part of segment at 2000 (size 1500)
	// Original segment: 2000-3500, write: 2500-2505
	// Should split into: 2000-2500 (500 bytes) + 2505-3500 (995 bytes)
	// Total change: -1500 + 500 + 995 = -5 bytes
	expectedAfterWrite := expectedTotal - 1500 + 500 + 995
	
	sf := fc.files[pathHash]
	if sf.Size != expectedAfterWrite {
		t.Errorf("Memory accounting after write split: expected %d, got %d",
			expectedAfterWrite, sf.Size)
	}
}

func TestMemoryPressureDuringConcurrentOps(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	
	// Very small limits to trigger frequent evictions
	fc := NewFileCache(8*1024, 32*1024, time.Hour, mockTime)
	
	// Simulate memory pressure with concurrent-like operations
	for round := 0; round < 10; round++ {
		for fileID := 0; fileID < 8; fileID++ {
			pathHash := uint32(fileID + 1000)
			
			// Add data that will cause evictions
			data := make([]byte, 5*1024) // 5KB per operation
			for i := range data {
				data[i] = byte(round*fileID + i%256)
			}
			
			offset := uint64(round * 10000) // Different offset each round
			fc.AddRead(pathHash, offset, data)
		}
		
		// Verify cache stays within global limit
		totalSize := uint64(0)
		for _, sf := range fc.files {
			totalSize += sf.Size
		}
		
		if totalSize > 32*1024 {
			t.Errorf("Round %d: Global limit exceeded: %d > %d", round, totalSize, 32*1024)
		}
		
		// Verify per-file limits
		for pathHash, sf := range fc.files {
			if sf.Size > 8*1024 {
				t.Errorf("Round %d, file %d: Per-file limit exceeded: %d > %d",
					round, pathHash, sf.Size, 8*1024)
			}
		}
	}
}

func TestEvictionOrderingLRU(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(32*1024, 64*1024, time.Hour, mockTime)
	
	// Add files with specific timing
	files := []struct {
		pathHash uint32
		delay    time.Duration
	}{
		{1001, 0},                    // Oldest
		{1002, 1 * time.Minute},      // Middle
		{1003, 2 * time.Minute},      // Newest
	}
	
	baseTime := time.Now()
	
	for _, file := range files {
		mockTime.time = baseTime.Add(file.delay)
		
		data := make([]byte, 20*1024) // 20KB each
		for i := range data {
			data[i] = byte(file.pathHash % 256)
		}
		fc.AddRead(file.pathHash, 0, data)
	}
	
	// All should fit (60KB < 64KB global limit)
	for _, file := range files {
		_, exists := fc.GetOldContent(file.pathHash, 0, 20*1024)
		if !exists {
			t.Errorf("File %d should exist before eviction test", file.pathHash)
		}
	}
	
	// Add another file that exceeds global limit
	mockTime.Advance(1 * time.Minute)
	pathHash4 := uint32(1004)
	data4 := make([]byte, 20*1024)
	fc.AddRead(pathHash4, 0, data4)
	
	// Should evict oldest file (1001)
	_, exists1 := fc.GetOldContent(1001, 0, 20*1024)
	_, exists2 := fc.GetOldContent(1002, 0, 20*1024)
	_, exists3 := fc.GetOldContent(1003, 0, 20*1024)
	_, exists4 := fc.GetOldContent(1004, 0, 20*1024)
	
	if exists1 {
		t.Error("Oldest file (1001) should be evicted first")
	}
	if !exists2 {
		t.Error("File 1002 should be retained")
	}
	if !exists3 {
		t.Error("File 1003 should be retained")
	}
	if !exists4 {
		t.Error("Newest file 1004 should be retained")
	}
}

func TestZeroLimitsHandling(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	
	// Test with zero limits
	fc := NewFileCache(0, 0, time.Hour, mockTime)
	
	pathHash := uint32(7777)
	data := []byte("test data")
	
	// Should handle zero limits gracefully (likely by not storing anything)
	fc.AddRead(pathHash, 0, data)
	
	_, exists := fc.GetOldContent(pathHash, 0, uint64(len(data)))
	
	// With zero limits, data should not be stored
	if exists {
		t.Error("Data should not be stored with zero limits")
	}
	
	// Cache should remain empty
	if fc.Size() != 0 {
		t.Error("Cache should be empty with zero limits")
	}
}
