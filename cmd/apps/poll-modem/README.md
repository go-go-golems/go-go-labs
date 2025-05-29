# Poll Modem TUI

A Terminal User Interface (TUI) application that polls a cable modem's network setup page and displays the channel information in a beautiful, real-time table format with persistent SQLite storage.

## Features

- **Real-time monitoring**: Continuously polls the modem endpoint at configurable intervals
- **Multiple views**: Switch between overview, downstream channels, upstream channels, and error statistics
- **Beautiful UI**: Uses charmbracelet's bubbletea framework for a modern TUI experience
- **Persistent storage**: All data is automatically stored in SQLite database (`~/.config/poll-modem/history.db`)
- **Flexible export**: Export current data, current session, or entire history to CSV
- **History view**: Browse historical data with channel-specific filtering
- **Session management**: Each application run creates a new session for data organization
- **Configurable**: Customize the modem URL and polling interval
- **Authentication**: Support for modem login with username/password
- **Error handling**: Graceful error handling with visual feedback and automatic re-authentication

## Installation

From the project root:

```bash
go build -o poll-modem ./cmd/apps/poll-modem
```

## Usage

### Basic usage with default settings:
```bash
./poll-modem
```

### Custom modem URL:
```bash
./poll-modem --url http://192.168.1.1
```

### With authentication:
```bash
./poll-modem --username admin --password yourpassword
```

### Custom polling interval:
```bash
./poll-modem --interval 1m
```

### Enable debug logging:
```bash
./poll-modem --debug
```

### All options:
```bash
./poll-modem --url http://192.168.1.1 --username admin --password yourpassword --interval 45s --debug
```

## Key Bindings

- **Tab / →**: Switch to next view
- **Shift+Tab / ←**: Switch to previous view  
- **r**: Manually refresh data
- **h**: Toggle between current data and history view
- **e**: Show export menu
- **1, 2, 3**: Select export mode (when export menu is shown)
- **q / Ctrl+C**: Quit application
- **?**: Show help

## Views

### 1. Overview
- Cable modem hardware information (model, vendor, versions, etc.)
- Channel summary statistics
- Lock status overview

### 2. Downstream Channels
- Channel ID, lock status, frequency
- Signal-to-Noise Ratio (SNR)
- Power levels and modulation

### 3. Upstream Channels  
- Channel ID, lock status, frequency
- Symbol rates, power levels
- Modulation and channel types

### 4. Error Codewords
- Per-channel error statistics
- Unerrored, correctable, and uncorrectable codewords

## Data Storage

All modem data is automatically stored in a SQLite database located at:
```
~/.config/poll-modem/history.db
```

The database includes:
- **Sessions**: Each application run creates a new session
- **Cable modem info**: Hardware details for each reading
- **Channel data**: Downstream, upstream, and error statistics
- **Timestamps**: All data is timestamped for historical analysis

## Export Options

Press **'e'** to show the export menu, then select:

1. **Export Current Data Only**: Latest reading from current session
2. **Export Current Session**: All data from the current application run
3. **Export All History**: Complete historical data from all sessions

Exported CSV files include:
- Session ID for data organization
- Timestamps for all readings
- Complete channel information
- Separate files for downstream, upstream, and error data

## Configuration

The application uses the following default settings:

- **URL**: `http://192.168.0.1`
- **Poll Interval**: 30 seconds
- **Debug**: Disabled
- **Export Mode**: Current Session

## Data Files

- **Database**: `~/.config/poll-modem/history.db` (SQLite database)
- **Cookies**: `~/.config/poll-modem/cookies.json` (Authentication cookies)
- **Exports**: CSV files saved in current directory

## Requirements

- Go 1.24.3 or later
- SQLite3 (automatically included via go-sqlite3 driver)
- Network access to the cable modem
- Terminal with color support for best experience

## Supported Modems

This application is designed to work with cable modems that provide network information in the HTML format similar to the Technicolor CGM4331COM. The HTML parser looks for specific CSS classes and table structures commonly used in cable modem web interfaces.

## Troubleshooting

### Connection Issues
- Ensure you can access the modem's web interface in a browser
- Check if the URL is correct for your modem model
- Verify network connectivity to the modem

### Authentication Issues
- Use `--username` and `--password` flags if your modem requires login
- Check credentials if you see "authentication failed" errors
- The app will automatically re-authenticate when sessions expire

### Database Issues
- Database files are created automatically in `~/.config/poll-modem/`
- Check disk space if you encounter database errors
- Database corruption is rare but can be fixed by deleting the database file

### Parsing Issues  
- Enable debug logging with `--debug` to see detailed error information
- Different modem models may use different HTML structures
- Check if the modem requires authentication (cookies/sessions)

### Performance
- Increase the polling interval if the modem responds slowly
- Use `--interval 2m` for less frequent updates
- Large history databases may slow down exports

## Example Output

```
┌─ Cable Modem Monitor ─┐

✓ Last updated: 14:32:15 | Exported 3 files (session mode)

Cable Modem Information

Model: CGM4331COM (XB7)
Vendor: Technicolor  
HW Version: 2.1
Core Version: 1.0
BOOT Version: S1TC-3.63.20.104
Download Version: Prod_23.2_231009 & Prod_23.2_231009
Flash Part: 4096 MB

Channel Summary

Downstream: 34/34 channels locked
Upstream: 5/5 channels locked
```

## Data Analysis

The SQLite database can be queried directly for advanced analysis:

```sql
-- View all sessions
SELECT * FROM sessions;

-- Channel performance over time
SELECT timestamp, channel_id, snr, power_level 
FROM downstream_channels 
WHERE channel_id = '1' 
ORDER BY timestamp;

-- Error trends
SELECT DATE(timestamp) as date, 
       SUM(CAST(uncorrectable_codewords AS INTEGER)) as total_errors
FROM error_channels 
GROUP BY DATE(timestamp);
``` 