package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"reflect"
)

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

	a, ok := NodeToFloat(aProcessed)
	if isNumberOperation && !ok {
		return nil, errors.New("could not convert first argument to float")
	}
	b, ok := NodeToFloat(bProcessed)
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
