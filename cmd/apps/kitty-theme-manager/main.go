package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Theme represents a kitty theme
type Theme struct {
	Name       string
	Path       string
	Background string
	Foreground string
	Cursor     string
	Colors     [16]string
	IsFavorite bool
}

// Implement list.Item interface for Theme
func (t Theme) FilterValue() string { return t.Name }
func (t Theme) Title() string       { 
	// Add star for favorites
	if t.IsFavorite {
		return "★ " + t.Name
	}
	return t.Name 
}
func (t Theme) Description() string { 
	// Show actual colors as colored blocks
	desc := ""
	if t.Background != "" {
		bgStyle := lipgloss.NewStyle().Background(lipgloss.Color(t.Background)).Foreground(lipgloss.Color(t.Background))
		desc += "BG:" + bgStyle.Render("███") + " "
	}
	if t.Foreground != "" {
		fgStyle := lipgloss.NewStyle().Background(lipgloss.Color(t.Foreground)).Foreground(lipgloss.Color(t.Foreground))
		desc += "FG:" + fgStyle.Render("███")
	}
	return desc
}

// Favorites storage
type Favorites struct {
	Themes []string `json:"themes"`
}

// Key bindings
type keyMap struct {
	Up            key.Binding
	Down          key.Binding
	Apply         key.Binding
	Preview       key.Binding
	Launch        key.Binding
	Rollback      key.Binding
	Favorite      key.Binding
	ToggleFavs    key.Binding
	Help          key.Binding
	Quit          key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Apply, k.Favorite, k.ToggleFavs, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Apply, k.Launch},
		{k.Preview, k.Rollback, k.Favorite, k.ToggleFavs},
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Apply: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "apply theme"),
	),
	Preview: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "preview theme"),
	),
	Launch: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("l", "launch kitty"),
	),
	Rollback: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "rollback"),
	),
	Favorite: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "toggle favorite"),
	),
	ToggleFavs: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "toggle favorites view"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// Model holds the application state
type Model struct {
	list           list.Model
	allThemes      []Theme
	currentTheme   Theme
	backupCount    int
	help           help.Model
	showHelp       bool
	showFavs       bool
	favorites      Favorites
	width          int
	height         int
}

// ThemeLoaded message type
type ThemeLoaded struct {
	themes []list.Item
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func initialModel() Model {
	// Create list with custom styling
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedStyle
	delegate.Styles.SelectedDesc = selectedDescStyle
	
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "F1: THEME LIST"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false) // Disable built-in filtering
	l.Styles.Title = headerStyle
	l.Styles.TitleBar = titleBarStyle
	
	return Model{
		list:        l,
		backupCount: getBackupCount(),
		help:        help.New(),
		showHelp:    false,
		showFavs:    false,
		favorites:   loadFavorites(),
	}
}

func (m Model) Init() tea.Cmd {
	return loadThemes
}

// Load themes from the kitty-themes directory
func loadThemes() tea.Msg {
	homeDir, _ := os.UserHomeDir()
	themesDir := filepath.Join(homeDir, ".config", "kitty", "kitty-themes", "themes")
	
	var themes []list.Item
	favorites := loadFavorites()
	
	err := filepath.Walk(themesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		if strings.HasSuffix(path, ".conf") {
			theme := parseTheme(path)
			if theme.Name != "" {
				// Check if this theme is a favorite
				for _, favName := range favorites.Themes {
					if theme.Name == favName {
						theme.IsFavorite = true
						break
					}
				}
				themes = append(themes, theme)
			}
		}
		return nil
	})
	
	if err != nil {
		return ThemeLoaded{themes: []list.Item{}}
	}
	
	// Sort themes: favorites first, then alphabetically within each group
	sort.Slice(themes, func(i, j int) bool {
		themeI := themes[i].(Theme)
		themeJ := themes[j].(Theme)
		
		if themeI.IsFavorite && !themeJ.IsFavorite {
			return true
		}
		if !themeI.IsFavorite && themeJ.IsFavorite {
			return false
		}
		return themeI.Name < themeJ.Name
	})
	
	return ThemeLoaded{themes: themes}
}

