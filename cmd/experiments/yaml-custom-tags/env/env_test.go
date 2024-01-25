package env

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEnvFrameWithVars(t *testing.T) {
	// Define the test cases
	tests := []struct {
		name           string
		parentVars     map[string]interface{}
		newVars        map[string]interface{}
		expectedVars   map[string]interface{}
		expectedFrames int
	}{
		{
			name:           "NewFrame with no parent",
			parentVars:     nil,
			newVars:        map[string]interface{}{"a": 1},
			expectedVars:   map[string]interface{}{"a": 1},
			expectedFrames: 1,
		},
		{
			name:           "NewFrame with parent",
			parentVars:     map[string]interface{}{"a": 1},
			newVars:        map[string]interface{}{"b": 2},
			expectedVars:   map[string]interface{}{"a": 1, "b": 2},
			expectedFrames: 1,
		},
		{
			name:           "NewFrame with overlapping Variables",
			parentVars:     map[string]interface{}{"a": 1},
			newVars:        map[string]interface{}{"a": 2},
			expectedVars:   map[string]interface{}{"a": 2},
			expectedFrames: 1,
		},
		{
			name:           "NewEnv with WithVars option on non-empty stack",
			parentVars:     map[string]interface{}{"a": 1},
			newVars:        map[string]interface{}{"b": 3, "c": 3},
			expectedVars:   map[string]interface{}{"a": 1, "b": 3, "c": 3},
			expectedFrames: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup parent frame if needed
			var parent *Frame
			if tc.parentVars != nil {
				parent = &Frame{Variables: tc.parentVars}
			}

			// Create new frame with newVars and parent
			frame := NewFrame(parent, tc.newVars)

			// Verify the frame's variables
			assert.Equal(t, tc.expectedVars, frame.Variables)

			// Create new environment with options
			env := NewEnv(WithVars(frame.Variables))

			// Verify the number of frames in the environment
			assert.Len(t, env.stack, tc.expectedFrames)

			if tc.expectedFrames > 0 {
				// Verify the variables in the current top frame
				currentFrame := env.GetCurrentFrame()
				assert.Equal(t, tc.expectedVars, currentFrame.Variables)
			}
		})
	}
}

func TestEnvPushPopGetCurrentFrame(t *testing.T) {
	type action struct {
		method string
		vars   map[string]interface{}
	}
	tests := []struct {
		name            string
		initialVars     map[string]interface{}
		actions         []action
		expectedVars    map[string]interface{}
		expectedFrames  int
		expectedTopVars map[string]interface{}
	}{
		{
			name:            "Push on empty stack",
			initialVars:     nil,
			actions:         []action{{method: "Push", vars: map[string]interface{}{"a": 1}}},
			expectedVars:    map[string]interface{}{"a": 1},
			expectedFrames:  1,
			expectedTopVars: map[string]interface{}{"a": 1},
		},
		{
			name:            "Push on non-empty stack",
			initialVars:     map[string]interface{}{"a": 1},
			actions:         []action{{method: "Push", vars: map[string]interface{}{"b": 2}}},
			expectedVars:    map[string]interface{}{"a": 1, "b": 2},
			expectedFrames:  2,
			expectedTopVars: map[string]interface{}{"a": 1, "b": 2},
		},
		{
			name:            "Pop removes top frame",
			initialVars:     map[string]interface{}{"a": 1},
			actions:         []action{{method: "Push", vars: map[string]interface{}{"b": 2}}, {method: "Pop"}},
			expectedVars:    map[string]interface{}{"a": 1},
			expectedFrames:  1,
			expectedTopVars: map[string]interface{}{"a": 1},
		},
		{
			name:            "Pop on empty stack does nothing",
			initialVars:     nil,
			actions:         []action{{method: "Pop"}},
			expectedVars:    nil,
			expectedFrames:  0,
			expectedTopVars: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create new environment and push initial frame if provided
			env := NewEnv()
			if tc.initialVars != nil {
				env.Push(tc.initialVars)
			}

			// Perform actions
			for _, act := range tc.actions {
				switch act.method {
				case "Push":
					env.Push(act.vars)
				case "Pop":
					env.Pop()
				}
			}

			// Verify the number of frames in the environment
			assert.Len(t, env.stack, tc.expectedFrames)

			// Verify the variables in the current top frame
			currentFrame := env.GetCurrentFrame()
			if currentFrame != nil {
				assert.Equal(t, tc.expectedTopVars, currentFrame.Variables)
			} else {
				assert.Nil(t, tc.expectedTopVars)
			}
		})
	}
}

