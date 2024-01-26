package main

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (ei *EmrichenInterpreter) handleLoop(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("!Loop requires a mapping node")
	}

	args, err := ei.parseArgs(node, []parsedVariable{
		{Name: "over", Required: true, Expand: true},
		{Name: "template", Required: true},
		{Name: "as"},
	})
	if err != nil {
		return nil, err
	}

	overNode, err := ei.Process(args["over"])
	if err != nil {
		return nil, err
	}

	templateNode := args["template"]
	varName := "item"
	if asNode, ok := args["as"]; ok {
		if asNode.Kind != yaml.ScalarNode {
			return nil, errors.New("!Loop 'as' argument must be a scalar")
		}
		varName = asNode.Value
	}

	var loopOutput []*yaml.Node

	// Handle sequence and mapping nodes separately
	if overNode.Kind == yaml.SequenceNode {
		for i := 0; i < len(overNode.Content); i++ {
			itemNode := overNode.Content[i]
			result, err := processItem(ei, itemNode, templateNode, varName)
			if err != nil {
				return nil, err
			}
			loopOutput = append(loopOutput, result)
		}
		return &yaml.Node{
			Kind:    yaml.SequenceNode,
			Tag:     "!!seq",
			Content: loopOutput,
		}, nil
	} else if overNode.Kind == yaml.MappingNode {
		for i := 0; i < len(overNode.Content); i += 2 {
			keyNode := overNode.Content[i]
			itemNode := overNode.Content[i+1]
			result, err := processItem(ei, itemNode, templateNode, varName)
			if err != nil {
				return nil, err
			}
			loopOutput = append(loopOutput, keyNode, result)
		}
		return &yaml.Node{
			Kind:    yaml.MappingNode,
			Tag:     "!!map",
			Content: loopOutput,
		}, nil

	} else {
		return nil, errors.New("!Loop 'over' must be a sequence or mapping node")
	}
}

func processItem(ei *EmrichenInterpreter, itemNode *yaml.Node, templateNode *yaml.Node, varName string) (*yaml.Node, error) {
	v, ok := NodeToInterface(itemNode)
	if !ok {
		return nil, errors.Errorf("could not get value for node: %v", itemNode)
	}
	ei.env.Push(map[string]interface{}{
		varName: v,
	})
	result, err := ei.Process(templateNode)
	ei.env.Pop()
	if err != nil {
		return nil, err
	}
	return result, nil
}
