package main

import "gopkg.in/yaml.v3"

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