func TestEnvGetVarAndLookupAll(t *testing.T) {
	type getVarTest struct {
		varName     string
		expectedVal interface{}
		expectedOk  bool
	}

	type lookupAllTest struct {
		expression    string
		expectedMatch []interface{}
		expectedErr   bool
	}

	tests := []struct {
		name           string
		initialVars    map[string]interface{}
		getVarTests    []getVarTest
		lookupAllTests []lookupAllTest
	}{
		{
			name:        "GetVar on empty stack",
			initialVars: nil,
			getVarTests: []getVarTest{
				{"a", nil, false},
			},
			lookupAllTests: []lookupAllTest{
				{"$.a", nil, false},
			},
		},
		{
			name:        "GetVar with variable found",
			initialVars: map[string]interface{}{"a": 1, "b": "test"},
			getVarTests: []getVarTest{
				{"a", 1, true},
				{"b", "test", true},
			},
			lookupAllTests: []lookupAllTest{
				{"$.a", []interface{}{1}, false},
				{"$.b", []interface{}{"test"}, false},
			},
		},
		{
			name:        "GetVar with variable not found",
			initialVars: map[string]interface{}{"a": 1},
			getVarTests: []getVarTest{
				{"b", nil, false},
			},
			lookupAllTests: []lookupAllTest{
				{"$.b", nil, true},
			},
		},
		{
			name:        "LookupAll with valid expression",
			initialVars: map[string]interface{}{"a": 1, "b": map[string]interface{}{"c": 2}},
			getVarTests: []getVarTest{
				{"a", 1, true},
			},
			lookupAllTests: []lookupAllTest{
				{"$.b.c", []interface{}{2}, false},
			},
		},
		{
			name:        "LookupAll with invalid expression",
			initialVars: map[string]interface{}{"a": 1},
			getVarTests: []getVarTest{
				{"a", 1, true},
			},
			lookupAllTests: []lookupAllTest{
				{"$.a[", nil, true},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create new environment and push initial frame if provided
			env := NewEnv()
			if tc.initialVars != nil {
				env.Push(tc.initialVars)
			}

			// Test GetVar
			for _, gvTest := range tc.getVarTests {
				val, ok := env.GetVar(gvTest.varName)
				assert.Equal(t, gvTest.expectedVal, val)
				assert.Equal(t, gvTest.expectedOk, ok)
			}

			// Test LookupAll
			for _, laTest := range tc.lookupAllTests {
				match, err := env.LookupAll(laTest.expression, false)
				if laTest.expectedErr {
					assert.Error(t, err)
					continue
				}
				require.NoError(t, err)
				assert.Equal(t, laTest.expectedMatch, match)
			}
		})
	}
}

