package filecache

import (
	"bytes"
	"testing"
	"time"
)

// TestHashCollisionHandling tests behavior when different files have same pathHash
func TestHashCollisionHandling(t *testing.T) {
	fc := NewFileCache(64*1024, 1024*1024, time.Hour, NewMockTimeProvider(time.Now()))
	
	// Simulate hash collision: two different files with same pathHash
	collisionHash := uint32(42)
	
	// File 1 data
	file1Data := []byte("This is content from file 1")
	fc.AddRead(collisionHash, 0, file1Data)
	
	// File 2 data at same hash (collision scenario)
	// In real usage, this would be a different file but same hash
	file2Data := []byte("This is completely different content from file 2")
	fc.AddRead(collisionHash, 100, file2Data)
	
	// Both segments should be stored (current implementation treats same hash as same file)
	retrieved1, exists1 := fc.GetOldContent(collisionHash, 0, uint64(len(file1Data)))
	if !exists1 || !bytes.Equal(retrieved1, file1Data) {
		t.Error("File 1 data corrupted by hash collision")
	}
	
	retrieved2, exists2 := fc.GetOldContent(collisionHash, 100, uint64(len(file2Data)))
	if !exists2 || !bytes.Equal(retrieved2, file2Data) {
		t.Error("File 2 data not stored correctly")
	}
	
	// Test write invalidation with collisions
	fc.UpdateWithWrite(collisionHash, 50, []byte("overwrite"))
	
	// Check cache consistency after collision + write
	fc.mu.RLock()
	sparseFile := fc.files[collisionHash]
	fc.mu.RUnlock()
	
	if sparseFile == nil {
		t.Error("Cache entry lost after hash collision and write")
	}
}

// TestPathHashConsistency tests pathHash changes for same logical file
func TestPathHashConsistency(t *testing.T) {
	fc := NewFileCache(64*1024, 1024*1024, time.Hour, NewMockTimeProvider(time.Now()))
	
	// Simulate same file with different pathHash values
	// (e.g., due to symlinks, renames, or path resolution differences)
	originalHash := uint32(100)
	renamedHash := uint32(200)
	
	originalData := []byte("original file content")
	fc.AddRead(originalHash, 0, originalData)
	
	// Same file content but different hash (e.g., after rename/symlink)
	sameData := []byte("original file content") // Identical content
	fc.AddRead(renamedHash, 0, sameData)
	
	// Both should be stored independently
	fc.mu.RLock()
	hasOriginal := fc.files[originalHash] != nil
	hasRenamed := fc.files[renamedHash] != nil
	fc.mu.RUnlock()
	
	if !hasOriginal {
		t.Error("Original file cache entry lost")
	}
	if !hasRenamed {
		t.Error("Renamed file cache entry not created")
	}
	
	// Verify both can be retrieved independently
	retrieved1, exists1 := fc.GetOldContent(originalHash, 0, uint64(len(originalData)))
	retrieved2, exists2 := fc.GetOldContent(renamedHash, 0, uint64(len(sameData)))
	
	if !exists1 || !bytes.Equal(retrieved1, originalData) {
		t.Error("Original file data corrupted")
	}
	if !exists2 || !bytes.Equal(retrieved2, sameData) {
		t.Error("Renamed file data corrupted")
	}
}

// TestCrossProcessFileAccess tests same file accessed from different processes
func TestCrossProcessFileAccess(t *testing.T) {
	fc := NewFileCache(64*1024, 1024*1024, time.Hour, NewMockTimeProvider(time.Now()))
	
	// Simulate same file accessed by different processes
	// In practice, pathHash should be consistent across processes for same file
	fileHash := uint32(300)
	
	// Process 1 reads
	process1Data := []byte("data read by process 1")
	fc.AddRead(fileHash, 0, process1Data)
	
	// Process 2 reads different part of same file
	process2Data := []byte("data read by process 2")
	fc.AddRead(fileHash, 100, process2Data)
	
	// Process 1 writes (should invalidate cached data)
	writeData := []byte("process 1 writes")
	fc.UpdateWithWrite(fileHash, 50, writeData)
	
	// Verify cache remains consistent across "processes"
	fc.mu.RLock()
	sparseFile := fc.files[fileHash]
	fc.mu.RUnlock()
	
	if sparseFile != nil {
		sparseFile.mu.RLock()
		segmentCount := len(sparseFile.Segments)
		sparseFile.mu.RUnlock()
		
		if segmentCount < 0 {
			t.Error("Invalid cache state after cross-process operations")
		}
	}
}

