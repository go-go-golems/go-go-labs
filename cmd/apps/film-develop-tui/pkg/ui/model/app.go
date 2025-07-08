package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg/data"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg/ui/keys"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg/ui/view"
)

// AppModel represents the main application model
type AppModel struct {
	data   *AppData
	keymap keys.KeyMap
	help   help.Model
	width  int
	height int
}

// NewAppModel creates a new application model
func NewAppModel() *AppModel {
	return &AppModel{
		data:   NewAppData(),
		keymap: keys.NewKeyMap(),
		help:   help.New(),
		width:  80,
		height: 24,
	}
}

// Init initializes the application
func (m *AppModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress handles keyboard input based on current state
func (m *AppModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys that work in all states
	switch {
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit
	}

	// State-specific key handling
	switch m.data.CurrentState {
	case StateMainScreen:
		return m.handleMainScreenKeys(msg)
	case StateFilmSelection:
		return m.handleFilmSelectionKeys(msg)
	case StateEISelection:
		return m.handleEISelectionKeys(msg)
	case StateRollSelection:
		return m.handleRollSelectionKeys(msg)
	case StateMixedRollInput:
		return m.handleMixedRollKeys(msg)
	case StateCalculatedScreen:
		return m.handleCalculatedScreenKeys(msg)
	}

	return m, nil
}

// handleMainScreenKeys handles keys for the main screen
func (m *AppModel) handleMainScreenKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Film):
		m.data.CurrentState = StateFilmSelection
		return m, nil
	case key.Matches(msg, m.keymap.Settings):
		m.data.CurrentState = StateSettings
		return m, nil
	case key.Matches(msg, m.keymap.UseFixer):
		m.data.CurrentState = StateFixerTracking
		return m, nil
	}
	return m, nil
}

// handleFilmSelectionKeys handles keys for film selection
func (m *AppModel) handleFilmSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Back):
		m.data.CurrentState = StateMainScreen
		return m, nil
	case key.Matches(msg, m.keymap.Key1):
		m.selectFilm(0)
		return m, nil
	case key.Matches(msg, m.keymap.Key2):
		m.selectFilm(1)
		return m, nil
	case key.Matches(msg, m.keymap.Key3):
		m.selectFilm(2)
		return m, nil
	case key.Matches(msg, m.keymap.Key4):
		m.selectFilm(3)
		return m, nil
	case key.Matches(msg, m.keymap.Key5):
		m.selectFilm(4)
		return m, nil
	case key.Matches(msg, m.keymap.Key6):
		m.selectFilm(5)
		return m, nil
	case key.Matches(msg, m.keymap.Key7):
		m.selectFilm(6)
		return m, nil
	}
	return m, nil
}

// handleEISelectionKeys handles keys for EI selection
func (m *AppModel) handleEISelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Back):
		m.data.CurrentState = StateFilmSelection
		return m, nil
	case key.Matches(msg, m.keymap.Key1):
		m.selectEI(0)
		return m, nil
	case key.Matches(msg, m.keymap.Key2):
		m.selectEI(1)
		return m, nil
	case key.Matches(msg, m.keymap.Key3):
		m.selectEI(2)
		return m, nil
	case key.Matches(msg, m.keymap.Key4):
		m.selectEI(3)
		return m, nil
	case key.Matches(msg, m.keymap.Key5):
		m.selectEI(4)
		return m, nil
	}
	return m, nil
}

// handleRollSelectionKeys handles keys for roll selection
func (m *AppModel) handleRollSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Back):
		m.data.CurrentState = StateEISelection
		return m, nil
	case key.Matches(msg, m.keymap.Mixed):
		m.data.CurrentState = StateMixedRollInput
		return m, nil
	// 35mm rolls
	case key.Matches(msg, m.keymap.Key1):
		m.selectRolls(1, 0)
		return m, nil
	case key.Matches(msg, m.keymap.Key2):
		m.selectRolls(2, 0)
		return m, nil
	case key.Matches(msg, m.keymap.Key3):
		m.selectRolls(3, 0)
		return m, nil
	case key.Matches(msg, m.keymap.Key4):
		m.selectRolls(4, 0)
		return m, nil
	case key.Matches(msg, m.keymap.Key5):
		m.selectRolls(5, 0)
		return m, nil
	case key.Matches(msg, m.keymap.Key6):
		m.selectRolls(6, 0)
		return m, nil
	// 120mm rolls
	case key.Matches(msg, m.keymap.KeyA):
		m.selectRolls(0, 1)
		return m, nil
	case key.Matches(msg, m.keymap.KeyB):
		m.selectRolls(0, 2)
		return m, nil
	case key.Matches(msg, m.keymap.KeyC):
		m.selectRolls(0, 3)
		return m, nil
	case key.Matches(msg, m.keymap.KeyD):
		m.selectRolls(0, 4)
		return m, nil
	case key.Matches(msg, m.keymap.KeyE):
		m.selectRolls(0, 5)
		return m, nil
	case key.Matches(msg, m.keymap.KeyF):
		m.selectRolls(0, 6)
		return m, nil
	}
	return m, nil
}

