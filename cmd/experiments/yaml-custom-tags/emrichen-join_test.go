package main

import (
	"testing"
)

func TestJoinTag(t *testing.T) {
	tests := []testCase{
		{
			name:      "Basic Join with Default Separator",
			inputYAML: `!Join [hello, world]`,
			expected:  "hello world",
		},
		{
			name:      "Join with Custom Separator",
			inputYAML: `!Join { items: [hello, world], separator: ", " }`,
			expected:  "hello, world",
		},
		{
			name:      "Empty List",
			inputYAML: `!Join []`,
			expected:  "\"\"",
		},
		{
			name:      "Single Element List",
			inputYAML: `!Join [hello]`,
			expected:  "hello",
		},
		{
			name:      "Non-String Elements",
			inputYAML: `!Join [1, 2, 3]`,
			expected:  "1 2 3",
		},
		{
			name:      "List with Null Elements",
			inputYAML: `!Join [hello, null, world]`,
			expected:  "hello world",
		},
		{
			name:      "Separator as Part of the Elements",
			inputYAML: `!Join { items: ["hello, world", "foo, bar"], separator: "; " }`,
			expected:  "hello, world; foo, bar",
		},
		{
			name:      "No Separator Provided",
			inputYAML: `!Join [hello, world]`,
			expected:  "hello world",
		},
		{
			name:      "Escaped Characters in Elements",
			inputYAML: `!Join ["hello\nworld", "foo\nbar"]`,
			expected:  "\"hello\\nworld foo\\nbar\"",
		},
		{
			name:      "Joining with a Complex Separator",
			inputYAML: `!Join { items: [hello, world], separator: "--**--" }`,
			expected:  "hello--**--world",
		},
		{
			name:      "Handling of Whitespace in Elements",
			inputYAML: `!Join ["  hello  ", "  world  "]`,
			expected:  "\"  hello     world  \"",
		},
	}

	runTests(t, tests)
}
