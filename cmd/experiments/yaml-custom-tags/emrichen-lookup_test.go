package main

import "testing"

func TestEmrichenLookupTag(t *testing.T) {
	tests := []testCase{
		{
			name:      "Lookup single node match",
			inputYAML: "!Lookup foo",
			expected:  "bar",
			initVars: map[string]interface{}{
				"foo": "bar",
			},
		},
		{
			name:      "Lookup multiple node match",
			inputYAML: "!Lookup items",
			expected:  "[first, second]",
			initVars: map[string]interface{}{
				"items": []interface{}{"first", "second"},
			},
		},
		{
			name:        "Lookup no match",
			inputYAML:   "!Lookup unknown",
			expected:    "",
			expectError: true,
		},
		{
			name:      "Lookup nested structures",
			inputYAML: "!Lookup foo.bar",
			expected:  "baz",
			initVars: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar": "baz",
				},
			},
		},
		// TODO(manuel, 2024-01-24) This is a limitation of the jsonpath library we use
		{
			name:      "Lookup special characters in keys",
			inputYAML: `!Lookup "foo-bar"`,
			expected:  "value",
			initVars: map[string]interface{}{
				"foo-bar": "value",
			},
		},
		{
			name:      "Lookup array index access",
			inputYAML: "!Lookup items[0]",
			expected:  "first",
			initVars: map[string]interface{}{
				"items": []interface{}{"first", "second"},
			},
		},
		{
			name:        "Lookup error handling",
			inputYAML:   "!Lookup [invalid]",
			expected:    "",
			expectError: true,
		},
		{
			name:      "Lookup node with list value",
			inputYAML: "!Lookup listOfItems",
			expected:  "[item1, item2, item3]",
			initVars: map[string]interface{}{
				"listOfItems": []interface{}{"item1", "item2", "item3"},
			},
		},
		{
			name:      "Lookup multiple node match with wildcard",
			inputYAML: "!Lookup items.*",
			expected:  "firstItem",
			initVars: map[string]interface{}{
				"items": []interface{}{
					"firstItem",
					"secondItem",
					"thirdItem",
				},
			},
		},
	}

	runTests(t, tests)
}
