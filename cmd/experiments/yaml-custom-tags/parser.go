package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// parseArgs extracts specific keys and their corresponding values from a given YAML node.
// This function is designed to generalize the process of parsing arguments from a YAML mapping node.
//
// Parameters:
// - node: A pointer to a yaml.Node that should be a mapping node from which the keys are to be extracted.
// - keys: A slice of strings representing the keys to be extracted from the mapping node.
//
// Returns:
// - A map[string]*yaml.Node where each key is a string from the provided keys slice, and the value is the corresponding yaml.Node.
// - An error if the node is not a mapping node or if any of the required keys are missing in the node.
//
// Note: The function will return an error if any of the required keys are not found in the node.
func parseArgs(node *yaml.Node, keys []string) (map[string]*yaml.Node, error) {
	argsMap := make(map[string]*yaml.Node)
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("expected a mapping node")
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		if keyNode.Kind == yaml.ScalarNode {
			for _, key := range keys {
				if keyNode.Value == key {
					argsMap[key] = valueNode
				}
			}
		}
	}

	// Check if all required keys are present
	for _, key := range keys {
		if _, ok := argsMap[key]; !ok {
			return nil, fmt.Errorf("required key '%s' not found", key)
		}
	}

	return argsMap, nil
}

// parseURLEncodeArgs extracts 'url' and 'query' parameters from a YAML node and organizes them suitably for URL encoding.
// This function is specifically tailored for extracting URL and query parameters for URL encoding purposes.
//
// Parameters:
// - node: A pointer to a yaml.Node that should contain the 'url' and 'query' parameters in a mapping structure.
//
// Returns:
// - A string representing the URL extracted from the node.
// - A map[string]string where keys and values represent query parameters.
// - An error if the node doesn't contain the necessary structure or required keys ('url', and optionally 'query').
//
// Note: The 'query' parameter is optional and can be a mapping node containing key-value pairs of query parameters.
func parseURLEncodeArgs(node *yaml.Node) (string, map[string]string, error) {
	args, err := parseArgs(node, []string{"url", "query"})
	if err != nil {
		return "", nil, err
	}

	urlStr := args["url"].Value
	queryParams := make(map[string]string)
	if queryNode, ok := args["query"]; ok && queryNode.Kind == yaml.MappingNode {
		for i := 0; i < len(queryNode.Content); i += 2 {
			paramKey := queryNode.Content[i].Value
			paramValue := queryNode.Content[i+1].Value
			queryParams[paramKey] = paramValue
		}
	}

	return urlStr, queryParams, nil
}
