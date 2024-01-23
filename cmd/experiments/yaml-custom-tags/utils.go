package main

import (
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

// toInt parses a YAML node to int.
func toInt(node *yaml.Node) (int, bool) {
	if node.Kind == yaml.ScalarNode && (node.Tag == "!!int" || node.Tag == "!!float") {
		val, err := strconv.Atoi(node.Value)
		if err == nil {
			return val, true
		}
	}
	return 0, false
}

// toFloat parses a YAML node to float.
func toFloat(node *yaml.Node) (float64, bool) {
	if node.Kind == yaml.ScalarNode && (node.Tag == "!!float" || node.Tag == "!!int") {
		val, err := strconv.ParseFloat(node.Value, 64)
		if err == nil {
			return val, true
		}
	}
	return 0.0, false
}

// toBool parses a YAML node to bool.
func toBool(node *yaml.Node) (bool, bool) {
	if node.Kind == yaml.ScalarNode && node.Tag == "!!bool" {
		val, err := strconv.ParseBool(node.Value)
		if err == nil {
			return val, true
		}
	}
	return false, false
}

// toString parses a YAML node to string.
func toString(node *yaml.Node) (string, bool) {
	if node.Kind == yaml.ScalarNode && node.Tag == "!!str" {
		return node.Value, true
	}
	return "", false
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

// deepCompareYamlNodes takes pointers to two yaml Node structures 'a' and 'b', and performs a deep comparison between them.
// It checks each aspect of the Nodes, including their Kind, Content, Value, Tag, and Alias (if applicable), and returns
// 'true' if they are identical, 'false' otherwise. When comparing Content (in the case of DocumentNode, SequenceNode, and
// MappingNode) it recursively calls itself on each element.
// Unknown node types result in a return value of 'false'.
func deepCompareYamlNodes(a, b *yaml.Node) bool {
	if a.Kind != b.Kind {
		return false
	}

	switch a.Kind {
	case yaml.DocumentNode, yaml.SequenceNode, yaml.MappingNode:
		if len(a.Content) != len(b.Content) {
			return false
		}
		for i := 0; i < len(a.Content); i++ {
			if !deepCompareYamlNodes(a.Content[i], b.Content[i]) {
				return false
			}
		}
		return true

	case yaml.ScalarNode:
		return a.Value == b.Value && a.Tag == b.Tag

	case yaml.AliasNode:
		// Handle alias nodes if necessary
		return reflect.DeepEqual(a.Alias, b.Alias)

	default:
		// Unknown node type
		return false
	}
}
