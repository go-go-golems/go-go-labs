package main

import "testing"

func TestEmrichenExistsTag(t *testing.T) {
	tests := []testCase{
		{
			name:      "Exists with empty vars",
			inputYAML: "!Exists foo",
			expected:  "false",
			initVars:  map[string]interface{}{},
		},
		{
			name:      "Exists single match",
			inputYAML: "!Exists foo",
			expected:  "true",
			initVars: map[string]interface{}{
				"foo": "bar",
			},
		},
		{
			name:      "Exists multiple matches",
			inputYAML: "!Exists foo",
			expected:  "true",
			initVars: map[string]interface{}{
				"foo":  "bar",
				"foo2": "baz",
			},
		},
		{
			name:      "Exists no match",
			inputYAML: "!Exists unknown",
			expected:  "false",
			initVars: map[string]interface{}{
				"foo": "bar",
			},
		},
		{
			name:      "Exists nested structure",
			inputYAML: "!Exists foo.bar",
			expected:  "true",
			initVars: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz",
				},
			},
		},
		// TODO(manuel, 2024-01-24) This is a limitation of our jsonpath library
		//{
		//	name:      "Exists special characters",
		//	inputYAML: "!Exists 'foo-bar'",
		//	expected:  "true",
		//	initVars: map[string]interface{}{
		//		"foo-bar": "baz",
		//	},
		//},

		{
			name:      "Exists complex JSONPath",
			inputYAML: "!Exists foo[*].bar",
			expected:  "true",
			initVars: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{"bar": "baz1"},
					map[string]interface{}{"bar": "baz2"},
				},
			},
		},
		{
			name:      "Exists variable types",
			inputYAML: "!Exists foo",
			expected:  "true",
			initVars: map[string]interface{}{
				"foo": 12345,
			},
		},
		{
			name:      "Exists null value",
			inputYAML: "!Exists foo",
			expected:  "true",
			initVars: map[string]interface{}{
				"foo": nil,
			},
		},
		{
			name:        "Exists invalid JSONPath",
			inputYAML:   "!Exists [invalid]",
			expected:    "false",
			expectError: true,
		},
	}

	runTests(t, tests)
}
