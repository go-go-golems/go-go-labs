package main

import "gopkg.in/yaml.v3"

func (ei *EmrichenInterpreter) handleIf(node *yaml.Node) (*yaml.Node, error) {
	args, err := ei.parseArgs(node, []parsedVariable{
		{Name: "test", Required: true},
		{Name: "then"},
		{Name: "else"},
	})
	if err != nil {
		return nil, err
	}

	testResult, err := ei.Process(args["test"])
	if err != nil {
		return nil, err
	}

	if isTruthy(testResult) {
		if args["then"] == nil {
			return ValueToNode(nil)
		}
		return ei.Process(args["then"])
	} else {
		if args["else"] == nil {
			return ValueToNode(nil)
		}
		return ei.Process(args["else"])
	}
}
