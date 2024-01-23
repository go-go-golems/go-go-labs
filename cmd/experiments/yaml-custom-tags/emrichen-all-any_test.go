package main

import (
	"testing"
)

func TestEmrichenTags(t *testing.T) {
	tests := []testCase{
		// https://chat.openai.com/c/ad6cb760-3417-4a9e-b76b-8bc590e366da
		{
			// Test with an empty list, expecting true since no false values exist
			name:      "All with empty list",
			inputYAML: "!All []",
			expected:  "true",
		},
		{
			// Test with all true values of different types (boolean, non-zero number, non-empty string)
			// Expecting true since all values are truthy
			name:      "All with various true values",
			inputYAML: "!All [true, 1, 'yes']",
			expected:  "true",
		},
		{
			// Test with mixed values, including a false value at the end
			// Expecting false since there's a false value in the list
			name:      "All with mixed values and false at the end",
			inputYAML: "!All [true, 1, 'yes', false]",
			expected:  "false",
		},
		{
			// Test with mixed values, including a false value at the beginning
			// Expecting false since there's a false value in the list
			name:      "All with mixed values and false at the beginning",
			inputYAML: "!All [false, true, 1, 'yes']",
			expected:  "false",
		},
		{
			name:      "All with all true values",
			inputYAML: "!All [true, true]",
			expected:  "true",
		},
		{
			name:      "All with one false value",
			inputYAML: "!All [true, false]",
			expected:  "false",
		},
		{
			name:      "Any with one true value",
			inputYAML: "!Any [false, true]",
			expected:  "true",
		},
		{
			name:      "Any with all false values",
			inputYAML: "!Any [false, false]",
			expected:  "false",
		},
		{
			// Test with all false values (boolean false, zero, empty string, null)
			// Expecting false since all values are falsy
			name:      "All with all false values",
			inputYAML: "!All [false, 0, '', null]",
			expected:  "false",
		},
		{
			// Test with a nested list
			// Expecting true or false depending on interpretation of nested list truthiness
			name:      "All with a nested list",
			inputYAML: "!All [[true, true], true]",
			expected:  "true", // or "false" if nested list is considered falsy
		},
		{
			name:      "All with nested list containing a falsy value",
			inputYAML: "!All [[false], true]",
			expected:  "true",
		},
		{
			name:      "All with empty nested list (falsy)",
			inputYAML: "!All [[], true]",
			expected:  "false",
		},
		{
			name: "All with a nested mapping",
			inputYAML: `!All
- a: true
  b: true
- true`,
			expected: "true",
		},
		{
			name: "All with a nested mapping containing a falsy value",
			inputYAML: `!All
- a: false
  b: true
- true`,
			expected: "true",
		},
		{
			name: "All with an empty nested mapping (falsy)",
			inputYAML: `!All
- {}
- true`,
			expected: "false",
		},
		{
			// Test with non-boolean values (string, number)
			// Expecting true since all are truthy values
			name:      "All with non-boolean values",
			inputYAML: "!All ['text', 123]",
			expected:  "true",
		},
		{
			// Test with non-boolean values including a falsy value (string, 0)
			// Expecting false since 0 is a falsy value
			name:      "All with mixed non-boolean values",
			inputYAML: "!All ['text', 0]",
			expected:  "false",
		},
		{
			// Test with a large list of true values
			// Expecting true since all values are truthy
			name:      "All with a large list of true values",
			inputYAML: "!All [true, true, true, true, true, true, true, true, true, true]",
			expected:  "true",
		},
	}

	runTests(t, tests)
}

func TestEmrichenAnyTag(t *testing.T) {
	tests := []testCase{
		{
			name:      "Any with empty list",
			inputYAML: "!Any []",
			expected:  "false",
		},
		{
			name:      "Any with all false values",
			inputYAML: "!Any [false, 0, '', null]",
			expected:  "false",
		},
		{
			name:      "Any with mixed values and true at the end",
			inputYAML: "!Any [false, 0, '', true]",
			expected:  "true",
		},
		{
			name:      "Any with mixed values and true at the beginning",
			inputYAML: "!Any [true, 0, '', false]",
			expected:  "true",
		},
		{
			name:      "Any with all true values",
			inputYAML: "!Any [true, 1, 'yes']",
			expected:  "true",
		},
		{
			name:      "Any with a nested list",
			inputYAML: "!Any [[false], true]",
			expected:  "true",
		},
		{
			name:      "Any with non-boolean values",
			inputYAML: "!Any ['text', 123]",
			expected:  "true",
		},
		{
			name:      "Any with single true value",
			inputYAML: "!Any [true]",
			expected:  "true",
		},
		{
			name:      "Any with single false value",
			inputYAML: "!Any [false]",
			expected:  "false",
		},
		{
			name:      "Any with boolean and non-boolean true values",
			inputYAML: "!Any [true, 'text']",
			expected:  "true",
		},
		{
			name:      "Any with strings containing special characters",
			inputYAML: "!Any ['@#$%', '']",
			expected:  "true",
		},
		{
			name:      "Any with very large list",
			inputYAML: "!Any [false, false, false, ... (1000 more times), true]",
			expected:  "true",
		},
		{
			name:      "Any with list including null values",
			inputYAML: "!Any [null, false, 0]",
			expected:  "false",
		},
	}

	runTests(t, tests)
}
