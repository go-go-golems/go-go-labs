# context.area: SEQUENCE_EDITOR
# Don't remove the comment above. It's important.

"""
Blender VSE Python API Test Script - Chapter 2: Importing Media

This script demonstrates core concepts from Chapter 2 of the Blender VSE Python API guide,
focusing on programmatically importing media into the Video Sequence Editor (VSE).

Key concepts covered:
1.  **Accessing the Sequence Editor**: Ensuring the VSE is ready for use.
2.  **Adding Movie Strips**: Importing video files using `seq_editor.strips.new_movie()`.
    -   Blender automatically determines video length.
    -   Note: `new_movie()` does NOT automatically add the audio track; it must be added separately.
3.  **Adding Image Strips**: Importing still images using `seq_editor.strips.new_image()`.
    -   Default duration for images can be adjusted via `frame_final_duration`.
4.  **Adding Sound Strips**: Importing audio files using `seq_editor.strips.new_sound()`.
    -   Can be used for standalone audio or to import audio from a video file.
5.  **Image Sequences**: Programmatically adding a sequence of images as individual strips.
6.  **Direct Data API vs. Operators**: This script favors direct `new_*` methods for more script control,
    as opposed to `bpy.ops.sequencer.*_strip_add()` operators which mimic UI actions.

Usage:
    -   Run this script from Blender's Text Editor or Python Console.
    -   Ensure you are in the Video Editing workspace.
"""

import bpy # type: ignore
import os
from mathutils import * # type: ignore

# Common shortcuts for Blender data and context
D = bpy.data
C = bpy.context

def get_active_scene():
    """
    Get the currently active scene.
    As per Chapter 1, the VSE is associated with a scene.
    
    Returns:
        bpy.types.Scene: The active scene object.
    """
    return C.scene

def ensure_sequence_editor(scene=None):
    """
    Ensure a scene has a sequence editor, creating one if it doesn't exist.
    Chapter 1 explains that `scene.sequence_editor_create()` initializes the VSE
    data-block if it's not already present.
    
    Args:
        scene (bpy.types.Scene, optional): The scene to check. If None, uses active scene.
        
    Returns:
        bpy.types.SequenceEditor: The sequence editor for the scene.
    """
    if scene is None:
        scene = get_active_scene()
    
    if not scene.sequence_editor:
        scene.sequence_editor_create()
    
    return scene.sequence_editor

def add_movie_strip(seq_editor, filepath, channel=1, frame_start=1, name=None):
    """
    Add a movie strip to the sequence editor using `seq_editor.strips.new_movie()`.
    
    As detailed in Chapter 2.1 of "Blender Video Python API for Dummies", this method
    directly creates a `MovieStrip` data-block. Blender determines the strip's length
    from the video file.
    
    Important: This method does *not* automatically create a linked sound strip,
    even if the video file contains audio. Audio must be added as a separate `SoundStrip`.
    The alternative, `bpy.ops.sequencer.movie_strip_add()`, has a `sound=True` option.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor to add the strip to.
        filepath (str): Absolute or relative path to the video file.
        channel (int): The channel (track) on the timeline to place the strip.
        frame_start (int): The frame number on the timeline where the strip will begin.
        name (str, optional): Name for the strip. If None, derived from the filename.
        
    Returns:
        bpy.types.MovieStrip: The created movie strip object.
    """
    if name is None:
        name = os.path.splitext(os.path.basename(filepath))[0]
    
    movie_strip = seq_editor.strips.new_movie(
        name=name,
        filepath=filepath,
        channel=channel,
        frame_start=frame_start
    )
    
    print(f"Added movie strip: {movie_strip.name} (Type: {movie_strip.type})")
    print(f"  File: {movie_strip.filepath}")
    print(f"  Duration: {movie_strip.frame_duration} frames (Source Media Length)")
    print(f"  Timeline: Channel {movie_strip.channel}, Start Frame {movie_strip.frame_start}, End Frame {movie_strip.frame_final_end}")
    
    return movie_strip

