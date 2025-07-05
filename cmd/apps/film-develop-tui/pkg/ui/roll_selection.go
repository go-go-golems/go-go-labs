package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg/models"
	"github.com/rs/zerolog/log"
)

// RollSelectionModel represents the roll selection screen
type RollSelectionModel struct {
	appState *models.AppModel
	styles   *Styles
}

// NewRollSelectionModel creates a new roll selection model
func NewRollSelectionModel(appState *models.AppModel, styles *Styles) *RollSelectionModel {
	return &RollSelectionModel{
		appState: appState,
		styles:   styles,
	}
}

// Update handles key presses for the roll selection screen
func (m *RollSelectionModel) Update(msg tea.KeyMsg) (*models.AppState, tea.Cmd) {
	switch msg.String() {
	case "1":
		m.appState.SetRolls(1, 0)
		log.Debug().Int("35mm", 1).Int("120mm", 0).Msg("Rolls selected")
		newState := models.StateCalculatedScreen
		return &newState, nil
	case "2":
		m.appState.SetRolls(2, 0)
		log.Debug().Int("35mm", 2).Int("120mm", 0).Msg("Rolls selected")
		newState := models.StateCalculatedScreen
		return &newState, nil
	case "3":
		m.appState.SetRolls(3, 0)
		log.Debug().Int("35mm", 3).Int("120mm", 0).Msg("Rolls selected")
		newState := models.StateCalculatedScreen
		return &newState, nil
	case "4":
		m.appState.SetRolls(4, 0)
		log.Debug().Int("35mm", 4).Int("120mm", 0).Msg("Rolls selected")
		newState := models.StateCalculatedScreen
		return &newState, nil
	case "5":
		m.appState.SetRolls(5, 0)
		log.Debug().Int("35mm", 5).Int("120mm", 0).Msg("Rolls selected")
		newState := models.StateCalculatedScreen
		return &newState, nil
	case "6":
		m.appState.SetRolls(6, 0)
		log.Debug().Int("35mm", 6).Int("120mm", 0).Msg("Rolls selected")
		newState := models.StateCalculatedScreen
		return &newState, nil
	case "a", "A":
		m.appState.SetRolls(0, 1)
		log.Debug().Int("35mm", 0).Int("120mm", 1).Msg("Rolls selected")
		newState := models.StateCalculatedScreen
		return &newState, nil
	case "b", "B":
		m.appState.SetRolls(0, 2)
		log.Debug().Int("35mm", 0).Int("120mm", 2).Msg("Rolls selected")
		newState := models.StateCalculatedScreen
		return &newState, nil
	case "c", "C":
		m.appState.SetRolls(0, 3)
		log.Debug().Int("35mm", 0).Int("120mm", 3).Msg("Rolls selected")
		newState := models.StateCalculatedScreen
		return &newState, nil
	case "d", "D":
		m.appState.SetRolls(0, 4)
		log.Debug().Int("35mm", 0).Int("120mm", 4).Msg("Rolls selected")
		newState := models.StateCalculatedScreen
		return &newState, nil
	case "e", "E":
		m.appState.SetRolls(0, 5)
		log.Debug().Int("35mm", 0).Int("120mm", 5).Msg("Rolls selected")
		newState := models.StateCalculatedScreen
		return &newState, nil
	case "f", "F":
		m.appState.SetRolls(0, 6)
		log.Debug().Int("35mm", 0).Int("120mm", 6).Msg("Rolls selected")
		newState := models.StateCalculatedScreen
		return &newState, nil
	case "m", "M":
		log.Debug().Msg("Navigating to mixed roll input")
		newState := models.StateMixedRollInput
		return &newState, nil
	}
	
	return nil, nil
}

// View renders the roll selection screen
func (m *RollSelectionModel) View() string {
	var b strings.Builder
	
	// Title
	title := m.styles.Title.Render("🎞️  Film Development Calculator")
	b.WriteString(title)
	b.WriteString("\n\n")
	
	// Film Setup Section
	filmSetup := m.renderFilmSetup()
	b.WriteString(filmSetup)
	b.WriteString("\n\n")
	
	// Roll Selection Section
	rollSelection := m.renderRollSelection()
	b.WriteString(rollSelection)
	b.WriteString("\n\n")
	
	// Actions Section
	actions := m.renderActions()
	b.WriteString(actions)
	
	return b.String()
}

func (m *RollSelectionModel) renderFilmSetup() string {
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
	rolls := "[ -- ]"
	tank := "[ --ml ]"
	
	line2 := fmt.Sprintf("│  Rolls:        %-32s Tank: %-14s │", rolls, tank)
	b.WriteString(line2)
	b.WriteString("\n")
	
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}

func (m *RollSelectionModel) renderRollSelection() string {
	var b strings.Builder
	
	b.WriteString("┌─── Number of Rolls ─────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	
	// 35mm and 120mm headers
	b.WriteString("│  35mm Rolls:                           120mm Rolls:                            │\n")
	
	// First row
	b.WriteString("│  [1] 1 Roll (300ml)  [4] 4 Rolls       [A] 1 Roll (500ml)  [D] 4 Rolls        │\n")
	
	// Second row
	b.WriteString("│  [2] 2 Rolls (500ml) [5] 5 Rolls       [B] 2 Rolls (700ml) [E] 5 Rolls        │\n")
	
	// Third row
	b.WriteString("│  [3] 3 Rolls (600ml) [6] 6 Rolls       [C] 3 Rolls (900ml) [F] 6 Rolls        │\n")
	
	b.WriteString("│                                                                                 │\n")
	
	// Mixed batches
	b.WriteString("│  Mixed batches: [M] Custom mix                                                  │\n")
	
	b.WriteString("│                                                                                 │\n")
	b.WriteString("│  [ESC] Back to EI selection                                                     │\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}

func (m *RollSelectionModel) renderActions() string {
	var b strings.Builder
	
	b.WriteString("┌─── Actions ─────────────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString("│  [1-6] 35mm    [A-F] 120mm    [M] Mixed    [ESC] Back    [Q] Quit              │\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}
