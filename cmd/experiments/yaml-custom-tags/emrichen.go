package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/yaml-custom-tags/env"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
)

type EmrichenInterpreter struct {
	env *env.Env
}

type EmrichenInterpreterOption func(*EmrichenInterpreter) error

func WithVars(vars map[string]interface{}) EmrichenInterpreterOption {
	return func(ei *EmrichenInterpreter) error {
		ei.env.Push(vars)
		return nil
	}
}

func NewEmrichenInterpreter(options ...EmrichenInterpreterOption) (*EmrichenInterpreter, error) {
	ret := &EmrichenInterpreter{
		env: env.NewEnv(),
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

func (ei *EmrichenInterpreter) LookupFirst(jsonPath string) (*yaml.Node, error) {
	v, err := ei.env.LookupFirst("$." + jsonPath)
	if err != nil {
		return nil, err
	}
	node, err := ValueToNode(v)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func (ei *EmrichenInterpreter) LookupAll(jsonPath string) (*yaml.Node, error) {
	v, err := ei.env.LookupAll("$."+jsonPath, true)
	if err != nil {
		return nil, err
	}
	node, err := ValueToNode(v)
	if err != nil {
		return nil, err
	}
	return node, nil
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

			case "!Base64":
				if node.Kind != yaml.ScalarNode {
					return nil, errors.New("!Base64 requires a scalar value")
				}
				return makeString(base64.StdEncoding.EncodeToString([]byte(node.Value))), nil

			case "!Concat":
				return ei.handleConcat(node)

			case "!Debug":
				v, err := ei.Process(node)
				if err != nil {
					return nil, err
				}
				toInterface, _ := NodeToInterface(v)
				fmt.Printf("DEBUG: %s\n", toInterface)
				return v, nil

			case "!Error":
				if node.Kind != yaml.ScalarNode {
					return nil, errors.New("!Error tag requires a scalar value for the error message")
				}
				errorString, err := ei.renderFormatString(node.Value)
				if err != nil {
					return nil, err
				}
				return nil, errors.New(errorString)

			case "!Exists":
				return ei.handleExists(node)

			case "!Format":
				return ei.handleFormat(node)

			case "!Filter":
				return ei.handleFilter(node)

			case "!Group":
				return ei.handleGroup(node)

			case "!If":
				return ei.handleIf(node)

			case "!Include":
				return ei.handleInclude(node)

			case "!Index":
				return ei.handleIndex(node)

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

			case "!Join":
				return ei.handleJoin(node)

			case "!Loop":
				return ei.handleLoop(node)

			case "!Lookup":
				return ei.handleLookup(node)

			case "!LookupAll":
				return ei.handleLookupAll(node)

			case "!MD5":
				if node.Kind != yaml.ScalarNode {
					return nil, errors.New("!MD5 requires a scalar value")
				}
				hash := md5.Sum([]byte(node.Value))
				return makeString(hex.EncodeToString(hash[:])), nil

			case "!Merge":
				return ei.handleMerge(node)

			case "!Not":
				return ei.handleNot(node)

			case "!Op":
				return ei.handleOp(node)

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

			case "!Var":
				if node.Kind == yaml.ScalarNode {
					varName := node.Value
					varValue, ok := ei.env.GetVar(varName)
					if !ok {
						return nil, fmt.Errorf("variable %s not found", varName)
					}
					v, err := ValueToNode(varValue)
					if err != nil {
						return nil, err
					}
					return v, nil
				}
				return nil, errors.New("variable definition must be !Var variable name")

			case "!URLEncode":
				return ei.handleURLEncode(node)

			case "!Void":
				return nil, nil

			case "!With":
				return ei.handleWith(node)

			default:
			}

			// TODO(manuel, 2024-01-25) This is where we need to handle void in sequences and mappings
			switch node.Kind {
			case yaml.SequenceNode:
				retContent := make([]*yaml.Node, 0)
				for i := range node.Content {
					v, err := ei.Process(node.Content[i])
					if err != nil {
						return nil, err
					}
					if v == nil {
						continue
					}
					retContent = append(retContent, v)
				}
				return &yaml.Node{
					Kind:    yaml.SequenceNode,
					Content: retContent,
					Tag:     "!!seq",
				}, nil
			case yaml.MappingNode:
				retContent := make([]*yaml.Node, 0)
				for i := 0; i < len(node.Content); i += 2 {
					key := node.Content[i]
					value := node.Content[i+1]

					v, err := ei.Process(value)
					if err != nil {
						return nil, err
					}
					if v == nil {
						continue
					}
					retContent = append(retContent, key, v)
				}
				return &yaml.Node{
					Kind:    yaml.MappingNode,
					Content: retContent,
					Tag:     "!!map",
				}, nil
			case yaml.ScalarNode:
				return node, nil
			case yaml.AliasNode:
				return nil, errors.New("alias nodes are not supported")
			case yaml.DocumentNode:
				return node, nil
			default:
				return nil, errors.Errorf("unknown node kind: %v", node.Kind)
			}
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
	vars := map[string]interface{}{}
	for i := range content {
		if i%2 == 0 {
			name = content[i].Value
			continue
		}
		node, err := ei.Process(content[i])
		if err != nil {
			return err
		}
		v, ok := NodeToInterface(node)
		if !ok {
			return errors.New("could not get node value")
		}
		vars[name] = v
	}

	ei.env.Push(vars)

	return nil
}
