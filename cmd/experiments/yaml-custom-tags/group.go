package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (ei *EmrichenInterpreter) handleGroup(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("!Group requires a mapping node")
	}

	args, err := parseArgs(node, []string{"over", "by"})
	if err != nil {
		return nil, err
	}

	overNode := args["over"]
	if overNode.Kind != yaml.SequenceNode && overNode.Kind != yaml.MappingNode {
		return nil, errors.New("!Group 'over' argument must be a sequence or mapping")
	}

	byNode, ok := args["by"]
	if !ok {
		return nil, errors.New("!Group requires a 'by' argument")
	}

	templateNode := args["template"]

	varName := "item"
	asNode, ok := args["as"]
	if ok {
		if asNode.Kind != yaml.ScalarNode {
			return nil, errors.New("!Group 'as' argument must be a scalar")
		}
		varName = asNode.Value
	}

	groups := make(map[interface{}][]*yaml.Node)
	var groupByMapping bool

	if overNode.Kind == yaml.MappingNode {
		groupByMapping = true
	}

	for i := 0; i < len(overNode.Content); i++ {
		_, err := func() (interface{}, error) {
			var itemNode *yaml.Node

			if groupByMapping {
				if i%2 != 0 { // Skip value nodes
					return nil, nil
				}
				itemNode = overNode.Content[i+1]
			} else {
				itemNode = overNode.Content[i]
			}

			item, ok := NodeToInterface(itemNode)
			if !ok {
				return nil, errors.Errorf("could not get item for node: %v", itemNode)
			}

			ei.env.Push(map[string]interface{}{
				varName: item,
			})
			defer ei.env.Pop()
			groupKeyNode, err := ei.Process(byNode)
			if err != nil {
				return nil, err
			}

			groupKey, ok := NodeToInterface(groupKeyNode)
			if !ok {
				return nil, errors.Errorf("could not get group key for node: %v", groupKeyNode)
			}

			var result *yaml.Node
			if templateNode != nil {
				result, err = ei.Process(templateNode)
				if err != nil {
					return nil, err
				}
			} else {
				result = itemNode
			}

			groups[groupKey] = append(groups[groupKey], result)
			return nil, nil
		}()
		if err != nil {
			return nil, err
		}
	}

	resultContent := make([]*yaml.Node, 0, len(groups)*2)
	for k, v := range groups {
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: fmt.Sprintf("%v", k),
		}
		valueNode := &yaml.Node{
			Kind:    yaml.SequenceNode,
			Tag:     "!!seq",
			Content: v,
		}
		resultContent = append(resultContent, keyNode, valueNode)
	}

	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Tag:     "!!map",
		Content: resultContent,
	}, nil
}
