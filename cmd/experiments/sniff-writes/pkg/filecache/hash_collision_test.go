package filecache

import (
	"bytes"
	"testing"
	"time"
)

func TestHashCollisionHandling(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	// Simulate hash collision: different logical files with same pathHash
	samePathHash := uint32(42)

	// File 1 content
	file1Data := []byte("content from file 1")
	fc.AddRead(samePathHash, 0, file1Data)

	// File 2 content (same pathHash, different logical file)
	file2Data := []byte("content from file 2 - different file")
	fc.AddRead(samePathHash, 100, file2Data)

	// Both should be stored in the same SparseFile due to hash collision
	retrieved1, exists1 := fc.GetOldContent(samePathHash, 0, uint64(len(file1Data)))
	retrieved2, exists2 := fc.GetOldContent(samePathHash, 100, uint64(len(file2Data)))

	if !exists1 || !exists2 {
		t.Fatal("Expected both files to be stored despite hash collision")
	}

	if !bytes.Equal(file1Data, retrieved1) {
		t.Error("File 1 data corrupted due to hash collision")
	}

	if !bytes.Equal(file2Data, retrieved2) {
		t.Error("File 2 data corrupted due to hash collision")
	}

	// Write to one "file" should only affect that range
	fc.UpdateWithWrite(samePathHash, 0, []byte("overwritten file 1"))

	// File 1 should be invalidated, file 2 should remain
	_, exists1After := fc.GetOldContent(samePathHash, 0, uint64(len(file1Data)))
	retrieved2After, exists2After := fc.GetOldContent(samePathHash, 100, uint64(len(file2Data)))

	if exists1After {
		t.Error("File 1 should be invalidated by write")
	}

	if !exists2After {
		t.Error("File 2 should not be affected by write to file 1 range")
	}

	if !bytes.Equal(file2Data, retrieved2After) {
		t.Error("File 2 data corrupted by write to file 1 range")
	}
}

func TestCrossProcessHashCollisions(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	fc := NewFileCache(64*1024, 256*1024, time.Hour, mockTime)

	// Simulate same file accessed by different processes
	pathHash := uint32(777)

	processes := []struct {
		pid  uint32
		fd   int32
		data []byte
	}{
		{1001, 10, []byte("process 1001 content")},
		{1002, 15, []byte("process 1002 different content")},
		{1003, 20, []byte("process 1003 yet another content")},
	}

	// Store content from each process using legacy API
	for i, proc := range processes {
		offset := uint64(i * 1000) // Different offsets
		fc.StoreReadContent(proc.pid, proc.fd, pathHash, proc.data, offset)
	}

	// All should be accessible via new API (pathHash only)
	for i, proc := range processes {
		offset := uint64(i * 1000)
		retrieved, exists := fc.GetOldContent(pathHash, offset, uint64(len(proc.data)))

		if !exists {
			t.Errorf("Content from process %d not found", proc.pid)
			continue
		}

		if !bytes.Equal(proc.data, retrieved) {
			t.Errorf("Content from process %d corrupted", proc.pid)
		}
	}
}
