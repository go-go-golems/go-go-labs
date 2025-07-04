package models

// ApplicationState represents the complete application state
type ApplicationState struct {
	SelectedFilm *Film                 `json:"selected_film"`
	SelectedEI   int                   `json:"selected_ei"`
	RollSetup    *RollSetup            `json:"roll_setup"`
	Dilution     string                `json:"dilution"`
	Calculations []DilutionCalculation `json:"calculations"`
	FixerState   *FixerState           `json:"fixer_state"`
	TimerState   *TimerState           `json:"timer_state"`
	FilmDB       *FilmDatabase         `json:"film_db"`
	TankDB       *TankDatabase         `json:"tank_db"`
	ChemicalDB   *ChemicalDatabase     `json:"chemical_db"`
}

// NewApplicationState creates a new application state with default values
func NewApplicationState() *ApplicationState {
	return &ApplicationState{
		Dilution: "1+9",
		FixerState: &FixerState{
			CapacityPerLiter: 24,
			UsedRolls:        0,
			TotalCapacity:    24,
		},
		FilmDB:     NewFilmDatabase(),
		TankDB:     NewTankDatabase(),
		ChemicalDB: NewChemicalDatabase(),
	}
}

// IsComplete checks if the application state is complete for calculation
func (as *ApplicationState) IsComplete() bool {
	return as.SelectedFilm != nil && as.SelectedEI > 0 && as.RollSetup != nil && as.RollSetup.TotalRolls() > 0
}

// GetDevelopmentTime returns the development time for the current selection
func (as *ApplicationState) GetDevelopmentTime() string {
	if as.SelectedFilm == nil {
		return "--:--"
	}

	dilutionTimes, ok := as.SelectedFilm.Times20C[as.Dilution]
	if !ok {
		return "--:--"
	}

	time, ok := dilutionTimes[as.SelectedEI]
	if !ok {
		return "--:--"
	}

	return time
}

// CalculateChemicals calculates the chemical dilutions for the current setup
func (as *ApplicationState) CalculateChemicals() {
	if !as.IsComplete() {
		return
	}

	totalVolume := as.RollSetup.TotalVolume
	as.Calculations = []DilutionCalculation{}

	// ILFOSOL 3
	as.Calculations = append(as.Calculations, calculateDilution("ILFOSOL 3", as.Dilution, totalVolume, as.GetDevelopmentTime()))

	// ILFOSTOP
	ilfostop, _ := as.ChemicalDB.GetChemical("ilfostop")
	as.Calculations = append(as.Calculations, calculateDilution("ILFOSTOP", "1+19", totalVolume, ilfostop.Time))

	// SPRINT FIXER
	fixer, _ := as.ChemicalDB.GetChemical("sprint_fixer")
	as.Calculations = append(as.Calculations, calculateDilution("SPRINT FIXER", "1+4", totalVolume, fixer.Time))

	// Create timer state from calculations
	as.TimerState = NewTimerState(as.Calculations)
}

// calculateDilution calculates the dilution for a given chemical
func calculateDilution(chemical, dilution string, totalVolume int, time string) DilutionCalculation {
	var concentrate, water int

	switch dilution {
	case "1+9":
		concentrate = totalVolume / 10
		water = totalVolume - concentrate
	case "1+14":
		concentrate = totalVolume / 15
		water = totalVolume - concentrate
	case "1+19":
		concentrate = totalVolume / 20
		water = totalVolume - concentrate
	case "1+4":
		concentrate = totalVolume / 5
		water = totalVolume - concentrate
	}

	return DilutionCalculation{
		Chemical:    chemical,
		Dilution:    dilution,
		TotalVolume: totalVolume,
		Concentrate: concentrate,
		Water:       water,
		Time:        time,
	}
}

// AppState represents the different states of the application
type AppState int

const (
	MainScreenState AppState = iota
	FilmSelectionState
	EISelectionState
	RollSelectionState
	MixedRollInputState
	CalculatedScreenState
	TimerScreenState
	FixerTrackingState
	SettingsState
)

// String returns the string representation of the app state
func (s AppState) String() string {
	switch s {
	case MainScreenState:
		return "MainScreen"
	case FilmSelectionState:
		return "FilmSelection"
	case EISelectionState:
		return "EISelection"
	case RollSelectionState:
		return "RollSelection"
	case MixedRollInputState:
		return "MixedRollInput"
	case CalculatedScreenState:
		return "CalculatedScreen"
	case TimerScreenState:
		return "TimerScreen"
	case FixerTrackingState:
		return "FixerTracking"
	case SettingsState:
		return "Settings"
	default:
		return "Unknown"
	}
}

// StateMachine manages the application state and transitions
type StateMachine struct {
	currentState AppState
	history      []AppState
	appState     *ApplicationState
}

