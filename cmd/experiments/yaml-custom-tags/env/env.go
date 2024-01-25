// Package env provides an environment for managing variable frames
// in an interpreter context, using a stack-based approach.
package env

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/util/jsonpath"
)

// Frame represents a single variable frame containing a map of variables.
type Frame struct {
	Variables map[string]interface{}
}

// NewFrame creates a new Frame with variables. It takes an optional parent frame
// and merges its variables (shallow copy) with newVars. If parent is nil, only newVars
// are used. It's primarily used internally by Env.
func NewFrame(parent *Frame, newVars map[string]interface{}) *Frame {
	mergedVars := make(map[string]interface{})
	if parent != nil {
		// Shallow copy of the parent's variables
		for k, v := range parent.Variables {
			mergedVars[k] = v
		}
	}

	// Add new variables
	for k, v := range newVars {
		mergedVars[k] = v
	}

	return &Frame{Variables: mergedVars}
}

// Env represents an environment with a stack of variable frames.
// It allows pushing and popping frames and querying variables.
type Env struct {
	stack []*Frame
}

type EnvOption func(*Env)

// NewEnv creates a new environment for managing variable frames.
// It accepts optional EnvOptions to configure the initial state.
func NewEnv(options ...EnvOption) *Env {
	env := &Env{stack: make([]*Frame, 0)}
	for _, option := range options {
		option(env)
	}
	return env
}

// WithVars is an option for NewEnv to initialize the environment
// with a set of variables. If used when the stack is empty, it creates
// a new Frame with these variables. Otherwise, it adds the variables
// to the current frame.
func WithVars(vars map[string]interface{}) EnvOption {
	return func(e *Env) {
		if len(e.stack) == 0 {
			e.stack = append(e.stack, &Frame{Variables: vars})
			return
		}
		currentFrame := e.GetCurrentFrame()
		for k, v := range vars {
			currentFrame.Variables[k] = v
		}
	}
}

// Push creates a new frame on top of the stack with newVars.
// It performs a shallow copy of the variables from the current top frame
// and merges them with newVars. If the stack is empty, newVars becomes
// the first frame.
func (e *Env) Push(newVars map[string]interface{}) {
	var parent *Frame
	if len(e.stack) > 0 {
		parent = e.stack[len(e.stack)-1]
	}
	e.stack = append(e.stack, NewFrame(parent, newVars))
}

// Pop removes the top frame from the stack. It does nothing if the stack is empty.
func (e *Env) Pop() {
	if len(e.stack) == 0 {
		return
	}
	e.stack = e.stack[:len(e.stack)-1]
}

// GetCurrentFrame returns the current top frame from the stack.
// Returns nil if the stack is empty.
func (e *Env) GetCurrentFrame() *Frame {
	if len(e.stack) == 0 {
		return nil
	}
	return e.stack[len(e.stack)-1]
}

// GetVar tries to retrieve a variable's value by its name from the current frame.
// Returns the value and a boolean indicating if the variable was found.
// If the stack is empty, it returns nil and false.
func (e *Env) GetVar(name string) (interface{}, bool) {
	v := e.GetCurrentFrame()
	if v == nil {
		return nil, false
	}
	val, ok := v.Variables[name]
	return val, ok
}

// LookupAll performs a jsonpath query on the variables of the current frame.
// It returns all matches as a slice of interface{} and an error if the query
// fails or if the current frame is nil. The function requires a valid jsonpath
// expression and uses the Kubernetes jsonpath package.
func (e *Env) LookupAll(expression string, allowMissingKeys bool) ([]interface{}, error) {
	v := e.GetCurrentFrame()
	if v == nil {
		return nil, nil
	}

	j := jsonpath.New("jsonpath")
	err := j.Parse("{" + expression + "}")
	if err != nil {
		return nil, err
	}

	results, err := j.AllowMissingKeys(allowMissingKeys).FindResults(v.Variables)
	if err != nil {
		return nil, err
	}

	var finalResults []interface{}
	for _, result := range results {
		for _, r := range result {
			finalResults = append(finalResults, r.Interface())
		}
	}

	return finalResults, nil
}

// LookupFirst performs a jsonpath query on the variables of the current frame.
// It returns the first match as an interface{} and an error if the query
// fails or if the current frame is nil, or if no matching node is found.
func (e *Env) LookupFirst(expression string) (interface{}, error) {
	res, err := e.LookupAll(expression, false)
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, errors.Errorf("no matching node found for expression %q", expression)
	}

	return res[0], nil
}
