package main

import (
	"errors"
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
			expectedError: errors.New("specified code block not found in the source"),
		},
		{
			name:          "WithEmptySourceLinesMultipleLocationLines",
			sourceLines:   []string{},
			locationLines: []string{"some code", "some other code"},
			expectedIndex: -1,
			expectedError: errors.New("specified code block not found in the source"),
		},
		{
			name:          "WithEmptyLocationLines",
			sourceLines:   []string{"some code"},
			locationLines: []string{},
			expectedIndex: -1,
			expectedError: errors.New("specified code block not found in the source"),
		},
		{
			name:          "WithEmptySourceAndLocationLines",
			sourceLines:   []string{},
			locationLines: []string{},
			expectedIndex: -1,
			expectedError: errors.New("specified code block not found in the source"),
		},
		{
			name:          "WithEmptyLocationLinesMultipleSourceLines",
			sourceLines:   []string{"some code", "some other code"},
			locationLines: []string{},
			expectedIndex: -1,
			expectedError: errors.New("specified code block not found in the source"),
		},
		{
			name:          "WithLocationNotFound",
			sourceLines:   []string{"some code"},
			locationLines: []string{"other code"},
			expectedIndex: -1,
			expectedError: errors.New("specified code block not found in the source"),
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, err := findLocation(tt.sourceLines, tt.locationLines)

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