func TestEnvStackOperations(t *testing.T) {
	type action struct {
		method string
		vars   map[string]interface{}
	}
	type getVarTest struct {
		varName     string
		expectedVal interface{}
		expectedOk  bool
	}
	type lookupTest struct {
		expression    string
		expectedMatch interface{}
		expectedErr   bool
	}
	tests := []struct {
		name             string
		actions          []action
		getVarTests      []getVarTest
		lookupAllTests   []lookupTest
		lookupFirstTests []lookupTest
	}{
		{
			name: "Push single frame and lookup",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"x": 10, "y": 20}},
			},
			getVarTests: []getVarTest{
				{"x", 10, true},
				{"y", 20, true},
			},
		},
		{
			name: "Push multiple frames and lookup",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"a": 1}},
				{method: "push", vars: map[string]interface{}{"b": 2}},
			},
			getVarTests: []getVarTest{
				{"a", 1, true},
				{"b", 2, true},
			},
		},
		{
			name: "Variable shadowing in stacked frames",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"var": "first"}},
				{method: "push", vars: map[string]interface{}{"var": "second"}},
			},
			getVarTests: []getVarTest{
				{"var", "second", true},
			},
		},
		{
			name: "Pop and variable restoration",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"var1": "value1"}},
				{method: "push", vars: map[string]interface{}{"var2": "value2"}},
				{method: "pop"},
			},
			getVarTests: []getVarTest{
				{"var1", "value1", true},
				{"var2", nil, false},
			},
		},
		{
			name: "Pop on empty stack",
			actions: []action{
				{method: "pop"},
			},
			getVarTests: []getVarTest{
				// No variables should be found since the stack is empty
				{"any", nil, false},
			},
		},
		{
			name:    "Lookup with empty stack",
			actions: []action{
				// No actions, as we want to test the empty stack
			},
			getVarTests: []getVarTest{
				{"any", nil, false},
			},
		},
		{
			name: "LookupAll with valid and invalid expressions",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"nested": map[string]interface{}{"key": "value"}}},
			},
			lookupAllTests: []lookupTest{
				{expression: "$.nested.key", expectedMatch: []interface{}{"value"}, expectedErr: false},
				{expression: "$.invalid[", expectedMatch: nil, expectedErr: true},
			},
		},
		{
			name: "LookupFirst with no matching node",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"a": 1}},
			},
			lookupFirstTests: []lookupTest{
				{expression: "$.nonexistent", expectedMatch: nil, expectedErr: true},
			},
		},
		{
			name: "LookupFirst with matching node",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"a": 1}},
			},
			lookupFirstTests: []lookupTest{
				{expression: "$.a", expectedMatch: 1, expectedErr: false},
			},
		},
		{
			name: "Push single frame with multiple variables and lookup non-existent variable",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"x": 10, "y": 20}},
			},
			getVarTests: []getVarTest{
				{"z", nil, false},
			},
		},
		{
			name: "Push multiple frames and lookup variable from non-top frame",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"a": 1}},
				{method: "push", vars: map[string]interface{}{"b": 2}},
			},
			getVarTests: []getVarTest{
				{"a", 1, true},
			},
		},
		{
			name: "Variable shadowing and unshadowing with multiple push and pop",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"var": "first"}},
				{method: "push", vars: map[string]interface{}{"var": "second"}},
				{method: "pop"},
			},
			getVarTests: []getVarTest{
				{"var", "first", true},
			},
		},
		{
			name: "Pop until empty stack and verify no variables are found",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"var1": "value1"}},
				{method: "pop"},
			},
			getVarTests: []getVarTest{
				{"var1", nil, false},
			},
		},
		{
			name: "Pop on stack with single frame",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"var": "value"}},
				{method: "pop"},
			},
			getVarTests: []getVarTest{
				{"var", nil, false},
			},
		},
		{
			name: "Lookup with stack containing a single empty frame",
			actions: []action{
				{method: "push", vars: map[string]interface{}{}}, // Empty variables map
			},
			getVarTests: []getVarTest{
				{"any", nil, false},
			},
		},
		{
			name: "LookupAll with nested structures and multiple matches",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"nested": map[string]interface{}{"key": []interface{}{"value1", "value2"}}}},
			},
			lookupAllTests: []lookupTest{
				{expression: "$.nested.key[*]", expectedMatch: []interface{}{"value1", "value2"}, expectedErr: false},
			},
		},
		{
			name: "LookupFirst with multiple matching nodes",
			actions: []action{
				{method: "push", vars: map[string]interface{}{"a": []interface{}{1, 2, 3}}},
			},
			lookupFirstTests: []lookupTest{
				// Using LookupAll to simulate LookupFirst behavior
				{expression: "$.a[*]", expectedMatch: 1, expectedErr: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := NewEnv()

			// Perform actions
			for _, act := range tt.actions {
				switch act.method {
				case "push":
					env.Push(act.vars)
				case "pop":
					env.Pop()
				}
			}

			// Test GetVar
			for _, gvTest := range tt.getVarTests {
				val, ok := env.GetVar(gvTest.varName)
				assert.Equal(t, gvTest.expectedVal, val)
				assert.Equal(t, gvTest.expectedOk, ok)
			}

			// Test LookupAll
			for _, laTest := range tt.lookupAllTests {
				match, err := env.LookupAll(laTest.expression, true)
				if laTest.expectedErr {
					assert.Error(t, err)
					continue
				}
				require.NoError(t, err)
				assert.Equal(t, match, laTest.expectedMatch)
			}

			// Test LookupFirst
			for _, laTest := range tt.lookupFirstTests {
				match, err := env.LookupFirst(laTest.expression)
				if laTest.expectedErr {
					assert.Error(t, err)
					continue
				}
				require.NoError(t, err)
				assert.Equal(t, laTest.expectedMatch, match)
			}
		})
	}
}
