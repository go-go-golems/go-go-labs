package main

import "testing"

func TestEmrichenOp(t *testing.T) {
	tests := []testCase{
		{
			name: "Op tag addition",
			inputYAML: `!Op
  a: 5
  op: "+"
  b: 3`,
			expected: "8",
		},
		{
			name: "Op tag subtraction",
			inputYAML: `!Op
  a: 10
  op: "-"
  b: 4`,
			expected: "6",
		},
		{
			name: "Op tag multiplication",
			inputYAML: `!Op
  a: 7
  op: "*"
  b: 6`,
			expected: "42",
		},
		{
			name: "Op tag division",
			inputYAML: `!Op
  a: 20.0
  op: "/"
  b: 4`,
			expected: "5.0",
		},
		{
			name: "Op tag modulo",
			inputYAML: `!Op
  a: 28
  op: "%"
  b: 5`,
			expected: "3",
		},
		{
			name: "Op tag greater than - true",
			inputYAML: `!Op
  a: 15
  op: ">"
  b: 10`,
			expected: "true",
		},
		{
			name: "Op tag greater than - false",
			inputYAML: `!Op
  a: 2
  op: ">"
  b: 3`,
			expected: "false",
		},
		{
			name: "Op tag less than - true",
			inputYAML: `!Op
  a: 3
  op: "<"
  b: 5`,
			expected: "true",
		},
		{
			name: "Op tag less than - false",
			inputYAML: `!Op
  a: 10
  op: "<"
  b: 5`,
			expected: "false",
		},
		{
			name: "Op tag equal to - true",
			inputYAML: `!Op
  a: 5
  op: "=="
  b: 5`,
			expected: "true",
		},
		{
			name: "Op tag equal to - false",
			inputYAML: `!Op
  a: 5
  op: "=="
  b: 3`,
			expected: "false",
		},
		{
			name: "Op tag not equal - true",
			inputYAML: `!Op
  a: 4
  op: "!="
  b: 5`,
			expected: "true",
		},
		{
			name: "Op tag not equal - false",
			inputYAML: `!Op
  a: 6
  op: "!="
  b: 6`,
			expected: "false",
		},
		{
			name: "Op tag greater than or equal - true",
			inputYAML: `!Op
  a: 5
  op: ">="
  b: 5`,
			expected: "true",
		},
		{
			name: "Op tag greater than or equal - false",
			inputYAML: `!Op
  a: 3
  op: ">="
  b: 4`,
			expected: "false",
		},
		{
			name: "Op tag less than or equal - true",
			inputYAML: `!Op
  a: 2
  op: "<="
  b: 3`,
			expected: "true",
		},
		{
			name: "Op tag less than or equal - false",
			inputYAML: `!Op
  a: 7
  op: "<="
  b: 6`,
			expected: "false",
		},
	}

	runTests(t, tests)
}
