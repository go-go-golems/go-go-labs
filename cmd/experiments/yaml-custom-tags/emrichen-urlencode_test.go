package main

import (
	"testing"
)

func TestURLEncodeTag(t *testing.T) {
	tests := []testCase{
		{
			name:      "Basic URL Encoding",
			inputYAML: `!URLEncode "hello world & special=characters"`,
			expected:  "hello+world+%26+special%3Dcharacters",
		},
		{
			name:      "Empty String",
			inputYAML: `!URLEncode ""`,
			expected:  "",
		},
		{
			name:      "Numerical and Boolean Values",
			inputYAML: `!URLEncode 12345`,
			expected:  "\"12345\"",
		},
		{
			name:      "Complex String",
			inputYAML: `!URLEncode "email=example@example.com&param=value"`,
			expected:  "email%3Dexample%40example.com%26param%3Dvalue",
		},
		{
			name:      "URL with Query Parameters",
			inputYAML: `!URLEncode { url: "https://example.com/", query: { param1: "value1", param2: "value2" } }`,
			expected:  "https://example.com/?param1=value1&param2=value2",
		},
		{
			name:               "Invalid Input Type",
			inputYAML:          `!URLEncode [1, 2, 3]`,
			expectError:        true,
			expectErrorMessage: "!URLEncode requires a scalar or mapping node",
		},
		{
			name:      "Null Input",
			inputYAML: `!URLEncode null`,
			expected:  "",
		},
		{
			name:      "Very Long String",
			inputYAML: `!URLEncode "longstringlongstringlongstringlongstringlongstring..."`, // replace with an actual long string
			expected:  "longstringlongstringlongstringlongstringlongstring...",              // replace with the encoded version of the long string
		},
	}

	runTests(t, tests)
}
