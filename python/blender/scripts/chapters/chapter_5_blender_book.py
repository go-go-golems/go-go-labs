# context.area: SEQUENCE_EDITOR
# Don't remove the comment above. It's important.

"""
Blender VSE Python API Test Script - Chapter 5: Effects and Adjustments

This script demonstrates core concepts from Chapter 5 of the Blender VSE Python API guide,
focusing on programmatically applying effects and adjustments to clips in the VSE.

Key concepts covered:
1.  **Transform Effects**: Adjusting position, scale, and rotation of clips
2.  **Color Effects**: Applying color adjustments using modifiers
3.  **Speed Control**: Changing playback speed of clips
4.  **Text Overlays**: Adding text titles and captions
5.  **Picture-in-Picture**: Creating PIP compositions
6.  **Glow Effects**: Applying glow/blur effects to clips

Usage:
    -   Run this script from Blender's Text Editor or Python Console.
    -   Ensure you are in the Video Editing workspace.
    -   The script creates a basic sequence and then demonstrates various effects.
"""

import bpy # type: ignore
import os
import sys
from math import radians

# Common shortcuts for Blender data and context
D = bpy.data
C = bpy.context

# Try to import utilities - first make sure our utilities path is in sys.path
scripts_dir = os.path.dirname(os.path.abspath(__file__))
utils_dir = os.path.join(os.path.dirname(scripts_dir), 'utils')
if utils_dir not in sys.path:
    sys.path.append(utils_dir)

# Now try to import our utilities, with fallback implementations if import fails
try:
    from effect_utils import (
        apply_transform_effect, create_picture_in_picture, apply_speed_effect,
        create_text_overlay, apply_color_balance, apply_glow_effect
    )
    print("Successfully imported effect utilities")
