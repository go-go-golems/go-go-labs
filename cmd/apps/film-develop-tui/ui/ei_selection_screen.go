package ui

import (
    "fmt"
    "strconv"
    "strings"

    statepkg "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/state"
    types "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/types"
)

// EISelectionScreen represents the EI selection screen
type EISelectionScreen struct{}

func (s *EISelectionScreen) Render(state *types.ApplicationState) string {
    var b strings.Builder

    // Title
    b.WriteString(TitleStyle.Render("🎞️  Film Development Calculator"))
    b.WriteString("\n\n")

    // Film Setup
    filmSetupContent := s.renderFilmSetup(state)
    b.WriteString(BoxStyle.Render(filmSetupContent))
    b.WriteString("\n\n")

    // EI Selection
    eiSelectionContent := s.renderEISelection(state)
    b.WriteString(BoxStyle.Render(eiSelectionContent))
    b.WriteString("\n\n")

    // Actions
    actionsContent := s.renderActions()
    b.WriteString(BoxStyle.Render(actionsContent))

    return b.String()
}

func (s *EISelectionScreen) renderFilmSetup(state *types.ApplicationState) string {
    var b strings.Builder
    b.WriteString(SectionStyle.Render("Film Setup"))
    b.WriteString("\n\n")

    filmType := "[ Not Selected ]"
    if state.SelectedFilm != nil {
        filmType = fmt.Sprintf("[ %s ]", state.SelectedFilm.Name)
    }

    ei := "[ Not Set ]"
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

    return b.String()
}

func (s *EISelectionScreen) renderEISelection(state *types.ApplicationState) string {
    var b strings.Builder
    b.WriteString(SectionStyle.Render("Select EI Rating"))
    b.WriteString("\n\n")

    if state.SelectedFilm == nil {
        b.WriteString(ErrorStyle.Render("No film selected"))
        return b.String()
    }

    dilution := "1+9"
    for i, ei := range state.SelectedFilm.EIRatings {
        time := "--:--"
        if dilutionTimes, ok := state.SelectedFilm.Times20C[dilution]; ok {
            if t, ok := dilutionTimes[ei]; ok {
                time = t
            }
        }

        description := ""
        switch {
        case ei <= 125:
            description = "🌞 Bright light, fine grain"
        case ei <= 400:
            description = "📷 Standard, most common"
        case ei <= 800:
            description = "🌆 Low light, pushed grain"
        default:
            description = "🌙 Very low light, high grain"
        }

        b.WriteString(fmt.Sprintf(ActionStyle.Render("[%d]")+" EI %-4d (%s @ %s)     %s\n",
            i+1, ei, time, dilution, description))
    }

    b.WriteString("\n")
    b.WriteString(DimStyle.Render("[ESC] Back to film selection"))

    return b.String()
}

func (s *EISelectionScreen) renderActions() string {
    var b strings.Builder
    b.WriteString(SectionStyle.Render("Actions"))
    b.WriteString("\n\n")

    b.WriteString(ActionStyle.Render("[1-9] Select EI    [ESC] Back    [Q] Quit"))

    return b.String()
}

func (s *EISelectionScreen) HandleInput(key string, sm *statepkg.StateMachine) bool {
    switch strings.ToLower(key) {
    case "1", "2", "3", "4", "5", "6", "7", "8", "9":
        index, _ := strconv.Atoi(key)
        appState := sm.GetApplicationState()
        if appState.SelectedFilm != nil && index > 0 && index <= len(appState.SelectedFilm.EIRatings) {
            ei := appState.SelectedFilm.EIRatings[index-1]
            sm.HandleEISelection(ei)
        }
        return true
    case "esc":
        sm.GoBack()
        return true
    case "q":
        return false
    }
    return true
} 