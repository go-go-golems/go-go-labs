package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDuration parses a duration string in MM:SS format
func ParseDuration(timeStr string) (time.Duration, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %s", parts[0])
	}

	seconds, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid seconds: %s", parts[1])
	}

	return time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, nil
}

// FormatDuration formats a duration as MM:SS
func FormatDuration(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// CalculateDilution calculates the dilution amounts for a chemical
func CalculateDilution(chemical, dilution string, totalVolume int, time string) DilutionCalculation {
	// Parse dilution ratio (e.g., "1+9" means 1 part concentrate + 9 parts water)
	parts := strings.Split(dilution, "+")
	if len(parts) != 2 {
		return DilutionCalculation{
			Chemical:    chemical,
			Dilution:    dilution,
			TotalVolume: totalVolume,
			Concentrate: 0,
			Water:       0,
			Time:        time,
		}
	}

	concentrateParts, err1 := strconv.Atoi(parts[0])
	waterParts, err2 := strconv.Atoi(parts[1])

	if err1 != nil || err2 != nil {
		return DilutionCalculation{
			Chemical:    chemical,
			Dilution:    dilution,
			TotalVolume: totalVolume,
			Concentrate: 0,
			Water:       0,
			Time:        time,
		}
	}

	totalParts := concentrateParts + waterParts
	concentrate := (totalVolume * concentrateParts) / totalParts
	water := totalVolume - concentrate

	return DilutionCalculation{
		Chemical:    chemical,
		Dilution:    dilution,
		TotalVolume: totalVolume,
		Concentrate: concentrate,
		Water:       water,
		Time:        time,
	}
}

// CalculateMixedTankSize calculates the tank size for mixed formats
func CalculateMixedTankSize(format35mm, format120mm int, tankDB *TankDatabase) int {
	size35mm := 0
	size120mm := 0

	if format35mm > 0 {
		if s, ok := tankDB.GetTankSize("35mm", format35mm); ok {
			size35mm = s
		}
	}

	if format120mm > 0 {
		if s, ok := tankDB.GetTankSize("120mm", format120mm); ok {
			size120mm = s
		}
	}

	// For mixed formats, we need to calculate based on the larger requirement
	// This is a simplified calculation - in reality it would be more complex
	if size35mm > size120mm {
		return size35mm
	}
	return size120mm
}
