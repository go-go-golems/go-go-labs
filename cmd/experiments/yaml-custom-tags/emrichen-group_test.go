package main

import (
	"testing"
)

func TestGroupTag(t *testing.T) {
	tests := []testCase{
		{
			name: "Basic Grouping",
			inputYAML: `!Group
  over:
    - name: manifold
      score: 7.8
    - name: John
      score: 9.9
    - name: John
      score: 9.8
  as: item
  by: !Lookup item.name
  template: !Lookup item.score`,
			expected: `{"manifold": [7.8], "John": [9.9, 9.8]}`,
		},
		{
			name: "Basic Grouping",
			inputYAML: `!Group
  over:
    - name: manifold
      score: 7.8
    - name: John
      score: 9.9
    - name: John
      score: 9.8
  as: item
  by: !Lookup item.name
  template: !Lookup item.score`,
			expected: `{"manifold": [7.8], "John": [9.9, 9.8]}`,
		},
		{
			name: "Grouping Empty List",
			inputYAML: `!Group
  over: []
  as: item
  by: !Lookup item.name
  template: !Lookup item.score`,
			expected: `{}`,
		},
		{
			name: "Grouping With Non-String Key",
			inputYAML: `!Group
  over:
    - category: 1
      value: "A"
    - category: 2
      value: "B"
    - category: 1
      value: "C"
  as: item
  by: !Lookup item.category
  template: !Lookup item.value`,
			expected: `{"1": ["A", "C"], "2": ["B"]}`,
		},

		{
			name: "Grouping With Duplicate Keys",
			inputYAML: `!Group
  over:
    - name: Alpha
      score: 5.5
    - name: Beta
      score: 6.6
    - name: Alpha
      score: 7.7
  as: item
  by: !Lookup item.name
  template: !Lookup item.score`,
			expected: `{"Alpha": [5.5, 7.7], "Beta": [6.6]}`,
		},
		{
			name: "Nested Objects Grouping",
			inputYAML: `!Group
  over:
    - category: { id: 1, name: "Category A" }
      value: "Item 1"
    - category: { id: 2, name: "Category B" }
      value: "Item 2"
    - category: { id: 1, name: "Category A" }
      value: "Item 3"
  as: item
  by: !Lookup item.category.id
  template: !Lookup item.value`,
			expected: `{"1": ["Item 1", "Item 3"], "2": ["Item 2"]}`,
		},

		{
			name: "Grouping Without 'template' Argument",
			inputYAML: `!Group
  over:
    - name: Alpha
      value: 1
    - name: Beta
      value: 2
    - name: Alpha
      value: 3
  as: item
  by: !Lookup item.name`,
			expected: `{"Alpha": [{"name": "Alpha", "value": 1}, {"name": "Alpha", "value": 3}], "Beta": [{"name": "Beta", "value": 2}]}`,
		},

		{
			name: "Invalid Key Path",
			inputYAML: `!Group
  over:
    - name: Alpha
      value: 1
    - name: Beta
      value: 2
  as: item
  by: !Lookup item.nonexistent
  template: !Lookup item.value`,
			expectError:        true,
			expectErrorMessage: "nonexistent is not found",
		},
	}

	// runTests function should be implemented to execute each test case
	runTests(t, tests)
}
