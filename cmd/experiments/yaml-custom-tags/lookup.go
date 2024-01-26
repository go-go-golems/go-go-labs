package main

import "gopkg.in/yaml.v3"

func (ei *EmrichenInterpreter) handleLookup(node *yaml.Node) (*yaml.Node, error) {
	// check that the value is a string
	v, err := ei.LookupFirst(node.Value)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (ei *EmrichenInterpreter) handleLookupAll(node *yaml.Node) (*yaml.Node, error) {
	v, err := ei.LookupAll(node.Value)
	if err != nil {
		return nil, err
	}
	return v, nil
}
