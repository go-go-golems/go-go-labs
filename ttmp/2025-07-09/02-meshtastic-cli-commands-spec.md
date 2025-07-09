# Meshtastic CLI Commands Specification

## Executive Summary

This specification defines a comprehensive set of CLI commands for interfacing with Meshtastic devices, mirroring the functionality of the Python-based `meshtastic` tool. The commands are organized into 8 priority groups covering the most commonly used functionality, with clear syntax, validation requirements, and implementation guidelines.

### Command Groups Overview

1. **Connection & Discovery** - Device connectivity and discovery
2. **Device Information** - Device status and node information
3. **Configuration Management** - Device configuration operations
4. **Channel Management** - Channel setup and QR code generation
5. **Messaging** - Send/receive text messages
6. **Position Management** - GPS position handling
7. **Device Management** - Device control operations
8. **Telemetry & Monitoring** - Network diagnostics and telemetry

## Implementation Priority

- **Priority 1 (MVP)**: Connection, Device Information, Basic Configuration
- **Priority 2 (Core)**: Channel Management, Messaging, Position Management
- **Priority 3 (Advanced)**: Device Management, Telemetry & Monitoring

---

## 1. Connection & Discovery

### 1.1 Device Connection Commands

#### `meshtastic connect`
Connect to a Meshtastic device using various connection methods.

**Syntax:**
```bash
meshtastic connect [OPTIONS]
```

**Options:**
- `--port PORT` - Serial port (e.g., `/dev/ttyUSB0`, `COM3`)
- `--host HOST` - TCP/IP host (e.g., `192.168.1.100`, `meshtastic.local`)
- `--ble-address ADDR` - BLE device address
- `--ble-scan` - Scan for BLE devices
- `--timeout SECONDS` - Connection timeout (default: 30)

**Examples:**
```bash
meshtastic connect --port /dev/ttyUSB0
meshtastic connect --host 192.168.1.100
meshtastic connect --ble-scan
```

**Output Format:**
```
Connected to: Meshtastic Device (T-Beam v1.1)
Firmware: 2.3.2.f4e6b2c
Hardware: TBEAM_0.7
Node ID: !a4c138f4
```

#### `meshtastic discover`
Discover available Meshtastic devices on all interfaces.

**Syntax:**
```bash
meshtastic discover [OPTIONS]
```

**Options:**
- `--serial-only` - Only scan serial ports
- `--tcp-only` - Only scan TCP/IP network
- `--ble-only` - Only scan BLE devices
- `--timeout SECONDS` - Discovery timeout (default: 10)

**Output Format:**
```
Discovered devices:
Serial:
  /dev/ttyUSB0 - T-Beam v1.1 (!a4c138f4)
  /dev/ttyACM0 - Heltec V3 (!b2d847a1)
Network:
  192.168.1.100 - T-Beam v1.1 (!a4c138f4)
BLE:
  A4:C1:38:F4:12:34 - Meshtastic A4C1 (!a4c138f4)
```

### 1.2 Connection Management

#### `meshtastic disconnect`
Disconnect from current device.

#### `meshtastic status`
Show current connection status.

**Output Format:**
```
Connection Status: Connected
Interface: Serial (/dev/ttyUSB0)
Device: T-Beam v1.1
Firmware: 2.3.2.f4e6b2c
Uptime: 2h 15m 32s
```

---

## 2. Device Information

### 2.1 Device Information Commands

#### `meshtastic info`
Display comprehensive device information.

**Syntax:**
```bash
meshtastic info [OPTIONS]
```

**Options:**
- `--json` - Output in JSON format
- `--yaml` - Output in YAML format

**Output Format:**
```
Device Information:
  Node ID: !a4c138f4
  User: User123 (U123)
  Hardware: TBEAM_0.7
  Firmware: 2.3.2.f4e6b2c
  Region: US
  Modem Preset: Long Range
  
Hardware Details:
  Battery: 3.85V (85%)
  Voltage: 4.12V
  Channel Utilization: 12.5%
  Air Time: 2.3%
  
Network:
  Mesh ID: !a4c138f4
  Nodes in mesh: 5
  Channels: 3
```

#### `meshtastic nodes`
Display information about nodes in the mesh network.

