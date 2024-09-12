package helpers

import (
	"fmt"
)

type UnitConverter struct {
	PPI float64
}

// ToPixels converts a unit expression to pixels
func (uc *UnitConverter) ToPixels(value string) (float64, error) {
	parser := &ExpressionParser{PPI: uc.PPI}
	return parser.Parse(value)
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
