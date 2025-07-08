package model

import (
	"fmt"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg/data"
)

// AppState represents the current state of the application
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

// FilmSetup holds the current film development setup
type FilmSetup struct {
	SelectedFilm     *data.Film
	SelectedEI       int
	SelectedDilution string
	Rolls35mm        int
	Rolls120mm       int
	TankSize         int
	DevelopmentTime  time.Duration
}

// FixerState tracks fixer usage
type FixerState struct {
	TotalCapacity   int // rolls per liter
	UsedRolls       int
	RemainingRolls  int
	Volume          int // in ml
}

// ChemicalCalc holds calculated chemical amounts
type ChemicalCalc struct {
	Name         string
	Dilution     string
	Concentrate  int // in ml
	Water        int // in ml
	Time         time.Duration
}

// AppData holds all application data
type AppData struct {
	CurrentState AppState
	FilmSetup    FilmSetup
	FixerState   FixerState
	Chemicals    []ChemicalCalc
	ErrorMessage string
}

// NewAppData creates a new AppData instance with default values
func NewAppData() *AppData {
	return &AppData{
		CurrentState: StateMainScreen,
		FilmSetup: FilmSetup{
			SelectedDilution: "1+9",
		},
		FixerState: FixerState{
			TotalCapacity:  data.ChemicalDatabase["sprint_fixer"].RollsPerL,
			RemainingRolls: data.ChemicalDatabase["sprint_fixer"].RollsPerL,
			Volume:         1000, // 1 liter
		},
		Chemicals: []ChemicalCalc{},
	}
}

// CalculateChemicals calculates the chemical amounts based on current setup
func (d *AppData) CalculateChemicals() {
	if d.FilmSetup.SelectedFilm == nil || d.FilmSetup.TankSize == 0 {
		d.Chemicals = []ChemicalCalc{}
		return
	}

	tankSize := d.FilmSetup.TankSize
	d.Chemicals = []ChemicalCalc{}

	// ILFOSOL 3
	ilfosol := data.ChemicalDatabase["ilfosol_3"]
	dilution := d.FilmSetup.SelectedDilution
	
	var concentrate, water int
	if dilution == "1+9" {
		concentrate = tankSize / 10
		water = tankSize - concentrate
	} else if dilution == "1+14" {
		concentrate = tankSize / 15
		water = tankSize - concentrate
	}
	
	// Get development time
	if times, exists := d.FilmSetup.SelectedFilm.Times20C[dilution]; exists {
		if devTime, exists := times[d.FilmSetup.SelectedEI]; exists {
			d.FilmSetup.DevelopmentTime = devTime
		}
	}
	
	d.Chemicals = append(d.Chemicals, ChemicalCalc{
		Name:        ilfosol.Name,
		Dilution:    dilution,
		Concentrate: concentrate,
		Water:       water,
		Time:        d.FilmSetup.DevelopmentTime,
	})

	// ILFOSTOP
	ilfostop := data.ChemicalDatabase["ilfostop"]
	stopConcentrate := tankSize / 20 // 1+19
	stopWater := tankSize - stopConcentrate
	
	d.Chemicals = append(d.Chemicals, ChemicalCalc{
		Name:        ilfostop.Name,
		Dilution:    "1+19",
		Concentrate: stopConcentrate,
		Water:       stopWater,
		Time:        ilfostop.Time,
	})

	// SPRINT FIXER
	fixer := data.ChemicalDatabase["sprint_fixer"]
	fixerConcentrate := tankSize / 5 // 1+4
	fixerWater := tankSize - fixerConcentrate
	
	d.Chemicals = append(d.Chemicals, ChemicalCalc{
		Name:        fixer.Name,
		Dilution:    "1+4",
		Concentrate: fixerConcentrate,
		Water:       fixerWater,
		Time:        fixer.Time,
	})
}

// UpdateTankSize calculates and updates the tank size based on current roll selection
func (d *AppData) UpdateTankSize() {
	if d.FilmSetup.Rolls35mm > 0 && d.FilmSetup.Rolls120mm > 0 {
		// Mixed batch
		d.FilmSetup.TankSize = data.CalculateCustomTankSize(d.FilmSetup.Rolls35mm, d.FilmSetup.Rolls120mm)
	} else if d.FilmSetup.Rolls35mm > 0 {
		// 35mm only
		d.FilmSetup.TankSize = data.GetTankSize("35mm", d.FilmSetup.Rolls35mm)
	} else if d.FilmSetup.Rolls120mm > 0 {
		// 120mm only
		d.FilmSetup.TankSize = data.GetTankSize("120mm", d.FilmSetup.Rolls120mm)
	} else {
		d.FilmSetup.TankSize = 0
	}
}

// UseFixer records fixer usage
func (d *AppData) UseFixer() {
	totalRolls := d.FilmSetup.Rolls35mm + d.FilmSetup.Rolls120mm
	if totalRolls > 0 && d.FixerState.RemainingRolls >= totalRolls {
		d.FixerState.UsedRolls += totalRolls
		d.FixerState.RemainingRolls -= totalRolls
	}
}

// GetRollsDescription returns a human-readable description of the current roll selection
func (d *AppData) GetRollsDescription() string {
	if d.FilmSetup.Rolls35mm > 0 && d.FilmSetup.Rolls120mm > 0 {
		return fmt.Sprintf("%dx 35mm + %dx 120mm", d.FilmSetup.Rolls35mm, d.FilmSetup.Rolls120mm)
	} else if d.FilmSetup.Rolls35mm > 0 {
		return fmt.Sprintf("%dx 35mm", d.FilmSetup.Rolls35mm)
	} else if d.FilmSetup.Rolls120mm > 0 {
		return fmt.Sprintf("%dx 120mm", d.FilmSetup.Rolls120mm)
	}
	return "--"
}

// GetCurrentRolls returns the total number of rolls being processed
func (d *AppData) GetCurrentRolls() int {
	return d.FilmSetup.Rolls35mm + d.FilmSetup.Rolls120mm
}
