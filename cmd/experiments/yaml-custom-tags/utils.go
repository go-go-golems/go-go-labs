package main

import (
	"github.com/pkg/errors"
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

// NodeToInt parses a YAML node to int.
func NodeToInt(node *yaml.Node) (int, bool) {
	if node == nil {
		return 0, false
	}
	if node.Kind == yaml.ScalarNode && (node.Tag == "!!int" || node.Tag == "!!float") {
		val, err := strconv.Atoi(node.Value)
		if err == nil {
			return val, true
		}
	}
	return 0, false
}

// NodeToFloat parses a YAML node to float.
func NodeToFloat(node *yaml.Node) (float64, bool) {
	if node == nil {
		return 0.0, false
	}
	if node.Kind == yaml.ScalarNode && (node.Tag == "!!float" || node.Tag == "!!int") {
		val, err := strconv.ParseFloat(node.Value, 64)
		if err == nil {
			return val, true
		}
	}
	return 0.0, false
}

// NodeToBool parses a YAML node to bool.
func NodeToBool(node *yaml.Node) (bool, bool) {
	if node == nil {
		return false, false
	}
	if node.Kind == yaml.ScalarNode && node.Tag == "!!bool" {
		val, err := strconv.ParseBool(node.Value)
		if err == nil {
			return val, true
		}
	}
	return false, false
}

// NodeToString parses a YAML node to string.
func NodeToString(node *yaml.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	if node.Kind == yaml.ScalarNode && node.Tag == "!!str" {
		return node.Value, true
	}
	return "", false
}

// NodeToInterface parses a YAML node into an interface{}.
func NodeToInterface(node *yaml.Node) (interface{}, bool) {
	switch node.Kind {
	case yaml.ScalarNode:
		return getScalarValue(node)
	case yaml.SequenceNode:
		return NodeToSlice(node)
	case yaml.MappingNode:
		return NodeToMap(node)
	default:
		return nil, false
	}
}

func getScalarValue(node *yaml.Node) (interface{}, bool) {
	if node == nil {
		return nil, false
	}
	if node.Tag == "!!null" {
		return nil, true
	}
	if val, ok := NodeToInt(node); ok {
		return val, true
	}
	if val, ok := NodeToFloat(node); ok {
		return val, true
	}
	if val, ok := NodeToBool(node); ok {
		return val, true
	}
	if val, ok := NodeToString(node); ok {
		return val, true
	}
	return nil, false
}

func NodeToSlice(node *yaml.Node) ([]interface{}, bool) {
	if node == nil {
		return nil, false
	}
	var slice []interface{}
	for _, n := range node.Content {
		if val, ok := NodeToInterface(n); ok {
			slice = append(slice, val)
		} else {
			return nil, false
		}
	}
	return slice, true
}

func NodeToMap(node *yaml.Node) (map[string]interface{}, bool) {
	if node == nil {
		return nil, false
	}
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
		if val, ok := NodeToInterface(valueNode); ok {
			m[key] = val
		} else {
			return nil, false
		}
	}
	return m, true
}

func ValueToNode(value interface{}) (*yaml.Node, error) {
	if value == nil {
		return makeNil(), nil
	}

	// Use reflection to handle dynamic types
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return makeInt(int(v.Int())), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return makeInt(int(v.Uint())), nil
	case reflect.Float32, reflect.Float64:
		return makeFloat(v.Float()), nil
	case reflect.Bool:
		return makeBool(v.Bool()), nil
	case reflect.String:
		return makeString(v.String()), nil
	case reflect.Slice, reflect.Array:
		return genericSliceToNode(v)
	case reflect.Map:
		return genericMapToNode(v)
	default:
		return nil, errors.Errorf("unsupported type %T", value)
	}
}

func genericSliceToNode(slice reflect.Value) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind: yaml.SequenceNode,
		Tag:  "!!seq",
	}
	for i := 0; i < slice.Len(); i++ {
		elem := slice.Index(i).Interface()
		elemNode, err := ValueToNode(elem)
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, elemNode)
	}
	return node, nil
}

func genericMapToNode(mapValue reflect.Value) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}
	for _, key := range mapValue.MapKeys() {
		keyNode, err := ValueToNode(key.Interface())
		if err != nil {
			return nil, err
		}
		valueNode, err := ValueToNode(mapValue.MapIndex(key).Interface())
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
