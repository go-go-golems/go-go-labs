package main

import (
	"bufio"
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

// File represents a file or directory with extended metadata
type File struct {
	Name     string
	Path     string
	IsDir    bool
	Size     int64
	ModTime  time.Time
	Mode     os.FileMode
	Selected bool
	Hidden   bool
}

// ViewState represents the current state of the file picker
type ViewState int

const (
	ViewStateNormal ViewState = iota
	ViewStateConfirmDelete
	ViewStateRename
	ViewStateCreateFile
	ViewStateCreateDir
	ViewStateSearch
)

// Operation represents file operations
type Operation int

const (
	OpNone Operation = iota
	OpCopy
	OpCut
)

// SortMode represents different sorting options
type SortMode int

const (
	SortByName SortMode = iota
	SortBySize
	SortByDate
	SortByType
)

// FilePicker represents the advanced file picker model
type FilePicker struct {
	currentPath   string
	files         []File
	filteredFiles []File
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
	searchInput  textinput.Model
	help         help.Model
	keys         keyMap

	// Tier 4 features
	showPreview    bool
	showHidden     bool
	detailedView   bool
	sortMode       SortMode
	searchQuery    string
	previewContent string
	previewWidth   int
}

// keyMap defines the key bindings for Tier 4
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

	// Tier 4 features
	TogglePreview key.Binding
	Search        key.Binding
	ToggleHidden  key.Binding
	ToggleDetail  key.Binding
	CycleSort     key.Binding

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
		{k.TogglePreview, k.Search, k.ToggleHidden, k.ToggleDetail},
		{k.CycleSort, k.Backspace, k.Escape, k.Help, k.Quit},
	}
}

// defaultKeyMap returns the default key bindings for Tier 4
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
		TogglePreview: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "toggle preview"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		ToggleHidden: key.NewBinding(
			key.WithKeys("f2"),
			key.WithHelp("f2", "toggle hidden"),
		),
		ToggleDetail: key.NewBinding(
			key.WithKeys("f3"),
			key.WithHelp("f3", "toggle details"),
		),
		CycleSort: key.NewBinding(
			key.WithKeys("f4"),
			key.WithHelp("f4", "cycle sort"),
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

// Styles for Tier 4
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

	hiddenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	previewStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	previewTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				Bold(true)

	searchStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("99"))

	confirmStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Background(lipgloss.Color("52")).
			Foreground(lipgloss.Color("255")).
			Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

// NewFilePicker creates a new advanced file picker (Tier 4)
func NewFilePicker(startPath string) *FilePicker {
	ti := textinput.New()
	ti.Placeholder = "Enter name..."
	ti.CharLimit = 255

	si := textinput.New()
	si.Placeholder = "Search files..."
	si.CharLimit = 100

	fp := &FilePicker{
		currentPath:   startPath,
		showIcons:     true,
		showSizes:     true,
		multiSelected: make(map[string]bool),
		clipboard:     []string{},
		clipboardOp:   OpNone,
		viewState:     ViewStateNormal,
		textInput:     ti,
		searchInput:   si,
		help:          help.New(),
		keys:          defaultKeyMap(),
		showPreview:   true,
		showHidden:    false,
		detailedView:  true,
		sortMode:      SortByName,
		previewWidth:  40,
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

// Update handles messages for Tier 4
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
		case ViewStateSearch:
			return fp.updateSearch(msg)
		}
	}

	return fp, cmd
}

