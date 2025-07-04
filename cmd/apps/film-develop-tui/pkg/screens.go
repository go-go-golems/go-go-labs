package pkg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/internal/models"
)

// Screen represents a screen in the application
type Screen interface {
	Render(state *models.ApplicationState) string
	HandleInput(key string, sm *models.StateMachine) bool
}

// Styles for the UI
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Align(lipgloss.Center).
			Width(81)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1).
			Width(79)

	mainBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1).
			Width(79)

	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("11"))

	highlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("10"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("9"))

	actionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14"))

	actionsOnlyStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("14")).
				Align(lipgloss.Center).
				Width(79)
)

// MainScreen represents the main screen
type MainScreen struct{}

func (s *MainScreen) Render(state *models.ApplicationState) string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🎞️  Film Development Calculator"))
	b.WriteString("\n\n")

	// Main content in single box
	mainContent := s.renderMainContent(state)
	b.WriteString(mainBoxStyle.Render(mainContent))
	b.WriteString("\n")

	// Actions Section (borderless)
	actionsContent := s.renderActions()
	b.WriteString(actionsOnlyStyle.Render(actionsContent))

	return b.String()
}

func (s *MainScreen) renderMainContent(state *models.ApplicationState) string {
	var b strings.Builder

	// Film Setup Section
	b.WriteString(sectionStyle.Render("Film Setup"))
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
	b.WriteString(sectionStyle.Render("Chemicals (20°C)"))
	b.WriteString("\n")
	b.WriteString(s.renderChemicalModels(state))
	b.WriteString("\n\n")

	// Fixer Usage Section
	b.WriteString(sectionStyle.Render("Fixer Usage"))
	b.WriteString("\n")
	capacity := state.FixerState.CapacityPerLiter
	used := state.FixerState.UsedRolls
	remaining := state.FixerState.RemainingCapacity()
	b.WriteString(fmt.Sprintf("Capacity: %d rolls per liter    Used: %d rolls    Remaining: %d rolls",
		capacity, used, remaining))

	return b.String()
}

func (s *MainScreen) renderChemicalModels(state *models.ApplicationState) string {
	// Convert models.DilutionCalculation to pkg.DilutionCalculation
	var calculations []models.DilutionCalculation
	for _, calc := range state.Calculations {
		calculations = append(calculations, models.DilutionCalculation{
			Chemical:    calc.Chemical,
			Dilution:    calc.Dilution,
			TotalVolume: calc.TotalVolume,
			Concentrate: calc.Concentrate,
			Water:       calc.Water,
			Time:        calc.Time,
		})
	}
	
	chemicals := models.GetCalculatedChemicals(calculations)
	components := models.ChemicalModelsToComponents(chemicals)
	
	return s.renderChemicalComponents(components, false)
}

