package main

import (
	"fmt"
	"reflect"
	"log"
	"strings"
)

// Pattern is the interface that all pattern types implement.
type Pattern interface {
	Match(input interface{}, bindings Bindings) (Bindings, error)
}

// Bindings represents the variable bindings resulting from a match.
type Bindings map[string]interface{}

// copyBindings creates a shallow copy of the bindings map.
func copyBindings(bindings Bindings) Bindings {
	newBindings := make(Bindings)
	for k, v := range bindings {
		newBindings[k] = v
	}
	return newBindings
}

// VariablePattern matches variables starting with '?'.
type VariablePattern struct {
	Name string
}

func (p *VariablePattern) Match(input interface{}, bindings Bindings) (Bindings, error) {
	if val, ok := bindings[p.Name]; ok {
		if reflect.DeepEqual(val, input) {
			return bindings, nil
		}
		return nil, fmt.Errorf("variable %s mismatch: %v vs %v", p.Name, val, input)
	}
	newBindings := copyBindings(bindings)
	newBindings[p.Name] = input
	return newBindings, nil
}

// ConstantPattern matches constants (non-variable atoms).
type ConstantPattern struct {
	Value interface{}
}

func (p *ConstantPattern) Match(input interface{}, bindings Bindings) (Bindings, error) {
	if reflect.DeepEqual(p.Value, input) {
		return bindings, nil
	}
	return nil, fmt.Errorf("constant %v does not match input %v", p.Value, input)
}

// ListPattern matches a sequence of patterns against a list of inputs.
type ListPattern struct {
	Patterns []Pattern
}

func (p *ListPattern) Match(input interface{}, bindings Bindings) (Bindings, error) {
	log.Printf("ListPattern.Match input type: %T, value: %v", input, input)
	
	var inputList []interface{}
	switch v := input.(type) {
	case []interface{}:
		inputList = v
	case []string:
		inputList = make([]interface{}, len(v))
		for i, s := range v {
			inputList[i] = s
		}
	default:
		if reflect.TypeOf(input).Kind() == reflect.Slice {
			val := reflect.ValueOf(input)
			inputList = make([]interface{}, val.Len())
			for i := 0; i < val.Len(); i++ {
				inputList[i] = val.Index(i).Interface()
			}
		} else {
			return nil, fmt.Errorf("input is not a list: %T %v", input, input)
		}
	}
	
	log.Printf("ListPattern.Match converted input: %v", inputList)
	return matchList(p.Patterns, inputList, bindings)
}

func matchList(patterns []Pattern, inputs []interface{}, bindings Bindings) (Bindings, error) {
	patternStrs := make([]string, len(patterns))
	for i, p := range patterns {
		patternStrs[i] = patternToString(p)
	}
	log.Printf("matchList patterns: [%s], inputs: %v", 
		strings.Join(patternStrs, ", "), 
		inputs)
	
	if len(patterns) == 0 && len(inputs) == 0 {
		return bindings, nil
	}
	if len(patterns) == 0 || (len(inputs) == 0 && !isSpecialPattern(patterns[0])) {
		return nil, fmt.Errorf("pattern and input list length mismatch")
	}

	firstPattern := patterns[0]
	restPatterns := patterns[1:]

	// Special handling for segment patterns
	if segPattern, ok := firstPattern.(*SegmentPattern); ok {
		log.Printf("Found segment pattern: %v", segPattern)
		return segPattern.Match(inputs, bindings)
	}

	// Special handling for ?if patterns - they don't consume input
	if singlePattern, ok := firstPattern.(*SinglePattern); ok && singlePattern.Operator == "?if" {
		log.Printf("Found ?if pattern, checking predicate")
		if singlePattern.Predicate(nil, bindings) {
			return matchList(restPatterns, inputs, bindings)
		}
		return nil, fmt.Errorf("?if predicate failed")
	}

	firstInput := inputs[0]
	restInputs := inputs[1:]
	
	b1, err1 := firstPattern.Match(firstInput, bindings)
	if err1 != nil {
		return nil, err1
	}
	return matchList(restPatterns, restInputs, b1)
}

// Helper function to identify patterns that don't consume input
func isSpecialPattern(p Pattern) bool {
	if sp, ok := p.(*SinglePattern); ok {
		return sp.Operator == "?if"
	}
	return false
}

