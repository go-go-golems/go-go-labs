package tui

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/go-go-golems/go-go-labs/cmd/apps/poll-modem/internal/modem"
	"github.com/rs/zerolog/log"
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
type loginStartMsg struct{}
type loginResultMsg struct {
	info *modem.ModemInfo
	err  error
}
type csvExportMsg struct {
	success bool
	files   []string
}

// App represents the main TUI application
type App struct {
	client            *modem.Client
	modemInfo         *modem.ModemInfo
	lastError         error
	pollInterval      time.Duration
	
	// UI components
	spinner      spinner.Model
	downTable    table.Model
	upTable      table.Model
	errorTable   table.Model
	
	// History tables
	downHistoryTable  table.Model
	upHistoryTable    table.Model
	errorHistoryTable table.Model
	
	// History data
	history           []modem.ModemInfo
	maxHistoryEntries int
	
	// State
	width        int
	height       int
	loading      bool
	loggingIn    bool // New field to track login state
	currentView  int // 0: overview, 1: downstream, 2: upstream, 3: errors
	showHistory  bool // Toggle between current and history view
	lastExport   string // Status of last CSV export
	selectedChannelID string // Track selected channel for history filtering
	
	// Key bindings
	keys keyMap
}

type keyMap struct {
	Quit        key.Binding
	Refresh     key.Binding
	NextView    key.Binding
	PrevView    key.Binding
	Help        key.Binding
	ToggleHistory key.Binding
	ExportCSV   key.Binding
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
	ToggleHistory: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "toggle history"),
	),
	ExportCSV: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "export CSV"),
	),
}

