package main

import (
	"reflect"
	"testing"
)

// TestPatMatch runs a series of pattern matching tests.
func TestPatMatch(t *testing.T) {
	tests := []struct {
		name             string
		pattern          Pattern
		input            interface{}
		expectedBindings Bindings
		shouldMatch      bool
	}{
		// 1. Variable Matching
		{
			name:             "Variable matching atom",
			pattern:          Var("?x"),
			input:            "a",
			expectedBindings: Bindings{"?x": "a"},
			shouldMatch:      true,
		},
		{
			name:             "Variable matching list",
			pattern:          Var("?x"),
			input:            []interface{}{"a", "b"},
			expectedBindings: Bindings{"?x": []interface{}{"a", "b"}},
			shouldMatch:      true,
		},

		// 2. Constant Matching
		{
			name:             "Constant matches same",
			pattern:          Const("a"),
			input:            "a",
			shouldMatch:      true,
			expectedBindings: nil,
		},
		{
			name:        "Constant does not match different",
			pattern:     Const("a"),
			input:       "b",
			shouldMatch: false,
		},

		// 3. Sequence Matching
		{
			name:        "Sequence matches same",
			pattern:     Seq(Const("a"), Const("b"), Const("c")),
			input:       []interface{}{"a", "b", "c"},
			shouldMatch: true,
		},
		{
			name:        "Sequence does not match different",
			pattern:     Seq(Const("a"), Const("b"), Const("c")),
			input:       []interface{}{"a", "b", "d"},
			shouldMatch: false,
		},

		// 4. Variable Binding
		{
			name:             "Variable binding",
			pattern:          Seq(Const("a"), Var("?x"), Const("c")),
			input:            []interface{}{"a", "b", "c"},
			expectedBindings: Bindings{"?x": "b"},
			shouldMatch:      true,
		},
		{
			name:        "Variable binding fails",
			pattern:     Seq(Const("a"), Var("?x"), Const("c")),
			input:       []interface{}{"a", "b", "d"},
			shouldMatch: false,
		},

		// 5. Segment Patterns '?*' matches zero or more elements
		{
			name:             "Segment pattern ?* matches empty list",
			pattern:          Seg("?x", nil, 0),
			input:            []interface{}{},
			expectedBindings: Bindings{"?x": []interface{}{}},
			shouldMatch:      true,
		},
		{
			name:             "Segment pattern ?* matches list",
			pattern:          Seg("?x", nil, 0),
			input:            []interface{}{"a", "b", "c"},
			expectedBindings: Bindings{"?x": []interface{}{"a", "b", "c"}},
			shouldMatch:      true,
		},

		// 6. Segment Patterns with Trailing Patterns
		{
			name:             "Segment pattern with trailing pattern matches",
			pattern:          Seq(Const("a"), Seg("?x", Seq(Const("d")), 0)),
			input:            []interface{}{"a", "b", "c", "d"},
			expectedBindings: Bindings{"?x": []interface{}{"b", "c"}},
			shouldMatch:      true,
		},
		{
			name:             "Segment pattern with trailing pattern matches empty segment",
			pattern:          Seq(Const("a"), Seg("?x", Seq(Const("d")), 0)),
			input:            []interface{}{"a", "d"},
			expectedBindings: Bindings{"?x": []interface{}{}},
			shouldMatch:      true,
		},

		// 7. Segment Patterns '?+' matches one or more elements
		{
			name:             "Segment pattern ?+ matches non-empty list",
			pattern:          Seq(Const("a"), Seg("?x", Seq(Const("d")), 1)),
			input:            []interface{}{"a", "b", "c", "d"},
			expectedBindings: Bindings{"?x": []interface{}{"b", "c"}},
			shouldMatch:      true,
		},
		{
			name:        "Segment pattern ?+ fails on empty segment",
			pattern:     Seq(Const("a"), Seg("?x", Seq(Const("d")), 1)),
			input:       []interface{}{"a", "d"},
			shouldMatch: false,
		},

		// 8. Segment Patterns '??' matches zero or one element
		{
			name:             "Segment pattern ?? matches one element",
			pattern:          Seq(Const("a"), Seg("?x", Seq(Const("d")), 0), Const("d")),
			input:            []interface{}{"a", "b", "d"},
			expectedBindings: Bindings{"?x": []interface{}{"b"}},
			shouldMatch:      true,
		},
		{
			name:             "Segment pattern ?? matches zero elements",
			pattern:          Seq(Const("a"), Seg("?x", Seq(Const("d")), 0), Const("d")),
			input:            []interface{}{"a", "d"},
			expectedBindings: Bindings{"?x": []interface{}{}},
			shouldMatch:      true,
		},

		// 9. Single Patterns '?is'
		{
			name:             "Single pattern ?is matches number",
			pattern:          Single("?is", Var("?n"), Const("numberp")),
			input:            3,
			expectedBindings: Bindings{"?n": 3},
			shouldMatch:      true,
		},
		{
			name:        "Single pattern ?is fails on non-number",
			pattern:     Single("?is", Var("?n"), Const("numberp")),
			input:       "a",
			shouldMatch: false,
		},

		// 10. Single Patterns '?and'
		{
			name: "Single pattern ?and succeeds",
			pattern: Single("?and",
				Single("?is", Var("?n"), Const("numberp")),
				Single("?is", Var("?n"), Const("oddp")),
			),
			input:            3,
			expectedBindings: Bindings{"?n": 3},
			shouldMatch:      true,
		},
		{
			name: "Single pattern ?and fails",
			pattern: Single("?and",
				Single("?is", Var("?n"), Const("numberp")),
				Single("?is", Var("?n"), Const("oddp")),
			),
			input:       4,
			shouldMatch: false,
		},

		// 11. Single Patterns '?or'
		{
			name:        "Single pattern ?or matches first",
			pattern:     Single("?or", Const("<"), Const("="), Const(">")),
			input:       "<",
			shouldMatch: true,
		},
		{
			name:        "Single pattern ?or matches none",
			pattern:     Single("?or", Const("<"), Const("="), Const(">")),
			input:       "!=",
			shouldMatch: false,
		},

		// 12. Single Patterns '?not'
		{
			name:        "Single pattern ?not fails when pattern matches",
			pattern:     Single("?not", Var("?x")),
			input:       "a",
			shouldMatch: false,
		},
		{
			name:        "Single pattern ?not succeeds when pattern does not match",
			pattern:     Single("?not", Const("a")),
			input:       "b",
			shouldMatch: true,
		},

		// 13. Patterns with '?if'
		{
			name: "Pattern with ?if succeeds",
			pattern: Seq(
				Var("?x"),
				Var("?op"),
				Var("?y"),
				SingleWithPredicate("?if", func(input interface{}, bindings Bindings) bool {
					x, ok1 := bindings["?x"].(int)
					y, ok2 := bindings["?y"].(int)
					return ok1 && ok2 && x > y
				}),
			),
			input: []interface{}{4, ">", 3},
			expectedBindings: Bindings{
				"?x":  4,
				"?op": ">",
				"?y":  3,
			},
			shouldMatch: true,
		},
		{
			name: "Pattern with ?if fails",
			pattern: Seq(
				Var("?x"),
				Var("?op"),
				Var("?y"),
				SingleWithPredicate("?if", func(input interface{}, bindings Bindings) bool {
					x, ok1 := bindings["?x"].(int)
					y, ok2 := bindings["?y"].(int)
					return ok1 && ok2 && x > y
				}),
			),
			input:       []interface{}{2, ">", 3},
			shouldMatch: false,
		},

		// 14. Nested Patterns
		{
			name: "Nested pattern succeeds",
			pattern: Single("?and",
				Single("?or", Const("a"), Const("b")),
				Single("?not", Const("c")),
			),
			input:       "a",
			shouldMatch: true,
		},
		{
			name: "Nested pattern fails",
			pattern: Single("?and",
				Single("?or", Const("a"), Const("b")),
				Single("?not", Const("c")),
			),
			input:       "c",
			shouldMatch: false,
		},

		// 15. Consistency of Variable Bindings
		{
			name:             "Variable binding consistent",
			pattern:          Seq(Var("?x"), Var("?x")),
			input:            []interface{}{"a", "a"},
			expectedBindings: Bindings{"?x": "a"},
			shouldMatch:      true,
		},
		{
			name:        "Variable binding inconsistent",
			pattern:     Seq(Var("?x"), Var("?x")),
			input:       []interface{}{"a", "b"},
			shouldMatch: false,
		},

		// 16. Failure Cases
		{
			name:        "Sequence longer than input fails",
			pattern:     Seq(Const("a"), Const("b"), Const("c")),
			input:       []interface{}{"a", "b"},
			shouldMatch: false,
		},
		{
			name:             "Variable pattern matches empty input",
			pattern:          Var("?x"),
			input:            []interface{}{},
			expectedBindings: Bindings{"?x": []interface{}{}},
			shouldMatch:      true,
		},

		// 17. Empty Patterns and Inputs
		{
			name:        "Empty pattern matches empty input",
			pattern:     Seq(),
			input:       []interface{}{},
			shouldMatch: true,
		},
		{
			name:        "Empty pattern does not match non-empty input",
			pattern:     Seq(),
			input:       []interface{}{"a"},
			shouldMatch: false,
		},

		// 18. Non-List Inputs
		{
			name:             "Variable matches atom",
			pattern:          Var("?x"),
			input:            "a",
			expectedBindings: Bindings{"?x": "a"},
			shouldMatch:      true,
		},
		{
			name:             "Variable matches list",
			pattern:          Var("?x"),
			input:            []interface{}{"a"},
			expectedBindings: Bindings{"?x": []interface{}{"a"}},
			shouldMatch:      true,
		},

		// 19. Variables Matching Lists
		{
			name:             "Variable matches list",
			pattern:          Var("?x"),
			input:            []interface{}{"a", "b", "c"},
			expectedBindings: Bindings{"?x": []interface{}{"a", "b", "c"}},
			shouldMatch:      true,
		},

		// 20. Complex Patterns
		{
			name: "Complex pattern with multiple segment variables",
			pattern: Seq(
				Const("a"),
				Seg("?x", Seq(
					Const("b"),
					Seg("?y", Seq(Const("c")), 0),
				), 0),
			),
			input: []interface{}{"a", 1, 2, 3, "b", 4, 5, "c"},
			expectedBindings: Bindings{
				"?x": []interface{}{1, 2, 3},
				"?y": []interface{}{4, 5},
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bindings := Bindings{}
			result, err := tt.pattern.Match(tt.input, bindings)
			if tt.shouldMatch {
				if err != nil {
					t.Errorf("Expected match but got error: %v", err)
				} else if tt.expectedBindings != nil && !reflect.DeepEqual(result, tt.expectedBindings) {
					t.Errorf("Expected bindings %v, got %v", tt.expectedBindings, result)
				}
			} else {
				if err == nil {
					t.Errorf("Expected no match but got bindings: %v", result)
				}
			}
		})
	}
}
