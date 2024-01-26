package main

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (ei *EmrichenInterpreter) handleConcat(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, errors.New("!Concat requires a sequence node")
	}

	concatenated := []*yaml.Node{}
	for _, listItem := range node.Content {
		resolvedListItem, err := ei.Process(listItem)
		if err != nil {
			return nil, err
		}
		if resolvedListItem.Kind != yaml.SequenceNode {
			return nil, errors.New("!Concat items must be sequences")
		}
		concatenated = append(concatenated, resolvedListItem.Content...)
	}

	return &yaml.Node{
		Kind:    yaml.SequenceNode,
		Tag:     "!!seq",
		Content: concatenated,
	}, nil
}
