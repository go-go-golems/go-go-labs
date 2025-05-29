package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/go-go-golems/go-go-labs/cmd/apps/poll-modem/internal/modem"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#F25D94")).
			Padding(0, 1).
			MarginTop(1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00"))

	tableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#874BFD"))
)

type tickMsg time.Time
type fetchResultMsg struct {
	info *modem.ModemInfo
	err  error
}

// App represents the main TUI application
type App struct {
	client       *modem.Client
	modemInfo    *modem.ModemInfo
	lastError    error
	pollInterval time.Duration
	
	// UI components
	spinner      spinner.Model
	downTable    table.Model
	upTable      table.Model
	errorTable   table.Model
	
	// State
	width        int
	height       int
	loading      bool
	currentView  int // 0: overview, 1: downstream, 2: upstream, 3: errors
	
	// Key bindings
	keys keyMap
}

type keyMap struct {
	Quit     key.Binding
	Refresh  key.Binding
	NextView key.Binding
	PrevView key.Binding
	Help     key.Binding
}

var defaultKeys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	NextView: key.NewBinding(
		key.WithKeys("tab", "right"),
		key.WithHelp("tab/→", "next view"),
	),
	PrevView: key.NewBinding(
		key.WithKeys("shift+tab", "left"),
		key.WithHelp("shift+tab/←", "prev view"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

// NewApp creates a new TUI application
func NewApp(url string, pollInterval time.Duration) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	app := &App{
		client:       modem.NewClient(url),
		pollInterval: pollInterval,
		spinner:      s,
		keys:         defaultKeys,
		loading:      true,
	}

	app.initTables()
	return app
}

func (a *App) initTables() {
	// Define common table styles
	tableStyles := table.Styles{
		Header: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false),
		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false),
	}

	// Downstream table - make it taller and scrollable for many channels
	downColumns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Status", Width: 8},
		{Title: "Frequency", Width: 12},
		{Title: "SNR", Width: 8},
		{Title: "Power", Width: 10},
		{Title: "Modulation", Width: 12},
	}
	a.downTable = table.New(
		table.WithColumns(downColumns),
		table.WithFocused(true),
	)
	a.downTable.SetStyles(tableStyles)

	// Upstream table
	upColumns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Status", Width: 8},
		{Title: "Frequency", Width: 12},
		{Title: "Symbol Rate", Width: 12},
		{Title: "Power", Width: 10},
		{Title: "Modulation", Width: 12},
		{Title: "Type", Width: 8},
	}
	a.upTable = table.New(
		table.WithColumns(upColumns),
		table.WithFocused(true),
	)
	a.upTable.SetStyles(tableStyles)

	// Error table - make it scrollable for many channels
	errorColumns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Unerrored", Width: 12},
		{Title: "Correctable", Width: 12},
		{Title: "Uncorrectable", Width: 14},
	}
	a.errorTable = table.New(
		table.WithColumns(errorColumns),
		table.WithFocused(true),
	)
	a.errorTable.SetStyles(tableStyles)
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.spinner.Tick,
		a.fetchData(),
		a.tickCmd(),
	)
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateTableHeights()
		return a, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, a.keys.Quit):
			return a, tea.Quit
		case key.Matches(msg, a.keys.Refresh):
			a.loading = true
			return a, a.fetchData()
		case key.Matches(msg, a.keys.NextView):
			a.currentView = (a.currentView + 1) % 4
			return a, nil
		case key.Matches(msg, a.keys.PrevView):
			a.currentView = (a.currentView - 1 + 4) % 4
			return a, nil
		default:
			// Forward navigation keys to the active table
			if a.modemInfo != nil {
				switch a.currentView {
				case 1: // Downstream
					var cmd tea.Cmd
					a.downTable, cmd = a.downTable.Update(msg)
					return a, cmd
				case 2: // Upstream
					var cmd tea.Cmd
					a.upTable, cmd = a.upTable.Update(msg)
					return a, cmd
				case 3: // Errors
					var cmd tea.Cmd
					a.errorTable, cmd = a.errorTable.Update(msg)
					return a, cmd
				}
			}
		}

	case tickMsg:
		if !a.loading {
			a.loading = true
			return a, tea.Batch(a.fetchData(), a.tickCmd())
		}
		return a, a.tickCmd()

	case fetchResultMsg:
		a.loading = false
		if msg.err != nil {
			a.lastError = msg.err
		} else {
			a.lastError = nil
			a.modemInfo = msg.info
			a.updateTables()
		}
		return a, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd
	}

	return a, nil
}

// View implements tea.Model
func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("Cable Modem Monitor"))
	content.WriteString("\n\n")

	// Status line
	if a.loading {
		content.WriteString(fmt.Sprintf("%s Fetching data...", a.spinner.View()))
	} else if a.lastError != nil {
		content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", a.lastError)))
	} else if a.modemInfo != nil {
		content.WriteString(successStyle.Render(fmt.Sprintf("Last updated: %s", a.modemInfo.LastUpdated.Format("15:04:05"))))
	}
	content.WriteString("\n\n")

	if a.modemInfo != nil {
		switch a.currentView {
		case 0:
			content.WriteString(a.renderOverview())
		case 1:
			content.WriteString(a.renderDownstream())
		case 2:
			content.WriteString(a.renderUpstream())
		case 3:
			content.WriteString(a.renderErrors())
		}
	}

	// Help
	content.WriteString("\n")
	if a.currentView == 0 {
		content.WriteString(infoStyle.Render("Press 'r' to refresh, 'tab' to switch views, 'q' to quit"))
	} else {
		content.WriteString(infoStyle.Render("Press 'r' to refresh, 'tab' to switch views, '↑/↓' or 'j/k' to navigate table, 'q' to quit"))
	}

	return content.String()
}

