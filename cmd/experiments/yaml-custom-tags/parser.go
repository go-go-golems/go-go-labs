package main

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type parsedVariable struct {
	Name     string
	Expand   bool
	Required bool
}

func (ei *EmrichenInterpreter) parseArgs(
	node *yaml.Node,
	variables []parsedVariable,
) (map[string]*yaml.Node, error) {
	argsMap := make(map[string]*yaml.Node)
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("expected a mapping node")
	}

	varMap := make(map[string]parsedVariable)
	for _, v := range variables {
		varMap[v.Name] = v
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		parsedVar, ok := varMap[keyNode.Value]
		if !ok {
			return nil, errors.Errorf("unknown key '%s'", keyNode.Value)
		}
		key, ok := NodeToString(keyNode)
		if !ok {
			return nil, errors.Errorf("expected scalar key '%s'", keyNode.Value)
		}

		if parsedVar.Expand {
			value, err := ei.Process(valueNode)
			if err != nil {
				return nil, err
			}
			valueNode = value
		}
		argsMap[key] = valueNode
	}

	for _, v := range variables {
		if v.Required {
			if _, ok := argsMap[v.Name]; !ok {
				return nil, errors.Errorf("required key '%s' not found", v.Name)
			}
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
func (ei *EmrichenInterpreter) parseURLEncodeArgs(node *yaml.Node) (string, map[string]interface{}, error) {
	args, err := ei.parseArgs(node, []parsedVariable{
		{Name: "url", Required: true},
		{Name: "query", Expand: true},
	})
	if err != nil {
		return "", nil, err
	}

	url, err := ei.Process(args["url"])
	if err != nil {
		return "", nil, err
	}
	urlStr, ok := NodeToString(url)
	if !ok {
		return "", nil, errors.New("url must be a string")
	}

	// TODO need to process node
	queryParams := make(map[string]interface{})
	if queryNode, ok := args["query"]; ok && queryNode.Kind == yaml.MappingNode {
		for i := 0; i < len(queryNode.Content); i += 2 {
			paramKey := queryNode.Content[i].Value
			param, err := ei.Process(queryNode.Content[i+1])
			if err != nil {
				return "", nil, err
			}
			paramValue, ok := NodeToScalarInterface(param)
			if !ok {
				return "", nil, errors.New("query parameter value must be a scalar")
			}
			queryParams[paramKey] = paramValue
		}
	}

	return urlStr, queryParams, nil
}