// handleMixedRollKeys handles keys for mixed roll input
func (m *AppModel) handleMixedRollKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Back):
		m.data.CurrentState = StateRollSelection
		return m, nil
	case key.Matches(msg, m.keymap.Enter):
		if m.data.FilmSetup.Rolls35mm > 0 || m.data.FilmSetup.Rolls120mm > 0 {
			m.data.UpdateTankSize()
			m.data.CalculateChemicals()
			m.data.CurrentState = StateCalculatedScreen
		}
		return m, nil
	case key.Matches(msg, m.keymap.Reset):
		m.data.FilmSetup.Rolls35mm = 0
		m.data.FilmSetup.Rolls120mm = 0
		m.data.UpdateTankSize()
		return m, nil
	case key.Matches(msg, m.keymap.Up):
		if m.data.FilmSetup.Rolls35mm < 6 {
			m.data.FilmSetup.Rolls35mm++
			m.data.UpdateTankSize()
		}
		return m, nil
	case key.Matches(msg, m.keymap.Down):
		if m.data.FilmSetup.Rolls35mm > 0 {
			m.data.FilmSetup.Rolls35mm--
			m.data.UpdateTankSize()
		}
		return m, nil
	case key.Matches(msg, m.keymap.Plus):
		if m.data.FilmSetup.Rolls120mm < 6 {
			m.data.FilmSetup.Rolls120mm++
			m.data.UpdateTankSize()
		}
		return m, nil
	case key.Matches(msg, m.keymap.Minus):
		if m.data.FilmSetup.Rolls120mm > 0 {
			m.data.FilmSetup.Rolls120mm--
			m.data.UpdateTankSize()
		}
		return m, nil
	}
	return m, nil
}

// handleCalculatedScreenKeys handles keys for calculated results screen
func (m *AppModel) handleCalculatedScreenKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.UseFixer):
		m.data.UseFixer()
		return m, nil
	case key.Matches(msg, m.keymap.ChangeRoll):
		m.data.CurrentState = StateRollSelection
		return m, nil
	case key.Matches(msg, m.keymap.Film):
		m.data.CurrentState = StateFilmSelection
		return m, nil
	}
	return m, nil
}

// selectFilm selects a film by index
func (m *AppModel) selectFilm(index int) {
	films := data.GetFilmList()
	if index >= 0 && index < len(films) {
		m.data.FilmSetup.SelectedFilm = &films[index]
		m.data.CurrentState = StateEISelection
	}
}

// selectEI selects an EI rating by index
func (m *AppModel) selectEI(index int) {
	if m.data.FilmSetup.SelectedFilm != nil {
		ratings := m.data.FilmSetup.SelectedFilm.EIRatings
		if index >= 0 && index < len(ratings) {
			m.data.FilmSetup.SelectedEI = ratings[index]
			m.data.CurrentState = StateRollSelection
		}
	}
}

// selectRolls selects roll counts and moves to calculated screen
func (m *AppModel) selectRolls(rolls35mm, rolls120mm int) {
	m.data.FilmSetup.Rolls35mm = rolls35mm
	m.data.FilmSetup.Rolls120mm = rolls120mm
	m.data.UpdateTankSize()
	m.data.CalculateChemicals()
	m.data.CurrentState = StateCalculatedScreen
}

