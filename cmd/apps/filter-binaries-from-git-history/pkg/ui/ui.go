package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/cmd/apps/filter-binaries-from-git-history/pkg/analyzer"
	"github.com/go-go-golems/go-go-labs/cmd/apps/filter-binaries-from-git-history/pkg/git"
	"github.com/rs/zerolog/log"
)

// State represents the current state of the TUI
type State int

const (
	StateLoading State = iota
	StateStats
	StateFileSelection
	StateConfirmation
	StateProcessing
	StateDone
	StateError
)

// Model represents the TUI model
type Model struct {
	state         State
	baseRef       string
	compareRef    string
	sizeThreshold int64
	repo          *git.Repository
	stats         *analyzer.Stats
	table         table.Model
	selected      map[int]bool
	cursor        int
	err           error
	message       string
	processing    bool
}

// fileItem represents a file in the selection list
type fileItem struct {
	file     analyzer.FileInfo
	selected bool
}

func (i fileItem) FilterValue() string { return i.file.Path }

// StartTUI initializes and starts the TUI
func StartTUI(baseRef, compareRef string, sizeThreshold int64) error {
	repo, err := git.OpenRepository("")
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	m := Model{
		state:         StateLoading,
		baseRef:       baseRef,
		compareRef:    compareRef,
		sizeThreshold: sizeThreshold,
		repo:          repo,
		selected:      make(map[int]bool),
	}

	p := tea.NewProgram(&m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}

	return nil
}

func (m *Model) Init() tea.Cmd {
	return m.loadStats
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case statsMsg:
		return m.handleStatsMsg(msg)
	case errorMsg:
		return m.handleErrorMsg(msg)
	case processCompleteMsg:
		return m.handleProcessCompleteMsg(msg)
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "enter":
		switch m.state {
		case StateStats:
			return m.transitionToFileSelection(), nil
		case StateFileSelection:
			return m.transitionToConfirmation(), nil
		case StateConfirmation:
			return m.transitionToProcessing(), m.processRemoval
		case StateDone, StateError:
			return m, tea.Quit
		}
	case " ":
		if m.state == StateFileSelection {
			m.toggleSelection()
		}
	case "up", "k":
		if m.state == StateFileSelection && m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.state == StateFileSelection && m.cursor < len(m.stats.Files)-1 {
			m.cursor++
		}
	case "a":
		if m.state == StateFileSelection {
			m.selectAll(true)
		}
	case "n":
		if m.state == StateFileSelection {
			m.selectAll(false)
		}
	}

	return m, nil
}

func (m *Model) handleStatsMsg(msg statsMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.state = StateError
		m.err = msg.err
		return m, nil
	}

	m.stats = msg.stats
	m.state = StateStats
	return m, nil
}

func (m *Model) handleErrorMsg(msg errorMsg) (tea.Model, tea.Cmd) {
	m.state = StateError
	m.err = msg.err
	return m, nil
}

func (m *Model) handleProcessCompleteMsg(msg processCompleteMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.state = StateError
		m.err = msg.err
		return m, nil
	}

	m.state = StateDone
	m.message = msg.message
	return m, nil
}

func (m *Model) View() string {
	switch m.state {
	case StateLoading:
		return m.loadingView()
	case StateStats:
		return m.statsView()
	case StateFileSelection:
		return m.fileSelectionView()
	case StateConfirmation:
		return m.confirmationView()
	case StateProcessing:
		return m.processingView()
	case StateDone:
		return m.doneView()
	case StateError:
		return m.errorView()
	default:
		return "Unknown state"
	}
}

func (m *Model) loadingView() string {
	return fmt.Sprintf("🔍 Analyzing git history between %s and %s...\n\nPlease wait...", 
		m.baseRef, m.compareRef)
}

func (m *Model) statsView() string {
	if m.stats == nil {
		return "No statistics available"
	}

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("📊 Git History Analysis Results"))
	sb.WriteString("\n\n")
	sb.WriteString(m.stats.Summary(m.sizeThreshold))
	sb.WriteString("\n\n")

	if len(m.stats.Files) > 0 {
		sb.WriteString("🔍 Top 10 Largest Files:\n")
		count := 10
		if len(m.stats.Files) < count {
			count = len(m.stats.Files)
		}

		for i := 0; i < count; i++ {
			file := m.stats.Files[i]
			binary := ""
			if analyzer.IsLikelyBinary(file.Path) {
				binary = " 📦"
			}
			sb.WriteString(fmt.Sprintf("  %d. %s (%s)%s\n", 
				i+1, file.Path, analyzer.FormatSize(file.Size), binary))
		}
	}

	sb.WriteString("\n\n")
	sb.WriteString(lipgloss.NewStyle().Faint(true).Render("Press Enter to select files for removal, or 'q' to quit"))

	return sb.String()
}

