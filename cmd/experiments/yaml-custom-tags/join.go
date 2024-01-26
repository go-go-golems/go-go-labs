package main

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"strings"
)

func (ei *EmrichenInterpreter) handleJoin(node *yaml.Node) (*yaml.Node, error) {
	separator := " " // Default separator
	itemsNode := node

	if node.Kind == yaml.MappingNode {
		args, err := parseArgs(node, []string{"items"})
		if err != nil {
			return nil, err
		}

		var ok bool
		itemsNode, ok = args["items"]
		if !ok || itemsNode.Kind != yaml.SequenceNode {
			return nil, errors.New("!Join requires a sequence node for 'items'")
		}

		if sepNode, ok := args["separator"]; ok && sepNode.Kind == yaml.ScalarNode {
			separator = sepNode.Value
		}
	}

	var items []string
	for _, itemNode := range itemsNode.Content {
		resolvedItem, err := ei.Process(itemNode)
		if err != nil {
			return nil, err
		}
		if resolvedItem.Kind != yaml.ScalarNode {
			return nil, errors.New("!Join items must be scalar values")
		}
		if resolvedItem.Tag == "!!null" {
			continue
		}
		items = append(items, resolvedItem.Value)
	}

	joinedStr := strings.Join(items, separator)

	return ValueToNode(joinedStr)
}
