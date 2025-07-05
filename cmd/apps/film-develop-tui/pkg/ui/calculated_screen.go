package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg/models"
	"github.com/rs/zerolog/log"
)

// CalculatedScreenModel represents the calculated results screen
type CalculatedScreenModel struct {
	appState *models.AppModel
	styles   *Styles
}

// NewCalculatedScreenModel creates a new calculated screen model
func NewCalculatedScreenModel(appState *models.AppModel, styles *Styles) *CalculatedScreenModel {
	return &CalculatedScreenModel{
		appState: appState,
		styles:   styles,
	}
}

// Update handles key presses for the calculated screen
func (m *CalculatedScreenModel) Update(msg tea.KeyMsg) (*models.AppState, tea.Cmd) {
	switch msg.String() {
	case "u":
		if m.appState.CanUseFixer() {
			m.appState.UseFixer()
			log.Debug().Int("rollsUsed", m.appState.GetRollsUsed()).Msg("Fixer used")
		}
	case "r":
		log.Debug().Msg("Navigating back to roll selection")
		newState := models.StateRollSelection
		return &newState, nil
	case "f":
		log.Debug().Msg("Navigating back to film selection")
		newState := models.StateFilmSelection
		return &newState, nil
	}
	
	return nil, nil
}

// View renders the calculated results screen
func (m *CalculatedScreenModel) View() string {
	var b strings.Builder
	
	// Title
	title := m.styles.Title.Render("🎞️  Film Development Calculator")
	b.WriteString(title)
	b.WriteString("\n\n")
	
	// Film Setup Section
	filmSetup := m.renderFilmSetup()
	b.WriteString(filmSetup)
	b.WriteString("\n\n")
	
	// Chemicals Section
	chemicals := m.renderChemicals()
	b.WriteString(chemicals)
	b.WriteString("\n\n")
	
	// Fixer Usage Section
	fixerUsage := m.renderFixerUsage()
	b.WriteString(fixerUsage)
	b.WriteString("\n\n")
	
	// Actions Section
	actions := m.renderActions()
	b.WriteString(actions)
	
	return b.String()
}

func (m *CalculatedScreenModel) renderFilmSetup() string {
	var b strings.Builder
	
	b.WriteString("┌─── Film Setup ──────────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	
	// Film Type and EI
	filmType := "[ Not Selected ]"
	ei := "[ -- ]"
	if m.appState.Film != nil {
		filmType = fmt.Sprintf("[ %s ]", m.appState.Film.Name)
		if m.appState.EIRating > 0 {
			ei = fmt.Sprintf("[ %d ]", m.appState.EIRating)
		}
	}
	
	line1 := fmt.Sprintf("│  Film Type:    %-32s EI:  %-16s │", filmType, ei)
	b.WriteString(line1)
	b.WriteString("\n")
	
	// Rolls and Tank
	rolls := fmt.Sprintf("[ %s ]", m.appState.GetRollDescription())
	tank := fmt.Sprintf("[ %dml ]", m.appState.TankSize)
	
	line2 := fmt.Sprintf("│  Rolls:        %-32s Tank: %-13s │", rolls, tank)
	b.WriteString(line2)
	b.WriteString("\n")
	
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}

func (m *CalculatedScreenModel) renderChemicals() string {
	var b strings.Builder
	
	b.WriteString("┌─── Chemicals (20°C) ────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	
	// Headers
	b.WriteString("│  ILFOSOL 3     │  ILFOSTOP      │  SPRINT FIXER                                │\n")
	
	// Dilutions
	dilution := fmt.Sprintf("%s dilution", m.appState.Dilution)
	b.WriteString(fmt.Sprintf("│  %-14s │  1+19 dilution │  1+4 dilution                                │\n", dilution))
	
	// Volumes
	volumes := m.appState.GetChemicalVolumes()
	
	// Concentrate volumes
	ilfosol := fmt.Sprintf("%dml conc", volumes["ilfosol_3"]["concentrate"])
	ilfostop := fmt.Sprintf("%dml conc", volumes["ilfostop"]["concentrate"])
	sprintFixer := fmt.Sprintf("%dml conc", volumes["sprint_fixer"]["concentrate"])
	
	b.WriteString(fmt.Sprintf("│  %-14s │  %-14s │  %-40s │\n", ilfosol, ilfostop, sprintFixer))
	
	// Water volumes
	ilfosolWater := fmt.Sprintf("%dml water", volumes["ilfosol_3"]["water"])
	ilfostopWater := fmt.Sprintf("%dml water", volumes["ilfostop"]["water"])
	sprintFixerWater := fmt.Sprintf("%dml water", volumes["sprint_fixer"]["water"])
	
	b.WriteString(fmt.Sprintf("│  %-14s │  %-14s │  %-40s │\n", ilfosolWater, ilfostopWater, sprintFixerWater))
	
	// Times
	devTime := "--:--"
	if time, err := m.appState.GetDevelopmentTime(); err == nil {
		devTime = time
	}
	
	timeLine := fmt.Sprintf("│  Time: %-7s │  Time: 0:10    │  Time: 2:30                                  │", devTime)
	b.WriteString(timeLine)
	b.WriteString("\n")
	
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}

func (m *CalculatedScreenModel) renderFixerUsage() string {
	var b strings.Builder
	
	b.WriteString("┌─── Fixer Usage ─────────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	
	used := m.appState.FixerTracker.UsedRolls
	remaining := m.appState.FixerTracker.GetRemainingCapacity()
	
	line1 := fmt.Sprintf("│  Capacity: 24 rolls per liter    Used: %d rolls    Remaining: %d rolls          │", used, remaining)
	b.WriteString(line1)
	b.WriteString("\n")
	
	rollsToUse := m.appState.GetRollsUsed()
	remainingAfterUse := remaining - rollsToUse
	
	line2 := fmt.Sprintf("│  This batch uses: %d roll         After use: %d rolls remaining                 │", rollsToUse, remainingAfterUse)
	b.WriteString(line2)
	b.WriteString("\n")
	
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}

func (m *CalculatedScreenModel) renderActions() string {
	var b strings.Builder
	
	b.WriteString("┌─── Actions ─────────────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	
	actionText := "│  [U] Use Fixer    [R] Change Rolls    [F] Change Film    [Q] Quit              │"
	if !m.appState.CanUseFixer() {
		actionText = "│  [!] Fixer Full   [R] Change Rolls    [F] Change Film    [Q] Quit              │"
	}
	
	b.WriteString(actionText)
	b.WriteString("\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}
