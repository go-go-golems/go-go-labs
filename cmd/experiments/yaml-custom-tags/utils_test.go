package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"strconv"
	"testing"
)

type testCase struct {
	name        string
	inputYAML   string
	expected    string
	initVars    map[string]interface{} // Adding a new field for initial variable bindings
	expectError bool
}

func runTests(t *testing.T, tests []testCase) {
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set initial variables using WithVars
			vars := make(map[string]*yaml.Node)
			for k, v := range tc.initVars {
				node := yaml.Node{}
				err := node.Encode(v)
				require.NoError(t, err)
				vars[k] = &node
			}

			ei := NewEmrichenInterpreter(WithVars(vars))

			// Parse input YAML
			inputNode := yaml.Node{}
			err := yaml.Unmarshal([]byte(tc.inputYAML), ei.CreateDecoder(&inputNode))
			if tc.expectError {
				require.Error(t, err, "Expected an error but got none")
				return
			} else {
				require.NoError(t, err, "Unexpected error encountered")
			}

			resultNode, err := ei.Process(&inputNode)

			// Process the input node
			if tc.expectError {
				require.Error(t, err, "Expected an error but got none")
			} else {
				require.NoError(t, err, "Unexpected error encountered")

				// Parse expected YAML
				expectedNode := yaml.Node{}
				err = yaml.Unmarshal([]byte(tc.expected), &expectedNode)
				require.NoError(t, err)

				expected_ := convertNodeToInterface(&expectedNode)
				actual_ := convertNodeToInterface(resultNode)

				assert.Equal(t, expected_, actual_)
			}
		})
	}
}

// convertNodeToInterface converts a yaml.Node into a corresponding Go type.
func convertNodeToInterface(node *yaml.Node) interface{} {
	switch node.Kind {
	case yaml.DocumentNode:
		// Assuming the document has a single root element
		if len(node.Content) == 1 {
			return convertNodeToInterface(node.Content[0])
		}

	case yaml.MappingNode:
		m := make(map[string]interface{})
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			value := convertNodeToInterface(node.Content[i+1])
			m[key] = value
		}
		return map[string]interface{}{
			"type":  "Mapping",
			"tag":   node.Tag,
			"value": m,
		}

	case yaml.SequenceNode:
		var s []interface{}
		for _, n := range node.Content {
			s = append(s, convertNodeToInterface(n))
		}
		return map[string]interface{}{
			"type":  "Sequence",
			"tag":   node.Tag,
			"value": s,
		}

	case yaml.ScalarNode:
		v := convertScalarValue(node)
		return map[string]interface{}{
			"type":  "Scalar",
			"tag":   node.Tag,
			"value": v,
		}

	case yaml.AliasNode:
		return map[string]interface{}{
			"type":  "Alias",
			"tag":   node.Tag,
			"value": node.Alias,
		}
	}

	return nil
}

// convertScalarValue converts a scalar YAML node to a primitive Go type.
func convertScalarValue(node *yaml.Node) interface{} {
	switch node.Tag {
	case "!!int":
		i, err := strconv.Atoi(node.Value)
		if err != nil {
			return node.Value
		}
		return i

	case "!!float":
		f, err := strconv.ParseFloat(node.Value, 64)
		if err != nil {
			return node.Value
		}
		return f

	case "!!bool":
		b, err := strconv.ParseBool(node.Value)
		if err != nil {
			return node.Value
		}
		return b

	case "!!str":
		return node.Value

	default:
		return node.Value
	}
}
