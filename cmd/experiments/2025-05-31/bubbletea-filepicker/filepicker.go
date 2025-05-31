package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// File represents a file or directory
type File struct {
	Name     string
	Path     string
	IsDir    bool
	Size     int64
	ModTime  time.Time
	Selected bool
}

// ViewState represents the current state of the file picker
type ViewState int

const (
	ViewStateNormal ViewState = iota
	ViewStateConfirmDelete
	ViewStateRename
	ViewStateCreateFile
	ViewStateCreateDir
)

// Operation represents file operations
type Operation int

const (
	OpNone Operation = iota
	OpCopy
	OpCut
)

// FilePicker represents the file picker model
type FilePicker struct {
	currentPath   string
	files         []File
	cursor        int
	selectedFiles []string // For final selection
	width         int
	height        int
	cancelled     bool
	showIcons     bool
	showSizes     bool
	err           error

	// Multi-selection state
	multiSelected map[string]bool

	// Operations
	clipboard   []string
	clipboardOp Operation

	// UI state
	viewState    ViewState
	confirmFiles []string
	textInput    textinput.Model
	help         help.Model
	keys         keyMap
}

// keyMap defines the key bindings
type keyMap struct {
	// Navigation
	Up   key.Binding
	Down key.Binding
	Home key.Binding
	End  key.Binding

	// Selection
	Enter          key.Binding
	Space          key.Binding
	SelectAll      key.Binding
	DeselectAll    key.Binding
	SelectAllFiles key.Binding

	// File operations
	Delete  key.Binding
	Copy    key.Binding
	Cut     key.Binding
	Paste   key.Binding
	Rename  key.Binding
	NewFile key.Binding
	NewDir  key.Binding

	// Navigation
	Escape    key.Binding
	Backspace key.Binding
	Refresh   key.Binding

	// System
	Help key.Binding
	Quit key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Home, k.End},
		{k.Enter, k.Space, k.SelectAll, k.DeselectAll},
		{k.Copy, k.Cut, k.Paste, k.Delete},
		{k.Rename, k.NewFile, k.NewDir, k.Refresh},
		{k.Backspace, k.Escape, k.Help, k.Quit},
	}
}

