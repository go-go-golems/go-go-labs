# Meshtastic CLI

A comprehensive command-line interface for interacting with Meshtastic devices via serial, TCP, and BLE connections.

## Features

- **Multiple Connection Types**: Serial, TCP/IP, and BLE support
- **Device Auto-Discovery**: Automatically finds connected Meshtastic devices
- **Comprehensive Command Set**: Complete device management and messaging
- **Output Formats**: JSON, YAML, and table formats
- **Robust Error Handling**: Detailed error messages and recovery
- **Real-time Operations**: Live monitoring and message listening
- **Network Diagnostics**: Ping and traceroute functionality

## Installation

### Prerequisites

- Go 1.23 or later
- A Meshtastic device (serial, TCP, or BLE)

### Building from Source

```bash
cd go-go-labs/cmd/apps/meshtastic
go build -o meshtastic-cli
```

## Usage

### Device Connection & Discovery

```bash
# Auto-discover devices
./meshtastic-cli discover

# Discover only serial devices
./meshtastic-cli discover --serial-only

# Connect to device
./meshtastic-cli connect --port /dev/ttyUSB0

# Connect via TCP/IP
./meshtastic-cli connect --host 192.168.1.100

# Get device information
./meshtastic-cli info
./meshtastic-cli info --json
```

### Node Management

```bash
# List all nodes
./meshtastic-cli nodes

# Show specific fields
./meshtastic-cli nodes --show-fields id,user,snr,distance

# Sort by SNR
./meshtastic-cli nodes --sort-by snr

# Live updating display
./meshtastic-cli nodes --live

# JSON output
./meshtastic-cli nodes --json
```

### Configuration Management

```bash
# Get all configuration
./meshtastic-cli config get --all

# Get specific field
./meshtastic-cli config get device.role

# Set configuration
./meshtastic-cli config set device.role CLIENT

# Export configuration
./meshtastic-cli config export --output config.yaml

# Import configuration
./meshtastic-cli config import --file config.yaml
```

### Channel Management

```bash
# List channels
./meshtastic-cli channel list

# Show encryption keys
./meshtastic-cli channel list --show-keys

# Add new channel
./meshtastic-cli channel add "Private Channel"

# Set channel with PSK
./meshtastic-cli channel add "Secure" --psk "base64key"

# Delete channel
./meshtastic-cli channel delete --index 2

# Enable/disable channel
./meshtastic-cli channel enable --index 1
./meshtastic-cli channel disable --index 1
```

### Messaging

```bash
# Send broadcast message
./meshtastic-cli message send "Hello, Meshtastic!"

# Send to specific node
./meshtastic-cli message send "Private message" --dest !a4c138f4

# Send on specific channel
./meshtastic-cli message send "Channel msg" --channel 1

# Request acknowledgment
./meshtastic-cli message send "Important" --want-ack

# Listen for messages
./meshtastic-cli message listen

# Listen on specific channel
./meshtastic-cli message listen --channel 1

# Listen with JSON output
./meshtastic-cli message listen --json

# Send private message
./meshtastic-cli message private !a4c138f4 "Secret message"

# Reply to last message
./meshtastic-cli message reply "Thanks!"
```

### Position Management

```bash
# Get current position
./meshtastic-cli position get

# Set fixed position
./meshtastic-cli position set --lat 37.7749 --lon -122.4194

# Set with altitude
./meshtastic-cli position set --lat 37.7749 --lon -122.4194 --alt 100

# Clear fixed position
./meshtastic-cli position clear

# Request position from node
./meshtastic-cli position request !a4c138f4

# Broadcast current position
./meshtastic-cli position broadcast
```

### Device Management

```bash
# Reboot device
./meshtastic-cli device reboot

# Shutdown device
./meshtastic-cli device shutdown

# Factory reset
./meshtastic-cli device factory-reset

# Set device owner
./meshtastic-cli device set-owner "John Doe"

# Set device time
./meshtastic-cli device set-time

# Get device metadata
./meshtastic-cli device metadata
```

### Telemetry & Monitoring

```bash
# Get telemetry data
./meshtastic-cli telemetry get

# Get specific telemetry type
./meshtastic-cli telemetry get --type device

# Request telemetry from node
./meshtastic-cli telemetry request !a4c138f4

# Monitor telemetry in real-time
./meshtastic-cli telemetry monitor

# Monitor with interval
./meshtastic-cli telemetry monitor --interval 30s
```