**Syntax:**
```bash
meshtastic nodes [OPTIONS]
```

**Options:**
- `--show-fields FIELDS` - Comma-separated list of fields to display
- `--sort-by FIELD` - Sort nodes by field (id, user, snr, distance, last_heard)
- `--json` - Output in JSON format
- `--live` - Live updating display

**Available Fields:**
- `id` - Node ID
- `user` - User name
- `hardware` - Hardware model
- `snr` - Signal-to-noise ratio
- `distance` - Distance from current node
- `last_heard` - Last heard timestamp
- `battery` - Battery level
- `position` - GPS coordinates
- `role` - Node role (CLIENT, ROUTER, REPEATER)

**Examples:**
```bash
meshtastic nodes
meshtastic nodes --show-fields id,user,snr,distance
meshtastic nodes --sort-by snr --show-fields id,user,snr
```

**Output Format:**
```
Mesh Network Nodes (5 nodes):
┌─────────────┬──────────┬─────────────┬─────┬──────────┬─────────────┐
│ Node ID     │ User     │ Hardware    │ SNR │ Distance │ Last Heard  │
├─────────────┼──────────┼─────────────┼─────┼──────────┼─────────────┤
│ !a4c138f4   │ User123  │ TBEAM_0.7   │ N/A │ 0.0 km   │ Self        │
│ !b2d847a1   │ Node2    │ HELTEC_V3   │ 8.2 │ 2.3 km   │ 2m ago      │
│ !c3e952b7   │ Repeater │ TBEAM_1.1   │ 5.1 │ 5.7 km   │ 5m ago      │
│ !d4f063c8   │ Mobile   │ TBEAM_0.7   │ 12.5│ 1.2 km   │ 1m ago      │
│ !e5g174d9   │ Base     │ STATION_G1  │ 15.3│ 8.9 km   │ 3m ago      │
└─────────────┴──────────┴─────────────┴─────┴──────────┴─────────────┘
```

---

## 3. Configuration Management

### 3.1 Configuration Read Commands

#### `meshtastic get`
Retrieve configuration values from the device.

**Syntax:**
```bash
meshtastic get [FIELD] [OPTIONS]
```

**Options:**
- `--all` - Get all configuration fields
- `--section SECTION` - Get all fields from a configuration section
- `--json` - Output in JSON format
- `--yaml` - Output in YAML format

**Configuration Sections:**
- `device` - Device configuration
- `position` - GPS/position settings
- `power` - Power management
- `network` - Network settings
- `display` - Display configuration
- `lora` - LoRa radio settings
- `bluetooth` - Bluetooth settings

**Examples:**
```bash
meshtastic get device.role
meshtastic get --section lora
meshtastic get --all
```

**Output Format:**
```bash
# Single field
meshtastic get device.role
device.role: CLIENT

# Section
meshtastic get --section lora
lora.use_preset: true
lora.modem_preset: LONG_RANGE
lora.bandwidth: 125
lora.spreading_factor: 12
lora.coding_rate: 4/5
lora.frequency_offset: 0
lora.hop_limit: 3
```

### 3.2 Configuration Write Commands

#### `meshtastic set`
Set configuration values on the device.

**Syntax:**
```bash
meshtastic set FIELD VALUE [OPTIONS]
```

**Options:**
- `--confirm` - Skip confirmation prompt
- `--dry-run` - Show what would be changed without applying

**Examples:**
```bash
meshtastic set device.role ROUTER
meshtastic set lora.modem_preset LONG_RANGE
meshtastic set power.ls_secs 300
```

**Output Format:**
```
Setting device.role = ROUTER
Confirming change... 
✓ Configuration updated successfully
Device will reboot to apply changes.
```

### 3.3 Configuration Import/Export

#### `meshtastic export-config`
Export device configuration to a file.

**Syntax:**
```bash
meshtastic export-config [FILENAME] [OPTIONS]
```

**Options:**
- `--format FORMAT` - Output format (yaml, json, txt)
- `--sections SECTIONS` - Comma-separated list of sections to export
- `--stdout` - Output to stdout instead of file

**Examples:**
```bash
meshtastic export-config my-config.yaml
meshtastic export-config --format json --stdout
```

#### `meshtastic import-config`
Import configuration from a file.

