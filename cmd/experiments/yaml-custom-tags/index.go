package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
)

func (ei *EmrichenInterpreter) handleIndex(node *yaml.Node) (*yaml.Node, error) {
	args, err := parseArgs(node, []string{"over", "by"})
	if err != nil {
		return nil, err
	}

	overNode, byNode := args["over"], args["by"]
	templateNode, templateExists := args["template"]
	duplicatesActionNode, duplicatesExists := args["duplicates"]
	duplicateAction := "error"
	if duplicatesExists {
		duplicateAction = duplicatesActionNode.Value
	}

	if overNode.Kind != yaml.SequenceNode {
		return nil, errors.New("!Index 'over' argument must be a sequence")
	}

	var resultVarName string
	if asNode, ok := args["as"]; ok && asNode.Kind == yaml.ScalarNode {
		resultVarName = asNode.Value
	} else {
		resultVarName = "item" // Default variable name
	}

	indexedResults := make(map[interface{}]*yaml.Node)
	duplicateKeys := make(map[interface{}]bool)

	for _, itemNode := range overNode.Content {
		v_, err := ei.Process(itemNode)
		if err != nil {
			return nil, err
		}
		v__, ok := NodeToInterface(v_)
		if !ok {
			return nil, errors.Errorf("could not get item for node: %v", itemNode)
		}

		err = ei.env.With(map[string]interface{}{resultVarName: v__}, func() error {
			keyNode, err := ei.Process(byNode)
			if err != nil {
				return err
			}
			if keyNode.Kind != yaml.ScalarNode {
				return errors.New("!Index 'by' expression must evaluate to a scalar")
			}
			key := keyNode.Value

			_, isDuplicate := duplicateKeys[key]
			if isDuplicate {
				switch duplicateAction {
				case "error":
					return errors.Errorf("Duplicate key encountered: %v", key)
				case "warn", "warning":
					_, _ = fmt.Fprintf(os.Stderr, "WARNING: Duplicate key encountered: %v\n", key)
					return nil
				case "ignore":
				default:
					return errors.Errorf("Unknown duplicate action: %v", duplicateAction)
				}
			}
			duplicateKeys[key] = true

			var resultNode *yaml.Node
			if templateExists {
				resultNode, err = ei.Process(templateNode)
				if err != nil {
					return err
				}
			} else {
				resultNode = itemNode
			}

			indexedResults[key] = resultNode

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// Convert the map to a sequence of YAML nodes
	content := make([]*yaml.Node, 0, len(indexedResults)*2)
	for k, v := range indexedResults {
		keyNode, err := ValueToNode(k)
		if err != nil {
			return nil, err
		}
		content = append(content, keyNode, v)
	}

	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Tag:     "!!map",
		Content: content,
	}, nil
}
