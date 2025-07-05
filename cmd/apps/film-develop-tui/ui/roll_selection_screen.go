package ui

import (
    "fmt"
    "strconv"
    "strings"

    statepkg "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/state"
    types "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/types"
)

// RollSelectionScreen represents the roll selection screen
type RollSelectionScreen struct{}

func (s *RollSelectionScreen) Render(state *types.ApplicationState) string {
    var b strings.Builder

    // Title
    b.WriteString(TitleStyle.Render("🎞️  Film Development Calculator"))
    b.WriteString("\n\n")

    // Film Setup
    filmSetupContent := s.renderFilmSetup(state)
    b.WriteString(BoxStyle.Render(filmSetupContent))
    b.WriteString("\n\n")

    // Roll Selection
    rollSelectionContent := s.renderRollSelection(state)
    b.WriteString(BoxStyle.Render(rollSelectionContent))
    b.WriteString("\n\n")

    // Actions
    actionsContent := s.renderActions()
    b.WriteString(BoxStyle.Render(actionsContent))

    return b.String()
}

func (s *RollSelectionScreen) renderFilmSetup(state *types.ApplicationState) string {
    var b strings.Builder
    b.WriteString(SectionStyle.Render("Film Setup"))
    b.WriteString("\n\n")

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

    return b.String()
}

func (s *RollSelectionScreen) renderRollSelection(state *types.ApplicationState) string {
    var b strings.Builder
    b.WriteString(SectionStyle.Render("Number of Rolls"))
    b.WriteString("\n\n")

    b.WriteString("35mm Rolls:                           120mm Rolls:\n")
    b.WriteString(fmt.Sprintf("%s 1 Roll (300ml)  %s 4 Rolls       %s 1 Roll (500ml)  %s 4 Rolls\n",
        ActionStyle.Render("[1]"), ActionStyle.Render("[4]"), ActionStyle.Render("[A]"), ActionStyle.Render("[D]")))
    b.WriteString(fmt.Sprintf("%s 2 Rolls (500ml) %s 5 Rolls       %s 2 Rolls (700ml) %s 5 Rolls\n",
        ActionStyle.Render("[2]"), ActionStyle.Render("[5]"), ActionStyle.Render("[B]"), ActionStyle.Render("[E]")))
    b.WriteString(fmt.Sprintf("%s 3 Rolls (600ml) %s 6 Rolls       %s 3 Rolls (900ml) %s 6 Rolls\n",
        ActionStyle.Render("[3]"), ActionStyle.Render("[6]"), ActionStyle.Render("[C]"), ActionStyle.Render("[F]")))
    b.WriteString("\n")
    b.WriteString(fmt.Sprintf("Mixed batches: %s Custom mix\n", ActionStyle.Render("[M]")))
    b.WriteString("\n")
    b.WriteString(DimStyle.Render("[ESC] Back to EI selection"))

    return b.String()
}

func (s *RollSelectionScreen) renderActions() string {
    var b strings.Builder
    b.WriteString(SectionStyle.Render("Actions"))
    b.WriteString("\n\n")

    b.WriteString(ActionStyle.Render("[1-6] 35mm    [A-F] 120mm    [M] Mixed    [ESC] Back    [Q] Quit"))

    return b.String()
}

func (s *RollSelectionScreen) HandleInput(key string, sm *statepkg.StateMachine) bool {
    switch strings.ToLower(key) {
    case "1", "2", "3", "4", "5", "6":
        rolls, _ := strconv.Atoi(key)
        sm.HandleRollSelection("35mm", rolls)
        return true
    case "a", "b", "c", "d", "e", "f":
        rollMap := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6}
        rolls := rollMap[key]
        sm.HandleRollSelection("120mm", rolls)
        return true
    case "m":
        sm.TransitionTo(statepkg.MixedRollInputState)
        return true
    case "esc":
        sm.GoBack()
        return true
    case "q":
        return false
    }
    return true
} 