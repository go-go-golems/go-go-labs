# AGENT.md - Kitty Configuration Repository

## Project Structure
- **Main config**: `kitty.conf` - Primary kitty terminal configuration
- **Theme config**: `current-theme.conf` - Active color theme (JetBrains Darcula)
- **Theme repository**: `kitty-themes/` - Git submodule with 180+ terminal themes
- **Backup**: `kitty.conf.bak` - Configuration backup

## Commands
- **Test theme**: `kitty @ set-colors -a "~/.config/kitty/kitty-themes/themes/ThemeName.conf"`
- **Preview theme**: `kitty -o include="~/.config/kitty/kitty-themes/themes/ThemeName.conf"`
- **Reload config**: `kitty @ load-config`
- **Remote control**: Uses unix socket at `/tmp/kitty-socket`

## Configuration Style
- **Format**: Standard kitty.conf syntax (`key value`)
- **Comments**: Use `#` for comments
- **Colors**: Hex format (`#rrggbb`)
- **Includes**: Use `include` directive for theme files
- **Font**: BerkeleyMonoVariable Nerd Font, size 9
- **Theme structure**: JetBrains Darcula with 16-color palette

## Important Settings
- Remote control enabled on unix socket
- Scrollback: 8000 lines, 100MB buffer
- Tab style: powerline
- Custom keybindings for tab navigation (Ctrl+Shift+1-9)
