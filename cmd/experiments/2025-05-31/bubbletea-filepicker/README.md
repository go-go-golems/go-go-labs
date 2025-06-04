# Bubbletea File Manager

A powerful Terminal User Interface (TUI) file manager built with [Charm's Bubbletea](https://github.com/charmbracelet/bubbletea) framework.

This implementation includes **Tier 1, Tier 2, Tier 3, and Tier 4** features from the comprehensive file picker specification, providing a full-featured file management experience in the terminal.

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

### Tier 3 - Multi-Selection & File Operations
- ✅ Multi-selection with Space key
- ✅ Visual indicators (✓ for selected, ▶ for cursor, ✓▶ for both)
- ✅ Select all (a), deselect all (A), select all files (Ctrl+A)
- ✅ File operations: delete (d), copy (c), cut (x), paste (v)
- ✅ Rename files and directories (r)
- ✅ Create new files (n) and directories (m)
- ✅ Confirmation dialogs for destructive operations
- ✅ Built-in help system with ? key
- ✅ Clipboard operations with visual feedback

### Tier 4 - Advanced Interface & Preview
- ✅ Dual-panel layout with file list and preview
- ✅ Tab to toggle preview panel on/off
- ✅ File content preview for text files
- ✅ File properties display (size, permissions, timestamps)
- ✅ Search functionality with / key and real-time filtering
- ✅ Advanced display options (F2 hidden files, F3 detailed view, F4 sort)
- ✅ Multiple sort modes (name, size, date, type)
- ✅ Extended file type icons and detection
- ✅ Hidden file support with visual indicators
- ✅ Directory content preview

## Controls

### Navigation
| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Home` | Jump to first item |
| `End` | Jump to last item |
| `Backspace` | Go up one directory |

### Selection
| Key | Action |
|-----|--------|
| `Enter` | Select file(s) or enter directory |
| `Space` | Toggle selection on current item |
| `a` | Select all items |
| `A` | Deselect all items |
| `Ctrl+A` | Select all files (not directories) |

### File Operations
| Key | Action |
|-----|--------|
| `d` | Delete selected files (with confirmation) |
| `c` | Copy selected files to clipboard |
| `x` | Cut selected files to clipboard |
| `v` | Paste files from clipboard |
| `r` | Rename current file/directory |
| `n` | Create new file |
| `m` | Create new directory |

### Advanced Features (Tier 4)
| Key | Action |
|-----|--------|
| `Tab` | Toggle preview panel |
| `/` | Enter search mode |
| `F2` | Toggle hidden files |
| `F3` | Toggle detailed view |
| `F4` | Cycle sort mode (name/size/date/type) |

### System
| Key | Action |
|-----|--------|
| `F5` | Refresh current directory |
| `?` | Toggle help display |
| `Esc` | Cancel current operation or exit |
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

## Visual Layout

### Tier 4 - Dual Panel with Preview
```
┌─ File Explorer ─────────────────────────────────┬─ Preview ─────────┐
│ Path: /home/user/documents           (search: doc) │ document.txt      │
├─────────────────────────────────────────────────┤ Size: 2.3 KB     │
│   📁 ..                                         │ Modified: Jan 15  │
│ ✓ 📁 projects                    Jan 20  drwxr- │ Permissions: -rw- │
│ ▶ 📄 document.txt     2.3 KB     Jan 15  -rw-r- │ ─────────────────  │
│ ✓ 📄 readme.md       1.1 KB     Jan 12  -rw-r- │ This is a sample  │
│   🖼️ photo.jpg       890 KB     Jan 08  -rw-r- │ text document     │
│                                                 │ with some content │
│ Search: [readme________]                        │ for demonstration │
├─────────────────────────────────────────────────┤ purposes.         │
│ 3 of 45 items | 2 selected | Sort: Name | Details,Preview     │                   
│ [Tab] Toggle Preview  [/] Search  [F2] Hidden  [F4] Sort       │
└─────────────────────────────────────────────────────────────────┘
```

### Multi-Selection Indicators
- `▶` - Current cursor position
- `✓` - Multi-selected item
- `✓▶` - Both selected and current cursor position

### File Type Icons (Extended)
- 📁 Directories / 👻 Hidden directories
- 📄 Text files / 📋 Documents / 📊 Spreadsheets / ▶️ Presentations  
- 🖼️ Images / 🎬 Videos / 🎵 Audio
- 📦 Archives / ⚙️ Executables / 💻 Code files
- 🔗 Symlinks / 🔒 Read-only / ❓ Unknown

## Key Features in Detail

### Search Functionality
- Press `/` to enter search mode
- Type to filter files in real-time
- Search matches file names (case-insensitive)
- Clear search with Esc or empty search term
- Search results counter in status bar

### Preview Panel
- Toggle with `Tab` key
- Shows file content for text files (up to 15 lines)
- Displays file properties (size, permissions, modification time)
- Directory previews show item counts
- Handles binary files gracefully

### Advanced Display Options
- `F2` toggles hidden file visibility
- `F3` switches between simple and detailed view
- `F4` cycles through sort modes:
  - Name (alphabetical)
  - Size (smallest to largest)
  - Date (newest first)
  - Type (by file extension)

### File Operations
- Copy/cut operations work across directories
- Visual feedback for clipboard operations
- Recursive directory copying
- Confirmation dialogs prevent accidental deletions
- Real-time input for rename and create operations

## Dependencies

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - Components (help, textinput)

## Demo Videos

The `demo/` directory contains VHS scripts to generate demonstration GIFs:

```bash
# Setup test environment and run all demos
./demo/run-all-demos.sh

# Or run individual demos
vhs demo/basic-navigation.tape
vhs demo/multi-selection.tape
vhs demo/preview-panel.tape
```

Available demos:
- `basic-navigation.gif` - Basic file picker usage
- `file-icons.gif` - File type icons demonstration
- `multi-selection.gif` - Multi-selection features
- `copy-paste.gif` - File copy operations
- `cut-paste.gif` - File move operations
- `delete-confirm.gif` - Delete with confirmation
- `create-files.gif` - Creating files and directories
- `rename-file.gif` - Renaming operations
- `help-system.gif` - Built-in help system
- `preview-panel.gif` - Preview panel features

## Implementation Notes

This implementation uses the standard Go filesystem APIs and supports:
- Cross-platform file system navigation
- Proper error handling for file system operations
- Responsive terminal resizing
- Scrolling for large directories
- Keyboard-only navigation
- Unicode file names and content
- Permission-aware operations

The file manager follows modern TUI design principles:
- Intuitive keyboard shortcuts
- Visual feedback for all operations
- Non-destructive operations by default
- Comprehensive help system
- Consistent styling and theming

## Architecture

The codebase is organized into clear functional areas:
- **Core Model**: File picker state and data structures
- **Key Bindings**: Comprehensive keyboard shortcut system
- **View Rendering**: Dual-panel layout with responsive design
- **File Operations**: Copy, move, delete, create operations
- **Search & Filter**: Real-time file filtering
- **Preview System**: Text file content and metadata display

This makes the code maintainable and extensible for additional features.
