package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"strings"
)

type VarResolver struct {
	vars map[string]*yaml.Node
}

func NewVarResolver() *VarResolver {
	return &VarResolver{
		vars: make(map[string]*yaml.Node),
	}
}

func (v *VarResolver) Resolve(node *yaml.Node) (*yaml.Node, error) {
	// check if node tag contains a variable name
	ss := strings.Split(node.Tag, " ")
	if len(ss) == 0 {
		return nil, errors.New("custom tag is empty")
	}
	switch ss[0] {
	case "!Defaults":
		if node.Kind == yaml.MappingNode {
			var err error
			name := ""
			for i := range node.Content {
				if i%2 == 0 {
					name = node.Content[i].Value
					continue
				}
				v.vars[name], err = resolveTags(node.Content[i])
				if err != nil {
					return nil, err
				}
			}
		}
		return node, nil

	case "!Var":
		if node.Kind == yaml.ScalarNode {
			varName := node.Value
			varValue, ok := v.vars[varName]
			if !ok {
				return nil, fmt.Errorf("variable %s not found", varName)
			}
			return varValue, nil
		}
		return nil, errors.New("variable definition must be !Var variable name")

	default:
		return nil, fmt.Errorf("unknown custom tag %s", ss[0])
	}
}
