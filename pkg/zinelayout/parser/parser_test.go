package parser

import (
	"math"
	"testing"
)

func TestExpressionParser(t *testing.T) {
	uc := &UnitConverter{PPI: 96}

	tests := []struct {
		name     string
		input    string
		expected float64
		hasError bool
	}{
		// Happy paths
		{"Simple inch", "10in", uc.FromInch(10), false},
		{"Simple cm", "2.54cm", uc.FromCentimeter(2.54), false},
		{"Simple px", "100px", 100, false},
		{"Addition", "1in + 2.54cm", uc.FromInch(1) + uc.FromCentimeter(2.54), false},
		{"Multiplication", "3mm * 4", uc.FromMillimeter(3) * 4, false},
		{"Division", "10pt / 2", uc.FromPoint(10) / 2, false},
		{"Parentheses", "(1in + 2.54cm) * 3", (uc.FromInch(1) + uc.FromCentimeter(2.54)) * 3, false},
		{"Em and rem", "5em - 2rem", uc.FromEm(5) - uc.FromEm(2), false},
		{"Multiple units", "1pc + 2pt + 3px", uc.FromPica(1) + uc.FromPoint(2) + 3, false},
		{"Negative value", "-5mm", -uc.FromMillimeter(5), false},
		{"Complex expression 1", "2.5in * 3 + 1cm", uc.FromInch(2.5)*3 + uc.FromCentimeter(1), false},
		{"Complex expression 2", "(10 + 5) * (2in - 1cm)", (10 + 5) * (uc.FromInch(2) - uc.FromCentimeter(1)), false},
		{"Division in expression", "100px / (2 + 3)", 100 / (2 + 3), false},
		{"Missing unit", "10", 10, false},
		{"Missing unit in second term", "1in + 2", uc.FromInch(1) + uc.FromInch(2), false},

		// Complex expressions
		{"Complex expression 3", "(1in + 2cm) * 3 - (4mm + 5pt) * 2", (uc.FromInch(1)+uc.FromCentimeter(2))*3 - (uc.FromMillimeter(4)+uc.FromPoint(5))*2, false},
		{"Complex expression 4", "10 * (5em - 2rem) + 3pc / (1 + 0.5)", 10*(uc.FromEm(5)-uc.FromEm(2)) + uc.FromPica(3)/(1+0.5), false},
		{"Complex expression 5", "(((1in + 2cm) * 3) - 4mm) - (5pt / 2)", (((uc.FromInch(1) + uc.FromCentimeter(2)) * 3) - uc.FromMillimeter(4)) - (uc.FromPoint(5) / 2), false},

		// Whitespace handling
		{"Extra whitespace", "  10in  +  5cm  ", uc.FromInch(10) + uc.FromCentimeter(5), false},
		{"No whitespace", "1in+2cm", uc.FromInch(1) + uc.FromCentimeter(2), false},
		{"Mixed whitespace", "3mm *\t4 + \n5px", uc.FromMillimeter(3)*4 + uc.FromPixel(5), false},

		// Unit variations
		{"Uppercase units", "1IN + 2CM", uc.FromInch(1) + uc.FromCentimeter(2), false},
		{"Mixed case units", "3Mm * 4", uc.FromMillimeter(3) * 4, false},
		{"Mixed case relative units", "5Em - 2rEm", uc.FromEm(5) - uc.FromEm(2), false},

		// Edge cases
		{"Zero value", "0in", 0, false},
		{"Very small value", "0.0000001mm", uc.FromMillimeter(0.0000001), false},
		{"Very large value", "9999999px", 9999999, false},
		{"Leading decimal", ".5in", uc.FromInch(0.5), false},
		{"Trailing decimal", "1.", 1, false},
		{"Single parentheses", "(1in)", uc.FromInch(1), false},
		{"Multiple parentheses", "((((1in))))", uc.FromInch(1), false},
		{"Multiple negatives", "-(-(-1in))", uc.FromInch(-1), false},

		// Failure cases
		{"Empty string", "", 0, true},
		{"Missing number", "in", 0, true},
		{"Invalid unit", "10kg", 0, true},
		{"Incomplete expression", "1in + ", 0, true},
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

		// New tests for unit handling
		{"Unit after parentheses", "(1 + 2) in", uc.FromInch(3), false},
		{"Unit after parentheses 2a", "(1 + 2) in + (2 + 3) cm", uc.FromInch(3) + uc.FromCentimeter(5), false},
		{"Unit after parentheses 2b", "(2 + 3) cm + 4", uc.FromCentimeter(5) + uc.FromCentimeter(4), false},
		{"Unit after parentheses 2b", "(2 + 3) cm + 4 px", uc.FromCentimeter(5) + uc.FromPixel(4), false},
		{"Unit after parentheses 2", "(1 + 2) in + (2 + 3) cm + 4 cm", uc.FromInch(3) + uc.FromCentimeter(2+3) + uc.FromCentimeter(4), false},
		{"Unit after parentheses 3", "(1 + 2) + (2 + 3) cm", uc.FromCentimeter(1+2) + uc.FromCentimeter(2+3), false},
		{"Unit after complex expression", "1 + 2 in", uc.FromInch(3), false},
		{"Mixed units in expression", "1 px + 2 in", uc.FromInch(2) + 1, false},
		{"Fraction with unit", "1/12 in", uc.FromInch(1.0 / 12.0), false},
		{"Complex expression with mixed units", "(1 in + 2 cm) * 3", (uc.FromInch(1) + uc.FromCentimeter(2)) * 3, false},
		{"Unit after each term", "1 in + 2 cm - 3 mm", uc.FromInch(1) + uc.FromCentimeter(2) - uc.FromMillimeter(3), false},

		// New tests for unit handling restrictions
		{"Multiplication with unitless", "2in * 3", uc.FromInch(6), false},
		{"Multiplication with unitless reversed", "3 * 2in", uc.FromInch(6), false},
		{"Division with unitless divisor", "6in / 2", uc.FromInch(3), false},

		// Error cases
		{"Multiplication with two units", "2in * 3cm", 0, true},
		{"Division with unit divisor", "6in / 2cm", 0, true},
		{"Complex expression with invalid multiplication", "(2in + 3cm) * 4mm", 0, true},
		{"Complex expression with invalid division", "((2in + 3cm) * 4) / 5mm", 0, true},
	}

	parser := &ExpressionParser{PPI: 96}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected an error, but got none, got %v instead", result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					expectedPixels := tt.expected
					resultPixels, err := parser.unitConverter.ToPixels(result.Val, result.Unit)
					if err != nil {
						t.Errorf("Error converting result value to pixels: %v", err)
					}
					if !almostEqual(resultPixels, expectedPixels, 1e-9) {
						t.Errorf("Expected %v pixels, but got %v pixels", expectedPixels, resultPixels)
					}
				}
			}
		})
	}
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}
