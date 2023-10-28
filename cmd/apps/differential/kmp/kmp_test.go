package kmp

import "testing"

func TestKMPSearch(t *testing.T) {
	tests := []struct {
		name     string
		text     []string
		pattern  []string
		expected int
	}{
		{
			name:     "EmptyText",
			text:     []string{},
			pattern:  []string{"apple", "banana"},
			expected: -1,
		},
		{
			name:     "EmptyPattern",
			text:     []string{"apple", "banana", "cherry", "date"},
			pattern:  []string{},
			expected: -1,
		},
		{
			name:     "EmptyTextAndPattern",
			text:     []string{},
			pattern:  []string{},
			expected: -1,
		},

		{
			name:     "SingleLineTextAndPatternMatch",
			text:     []string{"apple"},
			pattern:  []string{"apple"},
			expected: 0,
		},
		{
			name:     "SingleLineTextAndPatternNoMatch",
			text:     []string{"apple"},
			pattern:  []string{"banana"},
			expected: -1,
		},
		{
			name:     "PatternLongerThanText",
			text:     []string{"apple"},
			pattern:  []string{"apple", "banana"},
			expected: -1,
		},
		{
			name:     "AllLinesIdentical",
			text:     []string{"apple", "apple", "apple", "apple"},
			pattern:  []string{"apple", "apple"},
			expected: 0,
		},
		{
			name:     "NonAsciiCharacters",
			text:     []string{"こんにちは", "世界"},
			pattern:  []string{"こんにちは"},
			expected: 0,
		},

		{
			name:     "MatchAtBeginning",
			text:     []string{"apple", "banana", "cherry", "date"},
			pattern:  []string{"apple", "banana"},
			expected: 0,
		},
		{
			name:     "MatchAtEnd",
			text:     []string{"apple", "banana", "cherry", "date"},
			pattern:  []string{"cherry", "date"},
			expected: 2,
		},
		{
			name:     "MatchInMiddle",
			text:     []string{"apple", "banana", "cherry", "date"},
			pattern:  []string{"banana", "cherry"},
			expected: 1,
		},
		{
			name:     "MultipleMatches",
			text:     []string{"apple", "banana", "apple", "banana"},
			pattern:  []string{"apple", "banana"},
			expected: 0,
		},
		{
			name:     "NoMatch",
			text:     []string{"apple", "banana", "cherry", "date"},
			pattern:  []string{"kiwi", "lemon"},
			expected: -1,
		},
		{
			name:     "TextContainsEmptyLines",
			text:     []string{"apple", "", "banana"},
			pattern:  []string{""},
			expected: 1, // The pattern (an empty line) is found at index 1
		},
		{
			name:     "PatternIsEmptyLine",
			text:     []string{"apple", "banana", "cherry"},
			pattern:  []string{""},
			expected: -1,
		},
		{
			name:     "EmptyLineTextAndPattern",
			text:     []string{""},
			pattern:  []string{""},
			expected: 0, // Empty pattern is found at the start of the empty text
		},
		{
			name:     "NonEmptyTextEmptyPattern",
			text:     []string{"apple", "banana", "cherry"},
			pattern:  []string{""},
			expected: -1,
		},
		{
			name:     "ConsecutiveEmptyLinesInText",
			text:     []string{"", "", "apple", "", "banana"},
			pattern:  []string{""},
			expected: 0,
		},
		{
			name: "LongSequenceWithJumps",
			text: []string{
				"apple",
				"banana",
				"apple",
				"banana",
				"cherry",
				"date",
				"apple",
				"banana",
				"cherry",
				"disk",
			},
			pattern:  []string{"apple", "banana", "cherry", "disk"},
			expected: 6,
		},
		{
			name: "LongSequenceWithJumps",
			text: []string{
				"apple",
				"banana",
				"apple",
				"banana",
				"cherry",
				"date",
				"apple",
				"banana",
				"cherry",
			},
			pattern:  []string{"apple", "banana", "cherry", "disk"},
			expected: -1,
		},
		{
			name: "RepeatingLongSequenceWithJumps",
			text: []string{
				"apple",
				"banana",
				"apple",
				"banana",
				"cherry",
				"date",
				"apple",
				"banana",
				"cherry",
			},
			pattern:  []string{"apple", "banana", "apple", "banana", "cherry"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := KMPSearch(tt.text, tt.pattern)
			if result != tt.expected {
				t.Errorf("KMPSearch(%v, %v) = %v; want %v", tt.text, tt.pattern, result, tt.expected)
			}
		})
	}
}
