package pkg

import types "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/types"

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
)

// Wrapper constructors
func NewApplicationState() *ApplicationState { return types.NewApplicationState() }

// Function re-exports for convenience
var (
    GetDefaultChemicals     = types.GetDefaultChemicals
    GetCalculatedChemicals  = types.GetCalculatedChemicals
    ParseDuration           = types.ParseDuration
    FormatDuration          = types.FormatDuration
    CalculateMixedTankSize  = types.CalculateMixedTankSize
    GetFilmOrder            = types.GetFilmOrder
    ChemicalModelsToComponents = types.ChemicalModelsToComponents
) 