// updateNormal handles normal view state with Tier 4 features
func (fp *FilePicker) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, fp.keys.Quit):
		fp.cancelled = true
		return fp, tea.Quit

	case key.Matches(msg, fp.keys.Escape):
		if fp.searchQuery != "" {
			fp.searchQuery = ""
			fp.filterFiles()
		} else {
			fp.cancelled = true
			return fp, tea.Quit
		}

	case key.Matches(msg, fp.keys.Help):
		fp.help.ShowAll = !fp.help.ShowAll

	case key.Matches(msg, fp.keys.TogglePreview):
		fp.showPreview = !fp.showPreview

	case key.Matches(msg, fp.keys.Search):
		fp.searchInput.SetValue("")
		fp.searchInput.Focus()
		fp.viewState = ViewStateSearch
		return fp, textinput.Blink

	case key.Matches(msg, fp.keys.ToggleHidden):
		fp.showHidden = !fp.showHidden
		fp.loadDirectory()

	case key.Matches(msg, fp.keys.ToggleDetail):
		fp.detailedView = !fp.detailedView

	case key.Matches(msg, fp.keys.CycleSort):
		fp.sortMode = (fp.sortMode + 1) % 4
		fp.sortFiles()

	case key.Matches(msg, fp.keys.Up):
		if fp.cursor > 0 {
			fp.cursor--
			fp.updatePreview()
		}

	case key.Matches(msg, fp.keys.Down):
		if fp.cursor < len(fp.filteredFiles)-1 {
			fp.cursor++
			fp.updatePreview()
		}

	case key.Matches(msg, fp.keys.Home):
		fp.cursor = 0
		fp.updatePreview()

	case key.Matches(msg, fp.keys.End):
		if len(fp.filteredFiles) > 0 {
			fp.cursor = len(fp.filteredFiles) - 1
			fp.updatePreview()
		}

	case key.Matches(msg, fp.keys.Space):
		if len(fp.filteredFiles) > 0 {
			file := fp.filteredFiles[fp.cursor]
			if file.Name != ".." {
				if fp.multiSelected[file.Path] {
					delete(fp.multiSelected, file.Path)
				} else {
					fp.multiSelected[file.Path] = true
				}
			}
		}

	case key.Matches(msg, fp.keys.SelectAll):
		for _, file := range fp.filteredFiles {
			if file.Name != ".." {
				fp.multiSelected[file.Path] = true
			}
		}

	case key.Matches(msg, fp.keys.DeselectAll):
		fp.multiSelected = make(map[string]bool)

	case key.Matches(msg, fp.keys.SelectAllFiles):
		for _, file := range fp.filteredFiles {
			if !file.IsDir && file.Name != ".." {
				fp.multiSelected[file.Path] = true
			}
		}

	case key.Matches(msg, fp.keys.Enter):
		if len(fp.filteredFiles) > 0 {
			selectedFile := fp.filteredFiles[fp.cursor]
			if selectedFile.IsDir {
				if selectedFile.Name == ".." {
					fp.currentPath = filepath.Dir(fp.currentPath)
				} else {
					fp.currentPath = selectedFile.Path
				}
				fp.cursor = 0
				fp.multiSelected = make(map[string]bool)
				fp.searchQuery = ""
				fp.loadDirectory()
			} else {
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
		fp.searchQuery = ""
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
		if len(fp.filteredFiles) > 0 && fp.filteredFiles[fp.cursor].Name != ".." {
			fp.textInput.SetValue(fp.filteredFiles[fp.cursor].Name)
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

// updateSearch handles search input state
func (fp *FilePicker) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter", "esc":
		fp.searchQuery = fp.searchInput.Value()
		fp.searchInput.Blur()
		fp.viewState = ViewStateNormal
		fp.filterFiles()
		if fp.cursor >= len(fp.filteredFiles) {
			fp.cursor = 0
		}
		fp.updatePreview()

	default:
		fp.searchInput, cmd = fp.searchInput.Update(msg)
	}

	return fp, cmd
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

// filterFiles filters files based on search query
func (fp *FilePicker) filterFiles() {
	if fp.searchQuery == "" {
		fp.filteredFiles = fp.files
		return
	}

	fp.filteredFiles = []File{}
	query := strings.ToLower(fp.searchQuery)

	for _, file := range fp.files {
		if strings.Contains(strings.ToLower(file.Name), query) {
			fp.filteredFiles = append(fp.filteredFiles, file)
		}
	}
}

// sortFiles sorts files according to current sort mode
func (fp *FilePicker) sortFiles() {
	sort.Slice(fp.files, func(i, j int) bool {
		// Always keep parent directory at top
		if fp.files[i].Name == ".." {
			return true
		}
		if fp.files[j].Name == ".." {
			return false
		}

		// Directories first (except when sorting by type)
		if fp.sortMode != SortByType && fp.files[i].IsDir != fp.files[j].IsDir {
			return fp.files[i].IsDir
		}

		switch fp.sortMode {
		case SortBySize:
			return fp.files[i].Size < fp.files[j].Size
		case SortByDate:
			return fp.files[i].ModTime.After(fp.files[j].ModTime)
		case SortByType:
			extI := strings.ToLower(filepath.Ext(fp.files[i].Name))
			extJ := strings.ToLower(filepath.Ext(fp.files[j].Name))
			if extI != extJ {
				return extI < extJ
			}
			return strings.ToLower(fp.files[i].Name) < strings.ToLower(fp.files[j].Name)
		default: // SortByName
			return strings.ToLower(fp.files[i].Name) < strings.ToLower(fp.files[j].Name)
		}
	})

	fp.filterFiles()
}

// updatePreview updates the preview content for current file
func (fp *FilePicker) updatePreview() {
	if !fp.showPreview || len(fp.filteredFiles) == 0 {
		fp.previewContent = ""
		return
	}

	file := fp.filteredFiles[fp.cursor]

	if file.IsDir {
		fp.previewContent = fp.buildDirectoryPreview(file)
	} else {
		fp.previewContent = fp.buildFilePreview(file)
	}
}

// buildDirectoryPreview builds preview content for directories
func (fp *FilePicker) buildDirectoryPreview(file File) string {
	var content strings.Builder

	content.WriteString(previewTitleStyle.Render(file.Name) + "\n")
	content.WriteString("Type: Directory\n")
	content.WriteString(fmt.Sprintf("Modified: %s\n", file.ModTime.Format("Jan 02, 2006 15:04")))
	content.WriteString(fmt.Sprintf("Permissions: %s\n", file.Mode.String()))

	// Try to count items in directory
	if entries, err := os.ReadDir(file.Path); err == nil {
		visibleCount := 0
		hiddenCount := 0
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") {
				hiddenCount++
			} else {
				visibleCount++
			}
		}
		content.WriteString(fmt.Sprintf("Items: %d", visibleCount))
		if hiddenCount > 0 {
			content.WriteString(fmt.Sprintf(" (%d hidden)", hiddenCount))
		}
		content.WriteString("\n")
	}

	return content.String()
}