except ImportError as e:
    print(f"Warning: Could not import effect_utils: {e}")
    print("Using built-in fallback implementations")
    
    # --- Fallback utility implementations ---
    def apply_transform_effect(seq_editor, strip, offset_x=0, offset_y=0, scale_x=1.0, scale_y=1.0, rotation=0.0, channel=None):
        """Apply a transform effect to a strip"""
        if channel is None:
            channel = strip.channel + 1
        
        transform = seq_editor.strips.new_effect(
            name=f"Transform_{strip.name}",
            type='TRANSFORM',
            channel=channel,
            frame_start=strip.frame_start,
            frame_end=strip.frame_final_end,
            seq1=strip
        )
        
        transform.transform.offset_x = offset_x
        transform.transform.offset_y = offset_y
        transform.transform.scale_x = scale_x
        transform.transform.scale_y = scale_y
        transform.transform.rotation = radians(rotation)  # Convert degrees to radians
        
        return transform
    
    def create_picture_in_picture(seq_editor, main_strip, pip_strip, pip_scale=0.3, position='top-right', channel=None):
        """Create a picture-in-picture effect"""
        if channel is None:
            channel = max(main_strip.channel, pip_strip.channel) + 1
        
        # Calculate position based on specified corner
        width = C.scene.render.resolution_x
        height = C.scene.render.resolution_y
        padding = 20
        
        if position == 'top-right':
            offset_x = width/2 - (width * pip_scale)/2 - padding
            offset_y = height/2 - (height * pip_scale)/2 - padding
        elif position == 'top-left':
            offset_x = -(width/2 - (width * pip_scale)/2 - padding)
            offset_y = height/2 - (height * pip_scale)/2 - padding
        elif position == 'bottom-right':
            offset_x = width/2 - (width * pip_scale)/2 - padding
            offset_y = -(height/2 - (height * pip_scale)/2 - padding)
        elif position == 'bottom-left':
            offset_x = -(width/2 - (width * pip_scale)/2 - padding)
            offset_y = -(height/2 - (height * pip_scale)/2 - padding)
        else:  # 'center'
            offset_x = 0
            offset_y = 0
        
        return apply_transform_effect(
            seq_editor=seq_editor,
            strip=pip_strip,
            offset_x=offset_x,
            offset_y=offset_y,
            scale_x=pip_scale,
            scale_y=pip_scale,
            channel=channel
        )
    
    def apply_speed_effect(seq_editor, strip, speed_factor, channel=None):
        """Apply a speed effect to a strip"""
        if channel is None:
            channel = strip.channel + 1
        
        # Calculate new duration based on speed factor
        original_duration = strip.frame_final_duration
        new_duration = int(original_duration / speed_factor)
        
        speed_effect = seq_editor.strips.new_effect(
            name=f"Speed_{strip.name}_{speed_factor}x",
            type='SPEED',
            channel=channel,
            frame_start=strip.frame_start,
            frame_end=strip.frame_start + new_duration,
            seq1=strip
        )
        
        # Set stretch to input strip length if the property exists
        if hasattr(speed_effect, 'use_scale_to_length'):
            speed_effect.use_scale_to_length = True
        
        return speed_effect
    
    def create_text_overlay(seq_editor, text, frame_start, frame_end, channel=10,
                          position='center', size=50, color=(1,1,1,1)):
        """Create a text overlay"""
        text_strip = seq_editor.strips.new_effect(
            name=f"Text_{text[:10]}",
            type='TEXT',
            channel=channel,
            frame_start=frame_start,
            frame_end=frame_end
        )
        
        text_strip.text = text
        text_strip.font_size = size
        text_strip.color = color[:3]  # RGB only
        
        if position == 'center':
            text_strip.location = (0.5, 0.5)
        elif position == 'top':
            text_strip.location = (0.5, 0.8)
        elif position == 'bottom':
            text_strip.location = (0.5, 0.2)
        
        if hasattr(text_strip, 'align_x'):
            text_strip.align_x = 'CENTER'
        
        return text_strip
    
    def apply_color_balance(seq_editor, strip, lift=(1,1,1), gamma=(1,1,1), gain=(1,1,1)):
        """Apply color balance adjustments to a strip"""
        try:
            mod = strip.modifiers.new(name="ColorBalance", type='COLOR_BALANCE')
            mod.color_balance.lift = lift
            mod.color_balance.gamma = gamma
            mod.color_balance.gain = gain
            return mod
        except Exception as e:
            print(f"Error applying color balance: {e}")
            return None
    
    def apply_glow_effect(seq_editor, strip, threshold=0.5, blur_radius=3.0, quality=0.5, channel=None):
        """Apply a glow effect to a strip"""
        if channel is None:
            channel = strip.channel + 1
        
        glow = seq_editor.strips.new_effect(
            name=f"Glow_{strip.name}",
            type='GLOW',
            channel=channel,
            frame_start=strip.frame_start,
            frame_end=strip.frame_final_end,
            seq1=strip
        )
        
        if hasattr(glow, 'threshold'):
            glow.threshold = threshold
        if hasattr(glow, 'blur_radius'):
            glow.blur_radius = blur_radius
        if hasattr(glow, 'quality'):
            glow.quality = quality
        
        return glow

# --- Core utility functions from Chapter 2 & 3 ---

def get_active_scene():
    """Get the currently active scene."""
    return C.scene

def ensure_sequence_editor(scene=None):
    """Ensure a scene has a sequence editor, creating one if it doesn't exist."""
    if scene is None:
        scene = get_active_scene()
    
    if not scene.sequence_editor:
        scene.sequence_editor_create()
    
    return scene.sequence_editor

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

def check_and_set_fps(seq_editor, scene):
    """Check FPS of strips and ensure it matches the scene FPS."""
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

