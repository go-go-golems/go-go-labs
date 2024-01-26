package main

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (ei *EmrichenInterpreter) handleWith(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("!With requires a mapping node")
	}

	varsNode, templateNode := findWithNodes(node.Content)
	if varsNode == nil || templateNode == nil {
		return nil, errors.New("!With requires 'vars' and 'template' nodes")
	}

	if varsNode.Kind != yaml.MappingNode {
		return nil, errors.New("!With 'vars' node must be a mapping node")
	}

	if err := ei.updateVars(varsNode.Content); err != nil {
		return nil, err
	}

	defer ei.env.Pop()

	return ei.Process(templateNode)
}
