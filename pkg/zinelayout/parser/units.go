package parser

import (
	"fmt"
	"strings"
)

type UnitConverter struct {
	PPI float64
}

// ToPixels converts a value from a given unit to pixels
func (uc *UnitConverter) ToPixels(value float64, unit string) (float64, error) {
	switch strings.ToLower(unit) {
	case "mm":
		return uc.FromMillimeter(value), nil
	case "cm":
		return uc.FromCentimeter(value), nil
	case "in":
		return uc.FromInch(value), nil
	case "pc":
		return uc.FromPica(value), nil
	case "pt":
		return uc.FromPoint(value), nil
	case "px":
		return value, nil
	case "em", "rem":
		return uc.FromEm(value), nil
	case "":
		return value, nil // Assume pixels if no unit is specified
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}

// FromPixels converts a pixel value to the specified unit
func (uc *UnitConverter) FromPixels(pixels float64, unit string) (float64, error) {
	switch strings.ToLower(unit) {
	case "mm":
		return uc.ToMillimeter(pixels), nil
	case "cm":
		return uc.ToCentimeter(pixels), nil
	case "in":
		return uc.ToInch(pixels), nil
	case "pc":
		return uc.ToPica(pixels), nil
	case "pt":
		return uc.ToPoint(pixels), nil
	case "px":
		return pixels, nil
	case "em", "rem":
		return uc.ToEm(pixels), nil
	case "":
		return pixels, nil // Assume pixels if no unit is specified
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}

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

func (uc *UnitConverter) FromCentimeter(cm float64) float64 {
	return cm * uc.PPI / 2.54
}

func (uc *UnitConverter) ToCentimeter(pixels float64) float64 {
	return pixels * 2.54 / uc.PPI
}

func (uc *UnitConverter) FromEm(em float64) float64 {
	// Assuming 1em = 16px at 96 PPI, but this might need to be configurable
	return em * 16 * (uc.PPI / 96)
}

func (uc *UnitConverter) ToEm(pixels float64) float64 {
	// Assuming 1em = 16px at 96 PPI, but this might need to be configurable
	return pixels / (16 * (uc.PPI / 96))
}

func (uc *UnitConverter) FromPixel(px float64) float64 {
	return px
}

func (uc *UnitConverter) ToPixel(pixels float64) float64 {
	return pixels
}