// buildFilePreview builds preview content for files
func (fp *FilePicker) buildFilePreview(file File) string {
	var content strings.Builder

	content.WriteString(previewTitleStyle.Render(file.Name) + "\n")
	content.WriteString(fmt.Sprintf("Size: %s\n", fp.formatFileSize(file.Size)))
	content.WriteString(fmt.Sprintf("Modified: %s\n", file.ModTime.Format("Jan 02, 2006 15:04")))
	content.WriteString(fmt.Sprintf("Permissions: %s\n", file.Mode.String()))
	content.WriteString(strings.Repeat("─", 20) + "\n")

	// Try to preview file content
	if fp.isTextFile(file.Name) && file.Size < 10*1024 { // Only preview small text files
		if preview := fp.readFilePreview(file.Path); preview != "" {
			content.WriteString(preview)
		} else {
			content.WriteString("[Unable to read file]")
		}
	} else if fp.isImageFile(file.Name) {
		content.WriteString("[Image file]\n")
		if info, err := os.Stat(file.Path); err == nil {
			content.WriteString(fmt.Sprintf("Size: %dx? pixels\n", info.Size()))
		}
	} else if fp.isArchiveFile(file.Name) {
		content.WriteString("[Archive file]\n")
		content.WriteString("Use 'file' command for details")
	} else {
		content.WriteString("[Binary file]")
	}

	return content.String()
}

