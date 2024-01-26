package main

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
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
	expectPanic        bool
}

func runTests(t *testing.T, tests []testCase) {
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ei, err := NewEmrichenInterpreter(WithVars(tc.initVars))
			require.NoError(t, err)

			decoder := yaml.NewDecoder(bytes.NewReader([]byte(tc.inputYAML)))

			hadError := false
			var resultNode *yaml.Node
			if tc.expectPanic {
				resultNode, err = expectPanic(t, func() (*yaml.Node, error) {
					for {
						inputNode := yaml.Node{}
						// Parse input YAML
						err2 := decoder.Decode(ei.CreateDecoder(&inputNode))
						require.NoError(t, err2)
					}
				})
				require.Error(t, err)
				require.Equal(t, "paniced", err.Error())
				return
			}

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

func expectPanic(t *testing.T, f func() (*yaml.Node, error)) (*yaml.Node, error) {
	didPanic := false

	_, _ = func() (*yaml.Node, error) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("paniced")
				didPanic = true
			}
		}()
		return f()
	}()

	if !didPanic {
		t.Errorf("Expected a panic to occur, but none did")
	}

	return nil, errors.New("paniced")
}