def setup_test_sequence(seq_editor, test_media_dir):
    """Set up a test sequence with video clips for demonstration."""
    print("\nSetting up test sequence with video clips...")
    
    # Clear all existing strips first
    if seq_editor.sequences_all:
        print(f"Removing {len(seq_editor.sequences_all)} existing strips...")
        for strip in seq_editor.sequences_all:
            seq_editor.sequences.remove(strip)
        print("All existing strips removed.")
    
    # List of test video files to use
    video_files = [
        "SampleVideo_1280x720_2mb.mp4",
        "SampleVideo_1280x720_1mb.mp4",
        "SampleVideo_1280x720_5mb.mp4"
    ]
    
    # Track channels for video and audio
    video_channel = 1
    audio_channel = 2
    
    # Start with frame 1 and add spacing between clips
    current_frame = 1
    frame_padding = 10  # Space between clips
    
    # Name the channels
    if video_channel < len(seq_editor.channels):
        seq_editor.channels[video_channel-1].name = "Video"
    if audio_channel < len(seq_editor.channels):
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
    success, message = check_and_set_fps(seq_editor, C.scene)
    print(f"\nFPS Check: {message}")
    
    return added_clips

# --- Chapter 5 Demo Functions ---

def demonstrate_transform_effects(seq_editor, clips):
    """
    Demonstrate applying transform effects to clips.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
    """
    print("\n--- Demonstrating Transform Effects ---")
    
    # Need at least one clip
    if not clips:
        print("Need at least 1 clip to demonstrate transform effects. Skipping.")
        return
    
    # Get the first clip
    video1, _ = clips[0]
    
    # Apply the transform effect - scale down slightly and move up
    transform = apply_transform_effect(
        seq_editor=seq_editor,
        strip=video1,
        offset_x=0,  # No horizontal change
        offset_y=50,  # Move up 50 pixels
        scale_x=0.9,  # Scale to 90% width
        scale_y=0.9,  # Scale to 90% height
        rotation=5.0,  # Rotate 5 degrees
        channel=5     # Place on channel 5
    )
    
    print(f"Applied transform effect to '{video1.name}'")
    print(f"  Positioned at offset (0, 50) pixels")
    print(f"  Scaled to 90% of original size")
    print(f"  Rotated 5 degrees")
    print(f"  Effect placed on channel {transform.channel}")

def demonstrate_picture_in_picture(seq_editor, clips):
    """
    Demonstrate creating a picture-in-picture effect with two clips.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
    """
    print("\n--- Demonstrating Picture-in-Picture ---")
    
    # Need at least two clips
    if len(clips) < 2:
        print("Need at least 2 clips to demonstrate picture-in-picture. Skipping.")
        return
    
    # Get two clips
    video1, _ = clips[0]  # Main clip
    video2, _ = clips[1]  # PIP clip
    
    # Make sure video2 is at least as long as video1 for simplicity
    if video2.frame_final_end < video1.frame_final_end:
        print(f"Note: Adjusting second clip to match first clip's duration")
        # Extend video2's end to match video1
        # This is just for demo purposes, in real usage might need more care
        video2.frame_start = video1.frame_start
    
    # Create the picture-in-picture effect
    pip_effect = create_picture_in_picture(
        seq_editor=seq_editor,
        main_strip=video1,
        pip_strip=video2,
        pip_scale=0.3,   # 30% of original size
        position='top-right',
        channel=6
    )
    
    print(f"Created picture-in-picture effect")
    print(f"  Main clip: '{video1.name}'")
    print(f"  PIP clip: '{video2.name}' scaled to 30% in top-right corner")
    print(f"  Effect placed on channel {pip_effect.channel}")

