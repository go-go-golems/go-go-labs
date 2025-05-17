# VSE Utility functions for Blender Python scripts
# Contains reusable functions for Video Sequence Editor operations

import bpy
import os
from pathlib import Path

# ----- Core VSE Setup Functions -----

def get_active_scene():
    """Get the currently active scene."""
    return bpy.context.scene

def ensure_sequence_editor(scene=None):
    """Ensure a scene has a sequence editor, creating one if it doesn't exist."""
    if scene is None:
        scene = get_active_scene()
    
    if not scene.sequence_editor:
        scene.sequence_editor_create()
    
    return scene.sequence_editor

# ----- Media Handling Functions -----

def add_movie_strip(seq_editor, filepath, channel=1, frame_start=1, name=None):
    """Add a movie strip to the sequence editor."""
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
    print(f"  Duration: {movie_strip.frame_duration} frames")
    print(f"  Timeline: Channel {movie_strip.channel}, Start Frame {movie_strip.frame_start}, End Frame {movie_strip.frame_final_end}")
    
    return movie_strip

def add_sound_strip(seq_editor, filepath, channel=1, frame_start=1, name=None):
    """Add a sound strip to the sequence editor."""
    if name is None:
        name = os.path.splitext(os.path.basename(filepath))[0]
    
    sound_strip = seq_editor.strips.new_sound(
        name=name,
        filepath=filepath,
        channel=channel,
        frame_start=frame_start
    )
    
    print(f"Added sound strip: {sound_strip.name} (Type: {sound_strip.type})")
    if hasattr(sound_strip, 'sound') and sound_strip.sound and hasattr(sound_strip.sound, 'filepath'):
        print(f"  File: {sound_strip.sound.filepath}")
    else:
        print(f"  File: {filepath}")
    print(f"  Duration: {sound_strip.frame_duration} frames")
    print(f"  Volume: {sound_strip.volume}, Pan: {sound_strip.pan}")
    
    return sound_strip

# ----- Scene Management Functions -----

def clear_all_strips(seq_editor):
    """Remove all strips from the sequence editor."""
    if seq_editor.strips_all:
        print(f"Removing {len(seq_editor.strips_all)} existing strips...")
        # Create a copy of the list to avoid modification during iteration issues
        strips_to_remove = list(seq_editor.strips_all)
        for strip in strips_to_remove:
            seq_editor.sequences.remove(strip)
        print("All existing strips removed.")
    return True

def print_sequence_info(seq_editor, title="Current Sequence State"):
    """Print detailed information about all strips in the sequence editor."""
    print(f"\n--- {title} ---")
    if not seq_editor.strips_all:
        print("  No strips in sequence editor.")
        return
    
    print(f"Total strips: {len(seq_editor.strips_all)}")
    
    for i, strip in enumerate(seq_editor.strips_all, 1):
        print(f"{i}. '{strip.name}' (Type: {strip.type}):")
        print(f"   Channel: {strip.channel}")
        print(f"   Timeline: Frames {strip.frame_start}-{strip.frame_final_end} " +
              f"(Duration: {strip.frame_final_duration})")
        print(f"   Offsets: start={strip.frame_offset_start}, end={strip.frame_offset_end}")
        
        if strip.type == 'MOVIE':
            print(f"   Source: {strip.filepath}")
        elif strip.type == 'SOUND':
            print(f"   Volume: {strip.volume}, Pan: {strip.pan}")
        print()

# ----- Technical Functions -----

def check_and_set_fps(seq_editor, scene):
    """Check FPS of all strips and ensure they match the scene FPS.
    If strips have different FPS, use the most common one and set scene to match.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor to check
        scene (bpy.types.Scene): The scene to check/modify
        
    Returns:
        tuple: (bool, str) - (success, message)
    """
    if not seq_editor or not seq_editor.sequences_all:
        return True, "No sequences to check"
    
    # Collect FPS information from movie strips
    fps_counts = {}
    strip_fps = {}
    
    for strip in seq_editor.sequences_all:
        if strip.type == 'MOVIE':
            # Get source FPS using the new Blender 4.4 method
            src_fps = None
            if hasattr(strip.elements[0], 'orig_fps') and strip.elements[0].orig_fps:
                src_fps = strip.elements[0].orig_fps
            elif hasattr(strip, 'fps'):
                src_fps = strip.fps
                
            if src_fps:
                strip_fps[strip.name] = src_fps
                fps_counts[src_fps] = fps_counts.get(src_fps, 0) + 1
    
    if not fps_counts:
        return True, "No movie strips found"
    
    # Find most common FPS
    most_common_fps = max(fps_counts.items(), key=lambda x: x[1])[0]
    
    # Check if all strips have the same FPS
    if len(fps_counts) > 1:
        print("\nWARNING: Different FPS detected in strips:")
        for strip_name, fps in strip_fps.items():
            print(f"  u2022 {strip_name}: {fps} fps")
        print(f"\nUsing most common FPS: {most_common_fps}")
    
    # Check scene FPS
    scene_fps = scene.render.fps / scene.render.fps_base
    
    if abs(scene_fps - most_common_fps) > 0.01:  # Allow small floating-point differences
        print(f"\nAdjusting scene FPS from {scene_fps} to {most_common_fps}")
        scene.render.fps = int(most_common_fps)
        scene.render.fps_base = 1.0
        return True, f"Scene FPS adjusted to {most_common_fps}"
    
    return True, f"All good - Scene and strips using {scene_fps} fps"

