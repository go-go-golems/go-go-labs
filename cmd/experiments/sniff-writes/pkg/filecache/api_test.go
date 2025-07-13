package filecache

import (
	"testing"
	"time"
)

func TestAPICompatibility(t *testing.T) {
	cache := NewFileCache(1024, 4096, time.Hour, RealTimeProvider{})

	// Test basic API workflow
	pathHash := uint32(12345)
	offset := uint64(100)
	data := []byte("test content for API")

	// Add read data
	cache.AddRead(pathHash, offset, data)

	// Retrieve content
	result, exists := cache.GetOldContent(pathHash, offset, uint64(len(data)))
	if !exists {
		t.Fatal("expected content to exist")
	}

	if string(result) != string(data) {
		t.Errorf("expected %q, got %q", data, result)
	}

	// Update with write
	newData := []byte("updated content")
	cache.UpdateWithWrite(pathHash, offset, newData)

	// Verify updated content
	result, exists = cache.GetOldContent(pathHash, offset, uint64(len(newData)))
	if !exists {
		t.Fatal("expected updated content to exist")
	}

	if string(result) != string(newData) {
		t.Errorf("expected %q, got %q", newData, result)
	}
}

func TestNewFileCacheParameters(t *testing.T) {
	tests := []struct {
		name         string
		perFileLimit uint64
		globalLimit  uint64
		maxAge       time.Duration
		shouldPanic  bool
	}{
		{
			name:         "valid parameters",
			perFileLimit: 1024,
			globalLimit:  4096,
			maxAge:       time.Hour,
			shouldPanic:  false,
		},
		{
			name:         "zero per-file limit",
			perFileLimit: 0,
			globalLimit:  4096,
			maxAge:       time.Hour,
			shouldPanic:  false, // Should handle gracefully
		},
		{
			name:         "global limit smaller than per-file",
			perFileLimit: 2048,
			globalLimit:  1024,
			maxAge:       time.Hour,
			shouldPanic:  false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.shouldPanic {
					t.Errorf("unexpected panic: %v", r)
				} else if r == nil && tt.shouldPanic {
					t.Error("expected panic but none occurred")
				}
			}()

			cache := NewFileCache(tt.perFileLimit, tt.globalLimit, tt.maxAge, RealTimeProvider{})
			if cache == nil {
				t.Error("NewFileCache returned nil")
			}
		})
	}
}

func TestRealTimeProvider(t *testing.T) {
	provider := RealTimeProvider{}

	start := provider.Now()
	time.Sleep(1 * time.Millisecond)
	end := provider.Now()

	if !end.After(start) {
		t.Error("RealTimeProvider should return increasing time")
	}
}

func TestFileHashingAndIdentification(t *testing.T) {
	cache := NewFileCache(1024, 4096, time.Hour, RealTimeProvider{})

	// Test that different path hashes create separate files
	hash1 := uint32(1)
	hash2 := uint32(2)

	cache.AddRead(hash1, 0, []byte("file 1 content"))
	cache.AddRead(hash2, 0, []byte("file 2 content"))

	// Should have two separate files
	if len(cache.files) != 2 {
		t.Errorf("expected 2 files, got %d", len(cache.files))
	}

	// Each should have their own content
	result1, exists1 := cache.GetOldContent(hash1, 0, 14)
	result2, exists2 := cache.GetOldContent(hash2, 0, 14)

	if !exists1 || !exists2 {
		t.Fatal("both files should exist")
	}

	if string(result1) != "file 1 content" {
		t.Errorf("file 1 content mismatch: got %q", result1)
	}

	if string(result2) != "file 2 content" {
		t.Errorf("file 2 content mismatch: got %q", result2)
	}
}

func TestSegmentDataIntegrity(t *testing.T) {
	cache := NewFileCache(1024, 4096, time.Hour, RealTimeProvider{})

	// Test with binary data including null bytes
	binaryData := []byte{0x00, 0x01, 0xFF, 0xFE, 0x42, 0x00, 0x7F}
	pathHash := uint32(1)

	cache.AddRead(pathHash, 0, binaryData)

	result, exists := cache.GetOldContent(pathHash, 0, uint64(len(binaryData)))
	if !exists {
		t.Fatal("binary data should exist")
	}

	if len(result) != len(binaryData) {
		t.Fatalf("length mismatch: expected %d, got %d", len(binaryData), len(result))
	}

	for i, expected := range binaryData {
		if result[i] != expected {
			t.Errorf("byte %d: expected 0x%02X, got 0x%02X", i, expected, result[i])
		}
	}
}

func TestEmptyAndNilData(t *testing.T) {
	cache := NewFileCache(1024, 4096, time.Hour, RealTimeProvider{})

	t.Run("empty data", func(t *testing.T) {
		pathHash := uint32(1)
		cache.AddRead(pathHash, 0, []byte{})

		// Should handle empty data gracefully
		result, exists := cache.GetOldContent(pathHash, 0, 0)
		if exists && len(result) != 0 {
			t.Error("empty data should return empty result")
		}
	})

	t.Run("nil data", func(t *testing.T) {
		pathHash := uint32(2)
		cache.AddRead(pathHash, 0, nil)

		// Should handle nil data gracefully
		_, exists := cache.GetOldContent(pathHash, 0, 1)
		if exists {
			t.Error("nil data should not create segments")
		}
	})
}