**Syntax:**
```bash
meshtastic import-config FILENAME [OPTIONS]
```

**Options:**
- `--dry-run` - Show what would be changed without applying
- `--force` - Skip confirmation prompts
- `--sections SECTIONS` - Only import specified sections

**Examples:**
```bash
meshtastic import-config my-config.yaml
meshtastic import-config backup.json --dry-run
```

---

## 4. Channel Management

### 4.1 Channel Operations

#### `meshtastic channels`
List all configured channels.

**Syntax:**
```bash
meshtastic channels [OPTIONS]
```

**Options:**
- `--json` - Output in JSON format
- `--show-keys` - Show encryption keys (security warning)

**Output Format:**
```
Channels:
┌─────┬─────────────┬─────────────────┬─────────┬─────────────────┐
│ ID  │ Name        │ Mode           │ Role    │ PSK             │
├─────┼─────────────┼─────────────────┼─────────┼─────────────────┤
│ 0   │ LongRange   │ LONG_RANGE     │ PRIMARY │ default         │
│ 1   │ Emergency   │ VERY_LONG_RANGE │ SECONDARY│ ***encrypted*** │
│ 2   │ Local       │ MEDIUM_RANGE   │ DISABLED │ ***encrypted*** │
└─────┴─────────────┴─────────────────┴─────────┴─────────────────┘
```

#### `meshtastic channel add`
Add a new channel configuration.

**Syntax:**
```bash
meshtastic channel add [OPTIONS]
```

**Options:**
- `--name NAME` - Channel name
- `--psk PSK` - Pre-shared key (hex or base64)
- `--role ROLE` - Channel role (PRIMARY, SECONDARY, DISABLED)
- `--index INDEX` - Channel index (0-7)

**Examples:**
```bash
meshtastic channel add --name "Emergency" --psk "AQ==" --role SECONDARY
meshtastic channel add --name "Local" --index 2
```

#### `meshtastic channel set`
Modify an existing channel.

**Syntax:**
```bash
meshtastic channel set INDEX [OPTIONS]
```

**Options:**
- `--name NAME` - Channel name
- `--psk PSK` - Pre-shared key
- `--role ROLE` - Channel role
- `--uplink-enabled BOOL` - Enable uplink
- `--downlink-enabled BOOL` - Enable downlink

**Examples:**
```bash
meshtastic channel set 1 --name "NewName"
meshtastic channel set 2 --role DISABLED
```

#### `meshtastic channel delete`
Delete a channel configuration.

**Syntax:**
```bash
meshtastic channel delete INDEX [OPTIONS]
```

**Options:**
- `--confirm` - Skip confirmation prompt

### 4.2 Channel Sharing

#### `meshtastic qr`
Generate QR code for channel sharing.

**Syntax:**
```bash
meshtastic qr [OPTIONS]
```

**Options:**
- `--channel INDEX` - Channel index (default: 0)
- `--output FILE` - Save QR code to file
- `--format FORMAT` - Output format (png, svg, ascii)
- `--size SIZE` - QR code size in pixels

**Examples:**
```bash
meshtastic qr
meshtastic qr --channel 1 --output emergency.png
meshtastic qr --format ascii
```

**Output Format:**
```
Channel: LongRange
URL: https://meshtastic.org/e/#ChMSCUxvbmdSYW5nZQ...

QR Code:
████████████████████████████████
██      ██  ██    ██      ██
██  ██  ██  ████  ██  ██  ██
██  ██  ██  ██  ████  ██  ██
██  ██  ██  ██    ██  ██  ██
██      ██  ██  ████      ██
████████████████████████████████
```

---

## 5. Messaging

### 5.1 Send Messages

#### `meshtastic send`
Send a text message to the mesh network.

**Syntax:**
```bash
meshtastic send MESSAGE [OPTIONS]
```

**Options:**
- `--to NODE_ID` - Send to specific node (private message)
- `--channel INDEX` - Send on specific channel (default: 0)
- `--want-ack` - Request acknowledgment
- `--hop-limit HOPS` - Maximum hop limit

**Examples:**
```bash
meshtastic send "Hello mesh network!"
meshtastic send "Private message" --to !b2d847a1
meshtastic send "Emergency!" --channel 1 --want-ack
```

