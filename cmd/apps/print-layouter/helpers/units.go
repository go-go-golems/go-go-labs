package helpers

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
)

type UnitConverter struct {
	PPI float64
}

var unitRegex = regexp.MustCompile(`^(.*?)\s*([a-z%]+)$`)

func (uc *UnitConverter) ToPixels(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty input")
	}

	// Parse the expression and unit using regex
	matches := unitRegex.FindStringSubmatch(value)
	if len(matches) != 3 {
		return 0, fmt.Errorf("invalid input format: %s", value)
	}

	expression := strings.TrimSpace(matches[1])
	unit := strings.ToLower(matches[2])

	// Evaluate the expression
	env := map[string]interface{}{
		"ppi": uc.PPI,
	}

	result, err := expr.Eval(expression, env)
	if err != nil {
		return 0, fmt.Errorf("error evaluating expression: %v", err)
	}

	number, ok := cast.CastNumberInterfaceToFloat[float64](result)
	if !ok {
		return 0, fmt.Errorf("expression result is not a number")
	}

	// Convert to pixels based on the unit
	switch unit {
	case "mm":
		return number * uc.PPI / 25.4, nil
	case "cm":
		return number * uc.PPI / 2.54, nil
	case "in":
		return number * uc.PPI, nil
	case "pc":
		return number * uc.PPI / 6, nil
	case "pt":
		return number * uc.PPI / 72, nil
	case "px":
		return number, nil
	case "em", "rem":
		return number * 16 * (uc.PPI / 96), nil
	case "%":
		return number * 0.16 * (uc.PPI / 96), nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}

func (uc *UnitConverter) FromPixels(pixels float64, unit string) (string, error) {
	switch unit {
	case "mm":
		return fmt.Sprintf("%.2fmm", pixels*25.4/uc.PPI), nil
	case "in":
		return fmt.Sprintf("%.2fin", pixels/uc.PPI), nil
	case "pc":
		return fmt.Sprintf("%.2fpc", pixels*6/uc.PPI), nil
	case "px":
		return fmt.Sprintf("%.2fpx", pixels), nil
	default:
		return "", fmt.Errorf("unknown unit: %s", unit)
	}
}
