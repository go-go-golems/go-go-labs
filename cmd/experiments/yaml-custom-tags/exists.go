package main

import (
	"gopkg.in/yaml.v3"
	"strings"
)

func (ei *EmrichenInterpreter) handleExists(node *yaml.Node) (*yaml.Node, error) {
	v, err := ei.env.LookupAll("$."+node.Value, true)
	if err != nil {
		if strings.Contains(err.Error(), "unrecognized identifier ") {
			return makeBool(false), nil
		}
		return nil, err
	}
	return makeBool(len(v) >= 1), nil
}
