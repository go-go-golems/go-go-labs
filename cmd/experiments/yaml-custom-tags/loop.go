package main

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (ei *EmrichenInterpreter) handleLoop(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("!Loop requires a mapping node")
	}

	args, err := parseArgs(node, []string{"over", "template"})
	if err != nil {
		return nil, err
	}

	overNode := args["over"]
	templateNode := args["template"]

	// TODO: process templateNode

	if overNode.Kind != yaml.SequenceNode && overNode.Kind != yaml.MappingNode {
		return nil, errors.New("!Loop 'over' argument must be a sequence or mapping")
	}

	varName := "item"
	if asNode, ok := args["as"]; ok {
		if asNode.Kind != yaml.ScalarNode {
			return nil, errors.New("!Loop 'as' argument must be a scalar")
		}
		varName = asNode.Value
	}

	var loopOutput []*yaml.Node
	for i := 0; i < len(overNode.Content); i++ {
		itemNode := overNode.Content[i]

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
		if result != nil {
			loopOutput = append(loopOutput, result)
		}
	}

	resultNode := &yaml.Node{
		Kind:    yaml.SequenceNode,
		Tag:     "!!seq",
		Content: loopOutput,
	}

	return resultNode, nil
}