**Output Format:**
```
Sending message to mesh network...
✓ Message sent successfully
Message ID: 0x1234abcd
```

### 5.2 Receive Messages

#### `meshtastic listen`
Listen for incoming messages.

**Syntax:**
```bash
meshtastic listen [OPTIONS]
```

**Options:**
- `--channel INDEX` - Listen on specific channel
- `--from NODE_ID` - Only show messages from specific node
- `--timeout SECONDS` - Listen timeout (0 for infinite)
- `--json` - Output messages in JSON format

**Examples:**
```bash
meshtastic listen
meshtastic listen --channel 1 --timeout 60
meshtastic listen --from !b2d847a1
```

**Output Format:**
```
Listening for messages... (Press Ctrl+C to stop)
[2024-01-15 14:30:15] User123 (!a4c138f4): Hello mesh network!
[2024-01-15 14:30:22] Node2 (!b2d847a1) [Private]: How are you?
[2024-01-15 14:30:45] Repeater (!c3e952b7) [Ch1]: Emergency broadcast
```

### 5.3 Message Management

#### `meshtastic reply`
Reply to the last received message.

**Syntax:**
```bash
meshtastic reply MESSAGE [OPTIONS]
```

**Options:**
- `--want-ack` - Request acknowledgment

**Examples:**
```bash
meshtastic reply "Thanks for the message!"
```

---

## 6. Position Management

### 6.1 Position Configuration

#### `meshtastic position set`
Set the device's GPS position.

**Syntax:**
```bash
meshtastic position set [OPTIONS]
```

**Options:**
- `--latitude LAT` - Latitude in decimal degrees
- `--longitude LON` - Longitude in decimal degrees
- `--altitude ALT` - Altitude in meters
- `--fixed` - Mark position as fixed (not GPS)

**Examples:**
```bash
meshtastic position set --latitude 40.7128 --longitude -74.0060
meshtastic position set --latitude 40.7128 --longitude -74.0060 --altitude 10 --fixed
```

#### `meshtastic position get`
Get the current device position.

**Output Format:**
```
Current Position:
  Latitude: 40.7128°N
  Longitude: 74.0060°W
  Altitude: 10m
  Source: GPS
  Precision: ±3m
  Last Updated: 2024-01-15 14:30:45
```

#### `meshtastic position clear`
Clear the device's stored position.

### 6.2 Position Requests

#### `meshtastic position request`
Request position from other nodes.

**Syntax:**
```bash
meshtastic position request [NODE_ID] [OPTIONS]
```

**Options:**
- `--timeout SECONDS` - Request timeout (default: 30)
- `--all` - Request from all nodes

**Examples:**
```bash
meshtastic position request !b2d847a1
meshtastic position request --all
```

---

## 7. Device Management

### 7.1 Device Control

#### `meshtastic reboot`
Reboot the connected device.

**Syntax:**
```bash
meshtastic reboot [OPTIONS]
```

**Options:**
- `--confirm` - Skip confirmation prompt
- `--ota` - Reboot into OTA mode

#### `meshtastic shutdown`
Shutdown the connected device.

**Syntax:**
```bash
meshtastic shutdown [OPTIONS]
```

**Options:**
- `--confirm` - Skip confirmation prompt

#### `meshtastic factory-reset`
Factory reset the device.

**Syntax:**
```bash
meshtastic factory-reset [OPTIONS]
```

**Options:**
- `--confirm` - Skip confirmation prompt
- `--keep-bluetooth` - Keep Bluetooth settings

### 7.2 Device Configuration

#### `meshtastic set-owner`
Configure device owner information.

**Syntax:**
```bash
meshtastic set-owner [OPTIONS]
```

**Options:**
- `--long-name NAME` - Full device name
- `--short-name NAME` - Short device name (4 chars max)
- `--is-licensed` - Mark as licensed operator

**Examples:**
```bash
meshtastic set-owner --long-name "John's T-Beam" --short-name "J123"
```

---

## 8. Telemetry & Monitoring

### 8.1 Telemetry Commands

#### `meshtastic telemetry`
Display device telemetry data.

**Syntax:**
```bash
meshtastic telemetry [OPTIONS]
```

