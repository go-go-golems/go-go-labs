package main

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (ei *EmrichenInterpreter) handleMerge(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, errors.New("!Merge requires a sequence of mapping nodes")
	}

	mergedMap := make(map[string]*yaml.Node)
	for _, item := range node.Content {
		if item.Kind != yaml.MappingNode {
			return nil, errors.New("!Merge items must be mapping nodes")
		}

		tempMap := make(map[string]*yaml.Node)
		for i := 0; i < len(item.Content); i += 2 {
			keyNode := item.Content[i]
			valueNode := item.Content[i+1]

			resolvedValue, err := ei.Process(valueNode)
			if err != nil {
				return nil, err
			}

			tempMap[keyNode.Value] = resolvedValue
		}

		for k, v := range tempMap {
			mergedMap[k] = v
		}
	}

	mergedContent := []*yaml.Node{}
	for k, v := range mergedMap {
		mergedContent = append(mergedContent, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: k,
		}, v)
	}

	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: mergedContent,
	}, nil

}
