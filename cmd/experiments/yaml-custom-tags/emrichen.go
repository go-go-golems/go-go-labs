package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

type EmrichenInterpreter struct {
	vars map[string]*yaml.Node
}

func NewEmrichenInterpreter() *EmrichenInterpreter {
	return &EmrichenInterpreter{
		vars: make(map[string]*yaml.Node),
	}
}

type interpretHelper struct {
	target      interface{}
	interpreter *EmrichenInterpreter
}

func (ei *interpretHelper) UnmarshalYAML(value *yaml.Node) error {
	resolved, err := ei.interpreter.Process(value)
	if err != nil {
		return err
	}
	return resolved.Decode(ei.target)
}

func (ei *EmrichenInterpreter) CreateDecoder(target interface{}) *interpretHelper {
	return &interpretHelper{
		target:      target,
		interpreter: ei,
	}
}

func (ei *EmrichenInterpreter) Process(node *yaml.Node) (*yaml.Node, error) {
	tag := node.Tag
	ss := strings.Split(tag, ",")
	if len(ss) == 0 {
		return nil, errors.New("custom tag is empty")
	}

	switch ss[0] {
	case "!Defaults":
		if node.Kind == yaml.MappingNode {
			err := ei.updateVars(node.Content)
			if err != nil {
				return nil, err
			}
		}
		return node, nil

	case "!MD5":
		if node.Kind != yaml.ScalarNode {
			return nil, errors.New("!MD5 requires a scalar value")
		}
		hash := md5.Sum([]byte(node.Value))
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: hex.EncodeToString(hash[:]),
		}, nil

	case "!SHA1":
		if node.Kind != yaml.ScalarNode {
			return nil, errors.New("!SHA1 requires a scalar value")
		}
		hash := sha1.Sum([]byte(node.Value))
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: hex.EncodeToString(hash[:]),
		}, nil

	case "!SHA256":
		if node.Kind != yaml.ScalarNode {
			return nil, errors.New("!SHA256 requires a scalar value")
		}
		hash := sha256.Sum256([]byte(node.Value))
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: hex.EncodeToString(hash[:]),
		}, nil

	case "!Var":
		if node.Kind == yaml.ScalarNode {
			varName := node.Value
			varValue, ok := ei.vars[varName]
			if !ok {
				return nil, fmt.Errorf("variable %s not found", varName)
			}
			return varValue, nil
		}
		return nil, errors.New("variable definition must be !Var variable name")

	case "!Error":
		if node.Kind == yaml.ScalarNode {
			return nil, errors.New(node.Value)
		}
		return nil, errors.New("!Error tag requires a scalar value for the error message")

	case "!Format":
		return ei.handleFormat(node)

	case "!Include":
		return ei.handleInclude(node)

	case "!IsBoolean":
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: strconv.FormatBool(node.Kind == yaml.ScalarNode && (node.Value == "true" || node.Value == "false")),
		}, nil

	case "!IsDict":
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: strconv.FormatBool(node.Kind == yaml.MappingNode),
		}, nil

	case "!IsInteger":
		_, err := strconv.Atoi(node.Value)
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: strconv.FormatBool(err == nil && node.Kind == yaml.ScalarNode),
		}, nil

	case "!IsList":
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: strconv.FormatBool(node.Kind == yaml.SequenceNode),
		}, nil

	case "!IsNone":
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: strconv.FormatBool(node.Tag == "!!null" || node.Value == "null"),
		}, nil

	case "!IsNumber":
		_, err := strconv.ParseFloat(node.Value, 64)
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: strconv.FormatBool(err == nil && node.Kind == yaml.ScalarNode),
		}, nil

	case "!IsString":
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: strconv.FormatBool(node.Kind == yaml.ScalarNode),
		}, nil

	case "!Merge":
		return ei.handleMerge(node)

	case "!URLEncode":
		return ei.handleURLEncode(node)

	case "!Void":
		return nil, nil

	case "!With":
		return ei.handleWith(node)

	default:
	}

	if node.Kind == yaml.SequenceNode || node.Kind == yaml.MappingNode {
		var err error
		for i := range node.Content {
			node.Content[i], err = ei.Process(node.Content[i])
			if err != nil {
				return nil, err
			}
		}
	}

	return node, nil
}

