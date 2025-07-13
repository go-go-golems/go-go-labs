package filecache

import (
	"bytes"
	"testing"
)

func TestContentReconstruction(t *testing.T) {
	tests := []struct {
		name       string
		segments   []*Segment
		offset     uint64
		length     uint64
		gapByte    byte
		expectedOk bool
		expected   []byte
	}{
		{
			name: "exact single segment match",
			segments: []*Segment{
				{Start: 10, End: 20, Data: []byte("hello world")},
			},
			offset:     10,
			length:     10,
			gapByte:    0x00,
			expectedOk: true,
			expected:   []byte("hello worl"),
		},
		{
			name: "partial segment match",
			segments: []*Segment{
				{Start: 10, End: 30, Data: []byte("hello world testing 123")},
			},
			offset:     15,
			length:     10,
			gapByte:    0x00,
			expectedOk: true,
			expected:   []byte(" world tes"),
		},
		{
			name: "multiple adjacent segments",
			segments: []*Segment{
				{Start: 10, End: 15, Data: []byte("hello")},
				{Start: 15, End: 20, Data: []byte("world")},
			},
			offset:     10,
			length:     10,
			gapByte:    0x00,
			expectedOk: true,
			expected:   []byte("helloworld"),
		},
		{
			name: "segments with gap",
			segments: []*Segment{
				{Start: 10, End: 15, Data: []byte("hello")},
				{Start: 20, End: 25, Data: []byte("world")},
			},
			offset:     10,
			length:     15,
			gapByte:    '?',
			expectedOk: true,
			expected:   []byte("hello?????world"),
		},
		{
			name:       "request with no coverage",
			segments:   []*Segment{},
			offset:     100,
			length:     10,
			gapByte:    0xFF,
			expectedOk: false,
			expected:   nil,
		},
		{
			name: "partial coverage at start",
			segments: []*Segment{
				{Start: 15, End: 25, Data: []byte("partial")},
			},
			offset:     10,
			length:     10,
			gapByte:    '_',
			expectedOk: true,
			expected:   []byte("_____parti"),
		},
		{
			name: "partial coverage at end",
			segments: []*Segment{
				{Start: 10, End: 15, Data: []byte("start")},
			},
			offset:     10,
			length:     10,
			gapByte:    '-',
			expectedOk: true,
			expected:   []byte("start-----"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &SparseFile{
				Segments: make([]*Segment, len(tt.segments)),
			}

			// Copy segments
			for i, seg := range tt.segments {
				sf.Segments[i] = &Segment{
					Start: seg.Start,
					End:   seg.End,
					Data:  make([]byte, len(seg.Data)),
				}
				copy(sf.Segments[i].Data, seg.Data)
			}

			result, ok := sf.GetContentRange(tt.offset, tt.length, tt.gapByte)

			if ok != tt.expectedOk {
				t.Fatalf("expected ok=%v, got ok=%v", tt.expectedOk, ok)
			}

			if tt.expectedOk {
				if !bytes.Equal(result, tt.expected) {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestContentReconstructionComplexGaps(t *testing.T) {
	tests := []struct {
		name     string
		segments []*Segment
		offset   uint64
		length   uint64
		expected string // For easier reading in tests
	}{
		{
			name: "multiple gaps pattern",
			segments: []*Segment{
				{Start: 0, End: 10, Data: []byte("0123456789")},
				{Start: 30, End: 40, Data: []byte("abcdefghij")},
				{Start: 70, End: 80, Data: []byte("ABCDEFGHIJ")},
			},
			offset:   0,
			length:   80,
			expected: "0123456789--------------------abcdefghij------------------------------ABCDEFGHIJ",
		},
		{
			name: "interleaved gaps and data",
			segments: []*Segment{
				{Start: 5, End: 10, Data: []byte("AAAAA")},
				{Start: 15, End: 20, Data: []byte("BBBBB")},
				{Start: 25, End: 30, Data: []byte("CCCCC")},
			},
			offset:   0,
			length:   35,
			expected: "-----AAAAA-----BBBBB-----CCCCC-----",
		},
		{
			name: "overlapping reconstruction window",
			segments: []*Segment{
				{Start: 0, End: 50, Data: make([]byte, 50)},    // All zeros
				{Start: 100, End: 150, Data: make([]byte, 50)}, // All zeros
			},
			offset:   25,
			length:   100,
			expected: string(make([]byte, 25)) + string(bytes.Repeat([]byte("-"), 50)) + string(make([]byte, 25)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &SparseFile{
				Segments: make([]*Segment, len(tt.segments)),
			}

			// Copy segments
			for i, seg := range tt.segments {
				sf.Segments[i] = &Segment{
					Start: seg.Start,
					End:   seg.End,
					Data:  make([]byte, len(seg.Data)),
				}
				copy(sf.Segments[i].Data, seg.Data)
			}

			result, ok := sf.GetContentRange(tt.offset, tt.length, '-')
			if !ok {
				t.Fatal("expected reconstruction to succeed")
			}

			if string(result) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(result))
			}

			if uint64(len(result)) != tt.length {
				t.Errorf("expected result length %d, got %d", tt.length, len(result))
			}
		})
	}
}

func TestContentReconstructionEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		segments   []*Segment
		offset     uint64
		length     uint64
		shouldFail bool
	}{
		{
			name:       "zero length request",
			segments:   []*Segment{{Start: 10, End: 20, Data: []byte("test")}},
			offset:     15,
			length:     0,
			shouldFail: true,
		},
		{
			name:       "request beyond all segments",
			segments:   []*Segment{{Start: 10, End: 20, Data: []byte("test")}},
			offset:     100,
			length:     10,
			shouldFail: true,
		},
		{
			name:       "request before all segments",
			segments:   []*Segment{{Start: 100, End: 110, Data: []byte("test")}},
			offset:     0,
			length:     50,
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &SparseFile{
				Segments: make([]*Segment, len(tt.segments)),
			}

			// Copy segments
			for i, seg := range tt.segments {
				sf.Segments[i] = &Segment{
					Start: seg.Start,
					End:   seg.End,
					Data:  make([]byte, len(seg.Data)),
				}
				copy(sf.Segments[i].Data, seg.Data)
			}

			result, ok := sf.GetContentRange(tt.offset, tt.length, 0x00)

			if tt.shouldFail {
				if ok {
					t.Errorf("expected reconstruction to fail, but got result: %v", result)
				}
			} else {
				if !ok {
					t.Error("expected reconstruction to succeed, but it failed")
				}
			}
		})
	}
}
