# Messaging and Position Management Implementation

## Overview

This implementation adds comprehensive messaging and position management CLI commands to the Meshtastic application, following the specification document requirements.

## Implemented Commands

### Messaging Commands (`cmd/message.go`)

#### `message send [TEXT]`
- Send text messages to mesh network or specific destinations
- Flags: `--dest`, `--channel`, `--want-ack`, `--hop-limit`
- Supports broadcast and direct messaging
- Destination parsing for node IDs, hex format, and broadcast

#### `message listen`
- Listen for incoming messages with real-time display
- Flags: `--channel`, `--from`, `--timeout`, `--json`
- Message filtering by channel and sender
- Graceful shutdown with Ctrl+C
- Tracks last message for reply functionality

#### `message reply [TEXT]`
- Reply to the last received message
- Flags: `--want-ack`
- Automatically uses correct destination and channel

#### `message private [TEXT]`
- Send private messages to specific destinations
- Flags: `--dest` (required), `--want-ack`
- Forces channel 0 for private messaging

### Position Management Commands (`cmd/position.go`)

#### `position get`
- Get current GPS position from device
- Displays latitude, longitude, altitude, source, and precision
- Converts coordinates to human-readable format with direction indicators

#### `position set --lat LAT --lon LON [--alt ALT]`
- Set fixed GPS position
- Flags: `--lat`, `--lon`, `--alt`, `--fixed`
- Coordinate validation (-90 to 90 lat, -180 to 180 lon)
- Uses AdminMessage for device configuration

#### `position clear`
- Clear fixed position and return to GPS mode
- Uses AdminMessage to remove fixed position

#### `position request [--dest NODE | --all]`
- Request position from other nodes
- Flags: `--dest`, `--all`, `--timeout`
- Waits for responses and displays position information

#### `position broadcast`
- Broadcast current position to mesh network
- Retrieves current position and sends to all nodes

## Technical Implementation

### Key Features

1. **Robust Client Integration**: Uses existing `RobustMeshtasticClient` for reliable communication
2. **Proper Protobuf Handling**: Correctly constructs `MeshPacket` with `PayloadVariant`
3. **Destination Parsing**: Supports multiple formats:
   - Hex node ID: `!a4c138f4`
   - Decimal node ID: `12345`
   - Broadcast: `broadcast` or `all`
4. **Message Filtering**: Channel and sender filtering for listen command
5. **Position Validation**: Coordinate range validation and format conversion
6. **Error Handling**: Comprehensive error handling with meaningful messages
7. **Logging**: Structured logging with configurable levels

### Code Structure

- **Message State**: Tracks last message sender and channel for reply functionality
- **Destination Resolution**: Unified parsing for various node ID formats
- **Coordinate Conversion**: Proper handling of integer coordinates (multiply by 1e7)
- **Protocol Handling**: Uses correct port numbers and message types
- **Admin Messages**: Proper use of AdminMessage for position configuration

### Dependencies

- Uses existing robust client interfaces
- Leverages protobuf definitions from `pkg/pb`
- Integrates with cobra CLI framework
- Uses zerolog for structured logging

## Testing

- **Build Tests**: All commands compile successfully
- **Help Text**: Comprehensive help text with examples
- **Flag Validation**: Required flags are properly validated
- **Error Handling**: Graceful handling of connection errors

## Integration

The commands are integrated into the existing CLI structure:
- Added to `rootCmd` in `init()` functions
- Follow existing patterns for client creation and connection
- Use global configuration for ports, timeouts, and logging
- Compatible with existing flags and configuration

## Usage Examples

```bash
# Send a broadcast message
meshtastic message send "Hello mesh network!"

# Send to specific node
meshtastic message send "Private message" --dest !a4c138f4

# Listen for messages with filtering
meshtastic message listen --channel 1 --timeout 60s

# Set fixed position
meshtastic position set --lat 40.7128 --lon -74.0060 --alt 10

# Request position from all nodes
meshtastic position request --all --timeout 30s

# Broadcast current position
meshtastic position broadcast
```

## Future Enhancements

- Node name resolution for destinations
- DMS coordinate format support
- Message encryption status display
- Position history tracking
- Batch message operations
- Enhanced filtering options
