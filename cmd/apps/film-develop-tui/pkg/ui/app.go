package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg/models"
	"github.com/rs/zerolog/log"
)

// keyMap defines the key bindings for the application
type keyMap struct {
	Quit key.Binding
	Back key.Binding
}

// AppModel represents the main application model
type AppModel struct {
	appState       *models.AppModel
	keyMap         keyMap
	styles         *Styles
	width          int
	height         int
	
	// Sub-models
	mainScreen      *MainScreenModel
	filmSelection   *FilmSelectionModel
	eiSelection     *EISelectionModel
	rollSelection   *RollSelectionModel
	mixedRollInput  *MixedRollInputModel
	calculatedScreen *CalculatedScreenModel
}

// NewAppModel creates a new application model
func NewAppModel() *AppModel {
	appState := models.NewAppModel()
	styles := NewStyles()
	
	return &AppModel{
		appState:         appState,
		keyMap:           defaultKeyMap(),
		styles:           styles,
		mainScreen:       NewMainScreenModel(appState, styles),
		filmSelection:    NewFilmSelectionModel(appState, styles),
		eiSelection:      NewEISelectionModel(appState, styles),
		rollSelection:    NewRollSelectionModel(appState, styles),
		mixedRollInput:   NewMixedRollInputModel(appState, styles),
		calculatedScreen: NewCalculatedScreenModel(appState, styles),
	}
}

func defaultKeyMap() keyMap {
	return keyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

// Init initializes the application
func (m *AppModel) Init() tea.Cmd {
	log.Debug().Msg("Initializing application")
	return nil
}

// Update handles messages and updates the application state
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		log.Debug().Int("width", m.width).Int("height", m.height).Msg("Window resized")
		return m, nil
		
	case tea.KeyMsg:
		// Handle global key bindings
		if key.Matches(msg, m.keyMap.Quit) {
			log.Debug().Msg("Quit key pressed")
			return m, tea.Quit
		}
		
		if key.Matches(msg, m.keyMap.Back) {
			return m.handleBackNavigation()
		}
		
		// Forward to current screen
		return m.forwardToCurrentScreen(msg)
	}
	
	return m, cmd
}

// View renders the current view
func (m *AppModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}
	
	switch m.appState.CurrentState {
	case models.StateMainScreen:
		return m.mainScreen.View()
	case models.StateFilmSelection:
		return m.filmSelection.View()
	case models.StateEISelection:
		return m.eiSelection.View()
	case models.StateRollSelection:
		return m.rollSelection.View()
	case models.StateMixedRollInput:
		return m.mixedRollInput.View()
	case models.StateCalculatedScreen:
		return m.calculatedScreen.View()
	default:
		return m.styles.Error.Render("Unknown state: " + m.appState.CurrentState.String())
	}
}

// handleBackNavigation handles the back navigation logic
func (m *AppModel) handleBackNavigation() (*AppModel, tea.Cmd) {
	log.Debug().Str("currentState", m.appState.CurrentState.String()).Msg("Handling back navigation")
	
	switch m.appState.CurrentState {
	case models.StateFilmSelection:
		m.appState.CurrentState = models.StateMainScreen
	case models.StateEISelection:
		m.appState.CurrentState = models.StateFilmSelection
	case models.StateRollSelection:
		m.appState.CurrentState = models.StateEISelection
	case models.StateMixedRollInput:
		m.appState.CurrentState = models.StateRollSelection
	case models.StateCalculatedScreen:
		m.appState.CurrentState = models.StateRollSelection
	default:
		// Main screen or unknown state - don't navigate back
		return m, nil
	}
	
	return m, nil
}

// forwardToCurrentScreen forwards the message to the current screen
func (m *AppModel) forwardToCurrentScreen(msg tea.KeyMsg) (*AppModel, tea.Cmd) {
	var cmd tea.Cmd
	
	switch m.appState.CurrentState {
	case models.StateMainScreen:
		newState, newCmd := m.mainScreen.Update(msg)
		if newState != nil {
			m.appState.CurrentState = *newState
		}
		cmd = newCmd
		
	case models.StateFilmSelection:
		newState, newCmd := m.filmSelection.Update(msg)
		if newState != nil {
			m.appState.CurrentState = *newState
		}
		cmd = newCmd
		
	case models.StateEISelection:
		newState, newCmd := m.eiSelection.Update(msg)
		if newState != nil {
			m.appState.CurrentState = *newState
		}
		cmd = newCmd
		
	case models.StateRollSelection:
		newState, newCmd := m.rollSelection.Update(msg)
		if newState != nil {
			m.appState.CurrentState = *newState
		}
		cmd = newCmd
		
	case models.StateMixedRollInput:
		newState, newCmd := m.mixedRollInput.Update(msg)
		if newState != nil {
			m.appState.CurrentState = *newState
		}
		cmd = newCmd
		
	case models.StateCalculatedScreen:
		newState, newCmd := m.calculatedScreen.Update(msg)
		if newState != nil {
			m.appState.CurrentState = *newState
		}
		cmd = newCmd
	}
	
	return m, cmd
}

// Styles defines the styling for the application
type Styles struct {
	Title       lipgloss.Style
	Border      lipgloss.Style
	Section     lipgloss.Style
	Highlight   lipgloss.Style
	Error       lipgloss.Style
	Help        lipgloss.Style
	Success     lipgloss.Style
	Warning     lipgloss.Style
	Value       lipgloss.Style
	Label       lipgloss.Style
}

// NewStyles creates a new set of styles
func NewStyles() *Styles {
	return &Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Align(lipgloss.Center),
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1),
		Section: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		Highlight: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true),
		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")),
		Label: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
	}
}
