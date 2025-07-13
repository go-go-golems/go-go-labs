package types

import "fmt"

// RollSetup represents the roll configuration for development
type RollSetup struct {
	Format35mm  int `json:"format_35mm"`
	Format120mm int `json:"format_120mm"`
	TotalVolume int `json:"total_volume"`
}

// String returns a string representation of the roll setup
func (rs *RollSetup) String() string {
	if rs.Format35mm > 0 && rs.Format120mm > 0 {
		return fmt.Sprintf("%d×35mm + %d×120mm", rs.Format35mm, rs.Format120mm)
	} else if rs.Format35mm > 0 {
		return fmt.Sprintf("%d×35mm", rs.Format35mm)
	} else if rs.Format120mm > 0 {
		return fmt.Sprintf("%d×120mm", rs.Format120mm)
	}
	return "No rolls"
}

// TotalRolls returns the total number of rolls
func (rs *RollSetup) TotalRolls() int {
	return rs.Format35mm + rs.Format120mm
}
