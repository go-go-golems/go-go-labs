# context.area: SEQUENCE_EDITOR
# Don't remove the comment above. It's important.

"""
Blender VSE Python API Test Script - Chapter 4: Transitions and Fades

This script demonstrates core concepts from Chapter 4 of the Blender VSE Python API guide,
focusing on programmatically creating transitions between clips in the VSE.

Key concepts covered:
1.  **Crossfades**: Creating dissolve transitions using CROSS effect strips
2.  **Gamma Crossfades**: Using gamma-corrected crossfades for better visual transitions
3.  **Wipe Transitions**: Creating directional wipes between clips
4.  **Audio Fades**: Creating volume fades and crossfades for audio
5.  **Fade to Color**: Creating fades to/from black or other colors

Usage:
    -   Run this script from Blender's Text Editor or Python Console.
    -   Ensure you are in the Video Editing workspace.
    -   The script creates a basic sequence and then demonstrates various transitions.
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
    from transition_utils import (
        create_crossfade, create_gamma_crossfade, create_wipe, 
        create_audio_fade, create_audio_crossfade, create_fade_to_color
    )
    print("Successfully imported transition utilities")
except ImportError as e:
    print(f"Warning: Could not import transition_utils: {e}")
    print("Using built-in fallback implementations")
    
    # --- Fallback utility implementations ---
    def create_crossfade(seq_editor, strip1, strip2, transition_duration, channel=None):
        """Create a crossfade between two strips"""
        if channel is None:
            channel = max(strip1.channel, strip2.channel) + 1
        return seq_editor.strips.new_effect(
            name=f"Cross_{strip1.name}_{strip2.name}",
            type='CROSS',
            channel=channel,
            frame_start=strip2.frame_start,
            frame_end=strip2.frame_start + transition_duration,
            seq1=strip1,
            seq2=strip2
        )
    
    def create_gamma_crossfade(seq_editor, strip1, strip2, transition_duration, channel=None):
        """Create a gamma crossfade between two strips"""
        if channel is None:
            channel = max(strip1.channel, strip2.channel) + 1
        return seq_editor.strips.new_effect(
            name=f"GammaCross_{strip1.name}_{strip2.name}",
            type='GAMMA_CROSS',
            channel=channel,
            frame_start=strip2.frame_start,
            frame_end=strip2.frame_start + transition_duration,
            seq1=strip1,
            seq2=strip2
        )
    
    def create_wipe(seq_editor, strip1, strip2, transition_duration, wipe_type='SINGLE', angle=0.0, channel=None):
        """Create a wipe transition between two strips"""
        if channel is None:
            channel = max(strip1.channel, strip2.channel) + 1
        wipe = seq_editor.strips.new_effect(
            name=f"Wipe_{strip1.name}_{strip2.name}",
            type='WIPE',
            channel=channel,
            frame_start=strip2.frame_start,
            frame_end=strip2.frame_start + transition_duration,
            seq1=strip1,
            seq2=strip2
        )
        wipe.transition_type = wipe_type
        wipe.direction = 'IN' if angle == 0.0 else 'OUT'
        wipe.angle = angle
        return wipe
    
    def create_audio_fade(sound_strip, fade_type='IN', duration_frames=24):
        """Create a volume fade for a sound strip"""
        try:
            if fade_type == 'IN':
                start_frame = sound_strip.frame_start
                end_frame = start_frame + duration_frames
                start_vol, end_vol = 0.0, 1.0
            else:  # fade_type == 'OUT'
                end_frame = sound_strip.frame_final_end
                start_frame = end_frame - duration_frames
                start_vol, end_vol = 1.0, 0.0
                
            sound_strip.volume = start_vol
            sound_strip.keyframe_insert("volume", frame=start_frame)
            sound_strip.volume = end_vol
            sound_strip.keyframe_insert("volume", frame=end_frame)
            return True
        except Exception as e:
            print(f"Error in audio fade: {e}")
            return False
    
    def create_audio_crossfade(sound1, sound2, overlap_frames=24):
        """Create a crossfade between two audio strips"""
        try:
            fade_start = sound2.frame_start
            fade_end = fade_start + overlap_frames
            
            sound1.volume = 1.0
            sound1.keyframe_insert("volume", frame=fade_start)
            sound1.volume = 0.0
            sound1.keyframe_insert("volume", frame=fade_end)
            
            sound2.volume = 0.0
            sound2.keyframe_insert("volume", frame=fade_start)
            sound2.volume = 1.0
            sound2.keyframe_insert("volume", frame=fade_end)
            return True
        except Exception as e:
            print(f"Error in audio crossfade: {e}")
            return False
    
    def create_fade_to_color(seq_editor, strip, fade_duration, fade_type='IN', color=(0,0,0), channel=None):
        """Create a fade from/to a solid color"""
        if channel is None:
            channel = strip.channel + 1
            
        if fade_type == 'IN':
            color_start = strip.frame_start
            color_end = color_start + fade_duration
        else:  # fade_type == 'OUT'
            color_end = strip.frame_final_end
            color_start = color_end - fade_duration
        
        color_strip = seq_editor.strips.new_effect(
            name=f"FadeColor_{strip.name}",
            type='COLOR',
            channel=channel,
            frame_start=color_start,
            frame_end=color_end
        )
        color_strip.color = color
        
        if fade_type == 'IN':
            return seq_editor.strips.new_effect(
                name=f"FadeIn_{strip.name}",
                type='CROSS',
                channel=channel + 1,
                frame_start=color_start,
                frame_end=color_end,
                seq1=color_strip,
                seq2=strip
            )
        else:  # fade_type == 'OUT'
            return seq_editor.strips.new_effect(
                name=f"FadeOut_{strip.name}",
                type='CROSS',
                channel=channel + 1,
                frame_start=color_start,
                frame_end=color_end,
                seq1=strip,
                seq2=color_strip
            )

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
            print(f"  • {strip_name}: {fps} fps")
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
    """Set up a test sequence with a few video clips for demonstration."""
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
    
    # Start with frame 1 and add some padding between clips
    # For transitions, we'll need clips to overlap, so set initial spacing
    current_frame = 1
    clip_spacing = -30  # Negative spacing means overlap for transitions
    
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
            
            # Update frame position for next clip, adding a small overlap for transitions
            current_frame = video_strip.frame_final_end + clip_spacing
        else:
            print(f"Video file not found: {video_path}")
    
    print(f"Added {len(added_clips)} clips to the sequence.")
    
    # Check and set FPS after adding all clips
    success, message = check_and_set_fps(seq_editor, C.scene)
    print(f"\nFPS Check: {message}")
    
    return added_clips

# --- Chapter 4 Demo Functions ---

def demonstrate_crossfades(seq_editor, clips, transition_duration=24):
    """
    Demonstrate creating crossfade transitions between clips.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
        transition_duration: Duration of transitions in frames
    """
    print("\n--- Demonstrating Crossfades ---")
    
    # We need at least 2 clips for transitions
    if len(clips) < 2:
        print("Need at least 2 clips to demonstrate transitions. Skipping.")
        return
    
    # Create a crossfade between the first two clips
    video1, audio1 = clips[0]
    video2, audio2 = clips[1]
    
    # Make sure clips overlap by checking and possibly adjusting their positions
    # This is a safety check - our setup should already have created overlap
    if video2.frame_start >= video1.frame_final_end:
        print(f"Adjusting clip positions to create {transition_duration} frame overlap")
        video2.frame_start = video1.frame_final_end - transition_duration
        audio2.frame_start = video2.frame_start  # Keep audio in sync with video
    
    # Create the crossfade transition
    crossfade = create_crossfade(
        seq_editor=seq_editor,
        strip1=video1,
        strip2=video2,
        transition_duration=transition_duration,
        channel=3  # Place on channel above both clips
    )
    
    print(f"Created crossfade between '{video1.name}' and '{video2.name}'")
    print(f"  Transition duration: {transition_duration} frames")
    print(f"  Effect channel: {crossfade.channel}")
    
    # Also create an audio crossfade
    success = create_audio_crossfade(audio1, audio2, transition_duration)
    if success:
        print(f"Created audio crossfade between '{audio1.name}' and '{audio2.name}'")

def demonstrate_gamma_crossfade(seq_editor, clips, transition_duration=24):
    """
    Demonstrate creating gamma-corrected crossfade transitions.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
        transition_duration: Duration of transitions in frames
    """
    print("\n--- Demonstrating Gamma Crossfades ---")
    
    # We need at least 3 clips for this demo (to show a different transition type)
    if len(clips) < 3:
        print("Need at least 3 clips to demonstrate gamma crossfade. Skipping.")
        return
    
    # Create a gamma crossfade between clips 2 and 3
    video2, audio2 = clips[1]
    video3, audio3 = clips[2]
    
    # Ensure overlap
    if video3.frame_start >= video2.frame_final_end:
        print(f"Adjusting clip positions to create {transition_duration} frame overlap")
        video3.frame_start = video2.frame_final_end - transition_duration
        audio3.frame_start = video3.frame_start  # Keep audio in sync with video
    
    # Create the gamma crossfade transition
    gamma_crossfade = create_gamma_crossfade(
        seq_editor=seq_editor,
        strip1=video2,
        strip2=video3,
        transition_duration=transition_duration,
        channel=3  # Place on channel above both clips
    )
    
    print(f"Created gamma crossfade between '{video2.name}' and '{video3.name}'")
    print(f"  Transition duration: {transition_duration} frames")
    print(f"  Effect channel: {gamma_crossfade.channel}")
    
    # Also create an audio crossfade
    success = create_audio_crossfade(audio2, audio3, transition_duration)
    if success:
        print(f"Created audio crossfade between '{audio2.name}' and '{audio3.name}'")

def demonstrate_wipe_transitions(seq_editor, clips):
    """
    Demonstrate creating different types of wipe transitions.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
    """
    print("\n--- Demonstrating Wipe Transitions ---")
    
    # For this demo, we'll need to make a copy of a clip and create a wipe transition
    if len(clips) < 1:
        print("Need at least 1 clip to demonstrate wipe transitions. Skipping.")
        return
    
    # Get a source clip and duplicate it
    video1, audio1 = clips[0]
    
    # Create a new strip at the end of the timeline
    last_clip = clips[-1][0]
    start_frame = last_clip.frame_final_end + 30  # Add some spacing
    
    # Use the same filepath to create a new strip
    video_duplicate = add_movie_strip(
        seq_editor=seq_editor,
        filepath=video1.filepath,
        channel=video1.channel,
        frame_start=start_frame,
        name="WipeTarget"
    )
    
    audio_duplicate = add_sound_strip(
        seq_editor=seq_editor,
        filepath=audio1.filepath,
        channel=audio1.channel,
        frame_start=start_frame,
        name="WipeTargetAudio"
    )
    
    # Create a color strip to transition from
    color_strip = seq_editor.strips.new_effect(
        name="BlueBackground",
        type='COLOR',
        channel=video1.channel,
        frame_start=start_frame - 40,  # Start earlier to create overlap
        frame_end=start_frame + 20  # Extend past start of the duplicate clip
    )
    
    # Set the color (blue)
    color_strip.color = (0.0, 0.0, 0.8)
    
    # Create the wipe transition - 60 frames from color to video
    wipe_transition = create_wipe(
        seq_editor=seq_editor,
        strip1=color_strip,
        strip2=video_duplicate,
        transition_duration=60,
        wipe_type='CLOCK',  # Clock wipe
        angle=radians(45),  # 45-degree angle
        channel=4  # Place on channel above both clips
    )
    
    print(f"Created clock wipe transition from color strip to '{video_duplicate.name}'")
    print(f"  Transition duration: 60 frames")
    print(f"  Wipe type: CLOCK at 45-degree angle")
    print(f"  Effect channel: {wipe_transition.channel}")
    
    # Create an audio fade-in for the duplicated audio
    success = create_audio_fade(audio_duplicate, fade_type='IN', duration_frames=60)
    if success:
        print(f"Created audio fade-in for '{audio_duplicate.name}'")

def demonstrate_fade_to_from_black(seq_editor, clips):
    """
    Demonstrate fading to and from black.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
    """
    print("\n--- Demonstrating Fades to/from Black ---")
    
    # We need at least one clip
    if not clips:
        print("Need at least 1 clip to demonstrate fades. Skipping.")
        return
    
    # Get the first clip to fade in from black
    video1, audio1 = clips[0]
    
    # Create a fade-in from black
    fade_in_effect = create_fade_to_color(
        seq_editor=seq_editor,
        strip=video1,
        fade_duration=30,
        fade_type='IN',
        color=(0, 0, 0),  # Black
        channel=5
    )
    
    print(f"Created fade-in from black for '{video1.name}'")
    print(f"  Fade duration: 30 frames")
    print(f"  Effect channel: {fade_in_effect.channel}")
    
    # Create an audio fade-in
    success = create_audio_fade(audio1, fade_type='IN', duration_frames=30)
    if success:
        print(f"Created audio fade-in for '{audio1.name}'")
    
    # Get the last clip to fade out to black
    video_last, audio_last = clips[-1]
    
    # Create a fade-out to black
    fade_out_effect = create_fade_to_color(
        seq_editor=seq_editor,
        strip=video_last,
        fade_duration=30,
        fade_type='OUT',
        color=(0, 0, 0),  # Black
        channel=5
    )
    
    print(f"Created fade-out to black for '{video_last.name}'")
    print(f"  Fade duration: 30 frames")
    print(f"  Effect channel: {fade_out_effect.channel}")
    
    # Create an audio fade-out
    success = create_audio_fade(audio_last, fade_type='OUT', duration_frames=30)
    if success:
        print(f"Created audio fade-out for '{audio_last.name}'")

def main():
    """
    Main function demonstrating Chapter 4 concepts: Transitions and Fades.
    This function sets up a test sequence and then demonstrates various transition types.
    """
    print("Blender VSE Python API Test - Chapter 4: Transitions and Fades")
    print("===========================================================\n")
    
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
    
    # --- 1. Set up a basic sequence with slightly overlapping clips ---
    clips = setup_test_sequence(seq_editor, test_media_dir)
    
    if not clips:
        print("No clips were added. Make sure test_media_dir is correct.")
        return
    
    # --- 2. Demonstrate standard crossfade ---
    demonstrate_crossfades(seq_editor, clips)
    
    # --- 3. Demonstrate gamma crossfade ---
    demonstrate_gamma_crossfade(seq_editor, clips)
    
    # --- 4. Demonstrate wipe transition ---
    demonstrate_wipe_transitions(seq_editor, clips)
    
    # --- 5. Demonstrate fade to/from black ---
    demonstrate_fade_to_from_black(seq_editor, clips)
    
    # --- Recap the Final State ---
    print("\n--- Final State of Strips ---")
    for strip in seq_editor.strips_all:
        print(f"- '{strip.name}' (Type: {strip.type}): Channel {strip.channel}, "
              f"Frames {strip.frame_start}-{strip.frame_final_end}, "
              f"Duration: {strip.frame_final_duration}")
    
    print("\nTest for Chapter 4 completed! Review the VSE timeline and console output.")

# Run the main function to demonstrate Chapter 4 concepts
if __name__ == "__main__":
    main()