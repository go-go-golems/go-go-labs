package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg/models"
	"github.com/rs/zerolog/log"
)

// FilmSelectionModel represents the film selection screen
type FilmSelectionModel struct {
	appState *models.AppModel
	styles   *Styles
}

// NewFilmSelectionModel creates a new film selection model
func NewFilmSelectionModel(appState *models.AppModel, styles *Styles) *FilmSelectionModel {
	return &FilmSelectionModel{
		appState: appState,
		styles:   styles,
	}
}

// Update handles key presses for the film selection screen
func (m *FilmSelectionModel) Update(msg tea.KeyMsg) (*models.AppState, tea.Cmd) {
	films := models.GetFilmOptions()
	
	switch msg.String() {
	case "1":
		if len(films) >= 1 {
			m.appState.SetFilm(films[0].ID)
			log.Debug().Str("film", films[0].Name).Msg("Film selected")
			newState := models.StateEISelection
			return &newState, nil
		}
	case "2":
		if len(films) >= 2 {
			m.appState.SetFilm(films[1].ID)
			log.Debug().Str("film", films[1].Name).Msg("Film selected")
			newState := models.StateEISelection
			return &newState, nil
		}
	case "3":
		if len(films) >= 3 {
			m.appState.SetFilm(films[2].ID)
			log.Debug().Str("film", films[2].Name).Msg("Film selected")
			newState := models.StateEISelection
			return &newState, nil
		}
	case "4":
		if len(films) >= 4 {
			m.appState.SetFilm(films[3].ID)
			log.Debug().Str("film", films[3].Name).Msg("Film selected")
			newState := models.StateEISelection
			return &newState, nil
		}
	case "5":
		if len(films) >= 5 {
			m.appState.SetFilm(films[4].ID)
			log.Debug().Str("film", films[4].Name).Msg("Film selected")
			newState := models.StateEISelection
			return &newState, nil
		}
	case "6":
		if len(films) >= 6 {
			m.appState.SetFilm(films[5].ID)
			log.Debug().Str("film", films[5].Name).Msg("Film selected")
			newState := models.StateEISelection
			return &newState, nil
		}
	case "7":
		if len(films) >= 7 {
			m.appState.SetFilm(films[6].ID)
			log.Debug().Str("film", films[6].Name).Msg("Film selected")
			newState := models.StateEISelection
			return &newState, nil
		}
	}
	
	return nil, nil
}

// View renders the film selection screen
func (m *FilmSelectionModel) View() string {
	var b strings.Builder
	
	// Title
	title := m.styles.Title.Render("🎞️  Film Development Calculator")
	b.WriteString(title)
	b.WriteString("\n\n")
	
	// Film Selection Section
	filmSelection := m.renderFilmSelection()
	b.WriteString(filmSelection)
	b.WriteString("\n\n")
	
	// Actions Section
	actions := m.renderActions()
	b.WriteString(actions)
	
	return b.String()
}

func (m *FilmSelectionModel) renderFilmSelection() string {
	var b strings.Builder
	
	b.WriteString("┌─── Select Film Type ────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	
	films := models.GetFilmOptions()
	descriptions := []string{
		"📈 Most popular",
		"🎯 Fine grain",
		"🔍 Ultra fine",
		"⚖️  Versatile",
		"🌙 High speed",
		"💎 Finest grain",
		"🔴 Infrared",
	}
	
	for i, film := range films {
		if i < len(descriptions) {
			eiRange := m.formatEIRange(film.EIRatings)
			line := fmt.Sprintf("│  [%d] %-15s (EI %s)    %-31s │", i+1, film.Name, eiRange, descriptions[i])
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	
	b.WriteString("│                                                                                 │\n")
	b.WriteString("│  [ESC] Back                                                                     │\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}

func (m *FilmSelectionModel) formatEIRange(eiRatings []int) string {
	if len(eiRatings) == 1 {
		return fmt.Sprintf("%d", eiRatings[0])
	}
	
	var parts []string
	for _, ei := range eiRatings {
		parts = append(parts, fmt.Sprintf("%d", ei))
	}
	
	return strings.Join(parts, "/")
}

func (m *FilmSelectionModel) renderActions() string {
	var b strings.Builder
	
	b.WriteString("┌─── Actions ─────────────────────────────────────────────────────────────────────┐\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString("│  [1-7] Select Film    [ESC] Back    [Q] Quit                                   │\n")
	b.WriteString("│                                                                                 │\n")
	b.WriteString("└─────────────────────────────────────────────────────────────────────────────────┘")
	
	return b.String()
}