### Network Diagnostics

```bash
# Ping a node
./meshtastic-cli ping !a4c138f4

# Ping with custom count
./meshtastic-cli ping !a4c138f4 --count 10

# Traceroute to node
./meshtastic-cli traceroute !a4c138f4

# Traceroute with max hops
./meshtastic-cli traceroute !a4c138f4 --max-hops 5
```

### TUI Interface

```bash
# Launch interactive TUI
./meshtastic-cli tui

# TUI with custom settings
./meshtastic-cli tui --port /dev/ttyACM0 --log-level debug
```

## Global Options

- `--port`, `-p`: Serial port for Meshtastic device (default: `/dev/ttyUSB0`)
- `--host`: TCP/IP host for network connection
- `--timeout`: Operation timeout (default: `10s`)
- `--log-level`: Log level (debug, info, warn, error) (default: `info`)
- `--debug-serial`: Enable verbose serial communication logging
- `--hex-dump`: Enable hex dump logging of raw serial data

## Output Formats

Most commands support multiple output formats:

```bash
# Default table format
./meshtastic-cli nodes

# JSON format
./meshtastic-cli nodes --json

# YAML format (where supported)
./meshtastic-cli config get --yaml
```

## Examples

### Basic Device Setup

```bash
# Discover and connect to device
./meshtastic-cli discover
./meshtastic-cli info

# Set device owner
./meshtastic-cli device set-owner "MyCallsign"

# Configure device role
./meshtastic-cli config set device.role ROUTER
```

### Channel Configuration

```bash
# Create a private channel
./meshtastic-cli channel add "Private" --psk "your-base64-key"

# List all channels
./meshtastic-cli channel list

# Send message on specific channel
./meshtastic-cli message send "Hello private channel" --channel 1
```

### Monitoring Setup

```bash
# Start message listening in one terminal
./meshtastic-cli message listen --json

# Monitor telemetry in another terminal
./meshtastic-cli telemetry monitor --interval 60s

# Watch nodes in real-time
./meshtastic-cli nodes --live
```

### Network Diagnostics

```bash
# Test connectivity to all nodes
./meshtastic-cli nodes --json | jq -r '.[].id' | while read node; do
  echo "Testing $node..."
  ./meshtastic-cli ping "$node" --count 3
done
```

## Troubleshooting

### Common Issues

#### Device Not Found

```bash
# Check available ports
./meshtastic-cli discover --serial-only

# Try specific port
./meshtastic-cli info --port /dev/ttyACM0

# Enable debug logging
./meshtastic-cli info --log-level debug
```

#### Connection Timeout

```bash
# Increase timeout
./meshtastic-cli info --timeout 30s

# Check device is not in use
lsof /dev/ttyACM0
```

#### Permission Denied

```bash
# Add user to dialout group
sudo usermod -a -G dialout $USER

# Logout and login again
```

### Debug Information

Enable detailed logging for troubleshooting:

```bash
# Enable debug logging
./meshtastic-cli info --log-level debug

# Enable serial debugging
./meshtastic-cli info --debug-serial

# Enable hex dump
./meshtastic-cli info --hex-dump
```

## Performance

The CLI is optimized for reliability and ease of use:

- **Connection Setup**: 2-5 seconds with auto-discovery
- **Command Latency**: 100-500ms depending on device
- **Memory Usage**: ~10MB typical, ~20MB with debug logging
- **Concurrent Operations**: Supports multiple simultaneous commands

## Architecture

The application uses a layered architecture:

```
Meshtastic CLI
├── Commands Layer       # Cobra CLI commands
├── Client Layer        # Robust client with retry logic
├── Protocol Layer      # Meshtastic protocol handling
├── Transport Layer     # Serial/TCP/BLE communication
└── Device Layer        # Hardware abstraction
```

## Development

### Building

```bash
# Build the application
go build -o meshtastic-cli

# Build with race detection
go build -race -o meshtastic-cli

# Run tests
go test ./...
```

### Testing

```bash
# Run all tests
go test ./...

# Test specific package
go test ./pkg/client

# Test with race detection
go test -race ./...
```

## License

This project is part of the go-go-labs repository and follows the same licensing terms.
