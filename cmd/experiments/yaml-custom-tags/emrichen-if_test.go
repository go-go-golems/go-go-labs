package main

import (
	"testing"
)

func TestEmrichenIfAndFilterTags(t *testing.T) {
	tests := []testCase{
		{
			name: "If tag true condition",
			inputYAML: `!If
  test:
    a: 10
    op: ">"
    b: 5
  then: "True Condition"
  else: "False Condition"`,
			expected: "\"True Condition\"",
		},
		{
			name: "If tag false condition",
			inputYAML: `!If
  test:
    a: 3
    op: ">"
    b: 5
  then: "True Condition"
  else: "False Condition"`,
			expected: "\"False Condition\"",
		},
		{
			name: "Filter tag filtering numbers greater than 2",
			inputYAML: `!Filter
  test:
    a: !Var item
    op: ">"
    b: 2
  over: [1, 2, 3, 4]`,
			expected: "[3, 4]",
		},
	}

	runTests(t, tests)
}
