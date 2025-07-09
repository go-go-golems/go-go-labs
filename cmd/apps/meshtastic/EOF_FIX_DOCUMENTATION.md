# EOF Disconnection Fix Documentation

## Problem Description

The Meshtastic application was experiencing sudden disconnections with "Serial read error error=EOF" and "Device disconnected error=EOF" messages. This typically occurred right after receiving telemetry data, causing the application to lose connection permanently.

## Root Cause Analysis

1. **Small Buffer Size**: The original 256-byte buffer was insufficient for handling bursts of telemetry data
2. **No Reconnection Logic**: When EOF occurred, the connection was lost permanently with no recovery mechanism
3. **Poor Error Handling**: The reader loop would exit immediately on EOF instead of attempting recovery
4. **Buffer Overflow**: Large amounts of debug data could cause parsing issues

## Solution Implemented

### 1. Enhanced Serial Interface (`pkg/serial/interface.go`)

**Key Changes:**
- Increased buffer size from 256 to 1024 bytes
- Added automatic reconnection logic with configurable retry attempts
- Improved EOF error handling with specific recovery logic
- Added proper port cleanup on disconnection
- Reset reconnection counter on successful connection

**New Features:**
- `maxReconnectAttempts` (default: 5)
- `reconnectDelay` (default: 2 seconds)
- `attemptReconnect()` method for automatic recovery
- `GetReconnectAttempts()` and `ResetReconnectAttempts()` for monitoring

### 2. Improved Protocol Framing (`pkg/protocol/framing.go`)

**Key Changes:**
- Added panic recovery in `ProcessBytes()` method
- Enhanced payload length validation to prevent negative values
- Better error handling for invalid frame data

### 3. Better Client Error Handling (`pkg/client/client.go`)

**Key Changes:**
- Added proper EOF error detection
- Enhanced disconnect handler with EOF-specific logging
- Improved error context for debugging

## How It Works

1. **Normal Operation**: Data flows through the enlarged buffer to the frame parser
2. **EOF Detection**: When EOF is detected, the connection is marked as disconnected
3. **Automatic Reconnection**: The reader loop attempts to reconnect using exponential backoff
4. **Recovery**: On successful reconnection, the counter is reset and operation resumes
5. **Failure Handling**: After maximum attempts, the connection is considered permanently lost

## Configuration

The reconnection behavior can be customized:

```go
si := &SerialInterface{
    maxReconnectAttempts: 5,                // Maximum retry attempts
    reconnectDelay:       2 * time.Second,  // Delay between attempts
}
```

## Testing

To test the fix:

1. **Compile the application**: `go build -o meshtastic-tui .`
2. **Run with debug logging**: `./meshtastic-tui --log-level debug listen`
3. **Simulate disconnection**: Unplug and replug the device
4. **Observe reconnection**: Check logs for reconnection attempts

## Benefits

- **Automatic Recovery**: No manual intervention needed for temporary disconnections
- **Better Reliability**: Handles burst data and temporary connection issues
- **Improved Debugging**: Enhanced logging for troubleshooting
- **Graceful Degradation**: Continues operation even with intermittent connectivity issues

## Monitoring

The application now provides:
- Connection status logging
- Reconnection attempt counts
- EOF-specific error messages
- Recovery success/failure notifications

This fix should resolve the EOF disconnection issue while providing a robust, self-healing connection mechanism.
