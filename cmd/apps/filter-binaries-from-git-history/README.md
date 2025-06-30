# Filter Binaries from Git History

An interactive TUI tool built with [Charm's Bubbletea](https://github.com/charmbracelet/bubbletea) to analyze git history, identify large binary files, and selectively remove them from git history.

## Features

- 🔍 **Interactive Analysis**: Analyze differences between git references (default: origin/main vs main)
- 📊 **Statistics Display**: Shows file counts, sizes, and identifies likely binary files
- 🎯 **Smart Detection**: Identifies binary files by extension and size threshold
- ✅ **Selective Removal**: Choose which files to remove with an interactive checkbox interface
- 🛡️ **Safety First**: Confirmation step before performing irreversible operations
- 📦 **Git Integration**: Uses git filter-branch for history rewriting

## Usage

```bash
# Basic usage - analyze current branch vs origin/main
./filter-binaries-from-git-history

# Custom references
./filter-binaries-from-git-history --base HEAD~10 --compare HEAD

# Adjust size threshold (default 1MB)
./filter-binaries-from-git-history --size-threshold 5242880  # 5MB

# Enable debug logging
./filter-binaries-from-git-history --log-level debug
```

## Navigation

### Stats View
- **Enter**: Continue to file selection
- **q**: Quit

### File Selection View
- **↑/↓** or **j/k**: Navigate through files
- **Space**: Toggle file selection
- **a**: Select all files
- **n**: Deselect all files
- **Enter**: Continue to confirmation
- **q**: Quit

### Confirmation View
- **Enter**: Proceed with history rewrite
- **q**: Quit

## Warning

⚠️ **This tool modifies git history permanently!**

- Always backup your repository before using
- Force push will be required after history rewrite: `git push --force-with-lease`
- Coordinate with team members before rewriting shared history

## Installation

Built as part of the go-go-labs project. Build with:

```bash
go build -o filter-binaries-from-git-history cmd/apps/filter-binaries-from-git-history/main.go
```

## Architecture

- `cmd/`: Cobra CLI command structure
- `pkg/analyzer/`: File analysis and statistics
- `pkg/git/`: Git repository operations and history rewriting
- `pkg/ui/`: Bubbletea TUI components and state management