// defaultKeyMap returns the default key bindings
func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "first"),
		),
		End: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "last"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select/enter"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle selection"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "select all"),
		),
		DeselectAll: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "deselect all"),
		),
		SelectAllFiles: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "select all files"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Copy: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copy"),
		),
		Cut: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "cut"),
		),
		Paste: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "paste"),
		),
		Rename: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rename"),
		),
		NewFile: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new file"),
		),
		NewDir: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "new directory"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Backspace: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "up directory"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("f5"),
			key.WithHelp("f5", "refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// Styles
var (
	borderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230"))

	multiSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("33")).
				Foreground(lipgloss.Color("230"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	confirmStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Background(lipgloss.Color("52")).
			Foreground(lipgloss.Color("255")).
			Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

// NewFilePicker creates a new file picker
func NewFilePicker(startPath string) *FilePicker {
	ti := textinput.New()
	ti.Placeholder = "Enter name..."
	ti.CharLimit = 255

	fp := &FilePicker{
		currentPath:   startPath,
		showIcons:     true,
		showSizes:     true,
		multiSelected: make(map[string]bool),
		clipboard:     []string{},
		clipboardOp:   OpNone,
		viewState:     ViewStateNormal,
		textInput:     ti,
		help:          help.New(),
		keys:          defaultKeyMap(),
	}

	// Resolve the starting path
	if absPath, err := filepath.Abs(startPath); err == nil {
		fp.currentPath = absPath
	}

	fp.loadDirectory()
	return fp
}

// Init initializes the file picker
func (fp *FilePicker) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (fp *FilePicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		fp.width = msg.Width
		fp.height = msg.Height
		fp.help.Width = msg.Width

	case tea.KeyMsg:
		switch fp.viewState {
		case ViewStateNormal:
			return fp.updateNormal(msg)
		case ViewStateConfirmDelete:
			return fp.updateConfirmDelete(msg)
		case ViewStateRename, ViewStateCreateFile, ViewStateCreateDir:
			return fp.updateTextInput(msg)
		}
	}

	return fp, cmd
}

// updateNormal handles normal view state
func (fp *FilePicker) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, fp.keys.Quit):
		fp.cancelled = true
		return fp, tea.Quit

	case key.Matches(msg, fp.keys.Escape):
		fp.cancelled = true
		return fp, tea.Quit

	case key.Matches(msg, fp.keys.Help):
		fp.help.ShowAll = !fp.help.ShowAll

	case key.Matches(msg, fp.keys.Up):
		if fp.cursor > 0 {
			fp.cursor--
		}

	case key.Matches(msg, fp.keys.Down):
		if fp.cursor < len(fp.files)-1 {
			fp.cursor++
		}

	case key.Matches(msg, fp.keys.Home):
		fp.cursor = 0

	case key.Matches(msg, fp.keys.End):
		if len(fp.files) > 0 {
			fp.cursor = len(fp.files) - 1
		}

	case key.Matches(msg, fp.keys.Space):
		if len(fp.files) > 0 {
			file := fp.files[fp.cursor]
			if file.Name != ".." {
				if fp.multiSelected[file.Path] {
					delete(fp.multiSelected, file.Path)
				} else {
					fp.multiSelected[file.Path] = true
				}
			}
		}

	case key.Matches(msg, fp.keys.SelectAll):
		for _, file := range fp.files {
			if file.Name != ".." {
				fp.multiSelected[file.Path] = true
			}
		}

	case key.Matches(msg, fp.keys.DeselectAll):
		fp.multiSelected = make(map[string]bool)

	case key.Matches(msg, fp.keys.SelectAllFiles):
		for _, file := range fp.files {
			if !file.IsDir && file.Name != ".." {
				fp.multiSelected[file.Path] = true
			}
		}

	case key.Matches(msg, fp.keys.Enter):
		if len(fp.files) > 0 {
			selectedFile := fp.files[fp.cursor]
			if selectedFile.IsDir {
				// Enter directory
				if selectedFile.Name == ".." {
					fp.currentPath = filepath.Dir(fp.currentPath)
				} else {
					fp.currentPath = selectedFile.Path
				}
				fp.cursor = 0
				fp.multiSelected = make(map[string]bool)
				fp.loadDirectory()
			} else {
				// Select file (or multiple files if any are multi-selected)
				if len(fp.multiSelected) > 0 {
					fp.selectedFiles = make([]string, 0, len(fp.multiSelected))
					for path := range fp.multiSelected {
						fp.selectedFiles = append(fp.selectedFiles, path)
					}
				} else {
					fp.selectedFiles = []string{selectedFile.Path}
				}
				return fp, tea.Quit
			}
		}

	case key.Matches(msg, fp.keys.Backspace):
		fp.currentPath = filepath.Dir(fp.currentPath)
		fp.cursor = 0
		fp.multiSelected = make(map[string]bool)
		fp.loadDirectory()

	case key.Matches(msg, fp.keys.Refresh):
		fp.loadDirectory()

	case key.Matches(msg, fp.keys.Delete):
		filesToDelete := fp.getSelectedFiles()
		if len(filesToDelete) > 0 {
			fp.confirmFiles = filesToDelete
			fp.viewState = ViewStateConfirmDelete
		}

	case key.Matches(msg, fp.keys.Copy):
		fp.clipboard = fp.getSelectedFiles()
		fp.clipboardOp = OpCopy

	case key.Matches(msg, fp.keys.Cut):
		fp.clipboard = fp.getSelectedFiles()
		fp.clipboardOp = OpCut

	case key.Matches(msg, fp.keys.Paste):
		if len(fp.clipboard) > 0 {
			fp.performPaste()
		}

	case key.Matches(msg, fp.keys.Rename):
		if len(fp.files) > 0 && fp.files[fp.cursor].Name != ".." {
			fp.textInput.SetValue(fp.files[fp.cursor].Name)
			fp.textInput.CursorEnd()
			fp.textInput.Focus()
			fp.viewState = ViewStateRename
			return fp, textinput.Blink
		}

	case key.Matches(msg, fp.keys.NewFile):
		fp.textInput.SetValue("")
		fp.textInput.Focus()
		fp.viewState = ViewStateCreateFile
		return fp, textinput.Blink

	case key.Matches(msg, fp.keys.NewDir):
		fp.textInput.SetValue("")
		fp.textInput.Focus()
		fp.viewState = ViewStateCreateDir
		return fp, textinput.Blink
	}

	return fp, nil
}