// isTextFile checks if a file is likely a text file
func (fp *FilePicker) isTextFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	textExts := []string{
		".txt", ".md", ".readme", ".go", ".py", ".js", ".html", ".css",
		".json", ".xml", ".yml", ".yaml", ".toml", ".ini", ".conf",
		".sh", ".bat", ".ps1", ".php", ".rb", ".pl", ".java", ".cpp",
		".c", ".h", ".hpp", ".rs", ".swift", ".kt", ".scala", ".clj",
	}

	for _, textExt := range textExts {
		if ext == textExt {
			return true
		}
	}
	return false
}

// isImageFile checks if a file is an image
func (fp *FilePicker) isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp", ".tiff", ".ico"}

	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

// isArchiveFile checks if a file is an archive
func (fp *FilePicker) isArchiveFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	archiveExts := []string{".zip", ".tar", ".gz", ".rar", ".7z", ".bz2", ".xz", ".lz", ".lzma"}

	for _, archExt := range archiveExts {
		if ext == archExt {
			return true
		}
	}
	return false
}

// readFilePreview reads a preview of a text file
func (fp *FilePicker) readFilePreview(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineCount := 0
	maxLines := 15

	for scanner.Scan() && lineCount < maxLines {
		line := scanner.Text()
		if len(line) > 50 {
			line = line[:50] + "..."
		}
		lines = append(lines, line)
		lineCount++
	}

	if lineCount == maxLines {
		lines = append(lines, "...")
	}

	return strings.Join(lines, "\n")
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

	if len(fp.filteredFiles) > 0 && fp.filteredFiles[fp.cursor].Name != ".." {
		return []string{fp.filteredFiles[fp.cursor].Path}
	}

	return []string{}
}

// File operation methods (same as Tier 3, but working with filteredFiles)
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

func (fp *FilePicker) performRename(newName string) {
	if len(fp.filteredFiles) > 0 && fp.filteredFiles[fp.cursor].Name != ".." {
		oldPath := fp.filteredFiles[fp.cursor].Path
		newPath := filepath.Join(fp.currentPath, newName)

		if err := os.Rename(oldPath, newPath); err != nil {
			fp.err = fmt.Errorf("failed to rename: %v", err)
			return
		}

		delete(fp.multiSelected, oldPath)
		fp.loadDirectory()
	}
}

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

func (fp *FilePicker) performCreateDir(name string) {
	dirPath := filepath.Join(fp.currentPath, name)

	if err := os.Mkdir(dirPath, 0755); err != nil {
		fp.err = fmt.Errorf("failed to create directory: %v", err)
		return
	}

	fp.loadDirectory()
}

// View renders the advanced file picker (Tier 4)
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

// viewNormal renders the normal file picker view with preview panel
func (fp *FilePicker) viewNormal() string {
	// Calculate panel widths
	var fileListWidth, previewWidth int
	if fp.showPreview {
		previewWidth = (fp.width * fp.previewWidth) / 100
		if previewWidth < 20 {
			previewWidth = 2
		}
		if previewWidth > fp.width-40 {
			previewWidth = fp.width - 40
		}
		fileListWidth = fp.width - previewWidth - 6 // Account for borders and separator
	} else {
		fileListWidth = fp.width - 4
		previewWidth = 0
	}

	// Build file list panel
	filePanel := fp.buildFileListPanel(fileListWidth)

	if fp.showPreview {
		// Build preview panel
		previewPanel := fp.buildPreviewPanel(previewWidth)

		// Combine panels side by side
		panelsView := lipgloss.JoinHorizontal(
			lipgloss.Top,
			borderStyle.Width(fileListWidth).Render(filePanel),
			borderStyle.Width(previewWidth).Render(previewPanel),
		)

		// Add help below panels
		helpView := fp.help.View(fp.keys)
		if helpView != "" {
			return panelsView + "\n" + helpView
		}
		return panelsView
	} else {
		return borderStyle.Width(fileListWidth).Render(filePanel) + "\n" + fp.help.View(fp.keys)
	}
}

