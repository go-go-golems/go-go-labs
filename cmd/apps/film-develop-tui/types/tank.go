package types

// TankSize represents tank size requirements
type TankSize struct {
	Format string `json:"format"`
	Rolls  int    `json:"rolls"`
	Volume int    `json:"volume"` // in ml
}

// TankDatabase contains tank size information
type TankDatabase struct {
	Sizes map[string]map[int]int `json:"sizes"`
}

// GetTankSize returns the tank size for given format and roll count
func (td *TankDatabase) GetTankSize(format string, rolls int) (int, bool) {
	formatSizes, ok := td.Sizes[format]
	if !ok {
		return 0, false
	}
	size, ok := formatSizes[rolls]
	return size, ok
}