// updateConfirmDelete handles delete confirmation
func (fp *FilePicker) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		fp.performDelete()
		fp.viewState = ViewStateNormal
	case "n", "N", "esc":
		fp.viewState = ViewStateNormal
	}
	return fp, nil
}

// updateTextInput handles text input states
func (fp *FilePicker) updateTextInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(fp.textInput.Value())
		if name != "" {
			switch fp.viewState {
			case ViewStateRename:
				fp.performRename(name)
			case ViewStateCreateFile:
				fp.performCreateFile(name)
			case ViewStateCreateDir:
				fp.performCreateDir(name)
			}
		}
		fp.textInput.Blur()
		fp.viewState = ViewStateNormal

	case "esc":
		fp.textInput.Blur()
		fp.viewState = ViewStateNormal

	default:
		fp.textInput, cmd = fp.textInput.Update(msg)
	}

	return fp, cmd
}

// getSelectedFiles returns the list of selected files
func (fp *FilePicker) getSelectedFiles() []string {
	if len(fp.multiSelected) > 0 {
		files := make([]string, 0, len(fp.multiSelected))
		for path := range fp.multiSelected {
			files = append(files, path)
		}
		return files
	}

	if len(fp.files) > 0 && fp.files[fp.cursor].Name != ".." {
		return []string{fp.files[fp.cursor].Path}
	}

	return []string{}
}

// performDelete deletes the confirmed files
func (fp *FilePicker) performDelete() {
	for _, filePath := range fp.confirmFiles {
		if err := os.RemoveAll(filePath); err != nil {
			fp.err = fmt.Errorf("failed to delete %s: %v", filepath.Base(filePath), err)
			return
		}
		delete(fp.multiSelected, filePath)
	}
	fp.loadDirectory()
}

// performPaste performs copy or cut operation
func (fp *FilePicker) performPaste() {
	for _, src := range fp.clipboard {
		dst := filepath.Join(fp.currentPath, filepath.Base(src))

		if fp.clipboardOp == OpCopy {
			if err := fp.copyFile(src, dst); err != nil {
				fp.err = fmt.Errorf("failed to copy %s: %v", filepath.Base(src), err)
				return
			}
		} else if fp.clipboardOp == OpCut {
			if err := os.Rename(src, dst); err != nil {
				fp.err = fmt.Errorf("failed to move %s: %v", filepath.Base(src), err)
				return
			}
		}
	}

	if fp.clipboardOp == OpCut {
		fp.clipboard = []string{}
		fp.clipboardOp = OpNone
	}

	fp.loadDirectory()
}

// copyFile copies a file or directory
func (fp *FilePicker) copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return fp.copyDir(src, dst)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// copyDir copies a directory recursively
func (fp *FilePicker) copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if err := fp.copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

// performRename renames the current file
func (fp *FilePicker) performRename(newName string) {
	if len(fp.files) > 0 && fp.files[fp.cursor].Name != ".." {
		oldPath := fp.files[fp.cursor].Path
		newPath := filepath.Join(fp.currentPath, newName)

		if err := os.Rename(oldPath, newPath); err != nil {
			fp.err = fmt.Errorf("failed to rename: %v", err)
			return
		}

		delete(fp.multiSelected, oldPath)
		fp.loadDirectory()
	}
}

// performCreateFile creates a new file
func (fp *FilePicker) performCreateFile(name string) {
	filePath := filepath.Join(fp.currentPath, name)

	file, err := os.Create(filePath)
	if err != nil {
		fp.err = fmt.Errorf("failed to create file: %v", err)
		return
	}
	file.Close()

	fp.loadDirectory()
}

// performCreateDir creates a new directory
func (fp *FilePicker) performCreateDir(name string) {
	dirPath := filepath.Join(fp.currentPath, name)

	if err := os.Mkdir(dirPath, 0755); err != nil {
		fp.err = fmt.Errorf("failed to create directory: %v", err)
		return
	}

	fp.loadDirectory()
}

// View renders the file picker
func (fp *FilePicker) View() string {
	if fp.width == 0 || fp.height == 0 {
		return "Loading..."
	}

	switch fp.viewState {
	case ViewStateConfirmDelete:
		return fp.viewConfirmDelete()
	default:
		return fp.viewNormal()
	}
}

