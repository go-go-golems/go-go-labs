package types

// FixerState represents the state of fixer usage
type FixerState struct {
	CapacityPerLiter int `json:"capacity_per_liter"`
	UsedRolls        int `json:"used_rolls"`
	TotalCapacity    int `json:"total_capacity"`
}

// RemainingCapacity returns the remaining capacity of the fixer
func (fs *FixerState) RemainingCapacity() int {
	return fs.TotalCapacity - fs.UsedRolls
}

// CanProcess returns true if the fixer can process the given number of rolls
func (fs *FixerState) CanProcess(rolls int) bool {
	return fs.RemainingCapacity() >= rolls
}

// UseFixer uses the fixer for the given number of rolls
func (fs *FixerState) UseFixer(rolls int) {
	fs.UsedRolls += rolls
}
