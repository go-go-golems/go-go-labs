package main

import "gopkg.in/yaml.v3"

func (ei *EmrichenInterpreter) handleNot(node *yaml.Node) (*yaml.Node, error) {
	return makeBool(!isTruthy(node)), nil
}
