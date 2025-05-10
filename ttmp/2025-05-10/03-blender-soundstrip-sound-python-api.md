# Blender Python API: `SoundStrip` and `Sound` Reference

This document provides a reference for `bpy.types.Sound` and `bpy.types.SoundStrip` objects in the Blender Python API, focusing on their use in the Video Sequence Editor (VSE). Understanding these objects is crucial for programmatically manipulating audio in Blender.

**Key Distinction:**
-   `bpy.types.Sound`: Represents an audio data-block in Blender. It holds information about the sound file itself, like its filepath, sample rate, and channels. This is analogous to an Image data-block for images.
-   `bpy.types.SoundStrip`: Represents a strip in the Video Sequence Editor that uses a `Sound` data-block. It controls how and when the sound is played in a sequence, including its position on the timeline, volume, pan, etc.

Official Documentation Links:
-   [`Sound`](https://docs.blender.org/api/current/bpy.types.Sound.html)
-   [`SoundStrip`](https://docs.blender.org/api/current/bpy.types.SoundStrip.html)

---

## `bpy.types.Sound`

A `Sound` data-block stores the actual audio data or a link to an external audio file.

### Key Properties

| Property             | Type         | Description                                                                      | Readonly |
| -------------------- | ------------ | -------------------------------------------------------------------------------- | -------- |
| `filepath`           | `string`     | File path to the sound file.                                                     | No       |
| `packed_file`        | `PackedFile` | The packed sound data if the sound is packed into the .blend file.               | Yes      |
| `samplerate`         | `int`        | Sample rate of the sound in samples per second (e.g., 44100, 48000).              | Yes      |
| `channels`           | `int`        | Number of audio channels (e.g., 1 for mono, 2 for stereo).                       | Yes      |
| `use_memory_cache`   | `boolean`    | If true, the sound is cached in RAM for faster access.                           | No       |
| `use_mono`           | `boolean`    | If true, the sound is mixed down to a single mono channel.                       | No       |
| `name`               | `string`     | The name of the Sound data-block.                                                | No       |
| `users`              | `int`        | Number of users of this data-block (e.g., how many SoundStrips use this sound). | Yes      |
| `is_loaded`          | `boolean`    | True if the sound data has been successfully loaded.                             | Yes      |

### Key Methods

-   **`pack()`**:
    -   Description: Packs the sound into the current .blend file. This embeds the audio data within the .blend file, making it portable but increasing file size.
    -   Usage: `my_sound_data.pack()`

-   **`unpack(method='USE_LOCAL')`**:
    -   Description: Unpacks the sound from the .blend file to an external file.
    -   Arguments:
        -   `method` (enum): How to unpack. Common options:
            -   `'USE_LOCAL'`: Unpack to the `filepath` if it's set and relative, otherwise use a subfolder.
            -   `'WRITE_LOCAL'`: Unpack to `filepath`, overwriting if it exists.
            -   `'USE_ORIGINAL'`: Unpack to the original `filepath` if available.
            -   `'WRITE_ORIGINAL'`: Unpack to original `filepath`, overwriting.
    -   Usage: `my_sound_data.unpack(method='WRITE_LOCAL')`

### Examples for `bpy.types.Sound`

1.  **Loading a new sound file into Blender's data:**
    ```python
    import bpy
    import os

    # Define the filepath to your sound
    sound_filepath = "/path/to/your/audiofile.mp3" # XXX: Change this path

    if os.path.exists(sound_filepath):
        # Load the sound. This creates a new Sound data-block or returns an existing one if the path matches.
        # bpy.data.sounds.load() is the primary way to bring external sounds into Blender's data.
        sound_datablock = bpy.data.sounds.load(sound_filepath, check_existing=True)
        
        print(f"Loaded Sound data-block: '{sound_datablock.name}'")
        print(f"  Filepath: {sound_datablock.filepath}")
        print(f"  Samplerate: {sound_datablock.samplerate} Hz")
        print(f"  Channels: {sound_datablock.channels}")
        print(f"  Is Packed: {sound_datablock.packed_file is not None}")
        print(f"  Users: {sound_datablock.users}")
    else:
        print(f"Error: Sound file not found at '{sound_filepath}'")
    ```

2.  **Accessing an existing sound data-block by name:**
    ```python
    import bpy

    sound_name = "my_existing_sound" # Name of the sound data-block in Blender

    if sound_name in bpy.data.sounds:
        sound_datablock = bpy.data.sounds[sound_name]
        print(f"Found Sound: '{sound_datablock.name}'")
        print(f"  Filepath: {sound_datablock.filepath_raw}") # filepath_raw gives the path as entered
    else:
        print(f"Sound data-block '{sound_name}' not found.")
    ```

3.  **Packing and Unpacking a sound:**
    ```python
    import bpy

    sound_name = "audiofile.mp3" # Assuming "audiofile.mp3" is a loaded sound data-block

    if sound_name in bpy.data.sounds:
        sound_datablock = bpy.data.sounds[sound_name]
        
        if not sound_datablock.packed_file:
            print(f"Packing '{sound_datablock.name}'...")
            sound_datablock.pack()
            print(f"  Is Packed: {sound_datablock.packed_file is not None}")
        else:
            print(f"'{sound_datablock.name}' is already packed.")
            
        # To unpack (e.g., if you want to edit it externally again)
        # Make sure filepath is sensible or it might unpack to a default location
        if sound_datablock.packed_file:
            print(f"Unpacking '{sound_datablock.name}'...")
            try:
                # This will unpack to sound_datablock.filepath if it's a valid relative path
                # or create a subdirectory like 'sounds/'.
                sound_datablock.unpack(method='USE_LOCAL') 
                print(f"  Unpacked. New filepath (if changed): {sound_datablock.filepath}")
                print(f"  Is Packed: {sound_datablock.packed_file is not None}")
            except RuntimeError as e:
                print(f"  Error unpacking: {e}")
    else:
        print(f"Sound data-block '{sound_name}' not found.")
    ```

---

## `bpy.types.SoundStrip`

**Important Note on Accessing Strips:**
In recent Blender versions, the preferred way to access sequence strips via the Python API is through the `bpy.context.scene.sequence_editor.strips` collection for top-level strips, and `bpy.context.scene.sequence_editor.strips_all` for all strips (including those within meta-strips). The older attributes `sequences` and `sequences_all` are now deprecated. The examples below have been updated to use the `strips` attribute.

A `SoundStrip` is a type of sequence strip (`STRIP_TYPE_SOUND_RAM`) that plays audio in the Video Sequence Editor.

### Key Properties

| Property             | Type      | Description                                                                  | Range/Notes |
| ------------------- | --------- | ---------------------------------------------------------------------------- | ----------- |
| `name`              | `string`  | Unique name of the strip                                                     | |
| `type`              | `enum`    | Type of strip (will be `'SOUND'`)                                           | Read-only |
| `sound`             | `Sound`   | Reference to the sound data-block                                           | |
| `volume`            | `float`   | Playback volume multiplier                                                  | Default: 1.0 |
| `pan`               | `float`   | Stereo panning value                                                        | -2.0 to 2.0 |
| `sound_offset`      | `float`   | Time offset for the sound in seconds                                        | |
| `pitch`             | -         | **DEPRECATED** - No longer accessible in Python API                         | Use speed_factor instead |
| `speed_factor`      | `float`   | Playback speed multiplier                                                   | Default: 1.0 |
| `frame_start`       | `int`     | Start frame of the strip                                                    | |
| `frame_final_start` | `int`     | Start frame with handles and offsets applied                               | Read-only |
| `frame_final_end`   | `int`     | End frame with handles and offsets applied                                 | Read-only |
| `frame_duration`    | `int`     | Duration of the strip in frames                                            | Read-only |
| `channel`           | `int`     | The channel number this strip is on                                        | |
| `mute`              | `boolean` | Whether the strip is muted                                                  | |
| `lock`              | `boolean` | Whether the strip is locked                                                 | |
| `streamindex`       | `int`     | Stream index for multi-stream audio files                                  | |

### Animation Flags
The strip can have various flags set that indicate animation status:

```python
strip.flag & SEQ_AUDIO_VOLUME_ANIMATED  # Volume is animated
strip.flag & SEQ_AUDIO_PAN_ANIMATED     # Pan is animated
strip.flag & SEQ_AUDIO_DRAW_WAVEFORM    # Show waveform in the editor
```

### Sound Equalizer Modifier

Sound strips can have an equalizer modifier applied:

```python
class SoundEqualizerModifierData:
    modifier: SequenceModifierData
    graphics: ListBase  # Contains EQCurveMappingData for frequency control
```

### Examples for `bpy.types.SoundStrip`

These examples assume you have a Scene and its Sequence Editor ready.
```python
import bpy
import os

# Ensure you have an active scene with a sequence editor
scene = bpy.context.scene
if not scene.sequence_editor:
    scene.sequence_editor_create()
seq_editor = scene.sequence_editor

# Base directory for media files (XXX: Adjust this path)
media_dir = "/home/manuel/Movies/blender-movie-editor/" 
```

1.  **Adding a new sound strip from a filepath:**
    ```python
    # (Continuing from above setup)
    sound_file = os.path.join(media_dir, "SampleVideo_1280x720_2mb.mp4") # Can be a video file to extract audio
    
    if os.path.exists(sound_file):
        # This creates both the Sound data-block (if not existing for this path) and the SoundStrip
        new_sound_strip = seq_editor.strips.new_sound(
            name="MyNewAudio",
            filepath=sound_file,
            channel=1,
            frame_start=10
        )
        
        if new_sound_strip:
            print(f"Added SoundStrip: '{new_sound_strip.name}' on channel {new_sound_strip.channel}")
            # Accessing the filepath correctly:
            if new_sound_strip.sound:
                print(f"  Source File: {new_sound_strip.sound.filepath}")
            else:
                print(f"  Warning: SoundStrip '{new_sound_strip.name}' has no associated Sound data-block.")
            print(f"  Timeline Duration: {new_sound_strip.frame_final_duration} frames")
            print(f"  Volume: {new_sound_strip.volume}")
    else:
        print(f"Audio/Video file not found: {sound_file}")
    ```

2.  **Modifying an existing sound strip's properties:**
    ```python
    # (Continuing from above setup)
    strip_name_to_modify = "MyNewAudio" # Name of the strip added in the previous example

    if strip_name_to_modify in seq_editor.strips:
        sound_strip = seq_editor.strips[strip_name_to_modify]
        
        print(f"Modifying strip: '{sound_strip.name}'")
        sound_strip.volume = 0.5  # Set volume to 50%
        sound_strip.pan = -0.8    # Pan mostly to the left
        sound_strip.speed_factor = 1.2   # Increase speed slightly
        sound_strip.channel = 2   # Move to channel 2
        sound_strip.frame_start = 50 # Move start to frame 50
        
        # Trimming the start of the sound source by 1 second (assuming 25 FPS)
        # animation_offset_start is in frames of the source media
        # sound_offset is in seconds of the source media
        fps = scene.render.fps
        trim_start_seconds = 1.0
        sound_strip.sound_offset = trim_start_seconds 
        # Adjust frame_final_duration if you want to keep the perceived end point the same,
        # or let it shorten the strip.
        # original_media_duration_frames = sound_strip.sound.duration * fps # sound.duration is in seconds
        # new_final_duration = original_media_duration_frames - (trim_start_seconds * fps)
        # sound_strip.frame_final_duration = new_final_duration

        print(f"  New Volume: {sound_strip.volume}")
        print(f"  New Pan: {sound_strip.pan}")
        print(f"  New Speed Factor: {sound_strip.speed_factor}")
        print(f"  New Channel: {sound_strip.channel}")
        print(f"  New Frame Start: {sound_strip.frame_start}")
        print(f"  New Sound Offset (seconds): {sound_strip.sound_offset}")
        print(f"  New Timeline Duration: {sound_strip.frame_final_duration}")

    else:
        print(f"SoundStrip '{strip_name_to_modify}' not found.")
    ```

3.  **Accessing properties of the underlying `Sound` data-block from a `SoundStrip`:**
    ```python
    # (Continuing from above setup)
    strip_name_to_inspect = "MyNewAudio"

    if strip_name_to_inspect in seq_editor.strips:
        sound_strip = seq_editor.strips[strip_name_to_inspect]
        
        if sound_strip.sound: # Always check if the .sound attribute is valid
            sound_data = sound_strip.sound
            print(f"Inspecting sound data for strip '{sound_strip.name}':")
            print(f"  Sound Data Name: {sound_data.name}")
            print(f"  Filepath: {sound_data.filepath}")
            print(f"  Samplerate: {sound_data.samplerate} Hz")
            print(f"  Channels: {sound_data.channels}")
            print(f"  Is Packed: {sound_data.packed_file is not None}")
        else:
            print(f"Strip '{sound_strip.name}' does not have a linked Sound data-block.")
    else:
        print(f"SoundStrip '{strip_name_to_inspect}' not found.")
    ```
4. **Iterating through all sound strips and printing their source filepaths:**
    ```python
    # (Continuing from above setup)
    print("\n--- All Sound Strips and their Sources ---")
    for strip in seq_editor.strips:
        if strip.type == 'SOUND':
            print(f"Strip Name: '{strip.name}'")
            if strip.sound and strip.sound.filepath:
                print(f"  Source: {strip.sound.filepath}")
            elif strip.sound:
                print(f"  Source: (Sound data '{strip.sound.name}' has no filepath)")
            else:
                print(f"  Source: (No sound data linked)")
    ```

---

## Relationship and Common Pitfalls

-   **`SoundStrip.sound` is the Bridge**: The `.sound` property on a `SoundStrip` object is your gateway to the actual `Sound` data-block.
-   **Correct Filepath Access**: Always use `my_sound_strip.sound.filepath` to get the path to the audio file. `my_sound_strip.filepath` does not exist and will cause an `AttributeError`.
-   **Check for Existence**: Before accessing `strip.sound.some_attribute`, it's wise to check if `strip.sound` is not `None`, especially if strips could be misconfigured or newly created without a sound source yet.
-   **`frame_duration` vs. `frame_final_duration`**:
    -   `frame_duration`: Typically reflects the original length of the sound file in frames. For `SoundStrip`, this is often tied to the actual media length and may not be directly settable in the same way as an image strip.
    -   `frame_final_duration`: The duration of the strip as it appears on the timeline. This is what you edit to make the strip shorter or longer in the sequence.
-   **Trimming Sound**:
    -   `sound_offset` (in seconds): Use this to start playing the sound from a point *within* the source audio file. For example, `sound_strip.sound_offset = 2.0` will start playing the sound 2 seconds into the audio file.
    -   `animation_offset_start` / `animation_offset_end` (in frames): These properties, inherited from base Strip types, can also be used for trimming the source media. `animation_offset_start` defines how many frames to skip from the beginning of the source sound.
    - To change the strip's length on the timeline, modify `frame_final_duration`.

By using these properties and methods correctly, you can effectively manage and automate audio manipulation within Blender's Video Sequence Editor. 