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

# Import vse_utils for core functionality
try:
    from vse_utils import (
        get_active_scene, ensure_sequence_editor, add_movie_strip, add_sound_strip,
        add_crossfade, find_test_media_dir, print_sequence_info
    )
    print("Successfully imported VSE utilities")
except ImportError as e:
    print(f"Warning: Could not import vse_utils: {e}")
    raise ImportError("Required vse_utils module not found")

# Now try to import transition utilities
try:
    from transition_utils import (
        create_gamma_crossfade, create_wipe, 
        create_audio_fade, create_audio_crossfade, create_fade_to_color
    )
    print("Successfully imported transition utilities")
except ImportError as e:
    print(f"Warning: Could not import transition_utils: {e}")
    print("Using built-in fallback implementations")
    
    # --- Fallback utility implementations ---
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
        
        # Ensure frame numbers are integers
        start_frame = ensure_integer_frame(strip2.frame_start)
        end_frame = ensure_integer_frame(strip2.frame_start + transition_duration)
        
        wipe = seq_editor.strips.new_effect(
            name=f"Wipe_{strip1.name}_{strip2.name}",
            type='WIPE',
            channel=channel,
            frame_start=start_frame,
            frame_end=end_frame,
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

def ensure_integer_frame(value):
    """Helper function to ensure frame numbers are integers."""
    try:
        return int(float(value))
    except (TypeError, ValueError):
        return value

def safe_remove_strips(seq_editor):
    """Safely remove all strips from the sequence editor."""
    print("\nSafely removing all strips...")
    
    # First, deselect all strips
    for strip in seq_editor.sequences_all:
        strip.select = False
    
    # Remove strips in reverse order (effects first)
    strips_to_remove = sorted(
        seq_editor.sequences_all,
        key=lambda s: 1 if s.type in {'CROSS', 'GAMMA_CROSS', 'WIPE'} else 0
    )
    
    for strip in strips_to_remove:
        try:
            print(f"Removing strip: {strip.name} (Type: {strip.type})")
            seq_editor.sequences.remove(strip)
        except Exception as e:
            print(f"Warning: Could not remove strip {strip.name}: {e}")
    
    print("Strip removal completed.")

def setup_test_sequence(seq_editor, test_media_dir):
    """Set up a test sequence with a few video clips for demonstration."""
    print("\nSetting up test sequence with video clips...")
    
    # Clear all existing strips first
    if seq_editor.sequences_all:
        safe_remove_strips(seq_editor)
    
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
    
    print(f"\nInitial setup: current_frame={current_frame}, clip_spacing={clip_spacing}")
    
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
            print(f"\nAdding clip {i+1} from {video_path}")
            print(f"Starting at frame {current_frame}")
            
            # Add video strip
            video_strip = add_movie_strip(
                seq_editor, 
                video_path, 
                channel=video_channel, 
                frame_start=ensure_integer_frame(current_frame),
                name=f"Clip{i+1}"
            )
            
            print(f"Video strip added: {video_strip.name}")
            print(f"Frame range: {video_strip.frame_start}-{video_strip.frame_final_end}")
            
            # Add matching audio strip
            audio_strip = add_sound_strip(
                seq_editor, 
                video_path, 
                channel=audio_channel, 
                frame_start=ensure_integer_frame(current_frame),
                name=f"Audio{i+1}"
            )
            
            print(f"Audio strip added: {audio_strip.name}")
            print(f"Frame range: {audio_strip.frame_start}-{audio_strip.frame_final_end}")
            
            # Store the pair
            added_clips.append((video_strip, audio_strip))
            
            # Update frame position for next clip, adding a small overlap for transitions
            current_frame = ensure_integer_frame(video_strip.frame_final_end + clip_spacing)
            print(f"Next clip will start at frame {current_frame}")
        else:
            print(f"Video file not found: {video_path}")
    
    print(f"\nAdded {len(added_clips)} clips to the sequence.")
    
    # Print sequence info after adding all clips
    print_sequence_info(seq_editor, "Initial Sequence State")
    
    return added_clips

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
    
    print(f"\nPreparing crossfade between:")
    print(f"Video1: {video1.name} (frames {video1.frame_start}-{video1.frame_final_end})")
    print(f"Video2: {video2.name} (frames {video2.frame_start}-{video2.frame_final_end})")
    
    # Make sure clips overlap by checking and possibly adjusting their positions
    # This is a safety check - our setup should already have created overlap
    if video2.frame_start >= video1.frame_final_end:
        print(f"Adjusting clip positions to create {transition_duration} frame overlap")
        new_start = ensure_integer_frame(video1.frame_final_end - transition_duration)
        print(f"Moving {video2.name} from frame {video2.frame_start} to {new_start}")
        video2.frame_start = new_start
        audio2.frame_start = new_start  # Keep audio in sync with video
    
    # Create the crossfade transition using vse_utils
    print(f"\nCreating crossfade:")
    start_frame = ensure_integer_frame(video2.frame_start)
    end_frame = ensure_integer_frame(start_frame + transition_duration)
    print(f"Start frame: {start_frame}")
    print(f"End frame: {end_frame}")
    print(f"Duration: {transition_duration} frames")
    
    try:
        # Create the effect directly since add_crossfade is having issues
        crossfade = seq_editor.strips.new_effect(
            name=f"Cross_{video1.name}_{video2.name}",
            type='CROSS',
            channel=3,  # Place on channel above both clips
            frame_start=start_frame,
            frame_end=end_frame,
            seq1=video1,
            seq2=video2
        )
        
        print(f"Crossfade created successfully:")
        print(f"Name: {crossfade.name}")
        print(f"Channel: {crossfade.channel}")
        print(f"Frame range: {crossfade.frame_start}-{crossfade.frame_final_end}")
    except Exception as e:
        print(f"Error creating crossfade: {e}")
        import traceback
        traceback.print_exc()
        return
    
    # Also create an audio crossfade
    print("\nCreating audio crossfade...")
    success = create_audio_crossfade(audio1, audio2, ensure_integer_frame(transition_duration))
    if success:
        print(f"Created audio crossfade between '{audio1.name}' and '{audio2.name}'")
    else:
        print("Failed to create audio crossfade")

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
    
    print(f"\nPreparing gamma crossfade between:")
    print(f"Video2: {video2.name} (frames {video2.frame_start}-{video2.frame_final_end})")
    print(f"Video3: {video3.name} (frames {video3.frame_start}-{video3.frame_final_end})")
    
    # Ensure overlap
    if video3.frame_start >= video2.frame_final_end:
        print(f"Adjusting clip positions to create {transition_duration} frame overlap")
        new_start = ensure_integer_frame(video2.frame_final_end - transition_duration)
        print(f"Moving {video3.name} from frame {video3.frame_start} to {new_start}")
        video3.frame_start = new_start
        audio3.frame_start = new_start  # Keep audio in sync with video
    
    # Create the gamma crossfade transition
    print(f"\nCreating gamma crossfade:")
    print(f"Start frame: {ensure_integer_frame(video3.frame_start)}")
    print(f"End frame: {ensure_integer_frame(video3.frame_start + transition_duration)}")
    print(f"Duration: {transition_duration} frames")
    
    try:
        gamma_crossfade = create_gamma_crossfade(
            seq_editor=seq_editor,
            strip1=video2,
            strip2=video3,
            transition_duration=ensure_integer_frame(transition_duration),
            channel=3  # Place on channel above both clips
        )
        
        print(f"Gamma crossfade created successfully:")
        print(f"Name: {gamma_crossfade.name}")
        print(f"Channel: {gamma_crossfade.channel}")
        print(f"Frame range: {gamma_crossfade.frame_start}-{gamma_crossfade.frame_final_end}")
    except Exception as e:
        print(f"Error creating gamma crossfade: {e}")
        import traceback
        traceback.print_exc()
        return

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
    start_frame = ensure_integer_frame(last_clip.frame_final_end + 30)  # Add some spacing
    
    print(f"\nCreating wipe transition demo clips:")
    print(f"Using source clip: {video1.name}")
    print(f"Starting at frame: {start_frame}")
    
    # Get the filepath from the movie strip
    source_path = video1.filepath if hasattr(video1, 'filepath') else None
    if not source_path:
        print("Error: Could not get source filepath from video strip")
        return
    
    # Use the same filepath to create a new strip
    try:
        video_duplicate = add_movie_strip(
            seq_editor=seq_editor,
            filepath=source_path,
            channel=video1.channel,
            frame_start=start_frame,
            name="WipeTarget"
        )
        
        print(f"Created duplicate video strip: {video_duplicate.name}")
        print(f"Frame range: {video_duplicate.frame_start}-{video_duplicate.frame_final_end}")
        
        audio_duplicate = add_sound_strip(
            seq_editor=seq_editor,
            filepath=source_path,
            channel=audio1.channel,
            frame_start=start_frame,
            name="WipeTargetAudio"
        )
        
        print(f"Created duplicate audio strip: {audio_duplicate.name}")
        print(f"Frame range: {audio_duplicate.frame_start}-{audio_duplicate.frame_final_end}")
    except Exception as e:
        print(f"Error creating duplicate strips: {e}")
        import traceback
        traceback.print_exc()
        return
    
    # Create a color strip to transition from
    try:
        color_start = ensure_integer_frame(start_frame - 40)
        color_end = ensure_integer_frame(start_frame + 20)
        
        color_strip = seq_editor.strips.new_effect(
            name="BlueBackground",
            type='COLOR',
            channel=video1.channel,
            frame_start=color_start,  # Start earlier to create overlap
            frame_end=color_end  # Extend past start of the duplicate clip
        )
        
        # Set the color (blue)
        color_strip.color = (0.0, 0.0, 0.8)
        
        print(f"Created color strip: {color_strip.name}")
        print(f"Frame range: {color_strip.frame_start}-{color_strip.frame_final_end}")
    except Exception as e:
        print(f"Error creating color strip: {e}")
        import traceback
        traceback.print_exc()
        return
    
    # Create the wipe transition - 60 frames from color to video
    try:
        # Create the wipe effect directly since the utility function is having issues
        wipe_start = ensure_integer_frame(start_frame)
        wipe_end = ensure_integer_frame(start_frame + 60)
        
        wipe_transition = seq_editor.strips.new_effect(
            name=f"Wipe_{color_strip.name}_{video_duplicate.name}",
            type='WIPE',
            channel=4,  # Place on channel above both clips
            frame_start=wipe_start,
            frame_end=wipe_end,
            seq1=color_strip,
            seq2=video_duplicate
        )
        
        # Set wipe properties
        wipe_transition.transition_type = 'CLOCK'
        wipe_transition.direction = 'OUT'  # Since we have a non-zero angle
        wipe_transition.angle = radians(45)
        
        print(f"Created clock wipe transition from color strip to '{video_duplicate.name}'")
        print(f"  Transition duration: 60 frames")
        print(f"  Wipe type: CLOCK at 45-degree angle")
        print(f"  Effect channel: {wipe_transition.channel}")
        print(f"  Frame range: {wipe_transition.frame_start}-{wipe_transition.frame_final_end}")
    except Exception as e:
        print(f"Error creating wipe transition: {e}")
        import traceback
        traceback.print_exc()
        return
    
    # Create an audio fade-in for the duplicated audio
    print("\nCreating audio fade-in...")
    success = create_audio_fade(audio_duplicate, fade_type='IN', duration_frames=60)
    if success:
        print(f"Created audio fade-in for '{audio_duplicate.name}'")
    else:
        print("Failed to create audio fade-in")

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
    
    # Find test media directory using vse_utils helper
    test_media_dir = find_test_media_dir()
    
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
    
    # --- Print Final State ---
    print_sequence_info(seq_editor, "Final Sequence State")
    
    print("\nTest for Chapter 4 completed! Review the VSE timeline and console output.")

# Run the main function to demonstrate Chapter 4 concepts
if __name__ == "__main__":
    main()