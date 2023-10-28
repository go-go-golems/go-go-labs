package pkg

import (
	"testing"
)

func TestFindLocation(t *testing.T) {
	tests := []struct {
		name          string
		sourceLines   []string
		locationLines []string
		expectedIndex int
		expectedError error
	}{
		{
			name:          "WithEmptySourceLines",
			sourceLines:   []string{},
			locationLines: []string{"some code"},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name:          "WithEmptySourceLinesMultipleLocationLines",
			sourceLines:   []string{},
			locationLines: []string{"some code", "some other code"},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name:          "WithEmptyLocationLines",
			sourceLines:   []string{"some code"},
			locationLines: []string{},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name:          "WithEmptySourceAndLocationLines",
			sourceLines:   []string{},
			locationLines: []string{},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name:          "WithEmptyLocationLinesMultipleSourceLines",
			sourceLines:   []string{"some code", "some other code"},
			locationLines: []string{},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name:          "WithLocationNotFound",
			sourceLines:   []string{"some code"},
			locationLines: []string{"other code"},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name:          "WithLocationFoundAtBeginning",
			sourceLines:   []string{"location code", "some other code"},
			locationLines: []string{"location code"},
			expectedIndex: 0,
			expectedError: nil,
		},
		{
			name:          "WithLocationFoundAtMiddle",
			sourceLines:   []string{"some code", "location code", "some other code"},
			locationLines: []string{"location code"},
			expectedIndex: 1,
			expectedError: nil,
		},
		{
			name:          "WithLocationFoundAtEnd",
			sourceLines:   []string{"some code", "some other code", "location code"},
			locationLines: []string{"location code"},
			expectedIndex: 2,
			expectedError: nil,
		},
		{
			name:          "WithMultipleLocations",
			sourceLines:   []string{"location code", "some other code", "location code"},
			locationLines: []string{"location code"},
			expectedIndex: 0,
			expectedError: nil,
		},
		{
			name:          "WithSourceLinesHavingMultipleLines",
			sourceLines:   []string{"some code", "location code", "some other code"},
			locationLines: []string{"location code", "some other code"},
			expectedIndex: 1,
			expectedError: nil,
		},
		{
			name:          "WithLocationLinesHavingMultipleLines",
			sourceLines:   []string{"some code", "location code", "some other code"},
			locationLines: []string{"location code", "some other code"},
			expectedIndex: 1,
			expectedError: nil,
		},
		{
			name:          "WithLargeSourceAndLocationLines",
			sourceLines:   largeSourceLines(),
			locationLines: largeLocationLines(),
			expectedIndex: 50000,
			expectedError: nil,
		},
		{
			name:          "WithPartialMultipleLineMatch",
			sourceLines:   []string{"line one", "line two", "line three"},
			locationLines: []string{"line one", "non-matching line"},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name:          "WithMismatchedOrderOfLines",
			sourceLines:   []string{"line one", "line two", "line three"},
			locationLines: []string{"line two", "line one"},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name:          "WithEscapeSequences",
			sourceLines:   []string{"tab:\tend"},
			locationLines: []string{"tab:\tend"},
			expectedIndex: 0,
			expectedError: nil,
		},
		{
			name:          "WithCaseDifference",
			sourceLines:   []string{"Case Sensitive"},
			locationLines: []string{"case sensitive"},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name:          "WithEmptyStringLines",
			sourceLines:   []string{"", ""},
			locationLines: []string{""},
			expectedIndex: 0,
			expectedError: nil,
		},
		{
			name:          "WithSingleEmptyLineInLocation",
			sourceLines:   []string{"line one", "line two", "line three"},
			locationLines: []string{"line one", ""},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name: "WithMatchingEmptyLines",
			sourceLines: []string{
				"line one",
				"",
				"line three",
			},
			locationLines: []string{""},
			expectedIndex: 1,
			expectedError: nil,
		},
		{
			name: "WithMatchingEmptyLineAtEnd",
			sourceLines: []string{
				"line one",
				"line two",
				"",
			},
			locationLines: []string{""},
			expectedIndex: 2,
			expectedError: nil,
		},
		{
			name: "WithMatchingEmptyLineAtBeginning",
			sourceLines: []string{
				"",
				"line two",
				"line three",
			},
			locationLines: []string{""},
			expectedIndex: 0,
			expectedError: nil,
		},
		{
			name: "WithMultipleMatchingEmptyLinesAtBeginning",
			sourceLines: []string{
				"",
				"",
				"line three",
			},
			locationLines: []string{""},
			expectedIndex: 0,
			expectedError: nil,
		},
		{
			name: "WithMultipleMatchingEmptyLinesAtEnd",
			sourceLines: []string{
				"line one",
				"line two",
				"",
				"",
			},
			locationLines: []string{"", ""},
			expectedIndex: 2,
			expectedError: nil,
		},
		{
			name: "WithMultipleMatchingEmptyLinesInMiddle",
			sourceLines: []string{
				"line one",
				"",
				"",
				"line four",
			},
			locationLines: []string{"", ""},
			expectedIndex: 1,
			expectedError: nil,
		},
		{
			name:          "WithMultipleEmptyLinesInLocation",
			sourceLines:   []string{"line one", "line two", "line three"},
			locationLines: []string{"", "line two", ""},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name:          "WithEmptyLinesInBothSourceAndLocation",
			sourceLines:   []string{"line one", "", "line three"},
			locationLines: []string{"line one", ""},
			expectedIndex: 0, // if the function counts empty lines as valid lines
			expectedError: nil,
		},
		{
			name:          "WithLocationLinesCompletelyEmpty",
			sourceLines:   []string{"line one", "line two", "line three"},
			locationLines: []string{"", "", ""},
			expectedIndex: -1, // considering that completely empty location lines could be invalid
			expectedError: &ErrCodeBlock{},
		},
		{
			name:          "WithSourceInterspersedEmptyLines",
			sourceLines:   []string{"line one", "", "line three", "", "line five"},
			locationLines: []string{"", "line three", ""},
			expectedIndex: 1, // if the function considers empty lines as valid and part of the sequence
			expectedError: nil,
		},
		{
			name:          "WithSubstringInLocationNotFullLine",
			sourceLines:   []string{"This is a line of code"},
			locationLines: []string{"a line"},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name:          "WithLeadingOrTrailingSpaces",
			sourceLines:   []string{"    indented line", "line with space    "},
			locationLines: []string{"indented line", "line with space"},
			expectedIndex: -1, // if spaces are significant in matches
			expectedError: &ErrCodeBlock{},
		},
		{
			name:          "WithNonStandardLineBreaks",
			sourceLines:   []string{"line with \r", "another line"},
			locationLines: []string{"line with \r"},
			expectedIndex: 0, // if non-standard line breaks are handled correctly
			expectedError: nil,
		},
		{
			name: "WithSpaceAtBeginningOfLine",
			sourceLines: []string{
				"line one",
				" line two",
				"line three",
			},
			locationLines: []string{"line two"},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name: "WithPartialLineMatchAtBeginning",
			sourceLines: []string{
				"line one",
				"line two",
			},
			locationLines: []string{"line"},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name: "WithPartialLineMatchAtEnd",
			sourceLines: []string{
				"line one",
				"line two",
			},
			locationLines: []string{"two"},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
		{
			name: "WithPartialLineMatchInMiddle",
			sourceLines: []string{
				"line one",
				"line two",
			},
			locationLines: []string{"ne t"},
			expectedIndex: -1,
			expectedError: &ErrCodeBlock{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, err := FindLocation(tt.sourceLines, tt.locationLines)

			if index != tt.expectedIndex {
				t.Errorf("expected index %d, got %d", tt.expectedIndex, index)
			}

			if err != nil && tt.expectedError != nil {
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %s, got %s", tt.expectedError, err)
				}
			} else if err != tt.expectedError {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}

func largeSourceLines() []string {
	lines := make([]string, 100000)
	for i := range lines {
		if i == 50000 {
			lines[i] = "location code"
		} else {
			lines[i] = "some code"
		}
	}
	return lines
}

func largeLocationLines() []string {
	return []string{"location code"}
}
