package parser

import (
	"math"
	"strconv"
	"testing"
)

const epsilon = 1e-1

func TestToPixels(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		ppi      float64
	}{
		{"10mm", 118.11, 300},
		{"1in", 300, 300},
		{"1pc", 50, 300},
		{"100px", 100, 300},
		{"10cm", 1181.1, 300},
		{"0.5in", 150, 300},
		{"1000mm", 11811, 300},
		{"0.1mm", 1.1811, 300},
		// Additional test cases
		{"10cm", 1181.1, 300},
		{"12pt", 50, 300},
		{"1em", 16, 96},  // Assuming 1em = 16px at 96 PPI
		{"1rem", 16, 96}, // Assuming 1rem = 16px at 96 PPI
		{"0.5in", 150, 300},
		{"1000mm", 11811, 300},
		{"0.1mm", 1.1811, 300},
	}

	for _, test := range tests {
		uc := UnitConverter{PPI: test.ppi}
		parser := ExpressionParser{PPI: test.ppi, unitConverter: &uc}
		result, err := parser.Parse(test.input)
		if err != nil {
			t.Errorf("Error parsing %s: %v", test.input, err)
			continue
		}
		pixels, err := uc.ToPixels(result.Val, result.Unit)
		if err != nil {
			t.Errorf("Error converting %s to pixels: %v", test.input, err)
			continue
		}
		if !floatEquals(pixels, test.expected, epsilon) {
			t.Errorf("ToPixels(%s) = %f; want %f", test.input, pixels, test.expected)
		}
	}
}

func TestFromPixels(t *testing.T) {
	tests := []struct {
		input    float64
		unit     string
		expected string
		ppi      float64
	}{
		{118.11, "mm", "10.00mm", 300},
		{300, "in", "1.00in", 300},
		{50, "pc", "1.00pc", 300},
		{100, "px", "100.00px", 300},
		{1181.1, "mm", "100.00mm", 300},
		{150, "in", "0.50in", 300},
		{11811, "mm", "1000.00mm", 300},
		{1.1811, "mm", "0.10mm", 300},
	}

	for _, test := range tests {
		uc := UnitConverter{PPI: test.ppi}
		result, err := uc.FromPixels(test.input, test.unit)
		if err != nil {
			t.Errorf("Error converting %f to %s: %v", test.input, test.unit, err)
			continue
		}
		if !floatEquals(result, parseFloat(test.expected), epsilon) {
			t.Errorf("FromPixels(%f, %s) = %f; want %s", test.input, test.unit, result, test.expected)
		}
	}
}

func TestUnknownUnit(t *testing.T) {
	uc := UnitConverter{PPI: 300}
	parser := ExpressionParser{PPI: 300, unitConverter: &uc}
	_, err := parser.Parse("10unknown")
	if err == nil {
		t.Error("Expected error for unknown unit, got nil")
	}

	_, err = uc.FromPixels(100, "unknown")
	if err == nil {
		t.Error("Expected error for unknown unit, got nil")
	}
}

func TestEdgeCases(t *testing.T) {
	uc := UnitConverter{PPI: 300}
	parser := ExpressionParser{PPI: 300, unitConverter: &uc}

	// Test zero value
	result, err := parser.Parse("0mm")
	if err != nil {
		t.Errorf("Error parsing 0mm: %v", err)
	}
	pixels, _ := uc.ToPixels(result.Val, result.Unit)
	if !floatEquals(pixels, 0, epsilon) {
		t.Errorf("ToPixels(0mm) = %f; want 0", pixels)
	}

	// Test negative value
	result, err = parser.Parse("-10mm")
	if err != nil {
		t.Errorf("Error parsing -10mm: %v", err)
	}
	pixels, _ = uc.ToPixels(result.Val, result.Unit)
	if !floatEquals(pixels, -118.11, epsilon) {
		t.Errorf("ToPixels(-10mm) = %f; want -118.11", pixels)
	}

	// Additional edge cases
	_, err = parser.Parse("")
	if err == nil {
		t.Error("Expected error for empty string, got nil")
	}

	_, err = parser.Parse("mm")
	if err == nil {
		t.Error("Expected error for string with only units, got nil")
	}

	_, err = parser.Parse("100")
	if err == nil {
		t.Error("Expected error for string with only numbers, got nil")
	}

	// Add new tests for spaces
	result, err = parser.Parse("10 mm")
	if err != nil {
		t.Errorf("Unexpected error for string with space between number and unit: %v", err)
	}
	pixels, _ = uc.ToPixels(result.Val, result.Unit)
	if !floatEquals(pixels, 118.11, epsilon) {
		t.Errorf("ToPixels('10 mm') = %f; want 118.11", pixels)
	}

	result, err = parser.Parse(" 10 mm ")
	if err != nil {
		t.Errorf("Unexpected error for string with spaces before and after: %v", err)
	}
	pixels, _ = uc.ToPixels(result.Val, result.Unit)
	if !floatEquals(pixels, 118.11, epsilon) {
		t.Errorf("ToPixels(' 10 mm ') = %f; want 118.11", pixels)
	}

	_, err = parser.Parse("10MM")
	if err == nil {
		t.Errorf("Should have failed for string with uppercase units: %v", err)
	}

	_, err = parser.Parse("10Mm")
	if err == nil {
		t.Errorf("Should have failed for string with mixed case units: %v", err)
	}
}

