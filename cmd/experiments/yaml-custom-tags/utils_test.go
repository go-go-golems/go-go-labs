package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"strconv"
	"testing"
)

// test case structure
type testCase struct {
	name      string
	inputYAML string
	expected  string
}

func runTests(t *testing.T, tests []testCase) {
	// Initialize EmrichenInterpreter
	ei := NewEmrichenInterpreter()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Parse input YAML
			inputNode := yaml.Node{}
			err := yaml.Unmarshal([]byte(tc.inputYAML), ei.CreateDecoder(&inputNode))
			require.NoError(t, err)

			// Process the input node
			resultNode, err := ei.Process(&inputNode)
			require.NoError(t, err)

			// Parse expected YAML
			expectedNode := yaml.Node{}
			err = yaml.Unmarshal([]byte(tc.expected), &expectedNode)
			require.NoError(t, err)

			expected_ := convertNodeToInterface(&expectedNode)
			//s, err := yaml.Marshal(expected_)
			//require.NoError(t, err)
			//fmt.Println(string(s))

			actual_ := convertNodeToInterface(resultNode)

			assert.Equal(t, expected_, actual_)
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
	// Attempt to convert to int, float, or bool, else return as string
	if i, err := strconv.Atoi(node.Value); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(node.Value, 64); err == nil {
		return f
	}
	if b, err := strconv.ParseBool(node.Value); err == nil {
		return b
	}
	return node.Value
}
