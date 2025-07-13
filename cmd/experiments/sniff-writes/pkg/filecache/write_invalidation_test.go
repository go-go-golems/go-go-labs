package filecache

import (
	"testing"
)

func TestWriteInvalidation(t *testing.T) {
	tests := []struct {
		name      string
		initial   []*Segment
		writeOff  uint64
		writeData []byte
		expected  []*Segment
	}{
		{
			name: "write completely replaces segment",
			initial: []*Segment{
				{Start: 10, End: 20, Data: []byte("old_data")},
			},
			writeOff:  10,
			writeData: []byte("new_data"),
			expected: []*Segment{
				{Start: 10, End: 18, Data: []byte("new_data")},
			},
		},
		{
			name: "write splits segment",
			initial: []*Segment{
				{Start: 10, End: 30, Data: []byte("aaaaaaaaaaaaaaaaaaaa")},
			},
			writeOff:  15,
			writeData: []byte("NEW"),
			expected: []*Segment{
				{Start: 10, End: 15, Data: []byte("aaaaa")},
				{Start: 15, End: 18, Data: []byte("NEW")},
				{Start: 18, End: 30, Data: []byte("aaaaaaaaaaaa")},
			},
		},
		{
			name: "write overlaps segment start",
			initial: []*Segment{
				{Start: 20, End: 40, Data: []byte("xxxxxxxxxxxxxxxxxxxx")},
			},
			writeOff:  15,
			writeData: []byte("PREFIX"),
			expected: []*Segment{
				{Start: 15, End: 21, Data: []byte("PREFIX")},
				{Start: 21, End: 40, Data: []byte("xxxxxxxxxxxxxxxxxxx")},
			},
		},
		{
			name: "write overlaps segment end",
			initial: []*Segment{
				{Start: 10, End: 30, Data: []byte("yyyyyyyyyyyyyyyyyyyy")},
			},
			writeOff:  25,
			writeData: []byte("SUFFIX"),
			expected: []*Segment{
				{Start: 10, End: 25, Data: []byte("yyyyyyyyyyyyyyy")},
				{Start: 25, End: 31, Data: []byte("SUFFIX")},
			},
		},
		{
			name: "write spans multiple segments",
			initial: []*Segment{
				{Start: 10, End: 20, Data: []byte("first")},
				{Start: 30, End: 40, Data: []byte("second")},
				{Start: 50, End: 60, Data: []byte("third")},
			},
			writeOff:  15,
			writeData: []byte("REPLACEMENT"),
			expected: []*Segment{
				{Start: 10, End: 15, Data: []byte("first")},
				{Start: 15, End: 26, Data: []byte("REPLACEMENT")},
				{Start: 30, End: 40, Data: []byte("second")},
				{Start: 50, End: 60, Data: []byte("third")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("=== Test: %s ===", tt.name)
			t.Logf("Write: offset=%d, data=%q (%d bytes)", tt.writeOff, tt.writeData, len(tt.writeData))
			t.Logf("Write range: [%d,%d)", tt.writeOff, tt.writeOff+uint64(len(tt.writeData)))

			t.Log("Initial segments:")
			for i, seg := range tt.initial {
				t.Logf("  [%d] [%d,%d) %q (%d bytes)", i, seg.Start, seg.End, seg.Data, len(seg.Data))
			}

			sf := &SparseFile{
				Segments: make([]*Segment, len(tt.initial)),
			}

			// Copy initial segments
			for i, seg := range tt.initial {
				sf.Segments[i] = &Segment{
					Start: seg.Start,
					End:   seg.End,
					Data:  make([]byte, len(seg.Data)),
				}
				copy(sf.Segments[i].Data, seg.Data)
			}

			// Perform write
			sf.UpdateWithWrite(tt.writeOff, tt.writeData, RealTimeProvider{})

			t.Log("Actual segments after write:")
			for i, seg := range sf.Segments {
				t.Logf("  [%d] [%d,%d) %q (%d bytes)", i, seg.Start, seg.End, seg.Data, len(seg.Data))
			}

			t.Log("Expected segments:")
			for i, seg := range tt.expected {
				t.Logf("  [%d] [%d,%d) %q (%d bytes)", i, seg.Start, seg.End, seg.Data, len(seg.Data))
			}

			// Verify results
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

func TestWriteInvalidationComplexCases(t *testing.T) {
	tests := []struct {
		name      string
		initial   []*Segment
		writeOff  uint64
		writeData []byte
		expected  []*Segment
	}{
		{
			name: "write covers multiple complete segments",
			initial: []*Segment{
				{Start: 20, End: 30, Data: []byte("segment1")},
				{Start: 40, End: 50, Data: []byte("segment2")},
				{Start: 60, End: 70, Data: []byte("segment3")},
			},
			writeOff:  10,
			writeData: make([]byte, 70), // Write [10,80) covers all segments
			expected: []*Segment{
				{Start: 10, End: 80, Data: make([]byte, 70)},
			},
		},
		{
			name: "write creates holes in existing segments",
			initial: []*Segment{
				{Start: 20, End: 60, Data: make([]byte, 40)},
			},
			writeOff:  30,
			writeData: []byte("HOLE"),
			expected: []*Segment{
				{Start: 20, End: 30, Data: make([]byte, 10)},
				{Start: 30, End: 34, Data: []byte("HOLE")},
				{Start: 34, End: 60, Data: make([]byte, 26)},
			},
		},
		{
			name: "write at exact segment boundaries",
			initial: []*Segment{
				{Start: 10, End: 20, Data: []byte("boundary1")},
				{Start: 20, End: 30, Data: []byte("boundary2")},
			},
			writeOff:  20,
			writeData: []byte("EXACT"),
			expected: []*Segment{
				{Start: 10, End: 20, Data: []byte("boundary1")},
				{Start: 20, End: 25, Data: []byte("EXACT")},
				{Start: 25, End: 30, Data: []byte("ndary2")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := &SparseFile{
				Segments: make([]*Segment, len(tt.initial)),
			}

			// Copy initial segments
			for i, seg := range tt.initial {
				sf.Segments[i] = &Segment{
					Start: seg.Start,
					End:   seg.End,
					Data:  make([]byte, len(seg.Data)),
				}
				copy(sf.Segments[i].Data, seg.Data)
			}

			// Perform write
			sf.UpdateWithWrite(tt.writeOff, tt.writeData, RealTimeProvider{})

			// Verify results
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
				if len(actual.Data) != len(expected.Data) {
					t.Errorf("segment %d: expected data length %d, got %d", i, len(expected.Data), len(actual.Data))
				}
			}
		})
	}
}
