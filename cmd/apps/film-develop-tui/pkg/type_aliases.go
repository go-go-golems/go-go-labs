package pkg

import (
	types "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/types"
	statepkg "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/state"
	uipkg "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/ui"
)

// Type aliases to bridge existing code with new types package
type (
	Film                = types.Film
	FilmDatabase        = types.FilmDatabase
	TankDatabase        = types.TankDatabase
	TankSize            = types.TankSize
	Chemical            = types.Chemical
	ChemicalDatabase    = types.ChemicalDatabase
	DilutionCalculation = types.DilutionCalculation
	ChemicalModel       = types.ChemicalModel
	ChemicalComponent   = types.ChemicalComponent
	RollSetup           = types.RollSetup
	FixerState          = types.FixerState
	ApplicationState    = types.ApplicationState
	TimerState          = types.TimerState
	TimerStep           = types.TimerStep
	AppState            = statepkg.AppState
	StateMachine        = statepkg.StateMachine
	Screen              = uipkg.Screen
)

// Wrapper constructors
func NewApplicationState() *ApplicationState { return types.NewApplicationState() }
func NewStateMachine() *StateMachine { return statepkg.NewStateMachine() }

// Function re-exports for convenience
var (
	GetDefaultChemicals     = types.GetDefaultChemicals
	GetCalculatedChemicals  = types.GetCalculatedChemicals
	ParseDuration           = types.ParseDuration
	FormatDuration          = types.FormatDuration
	CalculateMixedTankSize  = types.CalculateMixedTankSize
	GetFilmOrder            = types.GetFilmOrder
	ChemicalModelsToComponents = types.ChemicalModelsToComponents
	GetScreenForState        = uipkg.GetScreenForState
)

const (
	MainScreenState      = statepkg.MainScreenState
	FilmSelectionState   = statepkg.FilmSelectionState
	EISelectionState     = statepkg.EISelectionState
	RollSelectionState   = statepkg.RollSelectionState
	MixedRollInputState  = statepkg.MixedRollInputState
	CalculatedScreenState = statepkg.CalculatedScreenState
	TimerScreenState     = statepkg.TimerScreenState
	FixerTrackingState   = statepkg.FixerTrackingState
	SettingsState        = statepkg.SettingsState
) 