def find_test_media_dir(default_path="/home/manuel/Movies/blender-movie-editor"):
    """Find a valid test media directory with test videos.
    
    Returns:
        str: Path to a directory containing test media files
    """
    # First check the default path
    if os.path.exists(default_path):
        return default_path
    
    # Try to find media directory relative to script location
    script_dir = os.path.dirname(os.path.abspath(__file__))
    alternative_paths = [
        os.path.join(script_dir, "../media"),
        os.path.join(script_dir, "media"),
        "/tmp/blender-test-media"
    ]
    
    for path in alternative_paths:
        if os.path.exists(path):
            print(f"Using alternative media path: {path}")
            return path
    
    # If no existing directory is found, create a default one
    print(f"Warning: Media directory not found. Using: {default_path}")
    os.makedirs(default_path, exist_ok=True)
    return default_path

def setup_test_sequence(seq_editor, test_media_dir, video_files=None):
    """Set up a test sequence with video clips for demonstration.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor to add strips to
        test_media_dir (str): Directory containing test media files
        video_files (list, optional): List of video filenames to use
        
    Returns:
        list: A list of (video_strip, audio_strip) tuples for the added clips
    """
    print("\nSetting up test sequence with video clips...")
    
    # Clear existing strips first
    clear_all_strips(seq_editor)
    
    # Default list of test video files to use
    if video_files is None:
        video_files = [
            "SampleVideo_1280x720_2mb.mp4",
            "SampleVideo_1280x720_1mb.mp4",
            "SampleVideo_1280x720_5mb.mp4"
        ]
    
    # Track channels for video and audio
    video_channel = 1
    audio_channel = 2
    
    # Start with frame 1 and add some padding between clips
    current_frame = 1
    frame_padding = 0  # No padding - place clips back-to-back
    
    # Name the channels
    if video_channel <= len(seq_editor.channels):
        seq_editor.channels[video_channel-1].name = "Video"
    if audio_channel <= len(seq_editor.channels):
        seq_editor.channels[audio_channel-1].name = "Audio"
    
    # Add clips and store them for later use
    added_clips = []
    
    for i, video_filename in enumerate(video_files):
        video_path = os.path.join(test_media_dir, video_filename)
        if os.path.exists(video_path):
            # Add video strip
            video_strip = add_movie_strip(
                seq_editor, 
                video_path, 
                channel=video_channel, 
                frame_start=current_frame,
                name=f"Clip{i+1}"
            )
            
            # Add matching audio strip
            audio_strip = add_sound_strip(
                seq_editor, 
                video_path, 
                channel=audio_channel, 
                frame_start=current_frame,
                name=f"Audio{i+1}"
            )
            
            # Store the pair
            added_clips.append((video_strip, audio_strip))
            
            # Update frame position for next clip
            current_frame = video_strip.frame_final_end + frame_padding
        else:
            print(f"Video file not found: {video_path}")
    
    print(f"Added {len(added_clips)} clips to the sequence.")
    
    # Check and set FPS after adding all clips
    success, message = check_and_set_fps(seq_editor, get_active_scene())
    print(f"\nFPS Check: {message}")
    
    return added_clips

# ----- Diagnostics Functions -----

def print_strip_details(strip, label="Strip"):
    """Print very detailed information about a strip, including internal properties."""
    print(f"\n{label}: '{strip.name}' (Type: {strip.type}, Ptr: {strip.as_pointer()})")
    print(f"  Position: Channel {strip.channel}, Frames {strip.frame_start}-{strip.frame_final_end}")
    print(f"  frame_start:        {strip.frame_start}")
    print(f"  frame_final_start:  {strip.frame_final_start}")
    print(f"  frame_duration:     {strip.frame_duration}")
    print(f"  frame_final_duration: {strip.frame_final_duration}")
    print(f"  frame_offset_start: {strip.frame_offset_start}")
    print(f"  frame_offset_end:   {strip.frame_offset_end}")
    print(f"  frame_still_start:  {strip.frame_still_start if hasattr(strip, 'frame_still_start') else 'N/A'}")
    print(f"  frame_still_end:    {strip.frame_still_end if hasattr(strip, 'frame_still_end') else 'N/A'}")