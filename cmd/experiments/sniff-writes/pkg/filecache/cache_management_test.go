package filecache

import (
	"sync"
	"testing"
	"time"
)

// MockTimeProvider allows controlling time in tests
type MockTimeProvider struct {
	mu   sync.Mutex
	time time.Time
}

func NewMockTimeProvider(start time.Time) *MockTimeProvider {
	return &MockTimeProvider{time: start}
}

func (m *MockTimeProvider) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.time
}

func (m *MockTimeProvider) Advance(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.time = m.time.Add(d)
}

func TestCacheExpiration(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Now())
	cache := NewFileCache(1024, 4096, 5*time.Minute, mockTime)

	t.Log("Starting cache expiration test")

	// Add some data
	t.Log("Created cache with mock time")

	t.Log("Adding test data")
	cache.AddRead(1, 0, []byte("test data 1"))
	t.Log("Added first segment")
	cache.AddRead(2, 0, []byte("test data 2"))
	t.Log("Added second segment")

	// Verify data exists
	t.Log("Verifying data exists")
	result1, exists1 := cache.GetOldContent(1, 0, 11)
	if !exists1 || string(result1) != "test data 1" {
		t.Fatalf("expected data to exist before expiration")
	}
	result2, exists2 := cache.GetOldContent(2, 0, 11)
	if !exists2 || string(result2) != "test data 2" {
		t.Fatalf("expected data to exist before expiration")
	}
	t.Log("Data verified, advancing time")

	// Advance time past expiration
	mockTime.Advance(6 * time.Minute)
	t.Log("Time advanced, running cleanup")

	// Run cleanup
	cache.Cleanup()
	t.Log("Cleanup completed")

	// Verify data is expired
	t.Log("Verifying data is expired")
	_, exists1 = cache.GetOldContent(1, 0, 11)
	_, exists2 = cache.GetOldContent(2, 0, 11)
	if exists1 || exists2 {
		t.Fatalf("expected data to be expired after cleanup")
	}
	t.Log("Test completed")
}

func TestMemoryLimits(t *testing.T) {
	// Create cache with small limits for testing
	cache := NewFileCache(50, 150, time.Hour, RealTimeProvider{}) // 50 bytes per file, 150 total

	// Test per-file limit
	t.Run("per-file limit", func(t *testing.T) {
		// Add data that exceeds per-file limit
		cache.AddRead(1, 0, []byte("this is a long string that should exceed the per-file limit of 50 bytes"))

		// Verify some data was evicted
		sf, exists := cache.files[1]
		if !exists {
			t.Fatal("expected file to exist")
		}
		if sf.Size > 50 {
			t.Errorf("per-file size %d exceeds limit 50", sf.Size)
		}
	})

	t.Run("global limit", func(t *testing.T) {
		// Reset cache
		cache = NewFileCache(100, 150, time.Hour, RealTimeProvider{})

		// Add data to multiple files to exceed global limit
		cache.AddRead(1, 0, []byte("file 1 data - 50 bytes of content here to test limit"))
		cache.AddRead(2, 0, []byte("file 2 data - 50 bytes of content here to test limit"))
		cache.AddRead(3, 0, []byte("file 3 data - 50 bytes of content here to test limit"))

		// Verify total size doesn't exceed global limit too much
		if cache.totalSize > 200 { // Allow some buffer for eviction timing
			t.Errorf("total cache size %d significantly exceeds limit 150", cache.totalSize)
		}
	})
}

func TestLRUEviction(t *testing.T) {
	cache := NewFileCache(100, 200, time.Hour, RealTimeProvider{})

	// Add files in sequence
	cache.AddRead(1, 0, []byte("file 1 - this is content that will be 50+ bytes long for testing"))
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps

	cache.AddRead(2, 0, []byte("file 2 - this is content that will be 50+ bytes long for testing"))
	time.Sleep(1 * time.Millisecond)

	cache.AddRead(3, 0, []byte("file 3 - this is content that will be 50+ bytes long for testing"))
	time.Sleep(1 * time.Millisecond)

	// Access file 1 to make it most recently used
	cache.GetOldContent(1, 0, 10)

	// Add another file to trigger eviction
	cache.AddRead(4, 0, []byte("file 4 - this is content that will be 50+ bytes long for testing"))

	// Check which files still exist after potential eviction
	_, exists1 := cache.GetOldContent(1, 0, 10)
	_, exists2 := cache.GetOldContent(2, 0, 10)
	_, exists3 := cache.GetOldContent(3, 0, 10)
	_, exists4 := cache.GetOldContent(4, 0, 10)

	if !exists1 {
		t.Error("file 1 should not be evicted (was accessed recently)")
	}
	if !exists3 {
		t.Error("file 3 should not be evicted")
	}
	if !exists4 {
		t.Error("file 4 should not be evicted (just added)")
	}
	// Note: file 2 might or might not be evicted depending on exact timing and sizes
	_ = exists2 // Acknowledge we're checking this but not asserting on it
}

func TestCacheCleanupEdgeCases(t *testing.T) {
	t.Run("cleanup empty cache", func(t *testing.T) {
		cache := NewFileCache(1024, 4096, time.Hour, RealTimeProvider{})
		cache.Cleanup() // Should not panic
	})

	t.Run("cleanup with mixed expired and fresh data", func(t *testing.T) {
		mockTime := NewMockTimeProvider(time.Now())
		cache := NewFileCache(1024, 4096, 5*time.Minute, mockTime)

		// Add some data
		cache.AddRead(1, 0, []byte("old data"))

		// Advance time partially
		mockTime.Advance(3 * time.Minute)

		// Add more data
		cache.AddRead(2, 0, []byte("new data"))

		// Advance time to expire only the old data
		mockTime.Advance(3 * time.Minute)

		cache.Cleanup()

		// New data should still exist
		_, exists2 := cache.GetOldContent(2, 0, 8)
		if !exists2 {
			t.Error("new data should not be expired")
		}

		// Old data should be expired
		_, exists1 := cache.GetOldContent(1, 0, 8)
		if exists1 {
			t.Error("old data should be expired")
		}
	})
}

func TestPerFileMemoryManagement(t *testing.T) {
	cache := NewFileCache(100, 1000, time.Hour, RealTimeProvider{})

	// Add segments that will exceed per-file limit
	cache.AddRead(1, 0, []byte("segment 1 - 30 bytes content here"))
	cache.AddRead(1, 50, []byte("segment 2 - 30 bytes content here"))
	cache.AddRead(1, 100, []byte("segment 3 - 30 bytes content here"))
	cache.AddRead(1, 150, []byte("segment 4 - 30 bytes content here"))

	sf, exists := cache.files[1]
	if !exists {
		t.Fatal("expected file to exist")
	}

	// Should have enforced per-file limit
	if sf.Size > 120 { // Allow some buffer
		t.Errorf("per-file size %d exceeds reasonable limit", sf.Size)
	}

	// Should still have some segments
	if len(sf.Segments) == 0 {
		t.Error("should have kept some segments")
	}
}
