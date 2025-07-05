package ui

import (
    "fmt"
    "strings"

    statepkg "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/state"
    types "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/types"
)

// TimerScreen represents the timer screen
type TimerScreen struct{}

func (s *TimerScreen) Render(state *types.ApplicationState) string {
    var b strings.Builder

    // Title
    b.WriteString(TitleStyle.Render("🎞️  Film Development Timer"))
    b.WriteString("\n\n")

    // Main content in single box
    mainContent := s.renderMainContent(state)
    b.WriteString(MainBoxStyle.Render(mainContent))
    b.WriteString("\n")

    // Actions Section (borderless)
    actionsContent := s.renderActions(state)
    b.WriteString(ActionsOnlyStyle.Render(actionsContent))

    return b.String()
}

func (s *TimerScreen) renderMainContent(state *types.ApplicationState) string {
    var b strings.Builder

    // Timer Display Section
    b.WriteString(SectionStyle.Render("Current Step"))
    b.WriteString("\n")
    b.WriteString(s.renderTimerContent(state))
    b.WriteString("\n\n")

    // Steps Progress Section
    b.WriteString(SectionStyle.Render("Development Steps"))
    b.WriteString("\n")
    b.WriteString(s.renderStepsContent(state))

    return b.String()
}

func (s *TimerScreen) renderTimerContent(state *types.ApplicationState) string {
    var b strings.Builder

    if state.TimerState == nil || len(state.TimerState.Steps) == 0 {
        b.WriteString(ErrorStyle.Render("No timer available"))
        return b.String()
    }

    currentStep := state.TimerState.GetCurrentStep()
    if currentStep == nil {
        b.WriteString(HighlightStyle.Render("🎉 All steps completed!"))
        return b.String()
    }

    elapsed := state.TimerState.GetCurrentElapsed()
    remaining := state.TimerState.GetRemainingTime()
    isOvertime := state.TimerState.IsCurrentStepOvertime()

    // Step name and target time
    b.WriteString(fmt.Sprintf("Step: %s\n", HighlightStyle.Render(currentStep.Name)))
    b.WriteString(fmt.Sprintf("Target Time: %s\n", types.FormatDuration(currentStep.Duration)))

    // Timer display
    elapsedStr := types.FormatDuration(elapsed)
    remainingStr := types.FormatDuration(remaining)

    if isOvertime {
        b.WriteString(fmt.Sprintf("Elapsed: %s ⚠️  OVERTIME\n", ErrorStyle.Render(elapsedStr)))
        b.WriteString(fmt.Sprintf("Overtime: %s\n", ErrorStyle.Render(types.FormatDuration(elapsed-currentStep.Duration))))
    } else {
        b.WriteString(fmt.Sprintf("Elapsed: %s\n", HighlightStyle.Render(elapsedStr)))
        b.WriteString(fmt.Sprintf("Remaining: %s\n", remainingStr))
    }

    // Status
    status := ""
    if state.TimerState.IsRunning {
        if state.TimerState.IsPaused {
            status = DimStyle.Render("⏸️  PAUSED")
        } else {
            status = HighlightStyle.Render("⏱️  RUNNING")
        }
    } else {
        status = DimStyle.Render("⏹️  STOPPED")
    }
    b.WriteString(fmt.Sprintf("Status: %s", status))

    return b.String()
}

func (s *TimerScreen) renderStepsContent(state *types.ApplicationState) string {
    var b strings.Builder

    if state.TimerState == nil || len(state.TimerState.Steps) == 0 {
        b.WriteString(DimStyle.Render("No steps available"))
        return b.String()
    }

    for i, step := range state.TimerState.Steps {
        icon := "○"
        style := DimStyle

        if step.Finished {
            icon = "✅"
            style = DimStyle
        } else if i == state.TimerState.CurrentStep {
            icon = "🔵"
            style = HighlightStyle
        }

        stepText := fmt.Sprintf("%s %s (%s)", icon, step.Name, types.FormatDuration(step.Duration))
        b.WriteString(style.Render(stepText))
        b.WriteString("\n")
    }

    return b.String()
}

func (s *TimerScreen) renderActions(state *types.ApplicationState) string {
    if state.TimerState == nil || state.TimerState.IsComplete {
        return "[R] Reset    [ESC] Back    [Q] Quit"
    }

    if state.TimerState.IsRunning {
        if state.TimerState.IsPaused {
            return "[Space] Resume    [N] Next Step    [S] Stop    [R] Reset    [ESC] Back"
        } else {
            return "[Space] Pause    [N] Next Step    [S] Stop    [R] Reset    [ESC] Back"
        }
    } else {
        return "[Space] Start    [N] Next Step    [R] Reset    [ESC] Back    [Q] Quit"
    }
}

func (s *TimerScreen) HandleInput(key string, sm *statepkg.StateMachine) bool {
    appState := sm.GetApplicationState()
    if appState.TimerState == nil {
        return true
    }

    switch key {
    case "space":
        if appState.TimerState.IsRunning {
            if appState.TimerState.IsPaused {
                appState.TimerState.ResumeTimer()
            } else {
                appState.TimerState.PauseTimer()
            }
        } else {
            appState.TimerState.StartTimer()
        }
        return true
    case "s":
        appState.TimerState.StopTimer()
        return true
    case "n":
        appState.TimerState.CompleteCurrentStep()
        return true
    case "r":
        appState.TimerState.Reset()
        return true
    case "esc":
        sm.GoBack()
        return true
    case "q":
        return false
    }
    return true
} 