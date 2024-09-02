package helpers

import (
	"fmt"
	"strconv"
	"strings"
)

type UnitConverter struct {
	PPI float64
}

func (uc *UnitConverter) ToPixels(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty input")
	}

	var number float64
	var unit string
	var err error

	// Split the value into number and unit
	for i, r := range value {
		if (r < '0' || r > '9') && r != '.' && r != '-' {
			number, err = strconv.ParseFloat(value[:i], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number format: %s", value[:i])
			}
			unit = strings.ToLower(strings.TrimSpace(value[i:]))
			break
		}
	}

	if unit == "" {
		return 0, fmt.Errorf("missing unit")
	}

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
