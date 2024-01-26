package main

import (
	"testing"
)

func TestVoidTag(t *testing.T) {
	tests := []testCase{
		{
			name:      "Basic Use Case",
			inputYAML: "key1: value1\nkey2: !Void",
			expected:  "key1: value1\n",
		},
		{
			name:      "Void in List",
			inputYAML: "- item1\n- !Void\n- item2",
			expected:  "- item1\n- item2\n",
		},
		{
			name:      "Void in List (single line syntax skipped node)",
			inputYAML: "[item1, !Void, item2]",
			expected:  "- item1\n",
		},
		{
			name:      "Void in List (single line syntax)",
			inputYAML: "[item1, !Void foo, item2]",
			expected:  "[item1, item2]",
		},
		{
			name:      "Void in List (remove node)",
			inputYAML: "- item1\n- !Void foo\n- item2",
			expected:  "- item1\n- item2\n",
		},
		{
			name: "Multiple Voids",
			inputYAML: `key1: !Void
key2: value2
list: 
  - item1
  - !Void
  - item2`,
			expected: "key2: value2\nlist: [item1, item2]\n",
		},
		{
			name:      "Nested Void",
			inputYAML: "outer: { inner: { key: !Void } }",
			expected:  "outer: { inner: { } }\n",
		},
		{
			name:      "Void Entire Document",
			inputYAML: "---\nkey1: value1\n---\n!Void\n---\nkey2: value2",
			expected:  "key2: value2",
		},
		{
			name: "Void with Conditional",
			inputYAML: `key1: value1
key2: !If 
  test: true
  then: !Void
  else: value2`,
			expected: "key1: value1\n",
		},
		{
			name:      "Void Non-Existent Key",
			inputYAML: "key1: value1\nkey3: !Void",
			expected:  "key1: value1\n",
		},
		{
			name: "Void in a Loop",
			inputYAML: `!Loop 
  over: [item1, item2]
  template: !If 
    test: !Op 
      a: !Var item
      op: ==
      b: item2
    then: !Void
    else: !Var item`,
			expected: "[item1]\n",
		},
		{
			name: "Void as a Value",
			inputYAML: `key1: !Void
list: 
- item1
- !Void`,
			expected: "list: [item1]\n",
		},
		{
			name: "Combination of Void and Other Tags",
			inputYAML: `key1: !If 
  test: true
  then: value1
  else: !Void
    key2: !Var myVar`,
			expected: "key1: value1",
			initVars: map[string]interface{}{"myVar": "valueFromVar"},
		},
	}

	runTests(t, tests)
}
