package parser

import (
	"math"
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
		result, err := uc.ToPixels(test.input)
		if err != nil {
			t.Errorf("Error converting %s: %v", test.input, err)
		}
		if !floatEquals(result, test.expected, epsilon) {
			t.Errorf("ToPixels(%s) = %f; want %f", test.input, result, test.expected)
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
		}
		if result != test.expected {
			t.Errorf("FromPixels(%f, %s) = %s; want %s", test.input, test.unit, result, test.expected)
		}
	}
}

func TestUnknownUnit(t *testing.T) {
	uc := UnitConverter{PPI: 300}
	_, err := uc.ToPixels("10unknown")
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

	// Test zero value
	result, err := uc.ToPixels("0mm")
	if err != nil {
		t.Errorf("Error converting 0mm: %v", err)
	}
	if !floatEquals(result, 0, epsilon) {
		t.Errorf("ToPixels(0mm) = %f; want 0", result)
	}

	// Test negative value
	result, err = uc.ToPixels("-10mm")
	if err != nil {
		t.Errorf("Error converting -10mm: %v", err)
	}
	if !floatEquals(result, -118.11, epsilon) {
		t.Errorf("ToPixels(-10mm) = %f; want -118.11", result)
	}

	// Additional edge cases
	_, err = uc.ToPixels("")
	if err == nil {
		t.Error("Expected error for empty string, got nil")
	}

	_, err = uc.ToPixels("mm")
	if err == nil {
		t.Error("Expected error for string with only units, got nil")
	}

	_, err = uc.ToPixels("100")
	if err == nil {
		t.Error("Expected error for string with only numbers, got nil")
	}

	// Add new tests for spaces
	result, err = uc.ToPixels("10 mm")
	if err != nil {
		t.Errorf("Unexpected error for string with space between number and unit: %v", err)
	}
	if !floatEquals(result, 118.11, epsilon) {
		t.Errorf("ToPixels('10 mm') = %f; want 118.11", result)
	}

	result, err = uc.ToPixels(" 10 mm ")
	if err != nil {
		t.Errorf("Unexpected error for string with spaces before and after: %v", err)
	}
	if !floatEquals(result, 118.11, epsilon) {
		t.Errorf("ToPixels(' 10 mm ') = %f; want 118.11", result)
	}

	_, err = uc.ToPixels("10MM")
	if err == nil {
		t.Errorf("Should have failed for string with uppercase units: %v", err)
	}

	_, err = uc.ToPixels("10Mm")
	if err == nil {
		t.Errorf("Should have failed for string with mixed case units: %v", err)
	}
}

func TestBoundaryValues(t *testing.T) {
	uc := UnitConverter{PPI: 300}

	// Test very large value
	result, err := uc.ToPixels("1000000mm")
	if err != nil {
		t.Errorf("Error converting 1000000mm: %v", err)
	}
	if !floatEquals(result, 11811023.622047244, epsilon) {
		t.Errorf("ToPixels(1000000mm) = %f; want 11811023.622047244", result)
	}

	// Test very small value
	result, err = uc.ToPixels("0.0001mm")
	if err != nil {
		t.Errorf("Error converting 0.0001mm: %v", err)
	}
	if !floatEquals(result, 0.0011811, epsilon) {
		t.Errorf("ToPixels(0.0001mm) = %f; want 0.0011811", result)
	}

	// Test maximum float64 value
	result, err = uc.ToPixels("1.7976931348623157e+308mm")
	if err != nil {
		t.Errorf("Error converting max float64 value: %v", err)
	}
	if !math.IsInf(result, 1) {
		t.Errorf("ToPixels(max float64) = %f; want %f", result, math.Inf(1))
	}

	// Test minimum float64 value
	result, err = uc.ToPixels("4.9406564584124654e-324mm")
	if err != nil {
		t.Errorf("Error converting min float64 value: %v", err)
	}
	if !floatEquals(result, 0, epsilon) {
		t.Errorf("ToPixels(min float64) = %f; want 0", result)
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
		result, err := uc.ToPixels(test.input)
		if err != nil {
			t.Errorf("Error converting %s: %v", test.input, err)
		}
		if !floatEquals(result, test.expected, epsilon) {
			t.Errorf("ToPixels(%s) = %f; want %f", test.input, result, test.expected)
		}
	}
}

func TestInvalidArithmeticExpressions(t *testing.T) {
	invalidTests := []string{
		"(10 + 5)mm + 2mm",
		"10mm + 5mm",
		"(10 + 5)mm px",
		"(10 + 5mm)",
		"(10 + 5)m m",
		"(10 + 5)",
		"mm",
		"(10 + 5)invalid",
	}

	uc := UnitConverter{PPI: 300}
	for _, test := range invalidTests {
		_, err := uc.ToPixels(test)
		if err == nil {
			t.Errorf("Expected error for invalid input %s, but got nil", test)
		}
	}
}