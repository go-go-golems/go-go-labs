package main

import (
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
)

func (ei *EmrichenInterpreter) handleInclude(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!Include requires a scalar value (the file path)")
	}

	filePath := node.Value
	decodedNodes, err := ei.loadYaml(filePath)
	if err != nil {
		return nil, err
	}
	if len(decodedNodes) == 0 {
		return nil, nil
	}
	if len(decodedNodes) == 1 {
		return decodedNodes[0], nil
	}
	return &yaml.Node{
		Kind:    yaml.SequenceNode,
		Tag:     "!!seq",
		Content: decodedNodes,
	}, nil
}

func (ei *EmrichenInterpreter) loadYaml(filePath string) ([]*yaml.Node, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file for !Include: %v", err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	decoder := yaml.NewDecoder(f)

	decodedNodes := make([]*yaml.Node, 0)
	for {
		includedNode := &yaml.Node{}

		err = decoder.Decode(ei.CreateRawDecoder(includedNode))
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		decodedNodes = append(decodedNodes, includedNode)
	}

	return decodedNodes, nil
}

func (ei *EmrichenInterpreter) handleIncludeBase64(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!IncludeBase64 requires a scalar value (the file path)")
	}

	filePath := node.Value
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file for !IncludeBase64: %v", err)
	}

	encodedContent := base64.StdEncoding.EncodeToString(fileContent)
	return makeString(encodedContent), nil
}

func (ei *EmrichenInterpreter) handleIncludeBinary(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!IncludeBinary requires a scalar value (the file path)")
	}

	filePath := node.Value
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file for !IncludeBinary: %v", err)
	}

	// The binary data needs to be properly handled as per your use case
	return makeString(string(fileContent)), nil
}

func (ei *EmrichenInterpreter) handleIncludeGlob(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!IncludeGlob requires a scalar value (the glob pattern)")
	}

	patterns := []string{node.Value}
	if node.Kind == yaml.SequenceNode {
		patterns = make([]string, len(node.Content))
		for i, n := range node.Content {
			if n.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("invalid glob pattern: %v", n.Value)
			}
			patterns[i] = n.Value
		}
	}

	var nodes []*yaml.Node
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("error in globbing pattern: %v", err)
		}
		for _, match := range matches {
			includedNodes, err := ei.loadYaml(match)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, includedNodes...)
		}
	}

	return &yaml.Node{
		Kind:    yaml.SequenceNode,
		Tag:     "!!seq",
		Content: nodes,
	}, nil
}

func (ei *EmrichenInterpreter) handleIncludeText(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!IncludeText requires a scalar value (the file path)")
	}

	filePath := node.Value
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file for !IncludeText: %v", err)
	}

	return makeString(string(fileContent)), nil
}
