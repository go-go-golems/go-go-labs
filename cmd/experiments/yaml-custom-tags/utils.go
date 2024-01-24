package main

import (
	"github.com/pkg/errors"
	"strconv"

	"gopkg.in/yaml.v3"
)

// GetInt parses a YAML node to int.
func GetInt(node *yaml.Node) (int, bool) {
	if node.Kind == yaml.ScalarNode && (node.Tag == "!!int" || node.Tag == "!!float") {
		val, err := strconv.Atoi(node.Value)
		if err == nil {
			return val, true
		}
	}
	return 0, false
}

// GetFloat parses a YAML node to float.
func GetFloat(node *yaml.Node) (float64, bool) {
	if node.Kind == yaml.ScalarNode && (node.Tag == "!!float" || node.Tag == "!!int") {
		val, err := strconv.ParseFloat(node.Value, 64)
		if err == nil {
			return val, true
		}
	}
	return 0.0, false
}

// GetBool parses a YAML node to bool.
func GetBool(node *yaml.Node) (bool, bool) {
	if node.Kind == yaml.ScalarNode && node.Tag == "!!bool" {
		val, err := strconv.ParseBool(node.Value)
		if err == nil {
			return val, true
		}
	}
	return false, false
}

// GetString parses a YAML node to string.
func GetString(node *yaml.Node) (string, bool) {
	if node.Kind == yaml.ScalarNode && node.Tag == "!!str" {
		return node.Value, true
	}
	return "", false
}

// GetValue parses a YAML node into an interface{}.
func GetValue(node *yaml.Node) (interface{}, bool) {
	switch node.Kind {
	case yaml.ScalarNode:
		return getScalarValue(node)
	case yaml.SequenceNode:
		return getSliceValue(node)
	case yaml.MappingNode:
		return getMapValue(node)
	default:
		return nil, false
	}
}

func getScalarValue(node *yaml.Node) (interface{}, bool) {
	if val, ok := GetInt(node); ok {
		return val, true
	}
	if val, ok := GetFloat(node); ok {
		return val, true
	}
	if val, ok := GetBool(node); ok {
		return val, true
	}
	if val, ok := GetString(node); ok {
		return val, true
	}
	return nil, false
}

func getSliceValue(node *yaml.Node) ([]interface{}, bool) {
	var slice []interface{}
	for _, n := range node.Content {
		if val, ok := GetValue(n); ok {
			slice = append(slice, val)
		} else {
			return nil, false
		}
	}
	return slice, true
}

func getMapValue(node *yaml.Node) (map[string]interface{}, bool) {
	if len(node.Content)%2 != 0 {
		return nil, false // Invalid map node
	}

	m := make(map[string]interface{})
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		if keyNode.Kind != yaml.ScalarNode {
			return nil, false // Invalid key node
		}

		key := keyNode.Value
		if val, ok := GetValue(valueNode); ok {
			m[key] = val
		} else {
			return nil, false
		}
	}
	return m, true
}

func makeValue(value interface{}) (*yaml.Node, error) {
	if value == nil {
		return makeNil(), nil
	}
	switch v := value.(type) {
	case int:
		return makeInt(v), nil
	case int8:
		return makeInt(int(v)), nil
	case int16:
		return makeInt(int(v)), nil
	case int32:
		return makeInt(int(v)), nil
	case int64:
		return makeInt(int(v)), nil
	case uint:
		return makeInt(int(v)), nil
	case uint8:
		return makeInt(int(v)), nil
	case uint16:
		return makeInt(int(v)), nil
	case uint32:
		return makeInt(int(v)), nil
	case uint64:
		return makeInt(int(v)), nil
	case float64:
		return makeFloat(v), nil
	case float32:
		return makeFloat(float64(v)), nil
	case bool:
		return makeBool(v), nil
	case string:
		return makeString(v), nil

		// TODO(manuel, 2024-01-24) This needs to handle all kinds of slices and sequences and such
	case []interface{}:
		return makeSlice(v)
	case map[string]interface{}:
		return makeMap(v)

	default:
		return nil, errors.New("unsupported value type")
	}
}

func makeSlice(slice []interface{}) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind: yaml.SequenceNode,
		Tag:  "!!seq",
	}
	for _, elem := range slice {
		elemNode, err := makeValue(elem)
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, elemNode)
	}
	return node, nil
}

func makeMap(m map[string]interface{}) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}
	for key, value := range m {
		keyNode, err := makeValue(key)
		if err != nil {
			return nil, err
		}
		valueNode, err := makeValue(value)
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, keyNode, valueNode)
	}
	return node, nil
}

// makeString converts a string value to a corresponding scalar YAML node.
func makeString(value string) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: value,
	}
}

// makeInt converts an int value to a corresponding scalar YAML node.
func makeInt(value int) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!int",
		Value: strconv.Itoa(value),
	}
}

// makeNil converts a nil value to a corresponding scalar YAML node.
func makeNil() *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!null",
		Value: "null",
	}
}

// makeFloat converts a float value to a corresponding scalar YAML node.
func makeFloat(value float64) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!float",
		Value: strconv.FormatFloat(value, 'f', -1, 64),
	}
}

// makeBool converts a bool value to a corresponding scalar YAML node.
func makeBool(value bool) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!bool",
		Value: strconv.FormatBool(value),
	}
}

// isTruthy checks if the given YAML node represents a truthy value.
// It handles scalar, sequence and mapping nodes. Other node types are
// considered falsy. For scalars, an empty string, "false", "null" or "0"
// are considered falsy, other values are truthy. Sequences and mappings
// are truthy if they contain 1 or more items.
func isTruthy(node *yaml.Node) bool {
	//exhaustive:ignore
	switch node.Kind {
	case yaml.ScalarNode:
		return node.Value != "" && node.Value != "false" && node.Value != "null" && node.Value != "0"
	case yaml.SequenceNode, yaml.MappingNode:
		return len(node.Content) > 0
	default:
		return false
	}
}

// findWithNodes searches for the 'vars' and 'template' nodes within the given 'content' YAML nodes.
// It returns two pointers to the 'vars' and 'template' nodes respectively.
func findWithNodes(content []*yaml.Node) (*yaml.Node, *yaml.Node) {
	var varsNode *yaml.Node
	var templateNode *yaml.Node
	for i := 0; i < len(content); i += 2 {
		keyNode := content[i]
		valueNode := content[i+1]
		if keyNode.Kind == yaml.ScalarNode {
			if keyNode.Value == "vars" {
				varsNode = valueNode
			} else if keyNode.Value == "template" {
				templateNode = valueNode
			}
		}
	}
	return varsNode, templateNode
}