func (a *App) renderOverview() string {
	if a.modemInfo == nil {
		return "No data available"
	}

	var content strings.Builder

	// Cable modem info
	content.WriteString(headerStyle.Render("Cable Modem Information"))
	content.WriteString("\n\n")

	modem := a.modemInfo.CableModem
	content.WriteString(fmt.Sprintf("Model: %s (%s)\n", modem.Model, modem.ProductType))
	content.WriteString(fmt.Sprintf("Vendor: %s\n", modem.Vendor))
	content.WriteString(fmt.Sprintf("HW Version: %s\n", modem.HWVersion))
	content.WriteString(fmt.Sprintf("Core Version: %s\n", modem.CoreVersion))
	content.WriteString(fmt.Sprintf("BOOT Version: %s\n", modem.BOOTVersion))
	content.WriteString(fmt.Sprintf("Download Version: %s\n", modem.DownloadVersion))
	content.WriteString(fmt.Sprintf("Flash Part: %s\n", modem.FlashPart))

	// Summary stats
	content.WriteString("\n")
	content.WriteString(headerStyle.Render("Channel Summary"))
	content.WriteString("\n\n")

	downLocked := 0
	for _, ch := range a.modemInfo.Downstream {
		if ch.LockStatus == "Locked" {
			downLocked++
		}
	}

	upLocked := 0
	for _, ch := range a.modemInfo.Upstream {
		if ch.LockStatus == "Locked" {
			upLocked++
		}
	}

	content.WriteString(fmt.Sprintf("Downstream: %d/%d channels locked\n", downLocked, len(a.modemInfo.Downstream)))
	content.WriteString(fmt.Sprintf("Upstream: %d/%d channels locked\n", upLocked, len(a.modemInfo.Upstream)))

	return content.String()
}

func (a *App) renderDownstream() string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("Downstream Channels"))
	content.WriteString("\n\n")
	content.WriteString(tableStyle.Render(a.downTable.View()))
	return content.String()
}

func (a *App) renderUpstream() string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("Upstream Channels"))
	content.WriteString("\n\n")
	content.WriteString(tableStyle.Render(a.upTable.View()))
	return content.String()
}

func (a *App) renderErrors() string {
	var content strings.Builder
	content.WriteString(headerStyle.Render("Error Codewords"))
	content.WriteString("\n\n")
	content.WriteString(tableStyle.Render(a.errorTable.View()))
	return content.String()
}

func (a *App) updateTables() {
	if a.modemInfo == nil {
		return
	}

	// Update downstream table
	var downRows []table.Row
	for _, ch := range a.modemInfo.Downstream {
		downRows = append(downRows, table.Row{
			ch.ChannelID,
			ch.LockStatus,
			ch.Frequency,
			ch.SNR,
			ch.PowerLevel,
			ch.Modulation,
		})
	}
	a.downTable.SetRows(downRows)

	// Update upstream table
	var upRows []table.Row
	for _, ch := range a.modemInfo.Upstream {
		upRows = append(upRows, table.Row{
			ch.ChannelID,
			ch.LockStatus,
			ch.Frequency,
			ch.SymbolRate,
			ch.PowerLevel,
			ch.Modulation,
			ch.ChannelType,
		})
	}
	a.upTable.SetRows(upRows)

	// Update error table
	var errorRows []table.Row
	for _, ch := range a.modemInfo.ErrorCodewords {
		errorRows = append(errorRows, table.Row{
			ch.ChannelID,
			ch.UnerroredCodewords,
			ch.CorrectableCodewords,
			ch.UncorrectableCodewords,
		})
	}
	a.errorTable.SetRows(errorRows)
}

func (a *App) fetchData() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		info, err := a.client.FetchModemInfo(ctx)
		return fetchResultMsg{info: info, err: err}
	}
}

func (a *App) tickCmd() tea.Cmd {
	return tea.Tick(a.pollInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (a *App) updateTableHeights() {
	if a.height == 0 {
		return
	}

	// Calculate space used by fixed UI elements:
	// - Title: 3 lines (title + 2 newlines)
	// - Status line: 2 lines (status + newline)
	// - Header for current view: 3 lines (header + 2 newlines)
	// - Help text: 2 lines (help + newline)
	// - Margins and spacing: 2 lines
	fixedHeight := 12

	// Calculate available height for the table
	availableHeight := a.height - fixedHeight
	
	// Ensure minimum height of 3 for usability
	if availableHeight < 3 {
		availableHeight = 3
	}

	// Update all table heights to use the available space
	a.downTable.SetHeight(availableHeight)
	a.upTable.SetHeight(availableHeight)
	a.errorTable.SetHeight(availableHeight)
} 