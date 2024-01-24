package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/PaesslerAG/jsonpath"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

type EmrichenInterpreter struct {
	vars map[string]interface{}
}

type EmrichenInterpreterOption func(*EmrichenInterpreter) error

func WithVars(vars map[string]interface{}) EmrichenInterpreterOption {
	return func(ei *EmrichenInterpreter) error {
		for k, v := range vars {
			value, err := cast.ToInterfaceValue(v)
			if err != nil {
				return err
			}
			ei.vars[k] = value
		}
		return nil
	}
}

func NewEmrichenInterpreter(options ...EmrichenInterpreterOption) (*EmrichenInterpreter, error) {
	ret := &EmrichenInterpreter{
		vars: make(map[string]interface{}),
	}

	for _, option := range options {
		err := option(ret)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
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

func (ei *EmrichenInterpreter) Lookup(jsonPath string) (interface{}, error) {
	if !strings.HasPrefix(jsonPath, "$.") {
		jsonPath = "$." + jsonPath
	}
	v, err := jsonpath.Get(jsonPath, ei.vars)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (ei *EmrichenInterpreter) Process(node *yaml.Node) (*yaml.Node, error) {
	tag := node.Tag
	ss := strings.Split(tag, ",")
	if len(ss) == 0 {
		return nil, errors.New("custom tag is empty")
	}

	for i, s := range ss[1:] {
		if !strings.HasPrefix(s, "!") {
			ss[i+1] = "!" + s
		}
	}

	// reverse ss
	for i := len(ss)/2 - 1; i >= 0; i-- {
		opp := len(ss) - 1 - i
		ss[i], ss[opp] = ss[opp], ss[i]
	}

	for _, verb := range ss {
		ret, err := func() (*yaml.Node, error) {
			//exhaustive:ignore
			switch verb {
			case "!Defaults":
				if node.Kind == yaml.MappingNode {
					err := ei.updateVars(node.Content)
					if err != nil {
						return nil, err
					}
				}
				return node, nil

			case "!All":
				return ei.handleAll(node)
			case "!Any":

				return ei.handleAny(node)

			case "!Filter":
				return ei.handleFilter(node)

			case "!If":
				return ei.handleIf(node)

			case "!Exists":
				return ei.handleExists(node)

			case "!Lookup":
				return ei.handleLookup(node)

			case "!LookupAll":
				return ei.handleLookup(node)

			case "!Concat":
				return ei.handleConcat(node)

			case "!Index":
				return ei.handleIndex(node)

			case "!Join":
				return ei.handleJoin(node)

			case "!Not":
				return ei.handleNot(node)

			case "!Op":
				return ei.handleOp(node)

			case "!MD5":
				if node.Kind != yaml.ScalarNode {
					return nil, errors.New("!MD5 requires a scalar value")
				}
				hash := md5.Sum([]byte(node.Value))
				return makeString(hex.EncodeToString(hash[:])), nil

			case "!SHA1":
				if node.Kind != yaml.ScalarNode {
					return nil, errors.New("!SHA1 requires a scalar value")
				}
				hash := sha1.Sum([]byte(node.Value))
				return makeString(hex.EncodeToString(hash[:])), nil

			case "!SHA256":
				if node.Kind != yaml.ScalarNode {
					return nil, errors.New("!SHA256 requires a scalar value")
				}
				hash := sha256.Sum256([]byte(node.Value))
				return makeString(hex.EncodeToString(hash[:])), nil

			case "!Base64":
				if node.Kind != yaml.ScalarNode {
					return nil, errors.New("!Base64 requires a scalar value")
				}
				return makeString(base64.StdEncoding.EncodeToString([]byte(node.Value))), nil

			case "!Var":
				if node.Kind == yaml.ScalarNode {
					varName := node.Value
					varValue, ok := ei.vars[varName]
					if !ok {
						return nil, fmt.Errorf("variable %s not found", varName)
					}
					v, err := makeValue(varValue)
					if err != nil {
						return nil, err
					}
					return v, nil
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
				return makeBool(node.Kind == yaml.ScalarNode && (node.Value == "true" || node.Value == "false")), nil

			case "!IsDict":
				return makeBool(node.Kind == yaml.MappingNode), nil

			case "!IsInteger":
				_, err := strconv.Atoi(node.Value)
				return makeBool(err == nil && node.Kind == yaml.ScalarNode), nil

			case "!IsList":
				return makeBool(node.Kind == yaml.SequenceNode), nil

			case "!IsNone":
				return makeBool(node.Tag == "!!null" || node.Value == "null"), nil

			case "!IsNumber":
				_, err := strconv.ParseFloat(node.Value, 64)
				return makeBool(err == nil && node.Kind == yaml.ScalarNode), nil

			case "!IsString":
				return makeBool(node.Kind == yaml.ScalarNode), nil

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
		}()

		if err != nil {
			return nil, err
		}

		node = ret
	}

	return node, nil
}

func (ei *EmrichenInterpreter) updateVars(content []*yaml.Node) error {
	name := ""
	for i := range content {
		if i%2 == 0 {
			name = content[i].Value
			continue
		}
		node, err := ei.Process(content[i])
		if err != nil {
			return err
		}
		v, ok := GetValue(node)
		if !ok {
			return errors.New("could not get node value")
		}
		v_, err := cast.ToInterfaceValue(v)
		if err != nil {
			return err
		}
		ei.vars[name] = v_
	}

	return nil
}

func (ei *EmrichenInterpreter) handleConcat(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, errors.New("!Concat requires a sequence node")
	}

	concatenated := []*yaml.Node{}
	for _, listItem := range node.Content {
		resolvedListItem, err := ei.Process(listItem)
		if err != nil {
			return nil, err
		}
		if resolvedListItem.Kind != yaml.SequenceNode {
			return nil, errors.New("!Concat items must be sequences")
		}
		concatenated = append(concatenated, resolvedListItem.Content...)
	}

	return &yaml.Node{
		Kind:    yaml.SequenceNode,
		Tag:     "!!seq",
		Content: concatenated,
	}, nil
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

func (ei *EmrichenInterpreter) handleExists(node *yaml.Node) (*yaml.Node, error) {
	_, err := ei.Lookup(node.Value)
	if err != nil {
		if strings.Contains(err.Error(), "unknown key") {
			return makeBool(false), nil
		}
		return nil, err
	}
	return makeBool(true), nil
}

func (ei *EmrichenInterpreter) handleLookup(node *yaml.Node) (*yaml.Node, error) {
	// check that the value is a string
	v, err := ei.Lookup(node.Value)
	if err != nil {
		return nil, err
	}
	// check if query contains *, and if the result is a list, return the first element
	if strings.Contains(node.Value, "*") && reflect.TypeOf(v).Kind() == reflect.Slice {
		v = reflect.ValueOf(v).Index(0).Interface()
	}
	return makeValue(v)
}

func (ei *EmrichenInterpreter) handleLookupAll(node *yaml.Node) (*yaml.Node, error) {
	v, err := ei.Lookup(node.Value)
	if err != nil {
		return nil, err
	}
	return makeValue(v)
}

func (ei *EmrichenInterpreter) handleFilter(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("!Filter requires a mapping node")
	}

	args, err := parseArgs(node, []string{"test", "over"})
	if err != nil {
		return nil, err
	}

	overNode := args["over"]
	if overNode.Kind != yaml.SequenceNode {
		return nil, errors.New("!Filter 'over' argument must be a sequence")
	}
	testNode := args["test"]

	varName := "item"
	asNode, ok := args["as"]
	if ok {
		if asNode.Kind != yaml.ScalarNode {
			return nil, errors.New("!Filter 'as' argument must be a scalar")
		}
		varName = asNode.Value
	}

	filtered := []*yaml.Node{}
	for _, item := range overNode.Content {
		ei.vars[varName] = item // Assuming 'item' is the variable name for each element
		result, err := ei.Process(testNode)
		if err != nil {
			return nil, err
		}
		if isTruthy(result) {
			filtered = append(filtered, item)
		}
	}
	delete(ei.vars, "item") // Clean up the temporary variable

	return &yaml.Node{
		Kind:    yaml.SequenceNode,
		Content: filtered,
	}, nil
}

func (ei *EmrichenInterpreter) handleIf(node *yaml.Node) (*yaml.Node, error) {
	args, err := parseArgs(node, []string{"test", "then", "else"})
	if err != nil {
		return nil, err
	}

	testResult, err := ei.Process(args["test"])
	if err != nil {
		return nil, err
	}

	if isTruthy(testResult) {
		return ei.Process(args["then"])
	} else {
		return ei.Process(args["else"])
	}
}

func (ei *EmrichenInterpreter) handleJoin(node *yaml.Node) (*yaml.Node, error) {
	args, err := parseArgs(node, []string{"items"})
	if err != nil {
		return nil, err
	}

	itemsNode, ok := args["items"]
	if !ok || itemsNode.Kind != yaml.SequenceNode {
		return nil, errors.New("!Join requires a sequence node for 'items'")
	}

	separator := " " // Default separator
	if sepNode, ok := args["separator"]; ok && sepNode.Kind == yaml.ScalarNode {
		separator = sepNode.Value
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
		items = append(items, resolvedItem.Value)
	}

	joinedStr := strings.Join(items, separator)

	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: joinedStr,
	}, nil
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

func (ei *EmrichenInterpreter) handleNot(node *yaml.Node) (*yaml.Node, error) {
	if len(node.Content) != 1 {
		return nil, errors.New("!Not requires exactly one argument")
	}

	result, err := ei.Process(node.Content[0])
	if err != nil {
		return nil, err
	}

	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: strconv.FormatBool(!isTruthy(result)),
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

	newVars := make(map[string]interface{})
	for k, v := range currentVars {
		value, err := cast.ToInterfaceValue(v)
		if err != nil {
			return nil, err
		}
		newVars[k] = value
	}

	ei.vars = newVars
	defer func() { ei.vars = currentVars }()

	if err := ei.updateVars(varsNode.Content); err != nil {
		return nil, err
	}

	return ei.Process(templateNode)
}

func (ei *EmrichenInterpreter) handleURLEncode(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind == yaml.ScalarNode {
		// Simple string encoding
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: url.QueryEscape(node.Value),
		}, nil
	} else if node.Kind == yaml.MappingNode {
		urlStr, queryParams, err := parseURLEncodeArgs(node)
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

func (ei *EmrichenInterpreter) handleOp(node *yaml.Node) (*yaml.Node, error) {
	args, err := parseArgs(node, []string{"a", "op", "b"})
	if err != nil {
		return nil, err
	}

	opNode := args["op"]
	if opNode.Kind != yaml.ScalarNode {
		return nil, errors.New("!Op 'op' argument must be a scalar")
	}

	aNode, bNode := args["a"], args["b"]
	aProcessed, err := ei.Process(aNode)
	if err != nil {
		return nil, err
	}
	bProcessed, err := ei.Process(bNode)
	if err != nil {
		return nil, err
	}

	isNumberOperation := false
	switch opNode.Value {
	case "+", "plus", "add", "-", "minus", "sub", "subtract", "*", "×", "mul", "times", "/", "÷", "div", "divide", "truediv", "//", "floordiv":
		isNumberOperation = true
	default:
	}

	a, ok := GetFloat(aProcessed)
	if isNumberOperation && !ok {
		return nil, errors.New("could not convert first argument to float")
	}
	b, ok := GetFloat(bProcessed)
	if isNumberOperation && !ok {
		return nil, errors.New("could not convert second argument to float")
	}

	bothInts := aProcessed.Tag == "!!int" && bProcessed.Tag == "!!int"

	// Handle different operators
	switch opNode.Value {
	case "=", "==", "===":
		return makeBool(reflect.DeepEqual(aProcessed.Value, bProcessed.Value)), nil
	case "≠", "!=", "!==":
		return makeBool(!reflect.DeepEqual(aProcessed.Value, bProcessed.Value)), nil

	// Less than, Greater than, Less than or equal to, Greater than or equal to
	case "<", "lt":
		return makeBool(a < b), nil
	case ">", "gt":
		return makeBool(a > b), nil
	case "<=", "le", "lte":
		return makeBool(a <= b), nil
	case ">=", "ge", "gte":
		return makeBool(a >= b), nil

	// Arithmetic operations
	case "+", "plus", "add":
		if bothInts {
			return makeInt(int(a) + int(b)), nil
		}
		return makeFloat(a + b), nil

	case "-", "minus", "sub", "subtract":
		if bothInts {
			return makeInt(int(a) - int(b)), nil
		}
		return makeFloat(a - b), nil

	case "*", "×", "mul", "times":
		if bothInts {
			return makeInt(int(a) * int(b)), nil
		}
		return makeFloat(a * b), nil
	case "/", "÷", "div", "divide", "truediv":
		return makeFloat(a / b), nil
	case "//", "floordiv":
		return makeInt(int(a) / int(b)), nil

	case "%", "mod", "modulo":
		return makeInt(int(a) % int(b)), nil

	// Membership tests
	// TODO(manuel, 2024-01-22) Implement the membership tests, in fact look up how they are supposed to work
	case "in", "∈":
		return nil, errors.New("not implemented")

	case "not in", "∉":
		return nil, errors.New("not implemented")

	default:
		return nil, fmt.Errorf("unsupported operator: %s", opNode.Value)
	}
}

func (ei *EmrichenInterpreter) handleIndex(node *yaml.Node) (*yaml.Node, error) {
	args, err := parseArgs(node, []string{"over", "by"})
	if err != nil {
		return nil, err
	}

	overNode, byNode := args["over"], args["by"]
	if overNode.Kind != yaml.SequenceNode {
		return nil, errors.New("!Index 'over' argument must be a sequence")
	}

	resultVarName := "item" // Default variable name
	if resultNode, ok := args["result_as"]; ok && resultNode.Kind == yaml.ScalarNode {
		resultVarName = resultNode.Value
	}

	indexedResults := make(map[string]*yaml.Node)
	for _, item := range overNode.Content {
		// Set the current item variable
		ei.vars[resultVarName] = item

		// Process the 'by' expression to determine the key
		keyNode, err := ei.Process(byNode)
		if err != nil {
			return nil, err
		}
		if keyNode.Kind != yaml.ScalarNode {
			return nil, errors.New("!Index 'by' expression must evaluate to a scalar")
		}
		key := keyNode.Value

		// Add the item to the indexed results
		indexedResults[key] = item
	}

	// Cleanup the result variable
	delete(ei.vars, resultVarName)

	// Convert the map to a sequence of YAML nodes
	content := make([]*yaml.Node, 0, len(indexedResults)*2)
	for k, v := range indexedResults {
		content = append(content, &yaml.Node{Kind: yaml.ScalarNode, Value: k}, v)
	}

	return &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: content,
	}, nil

}
