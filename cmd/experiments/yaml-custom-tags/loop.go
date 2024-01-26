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
		{Name: "index_as"},
		{Name: "previous_as"},
		{Name: "index_start", Expand: true},
		{Name: "as_documents"},
	})
	if err != nil {
		return nil, err
	}

	if _, ok := args["as_documents"]; ok {
		return nil, errors.New("!Loop 'as_documents' argument is not supported yet")
	}

	overNode := args["over"]

	templateNode := args["template"]

	asVarName := "item"
	if asNode, ok := args["as"]; ok {
		if asNode.Kind != yaml.ScalarNode {
			return nil, errors.New("!Loop 'as' argument must be a scalar")
		}
		asVarName = asNode.Value
	}

	indexAsVarName := ""
	if indexNode, ok := args["index_as"]; ok {
		if indexNode.Kind != yaml.ScalarNode {
			return nil, errors.New("!Loop 'index_as' argument must be a scalar")
		}
		indexAsVarName = indexNode.Value
	}

	previousAsVarName := ""
	if previousNode, ok := args["previous_as"]; ok {
		if previousNode.Kind != yaml.ScalarNode {
			return nil, errors.New("!Loop 'previous_as' argument must be a scalar")
		}
		previousAsVarName = previousNode.Value
	}
	indexStart := 0
	if indexStartNode, ok := args["index_start"]; ok {
		v, ok := NodeToInt(indexStartNode)
		if !ok {
			return nil, errors.Errorf("could not get value for node: %v", indexStartNode)
		}
		indexStart = v
	}

	var loopOutput []*yaml.Node

	previousNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}

	// Handle sequence and mapping nodes separately
	if overNode.Kind == yaml.SequenceNode {
		for i, itemNode := range overNode.Content {
			if i < indexStart {
				continue
			}
			v, ok := NodeToInterface(itemNode)
			if !ok {
				return nil, errors.Errorf("could not get value for node: %v", itemNode)
			}
			templateEnv := map[string]interface{}{
				asVarName: v,
			}
			if indexAsVarName != "" {
				templateEnv[indexAsVarName] = i
			}
			if previousAsVarName != "" {
				v, ok := NodeToInterface(previousNode)
				if !ok {
					return nil, errors.Errorf("could not get value for node: %v", previousNode)
				}
				templateEnv[previousAsVarName] = v
			}
			var resultNode *yaml.Node
			err = ei.env.With(templateEnv, func() error {
				resultNode, err = ei.Process(templateNode)
				return err
			})
			if err != nil {
				return nil, err
			}
			if resultNode == nil {
				continue
			}
			loopOutput = append(loopOutput, resultNode)
			previousNode = itemNode
		}
		return &yaml.Node{
			Kind:    yaml.SequenceNode,
			Tag:     "!!seq",
			Content: loopOutput,
		}, nil
	} else if overNode.Kind == yaml.MappingNode {
		for i := 0; i < len(overNode.Content); i += 2 {
			if i/2 < indexStart {
				continue
			}
			keyNode := overNode.Content[i]
			itemNode := overNode.Content[i+1]

			v, ok := NodeToInterface(itemNode)
			if !ok {
				return nil, errors.Errorf("could not get value for node: %v", itemNode)
			}
			templateEnv := map[string]interface{}{
				asVarName: v,
			}
			if indexAsVarName != "" {
				templateEnv[indexAsVarName] = keyNode.Value
			}
			if previousAsVarName != "" {
				v, ok := NodeToInterface(previousNode)
				if !ok {
					return nil, errors.Errorf("could not get value for node: %v", previousNode)
				}
				templateEnv[previousAsVarName] = v
			}
			var resultNode *yaml.Node
			err = ei.env.With(templateEnv, func() error {
				resultNode, err = ei.Process(templateNode)
				return err
			})
			if err != nil {
				return nil, err
			}
			if resultNode == nil {
				continue
			}
			loopOutput = append(loopOutput, keyNode, resultNode)
			previousNode = itemNode
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
