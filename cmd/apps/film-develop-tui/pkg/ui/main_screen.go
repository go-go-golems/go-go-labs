package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg/models"
	"github.com/rs/zerolog/log"
)

// MainScreenModel represents the main screen
type MainScreenModel struct {
	appState *models.AppModel
	styles   *Styles
}

// NewMainScreenModel creates a new main screen model
func NewMainScreenModel(appState *models.AppModel, styles *Styles) *MainScreenModel {
	return &MainScreenModel{
		appState: appState,
		styles:   styles,
	}
}

// Update handles key presses for the main screen
func (m *MainScreenModel) Update(msg tea.KeyMsg) (*models.AppState, tea.Cmd) {
	switch msg.String() {
	case "f":
		log.Debug().Msg("Navigating to film selection")
		newState := models.StateFilmSelection
		return &newState, nil
	case "u":
		log.Debug().Msg("Navigating to fixer tracking")
		newState := models.StateFixerTracking
		return &newState, nil
	case "s":
		log.Debug().Msg("Navigating to settings")
		newState := models.StateSettings
		return &newState, nil
	}
	
	return nil, nil
}

// View renders the main screen
func (m *MainScreenModel) View() string {
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

func (m *MainScreenModel) renderFilmSetup() string {
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
	
	line1 := fmt.Sprintf("│  Film Type:    %-32s EI:  %-17s │", filmType, ei)
	b.WriteString(line1)
	b.WriteString("\n")
	
	// Rolls and Tank
	rolls := "[ -- ]"
	tank := "[ --ml ]"
	if m.appState.TankSize > 0 {
		rolls = fmt.Sprintf("[ %s ]", m.appState.GetRollDescription())
		tank = fmt.Sprintf("[ %dml ]", m.appState.TankSize)
	}
	
	line2 := fmt.Sprintf("│  Rolls:        %-32s Tank: %-14s │", rolls, tank)
	b.WriteString(line2)
	b.WriteString("\n")
	
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}

func (m *MainScreenModel) renderChemicals() string {
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
	
	var ilfosol, ilfostop, sprintFixer string
	if len(volumes) > 0 {
		ilfosol = fmt.Sprintf("%dml conc", volumes["ilfosol_3"]["concentrate"])
		ilfostop = fmt.Sprintf("%dml conc", volumes["ilfostop"]["concentrate"])
		sprintFixer = fmt.Sprintf("%dml conc", volumes["sprint_fixer"]["concentrate"])
	} else {
		ilfosol = "--ml conc"
		ilfostop = "--ml conc"
		sprintFixer = "--ml conc"
	}
	
	b.WriteString(fmt.Sprintf("│  %-14s │  %-14s │  %-40s │\n", ilfosol, ilfostop, sprintFixer))
	
	// Water volumes
	var ilfosolWater, ilfostopWater, sprintFixerWater string
	if len(volumes) > 0 {
		ilfosolWater = fmt.Sprintf("%dml water", volumes["ilfosol_3"]["water"])
		ilfostopWater = fmt.Sprintf("%dml water", volumes["ilfostop"]["water"])
		sprintFixerWater = fmt.Sprintf("%dml water", volumes["sprint_fixer"]["water"])
	} else {
		ilfosolWater = "--ml water"
		ilfostopWater = "--ml water"
		sprintFixerWater = "--ml water"
	}
	
	b.WriteString(fmt.Sprintf("│  %-14s │  %-14s │  %-40s │\n", ilfosolWater, ilfostopWater, sprintFixerWater))
	
	// Times
	devTime := "--:--"
	if m.appState.IsComplete() {
		if time, err := m.appState.GetDevelopmentTime(); err == nil {
			devTime = time
		}
	}
	
	timeLine := fmt.Sprintf("│  Time: %-7s │  Time: 0:10    │  Time: 2:30                                  │", devTime)
	b.WriteString(timeLine)
	b.WriteString("\n")
	
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}

func (m *MainScreenModel) renderFixerUsage() string {
	var b strings.Builder
	
	b.WriteString("┌─── Fixer Usage ─────────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	
	used := m.appState.FixerTracker.UsedRolls
	remaining := m.appState.FixerTracker.GetRemainingCapacity()
	
	line := fmt.Sprintf("│  Capacity: 24 rolls per liter    Used: %d rolls    Remaining: %d rolls          │", used, remaining)
	b.WriteString(line)
	b.WriteString("\n")
	
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}

func (m *MainScreenModel) renderActions() string {
	var b strings.Builder
	
	b.WriteString("┌─── Actions ─────────────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString("│  [F] Film Type    [U] Fixer Usage    [S] Settings    [Q] Quit                  │\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}
