package models

import (
	"errors"
)

// AppState represents the different states of the application
type AppState int

const (
	StateMainScreen AppState = iota
	StateFilmSelection
	StateEISelection
	StateRollSelection
	StateMixedRollInput
	StateCalculatedScreen
	StateFixerTracking
	StateSettings
)

// String returns the string representation of the state
func (s AppState) String() string {
	switch s {
	case StateMainScreen:
		return "MainScreen"
	case StateFilmSelection:
		return "FilmSelection"
	case StateEISelection:
		return "EISelection"
	case StateRollSelection:
		return "RollSelection"
	case StateMixedRollInput:
		return "MixedRollInput"
	case StateCalculatedScreen:
		return "CalculatedScreen"
	case StateFixerTracking:
		return "FixerTracking"
	case StateSettings:
		return "Settings"
	default:
		return "Unknown"
	}
}

// AppModel represents the complete application state
type AppModel struct {
	CurrentState AppState
	Film         *Film
	EIRating     int
	Dilution     string
	Film35mm     int
	Film120mm    int
	TankSize     int
	FixerTracker *FixerTracker
	Error        error
}

// NewAppModel creates a new application model
func NewAppModel() *AppModel {
	return &AppModel{
		CurrentState: StateMainScreen,
		Dilution:     "1+9", // Default dilution
		FixerTracker: NewFixerTracker(),
	}
}

// SetFilm sets the selected film
func (m *AppModel) SetFilm(filmID string) error {
	film, ok := FilmDatabase[filmID]
	if !ok {
		return errors.New("invalid film ID")
	}
	m.Film = &film
	m.EIRating = 0 // Reset EI rating when film changes
	return nil
}

// SetEI sets the EI rating
func (m *AppModel) SetEI(ei int) error {
	if m.Film == nil {
		return errors.New("no film selected")
	}
	if !m.Film.HasEI(ei) {
		return errors.New("invalid EI rating for selected film")
	}
	m.EIRating = ei
	return nil
}

// SetRolls sets the number of rolls
func (m *AppModel) SetRolls(film35mm, film120mm int) {
	m.Film35mm = film35mm
	m.Film120mm = film120mm
	m.TankSize = CalculateTankSize(film35mm, film120mm)
}

// IsComplete checks if all required selections are made
func (m *AppModel) IsComplete() bool {
	return m.Film != nil && m.EIRating > 0 && m.TankSize > 0
}

// GetDevelopmentTime returns the development time for current selections
func (m *AppModel) GetDevelopmentTime() (string, error) {
	if !m.IsComplete() {
		return "", errors.New("incomplete setup")
	}
	
	timeStr, ok := m.Film.Times20C[m.Dilution][m.EIRating]
	if !ok {
		return "", errors.New("no time data for current combination")
	}
	
	return timeStr, nil
}

// GetChemicalVolumes returns the required chemical volumes
func (m *AppModel) GetChemicalVolumes() map[string]map[string]int {
	volumes := make(map[string]map[string]int)
	
	if m.TankSize == 0 {
		return volumes
	}
	
	// ILFOSOL 3
	conc, water := CalculateChemicalVolumes(m.TankSize, m.Dilution)
	volumes["ilfosol_3"] = map[string]int{
		"concentrate": conc,
		"water":       water,
	}
	
	// ILFOSTOP
	conc, water = CalculateChemicalVolumes(m.TankSize, "1+19")
	volumes["ilfostop"] = map[string]int{
		"concentrate": conc,
		"water":       water,
	}
	
	// SPRINT FIXER
	conc, water = CalculateChemicalVolumes(m.TankSize, "1+4")
	volumes["sprint_fixer"] = map[string]int{
		"concentrate": conc,
		"water":       water,
	}
	
	return volumes
}

// GetRollsUsed returns the number of rolls that will be used
func (m *AppModel) GetRollsUsed() int {
	return m.Film35mm + m.Film120mm
}

// CanUseFixer checks if the fixer can process the current batch
func (m *AppModel) CanUseFixer() bool {
	rollsUsed := m.GetRollsUsed()
	return m.FixerTracker.CanProcess(rollsUsed)
}

// UseFixer marks the fixer as used for the current batch
func (m *AppModel) UseFixer() {
	rollsUsed := m.GetRollsUsed()
	m.FixerTracker.UseRolls(rollsUsed)
}

// GetRollDescription returns a human-readable description of the roll configuration
func (m *AppModel) GetRollDescription() string {
	return FormatRollDescription(m.Film35mm, m.Film120mm)
}

// Reset resets the application state
func (m *AppModel) Reset() {
	m.CurrentState = StateMainScreen
	m.Film = nil
	m.EIRating = 0
	m.Film35mm = 0
	m.Film120mm = 0
	m.TankSize = 0
	m.Error = nil
}

// Common errors
var (
	ErrInvalidCombination = errors.New("invalid film/EI/dilution combination")
	ErrIncompleteSetup    = errors.New("incomplete setup")
)
