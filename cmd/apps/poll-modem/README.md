# Poll Modem TUI

A Terminal User Interface (TUI) application that polls a cable modem's network setup page and displays the channel information in a beautiful, real-time table format.

## Features

- **Real-time monitoring**: Continuously polls the modem endpoint at configurable intervals
- **Multiple views**: Switch between overview, downstream channels, upstream channels, and error statistics
- **Beautiful UI**: Uses charmbracelet's bubbletea framework for a modern TUI experience
- **Configurable**: Customize the modem URL and polling interval
- **Error handling**: Graceful error handling with visual feedback

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
./poll-modem --url http://192.168.1.1/network_setup.jst
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
./poll-modem --url http://192.168.1.1/network_setup.jst --interval 45s --debug
```

## Key Bindings

- **Tab / →**: Switch to next view
- **Shift+Tab / ←**: Switch to previous view  
- **r**: Manually refresh data
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

## Configuration

The application uses the following default settings:

- **URL**: `http://192.168.0.1/network_setup.jst`
- **Poll Interval**: 30 seconds
- **Debug**: Disabled

## Requirements

- Go 1.24.3 or later
- Network access to the cable modem
- Terminal with color support for best experience

## Supported Modems

This application is designed to work with cable modems that provide network information in the HTML format similar to the Technicolor CGM4331COM. The HTML parser looks for specific CSS classes and table structures commonly used in cable modem web interfaces.

## Troubleshooting

### Connection Issues
- Ensure you can access the modem's web interface in a browser
- Check if the URL is correct for your modem model
- Verify network connectivity to the modem

### Parsing Issues  
- Enable debug logging with `--debug` to see detailed error information
- Different modem models may use different HTML structures
- Check if the modem requires authentication (cookies/sessions)

### Performance
- Increase the polling interval if the modem responds slowly
- Use `--interval 2m` for less frequent updates

## Example Output

```
┌─ Cable Modem Monitor ─┐

✓ Last updated: 14:32:15

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