// NewApp creates a new TUI application
func NewApp(baseURL string, pollInterval time.Duration, username, password string) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	log.Info().Msg("Creating new app")

	client := modem.NewClient(baseURL)
	if username != "" && password != "" {
		client.SetCredentials(username, password)
	}

	app := &App{
		client:            client,
		pollInterval:      pollInterval,
		spinner:           s,
		keys:              defaultKeys,
		loading:           true,
		maxHistoryEntries: 100, // Keep last 100 readings
		history:           make([]modem.ModemInfo, 0),
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

	// History tables with timestamp column
	downHistoryColumns := []table.Column{
		{Title: "Time", Width: 10},
		{Title: "ID", Width: 4},
		{Title: "Status", Width: 8},
		{Title: "Frequency", Width: 12},
		{Title: "SNR", Width: 8},
		{Title: "Power", Width: 10},
		{Title: "Modulation", Width: 12},
	}
	a.downHistoryTable = table.New(
		table.WithColumns(downHistoryColumns),
		table.WithFocused(true),
	)
	a.downHistoryTable.SetStyles(tableStyles)

	upHistoryColumns := []table.Column{
		{Title: "Time", Width: 10},
		{Title: "ID", Width: 4},
		{Title: "Status", Width: 8},
		{Title: "Frequency", Width: 12},
		{Title: "Symbol Rate", Width: 12},
		{Title: "Power", Width: 10},
		{Title: "Modulation", Width: 12},
		{Title: "Type", Width: 8},
	}
	a.upHistoryTable = table.New(
		table.WithColumns(upHistoryColumns),
		table.WithFocused(true),
	)
	a.upHistoryTable.SetStyles(tableStyles)

	errorHistoryColumns := []table.Column{
		{Title: "Time", Width: 10},
		{Title: "ID", Width: 4},
		{Title: "Unerrored", Width: 12},
		{Title: "Correctable", Width: 12},
		{Title: "Uncorrectable", Width: 14},
	}
	a.errorHistoryTable = table.New(
		table.WithColumns(errorHistoryColumns),
		table.WithFocused(true),
	)
	a.errorHistoryTable.SetStyles(tableStyles)
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
			a.loggingIn = false
			return a, a.fetchData()
		case key.Matches(msg, a.keys.NextView):
			a.currentView = (a.currentView + 1) % 4
			return a, nil
		case key.Matches(msg, a.keys.PrevView):
			a.currentView = (a.currentView - 1 + 4) % 4
			return a, nil
		case key.Matches(msg, a.keys.ToggleHistory):
			if !a.showHistory {
				// Capture current selection before switching to history
				a.updateSelectedChannel()
			}
			a.showHistory = !a.showHistory
			// Refresh history tables to show filtered data
			a.updateHistoryTables()
			return a, nil
		case key.Matches(msg, a.keys.ExportCSV):
			return a, a.exportCSV()
		default:
			// Forward navigation keys to the active table
			if a.modemInfo != nil {
				switch a.currentView {
				case 1: // Downstream
					if a.showHistory {
						// Only history table receives events in history mode
						var cmd tea.Cmd
						a.downHistoryTable, cmd = a.downHistoryTable.Update(msg)
						return a, cmd
					} else {
						// Normal mode - update current table and track selection
						var cmd tea.Cmd
						a.downTable, cmd = a.downTable.Update(msg)
						a.updateSelectedChannel()
						return a, cmd
					}
				case 2: // Upstream
					if a.showHistory {
						// Only history table receives events in history mode
						var cmd tea.Cmd
						a.upHistoryTable, cmd = a.upHistoryTable.Update(msg)
						return a, cmd
					} else {
						// Normal mode - update current table and track selection
						var cmd tea.Cmd
						a.upTable, cmd = a.upTable.Update(msg)
						a.updateSelectedChannel()
						return a, cmd
					}
				case 3: // Errors
					if a.showHistory {
						// Only history table receives events in history mode
						var cmd tea.Cmd
						a.errorHistoryTable, cmd = a.errorHistoryTable.Update(msg)
						return a, cmd
					} else {
						// Normal mode - update current table and track selection
						var cmd tea.Cmd
						a.errorTable, cmd = a.errorTable.Update(msg)
						a.updateSelectedChannel()
						return a, cmd
					}
				}
			}
		}

	case tickMsg:
		if !a.loading && !a.loggingIn {
			a.loading = true
			return a, tea.Batch(a.fetchData(), a.tickCmd())
		}
		return a, a.tickCmd()

	case fetchResultMsg:
		a.loading = false
		if msg.err != nil {
			// Check if it's a logout error
			if modem.IsLogoutError(msg.err) {
				// Start login process if credentials are available
				if a.client != nil {
					// Check if we have credentials before attempting login
					// This is a simple check - the actual validation happens in LoginAndFetch
					a.loggingIn = true
					return a, a.performLogin()
				} else {
					a.lastError = errors.New("authentication required but no credentials provided - use --username and --password flags")
				}
			} else {
				a.lastError = msg.err
			}
		} else {
			a.lastError = nil
			a.modemInfo = msg.info
			a.addToHistory(*msg.info)
			a.updateTables()
			a.updateHistoryTables()
		}
		return a, nil

	case loginStartMsg:
		a.loggingIn = true
		return a, nil

	case loginResultMsg:
		a.loggingIn = false
		if msg.err != nil {
			a.lastError = msg.err
		} else {
			a.lastError = nil
			a.modemInfo = msg.info
			a.addToHistory(*msg.info)
			a.updateTables()
			a.updateHistoryTables()
		}
		return a, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd
		
	case csvExportMsg:
		if msg.success {
			a.lastExport = "CSV files exported successfully"
		} else {
			a.lastExport = "CSV export failed"
		}
		return a, nil
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
	} else if a.loggingIn {
		content.WriteString(fmt.Sprintf("%s Logging in...", a.spinner.View()))
	} else if a.lastError != nil {
		errorMsg := fmt.Sprintf("Error: %v", a.lastError)
		// Add helpful hints for common errors
		if modem.IsLogoutError(a.lastError) {
			errorMsg += "\nHint: Session expired - authentication will be attempted automatically if credentials are provided"
		} else if strings.Contains(a.lastError.Error(), "credentials provided") {
			errorMsg += "\nHint: Use --username and --password flags when starting the application"
		} else if strings.Contains(a.lastError.Error(), "forbidden") || strings.Contains(a.lastError.Error(), "403") {
			errorMsg += "\nHint: Try providing --username and --password flags for authentication"
		} else if strings.Contains(a.lastError.Error(), "connection refused") {
			errorMsg += "\nHint: Check if the modem URL is correct and accessible"
		} else if strings.Contains(a.lastError.Error(), "timeout") {
			errorMsg += "\nHint: The modem may be slow to respond, try increasing the timeout"
		} else if strings.Contains(a.lastError.Error(), "authentication failed") {
			errorMsg += "\nHint: Check if the username and password are correct"
		}
		content.WriteString(errorStyle.Render(errorMsg))
	} else if a.modemInfo != nil {
		statusText := fmt.Sprintf("Last updated: %s", a.modemInfo.LastUpdated.Format("15:04:05"))
		if a.lastExport != "" {
			statusText += fmt.Sprintf(" | %s", a.lastExport)
		}
		content.WriteString(successStyle.Render(statusText))
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
		content.WriteString(infoStyle.Render("Press 'r' to refresh, 'tab' to switch views, 'h' to toggle history, 'e' to export CSV, 'q' to quit"))
	} else {
		if a.showHistory {
			content.WriteString(infoStyle.Render("Press 'r' to refresh, 'tab' to switch views, '↑/↓' or 'j/k' to navigate history, 'h' to return to current data, 'e' to export CSV, 'q' to quit"))
		} else {
			content.WriteString(infoStyle.Render("Press 'r' to refresh, 'tab' to switch views, '↑/↓' or 'j/k' to navigate table, 'h' to view history, 'e' to export CSV, 'q' to quit"))
		}
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
	
	if a.showHistory && len(a.history) > 0 {
		// Show history view with descriptive header
		if a.selectedChannelID != "" {
			content.WriteString(headerStyle.Render(fmt.Sprintf("Downstream History - Channel %s", a.selectedChannelID)))
		} else {
			content.WriteString(headerStyle.Render("Downstream History - All Channels"))
		}
		content.WriteString("\n\n")
		content.WriteString(tableStyle.Render(a.downHistoryTable.View()))
	} else {
		// Show current data view
		content.WriteString(headerStyle.Render("Downstream Channels"))
		content.WriteString("\n\n")
		content.WriteString(tableStyle.Render(a.downTable.View()))
	}
	
	return content.String()
}

