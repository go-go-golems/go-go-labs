# Notification Sound Generator

This project provides multiple ways to generate short notification sounds programmatically. You can either generate sounds using mathematical synthesis or download free sounds from the internet.

## Features

- 🎵 **Programmatic sound generation** using sine wave synthesis
- 🔧 **Customizable parameters** (frequency, duration, volume)
- 📦 **Multiple output formats** (WAV files)
- 🎼 **Chord generation** (combine multiple frequencies)
- 🌐 **Download free sounds** from the internet
- 🛠️ **Both Go and Python implementations**

## Quick Start

### Using Go

1. **Generate a default notification sound:**
```bash
go run cmd/notification-sound/main.go generate -o notification.wav
```

2. **Generate a custom sound:**
```bash
go run cmd/notification-sound/main.go generate -o custom.wav -f 1000 -d 300ms -v 0.5
```

3. **List available sounds for download:**
```bash
go run cmd/notification-sound/main.go list
```

### Using Python

1. **Install dependencies:**
```bash
pip install numpy
```

2. **Generate a default notification sound:**
```bash
python python/notification_sound.py -o notification.wav
```

3. **Generate all presets:**
```bash
python python/notification_sound.py --generate-all
```

4. **Generate a chord:**
```bash
python python/notification_sound.py --chord 440 554.37 659.25 -o chord.wav
```

## Go Implementation

### Package Usage

```go
package main

import (
    "time"
    "github.com/wesen/corporate-headquarters/go-go-labs/pkg/audio"
)

func main() {
    // Create a notification sound generator
    ns := audio.NewNotificationSound()
    
    // Customize parameters
    ns.Frequency = 800.0  // Hz
    ns.Duration = 500 * time.Millisecond
    ns.Volume = 0.3       // 30% volume
    
    // Save to file
    err := ns.SaveToFile("notification.wav")
    if err != nil {
        panic(err)
    }
}
```

### CLI Tool

The command-line tool provides three main commands:

#### Generate Command
```bash
notification-sound generate [flags]

Flags:
  -o, --output string      Output file path (required)
  -f, --frequency float    Sound frequency in Hz (default: 800)
  -d, --duration duration  Sound duration (default: 500ms)
  -v, --volume float       Sound volume 0.0-1.0 (default: 0.3)
```

#### Download Command
```bash
notification-sound download [flags]

Flags:
  -o, --output string  Output file path (required)
  -t, --type string    Sound type to download (required)
```

#### List Command
```bash
notification-sound list
```

### Examples

Run the examples to see different notification sounds:
```bash
go run examples/notification-sound/main.go
```

This will generate:
- `sounds/default_notification.wav` - Standard notification
- `sounds/alert.wav` - High-pitched alert
- `sounds/low_notification.wav` - Low-pitched notification
- `sounds/double_beep.wav` - Double beep pattern

## Python Implementation

### Available Presets

The Python implementation includes several built-in presets:

| Preset  | Frequency | Duration | Amplitude | Use Case |
|---------|-----------|----------|-----------|----------|
| default | 800 Hz    | 0.5s     | 0.3       | General notifications |
| alert   | 1200 Hz   | 0.2s     | 0.4       | Urgent alerts |
| gentle  | 600 Hz    | 0.8s     | 0.2       | Subtle notifications |
| urgent  | 1000 Hz   | 0.15s    | 0.5       | Critical alerts |
| low     | 400 Hz    | 0.6s     | 0.25      | Background notifications |
| high    | 1500 Hz   | 0.3s     | 0.35      | Attention-grabbing |

### Usage Examples

```bash
# Generate using a preset
python python/notification_sound.py --preset alert -o alert.wav

# Generate custom sound
python python/notification_sound.py --frequency 1000 --duration 0.3 --amplitude 0.4 -o custom.wav

# Generate a major chord (C-E-G)
python python/notification_sound.py --chord 523.25 659.25 783.99 -o chord.wav

# Generate all presets and examples
python python/notification_sound.py --generate-all
```

## Sound Characteristics

### Frequency Guidelines

- **200-400 Hz**: Deep, bass-like tones (good for background notifications)
- **400-800 Hz**: Mid-range tones (pleasant for general use)
- **800-1200 Hz**: Higher tones (good for alerts)
- **1200+ Hz**: High-pitched tones (attention-grabbing, use sparingly)

### Duration Guidelines

- **50-150ms**: Very short beeps (urgent alerts)
- **200-500ms**: Standard notifications
- **500-1000ms**: Longer notifications (less intrusive)
- **1000ms+**: Extended tones (background ambience)

### Volume Guidelines

- **0.1-0.2**: Very quiet (background notifications)
- **0.2-0.4**: Normal volume (most use cases)
- **0.4-0.6**: Louder (alerts and important notifications)
- **0.6+**: Very loud (emergency alerts only)

## Playing Generated Sounds

### Linux
```bash
# Using aplay (ALSA)
aplay notification.wav

# Using paplay (PulseAudio)
paplay notification.wav

# Using mpv
mpv notification.wav
```

### macOS
```bash
# Using afplay
afplay notification.wav

# Using say (text-to-speech, for testing)
say "Notification sound generated"
```

### Windows
```bash
# Using start command
start notification.wav

# Using PowerShell
powershell -c "(New-Object Media.SoundPlayer 'notification.wav').PlaySync()"
```

## Integration Examples

### Web Applications

For web applications, you can serve the generated WAV files and play them using JavaScript:

```javascript
// Play notification sound
function playNotification() {
    const audio = new Audio('/sounds/notification.wav');
    audio.play().catch(e => console.log('Audio play failed:', e));
}
```

### Desktop Applications

For desktop applications, you can use the generated sounds with various audio libraries:

```go
// Example using a hypothetical audio library
func playNotificationSound() {
    player := audio.NewPlayer()
    player.Play("notification.wav")
}
```

## Free Sound Resources

If you prefer to use pre-made sounds instead of generating them, here are some good resources:

- **Freesound.org**: Large collection of Creative Commons licensed sounds
- **Zapsplat**: Professional sound effects (requires free account)
- **BBC Sound Effects**: Free sound effects from the BBC
- **YouTube Audio Library**: Royalty-free sounds and music

## Technical Details

### WAV File Format

The generated sounds use the WAV (Waveform Audio File Format) with these specifications:

- **Sample Rate**: 44.1 kHz (CD quality)
- **Bit Depth**: 16-bit
- **Channels**: 1 (mono)
- **Format**: PCM (uncompressed)

### Audio Processing

- **Fade In/Out**: 10ms fade applied to prevent audio clicks
- **Sine Wave Generation**: Pure mathematical sine wave synthesis
- **Volume Control**: Linear amplitude scaling
- **Anti-aliasing**: Proper sampling to prevent frequency artifacts

## Building and Installation

### Go Version

```bash
# Build the CLI tool
go build -o notification-sound cmd/notification-sound/main.go

# Install globally
go install cmd/notification-sound/main.go

# Run examples
go run examples/notification-sound/main.go
```

### Python Version

```bash
# Install dependencies
pip install numpy

# Make executable
chmod +x python/notification_sound.py

# Run directly
python python/notification_sound.py --help
```

## Contributing

Feel free to contribute improvements:

1. Add new sound generation algorithms (sawtooth, square wave, etc.)
2. Implement more audio effects (reverb, echo, etc.)
3. Add support for other audio formats (MP3, OGG, etc.)
4. Create more preset configurations
5. Improve the CLI interface

## License

This project is licensed under the same license as the parent repository. 