// viewNormal renders the normal file picker view
func (fp *FilePicker) viewNormal() string {
	// Calculate content dimensions
	helpHeight := lipgloss.Height(fp.help.View(fp.keys))
	contentWidth := fp.width - 4
	contentHeight := fp.height - helpHeight - 8 // Account for border, title, path, status, input

	var b strings.Builder

	// Title
	title := titleStyle.Render("File Manager")
	if len(fp.multiSelected) > 0 {
		title += statusStyle.Render(fmt.Sprintf(" (%d selected)", len(fp.multiSelected)))
	}
	b.WriteString(title + "\n")

	// Current path
	path := pathStyle.Render("Path: " + fp.currentPath)
	b.WriteString(path + "\n")

	// Separator
	b.WriteString(strings.Repeat("─", contentWidth) + "\n")

	// File list
	startIdx := 0
	endIdx := len(fp.files)

	// Calculate visible range for scrolling
	if len(fp.files) > contentHeight {
		if fp.cursor >= contentHeight/2 {
			startIdx = fp.cursor - contentHeight/2
			endIdx = startIdx + contentHeight
			if endIdx > len(fp.files) {
				endIdx = len(fp.files)
				startIdx = endIdx - contentHeight
				if startIdx < 0 {
					startIdx = 0
				}
			}
		} else {
			endIdx = contentHeight
		}
	}

	for i := startIdx; i < endIdx; i++ {
		file := fp.files[i]
		line := fp.formatFileEntry(file, i == fp.cursor, contentWidth)
		b.WriteString(line + "\n")
	}

	// Fill remaining space
	remaining := contentHeight - (endIdx - startIdx)
	for i := 0; i < remaining; i++ {
		b.WriteString(strings.Repeat(" ", contentWidth) + "\n")
	}

	// Separator
	b.WriteString(strings.Repeat("─", contentWidth) + "\n")

	// Status line
	status := fp.buildStatusLine()
	b.WriteString(status + "\n")

	// Text input (if active)
	if fp.viewState == ViewStateRename || fp.viewState == ViewStateCreateFile || fp.viewState == ViewStateCreateDir {
		var prompt string
		switch fp.viewState {
		case ViewStateRename:
			prompt = "Rename: "
		case ViewStateCreateFile:
			prompt = "New file: "
		case ViewStateCreateDir:
			prompt = "New directory: "
		}
		b.WriteString(prompt + fp.textInput.View() + "\n")
	} else {
		b.WriteString("\n")
	}

	// Error display
	if fp.err != nil {
		b.WriteString(errorStyle.Render("Error: "+fp.err.Error()) + "\n")
		fp.err = nil // Clear error after displaying
	} else {
		b.WriteString("\n")
	}

	// Help
	b.WriteString(fp.help.View(fp.keys))

	// Apply border
	content := b.String()
	return borderStyle.Width(contentWidth + 2).Render(content)
}

// viewConfirmDelete renders the delete confirmation dialog
func (fp *FilePicker) viewConfirmDelete() string {
	var b strings.Builder

	b.WriteString("Delete the following files?\n\n")

	for _, filePath := range fp.confirmFiles {
		b.WriteString("• " + filepath.Base(filePath) + "\n")
	}

	b.WriteString("\n[Y] Yes    [N] No")

	dialog := confirmStyle.Render(b.String())

	// Center the dialog
	return lipgloss.Place(fp.width, fp.height, lipgloss.Center, lipgloss.Center, dialog)
}

// buildStatusLine builds the status line
func (fp *FilePicker) buildStatusLine() string {
	var parts []string

	if len(fp.files) > 0 {
		selectedFile := fp.files[fp.cursor]
		parts = append(parts, fmt.Sprintf("Current: %s", selectedFile.Name))
	}

	if len(fp.multiSelected) > 0 {
		parts = append(parts, fmt.Sprintf("%d selected", len(fp.multiSelected)))
	}

	if len(fp.clipboard) > 0 {
		op := "copied"
		if fp.clipboardOp == OpCut {
			op = "cut"
		}
		parts = append(parts, fmt.Sprintf("%d %s", len(fp.clipboard), op))
	}

	parts = append(parts, fmt.Sprintf("Total: %d items", len(fp.files)))

	return statusStyle.Render(strings.Join(parts, " | "))
}

