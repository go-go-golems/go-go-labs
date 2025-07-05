package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/state"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/types"
)

// FilmSelectionScreen represents the film selection screen
type FilmSelectionScreen struct{}

func (s *FilmSelectionScreen) Render(appState *state.ApplicationState) string {
	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("🎞️  Film Development Calculator"))
	b.WriteString("\n\n")

	// Film Selection
	filmSelectionContent := s.renderFilmSelection(appState)
	b.WriteString(BoxStyle.Render(filmSelectionContent))
	b.WriteString("\n\n")

	// Actions
	actionsContent := s.renderActions()
	b.WriteString(BoxStyle.Render(actionsContent))

	return b.String()
}

func (s *FilmSelectionScreen) renderFilmSelection(appState *state.ApplicationState) string {
	var b strings.Builder
	b.WriteString(SectionStyle.Render("Select Film Type"))
	b.WriteString("\n\n")

	filmOrder := types.GetFilmOrder()
	for i, filmID := range filmOrder {
		if film, ok := appState.FilmDB.GetFilmByID(filmID); ok {
			eiRatings := make([]string, len(film.EIRatings))
			for j, ei := range film.EIRatings {
				eiRatings[j] = strconv.Itoa(ei)
			}
			eiStr := strings.Join(eiRatings, "/")
			if len(film.EIRatings) > 3 {
				eiStr = fmt.Sprintf("%d-%d", film.EIRatings[0], film.EIRatings[len(film.EIRatings)-1])
			}

			b.WriteString(fmt.Sprintf(ActionStyle.Render("[%d]")+" %-12s (EI %-10s) %s %s\n",
				i+1, film.Name, eiStr, film.Icon, film.Description))
		}
	}

	b.WriteString("\n")
	b.WriteString(DimStyle.Render("[ESC] Back"))

	return b.String()
}

func (s *FilmSelectionScreen) renderActions() string {
	var b strings.Builder
	b.WriteString(SectionStyle.Render("Actions"))
	b.WriteString("\n\n")

	b.WriteString(ActionStyle.Render("[1-7] Select Film    [ESC] Back    [Q] Quit"))

	return b.String()
}

func (s *FilmSelectionScreen) HandleInput(key string, sm *state.StateMachine) bool {
	switch strings.ToLower(key) {
	case "1", "2", "3", "4", "5", "6", "7":
		index, _ := strconv.Atoi(key)
		filmOrder := types.GetFilmOrder()
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