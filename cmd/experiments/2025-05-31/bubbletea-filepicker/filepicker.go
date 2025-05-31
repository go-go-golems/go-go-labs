package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
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

// FilePicker represents the file picker model
type FilePicker struct {
	currentPath string
	files       []File
	cursor      int
	selected    string
	width       int
	height      int
	cancelled   bool
	showIcons   bool
	showSizes   bool
	err         error
}

// keyMap defines the key bindings
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Escape    key.Binding
	Backspace key.Binding
	Refresh   key.Binding
	Home      key.Binding
	End       key.Binding
	Quit      key.Binding
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
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select/enter"),
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
		Home: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "first"),
		),
		End: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "last"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

var keys = defaultKeyMap()

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

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))
)

// NewFilePicker creates a new file picker
func NewFilePicker(startPath string) *FilePicker {
	fp := &FilePicker{
		currentPath: startPath,
		showIcons:   true,
		showSizes:   true,
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
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		fp.width = msg.Width
		fp.height = msg.Height
		return fp, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			fp.cancelled = true
			return fp, tea.Quit

		case key.Matches(msg, keys.Escape):
			fp.cancelled = true
			return fp, tea.Quit

		case key.Matches(msg, keys.Up):
			if fp.cursor > 0 {
				fp.cursor--
			}

		case key.Matches(msg, keys.Down):
			if fp.cursor < len(fp.files)-1 {
				fp.cursor++
			}

		case key.Matches(msg, keys.Home):
			fp.cursor = 0

		case key.Matches(msg, keys.End):
			if len(fp.files) > 0 {
				fp.cursor = len(fp.files) - 1
			}

		case key.Matches(msg, keys.Enter):
			if len(fp.files) > 0 {
				selectedFile := fp.files[fp.cursor]
				if selectedFile.IsDir {
					// Enter directory
					if selectedFile.Name == ".." {
						// Go up one directory
						fp.currentPath = filepath.Dir(fp.currentPath)
					} else {
						// Enter the selected directory
						fp.currentPath = selectedFile.Path
					}
					fp.cursor = 0
					fp.loadDirectory()
				} else {
					// Select file
					fp.selected = selectedFile.Path
					return fp, tea.Quit
				}
			}

		case key.Matches(msg, keys.Backspace):
			// Go up one directory
			fp.currentPath = filepath.Dir(fp.currentPath)
			fp.cursor = 0
			fp.loadDirectory()

		case key.Matches(msg, keys.Refresh):
			fp.loadDirectory()
		}
	}

	return fp, nil
}

// View renders the file picker
func (fp *FilePicker) View() string {
	if fp.width == 0 || fp.height == 0 {
		return "Loading..."
	}

	// Calculate content dimensions
	contentWidth := fp.width - 4 // Account for border
	contentHeight := fp.height - 6 // Account for border, title, path, status

	var b strings.Builder

	// Title
	title := titleStyle.Render("File Picker")
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
	status := ""
	if len(fp.files) > 0 {
		selectedFile := fp.files[fp.cursor]
		status = statusStyle.Render(fmt.Sprintf("Selected: %s", selectedFile.Name))
	}
	b.WriteString(status + "\n")

	// Help line
	help := helpStyle.Render("[Enter] Select  [Esc] Cancel  [F5] Refresh  [Backspace] Up Dir")
	b.WriteString(help)

	// Apply border
	content := b.String()
	return borderStyle.Width(contentWidth + 2).Height(fp.height - 2).Render(content)
}

// formatFileEntry formats a single file entry
func (fp *FilePicker) formatFileEntry(file File, selected bool, width int) string {
	var parts []string

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

	// Join parts
	line := strings.Join(parts, " ")

	// Truncate if too long
	if len(line) > width-2 {
		line = line[:width-5] + "..."
	}

	// Pad to full width
	line = fmt.Sprintf("%-*s", width-2, line)

	// Apply styling
	if selected {
		if file.IsDir {
			return selectedStyle.Render(line)
		}
		return selectedStyle.Render(line)
	} else {
		if file.IsDir {
			return dirStyle.Render(line)
		}
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

// GetSelected returns the selected file path
func (fp *FilePicker) GetSelected() (string, bool) {
	if fp.cancelled {
		return "", false
	}
	return fp.selected, fp.selected != ""
}

// GetError returns any error that occurred
func (fp *FilePicker) GetError() error {
	return fp.err
}