func (m *Model) fileSelectionView() string {
	if m.stats == nil || len(m.stats.Files) == 0 {
		return "No files to display"
	}

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("📁 Select Files to Remove from History"))
	sb.WriteString("\n\n")

	largeFiles := m.stats.GetLargeFiles(m.sizeThreshold)
	if len(largeFiles) == 0 {
		sb.WriteString("No large files found above threshold")
		return sb.String()
	}

	selectedCount := 0
	for i := range m.selected {
		if m.selected[i] {
			selectedCount++
		}
	}

	sb.WriteString(fmt.Sprintf("Selected: %d/%d files\n\n", selectedCount, len(largeFiles)))

	for i, file := range largeFiles {
		prefix := "  "
		if i == m.cursor {
			prefix = "→ "
		}

		checkbox := "☐"
		if m.selected[i] {
			checkbox = "☑"
		}

		binary := ""
		if analyzer.IsLikelyBinary(file.Path) {
			binary = " 📦"
		}

		line := fmt.Sprintf("%s%s %s (%s)%s", 
			prefix, checkbox, file.Path, analyzer.FormatSize(file.Size), binary)

		if i == m.cursor {
			line = lipgloss.NewStyle().Background(lipgloss.Color("240")).Render(line)
		}

		sb.WriteString(line + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Faint(true).Render(
		"Navigation: ↑/↓ or j/k • Toggle: Space • Select All: a • Deselect All: n • Continue: Enter • Quit: q"))

	return sb.String()
}

func (m *Model) confirmationView() string {
	selectedFiles := m.getSelectedFiles()
	if len(selectedFiles) == 0 {
		return "❌ No files selected for removal.\n\nPress Enter to go back to file selection."
	}

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("⚠️  CONFIRM HISTORY REWRITE"))
	sb.WriteString("\n\n")
	sb.WriteString("The following files will be PERMANENTLY removed from git history:\n\n")

	totalSize := int64(0)
	for _, file := range selectedFiles {
		totalSize += file.Size
		sb.WriteString(fmt.Sprintf("  • %s (%s)\n", file.Path, analyzer.FormatSize(file.Size)))
	}

	sb.WriteString(fmt.Sprintf("\nTotal size to be removed: %s\n", analyzer.FormatSize(totalSize)))
	sb.WriteString(fmt.Sprintf("Base reference: %s\n", m.baseRef))
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("⚠️  WARNING: This operation cannot be undone!"))
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("⚠️  Make sure you have a backup before proceeding!"))
	sb.WriteString("\n\n")
	sb.WriteString(lipgloss.NewStyle().Faint(true).Render("Press Enter to proceed with history rewrite, or 'q' to quit"))

	return sb.String()
}

func (m *Model) processingView() string {
	return "🔄 Rewriting git history...\n\nThis may take a while. Please wait..."
}

func (m *Model) doneView() string {
	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46")).Render("✅ History Rewrite Complete!"))
	sb.WriteString("\n\n")
	if m.message != "" {
		sb.WriteString(m.message)
		sb.WriteString("\n\n")
	}
	sb.WriteString("The selected files have been removed from git history.\n")
	sb.WriteString("Remember to force push your changes: git push --force-with-lease\n\n")
	sb.WriteString(lipgloss.NewStyle().Faint(true).Render("Press Enter or 'q' to exit"))
	return sb.String()
}

func (m *Model) errorView() string {
	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("❌ Error"))
	sb.WriteString("\n\n")
	if m.err != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n\n", m.err))
	}
	sb.WriteString(lipgloss.NewStyle().Faint(true).Render("Press Enter or 'q' to exit"))
	return sb.String()
}

// Helper methods

func (m *Model) transitionToFileSelection() *Model {
	m.state = StateFileSelection
	m.cursor = 0
	return m
}

func (m *Model) transitionToConfirmation() *Model {
	m.state = StateConfirmation
	return m
}

func (m *Model) transitionToProcessing() *Model {
	m.state = StateProcessing
	return m
}

func (m *Model) toggleSelection() {
	largeFiles := m.stats.GetLargeFiles(m.sizeThreshold)
	if m.cursor < len(largeFiles) {
		m.selected[m.cursor] = !m.selected[m.cursor]
	}
}

func (m *Model) selectAll(selected bool) {
	largeFiles := m.stats.GetLargeFiles(m.sizeThreshold)
	for i := range largeFiles {
		m.selected[i] = selected
	}
}

func (m *Model) getSelectedFiles() []analyzer.FileInfo {
	var selected []analyzer.FileInfo
	largeFiles := m.stats.GetLargeFiles(m.sizeThreshold)
	
	for i, file := range largeFiles {
		if m.selected[i] {
			selected = append(selected, file)
		}
	}
	return selected
}

// Commands and messages

type statsMsg struct {
	stats *analyzer.Stats
	err   error
}

type errorMsg struct {
	err error
}

type processCompleteMsg struct {
	message string
	err     error
}

func (m *Model) loadStats() tea.Msg {
	log.Info().Msg("Loading git statistics")
	stats, err := m.repo.AnalyzeDiff(m.baseRef, m.compareRef, m.sizeThreshold)
	return statsMsg{stats: stats, err: err}
}

func (m *Model) processRemoval() tea.Msg {
	selectedFiles := m.getSelectedFiles()
	if len(selectedFiles) == 0 {
		return errorMsg{err: fmt.Errorf("no files selected")}
	}

	var filePaths []string
	for _, file := range selectedFiles {
		filePaths = append(filePaths, file.Path)
	}

	log.Info().Strs("files", filePaths).Msg("Starting file removal from history")
	err := m.repo.RemoveFilesFromHistory(filePaths, m.baseRef)
	if err != nil {
		return errorMsg{err: err}
	}

	message := fmt.Sprintf("Successfully removed %d files from git history", len(selectedFiles))
	return processCompleteMsg{message: message}
}