def demonstrate_speed_effects(seq_editor, clips):
    """
    Demonstrate speed control effects.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
    """
    print("\n--- Demonstrating Speed Effects ---")
    
    # Need at least one clip
    if not clips:
        print("Need at least 1 clip to demonstrate speed effects. Skipping.")
        return
    
    # We'll need a fresh clip for this to avoid affecting other demos
    # Let's duplicate a clip and add it at the end
    last_video, last_audio = clips[-1]
    start_frame = last_video.frame_final_end + 50  # Add some spacing
    
    # Create a new strip using the same file
    slow_video = add_movie_strip(
        seq_editor=seq_editor,
        filepath=last_video.filepath,
        channel=last_video.channel,
        frame_start=start_frame,
        name="SlowClip"
    )
    
    # Also add a new strip for fast motion demo
    fast_video = add_movie_strip(
        seq_editor=seq_editor,
        filepath=last_video.filepath,
        channel=last_video.channel,
        frame_start=slow_video.frame_final_end + 50,  # After the slow clip
        name="FastClip"
    )
    
    # Create a slow-motion effect (0.5x speed)
    slow_effect = apply_speed_effect(
        seq_editor=seq_editor,
        strip=slow_video,
        speed_factor=0.5,  # Half speed (slow motion)
        channel=7
    )
    
    print(f"Applied slow-motion effect (0.5x) to '{slow_video.name}'")
    print(f"  Original duration: {slow_video.frame_final_duration} frames")
    print(f"  New duration: {slow_effect.frame_final_duration} frames")
    print(f"  Effect placed on channel {slow_effect.channel}")
    
    # Create a fast-motion effect (2x speed)
    fast_effect = apply_speed_effect(
        seq_editor=seq_editor,
        strip=fast_video,
        speed_factor=2.0,  # Double speed
        channel=7
    )
    
    print(f"Applied fast-motion effect (2.0x) to '{fast_video.name}'")
    print(f"  Original duration: {fast_video.frame_final_duration} frames")
    print(f"  New duration: {fast_effect.frame_final_duration} frames")
    print(f"  Effect placed on channel {fast_effect.channel}")

def demonstrate_text_overlays(seq_editor, clips):
    """
    Demonstrate text overlays and titles.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
    """
    print("\n--- Demonstrating Text Overlays ---")
    
    # Need at least one clip as reference for timeline
    if not clips:
        print("Need at least 1 clip as reference for text placement. Skipping.")
        return
    
    # Get reference to first and last clip for timeline reference
    first_video, _ = clips[0]
    last_video, _ = clips[-1]
    
    # Create a main title at the beginning
    title = create_text_overlay(
        seq_editor=seq_editor,
        text="Chapter 5: Effects and Adjustments",
        frame_start=first_video.frame_start,
        frame_end=first_video.frame_start + 90,  # 3-4 second title at 24-30fps
        channel=10,
        position='center',
        size=70,
        color=(1, 0.8, 0.2, 1)  # Golden yellow
    )
    
    print(f"Created main title overlay")
    print(f"  Text: '{title.text}'")
    print(f"  Duration: {title.frame_final_duration} frames")
    print(f"  Placed on channel {title.channel}")
    
    # Create a lower third subtitle for each clip
    for i, (video, _) in enumerate(clips):
        subtitle = create_text_overlay(
            seq_editor=seq_editor,
            text=f"Video Clip {i+1}",
            frame_start=video.frame_start + 10,  # Start a bit after clip starts
            frame_end=video.frame_start + 60,    # Short duration
            channel=9,
            position='bottom',
            size=40,
            color=(1, 1, 1, 1)  # White
        )
        
        print(f"Created lower third subtitle for '{video.name}'")
        print(f"  Text: '{subtitle.text}'")
    
    # Create an end credit
    credits = create_text_overlay(
        seq_editor=seq_editor,
        text="Created with Blender VSE Python API",
        frame_start=last_video.frame_final_end - 60,
        frame_end=last_video.frame_final_end,
        channel=10,
        position='bottom',
        size=50,
        color=(0.9, 0.9, 0.9, 1)  # Light gray
    )
    
    print(f"Created end credits overlay")
    print(f"  Text: '{credits.text}'")

