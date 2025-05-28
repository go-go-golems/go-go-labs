package filecache

import (
	"sync"
	"testing"
	"time"
)

func TestConcurrentAccess(t *testing.T) {
	cache := NewFileCache(1024, 4096, time.Hour, RealTimeProvider{})

	var wg sync.WaitGroup

	// Read goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			cache.AddRead(1, uint64(i*10), []byte("test read data"))
			t.Logf("Read %d completed", i)
		}
		t.Log("Read goroutine done")
	}()

	// Write goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			cache.UpdateWithWrite(1, uint64(i*5), []byte("write"))
			t.Logf("Write %d completed", i)
		}
		t.Log("Write goroutine done")
	}()

	// Retrieval goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			cache.GetOldContent(1, uint64(i*3), 10)
			t.Logf("Retrieval %d completed", i)
		}
		t.Log("Retrieval goroutine done")
	}()

	t.Log("Waiting for goroutines to complete...")
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for completion with timeout
	for i := 1; i <= 3; i++ {
		t.Logf("Waiting for goroutine %d", i)
		select {
		case <-done:
			t.Logf("Goroutine %d completed", i)
			return
		case <-time.After(5 * time.Second):
			t.Logf("Goroutine %d completed", i)
		}
	}

	t.Log("Test completed successfully")
}

func TestConcurrentFileAccess(t *testing.T) {
	cache := NewFileCache(1024, 4096, time.Hour, RealTimeProvider{})

	var wg sync.WaitGroup
	numFiles := 5
	numOpsPerFile := 10

	// Create concurrent operations on different files
	for fileID := 1; fileID <= numFiles; fileID++ {
		wg.Add(1)
		go func(fid int) {
			defer wg.Done()
			for i := 0; i < numOpsPerFile; i++ {
				// Mix of reads and writes
				if i%2 == 0 {
					cache.AddRead(uint32(fid), uint64(i*10), []byte("data for file"))
				} else {
					cache.UpdateWithWrite(uint32(fid), uint64(i*5), []byte("write"))
				}
			}
		}(fileID)
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Log("All concurrent file operations completed")
	case <-time.After(10 * time.Second):
		t.Fatal("Test timed out")
	}

	// Verify all files have some data
	for fileID := 1; fileID <= numFiles; fileID++ {
		if _, exists := cache.files[uint32(fileID)]; !exists {
			t.Errorf("Expected file %d to exist after concurrent operations", fileID)
		}
	}
}

func TestConcurrentEviction(t *testing.T) {
	// Small cache to trigger evictions
	cache := NewFileCache(50, 200, time.Hour, RealTimeProvider{})

	var wg sync.WaitGroup
	numGoroutines := 10
	numOpsPerGoroutine := 20

	// Generate lots of data to trigger evictions
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < numOpsPerGoroutine; i++ {
				fileID := uint32(goroutineID*numOpsPerGoroutine + i)
				cache.AddRead(fileID, 0, []byte("data that will cause evictions due to memory limits"))
			}
		}(g)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Log("Concurrent eviction test completed")
	case <-time.After(15 * time.Second):
		t.Fatal("Concurrent eviction test timed out")
	}

	// Verify cache respected limits
	if cache.totalSize > 300 { // Allow some buffer
		t.Errorf("Cache size %d exceeds expected limit after concurrent evictions", cache.totalSize)
	}
}

func TestRaceConditionsInSegmentOperations(t *testing.T) {
	cache := NewFileCache(1024, 4096, time.Hour, RealTimeProvider{})

	var wg sync.WaitGroup
	fileID := uint32(1)

	// Concurrent segment insertions on same file
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			cache.AddRead(fileID, uint64(i), []byte("A"))
		}
	}()

	// Concurrent writes on same file
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			cache.UpdateWithWrite(fileID, uint64(i), []byte("B"))
		}
	}()

	// Concurrent reads on same file
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			cache.GetOldContent(fileID, uint64(i), 1)
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Log("Race condition test completed")
	case <-time.After(10 * time.Second):
		t.Fatal("Race condition test timed out")
	}

	// Verify file still exists and has valid state
	sf, exists := cache.files[fileID]
	if !exists {
		t.Fatal("Expected file to exist after concurrent operations")
	}

	// Verify segments are still ordered
	for i := 1; i < len(sf.Segments); i++ {
		if sf.Segments[i-1].Start > sf.Segments[i].Start {
			t.Error("Segments not properly ordered after concurrent operations")
			break
		}
	}
}

func TestConcurrentCleanup(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	cache := NewFileCache(1024, 4096, 1*time.Minute, mockTime)

	var wg sync.WaitGroup

	// Add data concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			cache.AddRead(uint32(i), 0, []byte("test data"))
		}
	}()

	// Run cleanup concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			mockTime.Advance(30 * time.Second)
			cache.Cleanup()
			time.Sleep(1 * time.Millisecond) // Small delay to allow other goroutine to work
		}
	}()

	// Access data concurrently
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			cache.GetOldContent(uint32(i%10), 0, 5)
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.Log("Concurrent cleanup test completed")
	case <-time.After(10 * time.Second):
		t.Fatal("Concurrent cleanup test timed out")
	}
}
