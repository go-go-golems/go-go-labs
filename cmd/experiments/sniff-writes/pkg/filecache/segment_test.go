package filecache

import (
	"testing"
)

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
				{Start: 0, End: 8, Data: []byte("hello123")},
			},
			new: &Segment{Start: 5, End: 15, Data: []byte("worldabcde")},
			expected: []*Segment{
				{Start: 0, End: 15, Data: []byte("helloworldabcde")},
			},
		},
		{
			name: "identical segments merge",
			existing: []*Segment{
				{Start: 10, End: 20, Data: []byte("same content")},
			},
			new: &Segment{Start: 10, End: 20, Data: []byte("same content")},
			expected: []*Segment{
				{Start: 10, End: 20, Data: []byte("same content")},
			},
		},
		{
			name: "fully contained segment absorbed",
			existing: []*Segment{
				{Start: 0, End: 20, Data: []byte("12345678901234567890")},
			},
			new: &Segment{Start: 5, End: 10, Data: []byte("ABCDE")},
			expected: []*Segment{
				{Start: 0, End: 20, Data: []byte("12345ABCDE1234567890")},
			},
		},
		{
			name: "multiple segments merge cascade",
			existing: []*Segment{
				{Start: 0, End: 5, Data: []byte("hello")},
				{Start: 10, End: 15, Data: []byte("world")},
			},
			new: &Segment{Start: 3, End: 12, Data: []byte("BRIDGE123")},
			expected: []*Segment{
				{Start: 0, End: 15, Data: []byte("helBRIDGE123rld")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &SparseFile{
				Segments: make([]*Segment, len(tt.existing)),
			}

			// Copy existing segments
			for i, seg := range tt.existing {
				sf.Segments[i] = &Segment{
					Start: seg.Start,
					End:   seg.End,
					Data:  make([]byte, len(seg.Data)),
				}
				copy(sf.Segments[i].Data, seg.Data)
			}

			// Insert new segment
			newSeg := &Segment{
				Start: tt.new.Start,
				End:   tt.new.End,
				Data:  make([]byte, len(tt.new.Data)),
			}
			copy(newSeg.Data, tt.new.Data)
			sf.insertSegment(newSeg, RealTimeProvider{})

			if len(sf.Segments) != len(tt.expected) {
				t.Fatalf("expected %d segments, got %d", len(tt.expected), len(sf.Segments))
			}

			for i, expected := range tt.expected {
				actual := sf.Segments[i]
				if actual.Start != expected.Start {
					t.Errorf("segment %d: expected start %d, got %d", i, expected.Start, actual.Start)
				}
				if actual.End != expected.End {
					t.Errorf("segment %d: expected end %d, got %d", i, expected.End, actual.End)
				}
				if string(actual.Data) != string(expected.Data) {
					t.Errorf("segment %d: expected data %q, got %q", i, expected.Data, actual.Data)
				}
			}
		})
	}
}

func TestSegmentMergingEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		existing []*Segment
		new      *Segment
		expected []*Segment
	}{
		{
			name:     "zero-length segment ignored",
			existing: []*Segment{{Start: 10, End: 20, Data: []byte("hello")}},
			new:      &Segment{Start: 15, End: 15, Data: []byte{}}, // Invalid segment
			expected: []*Segment{{Start: 10, End: 20, Data: []byte("hello")}},
		},
		{
			name:     "single-byte segment",
			existing: []*Segment{},
			new:      &Segment{Start: 100, End: 101, Data: []byte("X")},
			expected: []*Segment{{Start: 100, End: 101, Data: []byte("X")}},
		},
		{
			name: "merge at boundaries",
			existing: []*Segment{
				{Start: 0, End: 10, Data: []byte("0123456789")},
				{Start: 20, End: 30, Data: []byte("abcdefghij")},
			},
			new: &Segment{Start: 10, End: 20, Data: []byte("BRIDGE1234")},
			expected: []*Segment{
				{Start: 0, End: 30, Data: []byte("0123456789BRIDGE1234abcdefghij")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &SparseFile{
				Segments: make([]*Segment, len(tt.existing)),
			}

			// Copy existing segments
			for i, seg := range tt.existing {
				sf.Segments[i] = &Segment{
					Start: seg.Start,
					End:   seg.End,
					Data:  make([]byte, len(seg.Data)),
				}
				copy(sf.Segments[i].Data, seg.Data)
			}

			// Insert new segment
			newSeg := &Segment{
				Start: tt.new.Start,
				End:   tt.new.End,
				Data:  make([]byte, len(tt.new.Data)),
			}
			copy(newSeg.Data, tt.new.Data)
			sf.insertSegment(newSeg, RealTimeProvider{})

			if len(sf.Segments) != len(tt.expected) {
				t.Fatalf("expected %d segments, got %d", len(tt.expected), len(sf.Segments))
			}

			for i, expected := range tt.expected {
				actual := sf.Segments[i]
				if actual.Start != expected.Start {
					t.Errorf("segment %d: expected start %d, got %d", i, expected.Start, actual.Start)
				}
				if actual.End != expected.End {
					t.Errorf("segment %d: expected end %d, got %d", i, expected.End, actual.End)
				}
				if string(actual.Data) != string(expected.Data) {
					t.Errorf("segment %d: expected data %q, got %q", i, expected.Data, actual.Data)
				}
			}
		})
	}
}

func TestInsertionOrder(t *testing.T) {
	tests := []struct {
		name      string
		segments  []*Segment
		expected  []*Segment
		finalSize int
	}{
		{
			name: "reverse order insertion merges correctly",
			segments: []*Segment{
				{Start: 80, End: 90, Data: []byte("segment4")},
				{Start: 60, End: 70, Data: []byte("segment3")},
				{Start: 40, End: 50, Data: []byte("segment2")},
				{Start: 20, End: 30, Data: []byte("segment1")},
				{Start: 0, End: 100, Data: make([]byte, 100)}, // Covers all
			},
			expected: []*Segment{
				{Start: 0, End: 100, Data: make([]byte, 100)},
			},
			finalSize: 1,
		},
		{
			name: "chain merging",
			segments: []*Segment{
				{Start: 40, End: 60, Data: []byte("middle")},
				{Start: 20, End: 40, Data: []byte("left")},
				{Start: 60, End: 80, Data: []byte("right")},
			},
			expected: []*Segment{
				{Start: 20, End: 80, Data: []byte("leftmiddleright")},
			},
			finalSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &SparseFile{}

			// Insert segments in order
			for _, seg := range tt.segments {
				newSeg := &Segment{
					Start: seg.Start,
					End:   seg.End,
					Data:  make([]byte, len(seg.Data)),
				}
				copy(newSeg.Data, seg.Data)
				sf.insertSegment(newSeg, RealTimeProvider{})
			}

			if len(sf.Segments) != tt.finalSize {
				t.Fatalf("expected %d final segments, got %d", tt.finalSize, len(sf.Segments))
			}

			// Verify segments are sorted by start offset
			for i := 1; i < len(sf.Segments); i++ {
				if sf.Segments[i-1].Start >= sf.Segments[i].Start {
					t.Errorf("segments not sorted: segment[%d].Start=%d >= segment[%d].Start=%d",
						i-1, sf.Segments[i-1].Start, i, sf.Segments[i].Start)
				}
			}
		})
	}
}
