package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"io"
	"strconv"
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

func TestNodeToInt(t *testing.T) {
	tests := []struct {
		name     string
		node     *yaml.Node
		expected int
		ok       bool
	}{
		{
			name: "Valid integer node",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!int",
				Value: "42",
			},
			expected: 42,
			ok:       true,
		},
		{
			name: "Valid float node truncated to int",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!float",
				Value: "42.5",
			},
			ok: false,
		},
		{
			name: "Non-integer string node",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!int",
				Value: "not_an_int",
			},
			expected: 0,
			ok:       false,
		},
		{
			name: "Non-integer, non-float node",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: "42",
			},
			expected: 0,
			ok:       false,
		},
		{
			name:     "Nil node",
			node:     nil,
			expected: 0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := NodeToInt(tt.node)
			require.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestNodeToFloat(t *testing.T) {
	tests := []struct {
		name     string
		node     *yaml.Node
		expected float64
		ok       bool
	}{
		{
			name: "Valid float node",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!float",
				Value: "42.5",
			},
			expected: 42.5,
			ok:       true,
		},
		{
			name: "Valid integer node as float",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!int",
				Value: "42",
			},
			expected: 42.0,
			ok:       true,
		},
		{
			name: "Non-float string node",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!float",
				Value: "not_a_float",
			},
			expected: 0.0,
			ok:       false,
		},
		{
			name:     "Nil node",
			node:     nil,
			expected: 0.0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := NodeToFloat(tt.node)
			require.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestNodeToBool(t *testing.T) {
	tests := []struct {
		name     string
		node     *yaml.Node
		expected bool
		ok       bool
	}{
		{
			name: "Valid boolean node true",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!bool",
				Value: "true",
			},
			expected: true,
			ok:       true,
		},
		{
			name: "Valid boolean node false",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!bool",
				Value: "false",
			},
			expected: false,
			ok:       true,
		},
		{
			name: "Non-boolean node",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: "true",
			},
			expected: false,
			ok:       false,
		},
		{
			name:     "Nil node",
			node:     nil,
			expected: false,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := NodeToBool(tt.node)
			require.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestNodeToString(t *testing.T) {
	tests := []struct {
		name     string
		node     *yaml.Node
		expected string
		ok       bool
	}{
		{
			name: "Valid string node",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: "hello",
			},
			expected: "hello",
			ok:       true,
		},
		{
			name: "Non-string node",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!int",
				Value: "123",
			},
			expected: "",
			ok:       false,
		},
		{
			name:     "Nil node",
			node:     nil,
			expected: "",
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := NodeToString(tt.node)
			require.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestNodeToInterface(t *testing.T) {
	tests := []struct {
		name     string
		node     *yaml.Node
		expected interface{}
		ok       bool
	}{
		{
			name: "Scalar int node",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!int",
				Value: "42",
			},
			expected: 42,
			ok:       true,
		},
		{
			name: "Scalar float node",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!float",
				Value: "42.5",
			},
			expected: 42.5,
			ok:       true,
		},
		{
			name: "Scalar bool node",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!bool",
				Value: "true",
			},
			expected: true,
			ok:       true,
		},
		{
			name: "Scalar string node",
			node: &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: "hello",
			},
			expected: "hello",
			ok:       true,
		},
		{
			name: "Sequence node with mixed types",
			node: &yaml.Node{
				Kind: yaml.SequenceNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Tag: "!!int", Value: "1"},
					{Kind: yaml.ScalarNode, Tag: "!!str", Value: "two"},
					{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"},
				},
			},
			expected: []interface{}{1, "two", true},
			ok:       true,
		},
		{
			name: "Mapping node with string keys",
			node: &yaml.Node{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Tag: "!!str", Value: "key1"},
					{Kind: yaml.ScalarNode, Tag: "!!int", Value: "10"},
					{Kind: yaml.ScalarNode, Tag: "!!str", Value: "key2"},
					{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "false"},
				},
			},
			expected: map[string]interface{}{"key1": 10, "key2": false},
			ok:       true,
		},
		{
			name: "Unsupported node kind",
			node: &yaml.Node{
				Kind: yaml.AliasNode, // AliasNode is not supported by NodeToInterface
			},
			expected: nil,
			ok:       false,
		},
		{
			name: "Sequence node with unsupported node type",
			node: &yaml.Node{
				Kind: yaml.SequenceNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Tag: "!!int", Value: "1"},
					{Kind: yaml.AliasNode}, // Unsupported node type
				},
			},
			expected: nil,
			ok:       false,
		},
		{
			name: "Mapping node with odd number of child nodes",
			node: &yaml.Node{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Tag: "!!str", Value: "key1"},
					{Kind: yaml.ScalarNode, Tag: "!!int", Value: "10"},
					{Kind: yaml.ScalarNode, Tag: "!!str", Value: "key2"},
				},
			},
			expected: nil,
			ok:       false,
		},
		{
			name: "Mapping node with invalid key node",
			node: &yaml.Node{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{Kind: yaml.SequenceNode}, // Invalid key node
					{Kind: yaml.ScalarNode, Tag: "!!int", Value: "10"},
				},
			},
			expected: nil,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := NodeToInterface(tt.node)
			require.Equal(t, tt.ok, ok)
			if ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
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