def add_image_strip(seq_editor, filepath, channel=1, frame_start=1, duration=25, name=None):
    """
    Add a single image strip to the sequence editor using `seq_editor.strips.new_image()`.
    
    Chapter 2.2 of the guide explains that `new_image()` adds an `ImageStrip`.
    By default, a single image strip has a default length (e.g., 1 second).
    This function allows explicit setting of the duration via `image_strip.frame_final_duration`.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor.
        filepath (str): Path to the image file (e.g., .png, .jpg).
        channel (int): Channel number for the strip.
        frame_start (int): Start frame on the timeline.
        duration (int): Desired duration of the image strip in frames.
        name (str, optional): Name for the strip. Defaults to filename.
        
    Returns:
        bpy.types.ImageStrip: The created image strip object.
    """
    if name is None:
        name = os.path.splitext(os.path.basename(filepath))[0]
    
    image_strip = seq_editor.strips.new_image(
        name=name,
        filepath=filepath,
        channel=channel,
        frame_start=frame_start
    )
    
    # Set the duration of the image strip on the timeline
    image_strip.frame_final_duration = duration
    
    print(f"Added image strip: {image_strip.name} (Type: {image_strip.type})")
    # Blender 4.4 removed the direct `filepath` attribute from `ImageStrip`.
    # The file path must be composed from the strip's `directory` and the first element's filename.
    if hasattr(image_strip, "elements") and image_strip.elements:
        first_elem = image_strip.elements[0]
        # `directory` already ends with a path separator ("/"), but use os.path.join for safety.
        img_source_path = os.path.join(image_strip.directory, first_elem.filename)
        print(f"  File: {img_source_path}")
    else:
        print("  File: (unable to determine – no strip elements)")
    print(f"  Duration: {image_strip.frame_final_duration} frames (Set Timeline Length)")
    print(f"  Timeline: Channel {image_strip.channel}, Start Frame {image_strip.frame_start}, End Frame {image_strip.frame_final_end}")
    
    return image_strip

def add_sound_strip(seq_editor, filepath, channel=1, frame_start=1, name=None):
    """
    Add a sound strip to the sequence editor using `seq_editor.strips.new_sound()`.
    
    As per Chapter 2.3, this creates a `SoundStrip`. The length is determined
    by the audio file's duration. This function can be used to add standalone audio
    or to extract and add the audio track from a video file (by providing the video file path).
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor.
        filepath (str): Path to the audio file (e.g., .mp3, .wav) or a video file to extract audio from.
        channel (int): Channel number for the strip.
        frame_start (int): Start frame on the timeline.
        name (str, optional): Name for the strip. Defaults to filename.
        
    Returns:
        bpy.types.SoundStrip: The created sound strip object.
    """
    if name is None:
        name = os.path.splitext(os.path.basename(filepath))[0]
    
    sound_strip = seq_editor.strips.new_sound(
        name=name,
        filepath=filepath,
        channel=channel,
        frame_start=frame_start
    )
    
    print(f"--- Introspecting SoundStrip: {sound_strip.name} ---")
    # Attempt to access filepath via sound_strip.sound.filepath
    actual_filepath = "N/A (Sound object not found or filepath missing)"
    if hasattr(sound_strip, 'sound') and sound_strip.sound:
        if hasattr(sound_strip.sound, 'filepath'):
            actual_filepath = sound_strip.sound.filepath
        else:
            actual_filepath = "N/A (sound_strip.sound has no filepath attribute)"
    print(f"  Attempting to get filepath via sound_strip.sound.filepath: {actual_filepath}")

    if hasattr(sound_strip, 'bl_rna'):
        print("RNA Properties:")
        for prop_name in sound_strip.bl_rna.properties.keys():
            prop = sound_strip.bl_rna.properties[prop_name]
            prop_value_str = "N/A"
            try:
                prop_value = getattr(sound_strip, prop_name)
                if prop_name == 'sound' and hasattr(prop_value, 'filepath'):
                     prop_value_str = f"{prop_value} (contains .filepath: {getattr(prop_value, 'filepath', 'Error accessing nested filepath')})"
                elif hasattr(prop_value, 'filepath'): # General check for other nested filepaths
                    prop_value_str = f"{prop_value} (contains .filepath: {getattr(prop_value, 'filepath', 'Error accessing nested filepath')})"
                else:
                    prop_value_str = str(prop_value)

            except AttributeError:
                prop_value_str = "<AttributeError>"
            except Exception as e:
                prop_value_str = f"<Exception: {e}>"
            
            print(f"  - {prop_name} (Type: {prop.type}, Value: {prop_value_str})")
    else:
        print("No bl_rna found. Attributes via dir():")
        for attr_name in dir(sound_strip):
            if not attr_name.startswith('_'):
                attr_value_str = "N/A"
                try:
                    attr_value = getattr(sound_strip, attr_name)
                    if attr_name == 'sound' and hasattr(attr_value, 'filepath'):
                         attr_value_str = f"{attr_value} (contains .filepath: {getattr(attr_value, 'filepath', 'Error accessing nested filepath')})"
                    elif hasattr(attr_value, 'filepath'): # General check for other nested filepaths
                        attr_value_str = f"{attr_value} (contains .filepath: {getattr(attr_value, 'filepath', 'Error accessing nested filepath')})"
                    else:
                        attr_value_str = str(attr_value)

                except AttributeError:
                    attr_value_str = "<AttributeError>"
                except Exception as e:
                    attr_value_str = f"<Exception: {e}>"
                print(f"  - {attr_name} (Value: {attr_value_str})")
    print(f"--- End Introspection ---")

    print(f"Added sound strip: {sound_strip.name} (Type: {sound_strip.type})")
    # Use sound_strip.sound.filepath as per Blender API documentation
    if hasattr(sound_strip, 'sound') and sound_strip.sound and hasattr(sound_strip.sound, 'filepath'):
        print(f"  File: {sound_strip.sound.filepath}")
    else:
        print(f"  File: {filepath} (Fallback: unable to access sound_strip.sound.filepath)")
    print(f"  Duration: {sound_strip.frame_duration} frames (Source Media Length)")
    # `pitch` was deprecated and removed from the Python API, but `pan` and `volume` remain.
    print(f"  Volume: {sound_strip.volume}, Pan: {sound_strip.pan}")
    
    return sound_strip

