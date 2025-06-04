# Bubbletea File Picker

A Terminal User Interface (TUI) file picker built with [Charm's Bubbletea](https://github.com/charmbracelet/bubbletea) framework.

This implementation includes Tier 1 and Tier 2 features from the file picker specification:

## Features

### Tier 1 - Basic File Selection
- ✅ Simple file picker for basic selection tasks
- ✅ Directory path display at the top
- ✅ File list with highlighting
- ✅ Basic navigation controls (↑/↓, Enter, Esc)
- ✅ File selection and cancellation

### Tier 2 - Enhanced Navigation  
- ✅ File type icons (📁📄🖼️📦⚙️💻)
- ✅ File sizes in human-readable format
- ✅ Parent directory (..) navigation
- ✅ Directory traversal with Enter key
- ✅ Backspace to go up one directory
- ✅ F5 to refresh current directory
- ✅ Home/End to jump to first/last item
- ✅ Status bar showing current selection

## Controls

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Enter` | Select file or enter directory |
| `Esc` | Cancel and exit |
| `Backspace` | Go up one directory |
| `F5` | Refresh current directory |
| `Home` | Jump to first item |
| `End` | Jump to last item |
| `q` / `Ctrl+C` | Quit |

## Usage

```bash
# Run from current directory
go run . 

# Run from specific directory
go run . /path/to/directory

# Build and run
go build -o filepicker .
./filepicker /home/user/documents
```

## File Type Icons

- 📁 Directories
- 📄 Text files (.txt, .md, etc.)
- 🖼️ Images (.jpg, .png, .gif, etc.)
- 📦 Archives (.zip, .tar, .gz, etc.)
- ⚙️ Executables (.sh, .exe, .bat, etc.)
- 💻 Code files (.go, .py, .js, etc.)

## Visual Layout

```
┌─ File Picker ────────────────────────────────────────┐
│ Path: /home/user/documents                           │
├──────────────────────────────────────────────────────┤
│ 📁 ..                                                │
│ 📁 projects                                          │
│ 📄 document.txt                          2.3 KB      │
│ 📄 readme.md                             1.1 KB      │
│ 🖼️ photo.jpg                             890 KB      │
│                                                      │
├──────────────────────────────────────────────────────┤
│ Selected: document.txt                               │
│ [Enter] Select  [Esc] Cancel  [F5] Refresh          │
└──────────────────────────────────────────────────────┘
```

## Dependencies

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - Key bindings

## Implementation Notes

This implementation uses the standard Go filesystem APIs and supports:
- Cross-platform file system navigation
- Proper error handling for file system operations
- Responsive terminal resizing
- Scrolling for large directories
- Keyboard-only navigation
