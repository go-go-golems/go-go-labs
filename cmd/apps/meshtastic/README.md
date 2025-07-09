# Meshtastic TUI

A terminal user interface for interacting with Meshtastic devices via serial connection.

## Features

- **Command Line Interface**: Debug and test protocol with CLI commands
- **Terminal User Interface**: Interactive TUI for real-time device management
- **Auto-discovery**: Automatically finds Meshtastic devices on common serial ports
- **Real-time Communication**: Send and receive messages in real-time
- **Node Management**: View connected nodes and their information
- **Device Status**: Monitor device health and configuration

## Installation

```bash
go mod download
make proto-gen  # Generate protobuf files
go build ./cmd/meshtastic-tui
```

## Usage

### CLI Commands

```bash
# Show device information
./meshtastic-tui info

# Send a text message
./meshtastic-tui send "Hello, Meshtastic!"

# Listen for incoming messages
./meshtastic-tui listen

# Launch the TUI interface
./meshtastic-tui tui
```

### Command Line Options

- `--port`: Specify serial port (default: `/dev/ttyUSB0`)
- `--log-level`: Set logging level (debug, info, warn, error)
- `--timeout`: Connection timeout duration

### TUI Interface

The TUI provides a tabbed interface with:

1. **Messages**: View incoming and outgoing messages
2. **Nodes**: See connected nodes and their status
3. **Status**: Monitor device information and health
4. **Compose**: Send messages to specific nodes or broadcast

#### Key Bindings

- `Tab/Shift+Tab`: Navigate between tabs
- `Enter`: Send message (in compose mode)
- `Esc`: Exit compose mode
- `q/Ctrl+C`: Quit application
- `?`: Show help

## Development

### Project Structure

Following [bobatea guidelines](./bobatea/docs/charmbracelet-bubbletea-guidelines.md):

```
cmd/meshtastic-tui/     # CLI entry point
pkg/
  ├── client/           # High-level Meshtastic client
  ├── serial/           # Serial communication layer
  ├── protocol/         # Binary protocol framing
  ├── pb/              # Generated protobuf files
  └── ui/              # TUI components
      ├── model/       # Bubble Tea models
      ├── view/        # Lipgloss styles
      ├── keys/        # Key bindings
      └── bubbles/     # Custom bubble components
```

### Building

```bash
# Build all packages
go build ./...

# Generate protobuf files
make proto-gen

# Run tests
go test ./...

# Format code
go fmt ./...
```

### Testing

To test without hardware:

```bash
# Test with debug logging
./meshtastic-tui --log-level debug info

# Test TUI (will warn about missing device but still launch)
./meshtastic-tui tui
```

## Device Connection

The application automatically discovers Meshtastic devices on common serial ports:

- `/dev/ttyACM0`, `/dev/ttyACM1` (Linux)
- `/dev/ttyUSB0`, `/dev/ttyUSB1` (Linux)
- `/dev/cu.usbmodem*` (macOS)
- `/dev/cu.usbserial*` (macOS)

### Supported Devices

- RAK4631 (VID: 0x239a)
- ESP32-based devices (VID: 0x303a)
- FTDI devices (VID: 0x0403)
- Silicon Labs CP2102 (VID: 0x10c4)

## Protocol

The application implements the Meshtastic binary protocol:

- **Framing**: START1 (0x94) + START2 (0xC3) + LENGTH + PAYLOAD
- **Encoding**: Protocol Buffers
- **Baud Rate**: 115200
- **Max Payload**: 512 bytes

### Message Types

- Text messages (TEXT_MESSAGE_APP)
- Node information (NODEINFO_APP)
- Position data (POSITION_APP)
- Telemetry (TELEMETRY_APP)
- Device administration (ADMIN_APP)

## Contributing

1. Follow the bobatea guidelines for TUI development
2. Use the existing protobuf definitions
3. Implement proper error handling
4. Add tests for new features
5. Update documentation

## License

This project is licensed under the MIT License.
