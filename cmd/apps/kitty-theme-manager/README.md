# Kitty Theme Manager

A retro 80s POS-style terminal UI for managing kitty terminal themes with Charm's Bubble Tea.

## Features

- 🎨 **Browse 180+ themes** from the kitty-themes collection
- 👀 **Live preview** themes in your current terminal
- 🚀 **Launch new kitty instances** with specific themes
- 💾 **Apply themes** with automatic backup system (.001, .002, etc.)
- ⏪ **Rollback support** to previous themes
- 🖥️ **80s POS-style interface** with green-on-black terminal aesthetic
- ⌨️ **Vim-like navigation** (j/k, arrows, home/end)

## Installation

```bash
# Build the binary
go build -o kitty-theme-manager

# Make it executable
chmod +x kitty-theme-manager

# Optional: Install to PATH
sudo mv kitty-theme-manager /usr/local/bin/
```

## Requirements

- Go 1.21+
- Kitty terminal with remote control enabled
- kitty-themes repository cloned to `~/.config/kitty/kitty-themes/`

## Setup

1. Enable kitty remote control by adding to your `kitty.conf`:
   ```
   allow_remote_control yes
   listen_on unix:/tmp/kitty-socket
   ```

2. Clone the themes repository:
   ```bash
   git clone --depth 1 https://github.com/dexpota/kitty-themes.git ~/.config/kitty/kitty-themes
   ```

## Usage

```bash
./kitty-theme-manager
```

### Controls

- **↑/↓** or **j/k** - Navigate theme list
- **Home/End** - Jump to first/last theme
- **Enter** - Apply theme to current-theme.conf (creates backup)
- **P** - Preview theme in current terminal (temporary)
- **L** - Launch new kitty window with selected theme
- **B** - Rollback to previous theme from backup
- **H** or **?** - Toggle help
- **Q** or **Ctrl+C** - Quit

## Backup System

When you apply a theme, the current `current-theme.conf` is automatically backed up as:
- `current-theme.conf.001`
- `current-theme.conf.002`
- `current-theme.conf.003`
- etc.

Use **B** to rollback to the most recent backup.

## Theme Preview

The interface shows:
- Theme name and background color
- 8-color palette preview
- Live color values (background, foreground, cursor)

## Architecture

- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling
- Parses kitty theme files (.conf format)
- Uses kitty's remote control API for live updates

## License

MIT