// Parse a theme file and extract color information
func parseTheme(path string) Theme {
	content, err := os.ReadFile(path)
	if err != nil {
		return Theme{}
	}
	
	theme := Theme{
		Name: strings.TrimSuffix(filepath.Base(path), ".conf"),
		Path: path,
	}
	
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		
		key := parts[0]
		value := parts[1]
		
		switch key {
		case "background":
			theme.Background = value
		case "foreground":
			theme.Foreground = value
		case "cursor":
			theme.Cursor = value
		case "color0", "color1", "color2", "color3", "color4", "color5", "color6", "color7",
			 "color8", "color9", "color10", "color11", "color12", "color13", "color14", "color15":
			if colorNum := strings.TrimPrefix(key, "color"); colorNum != "" {
				if num, err := strconv.Atoi(colorNum); err == nil && num < 16 {
					theme.Colors[num] = value
				}
			}
		}
	}
	
	return theme
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update list size - leave space for preview and help
		listHeight := (msg.Height - 4) / 2
		m.list.SetSize(msg.Width-4, listHeight)
		
	case ThemeLoaded:
		// Store all themes
		m.allThemes = make([]Theme, len(msg.themes))
		for i, item := range msg.themes {
			m.allThemes[i] = item.(Theme)
		}
		
		// Set initial list based on view mode
		m.updateList()
		
		if len(msg.themes) > 0 {
			m.currentTheme = msg.themes[0].(Theme)
		}
		
	case tea.KeyMsg:
		// Handle help toggle first
		if key.Matches(msg, keys.Help) {
			m.showHelp = !m.showHelp
			return m, nil
		}
		
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		
		if key.Matches(msg, keys.Apply) {
			// Apply theme in background
			go func() {
				cmd := exec.Command("sh", "-c", fmt.Sprintf(`
					cd ~/.config/kitty
					if [ -f current-theme.conf ]; then
						cp current-theme.conf current-theme.conf.%03d
					fi
					cp "%s" current-theme.conf
					kitty @ load-config
				`, m.backupCount+1, m.currentTheme.Path))
				cmd.Run()
			}()
			m.backupCount++
			return m, nil
		}
		
		if key.Matches(msg, keys.Preview) {
			return m, previewTheme(m.currentTheme)
		}
		
		if key.Matches(msg, keys.Launch) {
			return m, launchKittyWithTheme(m.currentTheme)
		}
		
		if key.Matches(msg, keys.Rollback) {
			return m, rollbackTheme()
		}
		
		if key.Matches(msg, keys.Favorite) {
			// Toggle favorite for current theme
			if selectedItem := m.list.SelectedItem(); selectedItem != nil {
				selectedTheme := selectedItem.(Theme)
				m.toggleFavorite(selectedTheme.Name)
				m.updateList()
			}
			return m, nil
		}
		
		if key.Matches(msg, keys.ToggleFavs) {
			m.showFavs = !m.showFavs
			m.updateList()
			return m, nil
		}
		
		// Update list and current theme
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		
		// Update current theme when selection changes
		if selectedItem := m.list.SelectedItem(); selectedItem != nil {
			m.currentTheme = selectedItem.(Theme)
		}
		
		return m, cmd
	}
	
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) updateList() {
	var items []list.Item
	
	for _, theme := range m.allThemes {
		if m.showFavs {
			// Only show favorites
			if theme.IsFavorite {
				items = append(items, theme)
			}
		} else {
			// Show all themes
			items = append(items, theme)
		}
	}
	
	// Update list title based on mode
	if m.showFavs {
		m.list.Title = "F1: FAVORITES (" + fmt.Sprintf("%d", len(items)) + ")"
	} else {
		m.list.Title = "F1: ALL THEMES (" + fmt.Sprintf("%d", len(items)) + ")"
	}
	
	m.list.SetItems(items)
}

func (m *Model) toggleFavorite(themeName string) {
	// Update in allThemes
	for i := range m.allThemes {
		if m.allThemes[i].Name == themeName {
			m.allThemes[i].IsFavorite = !m.allThemes[i].IsFavorite
			break
		}
	}
	
	// Update favorites struct
	found := false
	for i, favName := range m.favorites.Themes {
		if favName == themeName {
			// Remove from favorites
			m.favorites.Themes = append(m.favorites.Themes[:i], m.favorites.Themes[i+1:]...)
			found = true
			break
		}
	}
	
	if !found {
		// Add to favorites
		m.favorites.Themes = append(m.favorites.Themes, themeName)
	}
	
	// Save to file
	saveFavorites(m.favorites)
}

func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}
	
	// Main container
	mainStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(1)
	
	// Theme list section
	listSection := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4A90E2")).
		Padding(0, 1).
		Width(m.width/2 - 2).
		Height((m.height-8)/2)
	
	// Preview section
	previewSection := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4A90E2")).
		Padding(0, 1).
		Width(m.width/2 - 2).
		Height((m.height-8)/2)
	
	// Create the layout
	listView := listSection.Render(m.list.View())
	previewView := previewSection.Render(m.renderPreview())
	
	// Horizontal layout for list and preview
	topSection := lipgloss.JoinHorizontal(lipgloss.Top, listView, previewView)
	
	// Help section at bottom
	helpView := m.help.View(keys)
	helpSection := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(0, 1).
		Width(m.width - 4).
		Render(helpView)
	
	// Combine all sections
	content := lipgloss.JoinVertical(lipgloss.Left, topSection, helpSection)
	
	return mainStyle.Render(content)
}

