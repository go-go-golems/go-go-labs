package ui

import (
    "fmt"
    "strings"

    statepkg "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/state"
    types "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/types"
)

// MixedRollInputScreen represents the mixed roll input screen
type MixedRollInputScreen struct {
    rolls35mm  int
    rolls120mm int
}

func (s *MixedRollInputScreen) Render(state *types.ApplicationState) string {
    var b strings.Builder

    // Title
    b.WriteString(TitleStyle.Render("🎞️  Film Development Calculator"))
    b.WriteString("\n\n")

    // Mixed Roll Setup
    mixedRollContent := s.renderMixedRollSetup(state)
    b.WriteString(BoxStyle.Render(mixedRollContent))
    b.WriteString("\n\n")

    // Actions
    actionsContent := s.renderActions()
    b.WriteString(BoxStyle.Render(actionsContent))

    return b.String()
}

func (s *MixedRollInputScreen) renderMixedRollSetup(state *types.ApplicationState) string {
    var b strings.Builder
    b.WriteString(SectionStyle.Render("Custom Mix Setup"))
    b.WriteString("\n\n")

    totalVolume := types.CalculateMixedTankSize(s.rolls35mm, s.rolls120mm, state.TankDB)

    b.WriteString(fmt.Sprintf("35mm Rolls: %s    (↑/↓ or +/- to adjust)\n", HighlightStyle.Render(fmt.Sprintf("[ %d ]", s.rolls35mm))))
    b.WriteString(fmt.Sprintf("120mm Rolls: %s   (↑/↓ or +/- to adjust)\n", HighlightStyle.Render(fmt.Sprintf("[ %d ]", s.rolls120mm))))
    b.WriteString("\n")
    b.WriteString(fmt.Sprintf("Total Tank Size: %s\n", HighlightStyle.Render(fmt.Sprintf("[ %dml ]", totalVolume))))
    b.WriteString("\n")
    b.WriteString(fmt.Sprintf("%s Confirm    %s Back    %s Reset",
        ActionStyle.Render("[ENTER]"), ActionStyle.Render("[ESC]"), ActionStyle.Render("[R]")))

    return b.String()
}

func (s *MixedRollInputScreen) renderActions() string {
    var b strings.Builder
    b.WriteString(SectionStyle.Render("Actions"))
    b.WriteString("\n\n")

    b.WriteString(ActionStyle.Render("[↑↓] Adjust 35mm    [+/-] Adjust 120mm    [ENTER] Confirm    [ESC] Back"))

    return b.String()
}

func (s *MixedRollInputScreen) HandleInput(key string, sm *statepkg.StateMachine) bool {
    switch key {
    case "up":
        if s.rolls35mm < 6 {
            s.rolls35mm++
        }
        return true
    case "down":
        if s.rolls35mm > 0 {
            s.rolls35mm--
        }
        return true
    case "+":
        if s.rolls120mm < 6 {
            s.rolls120mm++
        }
        return true
    case "-":
        if s.rolls120mm > 0 {
            s.rolls120mm--
        }
        return true
    case "enter":
        sm.HandleMixedRollSetup(s.rolls35mm, s.rolls120mm)
        return true
    case "r":
        s.rolls35mm = 0
        s.rolls120mm = 0
        return true
    case "esc":
        sm.GoBack()
        return true
    case "q":
        return false
    }
    return true
} 