# Film Development Timer for Pico Display

A MicroPython-based film development timer application for the Raspberry Pi Pico with Pimoroni Pico Display. This application helps photographers track development and fixing times for film processing with an easy-to-use interface.

## Features

- Track separate development and fixing times
- Store multiple chemical presets
- Save experiment results
- Clear visual interface
- Simple 4-button control system
- Experiment numbering for record keeping

## Hardware Requirements

- Raspberry Pi Pico
- Pimoroni Pico Display (240x135 LCD)
- USB cable for power/programming

## Software Requirements

- MicroPython with Pimoroni Display Support
- Required libraries:
  - `picographics`
  - `pimoroni`
  - `machine`
  - `time`

## Installation

1. Flash MicroPython to your Pico
2. Copy the script to your Pico as `main.py`
3. Reset the Pico to start the application

## Interface

### Display Layout

The display shows different information based on the current stage:

#### Idle Stage
- Experiment number
- Selected developer
- Selected fixer
- Button controls

#### Development/Fixing Stage
- Current stage (DEVELOP/FIX)
- Selected chemical
- Timer display (MM:SS)
- Button controls

#### Complete Stage
- Final times for both stages
- Options to save or discard results

### Button Controls

The device uses four buttons labeled A, B, C, and D:

#### Idle Mode
- Button A: Cycle through developers
- Button B: Cycle through fixers
- Button C: Start development timer
- Button D: Increment experiment number

#### Development Mode
- Button C: Switch to fixing stage
- Button D: Cancel and return to idle

#### Fixing Mode
- Button C: Complete process
- Button D: Cancel and return to idle

#### Complete Mode
- Button C: Save results
- Button D: Discard and return to idle

## Chemical Presets

### Developers
- D-76
- Rodinal
- HC-110
- XTOL
- Pyro

### Fixers
- Rapid Fix
- TF-4
- Ilford
- Kodak

## Data Storage

Results are stored in a text file (`film_results.txt`) with the following format:
```
experiment_number,developer,fixer,develop_time,fix_time,timestamp
```

## Technical Details

### Timer Implementation
- Uses hardware Timer for accurate timing
- 100ms update frequency
- Time tracked in milliseconds for precision
- Displayed in MM:SS format

### Display Updates
- Screen refreshes on every timer tick
- Clear visual hierarchy with color coding
- Status information always visible

### Error Handling
- Graceful handling of file operations
- Results stored in memory if file writing fails
- Button debouncing to prevent accidental triggers

## Usage Example

1. Power on the device
2. Use buttons A and B to select your developer and fixer
3. Press C to start development timer
4. When development is complete, press C to switch to fixing
5. When fixing is complete, press C again
6. Choose to save (C) or discard (D) the results
7. Repeat for next experiment

## Tips

- Keep track of experiment numbers for consistent record keeping
- Save results after each session
- Note any special conditions in a separate log
- Clean the display between sessions to maintain visibility

## Customization

You can customize the chemical presets by modifying the `DEVELOPERS` and `FIXERS` lists in the code:

```python
DEVELOPERS = ["D-76", "Rodinal", "HC-110", "XTOL", "Pyro"]
FIXERS = ["Rapid Fix", "TF-4", "Ilford", "Kodak"]
```

## Troubleshooting

### Display Issues
- Check power connection
- Verify display initialization
- Reset device if display becomes unresponsive

### Timer Issues
- Check if timer callback is running
- Verify timer initialization parameters
- Reset device if timing becomes inaccurate

### Storage Issues
- Check available storage space
- Verify file permissions
- Results are kept in memory if file operations fail

## Future Improvements

Potential enhancements for future versions:

- Temperature compensation
- Custom chemical entry
- Multiple timer presets
- Data export functionality
- Temperature monitoring
- Agitation reminders
- Backlight control
- Battery monitoring (if battery-powered)

## Contributing

Feel free to fork and improve this project. Some areas that could use enhancement:

- Additional chemical presets
- Enhanced data logging
- UI improvements
- Temperature monitoring
- Network connectivity for data backup

## Credits

Developed using:
- MicroPython
- Pimoroni Display Libraries
- Raspberry Pi Pico Hardware

## Version History

- 1.0.0: Initial release
  - Basic timing functionality
  - Chemical presets
  - Result storage 