func (ei *EmrichenInterpreter) updateVars(content []*yaml.Node) error {
	var err error
	name := ""
	for i := range content {
		if i%2 == 0 {
			name = content[i].Value
			continue
		}
		ei.vars[name], err = ei.Process(content[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *EmrichenInterpreter) handleFormat(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.SequenceNode || len(node.Content) < 1 {
		return nil, errors.New("!Format requires at least one argument")
	}

	formatStringNode := node.Content[0]
	if formatStringNode.Kind != yaml.ScalarNode {
		return nil, errors.New("!Format first argument must be a scalar (the format string)")
	}

	var args []interface{}
	for _, argNode := range node.Content[1:] {
		resolvedArg, err := ei.Process(argNode)
		if err != nil {
			return nil, err
		}
		if resolvedArg.Kind != yaml.ScalarNode {
			return nil, errors.New("!Format arguments must be scalar values")
		}
		args = append(args, resolvedArg.Value)
	}

	tmpl, err := template.New("format").Parse(formatStringNode.Value)
	if err != nil {
		return nil, fmt.Errorf("error parsing format string: %v", err)
	}

	var formatted bytes.Buffer
	if err := tmpl.Execute(&formatted, args); err != nil {
		return nil, fmt.Errorf("error executing format template: %v", err)
	}

	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: formatted.String(),
	}, nil
}

func (ei *EmrichenInterpreter) handleInclude(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.ScalarNode {
		return nil, errors.New("!Include requires a scalar value (the file path)")
	}

	filePath := node.Value
	if !filepath.IsAbs(filePath) {
		// Handle relative path if necessary
		// filePath = filepath.Join(basePath, filePath)
	}

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

func (ei *EmrichenInterpreter) handleMerge(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, errors.New("!Merge requires a sequence of mapping nodes")
	}

	mergedMap := make(map[string]*yaml.Node)
	for _, item := range node.Content {
		if item.Kind != yaml.MappingNode {
			return nil, errors.New("!Merge items must be mapping nodes")
		}

		tempMap := make(map[string]*yaml.Node)
		for i := 0; i < len(item.Content); i += 2 {
			keyNode := item.Content[i]
			valueNode := item.Content[i+1]

			resolvedValue, err := ei.Process(valueNode)
			if err != nil {
				return nil, err
			}

			tempMap[keyNode.Value] = resolvedValue
		}

		for k, v := range tempMap {
			mergedMap[k] = v
		}
	}

	mergedContent := []*yaml.Node{}
	for k, v := range mergedMap {
		mergedContent = append(mergedContent, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: k,
		}, v)
	}

	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: mergedContent,
	}, nil

}

func (ei *EmrichenInterpreter) handleWith(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("!With requires a mapping node")
	}

	varsNode, templateNode := findWithNodes(node.Content)
	if varsNode == nil || templateNode == nil {
		return nil, errors.New("!With requires 'vars' and 'template' nodes")
	}

	currentVars := ei.vars

	newVars := make(map[string]*yaml.Node)
	for k, v := range currentVars {
		newVars[k] = v
	}

	ei.vars = newVars
	defer func() { ei.vars = currentVars }()

	if err := ei.updateVars(varsNode.Content); err != nil {
		return nil, err
	}

	return ei.Process(templateNode)
}

// Helper function to find 'vars' and 'template' nodes in !With
func findWithNodes(content []*yaml.Node) (*yaml.Node, *yaml.Node) {
	var varsNode *yaml.Node
	var templateNode *yaml.Node
	for i := 0; i < len(content); i += 2 {
		keyNode := content[i]
		valueNode := content[i+1]
		if keyNode.Kind == yaml.ScalarNode {
			if keyNode.Value == "vars" {
				varsNode = valueNode
			} else if keyNode.Value == "template" {
				templateNode = valueNode
			}
		}
	}
	return varsNode, templateNode
}

func (ei *EmrichenInterpreter) handleURLEncode(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind == yaml.ScalarNode {
		// Simple string encoding
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: url.QueryEscape(node.Value),
		}, nil
	} else if node.Kind == yaml.MappingNode {
		// Construct URL with query parameters
		urlStr, queryParams, err := parseURLEncodeArgs(node.Content)
		if err != nil {
			return nil, err
		}

		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing URL in !URLEncode: %v", err)
		}

		query := parsedURL.Query()
		for k, v := range queryParams {
			query.Set(k, v)
		}
		parsedURL.RawQuery = query.Encode()

		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: parsedURL.String(),
		}, nil
	}

	return nil, errors.New("!URLEncode requires a scalar or mapping node")

}

// Helper function to parse URL and query parameters from node content
func parseURLEncodeArgs(content []*yaml.Node) (
	urlStr string,
	queryParams map[string]string,
	err error
) {
	queryParams = make(map[string]string)

	for i := 0; i < len(content); i += 2 {
		keyNode := content[i]
		valueNode := content[i+1]
		if keyNode.Kind == yaml.ScalarNode {
			if keyNode.Value == "url" {
				urlStr = valueNode.Value
			} else if keyNode.Value == "query" && valueNode.Kind == yaml.MappingNode {
				for j := 0; j < len(valueNode.Content); j += 2 {
					paramKey := valueNode.Content[j].Value
					paramValue := valueNode.Content[j+1].Value
					queryParams[paramKey] = paramValue
				}
			}
		}
	}

	if urlStr == "" {
		err = errors.New("URL string is required for !URLEncode with query parameters")
	}

}

