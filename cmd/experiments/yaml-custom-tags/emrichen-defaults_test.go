package main

import "testing"

func TestEmrichenDefaultsTag(t *testing.T) {
	tests := []testCase{
		{
			name: "Basic Usage Test",
			inputYAML: `
!Defaults
var1: default1
var2: default2
---
var1: !Var var1
var2: !Var var2`,
			expected: `
var1: default1
var2: default2`,
		},
		{
			name: "Override Default Values Test",
			inputYAML: `
!Defaults
var1: default1
var2: default2
---
var1: newvalue1
var2: !Var var2`,
			expected: `
var1: newvalue1
var2: default2`,
		},
		{
			name: "Multiple Defaults Test",
			inputYAML: `
!Defaults
var1: default1
---
!Defaults
var1: newdefault1
var2: default2
---
var1: !Var var1
var2: !Var var2`,
			expected: `
var1: newdefault1
var2: default2`,
		},
		{
			name: "Empty Defaults Block Test",
			inputYAML: `
!Defaults
---
var1: !Var var1`,
			expected: `
var1: null`,
			expectError: true, // If the behavior is to throw an error when the variable is not defined
		},
		// Add more tests for Invalid Data Types, Nested Variables, Compatibility with Other Tags, etc.
	}

	runTests(t, tests)
}