func (a *App) renderUpstream() string {
	var content strings.Builder
	
	if a.showHistory && len(a.history) > 0 {
		// Show history view with descriptive header
		if a.selectedChannelID != "" {
			content.WriteString(headerStyle.Render(fmt.Sprintf("Upstream History - Channel %s", a.selectedChannelID)))
		} else {
			content.WriteString(headerStyle.Render("Upstream History - All Channels"))
		}
		content.WriteString("\n\n")
		content.WriteString(tableStyle.Render(a.upHistoryTable.View()))
	} else {
		// Show current data view
		content.WriteString(headerStyle.Render("Upstream Channels"))
		content.WriteString("\n\n")
		content.WriteString(tableStyle.Render(a.upTable.View()))
	}
	
	return content.String()
}

func (a *App) renderErrors() string {
	var content strings.Builder
	
	if a.showHistory && len(a.history) > 0 {
		// Show history view with descriptive header
		if a.selectedChannelID != "" {
			content.WriteString(headerStyle.Render(fmt.Sprintf("Error History - Channel %s", a.selectedChannelID)))
		} else {
			content.WriteString(headerStyle.Render("Error History - All Channels"))
		}
		content.WriteString("\n\n")
		content.WriteString(tableStyle.Render(a.errorHistoryTable.View()))
	} else {
		// Show current data view
		content.WriteString(headerStyle.Render("Error Codewords"))
		content.WriteString("\n\n")
		content.WriteString(tableStyle.Render(a.errorTable.View()))
	}
	
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
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		info, err := a.client.FetchModemInfo(ctx)
		return fetchResultMsg{info: info, err: err}
	}
}

func (a *App) performLogin() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		info, err := a.client.LoginAndFetch(ctx)
		return loginResultMsg{info: info, err: err}
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
	
	// History tables get the same full height since they're shown exclusively
	a.downHistoryTable.SetHeight(availableHeight)
	a.upHistoryTable.SetHeight(availableHeight)
	a.errorHistoryTable.SetHeight(availableHeight)
}

func (a *App) exportCSV() tea.Cmd {
	return func() tea.Msg {
		if len(a.history) == 0 {
			return csvExportMsg{success: false, files: []string{}}
		}

		timestamp := time.Now().Format("2006-01-02_15-04-05")
		var files []string
		
		// Export downstream data
		downFile := fmt.Sprintf("modem_downstream_%s.csv", timestamp)
		if err := a.exportDownstreamCSV(timestamp); err != nil {
			return fetchResultMsg{err: err}
		}
		files = append(files, downFile)
		
		// Export upstream data
		upFile := fmt.Sprintf("modem_upstream_%s.csv", timestamp)
		if err := a.exportUpstreamCSV(timestamp); err != nil {
			return fetchResultMsg{err: err}
		}
		files = append(files, upFile)
		
		// Export error data
		errorFile := fmt.Sprintf("modem_errors_%s.csv", timestamp)
		if err := a.exportErrorCSV(timestamp); err != nil {
			return fetchResultMsg{err: err}
		}
		files = append(files, errorFile)
		
		return csvExportMsg{success: true, files: files}
	}
}

