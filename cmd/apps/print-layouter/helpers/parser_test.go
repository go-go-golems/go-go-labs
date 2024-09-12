package helpers

import (
	"math"
	"testing"
)

func TestExpressionParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
		hasError bool
	}{
		// Happy paths
		{"Simple inch", "10in", 10 * 96, false},
		{"Simple cm", "2.54cm", 2.54 * 96 / 2.54, false},
		{"Simple px", "100px", 100, false},
		{"Addition", "1in + 2.54cm", 96 + (2.54 * 96 / 2.54), false},
		{"Multiplication", "3mm * 4", 3 * 96 / 25.4 * 4, false},
		{"Division", "10pt / 2", 10 * 96 / 72 / 2, false},
		{"Parentheses", "(1in + 2.54cm) * 3", (96 + (2.54 * 96 / 2.54)) * 3, false},
		{"Em and rem", "5em - 2rem", (5 * 16 * 96 / 96) - (2 * 16 * 96 / 96), false},
		{"Percentage", "10% + 20px", (10 * 0.16 * 96 / 96) + 20, false},
		{"Multiple units", "1pc + 2pt + 3px", (1 * 96 / 6) + (2 * 96 / 72) + 3, false},
		{"Negative value", "-5mm", -5 * 96 / 25.4, false},
		{"Complex expression 1", "2.5in * 3 + 1cm", (2.5 * 96 * 3) + (1 * 96 / 2.54), false},
		{"Complex expression 2", "(10px + 5) * (2in - 1cm)", (10 + 5) * ((2 * 96) - (1 * 96 / 2.54)), false},
		{"Multiple relative units", "1em + 2rem + 3%", (1 * 16 * 96 / 96) + (2 * 16 * 96 / 96) + (3 * 0.16 * 96 / 96), false},
		{"Division in expression", "100px / (2 + 3)", 100 / (2 + 3), false},

		// Complex expressions
		{"Complex expression 3", "(1in + 2cm) * 3 - (4mm + 5pt) * 2", ((96 + (2 * 96 / 2.54)) * 3) - ((4*96/25.4 + 5*96/72) * 2), false},
		{"Complex expression 4", "10px * (5em - 2rem) + 3pc / (1 + 0.5)", 10*((5*16*96/96)-(2*16*96/96)) + (3*96/6)/(1+0.5), false},
		{"Complex expression 5", "(((1in + 2cm) * 3) - 4mm) - (5pt / 2)", (((96 + (2 * 96 / 2.54)) * 3) - (4 * 96 / 25.4)) - ((5 * 96 / 72) / 2), false},
		{"Complex expression 6", "1em + 2rem - 3% * (4px + 5pt) / 6pc", (1 * 16 * 96 / 96) + (2 * 16 * 96 / 96) - (3*0.16*96/96)*(4+(5*96/72))/(6*96/6), false},
		{"Complex expression 7", "-((2in + 3cm) * (4mm - 5pt)) + 6pc", -((2*96 + (3 * 96 / 2.54)) * ((4 * 96 / 25.4) - (5 * 96 / 72))) + (6 * 96 / 6), false},

		// Whitespace handling
		{"Extra whitespace", "  10in  +  5cm  ", 1152, false},
		{"No whitespace", "1in+2cm", 168, false},
		{"Mixed whitespace", "3mm *\t4 + \n5px", 50.35433070866142, false},

		// Unit variations
		{"Uppercase units", "1IN + 2CM", 168, false},
		{"Mixed case units", "3Mm * 4", 45.35433070866142, false},
		{"Mixed case relative units", "5Em - 2rEm", 48, false},

		// Edge cases
		{"Zero value", "0in", 0, false},
		{"Very small value", "0.0000001mm", 0.00000037795275590551, false},
		{"Very large value", "9999999px", 9999999, false},
		{"Scientific notation small", "1e-6in", 0.000096, false},
		{"Scientific notation large", "1e6px", 1000000, false},
		{"Leading decimal", ".5in", 48, false},
		{"Trailing decimal", "1.", 1, false},
		{"Single parentheses", "(1in)", 96, false},
		{"Multiple parentheses", "((((1in))))", 96, false},
		{"Multiple negatives", "-(-(-1in))", -96, false},

		// Failure cases
		{"Empty string", "", 0, true},
		{"Missing unit", "10", 0, true},
		{"Missing number", "in", 0, true},
		{"Invalid unit", "10kg", 0, true},
		{"Incomplete expression", "1in + ", 0, true},
		{"Missing unit in second term", "1in + 2", 0, true},
		{"Missing operator", "1in 2cm", 0, true},
		{"Unclosed parenthesis", "(1in + 2cm", 0, true},
		{"Extra closing parenthesis", "1in + 2cm)", 0, true},
		{"Missing term before operator", "1in + *2cm", 0, true},
		{"Missing term after operator", "1in + 2cm *", 0, true},
		{"Division by zero", "1/0in", 0, true},
		{"Invalid character", "1in + #2cm", 0, true},
		{"Invalid number format 1", "1.2.3in", 0, true},
		{"Invalid number format 2", "1e10e20in", 0, true},
		{"Double negative", "--1in", 0, true},
		{"Incomplete expression 2", "1in +", 0, true},
		{"Empty parentheses", "()", 0, true},
		{"Empty parentheses in expression", "1in + ()", 0, true},
		{"Unclosed parenthesis in complex expression", "1in + (2cm", 0, true},
		{"Trailing operator", "1in + 2cm + ", 0, true},
	}

	parser := &ExpressionParser{PPI: 96}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if !almostEqual(result, tt.expected, 1e-9) {
					t.Errorf("Expected %v, but got %v", tt.expected, result)
				}
			}
		})
	}
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