def demonstrate_color_effects(seq_editor, clips):
    """
    Demonstrate color grading and effects.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
    """
    print("\n--- Demonstrating Color Effects ---")
    
    # Need at least one clip
    if not clips:
        print("Need at least 1 clip to demonstrate color effects. Skipping.")
        return
    
    # Apply color balance to the first clip
    video1, _ = clips[0]
    color_mod = apply_color_balance(
        seq_editor=seq_editor,
        strip=video1,
        lift=(1.1, 0.9, 0.9),    # Slightly reddish shadows
        gamma=(0.95, 1.05, 1.1),  # Slightly blue/green midtones
        gain=(1.0, 1.0, 1.1)      # Slightly blueish highlights
    )
    
    if color_mod:
        print(f"Applied color balance to '{video1.name}'")
        print(f"  Lift (shadows): (1.1, 0.9, 0.9) - reddish")
        print(f"  Gamma (midtones): (0.95, 1.05, 1.1) - blue/green tint")
        print(f"  Gain (highlights): (1.0, 1.0, 1.1) - slight blue tint")
    
    # If we have a second clip, apply a glow effect to it
    if len(clips) > 1:
        video2, _ = clips[1]
        glow = apply_glow_effect(
            seq_editor=seq_editor,
            strip=video2,
            threshold=0.7,     # Only brightest areas glow
            blur_radius=5.0,   # Moderate blur
            quality=0.8,       # High quality
            channel=8
        )
        
        print(f"Applied glow effect to '{video2.name}'")
        print(f"  Threshold: 0.7, Blur Radius: 5.0")
        print(f"  Effect placed on channel {glow.channel}")

def main():
    """
    Main function demonstrating Chapter 5 concepts: Effects and Adjustments.
    This function sets up a test sequence and then demonstrates various effects.
    """
    print("Blender VSE Python API Test - Chapter 5: Effects and Adjustments")
    print("=============================================================\n")
    
    scene = get_active_scene()
    seq_editor = ensure_sequence_editor(scene)
    
    print(f"Operating on scene: '{scene.name}' with Sequence Editor: {seq_editor}")
    
    # --- Configuration ---
    # Path to test media - try to find it in common locations
    test_media_dir = "/home/manuel/Movies/blender-movie-editor"
    
    # Check if path exists, if not look for alternative locations
    if not os.path.exists(test_media_dir):
        # Try relative path from current script
        script_dir = os.path.dirname(os.path.abspath(__file__))
        alternative_paths = [
            os.path.join(script_dir, "../media"),
            os.path.join(script_dir, "media"),
            "/tmp/blender-test-media"
        ]
        
        for path in alternative_paths:
            if os.path.exists(path):
                test_media_dir = path
                print(f"Using alternative media path: {test_media_dir}")
                break
        
        print(f"Warning: Media directory not found. Please download test videos to: {test_media_dir}")
        print("Sample videos can be downloaded from: https://sample-videos.com/")
    
    # --- 1. Set up a basic sequence with clips ---
    clips = setup_test_sequence(seq_editor, test_media_dir)
    
    if not clips:
        print("No clips were added. Make sure test_media_dir is correct.")
        return
    
    # --- 2. Demonstrate transform effects ---
    demonstrate_transform_effects(seq_editor, clips)
    
    # --- 3. Demonstrate picture-in-picture ---
    demonstrate_picture_in_picture(seq_editor, clips)
    
    # --- 4. Demonstrate speed effects ---
    demonstrate_speed_effects(seq_editor, clips)
    
    # --- 5. Demonstrate text overlays ---
    demonstrate_text_overlays(seq_editor, clips)
    
    # --- 6. Demonstrate color effects ---
    demonstrate_color_effects(seq_editor, clips)
    
    # --- Recap the Final State ---
    print("\n--- Final State of Strips ---")
    strip_count = 0
    for strip in seq_editor.strips_all:
        strip_count += 1
        if strip_count <= 10:  # Only show first 10 to avoid overwhelming output
            print(f"- '{strip.name}' (Type: {strip.type}): Channel {strip.channel}, "
                f"Frames {strip.frame_start}-{strip.frame_final_end}")
    
    if strip_count > 10:
        print(f"... and {strip_count - 10} more strips")
    
    print("\nTest for Chapter 5 completed! Review the VSE timeline and console output.")

# Run the main function to demonstrate Chapter 5 concepts
if __name__ == "__main__":
    main()