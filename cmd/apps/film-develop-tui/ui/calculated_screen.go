package ui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"

    statepkg "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/state"
    types "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/types"
)

// CalculatedScreen represents the calculated results screen
type CalculatedScreen struct{}

func (s *CalculatedScreen) Render(state *types.ApplicationState) string {
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

func (s *CalculatedScreen) renderMainContent(state *types.ApplicationState) string {
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

    batchRolls := 0
    if state.RollSetup != nil {
        batchRolls = state.RollSetup.TotalRolls()
    }

    b.WriteString(fmt.Sprintf("Capacity: %d rolls per liter    Used: %d rolls    Remaining: %d rolls\n",
        capacity, used, remaining))
    b.WriteString(fmt.Sprintf("This batch uses: %s         After use: %s remaining",
        HighlightStyle.Render(fmt.Sprintf("%d roll", batchRolls)),
        HighlightStyle.Render(fmt.Sprintf("%d rolls", remaining-batchRolls))))

    return b.String()
}

func (s *CalculatedScreen) renderChemicalModels(state *types.ApplicationState) string {
    chemicals := types.GetCalculatedChemicals(state.Calculations)
    components := types.ChemicalModelsToComponents(chemicals)
    
    return s.renderChemicalComponents(components, true)
}

// renderChemicalComponents renders chemical components with proper separation
func (s *CalculatedScreen) renderChemicalComponents(components []types.ChemicalComponent, highlight bool) string {
    var renderedComponents []string
    
    for _, component := range components {
        if highlight {
            renderedComponents = append(renderedComponents, component.RenderWithHighlight(HighlightStyle))
        } else {
            renderedComponents = append(renderedComponents, component.Render())
        }
    }
    
    // Join components horizontally
    return lipgloss.JoinHorizontal(lipgloss.Top, renderedComponents...)
}

func (s *CalculatedScreen) renderActions() string {
    return "[T] Timer    [U] Use Fixer    [R] Change Rolls    [F] Change Film    [Q] Quit"
}

func (s *CalculatedScreen) HandleInput(key string, sm *statepkg.StateMachine) bool {
    switch strings.ToLower(key) {
    case "u":
        sm.HandleFixerUsage()
        return true
    case "t":
        sm.TransitionTo(statepkg.TimerScreenState)
        return true
    case "r":
        sm.TransitionTo(statepkg.RollSelectionState)
        return true
    case "f":
        sm.TransitionTo(statepkg.FilmSelectionState)
        return true
    case "q":
        return false
    }
    return true
} 