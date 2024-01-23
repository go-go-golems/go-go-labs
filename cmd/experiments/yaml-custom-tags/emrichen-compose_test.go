package main

import (
	"encoding/base64"
	"testing"
)

func base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
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
			expected:  base64Encode(""), // Base64 of ""
			initVars: map[string]interface{}{
				"emptyString": "",
			},
		},
		{
			name:      "Base64 encoding of a null variable",
			inputYAML: "!Base64,Var nullVar",
			expected:  base64Encode(""), // Base64 of ""
			initVars: map[string]interface{}{
				"nullVar": nil,
			},
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
			name:      "Base64 encoding after variable substitution and formatting",
			inputYAML: `!Base64,Format "{var1} and {var2}",Var composedVars`,
			expected:  "SGVsbG8gYW5kIFdvcmxk", // Base64 of "Hello and World"
			initVars: map[string]interface{}{
				"composedVars": map[string]interface{}{
					"var1": "Hello",
					"var2": "World",
				},
			},
		},
		{
			name:      "Concatenation of lists after variable substitution",
			inputYAML: `!Concat,Var list1,Var list2`,
			expected:  "[1, 2, 3, 4, 5]", // Concatenation of two lists
			initVars: map[string]interface{}{
				"list1": []interface{}{1, 2, 3},
				"list2": []interface{}{4, 5},
			},
		},
		{
			name:      "Joining strings after base64 encoding and variable substitution",
			inputYAML: `!Join,Base64,Var joinVars`,
			expected:  "SGVsbG8gV29ybGQ=", // Joined string in Base64
			initVars: map[string]interface{}{
				"joinVars": []interface{}{"Hello", "World"},
			},
		},
		{
			name:      "URL encoding after format and variable substitution",
			inputYAML: `!URLEncode,Format "{protocol}://{domain}",Var urlVars`,
			expected:  "https%3A%2F%2Fexample.com", // URL encoded format
			initVars: map[string]interface{}{
				"urlVars": map[string]interface{}{
					"protocol": "https",
					"domain":   "example.com",
				},
			},
		},
		{
			name:      "Lookup after variable substitution in a nested mapping",
			inputYAML: `!Lookup people[0].name,Var peopleData`,
			expected:  "Alice",
			initVars: map[string]interface{}{
				"peopleData": map[string]interface{}{
					"people": []interface{}{
						map[string]interface{}{"name": "Alice"},
						map[string]interface{}{"name": "Bob"},
					},
				},
			},
		},
	}

	runTests(t, tests)
}
