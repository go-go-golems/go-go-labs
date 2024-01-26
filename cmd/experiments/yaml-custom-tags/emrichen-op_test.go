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
		{
			name: "Op tag with variable substitution in 'a'",
			inputYAML: `!Op
  a: !Var varA
  op: "+"
  b: 3`,
			expected: "8",
			initVars: map[string]interface{}{
				"varA": 5,
			},
		},
		{
			name: "Op tag with variable substitution in 'b'",
			inputYAML: `!Op
  a: 10
  op: "-"
  b: !Var varB`,
			expected: "6",
			initVars: map[string]interface{}{
				"varB": 4,
			},
		},
		{
			name: "Op tag with dynamic calculation in 'a'",
			inputYAML: `!Op
  a: !Op
    a: 3
    op: "*"
    b: 2
  op: "*"
  b: 6`,
			expected: "36", // (3*2)*6
		},
		{
			name: "Op tag with dynamic calculation in 'b'",
			inputYAML: `!Op
  a: 20.0
  op: "/"
  b: !Op
    a: 8
    op: "-"
    b: 4`,
			expected: "5.0", // 20 / (8-4)
		},
		{
			name: "Op tag with nested variable substitution",
			inputYAML: `!Op
  a: !Var varX
  op: "%"
  b: !Var varY`,
			expected: "3",
			initVars: map[string]interface{}{
				"varX": 28,
				"varY": 5,
			},
		},
		{
			name: "Op tag with mixed static and dynamic values",
			inputYAML: `!Op
  a: !Op
    a: !Var varX
    op: "+"
    b: 10
  op: ">"
  b: 20`,
			expected: "true",
			initVars: map[string]interface{}{
				"varX": 15,
			},
		},
		{
			name: "Op tag with variable substitution for operation",
			inputYAML: `!Op
  a: 10
  op: !Var operation
  b: 5`,
			expected: "15", // Assuming operation is addition
			initVars: map[string]interface{}{
				"operation": "+",
			},
		},
		{
			name: "Op tag with dynamic operation determination",
			inputYAML: `!Op
  a: 7
  op: !If
    test: !Var condition
    then: "*"
    else: "/"
  b: 3`,
			expected: "21", // Multiplication if condition is true, division otherwise
			initVars: map[string]interface{}{
				"condition": true,
			},
		},
		{
			name: "Op tag with operation based on another Op result",
			inputYAML: `!Op
  a: 20
  op: !If
    test: !Op
      a: 1
      op: "=="
      b: 1
    then: "/"
    else: "*"
  b: 4`,
			expected: "5", // Division if 1 == 1, otherwise multiplication
		},
		{
			name: "Op tag with operation from concatenated string",
			inputYAML: `!Op
  a: 15
  op: !Join 
    items: [">", "="]
    separator: ""
  b: 15`,
			expected: "true", // Greater than or equal operation
		},
		{
			name: "Op tag with operation based on conditional logic",
			inputYAML: `!Op
  a: 12
  op: !If
    test: !Op
      a: 5
      op: ">"
      b: 3
    then: "-"
    else: "+"
  b: 2`,
			expected: "10", // Subtraction if 5 > 3, otherwise addition
		},

		{
			name: "Op tag string equality - true",
			inputYAML: `!Op
  a: "Hello"
  op: "=="
  b: "Hello"`,
			expected: "true",
		},
		{
			name: "Op tag string equality - false",
			inputYAML: `!Op
  a: "Hello"
  op: "=="
  b: "World"`,
			expected: "false",
		},
		{
			name: "Op tag string inequality - true",
			inputYAML: `!Op
  a: "Hello"
  op: "!="
  b: "World"`,
			expected: "true",
		},
		{
			name: "Op tag string inequality - false",
			inputYAML: `!Op
  a: "Hello"
  op: "!="
  b: "Hello"`,
			expected: "false",
		},
		{
			name: "Op tag string comparison - unsupported",
			inputYAML: `!Op
  a: "abc"
  op: "<"
  b: "def"`,
			expectError: true,
		},
		{
			name: "Op tag string concatenation",
			inputYAML: `!Op
  a: "Hello, "
  op: "+"
  b: "World!"`,
			expectError: true,
		},
		{
			name: "Op tag string with number - addition",
			inputYAML: `!Op
  a: "123"
  op: "+"
  b: 456`,
			expectError: true, // Since string to number conversion is not supported for addition
		},
		{
			name: "Op tag list equality - true",
			inputYAML: `!Op
  a: [1, 2, 3]
  op: "=="
  b: [1, 2, 3]`,
			expected: "true",
		},
		{
			name: "Op tag list equality - false",
			inputYAML: `!Op
  a: [1, 2, 3]
  op: "=="
  b: [3, 2, 1]`,
			expected: "false",
		},
		{
			name: "Op tag dictionary equality - true",
			inputYAML: `!Op
  a: {key1: "value1", key2: "value2"}
  op: "=="
  b: {key1: "value1", key2: "value2"}`,
			expected: "true",
		},
		{
			name: "Op tag dictionary equality - false",
			inputYAML: `!Op
  a: {key1: "value1", key2: "value2"}
  op: "=="
  b: {key1: "value1", key3: "value3"}`,
			expected: "false",
		},
		{
			name: "Op tag nested structures equality - true",
			inputYAML: `!Op
  a: {key1: [1, 2, 3], key2: {subkey: "value"}}
  op: "=="
  b: {key1: [1, 2, 3], key2: {subkey: "value"}}`,
			expected: "true",
		},
		{
			name: "Op tag nested structures equality - false",
			inputYAML: `!Op
  a: {key1: [1, 2, 3], key2: {subkey: "value"}}
  op: "=="
  b: {key1: [1, 2, 4], key2: {subkey: "value"}}`,
			expected: "false",
		},
		{
			name: "Op tag with mixed list and scalar - equality",
			inputYAML: `!Op
  a: [1, 2, 3]
  op: "=="
  b: "not a list"`,
			expected: "false",
		},
		{
			name: "Op tag with mixed dictionary and scalar - equality",
			inputYAML: `!Op
  a: {key: "value"}
  op: "=="
  b: "not a dictionary"`,
			expected: "false",
		},
	}

	runTests(t, tests)
}