func TestBoundaryValues(t *testing.T) {
	uc := UnitConverter{PPI: 300}
	parser := ExpressionParser{PPI: 300, unitConverter: &uc}

	// Test very large value
	result, err := parser.Parse("1000000mm")
	if err != nil {
		t.Errorf("Error parsing 1000000mm: %v", err)
	}
	pixels, _ := uc.ToPixels(result.Val, result.Unit)
	if !floatEquals(pixels, 11811023.622047244, epsilon) {
		t.Errorf("ToPixels(1000000mm) = %f; want 11811023.622047244", pixels)
	}

	// Test very small value
	result, err = parser.Parse("0.0001mm")
	if err != nil {
		t.Errorf("Error parsing 0.0001mm: %v", err)
	}
	pixels, _ = uc.ToPixels(result.Val, result.Unit)
	if !floatEquals(pixels, 0.0011811, epsilon) {
		t.Errorf("ToPixels(0.0001mm) = %f; want 0.0011811", pixels)
	}

	// Test maximum float64 value
	result, err = parser.Parse("1.7976931348623157e+308mm")
	if err != nil {
		t.Errorf("Error parsing max float64 value: %v", err)
	}
	pixels, _ = uc.ToPixels(result.Val, result.Unit)
	if !math.IsInf(pixels, 1) {
		t.Errorf("ToPixels(max float64) = %f; want %f", pixels, math.Inf(1))
	}

	// Test minimum float64 value
	result, err = parser.Parse("4.9406564584124654e-324mm")
	if err != nil {
		t.Errorf("Error parsing min float64 value: %v", err)
	}
	pixels, _ = uc.ToPixels(result.Val, result.Unit)
	if !floatEquals(pixels, 0, epsilon) {
		t.Errorf("ToPixels(min float64) = %f; want 0", pixels)
	}
}

// Helper function for float comparison
func floatEquals(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestArithmeticExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		ppi      float64
	}{
		{"(10 + 5)mm", 177.165, 300},
		{"(20 - 5)mm", 177.165, 300},
		{"(4 * 3)mm", 141.732, 300},
		{"(12 / 3)mm", 47.244, 300},
		{"(10 + 5 * 2)mm", 236.22, 300},
		{"((10 + 5) * 2)mm", 354.33, 300},
		{"(10 + (5 * 2))mm", 236.22, 300},
		{"(10 - 5 + 3)mm", 94.488, 300},
		{"(2^3 + 1)mm", 106.299, 300},
		{"(10 % 3)mm", 11.811, 300},
		{"(10.5 + 1.5)cm", 1417.32, 300},
		{"(1 + 0.5)in", 450, 300},
		{"(20 + 4)pt", 100, 300},
		{"(100 / 2)px", 50, 300},
		{"(1 + 0.5)em", 24, 96},
		{"(50 + 25)%", 12, 96},
	}

	for _, test := range tests {
		uc := UnitConverter{PPI: test.ppi}
		parser := ExpressionParser{PPI: test.ppi, unitConverter: &uc}
		result, err := parser.Parse(test.input)
		if err != nil {
			t.Errorf("Error parsing %s: %v", test.input, err)
			continue
		}
		pixels, err := uc.ToPixels(result.Val, result.Unit)
		if err != nil {
			t.Errorf("Error converting %s to pixels: %v", test.input, err)
			continue
		}
		if !floatEquals(pixels, test.expected, epsilon) {
			t.Errorf("ToPixels(%s) = %f; want %f", test.input, pixels, test.expected)
		}
	}
}

func TestInvalidArithmeticExpressions(t *testing.T) {
	invalidTests := []string{
		"(10 + 5)mm + 2mm",
		"10mm + 5mm",
		"(10 + 5)mm px",
		"(10 + 5)m m",
		"(10 + 5)",
		"mm",
		"(10 + 5)invalid",
	}

	uc := UnitConverter{PPI: 300}
	parser := ExpressionParser{PPI: 300, unitConverter: &uc}
	for _, test := range invalidTests {
		_, err := parser.Parse(test)
		if err == nil {
			t.Errorf("Expected error for invalid input %s, but got nil", test)
		}
	}
}

// Helper function to parse float from string
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s[:len(s)-2], 64) // Remove unit (last 2 characters) before parsing
	return f
}
