package main

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (ei *EmrichenInterpreter) handleFilter(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("!Filter requires a mapping node")
	}

	args, err := ei.parseArgs(node, []parsedVariable{
		{Name: "over", Required: true, Expand: true},
		{Name: "test"},
		{Name: "as"},
	})
	if err != nil {
		return nil, err
	}

	overNode := args["over"]
	if overNode.Kind != yaml.SequenceNode && overNode.Kind != yaml.MappingNode {
		return nil, errors.New("!Filter 'over' argument must be a sequence or mapping")
	}
	testNode, hasTestNode := args["test"]

	varName := "item"
	asNode, ok := args["as"]
	if ok {
		if asNode.Kind != yaml.ScalarNode {
			return nil, errors.New("!Filter 'as' argument must be a scalar")
		}
		varName = asNode.Value
	}

	var filtered []*yaml.Node
	var filteredMap bool

	if overNode.Kind == yaml.MappingNode {
		filteredMap = true
	}

	for i := 0; i < len(overNode.Content); i++ {
		var item *yaml.Node
		var key *yaml.Node

		if filteredMap {
			if i%2 != 0 { // Skip value nodes
				continue
			}
			key = overNode.Content[i]
			item = overNode.Content[i+1]
		} else {
			item = overNode.Content[i]
		}

		var result *yaml.Node
		if hasTestNode {
			v, ok := NodeToInterface(item)
			if !ok {
				return nil, errors.Errorf("could not get value for node: %v", item)
			}
			ei.env.Push(map[string]interface{}{
				varName: v,
			})
			result, err = ei.Process(testNode)
			ei.env.Pop()
			if err != nil {
				return nil, err
			}
		} else {
			result = item
		}

		if isTruthy(result) {
			if filteredMap {
				filtered = append(filtered, key)
			}
			filtered = append(filtered, item)
		}
	}

	var resultNode *yaml.Node
	if filteredMap {
		resultNode = &yaml.Node{
			Kind:    yaml.MappingNode,
			Tag:     "!!map",
			Content: filtered,
		}
	} else {
		resultNode = &yaml.Node{
			Kind:    yaml.SequenceNode,
			Tag:     "!!seq",
			Content: filtered,
		}
	}

	return resultNode, nil
}
