package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"testing"
)

func TestHandleIncludeBase64InDepth(t *testing.T) {
	base64EncodedHelloWorld := base64.StdEncoding.EncodeToString([]byte("Hello World"))

	tests := []testCase{
		{
			name:               "Base64 valid test",
			inputYAML:          "!IncludeBase64 test-data/base64_test.bin",
			expected:           base64.StdEncoding.EncodeToString([]byte("Hello World")),
			initVars:           nil,
			expectError:        false,
			expectErrorMessage: "",
			expectPanic:        false,
		},
		{
			name:               "Base64 file not found",
			inputYAML:          "!IncludeBase64 test-data/nonexistent.bin",
			expected:           "",
			initVars:           nil,
			expectError:        true,
			expectErrorMessage: "error reading file for !IncludeBase64: open test-data/nonexistent.bin: no such file or directory",
			expectPanic:        false,
		},
		{
			name: "Base64 in nested map",
			inputYAML: `
nested:
  key: !IncludeBase64 test-data/base64_test.bin
            `,
			expected: fmt.Sprintf(`
nested:
  key: %s
            `, base64EncodedHelloWorld),
			initVars:           nil,
			expectError:        false,
			expectErrorMessage: "",
			expectPanic:        false,
		},
		{
			name: "Base64 in nested array",
			inputYAML: `
array:
  - item1
  - !IncludeBase64 test-data/base64_test.bin
  - item3
            `,
			expected: fmt.Sprintf(`
array:
  - item1
  - %s
  - item3
            `, base64EncodedHelloWorld),
			initVars:           nil,
			expectError:        false,
			expectErrorMessage: "",
			expectPanic:        false,
		},
		{
			name: "Base64 in deep nested structure",
			inputYAML: `
outer:
  inner:
    key: !IncludeBase64 test-data/base64_test.bin
            `,
			expected: fmt.Sprintf(`
outer:
  inner:
    key: %s
            `, base64EncodedHelloWorld),
			initVars:           nil,
			expectError:        false,
			expectErrorMessage: "",
			expectPanic:        false,
		},
		// Additional deep structure test cases here...
	}

	runTests(t, tests)
}

func TestHandleIncludeBinaryWithBase64(t *testing.T) {
	binaryContent, _ := os.ReadFile("test-data/binary_test.bin")
	base64EncodedBinaryContent := base64.StdEncoding.EncodeToString(binaryContent)

	tests := []testCase{
		{
			name:               "IncludeBinary with Base64 encoding",
			inputYAML:          `!Base64,IncludeBinary test-data/binary_test.bin`,
			expected:           base64EncodedBinaryContent,
			initVars:           nil,
			expectError:        false,
			expectErrorMessage: "",
			expectPanic:        false,
		},
		{
			name: "IncludeBinary with Base64 encoding in nested map",
			inputYAML: fmt.Sprintf(`
nested:
  key: !Base64,IncludeBinary test-data/binary_test.bin
            `),
			expected: fmt.Sprintf(`
nested:
  key: %s
            `, base64EncodedBinaryContent),
			initVars:           nil,
			expectError:        false,
			expectErrorMessage: "",
			expectPanic:        false,
		},
		{
			name: "IncludeBinary with Base64 encoding in nested array",
			inputYAML: fmt.Sprintf(`
array:
  - item1
  - !Base64,IncludeBinary test-data/binary_test.bin
  - item3
            `),
			expected: fmt.Sprintf(`
array:
  - item1
  - %s
  - item3
            `, base64EncodedBinaryContent),
			initVars:           nil,
			expectError:        false,
			expectErrorMessage: "",
			expectPanic:        false,
		},
	}

	runTests(t, tests)
}

func TestHandleIncludeGlob(t *testing.T) {
	tests := []testCase{
		{
			name: "IncludeGlob with variable resolution",
			inputYAML: `
!With
  vars:
    msg: Hello
    txt: World
  template: !IncludeGlob test-data/glob_test_*.yml
            `,
			expected: `
- message: Hello
- text: World
            `,
			initVars:           nil,
			expectError:        false,
			expectErrorMessage: "",
			expectPanic:        false,
		},
		{
			name:      "Glob pattern matching multiple files",
			inputYAML: `!IncludeGlob test-data/file_glob_*.yml`,
			expected: `
- item: file1
- item: file2
            `,
			initVars:           nil,
			expectError:        false,
			expectErrorMessage: "",
			expectPanic:        false,
		},
		{
			name:               "Glob pattern no matching files",
			inputYAML:          `!IncludeGlob test-data/nonexistent_*.yml`,
			expected:           "[]", // Expecting an empty list when no files match
			initVars:           nil,
			expectError:        false,
			expectErrorMessage: "",
			expectPanic:        false,
		},
		{
			name:        "Glob pattern with malformed file",
			inputYAML:   `!IncludeGlob test-data/glob_malformed*.yml`,
			expected:    "",
			initVars:    nil,
			expectError: true,
		},
	}

	runTests(t, tests)
}

func TestHandleInclude(t *testing.T) {
	tests := []testCase{
		{
			name:               "Glob valid test",
			inputYAML:          "!Include test-data/file_glob_1.yml",
			expected:           "item: file1",
			initVars:           nil,
			expectError:        false,
			expectErrorMessage: "",
			expectPanic:        false,
		},
		{
			name:      "Glob multi doc",
			inputYAML: "!Include test-data/multi-doc.yml",
			expected: `- foo: bla
- blop: blip
- hello: furg
`,
			initVars:           nil,
			expectError:        false,
			expectErrorMessage: "",
			expectPanic:        false,
		},
		// Additional test cases here...
	}

	runTests(t, tests)
}

func TestHandleIncludeText(t *testing.T) {
	tests := []testCase{
		{
			name:               "Text valid test",
			inputYAML:          "!IncludeText test-data/text_test.txt",
			expected:           "Sample Text Content",
			initVars:           nil,
			expectError:        false,
			expectErrorMessage: "",
			expectPanic:        false,
		},
		// Additional test cases here...
	}

	runTests(t, tests)
}