// SegmentPattern matches a segment of the input list.
type SegmentPattern struct {
	VarName string
	Rest    Pattern
	Min     int // Minimum number of elements to match (0 for ?*, 1 for ?+)
}

func (p *SegmentPattern) Match(input interface{}, bindings Bindings) (Bindings, error) {
	log.Printf("SegmentPattern.Match input type: %T, value: %v", input, input)
	
	var inputList []interface{}
	switch v := input.(type) {
	case []interface{}:
		inputList = v
	case []string:
		inputList = make([]interface{}, len(v))
		for i, s := range v {
			inputList[i] = s
		}
	default:
		if reflect.TypeOf(input).Kind() == reflect.Slice {
			val := reflect.ValueOf(input)
			inputList = make([]interface{}, val.Len())
			for i := 0; i < val.Len(); i++ {
				inputList[i] = val.Index(i).Interface()
			}
		} else {
			return nil, fmt.Errorf("input is not a list: %T %v", input, input)
		}
	}
	
	log.Printf("SegmentPattern.Match converted input: %v", inputList)
	
	if p.Rest == nil {
		if len(inputList) < p.Min {
			return nil, fmt.Errorf("input list too short")
		}
		segment := inputList
		newBindings := copyBindings(bindings)
		newBindings[p.VarName] = segment
		return newBindings, nil
	}

	// Try different splits of the input
	for i := p.Min; i <= len(inputList); i++ {
		segment := inputList[:i]
		restInput := inputList[i:]
		log.Printf("SegmentPattern trying split at %d: segment=%v, rest=%v", i, segment, restInput)
		
		newBindings := copyBindings(bindings)
		newBindings[p.VarName] = segment

		// Handle the rest pattern as a list pattern
		if listPattern, ok := p.Rest.(*ListPattern); ok {
			b2, err := listPattern.Match(restInput, newBindings)
			if err == nil {
				return b2, nil
			}
			log.Printf("ListPattern match failed: %v", err)
		} else {
			b2, err := p.Rest.Match(restInput, newBindings)
			if err == nil {
				return b2, nil
			}
			log.Printf("Rest pattern match failed: %v", err)
		}
	}
	return nil, fmt.Errorf("segment pattern did not match")
}

// SinglePattern handles special patterns like ?is, ?and, ?or, ?not.
type SinglePattern struct {
	Operator  string
	Args      []Pattern
	Predicate func(input interface{}, bindings Bindings) bool
}

func (p *SinglePattern) Match(input interface{}, bindings Bindings) (Bindings, error) {
	switch p.Operator {
	case "?is":
		if len(p.Args) != 2 {
			return nil, fmt.Errorf("?is requires two arguments")
		}
		varPattern, ok := p.Args[0].(*VariablePattern)
		if !ok {
			return nil, fmt.Errorf("?is first argument must be variable")
		}
		predicatePattern, ok := p.Args[1].(*ConstantPattern)
		if !ok {
			return nil, fmt.Errorf("?is second argument must be constant (predicate)")
		}
		predicateFuncName, ok := predicatePattern.Value.(string)
		if !ok {
			return nil, fmt.Errorf("predicate must be string")
		}
		newBindings, err := varPattern.Match(input, bindings)
		if err != nil {
			return nil, err
		}
		predicateFunc := getPredicateFunc(predicateFuncName)
		if predicateFunc == nil {
			return nil, fmt.Errorf("unknown predicate %s", predicateFuncName)
		}
		if predicateFunc(input) {
			return newBindings, nil
		}
		return nil, fmt.Errorf("predicate %s failed", predicateFuncName)
	case "?and":
		currentBindings := bindings
		for _, arg := range p.Args {
			b, err := arg.Match(input, currentBindings)
			if err != nil {
				return nil, err
			}
			currentBindings = b
		}
		return currentBindings, nil
	case "?or":
		for _, arg := range p.Args {
			b, err := arg.Match(input, bindings)
			if err == nil {
				return b, nil
			}
		}
		return nil, fmt.Errorf("?or: none of the patterns matched")
	case "?not":
		for _, arg := range p.Args {
			_, err := arg.Match(input, bindings)
			if err == nil {
				return nil, fmt.Errorf("?not: pattern matched")
			}
		}
		return bindings, nil
	case "?if":
		if p.Predicate == nil {
			return nil, fmt.Errorf("?if requires a predicate function")
		}
		if p.Predicate(input, bindings) {
			return bindings, nil
		}
		return nil, fmt.Errorf("?if predicate failed")
	default:
		return nil, fmt.Errorf("unknown operator %s", p.Operator)
	}
}

