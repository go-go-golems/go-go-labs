package main

import "testing"

func TestEmrichenVarTag(t *testing.T) {
	tests := []testCase{
		{
			name:      "Var with string value",
			inputYAML: "!Var stringVar",
			expected:  "Hello, World!",
			initVars:  map[string]interface{}{"stringVar": "Hello, World!"},
		},
		{
			name:      "Var with integer value",
			inputYAML: "!Var intVar",
			expected:  "42",
			initVars:  map[string]interface{}{"intVar": 42},
		},
		{
			name:      "Var with boolean value",
			inputYAML: "!Var boolVar",
			expected:  "true",
			initVars:  map[string]interface{}{"boolVar": true},
		},
		{
			name:      "Var with list value",
			inputYAML: "!Var listVar",
			expected:  "[1, 2, 3]",
			initVars:  map[string]interface{}{"listVar": []int{1, 2, 3}},
		},
		{
			name:      "Var with map value",
			inputYAML: "!Var mapVar",
			expected:  "key1: val1\nkey2: val2",
			initVars:  map[string]interface{}{"mapVar": map[string]string{"key1": "val1", "key2": "val2"}},
		},

		{
			name:        "Var with non-existent variable",
			inputYAML:   "!Var nonExistentVar",
			expected:    "", // or the expected error message, if error handling is implemented
			initVars:    map[string]interface{}{},
			expectError: true,
		},
		{
			name:        "Var with undefined but referenced variable",
			inputYAML:   "!Var undefinedVar",
			expected:    "", // or the expected error message
			initVars:    map[string]interface{}{"someOtherVar": "value"},
			expectError: true,
		},

		{
			name:      "Var with nested list",
			inputYAML: "!Var nestedListVar",
			expected:  "[[1, 2], [3, 4]]",
			initVars:  map[string]interface{}{"nestedListVar": [][]int{{1, 2}, {3, 4}}},
		},
		{
			name:      "Var with nested map",
			inputYAML: "!Var nestedMapVar",
			expected:  "outerKey: {innerKey: value}",
			initVars:  map[string]interface{}{"nestedMapVar": map[string]map[string]string{"outerKey": {"innerKey": "value"}}},
		},

		{
			name:      "Var with explicitly set empty string",
			inputYAML: "!Var emptyString",
			expected:  "\"\"",
			initVars:  map[string]interface{}{"emptyString": ""},
		},
		{
			name:      "Var with explicitly set null value",
			inputYAML: "!Var nullVar",
			expected:  "null", // or the appropriate representation of null in your system
			initVars:  map[string]interface{}{"nullVar": nil},
		},
	}

	runTests(t, tests)
}