// formatFileEntry formats a single file entry
func (fp *FilePicker) formatFileEntry(file File, isCursor bool, width int) string {
	var parts []string

	// Selection indicators
	indicator := " "
	if isCursor && fp.multiSelected[file.Path] {
		indicator = "✓▶"
	} else if isCursor {
		indicator = "▶"
	} else if fp.multiSelected[file.Path] {
		indicator = "✓"
	}
	parts = append(parts, indicator)

	// Icon
	if fp.showIcons {
		icon := fp.getFileIcon(file)
		parts = append(parts, icon)
	}

	// Name
	name := file.Name
	parts = append(parts, name)

	// Size (if enabled and not a directory)
	if fp.showSizes && !file.IsDir && file.Name != ".." {
		size := fp.formatFileSize(file.Size)
		parts = append(parts, size)
	}

	// Modification date
	if file.Name != ".." {
		modTime := file.ModTime.Format("Jan 02")
		parts = append(parts, modTime)
	}

	// Join parts
	line := strings.Join(parts, " ")

	// Truncate if too long
	if len(line) > width-2 {
		line = line[:width-5] + "..."
	}

	// Pad to full width
	line = fmt.Sprintf("%-*s", width-2, line)

	// Apply styling
	if isCursor && fp.multiSelected[file.Path] {
		return multiSelectedStyle.Render(line)
	} else if isCursor {
		return selectedStyle.Render(line)
	} else if fp.multiSelected[file.Path] {
		return multiSelectedStyle.Render(line)
	} else if file.IsDir {
		return dirStyle.Render(line)
	} else {
		return normalStyle.Render(line)
	}
}

// getFileIcon returns an appropriate icon for the file
func (fp *FilePicker) getFileIcon(file File) string {
	if file.Name == ".." {
		return "📁"
	}
	if file.IsDir {
		return "📁"
	}

	// Basic file type detection based on extension
	ext := strings.ToLower(filepath.Ext(file.Name))
	switch ext {
	case ".txt", ".md", ".readme":
		return "📄"
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg":
		return "🖼️"
	case ".zip", ".tar", ".gz", ".rar", ".7z":
		return "📦"
	case ".sh", ".exe", ".bat", ".cmd":
		return "⚙️"
	case ".go", ".py", ".js", ".html", ".css", ".java", ".cpp", ".c":
		return "💻"
	default:
		return "📄"
	}
}

// formatFileSize formats file size in human-readable format
func (fp *FilePicker) formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// loadDirectory loads the contents of the current directory
func (fp *FilePicker) loadDirectory() {
	fp.files = []File{}
	fp.err = nil

	entries, err := os.ReadDir(fp.currentPath)
	if err != nil {
		fp.err = err
		return
	}

	// Add parent directory entry if not at root
	if fp.currentPath != "/" && fp.currentPath != "\\" {
		fp.files = append(fp.files, File{
			Name:  "..",
			Path:  filepath.Dir(fp.currentPath),
			IsDir: true,
		})
	}

	// Add directory entries
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		file := File{
			Name:    entry.Name(),
			Path:    filepath.Join(fp.currentPath, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		fp.files = append(fp.files, file)
	}

	// Sort files: directories first, then files, both alphabetically
	sort.Slice(fp.files, func(i, j int) bool {
		// Special case for parent directory
		if fp.files[i].Name == ".." {
			return true
		}
		if fp.files[j].Name == ".." {
			return false
		}

		// Directories first
		if fp.files[i].IsDir != fp.files[j].IsDir {
			return fp.files[i].IsDir
		}

		// Then alphabetically
		return strings.ToLower(fp.files[i].Name) < strings.ToLower(fp.files[j].Name)
	})

	// Reset cursor if out of bounds
	if fp.cursor >= len(fp.files) {
		fp.cursor = len(fp.files) - 1
	}
	if fp.cursor < 0 {
		fp.cursor = 0
	}
}

// GetSelected returns the selected file paths
func (fp *FilePicker) GetSelected() ([]string, bool) {
	if fp.cancelled {
		return nil, false
	}
	return fp.selectedFiles, len(fp.selectedFiles) > 0
}

// GetError returns any error that occurred
func (fp *FilePicker) GetError() error {
	return fp.err
}
