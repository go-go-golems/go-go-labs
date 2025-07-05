package models

import "time"

// ChemicalType represents the type of chemical
type ChemicalType string

const (
	ChemicalTypeOneShot  ChemicalType = "one_shot"
	ChemicalTypeReusable ChemicalType = "reusable"
)

// Chemical represents a chemical used in film development
type Chemical struct {
	Name      string
	Dilution  string
	Time      time.Duration
	Type      ChemicalType
	Capacity  string // Human readable capacity description
}

// ChemicalDatabase contains all supported chemicals
var ChemicalDatabase = map[string]Chemical{
	"ilfosol_3": {
		Name:     "ILFOSOL 3",
		Dilution: "1+9", // Default dilution
		Time:     0,     // Time depends on film/EI combination
		Type:     ChemicalTypeOneShot,
		Capacity: "Single use",
	},
	"ilfostop": {
		Name:     "ILFOSTOP",
		Dilution: "1+19",
		Time:     10 * time.Second,
		Type:     ChemicalTypeReusable,
		Capacity: "15 rolls per liter",
	},
	"sprint_fixer": {
		Name:     "SPRINT FIXER",
		Dilution: "1+4",
		Time:     2*time.Minute + 30*time.Second,
		Type:     ChemicalTypeReusable,
		Capacity: "24 rolls per liter",
	},
}

// DilutionRatio represents a dilution ratio
type DilutionRatio struct {
	Concentrate int
	Water       int
}

// ParseDilution parses a dilution string like "1+9" into a DilutionRatio
func ParseDilution(dilution string) DilutionRatio {
	switch dilution {
	case "1+9":
		return DilutionRatio{Concentrate: 1, Water: 9}
	case "1+14":
		return DilutionRatio{Concentrate: 1, Water: 14}
	case "1+19":
		return DilutionRatio{Concentrate: 1, Water: 19}
	case "1+4":
		return DilutionRatio{Concentrate: 1, Water: 4}
	default:
		return DilutionRatio{Concentrate: 1, Water: 9} // Default
	}
}

// CalculateChemicalVolumes calculates the required volumes for a given tank size
func CalculateChemicalVolumes(tankSize int, dilution string) (concentrate, water int) {
	ratio := ParseDilution(dilution)
	totalParts := ratio.Concentrate + ratio.Water
	
	concentrate = (tankSize * ratio.Concentrate) / totalParts
	water = tankSize - concentrate
	
	return concentrate, water
}

// FixerTracker tracks fixer usage
type FixerTracker struct {
	CapacityPerLiter int
	UsedRolls        int
	TankSizeML       int
}

// NewFixerTracker creates a new fixer tracker
func NewFixerTracker() *FixerTracker {
	return &FixerTracker{
		CapacityPerLiter: 24,
		UsedRolls:        0,
		TankSizeML:       1000, // Default 1L
	}
}

// GetRemainingCapacity returns the remaining fixer capacity
func (ft *FixerTracker) GetRemainingCapacity() int {
	totalCapacity := (ft.TankSizeML * ft.CapacityPerLiter) / 1000
	return totalCapacity - ft.UsedRolls
}

// CanProcess checks if the fixer can process the given number of rolls
func (ft *FixerTracker) CanProcess(rolls int) bool {
	return ft.GetRemainingCapacity() >= rolls
}

// UseRolls marks the specified number of rolls as used
func (ft *FixerTracker) UseRolls(rolls int) {
	ft.UsedRolls += rolls
}

// Reset resets the fixer tracker (when replacing fixer)
func (ft *FixerTracker) Reset() {
	ft.UsedRolls = 0
}

// GetTotalCapacity returns the total fixer capacity
func (ft *FixerTracker) GetTotalCapacity() int {
	return (ft.TankSizeML * ft.CapacityPerLiter) / 1000
}
