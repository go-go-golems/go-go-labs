package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg/models"
	"github.com/rs/zerolog/log"
)

// EISelectionModel represents the EI selection screen
type EISelectionModel struct {
	appState *models.AppModel
	styles   *Styles
}

// NewEISelectionModel creates a new EI selection model
func NewEISelectionModel(appState *models.AppModel, styles *Styles) *EISelectionModel {
	return &EISelectionModel{
		appState: appState,
		styles:   styles,
	}
}

// Update handles key presses for the EI selection screen
func (m *EISelectionModel) Update(msg tea.KeyMsg) (*models.AppState, tea.Cmd) {
	if m.appState.Film == nil {
		return nil, nil
	}
	
	switch msg.String() {
	case "1":
		if len(m.appState.Film.EIRatings) >= 1 {
			m.appState.SetEI(m.appState.Film.EIRatings[0])
			log.Debug().Int("ei", m.appState.Film.EIRatings[0]).Msg("EI selected")
			newState := models.StateRollSelection
			return &newState, nil
		}
	case "2":
		if len(m.appState.Film.EIRatings) >= 2 {
			m.appState.SetEI(m.appState.Film.EIRatings[1])
			log.Debug().Int("ei", m.appState.Film.EIRatings[1]).Msg("EI selected")
			newState := models.StateRollSelection
			return &newState, nil
		}
	case "3":
		if len(m.appState.Film.EIRatings) >= 3 {
			m.appState.SetEI(m.appState.Film.EIRatings[2])
			log.Debug().Int("ei", m.appState.Film.EIRatings[2]).Msg("EI selected")
			newState := models.StateRollSelection
			return &newState, nil
		}
	case "4":
		if len(m.appState.Film.EIRatings) >= 4 {
			m.appState.SetEI(m.appState.Film.EIRatings[3])
			log.Debug().Int("ei", m.appState.Film.EIRatings[3]).Msg("EI selected")
			newState := models.StateRollSelection
			return &newState, nil
		}
	case "5":
		if len(m.appState.Film.EIRatings) >= 5 {
			m.appState.SetEI(m.appState.Film.EIRatings[4])
			log.Debug().Int("ei", m.appState.Film.EIRatings[4]).Msg("EI selected")
			newState := models.StateRollSelection
			return &newState, nil
		}
	}
	
	return nil, nil
}

// View renders the EI selection screen
func (m *EISelectionModel) View() string {
	var b strings.Builder
	
	// Title
	title := m.styles.Title.Render("🎞️  Film Development Calculator")
	b.WriteString(title)
	b.WriteString("\n\n")
	
	// Film Setup Section
	filmSetup := m.renderFilmSetup()
	b.WriteString(filmSetup)
	b.WriteString("\n\n")
	
	// EI Selection Section
	eiSelection := m.renderEISelection()
	b.WriteString(eiSelection)
	b.WriteString("\n\n")
	
	// Actions Section
	actions := m.renderActions()
	b.WriteString(actions)
	
	return b.String()
}

func (m *EISelectionModel) renderFilmSetup() string {
	var b strings.Builder
	
	b.WriteString("┌─── Film Setup ──────────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	
	// Film Type and EI
	filmType := "[ Not Selected ]"
	ei := "[ Not Set ]"
	if m.appState.Film != nil {
		filmType = fmt.Sprintf("[ %s ]", m.appState.Film.Name)
	}
	
	line1 := fmt.Sprintf("│  Film Type:    %-32s EI:  %-15s │", filmType, ei)
	b.WriteString(line1)
	b.WriteString("\n")
	
	// Rolls and Tank
	rolls := "[ -- ]"
	tank := "[ --ml ]"
	
	line2 := fmt.Sprintf("│  Rolls:        %-32s Tank: %-14s │", rolls, tank)
	b.WriteString(line2)
	b.WriteString("\n")
	
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}

func (m *EISelectionModel) renderEISelection() string {
	var b strings.Builder
	
	b.WriteString("┌─── Select EI Rating ────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	
	if m.appState.Film == nil {
		b.WriteString("│  No film selected                                                              │\n")
		b.WriteString("│                                                                                 │\n")
		b.WriteString("│  [ESC] Back to film selection                                                   │\n")
		b.WriteString("│                                                                                 │\n")
		b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
		return b.String()
	}
	
	descriptions := []string{
		"🌞 Bright light, fine grain",
		"📷 Standard, most common",
		"🌆 Low light, pushed grain",
		"🌙 High speed, grainy",
		"🔥 Extreme high speed",
	}
	
	for i, ei := range m.appState.Film.EIRatings {
		// Get development time if available
		timeStr := "--:--"
		if time, ok := m.appState.Film.Times20C[m.appState.Dilution][ei]; ok {
			timeStr = time
		}
		
		desc := ""
		if i < len(descriptions) {
			desc = descriptions[i]
		}
		
		line := fmt.Sprintf("│  [%d] EI %d  (%s @ %s)     %-35s │", i+1, ei, timeStr, m.appState.Dilution, desc)
		b.WriteString(line)
		b.WriteString("\n")
	}
	
	b.WriteString("│                                                                                 │\n")
	b.WriteString("│  [ESC] Back to film selection                                                   │\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}

func (m *EISelectionModel) renderActions() string {
	var b strings.Builder
	
	maxOptions := 0
	if m.appState.Film != nil {
		maxOptions = len(m.appState.Film.EIRatings)
	}
	
	actionText := fmt.Sprintf("│  [1-%d] Select EI    [ESC] Back    [Q] Quit                                     │", maxOptions)
	if maxOptions == 0 {
		actionText = "│  [ESC] Back    [Q] Quit                                                         │"
	}
	
	b.WriteString("┌─── Actions ─────────────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString(actionText)
	b.WriteString("\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}