// NewStateMachine creates a new state machine
func NewStateMachine() *StateMachine {
	return &StateMachine{
		currentState: MainScreenState,
		history:      []AppState{},
		appState:     NewApplicationState(),
	}
}

// GetCurrentState returns the current state
func (sm *StateMachine) GetCurrentState() AppState {
	return sm.currentState
}

// GetApplicationState returns the application state
func (sm *StateMachine) GetApplicationState() *ApplicationState {
	return sm.appState
}

// TransitionTo transitions to a new state
func (sm *StateMachine) TransitionTo(newState AppState) {
	sm.history = append(sm.history, sm.currentState)
	sm.currentState = newState
}

// GoBack goes back to the previous state
func (sm *StateMachine) GoBack() {
	if len(sm.history) > 0 {
		sm.currentState = sm.history[len(sm.history)-1]
		sm.history = sm.history[:len(sm.history)-1]
	}
}

// HandleFilmSelection handles film selection
func (sm *StateMachine) HandleFilmSelection(filmID string) {
	if film, ok := sm.appState.FilmDB.GetFilmByID(filmID); ok {
		sm.appState.SelectedFilm = &film
		sm.TransitionTo(EISelectionState)
	}
}

// HandleEISelection handles EI selection
func (sm *StateMachine) HandleEISelection(ei int) {
	if sm.appState.SelectedFilm != nil {
		// Check if the selected EI is valid for the selected film
		for _, validEI := range sm.appState.SelectedFilm.EIRatings {
			if validEI == ei {
				sm.appState.SelectedEI = ei
				sm.TransitionTo(RollSelectionState)
				return
			}
		}
	}
}

// HandleRollSelection handles roll selection
func (sm *StateMachine) HandleRollSelection(format string, rolls int) {
	tankSize := 0
	rollSetup := &RollSetup{}

	if format == "35mm" {
		rollSetup.Format35mm = rolls
		if size, ok := sm.appState.TankDB.GetTankSize("35mm", rolls); ok {
			tankSize = size
		}
	} else if format == "120mm" {
		rollSetup.Format120mm = rolls
		if size, ok := sm.appState.TankDB.GetTankSize("120mm", rolls); ok {
			tankSize = size
		}
	}

	rollSetup.TotalVolume = tankSize
	sm.appState.RollSetup = rollSetup
	sm.appState.CalculateChemicals()
	sm.TransitionTo(CalculatedScreenState)
}

// HandleMixedRollSetup handles mixed roll setup
func (sm *StateMachine) HandleMixedRollSetup(format35mm, format120mm int) {
	tankSize := CalculateMixedTankSize(format35mm, format120mm, sm.appState.TankDB)
	rollSetup := &RollSetup{
		Format35mm:  format35mm,
		Format120mm: format120mm,
		TotalVolume: tankSize,
	}

	sm.appState.RollSetup = rollSetup
	sm.appState.CalculateChemicals()
	sm.TransitionTo(CalculatedScreenState)
}

// HandleFixerUsage handles fixer usage
func (sm *StateMachine) HandleFixerUsage() {
	if sm.appState.RollSetup != nil {
		totalRolls := sm.appState.RollSetup.TotalRolls()
		sm.appState.FixerState.UseFixer(totalRolls)
	}
}

// Reset resets the application state
func (sm *StateMachine) Reset() {
	sm.currentState = MainScreenState
	sm.history = []AppState{}
	sm.appState = NewApplicationState()
}

// GetValidTransitions returns valid transitions from the current state
func (sm *StateMachine) GetValidTransitions() []AppState {
	switch sm.currentState {
	case MainScreenState:
		return []AppState{FilmSelectionState, FixerTrackingState, SettingsState}
	case FilmSelectionState:
		return []AppState{MainScreenState, EISelectionState}
	case EISelectionState:
		return []AppState{FilmSelectionState, RollSelectionState}
	case RollSelectionState:
		return []AppState{EISelectionState, MixedRollInputState, CalculatedScreenState}
	case MixedRollInputState:
		return []AppState{RollSelectionState, CalculatedScreenState}
	case CalculatedScreenState:
		return []AppState{RollSelectionState, FilmSelectionState, TimerScreenState, MainScreenState}
	case TimerScreenState:
		return []AppState{CalculatedScreenState, MainScreenState}
	case FixerTrackingState:
		return []AppState{MainScreenState}
	case SettingsState:
		return []AppState{MainScreenState}
	default:
		return []AppState{MainScreenState}
	}
}

// CanTransitionTo checks if a transition to the given state is valid
func (sm *StateMachine) CanTransitionTo(state AppState) bool {
	validTransitions := sm.GetValidTransitions()
	for _, validState := range validTransitions {
		if validState == state {
			return true
		}
	}
	return false
}