def add_image_sequence(seq_editor, directory, pattern="*.png", channel=1, frame_start=1, duration_per_image=25):
    """
    Add a sequence of images from a directory as individual, consecutive image strips.
    
    Chapter 2.2 provides an example of manual sequencing. This function automates
    finding images in a directory matching a pattern (e.g., "frame_*.png") and
    adding each as an `ImageStrip` with a specified duration, placed one after another.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor.
        directory (str): Path to the directory containing the image files.
        pattern (str): Glob pattern to match image files (e.g., "*.png", "render_output_####.jpg").
        channel (int): Channel number for the image strips.
        frame_start (int): Start frame on the timeline for the first image.
        duration_per_image (int): Duration in frames for each individual image strip.
        
    Returns:
        list[bpy.types.ImageStrip]: A list of the created image strip objects.
    """
    import glob # Keep import local as it's specific to this function's logic
    
    # Construct the full search path for glob
    search_path = os.path.join(directory, pattern)
    # Find and sort image files to ensure correct order
    image_files = sorted(glob.glob(search_path))
    
    if not image_files:
        print(f"No image files found in '{directory}' matching pattern '{pattern}'.")
        return []
    
    print(f"\nAttempting to add image sequence from '{directory}' (pattern: '{pattern}')")
    created_strips = []
    current_timeline_frame = frame_start
    
    for img_filepath in image_files:
        # Use the existing add_image_strip function for each image
        img_strip = add_image_strip(
            seq_editor=seq_editor,
            filepath=img_filepath,
            channel=channel,
            frame_start=current_timeline_frame,
            duration=duration_per_image
            # Name will be auto-generated by add_image_strip
        )
        created_strips.append(img_strip)
        # Advance the timeline for the next strip
        current_timeline_frame += duration_per_image 
        
    if created_strips:
        print(f"Successfully added {len(created_strips)} images as a sequence on channel {channel}.")
    else:
        print("No image strips were added for the sequence.")
        
    return created_strips