func (m Model) renderPreview() string {
	if m.currentTheme.Name == "" {
		return "No theme selected"
	}
	
	// Title
	title := headerStyle.Render("F2: THEME PREVIEW")
	
	// Theme info
	info := fmt.Sprintf("Theme: %s\n", m.currentTheme.Name)
	if m.currentTheme.Background != "" {
		info += fmt.Sprintf("Background: %s\n", m.currentTheme.Background)
	}
	if m.currentTheme.Foreground != "" {
		info += fmt.Sprintf("Foreground: %s\n", m.currentTheme.Foreground)
	}
	if m.currentTheme.Cursor != "" {
		info += fmt.Sprintf("Cursor: %s\n", m.currentTheme.Cursor)
	}
	
	// Fake terminal preview
	terminalContent := m.renderFakeTerminal()
	
	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		info,
		"",
		terminalContent,
	)
}

func (m Model) renderFakeTerminal() string {
	bg := "#000000"
	fg := "#ffffff"
	if m.currentTheme.Background != "" {
		bg = m.currentTheme.Background
	}
	if m.currentTheme.Foreground != "" {
		fg = m.currentTheme.Foreground
	}
	
	// Terminal window style
	terminalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(fg)).
		Background(lipgloss.Color(bg)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#888888")).
		Padding(1).
		Width(40).
		Height(8)
	
	// Sample content with colors
	content := "$ ls -la\n"
	
	// Directory (color2 - usually green)
	if m.currentTheme.Colors[2] != "" {
		dirStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.currentTheme.Colors[2]))
		content += dirStyle.Render("drwxr-xr-x  docs/") + "\n"
	} else {
		content += "drwxr-xr-x  docs/\n"
	}
	
	// Executable (color1 - usually red)
	if m.currentTheme.Colors[1] != "" {
		exeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.currentTheme.Colors[1]))
		content += exeStyle.Render("-rwxr-xr-x  script.sh") + "\n"
	} else {
		content += "-rwxr-xr-x  script.sh\n"
	}
	
	content += "-rw-r--r--  config.yaml\n"
	content += "$ echo 'Hello World!'\n"
	content += "Hello World!\n"
	
	// Cursor
	if m.currentTheme.Cursor != "" {
		cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.currentTheme.Cursor))
		content += cursorStyle.Render("$ █")
	} else {
		content += "$ █"
	}
	
	return terminalStyle.Render(content)
}

func (m Model) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4A90E2")).
		Padding(2).
		Width(m.width - 4)
	
	title := headerStyle.Render("KITTY THEME MANAGER - HELP")
	content := m.help.View(keys)
	
	return helpStyle.Render(lipgloss.JoinVertical(lipgloss.Left, title, "", content))
}

// Styles
var (
	headerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Background(lipgloss.Color("#000080")).
		Bold(true).
		Padding(0, 1)
	
	titleBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#000080")).
		Foreground(lipgloss.Color("#FFFFFF"))
	
	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4A90E2")).
		Bold(true)
	
	selectedDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E0E0E0")).
		Background(lipgloss.Color("#4A90E2"))
)

// Commands
func previewTheme(theme Theme) tea.Cmd {
	return tea.ExecProcess(
		exec.Command("kitty", "@", "set-colors", "-a", theme.Path),
		func(err error) tea.Msg {
			return nil
		},
	)
}

func launchKittyWithTheme(theme Theme) tea.Cmd {
	return tea.ExecProcess(
		exec.Command("kitty", "-o", fmt.Sprintf("include=%s", theme.Path)),
		func(err error) tea.Msg {
			return nil
		},
	)
}

func rollbackTheme() tea.Cmd {
	return tea.ExecProcess(
		exec.Command("sh", "-c", `
			cd ~/.config/kitty
			LATEST=$(ls -1 current-theme.conf.[0-9][0-9][0-9] 2>/dev/null | sort -r | head -1)
			if [ -n "$LATEST" ]; then
				cp "$LATEST" current-theme.conf
				kitty @ load-config
			fi
		`),
		func(err error) tea.Msg {
			return nil
		},
	)
}

func getBackupCount() int {
	homeDir, _ := os.UserHomeDir()
	kittyDir := filepath.Join(homeDir, ".config", "kitty")
	
	files, err := filepath.Glob(filepath.Join(kittyDir, "current-theme.conf.[0-9][0-9][0-9]"))
	if err != nil {
		return 0
	}
	
	return len(files)
}

func getFavoritesPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "kitty", "favorites.json")
}

func loadFavorites() Favorites {
	favPath := getFavoritesPath()
	
	data, err := os.ReadFile(favPath)
	if err != nil {
		// File doesn't exist, return empty favorites
		return Favorites{Themes: []string{}}
	}
	
	var favorites Favorites
	err = json.Unmarshal(data, &favorites)
	if err != nil {
		// Invalid JSON, return empty favorites
		return Favorites{Themes: []string{}}
	}
	
	return favorites
}

func saveFavorites(favorites Favorites) {
	favPath := getFavoritesPath()
	
	data, err := json.MarshalIndent(favorites, "", "  ")
	if err != nil {
		return
	}
	
	os.WriteFile(favPath, data, 0644)
}
