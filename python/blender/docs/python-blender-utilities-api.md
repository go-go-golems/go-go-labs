# Blender Python Utilities API Reference

This document provides quick reference for the VSE (Video Sequence Editor) utility functions available in the `vse_utils.py` module. Use this as a quick reference instead of having to read through the entire file.

## Importing VSE Utilities

```python
# Add utility directory to path
import sys, os, importlib
scripts_dir = os.path.dirname(os.path.abspath(__file__))
utils_dir = os.path.join(scripts_dir, 'utils')
for path in [scripts_dir, utils_dir]:
    if path not in sys.path: sys.path.append(path)

# Import and reload the VSE utilities
from utils import vse_utils
importlib.reload(vse_utils)
```

## Core VSE Setup Functions

### `get_active_scene()`
- **Returns**: The currently active scene in Blender

### `ensure_sequence_editor(scene=None)`
- **Parameters**:
  - `scene`: Optional scene object (defaults to active scene)
- **Returns**: The sequence editor for the scene (creates one if it doesn't exist)

## Media Handling Functions

### `add_movie_strip(seq_editor, filepath, channel=1, frame_start=1, name=None)`
- **Parameters**:
  - `seq_editor`: Sequence editor to add the strip to
  - `filepath`: Path to the movie file
  - `channel`: Channel number (default: 1)
  - `frame_start`: Starting frame (default: 1)
  - `name`: Optional custom name (defaults to filename)
- **Returns**: The created movie strip

### `add_sound_strip(seq_editor, filepath, channel=1, frame_start=1, name=None)`
- **Parameters**:
  - `seq_editor`: Sequence editor to add the strip to
  - `filepath`: Path to the sound file
  - `channel`: Channel number (default: 1)
  - `frame_start`: Starting frame (default: 1)
  - `name`: Optional custom name (defaults to filename)
- **Returns**: The created sound strip

## Scene Management Functions

### `clear_all_strips(seq_editor)`
- **Parameters**:
  - `seq_editor`: Sequence editor to clear
- **Returns**: `True` when complete

### `print_sequence_info(seq_editor, title="Current Sequence State")`
- **Parameters**:
  - `seq_editor`: Sequence editor to analyze
  - `title`: Optional title for the output (default: "Current Sequence State")
- **Effect**: Prints detailed information about all strips

## Technical Functions

### `check_and_set_fps(seq_editor, scene)`
- **Parameters**:
  - `seq_editor`: Sequence editor to check
  - `scene`: Scene to check/modify
- **Returns**: Tuple of `(success, message)`
- **Effect**: Checks FPS of all strips and sets scene FPS to match the most common

### `find_test_media_dir(default_path="/home/manuel/Movies/blender-movie-editor")`
- **Parameters**:
  - `default_path`: Default path to look for test media
- **Returns**: Path to a directory with test media files

### `setup_test_sequence(seq_editor, test_media_dir, video_files=None)`
- **Parameters**:
  - `seq_editor`: Sequence editor to add strips to
  - `test_media_dir`: Directory with test media files
  - `video_files`: Optional list of video filenames (defaults to sample videos)
- **Returns**: List of (video_strip, audio_strip) tuples for added clips
- **Effect**: Sets up a test sequence with video clips and matches audio

## Diagnostics Functions

### `print_strip_details(strip, label="Strip")`
- **Parameters**:
  - `strip`: The strip to analyze
  - `label`: Optional label for the output (default: "Strip")
- **Effect**: Prints very detailed information about the strip

## Usage Example

```python
import bpy
from utils import vse_utils

# Setup VSE
scene = vse_utils.get_active_scene()
seq_editor = vse_utils.ensure_sequence_editor(scene)

# Find test media
media_dir = vse_utils.find_test_media_dir()

# Setup a test sequence
clips = vse_utils.setup_test_sequence(seq_editor, media_dir)

# Print sequence information
vse_utils.print_sequence_info(seq_editor, "Test Sequence Setup") 