def main():
    """
    Main function demonstrating Chapter 2 concepts: Importing Media.
    This function sets up the VSE and then calls various strip-adding functions.
    """
    print("Blender VSE Python API Test - Chapter 2: Importing Media")
    print("========================================================")
    
    scene = get_active_scene()
    seq_editor = ensure_sequence_editor(scene)
    
    print(f"\nOperating on scene: '{scene.name}' with Sequence Editor: {seq_editor}")
    
    # --- Configuration ---
    test_media_dir = "/home/manuel/Movies/blender-movie-editor"
    
    current_channel = 1
    current_frame = 1
    frame_padding = 10 # Add some padding between different types of imports for clarity

    # --- Name Channels for Clarity ---
    # Set names for the first few channels (adjust as needed for your project)
    channel_names = [
        "Main Video",   # Channel 1
        "Audio",        # Channel 2
        "Music",        # Channel 3
        "Images",       # Channel 4
        "FX",           # Channel 5
        "Titles"        # Channel 6
    ]

    # --- 1. Add Movie Strips ---
    print(f"\n--- Adding Movie Strips ---")
    video_files = [
        "SampleVideo_1280x720_2mb.mp4",
        "SampleVideo_1280x720_1mb.mp4",
        "SampleVideo_1280x720_5mb.mp4"
    ]
    
    for video_filename in video_files:
        video_path = os.path.join(test_media_dir, video_filename)
        if os.path.exists(video_path):
            # Name the video channel
            if current_channel < len(seq_editor.channels):
                seq_editor.channels[current_channel].name = f"Video: {os.path.splitext(video_filename)[0]}"
            
            movie_strip = add_movie_strip(seq_editor, video_path, channel=current_channel, frame_start=current_frame)
            movie_strip.name = f"Movie: {os.path.splitext(video_filename)[0]}"
            
            # Name the audio channel and add audio strip
            if current_channel + 1 < len(seq_editor.channels):
                seq_editor.channels[current_channel+1].name = f"Audio: {os.path.splitext(video_filename)[0]}"
            
            audio_strip = add_sound_strip(seq_editor, video_path, channel=current_channel + 1, frame_start=current_frame)
            audio_strip.name = f"Audio: {os.path.splitext(video_filename)[0]}"
            
            current_frame += movie_strip.frame_duration + frame_padding
            current_channel += 2
        else:
            print(f"Movie file not found: {video_path}. Skipping movie strip import.")
    
    # --- 2. Add Extracted Audio Strip ---
    print(f"\n--- Adding Extracted Audio Strip ---")
    audio_path = os.path.join(test_media_dir, "extracted/audio/sample_2mb.aac")
    if os.path.exists(audio_path):
        # Name the music channel
        if current_channel < len(seq_editor.channels):
            seq_editor.channels[current_channel].name = "Music Track"
            
        audio_strip = add_sound_strip(seq_editor, audio_path, channel=current_channel, frame_start=1)
        audio_strip.name = f"Music: {os.path.splitext(os.path.basename(audio_path))[0]}"
        current_channel += 1
    else:
        print(f"Audio file not found: {audio_path}. Skipping audio strip import.")

    # --- 3. Add Image Sequence ---
    print(f"\n--- Adding Image Sequence ---")
    frames_dir = os.path.join(test_media_dir, "extracted/frames")
    
    # Name the image sequence channel
    if current_channel < len(seq_editor.channels):
        seq_editor.channels[current_channel].name = "Image Sequence"
        
    image_strips = add_image_sequence(
        seq_editor=seq_editor,
        directory=frames_dir,
        pattern="frame_*.png",
        channel=current_channel,
        frame_start=1,
        duration_per_image=25  # Each frame lasts 1 second (25 frames)
    )
    
    # Name the image sequence strips
    for i, strip in enumerate(image_strips):
        # Get the filepath using the new Blender 4.4+ method
        if hasattr(strip, "elements") and strip.elements:
            first_elem = strip.elements[0]
            img_source_path = os.path.join(strip.directory, first_elem.filename)
            strip.name = f"Frame {i+1}: {os.path.splitext(os.path.basename(img_source_path))[0]}"
        else:
            strip.name = f"Frame {i+1}: (Unable to determine filename)"
        
    # --- Recap Added Strips ---
    print(f"\n--- Recap of All Strips Added ---")
    if seq_editor.strips:
        print("\nChannel Configuration:")
        for idx, ch in enumerate(seq_editor.channels, start=1):
            print(f"Channel {idx}: {ch.name}")
            
        print("\nStrip Configuration:")
        for strip in seq_editor.strips_all:
            print(f"- Name: '{strip.name}', Type: {strip.type}, Channel: {strip.channel}, "
                  f"Start: {strip.frame_start}, End: {strip.frame_final_end}, Duration: {strip.frame_final_duration}")
    else:
        print("No strips were added to the sequence editor.")

    print(f"\nTest for Chapter 2 completed! Review the VSE timeline and console output.")

main() 