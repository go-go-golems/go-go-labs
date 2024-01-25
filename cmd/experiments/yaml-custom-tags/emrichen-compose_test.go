package main

import (
	"encoding/base64"
	"testing"
)

func base64Encode(data string) string {
	ret := base64.StdEncoding.EncodeToString([]byte(data))
	return ret
}

func TestEmrichenBase64VarComposition(t *testing.T) {
	tests := []testCase{
		{
			name:      "Base64 encoding of a simple string variable",
			inputYAML: "!Base64,Var simpleString",
			expected:  base64Encode("Hello World"),
			initVars: map[string]interface{}{
				"simpleString": "Hello World",
			},
		},
		{
			name:      "Base64 encoding of an integer variable",
			inputYAML: "!Base64,Var integer",
			expected:  base64Encode("123"), // Base64 of "123"
			initVars: map[string]interface{}{
				"integer": 123,
			},
		},
		{
			name:      "Base64 encoding of a boolean variable",
			inputYAML: "!Base64,Var boolean",
			expected:  base64Encode("true"), // Base64 of "true"
			initVars: map[string]interface{}{
				"boolean": true,
			},
		},
		{
			name:      "Base64 encoding of an empty string variable",
			inputYAML: "!Base64,Var emptyString",
			expected:  "\"\"",
			initVars: map[string]interface{}{
				"emptyString": "",
			},
		},
		{
			name:      "Base64 encoding of a null variable",
			inputYAML: "!Base64,Var nullVar",
			initVars: map[string]interface{}{
				"nullVar": nil,
			},
			expected: "bnVsbA==", // Base64 of "null"
		},
		{
			name:        "Error case with undefined variable",
			inputYAML:   "!Base64,Var undefinedVar",
			expected:    "",
			initVars:    map[string]interface{}{}, // No variable defined
			expectError: true,
		},

		{
			name:               "Error in the middle of a composition chain",
			inputYAML:          `!Base64,Var,Error "Error message"`,
			expected:           "",
			expectError:        true,
			expectErrorMessage: "Error message",
		},
		{
			name:               "Error at the beginning of a composition chain",
			inputYAML:          `!Var,Base64,Error "Early error"`,
			expected:           "",
			expectError:        true,
			expectErrorMessage: "Early error",
		},
		{
			name:               "Error at the end of a composition chain",
			inputYAML:          `!Error "Final error",Var,Base64`,
			expected:           "",
			expectError:        true,
			expectErrorMessage: "Final error",
		},
		{
			name:               "Error with other tags in a long composition chain",
			inputYAML:          `!Base64,Concat,Var,Error "Complex error",Join,Lookup`,
			expected:           "",
			expectError:        true,
			expectErrorMessage: "Complex error",
		},

		{
			name:      "Base64 strings after joining and variable substitution",
			inputYAML: `!Base64,Join,Var joinVars`,
			expected:  "SGVsbG8gV29ybGQ=", // Joined string in Base64
			initVars: map[string]interface{}{
				"joinVars": map[string]interface{}{
					"items": []interface{}{"Hello", "World"},
				},
			},
		},
	}

	runTests(t, tests)
}

func TestEmrichenNotAndIsTags(t *testing.T) {
	tests := []testCase{
		{
			name:      "Not with IsBoolean - Direct",
			inputYAML: `!Not,IsBoolean true`,
			expected:  "false",
		},
		{
			name:      "Not with IsBoolean - With Var",
			inputYAML: `!Not,IsBoolean,Var booleanVar`,
			expected:  "false",
			initVars:  map[string]interface{}{"booleanVar": true},
		},
		{
			name:      "Not with IsDict - Direct",
			inputYAML: `!Not,IsDict {a: 1, b: 2}`,
			expected:  "false",
		},
		{
			name:      "Not with IsDict - With Var",
			inputYAML: `!Not,IsDict,Var dictVar`,
			expected:  "false",
			initVars:  map[string]interface{}{"dictVar": map[string]interface{}{"key": "value"}},
		},
		{
			name:      "Not with IsInteger - Direct",
			inputYAML: `!Not,IsInteger 42`,
			expected:  "false",
		},
		{
			name:      "Not with IsInteger - With Var",
			inputYAML: `!Not,IsInteger,Var intVar`,
			expected:  "false",
			initVars:  map[string]interface{}{"intVar": 123},
		},
		{
			name:      "Not with IsList - Direct",
			inputYAML: `!Not,IsList [1, 2, 3]`,
			expected:  "false",
		},
		{
			name:      "Not with IsList - With Var",
			inputYAML: `!Not,IsList,Var listVar`,
			expected:  "false",
			initVars:  map[string]interface{}{"listVar": []interface{}{1, 2, 3}},
		},
	}

	runTests(t, tests)
}
