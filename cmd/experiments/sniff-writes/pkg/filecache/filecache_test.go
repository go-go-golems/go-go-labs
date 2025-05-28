package filecache

import (
	"bytes"
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

func TestSegmentMerging(t *testing.T) {
	tests := []struct {
		name     string
		existing []*Segment
		new      *Segment
		expected []*Segment
	}{
		{
			name:     "single segment to empty",
			existing: []*Segment{},
			new:      &Segment{Start: 10, End: 20, Data: []byte("hello")},
			expected: []*Segment{{Start: 10, End: 20, Data: []byte("hello")}},
		},
		{
			name: "adjacent segments merge",
			existing: []*Segment{
				{Start: 0, End: 5, Data: []byte("hello")},
			},
			new: &Segment{Start: 5, End: 10, Data: []byte("world")},
			expected: []*Segment{
				{Start: 0, End: 10, Data: []byte("helloworld")},
			},
		},
		{
			name: "overlapping segments merge",
			existing: []*Segment{
				{Start: 0, End: 11, Data: []byte("hello_world")},
			},
			new: &Segment{Start: 6, End: 17, Data: []byte("world_again")},
			expected: []*Segment{
				{Start: 0, End: 17, Data: []byte("hello_world_again")},
			},
		},
		{
			name: "identical segments merge",
			existing: []*Segment{
				{Start: 10, End: 14, Data: []byte("same")},
			},
			new: &Segment{Start: 10, End: 14, Data: []byte("same")},
			expected: []*Segment{
				{Start: 10, End: 14, Data: []byte("same")},
			},
		},
		{
			name: "fully contained segment absorbed",
			existing: []*Segment{
				{Start: 10, End: 30, Data: bytes.Repeat([]byte("a"), 20)},
			},
			new: &Segment{Start: 15, End: 20, Data: []byte("inner")},
			expected: []*Segment{
				{Start: 10, End: 30, Data: append(append(bytes.Repeat([]byte("a"), 5), []byte("inner")...), bytes.Repeat([]byte("a"), 10)...)},
			},
		},
		{
			name: "multiple segments merge cascade",
			existing: []*Segment{
				{Start: 0, End: 5, Data: []byte("first")},
				{Start: 25, End: 30, Data: []byte("third")},
			},
			new: &Segment{Start: 5, End: 25, Data: bytes.Repeat([]byte("b"), 20)},
			expected: []*Segment{
				{Start: 0, End: 30, Data: append(append([]byte("first"), bytes.Repeat([]byte("b"), 20)...), []byte("third")...)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &SparseFile{
				Segments: make([]*Segment, len(tt.existing)),
			}
			for i, seg := range tt.existing {
				sf.Segments[i] = &Segment{
					Start: seg.Start,
					End:   seg.End,
					Data:  make([]byte, len(seg.Data)),
				}
				copy(sf.Segments[i].Data, seg.Data)
			}

			newSeg := &Segment{
				Start: tt.new.Start,
				End:   tt.new.End,
				Data:  make([]byte, len(tt.new.Data)),
			}
			copy(newSeg.Data, tt.new.Data)
			sf.insertSegment(newSeg)

			if len(sf.Segments) != len(tt.expected) {
				t.Fatalf("expected %d segments, got %d", len(tt.expected), len(sf.Segments))
			}

			for i, expected := range tt.expected {
				actual := sf.Segments[i]
				if actual.Start != expected.Start || actual.End != expected.End {
					t.Errorf("segment %d: expected [%d,%d), got [%d,%d)", i, expected.Start, expected.End, actual.Start, actual.End)
				}
				if !bytes.Equal(actual.Data, expected.Data) {
					t.Errorf("segment %d: expected data %q, got %q", i, expected.Data, actual.Data)
				}
			}
		})
	}
}

func TestWriteInvalidation(t *testing.T) {
	tests := []struct {
		name     string
		existing []*Segment
		write    struct {
			offset uint64
			data   []byte
		}
		expected []*Segment
	}{
		{
			name: "write completely replaces segment",
			existing: []*Segment{
				{Start: 10, End: 20, Data: []byte("old_data")},
			},
			write: struct {
				offset uint64
				data   []byte
			}{offset: 10, data: []byte("new_data")},
			expected: []*Segment{
				{Start: 10, End: 18, Data: []byte("new_data")},
			},
		},
		{
			name: "write splits segment",
			existing: []*Segment{
				{Start: 10, End: 30, Data: bytes.Repeat([]byte("a"), 20)},
			},
			write: struct {
				offset uint64
				data   []byte
			}{offset: 15, data: []byte("NEW")},
			expected: []*Segment{
				{Start: 10, End: 15, Data: bytes.Repeat([]byte("a"), 5)},
				{Start: 15, End: 18, Data: []byte("NEW")},
				{Start: 18, End: 30, Data: bytes.Repeat([]byte("a"), 12)},
			},
		},
		{
			name: "write overlaps segment start",
			existing: []*Segment{
				{Start: 20, End: 40, Data: bytes.Repeat([]byte("x"), 20)},
			},
			write: struct {
				offset uint64
				data   []byte
			}{offset: 15, data: []byte("PREFIX")},
			expected: []*Segment{
				{Start: 15, End: 21, Data: []byte("PREFIX")},
				{Start: 21, End: 40, Data: bytes.Repeat([]byte("x"), 19)},
			},
		},
		{
			name: "write overlaps segment end",
			existing: []*Segment{
				{Start: 10, End: 30, Data: bytes.Repeat([]byte("y"), 20)},
			},
			write: struct {
				offset uint64
				data   []byte
			}{offset: 25, data: []byte("SUFFIX")},
			expected: []*Segment{
				{Start: 10, End: 25, Data: bytes.Repeat([]byte("y"), 15)},
				{Start: 25, End: 31, Data: []byte("SUFFIX")},
			},
		},
		{
			name: "write spans multiple segments",
			existing: []*Segment{
				{Start: 10, End: 20, Data: []byte("first")},
				{Start: 30, End: 40, Data: []byte("second")},
				{Start: 50, End: 60, Data: []byte("third")},
			},
			write: struct {
				offset uint64
				data   []byte
			}{offset: 15, data: []byte("REPLACEMENT")},
			expected: []*Segment{
				{Start: 10, End: 15, Data: []byte("first")[:5]},
				{Start: 15, End: 26, Data: []byte("REPLACEMENT")},
				{Start: 50, End: 60, Data: []byte("third")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &SparseFile{
				Segments: make([]*Segment, len(tt.existing)),
			}
			for i, seg := range tt.existing {
				sf.Segments[i] = &Segment{
					Start: seg.Start,
					End:   seg.End,
					Data:  make([]byte, len(seg.Data)),
				}
				copy(sf.Segments[i].Data, seg.Data)
			}

			sf.UpdateWithWrite(tt.write.offset, tt.write.data)

			if len(sf.Segments) != len(tt.expected) {
				t.Logf("Actual segments:")
				for i, seg := range sf.Segments {
					t.Logf("  [%d] [%d,%d) %q", i, seg.Start, seg.End, seg.Data)
				}
				t.Logf("Expected segments:")
				for i, seg := range tt.expected {
					t.Logf("  [%d] [%d,%d) %q", i, seg.Start, seg.End, seg.Data)
				}
				t.Fatalf("expected %d segments, got %d", len(tt.expected), len(sf.Segments))
			}

			for i, expected := range tt.expected {
				actual := sf.Segments[i]
				if actual.Start != expected.Start || actual.End != expected.End {
					t.Errorf("segment %d: expected [%d,%d), got [%d,%d)", i, expected.Start, expected.End, actual.Start, actual.End)
				}
				if !bytes.Equal(actual.Data, expected.Data) {
					t.Errorf("segment %d: expected data %q, got %q", i, expected.Data, actual.Data)
				}
			}
		})
	}
}

func TestContentReconstruction(t *testing.T) {
	tests := []struct {
		name      string
		segments  []*Segment
		request   struct {
			offset uint64
			length uint64
		}
		expected    []byte
		shouldExist bool
		gapByte     byte
	}{
		{
			name: "exact single segment match",
			segments: []*Segment{
				{Start: 10, End: 20, Data: []byte("helloworld")},
			},
			request:     struct{ offset, length uint64 }{offset: 10, length: 10},
			expected:    []byte("helloworld"),
			shouldExist: true,
		},
		{
			name: "partial segment match",
			segments: []*Segment{
				{Start: 10, End: 30, Data: []byte("helloworldagain")},
			},
			request:     struct{ offset, length uint64 }{offset: 15, length: 5},
			expected:    []byte("world"),
			shouldExist: true,
		},
		{
			name: "multiple adjacent segments",
			segments: []*Segment{
				{Start: 0, End: 5, Data: []byte("hello")},
				{Start: 5, End: 10, Data: []byte("world")},
			},
			request:     struct{ offset, length uint64 }{offset: 0, length: 10},
			expected:    []byte("helloworld"),
			shouldExist: true,
		},
		{
			name: "segments with gap",
			segments: []*Segment{
				{Start: 0, End: 5, Data: []byte("hello")},
				{Start: 10, End: 15, Data: []byte("world")},
			},
			request:     struct{ offset, length uint64 }{offset: 0, length: 15},
			expected:    []byte("hello?????world"),
			shouldExist: true,
			gapByte:     '?',
		},
		{
			name: "request with no coverage",
			segments: []*Segment{
				{Start: 100, End: 110, Data: []byte("faraway")},
			},
			request:     struct{ offset, length uint64 }{offset: 0, length: 10},
			expected:    nil,
			shouldExist: false,
		},
		{
			name: "partial coverage at start",
			segments: []*Segment{
				{Start: 5, End: 15, Data: []byte("helloworld")},
			},
			request:     struct{ offset, length uint64 }{offset: 0, length: 10},
			expected:    []byte("?????hello"),
			shouldExist: true,
			gapByte:     '?',
		},
		{
			name: "partial coverage at end",
			segments: []*Segment{
				{Start: 0, End: 5, Data: []byte("hello")},
			},
			request:     struct{ offset, length uint64 }{offset: 0, length: 10},
			expected:    []byte("hello?????"),
			shouldExist: true,
			gapByte:     '?',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &SparseFile{
				Segments: make([]*Segment, len(tt.segments)),
			}
			for i, seg := range tt.segments {
				sf.Segments[i] = &Segment{
					Start: seg.Start,
					End:   seg.End,
					Data:  make([]byte, len(seg.Data)),
				}
				copy(sf.Segments[i].Data, seg.Data)
			}

			gapByte := tt.gapByte
			if gapByte == 0 {
				gapByte = 0x00
			}

			result, exists := sf.GetContentRange(tt.request.offset, tt.request.length, gapByte)

			if exists != tt.shouldExist {
				t.Errorf("expected exists=%v, got %v", tt.shouldExist, exists)
				return
			}

			if !exists {
				return
			}

			if !bytes.Equal(result, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}

			if uint64(len(result)) != tt.request.length {
				t.Errorf("expected length %d, got %d", tt.request.length, len(result))
			}
		})
	}
}

func TestCacheExpiration(t *testing.T) {
	mockTime := NewMockTimeProvider(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
	fc := NewWithTimeProvider(mockTime)
	fc.maxAge = 1 * time.Minute

	// Add some test data
	fc.AddRead(123, 0, []byte("hello"))
	fc.AddRead(123, 10, []byte("world"))

	// Verify data exists
	content, exists := fc.GetOldContent(123, 0, 15)
	if !exists {
		t.Fatal("content should exist initially")
	}
	if !bytes.Contains(content, []byte("hello")) {
		t.Error("should contain hello")
	}

	// Advance time past expiration
	mockTime.Advance(2 * time.Minute)

	// Cleanup expired entries
	fc.Cleanup()

	// Verify data is gone
	_, exists = fc.GetOldContent(123, 0, 15)
	if exists {
		t.Error("content should be expired")
	}
}

func TestMemoryLimits(t *testing.T) {
	fc := New()
	fc.perFileLimit = 100 // 100 bytes per file
	fc.globalLimit = 300  // 300 bytes total

	pathHash1 := uint32(111)
	pathHash2 := uint32(222)

	// Fill up first file to limit
	data50 := bytes.Repeat([]byte("a"), 50)
	fc.AddRead(pathHash1, 0, data50)
	fc.AddRead(pathHash1, 50, data50)

	// Verify both segments exist
	content, exists := fc.GetOldContent(pathHash1, 0, 100)
	if !exists || len(content) != 100 {
		t.Error("should have full content")
	}

	// Add more data that should trigger per-file eviction
	data60 := bytes.Repeat([]byte("b"), 60)
	fc.AddRead(pathHash1, 100, data60)

	// First segment should be evicted
	content, exists = fc.GetOldContent(pathHash1, 0, 50)
	if exists && bytes.Contains(content, []byte("a")) {
		t.Error("old data should be evicted due to per-file limit")
	}

	// Fill up second file
	fc.AddRead(pathHash2, 0, data50)
	fc.AddRead(pathHash2, 50, data50)

	// Add third file that should trigger global eviction
	pathHash3 := uint32(333)
	fc.AddRead(pathHash3, 0, data50)
	fc.AddRead(pathHash3, 50, data50)

	// Check that total cache size respects global limit
	if fc.Size() > int(fc.globalLimit) {
		t.Errorf("cache size %d exceeds global limit %d", fc.Size(), fc.globalLimit)
	}
}

func TestConcurrentAccess(t *testing.T) {
	fc := New()
	pathHash := uint32(123)

	done := make(chan bool, 3)

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			data := []byte("concurrent_read_data")
			fc.AddRead(pathHash, uint64(i*20), data)
		}
		done <- true
	}()

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			data := []byte("concurrent_write_data")
			fc.UpdateWithWrite(pathHash, uint64(i*15), data)
		}
		done <- true
	}()

	// Concurrent content retrieval
	go func() {
		for i := 0; i < 100; i++ {
			fc.GetOldContent(pathHash, uint64(i*10), 20)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify cache is still consistent
	if fc.Size() < 0 {
		t.Error("cache size should not be negative")
	}
}

func TestAPICompatibility(t *testing.T) {
	fc := New()

	// Test old API still works
	fc.StoreReadContent(123, 45, 789, []byte("test"), 0)
	content, exists := fc.GetContentForDiff(123, 45, 789, 0)
	if !exists {
		t.Error("old API should still work")
	}
	if !bytes.Equal(content.Content, []byte("test")) {
		t.Error("old API should return correct content")
	}

	// Test diff generation still works
	diff, hasDiff := fc.GenerateDiff(123, 45, 789, 0, []byte("modified"))
	if !hasDiff {
		t.Error("should generate diff")
	}
	if len(diff) == 0 {
		t.Error("diff should not be empty")
	}
}

func BenchmarkSegmentInsertion(b *testing.B) {
	sf := &SparseFile{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := []byte("benchmark_data")
		seg := &Segment{
			Start: uint64(i * 20),
			End:   uint64(i*20 + len(data)),
			Data:  data,
		}
		sf.insertSegment(seg)
	}
}

func BenchmarkContentReconstruction(b *testing.B) {
	sf := &SparseFile{}

	// Prepare segments
	for i := 0; i < 100; i++ {
		data := bytes.Repeat([]byte("x"), 100)
		seg := &Segment{
			Start: uint64(i * 100),
			End:   uint64((i + 1) * 100),
			Data:  data,
		}
		sf.insertSegment(seg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sf.GetContentRange(0, 10000, 0x00)
	}
}