// renderChemicalComponents renders chemical components with proper separation
func (s *MainScreen) renderChemicalComponents(components []models.ChemicalComponent, highlight bool) string {
	if len(components) == 0 {
		return ""
	}
	
	// Get component lines
	var componentLines [][]string
	for _, component := range components {
		var rendered string
		if highlight {
			rendered = component.RenderWithHighlight(highlightStyle)
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

func (s *MainScreen) HandleInput(key string, sm *models.StateMachine) bool {
	switch strings.ToLower(key) {
	case "f":
		sm.TransitionTo(models.FilmSelectionState)
		return true
	case "u":
		sm.TransitionTo(models.FixerTrackingState)
		return true
	case "s":
		sm.TransitionTo(models.SettingsState)
		return true
	case "q":
		return false
	}
	return true
}

// FilmSelectionScreen represents the film selection screen
type FilmSelectionScreen struct{}

func (s *FilmSelectionScreen) Render(state *models.ApplicationState) string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🎞️  Film Development Calculator"))
	b.WriteString("\n\n")

	// Film Selection
	filmSelectionContent := s.renderFilmSelection(state)
	b.WriteString(boxStyle.Render(filmSelectionContent))
	b.WriteString("\n\n")

	// Actions
	actionsContent := s.renderActions()
	b.WriteString(boxStyle.Render(actionsContent))

	return b.String()
}

func (s *FilmSelectionScreen) renderFilmSelection(state *models.ApplicationState) string {
	var b strings.Builder
	b.WriteString(sectionStyle.Render("Select Film Type"))
	b.WriteString("\n\n")

	filmOrder := models.GetFilmOrder()
	for i, filmID := range filmOrder {
		if film, ok := state.FilmDB.GetFilmByID(filmID); ok {
			eiRatings := make([]string, len(film.EIRatings))
			for j, ei := range film.EIRatings {
				eiRatings[j] = strconv.Itoa(ei)
			}
			eiStr := strings.Join(eiRatings, "/")
			if len(film.EIRatings) > 3 {
				eiStr = fmt.Sprintf("%d-%d", film.EIRatings[0], film.EIRatings[len(film.EIRatings)-1])
			}

			b.WriteString(fmt.Sprintf(actionStyle.Render("[%d]")+" %-12s (EI %-10s) %s %s\n",
				i+1, film.Name, eiStr, film.Icon, film.Description))
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("[ESC] Back"))

	return b.String()
}

func (s *FilmSelectionScreen) renderActions() string {
	var b strings.Builder
	b.WriteString(sectionStyle.Render("Actions"))
	b.WriteString("\n\n")

	b.WriteString(actionStyle.Render("[1-7] Select Film    [ESC] Back    [Q] Quit"))

	return b.String()
}

func (s *FilmSelectionScreen) HandleInput(key string, sm *models.StateMachine) bool {
	switch strings.ToLower(key) {
	case "1", "2", "3", "4", "5", "6", "7":
		index, _ := strconv.Atoi(key)
		filmOrder := models.GetFilmOrder()
		if index > 0 && index <= len(filmOrder) {
			sm.HandleFilmSelection(filmOrder[index-1])
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

// EISelectionScreen represents the EI selection screen
type EISelectionScreen struct{}

func (s *EISelectionScreen) Render(state *models.ApplicationState) string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🎞️  Film Development Calculator"))
	b.WriteString("\n\n")

	// Film Setup
	filmSetupContent := s.renderFilmSetup(state)
	b.WriteString(boxStyle.Render(filmSetupContent))
	b.WriteString("\n\n")

	// EI Selection
	eiSelectionContent := s.renderEISelection(state)
	b.WriteString(boxStyle.Render(eiSelectionContent))
	b.WriteString("\n\n")

	// Actions
	actionsContent := s.renderActions()
	b.WriteString(boxStyle.Render(actionsContent))

	return b.String()
}

func (s *EISelectionScreen) renderFilmSetup(state *models.ApplicationState) string {
	var b strings.Builder
	b.WriteString(sectionStyle.Render("Film Setup"))
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

func (s *EISelectionScreen) renderEISelection(state *models.ApplicationState) string {
	var b strings.Builder
	b.WriteString(sectionStyle.Render("Select EI Rating"))
	b.WriteString("\n\n")

	if state.SelectedFilm == nil {
		b.WriteString(errorStyle.Render("No film selected"))
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

		b.WriteString(fmt.Sprintf(actionStyle.Render("[%d]")+" EI %-4d (%s @ %s)     %s\n",
			i+1, ei, time, dilution, description))
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("[ESC] Back to film selection"))

	return b.String()
}

func (s *EISelectionScreen) renderActions() string {
	var b strings.Builder
	b.WriteString(sectionStyle.Render("Actions"))
	b.WriteString("\n\n")

	b.WriteString(actionStyle.Render("[1-9] Select EI    [ESC] Back    [Q] Quit"))

	return b.String()
}

func (s *EISelectionScreen) HandleInput(key string, sm *models.StateMachine) bool {
	switch strings.ToLower(key) {
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		index, _ := strconv.Atoi(key)
		if sm.GetApplicationState().SelectedFilm != nil && index > 0 && index <= len(sm.GetApplicationState().SelectedFilm.EIRatings) {
			ei := sm.GetApplicationState().SelectedFilm.EIRatings[index-1]
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

// RollSelectionScreen represents the roll selection screen
type RollSelectionScreen struct{}

func (s *RollSelectionScreen) Render(state *models.ApplicationState) string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🎞️  Film Development Calculator"))
	b.WriteString("\n\n")

	// Film Setup
	filmSetupContent := s.renderFilmSetup(state)
	b.WriteString(boxStyle.Render(filmSetupContent))
	b.WriteString("\n\n")

	// Roll Selection
	rollSelectionContent := s.renderRollSelection(state)
	b.WriteString(boxStyle.Render(rollSelectionContent))
	b.WriteString("\n\n")

	// Actions
	actionsContent := s.renderActions()
	b.WriteString(boxStyle.Render(actionsContent))

	return b.String()
}

func (s *RollSelectionScreen) renderFilmSetup(state *models.ApplicationState) string {
	var b strings.Builder
	b.WriteString(sectionStyle.Render("Film Setup"))
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

func (s *RollSelectionScreen) renderRollSelection(state *models.ApplicationState) string {
	var b strings.Builder
	b.WriteString(sectionStyle.Render("Number of Rolls"))
	b.WriteString("\n\n")

	b.WriteString("35mm Rolls:                           120mm Rolls:\n")
	b.WriteString(fmt.Sprintf("%s 1 Roll (300ml)  %s 4 Rolls       %s 1 Roll (500ml)  %s 4 Rolls\n",
		actionStyle.Render("[1]"), actionStyle.Render("[4]"), actionStyle.Render("[A]"), actionStyle.Render("[D]")))
	b.WriteString(fmt.Sprintf("%s 2 Rolls (500ml) %s 5 Rolls       %s 2 Rolls (700ml) %s 5 Rolls\n",
		actionStyle.Render("[2]"), actionStyle.Render("[5]"), actionStyle.Render("[B]"), actionStyle.Render("[E]")))
	b.WriteString(fmt.Sprintf("%s 3 Rolls (600ml) %s 6 Rolls       %s 3 Rolls (900ml) %s 6 Rolls\n",
		actionStyle.Render("[3]"), actionStyle.Render("[6]"), actionStyle.Render("[C]"), actionStyle.Render("[F]")))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Mixed batches: %s Custom mix\n", actionStyle.Render("[M]")))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("[ESC] Back to EI selection"))

	return b.String()
}

func (s *RollSelectionScreen) renderActions() string {
	var b strings.Builder
	b.WriteString(sectionStyle.Render("Actions"))
	b.WriteString("\n\n")

	b.WriteString(actionStyle.Render("[1-6] 35mm    [A-F] 120mm    [M] Mixed    [ESC] Back    [Q] Quit"))

	return b.String()
}

func (s *RollSelectionScreen) HandleInput(key string, sm *models.StateMachine) bool {
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
		sm.TransitionTo(models.MixedRollInputState)
		return true
	case "esc":
		sm.GoBack()
		return true
	case "q":
		return false
	}
	return true
}

// MixedRollInputScreen represents the mixed roll input screen
type MixedRollInputScreen struct {
	rolls35mm  int
	rolls120mm int
}

func (s *MixedRollInputScreen) Render(state *models.ApplicationState) string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🎞️  Film Development Calculator"))
	b.WriteString("\n\n")

	// Mixed Roll Setup
	mixedRollContent := s.renderMixedRollSetup(state)
	b.WriteString(boxStyle.Render(mixedRollContent))
	b.WriteString("\n\n")

	// Actions
	actionsContent := s.renderActions()
	b.WriteString(boxStyle.Render(actionsContent))

	return b.String()
}

func (s *MixedRollInputScreen) renderMixedRollSetup(state *models.ApplicationState) string {
	var b strings.Builder
	b.WriteString(sectionStyle.Render("Custom Mix Setup"))
	b.WriteString("\n\n")

	// Convert models.TankDatabase to pkg.TankDatabase
	tankDB := &models.TankDatabase{
		Sizes: state.TankDB.Sizes,
	}
	totalVolume := models.CalculateMixedTankSize(s.rolls35mm, s.rolls120mm, tankDB)

	b.WriteString(fmt.Sprintf("35mm Rolls: %s    (↑/↓ or +/- to adjust)\n", highlightStyle.Render(fmt.Sprintf("[ %d ]", s.rolls35mm))))
	b.WriteString(fmt.Sprintf("120mm Rolls: %s   (↑/↓ or +/- to adjust)\n", highlightStyle.Render(fmt.Sprintf("[ %d ]", s.rolls120mm))))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Total Tank Size: %s\n", highlightStyle.Render(fmt.Sprintf("[ %dml ]", totalVolume))))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("%s Confirm    %s Back    %s Reset",
		actionStyle.Render("[ENTER]"), actionStyle.Render("[ESC]"), actionStyle.Render("[R]")))

	return b.String()
}

func (s *MixedRollInputScreen) renderActions() string {
	var b strings.Builder
	b.WriteString(sectionStyle.Render("Actions"))
	b.WriteString("\n\n")

	b.WriteString(actionStyle.Render("[↑↓] Adjust 35mm    [+/-] Adjust 120mm    [ENTER] Confirm    [ESC] Back"))

	return b.String()
}

func (s *MixedRollInputScreen) HandleInput(key string, sm *models.StateMachine) bool {
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

// CalculatedScreen represents the calculated results screen
type CalculatedScreen struct{}

func (s *CalculatedScreen) Render(state *models.ApplicationState) string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🎞️  Film Development Calculator"))
	b.WriteString("\n\n")

	// Main content in single box
	mainContent := s.renderMainContent(state)
	b.WriteString(mainBoxStyle.Render(mainContent))
	b.WriteString("\n")

	// Actions Section (borderless)
	actionsContent := s.renderActions()
	b.WriteString(actionsOnlyStyle.Render(actionsContent))

	return b.String()
}

func (s *CalculatedScreen) renderMainContent(state *models.ApplicationState) string {
	var b strings.Builder

	// Film Setup Section
	b.WriteString(sectionStyle.Render("Film Setup"))
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
	b.WriteString(sectionStyle.Render("Chemicals (20°C)"))
	b.WriteString("\n")
	b.WriteString(s.renderChemicalModels(state))
	b.WriteString("\n\n")

	// Fixer Usage Section
	b.WriteString(sectionStyle.Render("Fixer Usage"))
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
		highlightStyle.Render(fmt.Sprintf("%d roll", batchRolls)),
		highlightStyle.Render(fmt.Sprintf("%d rolls", remaining-batchRolls))))

	return b.String()
}

func (s *CalculatedScreen) renderChemicalModels(state *models.ApplicationState) string {
	// Convert models.DilutionCalculation to pkg.DilutionCalculation
	var calculations []models.DilutionCalculation
	for _, calc := range state.Calculations {
		calculations = append(calculations, models.DilutionCalculation{
			Chemical:    calc.Chemical,
			Dilution:    calc.Dilution,
			TotalVolume: calc.TotalVolume,
			Concentrate: calc.Concentrate,
			Water:       calc.Water,
			Time:        calc.Time,
		})
	}
	
	chemicals := models.GetCalculatedChemicals(calculations)
	components := models.ChemicalModelsToComponents(chemicals)
	
	return s.renderChemicalComponents(components, true)
}

// renderChemicalComponents renders chemical components with proper separation
func (s *CalculatedScreen) renderChemicalComponents(components []models.ChemicalComponent, highlight bool) string {
	if len(components) == 0 {
		return ""
	}
	
	// Get component lines
	var componentLines [][]string
	for _, component := range components {
		var rendered string
		if highlight {
			rendered = component.RenderWithHighlight(highlightStyle)
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

func (s *CalculatedScreen) renderActions() string {
	return "[T] Timer    [U] Use Fixer    [R] Change Rolls    [F] Change Film    [Q] Quit"
}

func (s *CalculatedScreen) HandleInput(key string, sm *models.StateMachine) bool {
	switch strings.ToLower(key) {
	case "u":
		sm.HandleFixerUsage()
		return true
	case "t":
		sm.TransitionTo(models.TimerScreenState)
		return true
	case "r":
		sm.TransitionTo(models.RollSelectionState)
		return true
	case "f":
		sm.TransitionTo(models.FilmSelectionState)
		return true
	case "q":
		return false
	}
	return true
}

// TimerScreen represents the timer screen
type TimerScreen struct{}

func (s *TimerScreen) Render(state *models.ApplicationState) string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🎞️  Film Development Timer"))
	b.WriteString("\n\n")

	// Main content in single box
	mainContent := s.renderMainContent(state)
	b.WriteString(mainBoxStyle.Render(mainContent))
	b.WriteString("\n")

	// Actions Section (borderless)
	actionsContent := s.renderActions(state)
	b.WriteString(actionsOnlyStyle.Render(actionsContent))

	return b.String()
}

func (s *TimerScreen) renderMainContent(state *models.ApplicationState) string {
	var b strings.Builder

	// Timer Display Section
	b.WriteString(sectionStyle.Render("Current Step"))
	b.WriteString("\n")
	b.WriteString(s.renderTimerContent(state))
	b.WriteString("\n\n")

	// Steps Progress Section
	b.WriteString(sectionStyle.Render("Development Steps"))
	b.WriteString("\n")
	b.WriteString(s.renderStepsContent(state))

	return b.String()
}

func (s *TimerScreen) renderTimerContent(state *models.ApplicationState) string {
	var b strings.Builder

	if state.TimerState == nil || len(state.TimerState.Steps) == 0 {
		b.WriteString(errorStyle.Render("No timer available"))
		return b.String()
	}

	currentStep := state.TimerState.GetCurrentStep()
	if currentStep == nil {
		b.WriteString(highlightStyle.Render("🎉 All steps completed!"))
		return b.String()
	}

	elapsed := state.TimerState.GetCurrentElapsed()
	remaining := state.TimerState.GetRemainingTime()
	isOvertime := state.TimerState.IsCurrentStepOvertime()

	// Step name and target time
	b.WriteString(fmt.Sprintf("Step: %s\n", highlightStyle.Render(currentStep.Name)))
	b.WriteString(fmt.Sprintf("Target Time: %s\n", models.FormatDuration(currentStep.Duration)))

	// Timer display
	elapsedStr := models.FormatDuration(elapsed)
	remainingStr := models.FormatDuration(remaining)

	if isOvertime {
		b.WriteString(fmt.Sprintf("Elapsed: %s ⚠️  OVERTIME\n", errorStyle.Render(elapsedStr)))
		b.WriteString(fmt.Sprintf("Overtime: %s\n", errorStyle.Render(models.FormatDuration(elapsed-currentStep.Duration))))
	} else {
		b.WriteString(fmt.Sprintf("Elapsed: %s\n", highlightStyle.Render(elapsedStr)))
		b.WriteString(fmt.Sprintf("Remaining: %s\n", remainingStr))
	}

	// Status
	status := ""
	if state.TimerState.IsRunning {
		if state.TimerState.IsPaused {
			status = dimStyle.Render("⏸️  PAUSED")
		} else {
			status = highlightStyle.Render("⏱️  RUNNING")
		}
	} else {
		status = dimStyle.Render("⏹️  STOPPED")
	}
	b.WriteString(fmt.Sprintf("Status: %s", status))

	return b.String()
}

func (s *TimerScreen) renderStepsContent(state *models.ApplicationState) string {
	var b strings.Builder

	if state.TimerState == nil || len(state.TimerState.Steps) == 0 {
		b.WriteString(dimStyle.Render("No steps available"))
		return b.String()
	}

	for i, step := range state.TimerState.Steps {
		icon := "○"
		style := dimStyle

		if step.Finished {
			icon = "✅"
			style = dimStyle
		} else if i == state.TimerState.CurrentStep {
			icon = "🔵"
			style = highlightStyle
		}

		stepText := fmt.Sprintf("%s %s (%s)", icon, step.Name, models.FormatDuration(step.Duration))
		b.WriteString(style.Render(stepText))
		b.WriteString("\n")
	}

	return b.String()
}

func (s *TimerScreen) renderActions(state *models.ApplicationState) string {
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

func (s *TimerScreen) HandleInput(key string, sm *models.StateMachine) bool {
	if sm.GetApplicationState().TimerState == nil {
		return true
	}

	switch key {
	case "space":
		if sm.GetApplicationState().TimerState.IsRunning {
			if sm.GetApplicationState().TimerState.IsPaused {
				sm.GetApplicationState().TimerState.ResumeTimer()
			} else {
				sm.GetApplicationState().TimerState.PauseTimer()
			}
		} else {
			sm.GetApplicationState().TimerState.StartTimer()
		}
		return true
	case "s":
		sm.GetApplicationState().TimerState.StopTimer()
		return true
	case "n":
		sm.GetApplicationState().TimerState.CompleteCurrentStep()
		return true
	case "r":
		sm.GetApplicationState().TimerState.Reset()
		return true
	case "esc":
		sm.GoBack()
		return true
	case "q":
		return false
	}
	return true
}

// GetScreenForState returns the appropriate screen for the given state
func GetScreenForState(state models.AppState) Screen {
	switch state {
	case models.MainScreenState:
		return &MainScreen{}
	case models.FilmSelectionState:
		return &FilmSelectionScreen{}
	case models.EISelectionState:
		return &EISelectionScreen{}
	case models.RollSelectionState:
		return &RollSelectionScreen{}
	case models.MixedRollInputState:
		return &MixedRollInputScreen{}
	case models.CalculatedScreenState:
		return &CalculatedScreen{}
	case models.TimerScreenState:
		return &TimerScreen{}
	default:
		return &MainScreen{}
	}
}
