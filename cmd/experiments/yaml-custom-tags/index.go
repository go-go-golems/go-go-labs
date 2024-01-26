package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
)

func (ei *EmrichenInterpreter) handleIndex(node *yaml.Node) (*yaml.Node, error) {
	args, err := ei.parseArgs(node, []parsedVariable{
		{Name: "over", Required: true, Expand: true},
		{Name: "by", Required: true},
		{Name: "template"},
		{Name: "as"},
		{Name: "duplicates"},
		{Name: "result_as"},
	})
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

	var asVarName string
	if asNode, ok := args["as"]; ok && asNode.Kind == yaml.ScalarNode {
		asVarName = asNode.Value
	} else {
		asVarName = "item" // Default variable name
	}
	var resultVarName string
	if resultAsNode, ok := args["result_as"]; ok && resultAsNode.Kind == yaml.ScalarNode {
		resultVarName = resultAsNode.Value
	} else {
		resultVarName = "" // Default variable name
	}

	indexedResults := make(map[string]*yaml.Node)
	duplicateKeys := make(map[string]bool)

	for _, itemNode := range overNode.Content {
		v__, ok := NodeToInterface(itemNode)
		if !ok {
			return nil, errors.Errorf("could not get item for node: %v", itemNode)
		}

		err = ei.env.With(map[string]interface{}{asVarName: v__}, func() error {
			var resultNode *yaml.Node
			if templateExists {
				resultNode, err = ei.Process(templateNode)
				if err != nil {
					return err
				}
			} else {
				resultNode = itemNode
			}

			// Expand byNode
			byEnv := map[string]interface{}{}
			if resultVarName != "" {
				v, ok := NodeToInterface(resultNode)
				if !ok {
					return errors.Errorf("could not get result for node: %v", resultNode)
				}
				byEnv[resultVarName] = v
			}
			var processedByNode *yaml.Node
			err := ei.env.With(byEnv, func() error {
				processedByNode, err = ei.Process(byNode)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}
			if processedByNode.Kind != yaml.ScalarNode {
				return errors.New("!Index 'by' expression must evaluate to a scalar")
			}
			by := processedByNode.Value
			_, isDuplicate := duplicateKeys[by]
			if isDuplicate {
				switch duplicateAction {
				case "error":
					return errors.Errorf("Duplicate key encountered: %v", by)
				case "warn", "warning":
					_, _ = fmt.Fprintf(os.Stderr, "WARNING: Duplicate key encountered: %v\n", by)
					return nil
				case "ignore":
				default:
					return errors.Errorf("Unknown duplicate action: %v", duplicateAction)
				}
			}
			duplicateKeys[by] = true

			indexedResults[by] = resultNode

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
