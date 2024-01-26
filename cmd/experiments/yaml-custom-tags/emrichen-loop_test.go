package main

import "testing"

func TestLoopTag(t *testing.T) {
	tests := []testCase{
		{
			name: "Basic Loop Over List",
			inputYAML: `!Loop
  over: [1, 2, 3]
  as: item
  template: !Var item`,
			expected: `[1, 2, 3]`,
		},
		{
			name: "Loop Over Empty List",
			inputYAML: `!Loop
  over: []
  as: item
  template: !Var item`,
			expected: `[]`,
		},
		{
			name: "Loop Over List with Various Data Types",
			inputYAML: `!Loop
  over: [1, "two", true]
  as: item
  template: !Var item`,
			expected: `[1, "two", true]`,
		},
		{
			name: "Loop Over Dictionary",
			inputYAML: `!Loop
  over: {a: 1, b: 2}
  as: item
  template: !Var item`,
			expected: `{a: 1, b: 2}`,
		},
		{
			name: "Loop With Custom Variable Name",
			inputYAML: `!Loop
  over: [10, 20, 30]
  as: num
  template: !Var num`,
			expected: `[10, 20, 30]`,
		},
		{
			name: "Nested Loop With Custom Variable Names",
			inputYAML: `!Loop
  over: [{values: [1, 2]}, {values: [3, 4]}]
  as: group
  template: !Loop
    over: !Lookup group.values
    as: val
    template: !Var val`,
			expected: `[[1, 2], [3, 4]]`,
		},
		{
			name: "Loop With `index_as` and Custom Variable Name",
			inputYAML: `!Loop
  over: ["apple", "banana", "cherry"]
  as: fruit
  index_as: idx
  template: !Format "{idx}: !Var fruit"`,
			expected: `["0: apple", "1: banana", "2: cherry"]`,
		},
		{
			name: "Loop With `previous_as` and Custom Variable Name",
			inputYAML: `!Loop
  over: [4, 5, 6]
  as: current
  previous_as: prev
  template: !Format "prev: !Var prev, current: !Var current"`,
			expected: `["prev: null, current: 4", "prev: 4, current: 5", "prev: 5, current: 6"]`,
		},
		{
			name: "Loop With `index_start` and Custom Variable Name",
			inputYAML: `!Loop
  over: ["first", "second", "third"]
  as: position
  index_as: num
  index_start: 1
  template: !Format "Position !Var num: !Var position"`,
			expected: `["Position 1: first", "Position 2: second", "Position 3: third"]`,
		},
		{
			name: "Nested Loops",
			inputYAML: `!Loop
  over: [{name: "group1", items: [1, 2]}, {name: "group2", items: [3, 4]}]
  as: group
  template:
    name: !Lookup group.name
    items: !Loop
      over: !Lookup group.items
      as: item
      template: !Var item`,
			expected: `[
  {
    "name": "group1",
    "items": [1, 2]
  },
  {
    "name": "group2",
    "items": [3, 4]
  }
]`,
		},
		{
			name: "Loop with 'index_as' Option",
			inputYAML: `!Loop
  over: ["a", "b", "c"]
  as: letter
  index_as: idx
  template: !Format "{idx}: !Var letter"`,
			expected: `["0: a", "1: b", "2: c"]`,
		},
		{
			name: "Loop with 'previous_as' Option",
			inputYAML: `!Loop
  over: [10, 20, 30]
  as: current
  previous_as: prev
  template: !Format "prev: !Var prev, current: !Var current"`,
			expected: `["prev: null, current: 10", "prev: 10, current: 20", "prev: 20, current: 30"]`,
		},
		{
			name: "Loop with 'index_start' Option",
			inputYAML: `!Loop
  over: ["first", "second", "third"]
  as: item
  index_as: num
  index_start: 1
  template: !Format "Item !Var num: !Var item"`,
			expected: `["Item 1: first", "Item 2: second", "Item 3: third"]`,
		},
		{
			name: "Loop with 'as_documents' Option",
			inputYAML: `!Loop
  over: ["one", "two", "three"]
  as: item
  as_documents: true
  template: !Var item`,
			expected: `---
"one"
---
"two"
---
"three"
---`,
		},
		{
			name: "Loop with Conditional Logic Inside",
			inputYAML: `!Loop
  over: [1, 2, 3, 4, 5]
  as: num
  template: !If
    test: !Op {a: !Var num, op: "<", b: 4}
    then: !Format "Number !Var num is less than 4"
    else: !Format "Number !Var num is 4 or greater"`,
			expected: `[
  "Number 1 is less than 4",
  "Number 2 is less than 4",
  "Number 3 is less than 4",
  "Number 4 is 4 or greater",
  "Number 5 is 4 or greater"
]`,
		},
		{
			name: "Loop with Variable Substitution Inside",
			inputYAML: `!Loop
  over: ["x", "y", "z"]
  as: letter
  template: !Format "The letter is: !Var letter"`,
			expected: `[
  "The letter is: x",
  "The letter is: y",
  "The letter is: z"
]`,
		},
		{
			name: "Loop with Error Handling - Missing Variable",
			inputYAML: `!Loop
  over: [1, 2, 3]
  as: num
  template: !Var missingVariable`,
			expectError:        true,
			expectErrorMessage: "Variable 'missingVariable' not found",
		},
		{
			name: "Loop with Error Handling - Invalid Operation",
			inputYAML: `!Loop
  over: [1, 2, "three"]
  as: num
  template: !Op {a: !Var num, op: "+", b: 10}`,
			expectError:        true,
			expectErrorMessage: "Cannot perform operation '+' on a string and an integer",
		},
	}

	// runTests function should be implemented to execute each test case
	runTests(t, tests)
}
