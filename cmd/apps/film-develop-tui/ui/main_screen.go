package ui

import (
    "fmt"
    "strings"

    statepkg "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/state"
    types "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/types"
)

// MainScreen represents the main screen
type MainScreen struct{}

func (s *MainScreen) Render(state *types.ApplicationState) string {
    var b strings.Builder

    // Title
    b.WriteString(TitleStyle.Render("🎞️  Film Development Calculator"))
    b.WriteString("\n\n")

    // Main content in single box
    mainContent := s.renderMainContent(state)
    b.WriteString(MainBoxStyle.Render(mainContent))
    b.WriteString("\n")

    // Actions Section (borderless)
    actionsContent := s.renderActions()
    b.WriteString(ActionsOnlyStyle.Render(actionsContent))

    return b.String()
}

func (s *MainScreen) renderMainContent(state *types.ApplicationState) string {
    var b strings.Builder

    // Film Setup Section
    b.WriteString(SectionStyle.Render("Film Setup"))
    b.WriteString("\n")
    filmType := "[ Not Selected ]"
    if state.SelectedFilm != nil {
        filmType = fmt.Sprintf("[ %s ]", state.SelectedFilm.Name)
    }

    ei := "[ -- ]"
    if state.SelectedEI > 0 {
        ei = fmt.Sprintf("[ %d ]", state.SelectedEI)
    }

    rolls := "[ -- ]"
    tank := "[ --ml ]"
    if state.RollSetup != nil {
        rolls = fmt.Sprintf("[ %s ]", state.RollSetup.String())
        tank = fmt.Sprintf("[ %dml ]", state.RollSetup.TotalVolume)
    }

    b.WriteString(fmt.Sprintf("Film Type:    %-30s EI:  %s\n", filmType, ei))
    b.WriteString(fmt.Sprintf("Rolls:        %-30s Tank: %s", rolls, tank))
    b.WriteString("\n\n")

    // Chemicals Section
    b.WriteString(SectionStyle.Render("Chemicals (20°C)"))
    b.WriteString("\n")
    b.WriteString(s.renderChemicalModels(state))
    b.WriteString("\n\n")

    // Fixer Usage Section
    b.WriteString(SectionStyle.Render("Fixer Usage"))
    b.WriteString("\n")
    capacity := state.FixerState.CapacityPerLiter
    used := state.FixerState.UsedRolls
    remaining := state.FixerState.RemainingCapacity()
    b.WriteString(fmt.Sprintf("Capacity: %d rolls per liter    Used: %d rolls    Remaining: %d rolls",
        capacity, used, remaining))

    return b.String()
}

func (s *MainScreen) renderChemicalModels(state *types.ApplicationState) string {
    chemicals := types.GetCalculatedChemicals(state.Calculations)
    components := types.ChemicalModelsToComponents(chemicals)
    
    return s.renderChemicalComponents(components, false)
}

// renderChemicalComponents renders chemical components with proper separation
func (s *MainScreen) renderChemicalComponents(components []types.ChemicalComponent, highlight bool) string {
    if len(components) == 0 {
        return ""
    }
    
    // Get component lines
    var componentLines [][]string
    for _, component := range components {
        var rendered string
        if highlight {
            rendered = component.RenderWithHighlight(HighlightStyle)
        } else {
            rendered = component.Render()
        }
        componentLines = append(componentLines, strings.Split(rendered, "\n"))
    }
    
    // Build output by joining lines horizontally
    var result strings.Builder
    maxLines := 5 // Name, Dilution, Concentrate, Water, Time
    
    for line := 0; line < maxLines; line++ {
        for i, componentLine := range componentLines {
            if i > 0 {
                result.WriteString(" │  ")
            }
            if line < len(componentLine) {
                result.WriteString(componentLine[line])
            }
        }
        if line < maxLines-1 {
            result.WriteString("\n")
        }
    }
    
    return result.String()
}

func (s *MainScreen) renderActions() string {
    return "[F] Film Type    [U] Fixer Usage    [S] Settings    [Q] Quit"
}

func (s *MainScreen) HandleInput(key string, sm *statepkg.StateMachine) bool {
    switch strings.ToLower(key) {
    case "f":
        sm.TransitionTo(statepkg.FilmSelectionState)
        return true
    case "u":
        sm.TransitionTo(statepkg.FixerTrackingState)
        return true
    case "s":
        sm.TransitionTo(statepkg.SettingsState)
        return true
    case "q":
        return false
    }
    return true
} 