// buildFileListPanel builds the file list panel content
func (fp *FilePicker) buildFileListPanel(width int) string {
	var b strings.Builder
	contentWidth := width - 2

	// Title with status
	title := titleStyle.Render("File Explorer")
	if fp.searchQuery != "" {
		title += statusStyle.Render(fmt.Sprintf(" (search: %s)", fp.searchQuery))
	}
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
	contentHeight := fp.height - 9 // Account for title, path, status, search, help

	startIdx := 0
	endIdx := len(fp.filteredFiles)

	// Calculate visible range for scrolling
	if len(fp.filteredFiles) > contentHeight {
		if fp.cursor >= contentHeight/2 {
			startIdx = fp.cursor - contentHeight/2
			endIdx = startIdx + contentHeight
			if endIdx > len(fp.filteredFiles) {
				endIdx = len(fp.filteredFiles)
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
		file := fp.filteredFiles[i]
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

	// Search input (if active)
	if fp.viewState == ViewStateSearch {
		b.WriteString(searchStyle.Render("Search: ") + fp.searchInput.View())
	} else if fp.searchQuery != "" {
		b.WriteString(searchStyle.Render(fmt.Sprintf("Search: %s (%d matches)", fp.searchQuery, len(fp.filteredFiles))))
	}

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
		if fp.viewState == ViewStateSearch {
			b.WriteString("\n")
		}
		b.WriteString(prompt + fp.textInput.View())
	}

	// Add line break only if we had search or text input
	if fp.viewState == ViewStateSearch || fp.viewState == ViewStateRename ||
		fp.viewState == ViewStateCreateFile || fp.viewState == ViewStateCreateDir ||
		fp.searchQuery != "" {
		b.WriteString("\n")
	}

	// Status line
	status := fp.buildStatusLine()
	b.WriteString(status)

	// Error display
	if fp.err != nil {
		b.WriteString("\n" + errorStyle.Render("Error: "+fp.err.Error()))
		fp.err = nil
	}

	return b.String()
}

// buildPreviewPanel builds the preview panel content
func (fp *FilePicker) buildPreviewPanel(width int) string {
	var b strings.Builder
	contentWidth := width - 2

	// Preview title
	b.WriteString(previewTitleStyle.Render("Preview") + "\n")
	b.WriteString(strings.Repeat("─", contentWidth) + "\n")

	// Preview content with better formatting
	if fp.previewContent != "" {
		lines := strings.Split(fp.previewContent, "\n")
		maxLines := fp.height - 8

		for i, line := range lines {
			if i >= maxLines {
				b.WriteString("...\n")
				break
			}
			if len(line) > contentWidth-1 {
				line = line[:contentWidth-3] + "..."
			}
			b.WriteString(" " + line + "\n") // Add leading space for readability
		}
	} else {
		b.WriteString(" No preview available\n")
	}

	return b.String()
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

	return lipgloss.Place(fp.width, fp.height, lipgloss.Center, lipgloss.Center, dialog)
}

// buildStatusLine builds the status line with Tier 4 info
func (fp *FilePicker) buildStatusLine() string {
	var parts []string

	// File count and filtering info
	if fp.searchQuery != "" {
		parts = append(parts, fmt.Sprintf("%d of %d items", len(fp.filteredFiles), len(fp.files)))
	} else {
		parts = append(parts, fmt.Sprintf("%d items", len(fp.files)))
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

	// Sort mode
	sortModes := []string{"Name", "Size", "Date", "Type"}
	parts = append(parts, fmt.Sprintf("Sort: %s", sortModes[fp.sortMode]))

	// View options
	var options []string
	if fp.showHidden {
		options = append(options, "Hidden")
	}
	if fp.detailedView {
		options = append(options, "Details")
	}
	if fp.showPreview {
		options = append(options, "Preview")
	}
	if len(options) > 0 {
		parts = append(parts, strings.Join(options, ","))
	}

	return statusStyle.Render(strings.Join(parts, " | "))
}

// formatFileEntry formats a single file entry with proper table columns
func (fp *FilePicker) formatFileEntry(file File, isCursor bool, width int) string {
	if file.Hidden && !fp.showHidden {
		return "" // Skip hidden files if not showing them
	}

	// Column widths - responsive design with column hiding (no permissions)
	const (
		indicatorWidth = 4  // "✓▶  "
		iconWidth      = 4  // "📁  "
		sizeWidth      = 12 // "  1.23 GB  " (can get quite wide)
		dateWidth      = 10 // " Jan 02   "
		spacerWidth    = 2  // Extra spacing between sections
		sizeDateSpacer = 4  // Extra spacing between size and date (wider files)
		minNameWidth   = 25 // Minimum name column width
	)

	// Calculate which columns to show based on available width
	baseWidth := indicatorWidth + iconWidth + spacerWidth + minNameWidth + 8 // 8 for padding/safety
	showSize := fp.detailedView && fp.showSizes
	showDate := fp.detailedView

	// Progressive column hiding based on available width (no permissions column)
	fullWidth := baseWidth + sizeWidth + sizeDateSpacer + dateWidth
	if width < fullWidth {
		// Not enough space for all columns, start hiding
		if width < baseWidth + sizeWidth + sizeDateSpacer {
			showDate = false // Hide date first
		}
		if width < baseWidth + sizeWidth {
			showSize = false // Hide size last
		}
	}

	// Calculate actual fixed width based on what we're showing
	fixedWidth := indicatorWidth + iconWidth + spacerWidth
	if showSize {
		fixedWidth += sizeWidth + sizeDateSpacer
	}
	if showDate {
		fixedWidth += dateWidth
	}

	nameWidth := width - fixedWidth - 8 // -8 for border padding and extra safety margin
	if nameWidth < minNameWidth {
		nameWidth = minNameWidth
	}

	var line strings.Builder

	// Selection indicators (fixed width)
	indicator := "   "
	if isCursor && fp.multiSelected[file.Path] {
		indicator = "✓▶ "
	} else if isCursor {
		indicator = "▶  "
	} else if fp.multiSelected[file.Path] {
		indicator = "✓  "
	}
	line.WriteString(fmt.Sprintf("%-*s", indicatorWidth, indicator))

	// Icon (fixed width)
	if fp.showIcons {
		icon := fp.getFileIcon(file)
		line.WriteString(fmt.Sprintf("%-*s", iconWidth, icon))
	}

	// Spacer after icon
	line.WriteString(strings.Repeat(" ", spacerWidth))

	// Name (variable width, left-aligned)
	name := file.Name
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}
	line.WriteString(fmt.Sprintf("%-*s", nameWidth, name))

	// Add detail columns based on available space
	if showSize {
		// Size column (right-aligned with extra spacing after)
		if !file.IsDir && file.Name != ".." {
			size := fp.formatFileSize(file.Size)
			line.WriteString(fmt.Sprintf("%*s", sizeWidth, size))
		} else {
			line.WriteString(fmt.Sprintf("%*s", sizeWidth, ""))
		}
		// Extra spacer after size column (since sizes can be wide)
		line.WriteString(strings.Repeat(" ", sizeDateSpacer))
	}

	if showDate {
		// Date column (left-aligned)
		if file.Name != ".." {
			modTime := file.ModTime.Format("Jan 02")
			line.WriteString(fmt.Sprintf("%-*s", dateWidth, modTime))
		} else {
			line.WriteString(fmt.Sprintf("%-*s", dateWidth, ""))
		}
	}

	result := line.String()

	// Ensure we don't exceed width
	if len(result) > width-2 {
		result = result[:width-2]
	}

	// Apply styling
	if file.Hidden {
		result = hiddenStyle.Render(result)
	} else if isCursor && fp.multiSelected[file.Path] {
		result = multiSelectedStyle.Render(result)
	} else if isCursor {
		result = selectedStyle.Render(result)
	} else if fp.multiSelected[file.Path] {
		result = multiSelectedStyle.Render(result)
	} else if file.IsDir {
		result = dirStyle.Render(result)
	} else {
		result = normalStyle.Render(result)
	}

	return result
}

// getFileIcon returns an appropriate icon for the file (extended for Tier 4)
func (fp *FilePicker) getFileIcon(file File) string {
	if file.Name == ".." {
		return "📁"
	}
	if file.IsDir {
		if file.Hidden {
			return "👻"
		}
		return "📁"
	}

	if file.Hidden {
		return "👻"
	}

	// Detailed file type detection for Tier 4
	ext := strings.ToLower(filepath.Ext(file.Name))
	filename := strings.ToLower(file.Name)

	// Archives
	switch ext {
	case ".zip", ".tar", ".gz", ".rar", ".7z", ".bz2", ".xz":
		return "📦"
	}

	// Documents
	switch ext {
	case ".pdf", ".doc", ".docx", ".odt":
		return "📋"
	case ".xls", ".xlsx", ".ods", ".csv":
		return "📊"
	case ".ppt", ".pptx", ".odp":
		return "▶️"
	}

	// Media
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp", ".tiff":
		return "🖼️"
	case ".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm":
		return "🎬"
	case ".mp3", ".wav", ".flac", ".ogg", ".aac", ".m4a":
		return "🎵"
	}

	// Code files
	switch ext {
	case ".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", ".rs", ".swift":
		return "💻"
	case ".html", ".css", ".scss", ".less":
		return "💻"
	case ".json", ".xml", ".yml", ".yaml", ".toml":
		return "💻"
	}

	// Scripts and executables
	switch ext {
	case ".sh", ".bat", ".ps1", ".cmd":
		return "⚙️"
	case ".exe", ".msi", ".deb", ".rpm", ".dmg", ".app":
		return "⚙️"
	}

	// Text files
	switch ext {
	case ".txt", ".md", ".readme", ".log":
		return "📄"
	}

	// Symlinks (would need additional detection)
	if file.Mode&os.ModeSymlink != 0 {
		return "🔗"
	}

	// Read-only files
	if file.Mode&0200 == 0 {
		return "🔒"
	}

	// Special files
	if strings.Contains(filename, "readme") {
		return "📄"
	}
	if strings.Contains(filename, "license") {
		return "📄"
	}
	if strings.Contains(filename, "makefile") || strings.Contains(filename, "dockerfile") {
		return "⚙️"
	}

	return "❓" // Unknown file type
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

// loadDirectory loads the contents of the current directory (enhanced for Tier 4)
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
		if info, err := os.Stat(filepath.Dir(fp.currentPath)); err == nil {
			fp.files = append(fp.files, File{
				Name:    "..",
				Path:    filepath.Dir(fp.currentPath),
				IsDir:   true,
				Size:    0,
				ModTime: info.ModTime(),
				Mode:    info.Mode(),
				Hidden:  false,
			})
		}
	}

	// Add directory entries
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		isHidden := strings.HasPrefix(entry.Name(), ".")

		// Skip hidden files if not showing them
		if isHidden && !fp.showHidden {
			continue
		}

		file := File{
			Name:    entry.Name(),
			Path:    filepath.Join(fp.currentPath, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Mode:    info.Mode(),
			Hidden:  isHidden,
		}
		fp.files = append(fp.files, file)
	}

	// Sort files
	fp.sortFiles()

	// Reset cursor if out of bounds
	if fp.cursor >= len(fp.filteredFiles) {
		fp.cursor = len(fp.filteredFiles) - 1
	}
	if fp.cursor < 0 {
		fp.cursor = 0
	}

	// Update preview for current file
	fp.updatePreview()
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
