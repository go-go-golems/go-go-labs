package main

import "testing"

func TestEmrichenConcatTag(t *testing.T) {
	tests := []testCase{
		{
			name:      "Concatenate Empty Lists",
			inputYAML: "!Concat [[], []]",
			expected:  "[]",
		},
		{
			name:      "Concatenate Single Element Lists",
			inputYAML: "!Concat [[1], ['a']]",
			expected:  "[1, 'a']",
		},
		{
			name:      "Concatenate Multiple Element Lists",
			inputYAML: "!Concat [[1, 2, 3], [4, 5]]",
			expected:  "[1, 2, 3, 4, 5]",
		},
		{
			name:      "Concatenate Lists with Different Types",
			inputYAML: "!Concat [[1, 'hello'], [true, 3.14]]",
			expected:  "[1, 'hello', true, 3.14]",
		},
		{
			name:      "Concatenate Nested Lists",
			inputYAML: "!Concat [[[1, 2]], [[3, 4]]]",
			expected:  "[[1, 2], [3, 4]]",
		},
		{
			name:      "Concatenate with Empty List",
			inputYAML: "!Concat [[1, 2, 3], []]",
			expected:  "[1, 2, 3]",
		},
		{
			name:      "Concatenate Lists with Null Values",
			inputYAML: "!Concat [[null], [1, null]]",
			expected:  "[null, 1, null]",
		},
		{
			name:      "Concatenate Lists with Variable References",
			inputYAML: "!Concat [!Var list1, !Var list2]",
			expected:  "[1, 2, 3, 4, 5]",
			initVars: map[string]interface{}{
				"list1": []interface{}{1, 2, 3},
				"list2": []interface{}{4, 5},
			},
		},
		{
			name:        "Concatenating Non-List Items",
			inputYAML:   "!Concat ['hello', [1, 2, 3]]",
			expected:    "Error",
			expectError: true,
		},
		{
			name:      "Concatenate Lists with Special Characters",
			inputYAML: "!Concat [['one', 'two@#'], ['three%$']]",
			expected:  "['one', 'two@#', 'three%$']",
		},
		{
			name:      "Concatenate Lists with Duplicate Elements",
			inputYAML: "!Concat [[1, 2, 2], [2, 3]]",
			expected:  "[1, 2, 2, 2, 3]",
		},
		{
			name:        "Undefined or Missing Lists",
			inputYAML:   "!Concat [!Var undefinedList, [1, 2, 3]]",
			expected:    "Error",
			expectError: true,
			initVars: map[string]interface{}{
				"undefinedList": nil,
			},
		},
	}

	runTests(t, tests)
}
