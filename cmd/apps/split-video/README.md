# Split Video Tool

A powerful command-line video splitting tool with both CLI and TUI (Terminal User Interface) modes.

## Features

- **Multiple Split Modes**:
  - Equal segments with configurable overlap
  - Time-based splitting at specific intervals
  - Duration-based splitting with fixed segment lengths
  
- **Audio Extraction**: Optionally extract audio from video segments in various formats (MP3, WAV, AAC, FLAC)

- **Interactive TUI**: Beautiful terminal interface using Bubbletea for easy configuration

- **CLI Commands**: Full command-line interface for scripting and automation

- **Comprehensive Logging**: Structured logging with configurable levels using zerolog

## Installation

```bash
# Clone and build
git clone <repository-url>
cd split-video
go build -o split-video cmd/split-video/main.go

# Or install directly
go install ./cmd/split-video
```

## Usage

### Interactive TUI Mode

Launch the interactive interface:

```bash
# Launch TUI without a file (browse/type to select)
./split-video

# Launch TUI with a pre-loaded video file
./split-video myvideo.mp4
```

The TUI provides:
- Single-screen POS-style interface with all options visible
- Real-time preview of what will be created
- Function key shortcuts (F1-F4) for quick navigation
- Professional terminal interface for efficient operation

### CLI Commands

#### Equal Segments Split

Split a video into equal segments with optional overlap:

```bash
# Split into 5 equal segments with 5-minute overlap
./split-video equal video.mp4 --segments 5 --overlap 5m

# Extract audio from each segment
./split-video equal video.mp4 --segments 3 --extract-audio --audio-format mp3
```

#### Time-based Split

Split video at specific time intervals:

```bash
# Split at 10m, 25m, and 45m marks
./split-video time video.mp4 --intervals 10m,25m,45m

# With audio extraction
./split-video time video.mp4 --intervals 15m,30m --extract-audio
```

#### Duration-based Split

Split into segments of fixed duration:

```bash
# 15-minute segments with 2-minute overlap
./split-video duration video.mp4 --duration 15m --overlap 2m

# 10-minute segments with audio extraction
./split-video duration video.mp4 --duration 10m --extract-audio --audio-format wav
```

#### Audio Extraction Only

Extract audio from video without splitting:

```bash
# Extract as MP3
./split-video audio video.mp4 --format mp3

# Custom output file
./split-video audio video.mp4 --output soundtrack.wav --format wav
```

### Global Options

- `--output, -d`: Output directory (default: current directory)
- `--verbose`: Enable verbose logging
- `--log-level`: Set log level (debug, info, warn, error)

### Audio Formats

Supported audio formats:
- `mp3`: MP3 format (default)
- `wav`: WAV format
- `aac`: AAC format  
- `flac`: FLAC format

## Requirements

- **FFmpeg**: Must be installed and available in PATH
- **Go 1.21+**: For building from source

## Examples

### Basic Usage

```bash
# Launch TUI (empty)
./split-video

# Launch TUI with pre-loaded file
./split-video myvideo.mp4

# Quick 5-segment split via CLI
./split-video equal myvideo.mp4 --segments 5

# Split with overlap and audio extraction via CLI
./split-video equal myvideo.mp4 --segments 3 --overlap 5m --extract-audio
```

### Advanced Usage

```bash
# Custom output directory with verbose logging
./split-video equal myvideo.mp4 --segments 4 --output ./segments --verbose

# Duration-based with specific audio format
./split-video duration myvideo.mp4 --duration 20m --extract-audio --audio-format flac

# Time-based split at specific points
./split-video time myvideo.mp4 --intervals 5m30s,15m45s,32m10s
```

## Architecture

The tool is organized into several packages:

- `cmd/split-video/`: Main application entry point
- `pkg/config/`: Configuration structures and validation
- `pkg/video/`: Video processing logic using FFmpeg
- `pkg/tui/`: Terminal user interface using Bubbletea

## Development

```bash
# Install dependencies
go mod tidy

# Run with TUI
go run cmd/split-video/main.go

# Run CLI command
go run cmd/split-video/main.go equal test.mp4 --segments 2

# Build
go build -o split-video cmd/split-video/main.go
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details.