**Options:**
- `--type TYPE` - Telemetry type (device, environment, power)
- `--from NODE_ID` - Get telemetry from specific node
- `--live` - Live updating display
- `--json` - Output in JSON format

**Examples:**
```bash
meshtastic telemetry
meshtastic telemetry --type device --live
meshtastic telemetry --from !b2d847a1
```

**Output Format:**
```
Device Telemetry:
  Battery: 3.85V (85%)
  Voltage: 4.12V
  Channel Utilization: 12.5%
  Air Time: 2.3%
  Uptime: 2h 15m 32s
  Temperature: 24.5°C

Environment Telemetry:
  Temperature: 22.1°C
  Humidity: 45%
  Pressure: 1013.2 hPa
  Gas Resistance: 50.2 kΩ
```

#### `meshtastic telemetry request`
Request telemetry from other nodes.

**Syntax:**
```bash
meshtastic telemetry request [NODE_ID] [OPTIONS]
```

**Options:**
- `--type TYPE` - Telemetry type to request
- `--timeout SECONDS` - Request timeout

### 8.2 Network Diagnostics

#### `meshtastic traceroute`
Trace route to a destination node.

**Syntax:**
```bash
meshtastic traceroute NODE_ID [OPTIONS]
```

**Options:**
- `--timeout SECONDS` - Timeout per hop (default: 30)
- `--max-hops HOPS` - Maximum hops (default: 7)

**Examples:**
```bash
meshtastic traceroute !b2d847a1
```

**Output Format:**
```
Traceroute to !b2d847a1:
1. !a4c138f4 (self) - 0ms
2. !c3e952b7 (Repeater) - 1234ms SNR: 8.2dB
3. !b2d847a1 (Node2) - 2456ms SNR: 5.1dB
```

---

## Implementation Requirements

### Integration with Robust Client

All commands must integrate with the existing robust client implementation:

1. **Connection Management**: Use established connection patterns
2. **Error Handling**: Implement consistent error handling and retry logic
3. **Logging**: Use structured logging with configurable levels
4. **Configuration**: Support configuration files and environment variables

### Input Validation

1. **Node ID Format**: Validate node IDs match pattern `!([a-f0-9]{8})`
2. **Coordinates**: Validate GPS coordinates are within valid ranges
3. **Frequencies**: Validate radio frequencies are within allowed bands
4. **Channel Numbers**: Validate channel indices are 0-7
5. **Text Length**: Validate message length limits (237 bytes)

### Error Handling

1. **Connection Errors**: Clear messages for connection failures
2. **Timeout Handling**: Appropriate timeouts for all operations
3. **Validation Errors**: Helpful error messages for invalid input
4. **Device Errors**: Handle device-specific error responses

### Output Formatting

1. **Table Format**: Consistent table formatting with borders
2. **JSON/YAML**: Structured output for machine parsing
3. **Colors**: Optional color coding for status indicators
4. **Timestamps**: Consistent timestamp formatting

### Testing Strategy

1. **Unit Tests**: Test individual command parsing and validation
2. **Integration Tests**: Test with mock device responses
3. **End-to-End Tests**: Test with actual hardware when available
4. **Error Scenarios**: Test error handling and edge cases

### Future Extensibility

1. **Plugin System**: Design for future plugin architecture
2. **Custom Commands**: Support for user-defined commands
3. **Scripting**: Enable command scripting and automation
4. **REST API**: Potential REST API wrapper for commands

---

## Command Priority Matrix

| Command Group | Priority | Complexity | Dependencies |
|---------------|----------|------------|--------------|
| Connection & Discovery | 1 | Medium | Serial, TCP, BLE clients |
| Device Information | 1 | Low | Connection established |
| Configuration Management | 1 | Medium | Protobuf definitions |
| Channel Management | 2 | Medium | QR code generation |
| Messaging | 2 | Low | Channel management |
| Position Management | 2 | Low | GPS utilities |
| Device Management | 3 | Low | Device control APIs |
| Telemetry & Monitoring | 3 | Medium | Telemetry parsers |

This specification provides a comprehensive foundation for implementing a full-featured Meshtastic CLI tool that mirrors the Python implementation while leveraging Go's strengths in systems programming and CLI development.
