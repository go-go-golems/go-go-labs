# VHS Demo Scripts

This directory contains [VHS](https://github.com/charmbracelet/vhs) scripts to demonstrate the Bubbletea File Manager functionality.

## Prerequisites

Install VHS:
```bash
# macOS
brew install vhs

# Other platforms
go install github.com/charmbracelet/vhs@latest
```

## Quick Start

Run all demos at once:
```bash
./demo/run-all-demos.sh
```

Or run individual demos:
```bash
# Build the app first
go build -o filepicker .

# Setup test environment
./demo/setup-test-env.sh

# Run specific demo
vhs demo/basic-navigation.tape
```

## Available Demos

### Core Navigation
- **`basic-navigation.tape`** - Basic file picker launching and navigation
- **`file-icons.tape`** - File type icons and size display demonstration

### Multi-Selection Features
- **`multi-selection.tape`** - Multi-selection with Space key, select all/deselect all
- **`copy-paste.tape`** - File copy and paste operations
- **`cut-paste.tape`** - File cut (move) and paste operations

### File Operations
- **`delete-confirm.tape`** - Delete operation with confirmation dialog
- **`create-files.tape`** - Creating new files and directories
- **`rename-file.tape`** - Renaming files and directories

### UI Features
- **`help-system.tape`** - Built-in help system demonstration
- **`overview.tape`** - Quick overview of main features

## Generated Output

Each script generates a GIF file in the `demo/` directory:
- `demo/basic-navigation.gif`
- `demo/file-icons.gif`
- `demo/multi-selection.gif`
- `demo/copy-paste.gif`
- `demo/cut-paste.gif`
- `demo/delete-confirm.gif`
- `demo/create-files.gif`
- `demo/rename-file.gif`
- `demo/help-system.gif`
- `demo/overview.gif`

## Using the GIFs

These GIFs are perfect for:
- README documentation
- GitHub wiki pages
- Blog posts and tutorials
- Feature demonstrations

## Customization

To modify a demo:
1. Edit the corresponding `.tape` file
2. Adjust timing with `Sleep` commands
3. Change dimensions with `Set Width` and `Set Height`
4. Modify theme with `Set Theme`

Example VHS script structure:
```tape
# Setup
Output demo/my-demo.gif
Set FontSize 14
Set Width 800
Set Height 600
Set Theme "Dracula"

# Commands
Type "./filepicker test-files"
Enter
Sleep 1s
Down Down
Space
Sleep 500ms
Escape
```

## Tips

- Use `Sleep` commands to control timing between actions
- Keep demos short (10-15 seconds) for README usage
- Use consistent themes across all demos
- Test scripts multiple times to ensure smooth playback
