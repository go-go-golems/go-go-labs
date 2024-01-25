package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"io"
	"testing"
)

type testCase struct {
	name               string
	inputYAML          string
	expected           string
	initVars           map[string]interface{} // Adding a new field for initial variable bindings
	expectError        bool
	expectErrorMessage string
}

func runTests(t *testing.T, tests []testCase) {
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ei, err := NewEmrichenInterpreter(WithVars(tc.initVars))
			require.NoError(t, err)

			decoder := yaml.NewDecoder(bytes.NewReader([]byte(tc.inputYAML)))

			hadError := false
			var resultNode *yaml.Node
			for {
				inputNode := yaml.Node{}
				// Parse input YAML
				err2 := decoder.Decode(ei.CreateDecoder(&inputNode))
				if err2 == io.EOF {
					break
				}
				err = err2
				if err != nil {
					hadError = true
					break
				}

				resultNode, err = ei.Process(&inputNode)
				if err != nil {
					hadError = true
					break
				}
			}

			if hadError {
				if tc.expectError {
					require.Error(t, err, "Expected an error but got none")
					if tc.expectErrorMessage != "" {
						assert.Equal(t, tc.expectErrorMessage, err.Error())
					}
				} else {
					require.NoError(t, err, "Unexpected error encountered", err)
				}
				return
			} else {
				require.NoError(t, err, "Unexpected error encountered", err)
			}

			expectedNode := yaml.Node{}
			err = yaml.Unmarshal([]byte(tc.expected), &expectedNode)
			require.NoError(t, err)

			expected_ := convertNodeToInterface(&expectedNode)
			actual_ := convertNodeToInterface(resultNode)

			assert.Equal(t, expected_, actual_)
		})
	}
}
