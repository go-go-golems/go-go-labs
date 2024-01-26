package main

import "testing"

func TestEmrichenFilterTag(t *testing.T) {
	tests := []testCase{
		{
			name: "Filter with simple values",
			inputYAML: `!Filter
  over:
  - valid
  - hello
  - 0
  - SSEJ
  - false
  - null`,
			expected: "[valid, hello, SSEJ]", // Adjust this expected output based on your logic
		},
		{
			name: "Filter using !Not and !Var on a dictionary",
			inputYAML: `!Filter
  as: i
  test: !Not,Var i
  over:
    'yes': true
    no: 0
    nope: false
    oui: 1`,
			expected: "{'no': 0, nope: false}", // Adjust this expected output based on your logic
		},
		{
			name: "Filter with !Op on a list of integers",
			inputYAML: `!Filter
  test: !Op
    a: !Var item
    op: gt
    b: 4
  over: [1, 7, 2, 5]`,
			expected: "[7, 5]", // This should filter out elements greater than 4
		},
		{
			name: "Filter with dynamic list from !Loop",
			inputYAML: `!Filter
  test: !Op
    a: !Var item
    op: lt
    b: 5
  over: !Loop
    over: [1, 2, 3, 4, 5, 6, 7]
    template: !Var item`,
			expected: "[1, 2, 3, 4]", // Only items less than 5
		},
		{
			name: "Filter with list expansion using !Concat",
			inputYAML: `!Filter
  test: !Op
    a: !Var item
    op: ne
    b: 3
  over: !Concat [[1, 2], [3, 4]]`,
			expected: "[1, 2, 4]", // Filter out the number 3
		},
	}

	runTests(t, tests)
}
