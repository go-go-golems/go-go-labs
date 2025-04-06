# RTC Transcribe

A real-time audio transcription server using WebRTC and OpenAI's Whisper API.

## Features

- Browser-based audio capture using WebRTC
- Real-time audio streaming to Go server
- Robust Opus audio decoding with libopus support
- Detailed structured logging for debugging and monitoring
- Audio transcription using OpenAI's Whisper API
- Real-time text streaming back to the browser using Server-Sent Events (SSE)
- Clean, simple user interface

## Architecture

RTC Transcribe is built with the following components:

1. **Frontend**: HTML/CSS/JavaScript web client that:
   - Captures microphone audio
   - Establishes WebRTC connection with the server
   - Streams the audio in real-time
   - Receives and displays transcription results

2. **Backend Server**: Go application that:
   - Handles WebRTC signaling and establishes peer connections
   - Receives and decodes Opus audio streams
   - Buffers audio for optimal transcription
   - Transcribes audio using Whisper API
   - Streams results back to the client using SSE

## Requirements

- Go 1.23+
- OpenAI API key for the Whisper API

### Production Requirements

For production use with real Opus decoding (recommended), you'll need:
- libopus and libopusfile development libraries
  - Ubuntu/Debian: `sudo apt-get install libopus-dev libopusfile-dev`
  - macOS: `brew install opus opusfile`
  - Windows: Use MSYS2/MinGW or vcpkg

## Building

### Build with real Opus support (recommended for production)

```bash
# First install the required dependencies
sudo apt-get install libopus-dev libopusfile-dev

# Then build
go build -o rtc-transcribe ./cmd/apps/rtc-transcribe/
```

### Build with mock Opus decoder (no dependencies, reduced audio quality)

```bash
go build -tags noopus -o rtc-transcribe ./cmd/apps/rtc-transcribe/
```

## Usage

Run the application:

```bash
# Set your OpenAI API key
export OPENAI_API_KEY=your_api_key_here

# Run with default settings
./rtc-transcribe

# Run with debug logging
./rtc-transcribe --log-level debug

# Run on a different port
./rtc-transcribe --port 9000

# Specify OpenAI API key directly
./rtc-transcribe --api-key sk-xxxxxxxxxxxxx
```

### Command Line Flags

- `-p, --port`: HTTP server port (default: 8080)
- `-l, --log-level`: Log level (debug, info, warn, error) (default: info)
- `-k, --api-key`: OpenAI API key (defaults to OPENAI_API_KEY env var)
- `-m, --mode`: Transcription mode (currently only "api" is supported)

## Implementation Details

The application is organized into the following packages:

- **webrtc**: Handles WebRTC peer connections and audio processing
- **transcribe**: Manages audio transcription using Whisper API
- **sse**: Handles Server-Sent Events for streaming text back to the client

## Debugging

The application uses structured logging with zerolog. To enable debug logs:

```bash
./rtc-transcribe --log-level debug
```

Important log components:
- `WebRTCSignaling`: WebRTC signaling and connection setup
- `WebRTCPeer`: Peer connection management
- `AudioTrackHandler`: Audio track processing
- `AudioProcessor`: Audio processing and buffer management
- `OpusDecoder`: Audio decoding (real or mock)
- `WhisperAPIClient`: API communication
- `TranscriptionService`: Transcription orchestration
- `SSE`: Server-sent events management

The application buffers 3 seconds of audio before sending it for transcription, which balances responsiveness with transcription accuracy.

## Limitations and Future Work

- Currently only supports the OpenAI Whisper API for transcription
- A local mode using whisper.cpp could be added in the future
- No data persistence or session management
- Basic error handling

## License

See the LICENSE file at the root of the repository.