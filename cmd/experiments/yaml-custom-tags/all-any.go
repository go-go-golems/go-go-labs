package main

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (ei *EmrichenInterpreter) handleAll(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, errors.New("!All requires a sequence node")
	}

	for _, item := range node.Content {
		resolvedItem, err := ei.Process(item)
		if err != nil {
			return nil, err
		}
		if !isTruthy(resolvedItem) {
			return makeBool(false), nil
		}
	}
	return makeBool(true), nil
}

func (ei *EmrichenInterpreter) handleAny(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, errors.New("!Any requires a sequence node")
	}

	for _, item := range node.Content {
		resolvedItem, err := ei.Process(item)
		if err != nil {
			return nil, err
		}
		if isTruthy(resolvedItem) {
			return makeBool(true), nil
		}
	}
	return makeBool(false), nil
}
