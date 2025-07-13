package state

import "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/types"

// ApplicationState represents the overall state of the application
type ApplicationState struct {
	SelectedFilm *types.Film                 `json:"selected_film"`
	SelectedEI   int                         `json:"selected_ei"`
	RollSetup    *types.RollSetup            `json:"roll_setup"`
	Dilution     string                      `json:"dilution"`
	Calculations []types.DilutionCalculation `json:"calculations"`
	FixerState   *types.FixerState           `json:"fixer_state"`
	TimerState   *types.TimerState           `json:"timer_state"`
	FilmDB       *types.FilmDatabase         `json:"film_db"`
	TankDB       *types.TankDatabase         `json:"tank_db"`
	ChemicalDB   *types.ChemicalDatabase     `json:"chemical_db"`
}

// NewApplicationState creates a new application state
func NewApplicationState() *ApplicationState {
	return &ApplicationState{
		SelectedFilm: nil,
		SelectedEI:   0,
		RollSetup:    nil,
		Dilution:     "1+9",
		Calculations: []types.DilutionCalculation{},
		FixerState:   &types.FixerState{CapacityPerLiter: 24, UsedRolls: 0, TotalCapacity: 24},
		TimerState:   nil,
		FilmDB:       types.NewFilmDatabase(),
		TankDB:       types.NewTankDatabase(),
		ChemicalDB:   types.NewChemicalDatabase(),
	}
}

// IsComplete returns true if all required fields are set
func (as *ApplicationState) IsComplete() bool {
	return as.SelectedFilm != nil && as.SelectedEI > 0 && as.RollSetup != nil
}

// GetDevelopmentTime returns the development time for the selected film and EI
func (as *ApplicationState) GetDevelopmentTime() string {
	if as.SelectedFilm == nil || as.SelectedEI == 0 {
		return "--:--"
	}

	if dilutionTimes, ok := as.SelectedFilm.Times20C[as.Dilution]; ok {
		if time, ok := dilutionTimes[as.SelectedEI]; ok {
			return time
		}
	}

	return "--:--"
}

// CalculateChemicals calculates the chemical dilutions based on current state
func (as *ApplicationState) CalculateChemicals() {
	if !as.IsComplete() {
		return
	}

	totalVolume := as.RollSetup.TotalVolume
	developmentTime := as.GetDevelopmentTime()

	// Calculate dilutions for each chemical
	as.Calculations = []types.DilutionCalculation{
		types.CalculateDilution("ILFOSOL 3", as.Dilution, totalVolume, developmentTime),
		types.CalculateDilution("ILFOSTOP", "1+19", totalVolume, "0:10"),
		types.CalculateDilution("SPRINT FIXER", "1+4", totalVolume, "2:30"),
	}

	// Create timer state if calculations are available
	if len(as.Calculations) > 0 {
		as.TimerState = types.NewTimerState(as.Calculations)
	}
}