// View renders the current view
func (m *AppModel) View() string {
	switch m.data.CurrentState {
	case StateMainScreen:
		return m.renderMainScreen()
	case StateFilmSelection:
		return m.renderFilmSelection()
	case StateEISelection:
		return m.renderEISelection()
	case StateRollSelection:
		return m.renderRollSelection()
	case StateMixedRollInput:
		return m.renderMixedRollInput()
	case StateCalculatedScreen:
		return m.renderCalculatedScreen()
	default:
		return "Unknown state"
	}
}

// renderMainScreen renders the main screen
func (m *AppModel) renderMainScreen() string {
	var sections []string
	
	// Title
	sections = append(sections, view.RenderTitle())
	
	// Film Setup section
	filmName := "--"
	if m.data.FilmSetup.SelectedFilm != nil {
		filmName = m.data.FilmSetup.SelectedFilm.Name
	}
	
	eiValue := "--"
	if m.data.FilmSetup.SelectedEI > 0 {
		eiValue = strconv.Itoa(m.data.FilmSetup.SelectedEI)
	}
	
	tankValue := "--ml"
	if m.data.FilmSetup.TankSize > 0 {
		tankValue = fmt.Sprintf("%dml", m.data.FilmSetup.TankSize)
	}
	
	setupContent := fmt.Sprintf("Film Type:    %s                    EI:  %s\nRolls:        %s                              Tank: %s",
		view.RenderValue(filmName),
		view.RenderValue(eiValue),
		view.RenderValue(m.data.GetRollsDescription()),
		view.RenderValue(tankValue))
	
	sections = append(sections, view.RenderSection("Film Setup", setupContent))
	
	// Chemicals section
	chemicalsContent := m.renderChemicalsTable()
	sections = append(sections, view.RenderSection("Chemicals (20°C)", chemicalsContent))
	
	// Fixer Usage section
	fixerContent := fmt.Sprintf("Capacity: %d rolls per liter    Used: %d rolls    Remaining: %d rolls",
		m.data.FixerState.TotalCapacity,
		m.data.FixerState.UsedRolls,
		m.data.FixerState.RemainingRolls)
	sections = append(sections, view.RenderSection("Fixer Usage", fixerContent))
	
	// Actions section
	actionsContent := view.RenderKeyBinding("F", "Film Type") + "    " +
		view.RenderKeyBinding("U", "Fixer Usage") + "    " +
		view.RenderKeyBinding("S", "Settings") + "    " +
		view.RenderKeyBinding("Q", "Quit")
	sections = append(sections, view.RenderSection("Actions", actionsContent))
	
	return view.Styles.MainContainer.Render(strings.Join(sections, "\n"))
}