func (a *App) exportDownstreamCSV(timestamp string) error {
	filename := fmt.Sprintf("modem_downstream_%s.csv", timestamp)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Timestamp", "Channel_ID", "Lock_Status", "Frequency", "SNR", "Power_Level", "Modulation"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, info := range a.history {
		for _, ch := range info.Downstream {
			record := []string{
				info.LastUpdated.Format("2006-01-02 15:04:05"),
				ch.ChannelID,
				ch.LockStatus,
				ch.Frequency,
				ch.SNR,
				ch.PowerLevel,
				ch.Modulation,
			}
			if err := writer.Write(record); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *App) exportUpstreamCSV(timestamp string) error {
	filename := fmt.Sprintf("modem_upstream_%s.csv", timestamp)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Timestamp", "Channel_ID", "Lock_Status", "Frequency", "Symbol_Rate", "Power_Level", "Modulation", "Channel_Type"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, info := range a.history {
		for _, ch := range info.Upstream {
			record := []string{
				info.LastUpdated.Format("2006-01-02 15:04:05"),
				ch.ChannelID,
				ch.LockStatus,
				ch.Frequency,
				ch.SymbolRate,
				ch.PowerLevel,
				ch.Modulation,
				ch.ChannelType,
			}
			if err := writer.Write(record); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *App) exportErrorCSV(timestamp string) error {
	filename := fmt.Sprintf("modem_errors_%s.csv", timestamp)
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"Timestamp", "Channel_ID", "Unerrored_Codewords", "Correctable_Codewords", "Uncorrectable_Codewords"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, info := range a.history {
		for _, ch := range info.ErrorCodewords {
			record := []string{
				info.LastUpdated.Format("2006-01-02 15:04:05"),
				ch.ChannelID,
				ch.UnerroredCodewords,
				ch.CorrectableCodewords,
				ch.UncorrectableCodewords,
			}
			if err := writer.Write(record); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *App) addToHistory(info modem.ModemInfo) {
	if len(a.history) >= a.maxHistoryEntries {
		a.history = a.history[1:]
	}
	a.history = append(a.history, info)
}

func (a *App) updateHistoryTables() {
	if len(a.history) == 0 {
		return
	}

	// If we have a selected channel and we're showing history, filter by that channel
	// Otherwise show all channels (for when history mode is first enabled)
	
	// Update downstream history table
	var downHistoryRows []table.Row
	for _, info := range a.history {
		for _, ch := range info.Downstream {
			// Only add rows for the selected channel if we have one selected
			if a.selectedChannelID == "" || ch.ChannelID == a.selectedChannelID {
				downHistoryRows = append(downHistoryRows, table.Row{
					info.LastUpdated.Format("15:04:05"),
					ch.ChannelID,
					ch.LockStatus,
					ch.Frequency,
					ch.SNR,
					ch.PowerLevel,
					ch.Modulation,
				})
			}
		}
	}
	a.downHistoryTable.SetRows(downHistoryRows)

	// Update upstream history table
	var upHistoryRows []table.Row
	for _, info := range a.history {
		for _, ch := range info.Upstream {
			// Only add rows for the selected channel if we have one selected
			if a.selectedChannelID == "" || ch.ChannelID == a.selectedChannelID {
				upHistoryRows = append(upHistoryRows, table.Row{
					info.LastUpdated.Format("15:04:05"),
					ch.ChannelID,
					ch.LockStatus,
					ch.Frequency,
					ch.SymbolRate,
					ch.PowerLevel,
					ch.Modulation,
					ch.ChannelType,
				})
			}
		}
	}
	a.upHistoryTable.SetRows(upHistoryRows)

	// Update error history table
	var errorHistoryRows []table.Row
	for _, info := range a.history {
		for _, ch := range info.ErrorCodewords {
			// Only add rows for the selected channel if we have one selected
			if a.selectedChannelID == "" || ch.ChannelID == a.selectedChannelID {
				errorHistoryRows = append(errorHistoryRows, table.Row{
					info.LastUpdated.Format("15:04:05"),
					ch.ChannelID,
					ch.UnerroredCodewords,
					ch.CorrectableCodewords,
					ch.UncorrectableCodewords,
				})
			}
		}
	}
	a.errorHistoryTable.SetRows(errorHistoryRows)
}

func (a *App) updateSelectedChannel() {
	if a.modemInfo != nil {
		switch a.currentView {
		case 1:
			if row := a.downTable.SelectedRow(); len(row) > 0 {
				a.selectedChannelID = row[0]
			}
		case 2:
			if row := a.upTable.SelectedRow(); len(row) > 0 {
				a.selectedChannelID = row[0]
			}
		case 3:
			if row := a.errorTable.SelectedRow(); len(row) > 0 {
				a.selectedChannelID = row[0]
			}
		}
		// Update history tables to reflect the new selection
		a.updateHistoryTables()
	}
} 