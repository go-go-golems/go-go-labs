package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"testing"
)

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
