package main

import "testing"

func TestWithBasicFunctionality(t *testing.T) {
	tests := []testCase{
		{
			name: "Basic With Tag Functionality",
			inputYAML: `!With
  vars:
    localVar: "local value"
  template: !Var localVar`,
			expected: `"local value"`,
		},
		{
			name: "Access Outside Variable Inside With",
			inputYAML: `!With
  vars:
    localVar: "local value"
  template: !Var globalVar`,
			expected: `"global value"`,
			initVars: map[string]interface{}{"globalVar": "global value"},
		},
		// 3. Nested `!With` Blocks Test
		{
			name: "Nested With Blocks Different Variables",
			inputYAML: `!With
  vars:
    outerVar: "outer"
  template: !With
    vars:
      innerVar: "inner"
    template: !Join [!Var outerVar, !Var innerVar]`,
			expected: `"outer inner"`,
		},
		{
			name: "Nested With Blocks Same Variable",
			inputYAML: `!With
  vars:
    var: "outer"
  template: !With
    vars:
      var: "inner"
    template: !Var var`,
			expected: `"inner"`,
		},

		// 4. Empty `!With` Block Test
		{
			name: "Empty With Block",
			inputYAML: `!With
  vars: {}
  template: "No vars"`,
			expected: `"No vars"`,
		},

		// 5. Invalid Variable Reference Test
		{
			name: "Invalid Variable Reference Inside With",
			inputYAML: `!With
  vars:
    validVar: "valid"
  template: !Var invalidVar`,
			expectError:        true,
			expectErrorMessage: "variable invalidVar not found",
		},

		// 6. Complex Variable Expressions Test
		{
			name: "Complex Expression Inside With",
			inputYAML: `!With
  vars:
    complexVar: !Join ["Hello", "World"]
  template: !Var complexVar`,
			expected: `"Hello World"`,
		},

		// 7. Interaction with Other Tags Test
		{
			name: "With Interaction with Loop",
			inputYAML: `!With
  vars:
    items: ["one", "two", "three"]
  template: !Loop
    over: !Var items
    as: item
    template: !Var item`,
			expected: `["one", "two", "three"]`,
		},
		{
			name: "With Interaction with If",
			inputYAML: `!With
  vars:
    condition: true
  template: !If
    test: !Var condition
    then: "Yes"
    else: "No"`,
			expected: `"Yes"`,
		},

		// 8. Reusing `!With` Blocks in Different Contexts Test
		{
			name: "Reusing With Block with Different Vars",
			inputYAML: `!Join 
  items: 
    - !With
        vars:
          localVar: "First"
        template: !Var localVar
    - !With
        vars:
          localVar: "Second"
        template: !Var localVar`,
			expected: `"First Second"`,
		},

		// 9. Error Handling Test
		{
			name: "Error Inside With Block",
			inputYAML: `!With
  vars:
    someVar: "value"
  template: !Error "An error occurred"`,
			expectError:        true,
			expectErrorMessage: "An error occurred",
		},
	}

	runTests(t, tests)
}