// getPredicateFunc retrieves predefined predicate functions.
var predicateFuncs = map[string]func(interface{}) bool{
	"numberp": func(v interface{}) bool {
		switch v.(type) {
		case int, float64:
			return true
		default:
			return false
		}
	},
	"oddp": func(v interface{}) bool {
		n, ok := v.(int)
		return ok && n%2 == 1
	},
}

func getPredicateFunc(name string) func(interface{}) bool {
	return predicateFuncs[name]
}

// Helper functions for creating patterns in a terse DSL.
func Var(name string) Pattern {
	return &VariablePattern{Name: name}
}

func Const(value interface{}) Pattern {
	return &ConstantPattern{Value: value}
}

func Seq(patterns ...Pattern) Pattern {
	return &ListPattern{Patterns: patterns}
}

func Seg(varName string, rest Pattern, min int) Pattern {
	return &SegmentPattern{VarName: varName, Rest: rest, Min: min}
}

func Single(operator string, args ...Pattern) Pattern {
	return &SinglePattern{Operator: operator, Args: args}
}

func SingleWithPredicate(operator string, predicate func(input interface{}, bindings Bindings) bool) Pattern {
	return &SinglePattern{Operator: operator, Predicate: predicate}
}

// Example usage of the pattern matcher.
func main() {
	// Example 1: Matching '(x = (?is ?n numberp))' against '(x = 34)'
	pattern1 := Seq(Const("x"), Const("="), Single("?is", Var("?n"), Const("numberp")))
	input1 := []interface{}{"x", "=", 34}
	bindings := Bindings{}
	result1, err := pattern1.Match(input1, bindings)
	if err != nil {
		fmt.Println("No match:", err)
	} else {
		fmt.Println("Match 1:", result1)
	}

	// Example 2: Matching '(a (?* ?x) d)' against '(a b c d)'
	pattern2 := Seq(
		Const("a"),
		Seg("?x",
			Seq(
				Const("d"),
			), 0))
	input2 := []interface{}{"a", "b", "c", "d"}
	result2, err := pattern2.Match(input2, bindings)
	if err != nil {
		fmt.Println("No match:", err)
	} else {
		fmt.Println("Match 2:", result2)
	}

	// Example 3: Matching '(?x > ?y (?if (> ?x ?y)))' against '(4 > 3)'
	pattern3 := Seq(Var("?x"), Const(">"), Var("?y"), SingleWithPredicate("?if", func(input interface{}, bindings Bindings) bool {
		x, ok1 := bindings["?x"].(int)
		y, ok2 := bindings["?y"].(int)
		return ok1 && ok2 && x > y
	}))
	input3 := []interface{}{4, ">", 3}
	result3, err := pattern3.Match(input3, bindings)
	if err != nil {
		fmt.Println("No match:", err)
	} else {
		fmt.Println("Match 3:", result3)
	}
}

// Add this helper function near the top of the file
func patternToString(p Pattern) string {
	switch v := p.(type) {
	case *VariablePattern:
		return fmt.Sprintf("Var(%s)", v.Name)
	case *ConstantPattern:
		return fmt.Sprintf("Const(%v)", v.Value)
	case *ListPattern:
		patterns := make([]string, len(v.Patterns))
		for i, p := range v.Patterns {
			patterns[i] = patternToString(p)
		}
		return fmt.Sprintf("List(%v)", patterns)
	case *SegmentPattern:
		if v.Rest == nil {
			return fmt.Sprintf("Seg(%s, nil, %d)", v.VarName, v.Min)
		}
		return fmt.Sprintf("Seg(%s, %s, %d)", v.VarName, patternToString(v.Rest), v.Min)
	case *SinglePattern:
		args := make([]string, len(v.Args))
		for i, p := range v.Args {
			args[i] = patternToString(p)
		}
		return fmt.Sprintf("Single(%s, %v)", v.Operator, args)
	default:
		return fmt.Sprintf("%T", p)
	}
}
