package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg/models"
	"github.com/rs/zerolog/log"
)

// MixedRollInputModel represents the mixed roll input screen
type MixedRollInputModel struct {
	appState *models.AppModel
	styles   *Styles
	film35mm int
	film120mm int
}

// NewMixedRollInputModel creates a new mixed roll input model
func NewMixedRollInputModel(appState *models.AppModel, styles *Styles) *MixedRollInputModel {
	return &MixedRollInputModel{
		appState: appState,
		styles:   styles,
		film35mm: 0,
		film120mm: 0,
	}
}

// Update handles key presses for the mixed roll input screen
func (m *MixedRollInputModel) Update(msg tea.KeyMsg) (*models.AppState, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.film35mm < 6 {
			m.film35mm++
			log.Debug().Int("35mm", m.film35mm).Msg("35mm rolls increased")
		}
	case "down", "j":
		if m.film35mm > 0 {
			m.film35mm--
			log.Debug().Int("35mm", m.film35mm).Msg("35mm rolls decreased")
		}
	case "+", "=":
		if m.film120mm < 6 {
			m.film120mm++
			log.Debug().Int("120mm", m.film120mm).Msg("120mm rolls increased")
		}
	case "-", "_":
		if m.film120mm > 0 {
			m.film120mm--
			log.Debug().Int("120mm", m.film120mm).Msg("120mm rolls decreased")
		}
	case "enter":
		if m.film35mm > 0 || m.film120mm > 0 {
			m.appState.SetRolls(m.film35mm, m.film120mm)
			log.Debug().Int("35mm", m.film35mm).Int("120mm", m.film120mm).Msg("Mixed rolls confirmed")
			newState := models.StateCalculatedScreen
			return &newState, nil
		}
	case "r":
		m.film35mm = 0
		m.film120mm = 0
		log.Debug().Msg("Mixed rolls reset")
	}
	
	return nil, nil
}

// View renders the mixed roll input screen
func (m *MixedRollInputModel) View() string {
	var b strings.Builder
	
	// Title
	title := m.styles.Title.Render("🎞️  Film Development Calculator")
	b.WriteString(title)
	b.WriteString("\n\n")
	
	// Custom Mix Setup Section
	customMix := m.renderCustomMixSetup()
	b.WriteString(customMix)
	b.WriteString("\n\n")
	
	// Actions Section
	actions := m.renderActions()
	b.WriteString(actions)
	
	return b.String()
}

func (m *MixedRollInputModel) renderCustomMixSetup() string {
	var b strings.Builder
	
	b.WriteString("┌─── Custom Mix Setup ────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	
	// 35mm rolls
	line1 := fmt.Sprintf("│  35mm Rolls: [ %d ]    (↑/↓ or +/- to adjust)                                  │", m.film35mm)
	b.WriteString(line1)
	b.WriteString("\n")
	
	// 120mm rolls
	line2 := fmt.Sprintf("│  120mm Rolls: [ %d ]   (↑/↓ or +/- to adjust)                                  │", m.film120mm)
	b.WriteString(line2)
	b.WriteString("\n")
	
	b.WriteString("│                                                                                 │\n")
	
	// Total tank size
	tankSize := models.CalculateTankSize(m.film35mm, m.film120mm)
	line3 := fmt.Sprintf("│  Total Tank Size: [ %dml ]                                                      │", tankSize)
	b.WriteString(line3)
	b.WriteString("\n")
	
	b.WriteString("│                                                                                 │\n")
	
	// Controls
	b.WriteString("│  [ENTER] Confirm    [ESC] Back    [R] Reset                                    │\n")
	
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}

func (m *MixedRollInputModel) renderActions() string {
	var b strings.Builder
	
	b.WriteString("┌─── Actions ─────────────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString("│  [↑↓] Adjust 35mm    [+/-] Adjust 120mm    [ENTER] Confirm    [ESC] Back       │\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}
