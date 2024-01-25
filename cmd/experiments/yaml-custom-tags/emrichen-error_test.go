package main

import "testing"

func TestEmrichenErrorTag(t *testing.T) {
	tests := []testCase{
		{
			name:               "Basic Error Triggering",
			inputYAML:          "!Error 'Test error'",
			expectError:        true,
			expectErrorMessage: "Test error",
		},
		{
			name:               "Conditional Error Triggering",
			inputYAML:          `!If {test: true, then: !Error "Conditional error", else: No error}`,
			expectError:        true,
			expectErrorMessage: "Conditional error",
		},
		{
			name:      "No Error on False Condition",
			inputYAML: `!If {test: false, then: !Error "Should not occur", else: 'No error'}`,
			expected:  "No error",
		},
		{
			name:               "Error in a Loop",
			inputYAML:          `!Loop {over: [1, 2, 3], template: !If {test: item == 2, then: !Error "Error in loop", else: item}}`,
			expectError:        true,
			expectErrorMessage: "Error in loop",
		},
		{
			name:               "Nested Error Tags",
			inputYAML:          `!Any [!Error "Inner error", 1]`,
			expectError:        true,
			expectErrorMessage: "Inner error",
		},
		{
			name:               "Error with Dynamic Message",
			inputYAML:          `!Error "Dynamic {{.message}}"`,
			expectError:        true,
			expectErrorMessage: "Dynamic error message",
			initVars: map[string]interface{}{
				"message": "error message",
			},
		},
		{
			name:               "Error in Nested Structures",
			inputYAML:          `!With {vars: {x: 5}, template: !If {test: x > 3, then: !Error "Too large", else: x}}`,
			expectError:        true,
			expectErrorMessage: "Too large",
		},
		{
			name:        "Error with Invalid Syntax",
			inputYAML:   `!Error {message: 123}`,
			expectError: true,
		},
		{
			name:               "Multiple Error Tags",
			inputYAML:          `!Error 'First error'; !Error 'Second error'`,
			expectError:        true,
			expectErrorMessage: "First error",
		},
		{
			name:               "Error Tag with No Message",
			inputYAML:          `!Error`,
			expectError:        true,
			expectErrorMessage: "",
		},
	}

	runTests(t, tests)
}
