#!/usr/bin/env python3
"""
Notification Sound Generator

This script provides an alternative Python implementation for generating
notification sounds using numpy and scipy.

Requirements:
    pip install numpy scipy

Usage:
    python notification_sound.py
"""

import numpy as np
import wave
import struct
import argparse
import os
from typing import Tuple


def generate_sine_wave(frequency: float, duration: float, sample_rate: int = 44100, 
                      amplitude: float = 0.3) -> np.ndarray:
    """
    Generate a sine wave with the given parameters.
    
    Args:
        frequency: Frequency in Hz
        duration: Duration in seconds
        sample_rate: Sample rate in Hz
        amplitude: Amplitude (0.0 to 1.0)
    
    Returns:
        numpy array of audio samples
    """
    t = np.linspace(0, duration, int(sample_rate * duration), False)
    
    # Generate sine wave
    wave_data = amplitude * np.sin(2 * np.pi * frequency * t)
    
    # Apply fade in/out to prevent clicks
    fade_samples = int(0.01 * sample_rate)  # 10ms fade
    if len(wave_data) > 2 * fade_samples:
        # Fade in
        fade_in = np.linspace(0, 1, fade_samples)
        wave_data[:fade_samples] *= fade_in
        
        # Fade out
        fade_out = np.linspace(1, 0, fade_samples)
        wave_data[-fade_samples:] *= fade_out
    
    return wave_data


def save_wav(filename: str, audio_data: np.ndarray, sample_rate: int = 44100):
    """
    Save audio data as a WAV file.
    
    Args:
        filename: Output filename
        audio_data: Audio samples as numpy array
        sample_rate: Sample rate in Hz
    """
    # Ensure directory exists
    os.makedirs(os.path.dirname(filename) if os.path.dirname(filename) else '.', exist_ok=True)
    
    # Convert to 16-bit integers
    audio_16bit = (audio_data * 32767).astype(np.int16)
    
    with wave.open(filename, 'w') as wav_file:
        wav_file.setnchannels(1)  # Mono
        wav_file.setsampwidth(2)  # 16-bit
        wav_file.setframerate(sample_rate)
        wav_file.writeframes(audio_16bit.tobytes())


def generate_notification_presets() -> dict:
    """
    Generate various notification sound presets.
    
    Returns:
        Dictionary of preset configurations
    """
    return {
        'default': {'frequency': 800, 'duration': 0.5, 'amplitude': 0.3},
        'alert': {'frequency': 1200, 'duration': 0.2, 'amplitude': 0.4},
        'gentle': {'frequency': 600, 'duration': 0.8, 'amplitude': 0.2},
        'urgent': {'frequency': 1000, 'duration': 0.15, 'amplitude': 0.5},
        'low': {'frequency': 400, 'duration': 0.6, 'amplitude': 0.25},
        'high': {'frequency': 1500, 'duration': 0.3, 'amplitude': 0.35},
    }


def generate_chord(frequencies: list, duration: float, sample_rate: int = 44100,
                  amplitude: float = 0.3) -> np.ndarray:
    """
    Generate a chord by combining multiple frequencies.
    
    Args:
        frequencies: List of frequencies in Hz
        duration: Duration in seconds
        sample_rate: Sample rate in Hz
        amplitude: Amplitude per frequency
    
    Returns:
        numpy array of combined audio samples
    """
    t = np.linspace(0, duration, int(sample_rate * duration), False)
    combined = np.zeros_like(t)
    
    for freq in frequencies:
        wave = amplitude * np.sin(2 * np.pi * freq * t) / len(frequencies)
        combined += wave
    
    # Apply fade in/out
    fade_samples = int(0.01 * sample_rate)
    if len(combined) > 2 * fade_samples:
        fade_in = np.linspace(0, 1, fade_samples)
        combined[:fade_samples] *= fade_in
        
        fade_out = np.linspace(1, 0, fade_samples)
        combined[-fade_samples:] *= fade_out
    
    return combined


def main():
    parser = argparse.ArgumentParser(description='Generate notification sounds')
    parser.add_argument('--preset', choices=list(generate_notification_presets().keys()),
                       help='Use a preset configuration')
    parser.add_argument('--frequency', type=float, default=800,
                       help='Frequency in Hz (default: 800)')
    parser.add_argument('--duration', type=float, default=0.5,
                       help='Duration in seconds (default: 0.5)')
    parser.add_argument('--amplitude', type=float, default=0.3,
                       help='Amplitude 0.0-1.0 (default: 0.3)')
    parser.add_argument('--output', '-o', default='notification.wav',
                       help='Output filename (default: notification.wav)')
    parser.add_argument('--chord', nargs='+', type=float,
                       help='Generate chord with multiple frequencies')
    parser.add_argument('--generate-all', action='store_true',
                       help='Generate all presets')
    
    args = parser.parse_args()
    
    if args.generate_all:
        print("🎵 Generating all notification sound presets...")
        presets = generate_notification_presets()
        
        for name, config in presets.items():
            filename = f"sounds/{name}_notification.wav"
            audio = generate_sine_wave(**config)
            save_wav(filename, audio)
            print(f"✅ {name.capitalize()} notification saved to {filename}")
        
        # Generate some chord examples
        print("\n🎼 Generating chord examples...")
        
        # Major chord (C-E-G)
        major_chord = generate_chord([523.25, 659.25, 783.99], 0.8, amplitude=0.2)
        save_wav("sounds/major_chord.wav", major_chord)
        print("✅ Major chord saved to sounds/major_chord.wav")
        
        # Minor chord (A-C-E)
        minor_chord = generate_chord([440.00, 523.25, 659.25], 0.8, amplitude=0.2)
        save_wav("sounds/minor_chord.wav", minor_chord)
        print("✅ Minor chord saved to sounds/minor_chord.wav")
        
        print("\n🎵 All sounds generated! You can play them with:")
        print("  Linux: aplay sounds/*.wav")
        print("  macOS: afplay sounds/default_notification.wav")
        print("  Windows: start sounds\\default_notification.wav")
        
    elif args.chord:
        print(f"🎼 Generating chord with frequencies: {args.chord}")
        audio = generate_chord(args.chord, args.duration, amplitude=args.amplitude)
        save_wav(args.output, audio)
        print(f"✅ Chord saved to {args.output}")
        
    else:
        # Use preset or custom parameters
        if args.preset:
            presets = generate_notification_presets()
            config = presets[args.preset]
            print(f"🔊 Generating '{args.preset}' preset notification...")
        else:
            config = {
                'frequency': args.frequency,
                'duration': args.duration,
                'amplitude': args.amplitude
            }
            print(f"🔊 Generating custom notification...")
        
        print(f"  Frequency: {config['frequency']} Hz")
        print(f"  Duration: {config['duration']} seconds")
        print(f"  Amplitude: {config['amplitude']}")
        
        audio = generate_sine_wave(**config)
        save_wav(args.output, audio)
        print(f"✅ Notification sound saved to {args.output}")


if __name__ == "__main__":
    main() 