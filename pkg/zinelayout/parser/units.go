package parser

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

// New methods

func (uc *UnitConverter) FromPica(picas float64) float64 {
	return picas * uc.PPI / 6
}

func (uc *UnitConverter) ToPica(pixels float64) float64 {
	return pixels * 6 / uc.PPI
}

func (uc *UnitConverter) FromInch(inches float64) float64 {
	return inches * uc.PPI
}

func (uc *UnitConverter) ToInch(pixels float64) float64 {
	return pixels / uc.PPI
}

func (uc *UnitConverter) FromMillimeter(mm float64) float64 {
	return mm * uc.PPI / 25.4
}

func (uc *UnitConverter) ToMillimeter(pixels float64) float64 {
	return pixels * 25.4 / uc.PPI
}

func (uc *UnitConverter) FromPoint(points float64) float64 {
	return points * uc.PPI / 72
}

func (uc *UnitConverter) ToPoint(pixels float64) float64 {
	return pixels * 72 / uc.PPI
}

// Add these new methods
func (uc *UnitConverter) FromCentimeter(cm float64) float64 {
	return cm * uc.PPI / 2.54
}

func (uc *UnitConverter) FromEm(em float64) float64 {
	// Assuming 1em = 16px, but this might need to be configurable
	return em * 16
}

func (uc *UnitConverter) FromRem(rem float64) float64 {
	// Assuming 1rem = 16px, but this might need to be configurable
	return rem * 16
}

func (uc *UnitConverter) FromPixel(px float64) float64 {
	return px
}