// renderChemicalsTable renders the chemicals table
func (m *AppModel) renderChemicalsTable() string {
	if len(m.data.Chemicals) == 0 {
		return view.Styles.Placeholder.Render("Select film type to see chemical calculations")
	}
	
	var columns []string
	for _, chemical := range m.data.Chemicals {
		concentrate := fmt.Sprintf("%dml conc", chemical.Concentrate)
		water := fmt.Sprintf("%dml water", chemical.Water)
		time := formatDuration(chemical.Time)
		
		columns = append(columns, view.RenderChemicalColumn(
			chemical.Name,
			chemical.Dilution,
			concentrate,
			water,
			time))
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

// formatDuration formats a duration as MM:SS
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "--:--"
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// renderFilmSelection renders the film selection screen
func (m *AppModel) renderFilmSelection() string {
	var sections []string
	
	// Title
	sections = append(sections, view.RenderTitle())
	
	// Film options
	var filmOptions []string
	films := data.GetFilmList()
	for i, film := range films {
		var ratings []string
		for _, rating := range film.EIRatings {
			ratings = append(ratings, strconv.Itoa(rating))
		}
		ratingsStr := "EI " + strings.Join(ratings, "/")
		
		filmOptions = append(filmOptions, view.RenderFilmOption(
			strconv.Itoa(i+1),
			film.Name,
			ratingsStr,
			film.Description,
			film.Icon))
	}
	
	filmOptions = append(filmOptions, "")
	filmOptions = append(filmOptions, view.RenderKeyBinding("ESC", "Back"))
	
	filmContent := strings.Join(filmOptions, "\n")
	sections = append(sections, view.RenderSection("Select Film Type", filmContent))
	
	// Actions
	actionsContent := view.RenderKeyBinding("1-7", "Select Film") + "    " +
		view.RenderKeyBinding("ESC", "Back") + "    " +
		view.RenderKeyBinding("Q", "Quit")
	sections = append(sections, view.RenderSection("Actions", actionsContent))
	
	return view.Styles.MainContainer.Render(strings.Join(sections, "\n"))
}

// renderEISelection renders the EI selection screen
func (m *AppModel) renderEISelection() string {
	var sections []string
	
	// Title
	sections = append(sections, view.RenderTitle())
	
	// Film Setup section
	filmName := "--"
	if m.data.FilmSetup.SelectedFilm != nil {
		filmName = m.data.FilmSetup.SelectedFilm.Name
	}
	
	setupContent := fmt.Sprintf("Film Type:    %s                        EI:  %s\nRolls:        %s                              Tank: %s",
		view.RenderValue(filmName),
		view.RenderValue("Not Set"),
		view.RenderValue("--"),
		view.RenderValue("--ml"))
	
	sections = append(sections, view.RenderSection("Film Setup", setupContent))
	
	// EI options
	var eiOptions []string
	if m.data.FilmSetup.SelectedFilm != nil {
		for i, ei := range m.data.FilmSetup.SelectedFilm.EIRatings {
			// Get development time for this EI
			timeStr := "--:--"
			if times, exists := m.data.FilmSetup.SelectedFilm.Times20C[m.data.FilmSetup.SelectedDilution]; exists {
				if devTime, exists := times[ei]; exists {
					timeStr = formatDuration(devTime)
				}
			}
			
			description := fmt.Sprintf("(%s @ %s)", timeStr, m.data.FilmSetup.SelectedDilution)
			eiOptions = append(eiOptions, view.RenderKeyBinding(strconv.Itoa(i+1), fmt.Sprintf("EI %d  %s", ei, description)))
		}
	}
	
	eiOptions = append(eiOptions, "")
	eiOptions = append(eiOptions, view.RenderKeyBinding("ESC", "Back to film selection"))
	
	eiContent := strings.Join(eiOptions, "\n")
	sections = append(sections, view.RenderSection("Select EI Rating", eiContent))
	
	// Actions
	actionsContent := view.RenderKeyBinding("1-5", "Select EI") + "    " +
		view.RenderKeyBinding("ESC", "Back") + "    " +
		view.RenderKeyBinding("Q", "Quit")
	sections = append(sections, view.RenderSection("Actions", actionsContent))
	
	return view.Styles.MainContainer.Render(strings.Join(sections, "\n"))
}

// renderRollSelection renders the roll selection screen
func (m *AppModel) renderRollSelection() string {
	var sections []string
	
	// Title
	sections = append(sections, view.RenderTitle())
	
	// Film Setup section
	filmName := "--"
	if m.data.FilmSetup.SelectedFilm != nil {
		filmName = m.data.FilmSetup.SelectedFilm.Name
	}
	
	eiValue := "--"
	if m.data.FilmSetup.SelectedEI > 0 {
		eiValue = strconv.Itoa(m.data.FilmSetup.SelectedEI)
	}
	
	setupContent := fmt.Sprintf("Film Type:    %s                        EI:  %s\nRolls:        %s                              Tank: %s",
		view.RenderValue(filmName),
		view.RenderValue(eiValue),
		view.RenderValue("--"),
		view.RenderValue("--ml"))
	
	sections = append(sections, view.RenderSection("Film Setup", setupContent))
	
	// Roll options
	rollsContent := "35mm Rolls:                           120mm Rolls:\n" +
		"[1] 1 Roll (300ml)  [4] 4 Rolls       [A] 1 Roll (500ml)  [D] 4 Rolls\n" +
		"[2] 2 Rolls (500ml) [5] 5 Rolls       [B] 2 Rolls (700ml) [E] 5 Rolls\n" +
		"[3] 3 Rolls (600ml) [6] 6 Rolls       [C] 3 Rolls (900ml) [F] 6 Rolls\n\n" +
		"Mixed batches: [M] Custom mix\n\n" +
		"[ESC] Back to EI selection"
	
	sections = append(sections, view.RenderSection("Number of Rolls", rollsContent))
	
	// Actions
	actionsContent := view.RenderKeyBinding("1-6", "35mm") + "    " +
		view.RenderKeyBinding("A-F", "120mm") + "    " +
		view.RenderKeyBinding("M", "Mixed") + "    " +
		view.RenderKeyBinding("ESC", "Back") + "    " +
		view.RenderKeyBinding("Q", "Quit")
	sections = append(sections, view.RenderSection("Actions", actionsContent))
	
	return view.Styles.MainContainer.Render(strings.Join(sections, "\n"))
}

// renderMixedRollInput renders the mixed roll input screen
func (m *AppModel) renderMixedRollInput() string {
	var sections []string
	
	// Title
	sections = append(sections, view.RenderTitle())
	
	// Custom Mix Setup
	mixContent := fmt.Sprintf("35mm Rolls: %s    (↑/↓ to adjust)\n120mm Rolls: %s   (+/- to adjust)\n\nTotal Tank Size: %s\n\n%s    %s    %s",
		view.RenderValue(strconv.Itoa(m.data.FilmSetup.Rolls35mm)),
		view.RenderValue(strconv.Itoa(m.data.FilmSetup.Rolls120mm)),
		view.RenderValue(fmt.Sprintf("%dml", m.data.FilmSetup.TankSize)),
		view.RenderKeyBinding("ENTER", "Confirm"),
		view.RenderKeyBinding("ESC", "Back"),
		view.RenderKeyBinding("R", "Reset"))
	
	sections = append(sections, view.RenderSection("Custom Mix Setup", mixContent))
	
	// Actions
	actionsContent := view.RenderKeyBinding("↑↓", "Adjust 35mm") + "    " +
		view.RenderKeyBinding("+/-", "Adjust 120mm") + "    " +
		view.RenderKeyBinding("ENTER", "Confirm") + "    " +
		view.RenderKeyBinding("ESC", "Back")
	sections = append(sections, view.RenderSection("Actions", actionsContent))
	
	return view.Styles.MainContainer.Render(strings.Join(sections, "\n"))
}

// renderCalculatedScreen renders the calculated results screen
func (m *AppModel) renderCalculatedScreen() string {
	var sections []string
	
	// Title
	sections = append(sections, view.RenderTitle())
	
	// Film Setup section
	filmName := "--"
	if m.data.FilmSetup.SelectedFilm != nil {
		filmName = m.data.FilmSetup.SelectedFilm.Name
	}
	
	eiValue := "--"
	if m.data.FilmSetup.SelectedEI > 0 {
		eiValue = strconv.Itoa(m.data.FilmSetup.SelectedEI)
	}
	
	tankValue := "--ml"
	if m.data.FilmSetup.TankSize > 0 {
		tankValue = fmt.Sprintf("%dml", m.data.FilmSetup.TankSize)
	}
	
	setupContent := fmt.Sprintf("Film Type:    %s                        EI:  %s\nRolls:        %s                        Tank: %s",
		view.RenderValue(filmName),
		view.RenderValue(eiValue),
		view.RenderValue(m.data.GetRollsDescription()),
		view.RenderValue(tankValue))
	
	sections = append(sections, view.RenderSection("Film Setup", setupContent))
	
	// Chemicals section
	chemicalsContent := m.renderChemicalsTable()
	sections = append(sections, view.RenderSection("Chemicals (20°C)", chemicalsContent))
	
	// Fixer Usage section
	currentRolls := m.data.GetCurrentRolls()
	afterUse := m.data.FixerState.RemainingRolls - currentRolls
	
	fixerContent := fmt.Sprintf("Capacity: %d rolls per liter    Used: %d rolls    Remaining: %d rolls\nThis batch uses: %d roll(s)         After use: %d rolls remaining",
		m.data.FixerState.TotalCapacity,
		m.data.FixerState.UsedRolls,
		m.data.FixerState.RemainingRolls,
		currentRolls,
		afterUse)
	sections = append(sections, view.RenderSection("Fixer Usage", fixerContent))
	
	// Actions section
	actionsContent := view.RenderKeyBinding("U", "Use Fixer") + "    " +
		view.RenderKeyBinding("R", "Change Rolls") + "    " +
		view.RenderKeyBinding("F", "Change Film") + "    " +
		view.RenderKeyBinding("Q", "Quit")
	sections = append(sections, view.RenderSection("Actions", actionsContent))
	
	return view.Styles.MainContainer.Render(strings.Join(sections, "\n"))
}
