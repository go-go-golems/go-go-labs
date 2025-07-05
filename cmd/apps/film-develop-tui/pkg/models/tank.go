package models

import "fmt"

// TankSize represents the required tank size for different roll configurations
type TankSize struct {
	Film35mm int
	Film120mm int
	TotalML  int
}

// TankSizes defines the tank sizes for different roll configurations
var TankSizes = map[string]map[int]int{
	"35mm": {
		1: 300,
		2: 500,
		3: 600,
		4: 700,
		5: 800,
		6: 900,
	},
	"120mm": {
		1: 500,
		2: 700,
		3: 900,
		4: 1000,
		5: 1200,
		6: 1400,
	},
}

// CalculateTankSize calculates the required tank size for a given roll configuration
func CalculateTankSize(film35mm, film120mm int) int {
	var totalML int
	
	// Calculate contribution from 35mm rolls
	if film35mm > 0 {
		// For mixed batches, we need to calculate based on the largest format requirement
		if film120mm > 0 {
			// Mixed batch - use 120mm sizing as base and add 35mm requirements
			totalML = TankSizes["120mm"][film120mm]
			// Add additional capacity for 35mm rolls (rough estimation)
			totalML += film35mm * 50
		} else {
			// Pure 35mm batch
			totalML = TankSizes["35mm"][film35mm]
		}
	} else if film120mm > 0 {
		// Pure 120mm batch
		totalML = TankSizes["120mm"][film120mm]
	}
	
	return totalML
}

// GetMaxRolls returns the maximum number of rolls supported for each format
func GetMaxRolls() (int, int) {
	return 6, 6 // 35mm, 120mm
}

// FormatRollDescription returns a human-readable description of the roll configuration
func FormatRollDescription(film35mm, film120mm int) string {
	if film35mm == 0 && film120mm == 0 {
		return "No rolls"
	}
	
	if film35mm > 0 && film120mm > 0 {
		return fmt.Sprintf("%dx 35mm + %dx 120mm", film35mm, film120mm)
	}
	
	if film35mm > 0 {
		return fmt.Sprintf("%dx 35mm", film35mm)
	}
	
	return fmt.Sprintf("%dx 120mm", film120mm)
}
