package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
)

func (ei *EmrichenInterpreter) handleInclude(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!Include requires a scalar value (the file path)")
	}

	filePath := node.Value

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file for !Include: %v", err)
	}

	includedNode := &yaml.Node{}
	if err := yaml.Unmarshal(fileContent, includedNode); err != nil {
		return nil, fmt.Errorf("error unmarshalling included file: %v", err)
	}

	return ei.Process(includedNode)
}