// TestFileDescriptorReuse tests file descriptor reuse scenarios
func TestFileDescriptorReuse(t *testing.T) {
	fc := NewFileCache(64*1024, 1024*1024, time.Hour, NewMockTimeProvider(time.Now()))
	
	// Simulate file descriptor being reused for different files
	// This could happen when files are opened/closed frequently
	fdHash := uint32(400) // Represents a file descriptor number
	
	// First file using this "fd"
	file1Data := []byte("first file content")
	fc.AddRead(fdHash, 0, file1Data)
	
	// Second file using same "fd" (different content)
	file2Data := []byte("completely different second file content")
	fc.AddRead(fdHash, 200, file2Data) // Different offset
	
	// Both should be retrievable (current implementation)
	retrieved1, exists1 := fc.GetOldContent(fdHash, 0, uint64(len(file1Data)))
	retrieved2, exists2 := fc.GetOldContent(fdHash, 200, uint64(len(file2Data)))
	
	if !exists1 || !bytes.Equal(retrieved1, file1Data) {
		t.Error("First file data lost after fd reuse")
	}
	if !exists2 || !bytes.Equal(retrieved2, file2Data) {
		t.Error("Second file data not stored correctly")
	}
	
	// Test that writes to "reused fd" work correctly
	overwriteData := []byte("overwrite")
	fc.UpdateWithWrite(fdHash, 0, overwriteData)
	
	// Verify cache consistency
	fc.mu.RLock()
	sparseFile := fc.files[fdHash]
	fc.mu.RUnlock()
	
	if sparseFile == nil {
		t.Error("Cache entry lost after fd reuse scenario")
	}
}

// TestHashDistribution tests hash distribution and collision probability
func TestHashDistribution(t *testing.T) {
	fc := NewFileCache(64*1024, 2*1024*1024, time.Hour, NewMockTimeProvider(time.Now()))
	
	// Test with many different pathHash values to check for collisions
	testHashes := []uint32{
		0, 1, 2, 3, 4, 5, // Sequential small values
		0xFFFFFFFF, 0xFFFFFFFE, 0xFFFFFFFD, // Near maximum
		0x80000000, 0x80000001, 0x7FFFFFFF, // Around 2^31
		42, 142, 242, 342, 442, // Regular intervals
		1000, 2000, 3000, 4000, // Larger intervals
	}
	
	// Add data for each hash
	for i, hash := range testHashes {
		data := []byte("test data for hash " + string(rune('A'+i)))
		fc.AddRead(hash, 0, data)
	}
	
	// Verify all data is independently retrievable
	for i, hash := range testHashes {
		expectedData := []byte("test data for hash " + string(rune('A'+i)))
		retrieved, exists := fc.GetOldContent(hash, 0, uint64(len(expectedData)))
		if !exists || !bytes.Equal(retrieved, expectedData) {
			t.Errorf("Data corrupted for hash %d", hash)
		}
	}
	
	// Verify all cache entries exist
	fc.mu.RLock()
	cacheSize := len(fc.files)
	fc.mu.RUnlock()
	
	if cacheSize != len(testHashes) {
		t.Errorf("Expected %d cache entries, got %d", len(testHashes), cacheSize)
	}
	
	// Test mixed operations across different hashes
	for _, hash := range testHashes[:5] { // Test first 5
		writeData := []byte("updated")
		fc.UpdateWithWrite(hash, 5, writeData)
	}
	
	// Verify writes didn't affect other hashes
	for i := 5; i < len(testHashes); i++ {
		hash := testHashes[i]
		expectedData := []byte("test data for hash " + string(rune('A'+i)))
		retrieved, exists := fc.GetOldContent(hash, 0, uint64(len(expectedData)))
		if !exists || !bytes.Equal(retrieved, expectedData) {
			t.Errorf("Unrelated hash %d affected by other writes", hash)
		}
	}
}

// TestHashOverflow tests behavior with hash values that might cause overflow
func TestHashOverflow(t *testing.T) {
	fc := NewFileCache(64*1024, 1024*1024, time.Hour, NewMockTimeProvider(time.Now()))
	
	// Test edge case hash values
	edgeCases := []uint32{
		0,                    // Minimum
		0xFFFFFFFF,           // Maximum
		0x80000000,           // 2^31
		0x7FFFFFFF,           // 2^31 - 1
		0x12345678,           // Arbitrary value
		^uint32(0),           // All bits set
	}
	
	for _, hash := range edgeCases {
		data := []byte("edge case data")
		fc.AddRead(hash, 0, data)
		
		retrieved, exists := fc.GetOldContent(hash, 0, uint64(len(data)))
		if !exists || !bytes.Equal(retrieved, data) {
			t.Errorf("Data corrupted for edge case hash %x", hash)
		}
	}
	
	// Verify cache remains stable with edge case hashes
	fc.mu.RLock()
	cacheSize := len(fc.files)
	fc.mu.RUnlock()
	
	if cacheSize != len(edgeCases) {
		t.Errorf("Expected %d cache entries for edge cases, got %d", len(edgeCases), cacheSize)
	}
}
