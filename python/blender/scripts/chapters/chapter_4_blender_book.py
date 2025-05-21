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
import traceback # For detailed error logging

# Common shortcuts for Blender data and context
D = bpy.data
C = bpy.context

# --- Utility Path Setup ---
# Try to import utilities - first make sure our utilities path is in sys.path
# IMPORTANT: Never use __file__ in Blender scripts that might be run from text editor.
# Instead, construct absolute path based on a known project structure or pass it.
# For this example, assuming a specific structure relative to a known base path.
# If running from Blender Text editor, this needs to be robust.
# For addon context, bpy.utils.user_resource('SCRIPTS') or similar might be better.

# Construct path to 'scripts' dir from a known point (e.g. current working directory if script is run from CLI)
# or a hardcoded/configurable base path for the project.
scripts_dir_path = None
try:
    # Attempt to get addon preferences if this script is part of an addon
    # This is just an example, actual addon preferences would be structured differently.
    # preferences = C.preferences.addons[__package__].preferences
    # scripts_dir_path = preferences.scripts_directory 
    # Fallback: if not an addon or prefs not set, try to deduce from current file if possible (less reliable in Blender)
    # For robust execution, especially when running as a script in Blender, consider hardcoding 
    # or using an environment variable, or a custom property on the scene/addon.
    
    # For this specific context (corporate-headquarters/go-go-labs/python/blender/scripts/chapters/chapter_4_blender_book.py)
    # we know the structure. If this script is moved, this will break.
    # This is a common challenge with Blender scripts not part of an addon.
    current_script_path = bpy.context.space_data.text.filepath if bpy.context.space_data and bpy.context.space_data.type == 'TEXT_EDITOR' and bpy.context.space_data.text else ""
    if current_script_path:
        # /path/to/.../blender/scripts/chapters/chapter_4_blender_book.py -> /path/to/.../blender/scripts/
        scripts_dir_path = os.path.dirname(os.path.dirname(os.path.abspath(current_script_path)))
    else:
        # Fallback if not in text editor or filepath not available
        # This is a less reliable method, assumes script is run from a certain PWD
        # For the tool execution, this is typically /home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender
        scripts_dir_path = os.path.abspath(os.path.join(os.getcwd(), "scripts")) 
        # Check if this is correct, might need to adjust based on actual PWD when Blender executes.
        # A more robust way for external execution is to pass this path as an argument or env var.
        # For agent execution, let's assume the agent sets the PWD to the project root or similar.
        # The agent log shows PWD is `/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender`
        # So, `os.path.join(os.getcwd(), "scripts")` will be `/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender/scripts`

    if not scripts_dir_path or not os.path.isdir(scripts_dir_path):
        # Final fallback for development/testing if above fails - use a known absolute path
        # THIS SHOULD BE REMOVED OR MADE CONFIGURABLE FOR DISTRIBUTION
        scripts_dir_path = "/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender/scripts" # XXX: Hardcoded path
        print(f"Warning: Using hardcoded scripts_dir_path: {scripts_dir_path}")

    utils_dir = os.path.join(scripts_dir_path, 'utils')
    if utils_dir not in sys.path:
        sys.path.append(utils_dir)
    print(f"Attempting to use utils_dir: {utils_dir}")
    print(f"Current sys.path: {sys.path}")

except Exception as e:
    print(f"Error setting up sys.path for utils: {e}")
    utils_dir = None # Ensure it's None if setup fails

# Import vse_utils for core functionality
# These are expected to be in the utils_dir added to sys.path
try:
    from vse_utils import (
        get_active_scene, ensure_sequence_editor, add_movie_strip, add_sound_strip,
        find_test_media_dir, print_sequence_info
    )
    print("Successfully imported VSE utilities (vse_utils)")
except ImportError as e:
    print(f"Fatal Error: Could not import vse_utils: {e}. Check utils_dir in sys.path and file existence.")
    # Depending on script importance, might raise or exit here.
    # For this demo, we'll try to continue if transition_utils also fails, but it's unlikely to work.
    raise # Re-raise to stop execution if core VSE utils are missing

# Now try to import transition utilities
# These functions are now canonical and robust, no fallbacks needed here.
try:
    from transition_utils import (
        ensure_integer_frame, # Import the helper if needed directly, though it's mostly internal to transition_utils
        create_crossfade,
        create_gamma_crossfade,
        create_wipe, 
        create_audio_fade,
        create_audio_crossfade,
        create_fade_to_color
    )
    print("Successfully imported robust transition utilities (transition_utils)")
except ImportError as e:
    print(f"Fatal Error: Could not import transition_utils: {e}. Check utils_dir in sys.path and file existence.")
    # This is critical for the script, so re-raise.
    raise

def safe_remove_strips(seq_editor: bpy.types.SequenceEditor):
    """Safely remove all strips from the sequence editor, effects first."""
    if not seq_editor:
        print("safe_remove_strips: No sequence editor found.")
        return
    print("\nSafely removing all strips...")
    
    # Deselect all strips first to avoid potential issues with Blender's UI state
    for s in seq_editor.sequences_all:
        s.select = False
    
    # Sort strips: effects (like CROSS, WIPE, GAMMA_CROSS) are removed before base strips (MOVIE, SOUND, COLOR)
    # This helps prevent errors if an effect's input strip is removed first.
    # Simple sort: place known effect types later in the sort order (so they are processed first by reversed()).
    # A more robust way might involve checking strip.type against a set of known effect types.
    effect_types = {'CROSS', 'GAMMA_CROSS', 'WIPE', 'COLOR'} # Color can be an effect input
    
    # We iterate and remove. This can be tricky if the collection changes.
    # Best to collect names or indices first, then remove.
    # However, Blender's remove operation might be safe if done carefully.
    # Let's try removing in a loop, but be mindful of collection changes.
    # A safer approach: build a list of strips to remove, then iterate that list.
    
    strips_to_process = list(seq_editor.sequences_all) # Shallow copy
    
    # Sort by type: non-effects first, then effects. When reversing, effects are removed first.
    strips_sorted_for_removal = sorted(
        strips_to_process, 
        key=lambda s: 0 if s.type not in effect_types else 1 
        # MOVIE, SOUND (0) will be at the start
        # CROSS, GAMMA_CROSS, WIPE, COLOR (1) will be at the end
    )

    # Iterate in reverse over the sorted list (effects first, then their inputs)
    for strip in reversed(strips_sorted_for_removal):
        try:
            if strip and strip.name in seq_editor.sequences_all:
                 # Check name because strip object might become invalid after other removals
                print(f"Attempting to remove strip: '{strip.name}' (Type: {strip.type}, Channel: {strip.channel})")
                seq_editor.sequences.remove(strip)
                print(f"  Successfully removed '{strip.name}'")
            elif strip:
                print(f"  Skipping removal of '{strip.name}' (Type: {strip.type}) - no longer in sequences_all (already removed or invalid). ")
            else:
                print("  Skipping removal of an invalid/None strip object.")
        except Exception as e:
            # Use strip.name if strip object is still valid enough, otherwise placeholder
            strip_name_for_error = strip.name if strip and hasattr(strip, 'name') else "[Unknown/Invalid Strip]"
            print(f"Warning: Could not remove strip '{strip_name_for_error}': {e}")
            # traceback.print_exc() # Optionally print full traceback for debugging this specific error
    
    print("Strip removal process completed.")
    # Verify (optional)
    if seq_editor.sequences_all:
        print(f"Warning: {len(seq_editor.sequences_all)} strips still remain after cleanup:")
        for s in seq_editor.sequences_all:
            print(f"  - '{s.name}' (Type: {s.type})")
    else:
        print("All strips successfully cleared from the sequence editor.")

def setup_test_sequence(seq_editor: bpy.types.SequenceEditor, test_media_dir: str) -> list[tuple[bpy.types.Sequence, bpy.types.Sequence]]:
    """Set up a test sequence with a few video clips for demonstration."""
    print("\nSetting up test sequence with video clips...")
    
    if seq_editor.sequences_all:
        safe_remove_strips(seq_editor)
    
    video_files = [
        "SampleVideo_1280x720_2mb.mp4",
        "SampleVideo_1280x720_1mb.mp4",
        "SampleVideo_1280x720_5mb.mp4"
    ]
    
    video_channel = 1
    audio_channel = 2
    current_frame = 1
    clip_spacing = -30  # Negative overlap for transitions
    
    print(f"Initial setup: current_frame={current_frame}, clip_spacing={clip_spacing}")
    
    if video_channel < len(seq_editor.channels):
        seq_editor.channels[video_channel-1].name = "VideoCh"
    if audio_channel < len(seq_editor.channels):
        seq_editor.channels[audio_channel-1].name = "AudioCh"
    
    added_clips: list[tuple[bpy.types.Sequence, bpy.types.Sequence]] = []
    
    for i, video_filename in enumerate(video_files):
        video_path = os.path.join(test_media_dir, video_filename)
        if os.path.exists(video_path):
            clip_name_base = f"Clip{i+1}"
            print(f"\nAdding '{clip_name_base}' from {video_path}, starting at frame {current_frame}")
            
            frame_start_int = ensure_integer_frame(current_frame)
            
            video_strip = add_movie_strip(
                seq_editor, 
                video_path, 
                channel=video_channel, 
                frame_start=frame_start_int,
                name=clip_name_base
            )
            if not video_strip:
                print(f"Error: Failed to add video strip for {video_path}")
                continue # Skip this file
            
            print(f"  Video strip '{video_strip.name}' added: frames {video_strip.frame_start}-{video_strip.frame_final_end}")
            
            audio_strip = add_sound_strip(
                seq_editor, 
                video_path, 
                channel=audio_channel, 
                frame_start=frame_start_int,
                name=f"Audio{i+1}"
            )
            if not audio_strip:
                print(f"Error: Failed to add audio strip for {video_path}")
                # We might have a video strip without audio, decide if this is critical
                # For now, let's add a placeholder if audio fails, or handle it gracefully
                added_clips.append((video_strip, None)) # type: ignore
                current_frame = ensure_integer_frame(video_strip.frame_final_end + clip_spacing)
                continue
            
            print(f"  Audio strip '{audio_strip.name}' added: frames {audio_strip.frame_start}-{audio_strip.frame_final_end}")
            added_clips.append((video_strip, audio_strip))
            current_frame = ensure_integer_frame(video_strip.frame_final_end + clip_spacing)
            print(f"  Next clip planned to start at frame {current_frame}")
        else:
            print(f"Video file not found and skipped: {video_path}")
    
    print(f"\nAdded {len(added_clips)} clip pairs to the sequence.")
    print_sequence_info(seq_editor, "Initial Sequence State After Setup")
    return added_clips

def demonstrate_crossfades(seq_editor: bpy.types.SequenceEditor, clips: list, transition_duration: int = 24):
    """Demonstrate standard crossfade using transition_utils."""
    print("\n--- Demonstrating Crossfades (using transition_utils) ---")
    if len(clips) < 2:
        print("Need at least 2 clips for crossfades. Skipping.")
        return

    video1, audio1 = clips[0]
    video2, audio2 = clips[1]

    if not video1 or not video2:
        print("Error: Missing video strips for crossfade. Skipping.")
        return
        
    # Ensure strips are positioned for overlap BEFORE calling create_crossfade
    # transition_utils expects caller to manage layout.
    # For a crossfade of `transition_duration` starting at video2.frame_start,
    # video1 must extend to at least video2.frame_start + transition_duration.
    # video2 must start such that video1 has `transition_duration` frames of overlap.
    # A common setup: video2.frame_start = video1.frame_final_end - transition_duration
    
    # Adjust video2 start if it's not already overlapping correctly
    # This logic is specific to this demo script's layout goals.
    required_video2_start = ensure_integer_frame(video1.frame_final_end - transition_duration)
    if video2.frame_start > required_video2_start:
        print(f"Adjusting '{video2.name}' start from {video2.frame_start} to {required_video2_start} for overlap.")
        video2.frame_start = required_video2_start
        if audio2: audio2.frame_start = required_video2_start # Keep audio synced
    
    print(f"Preparing crossfade between '{video1.name}' ({video1.frame_start}-{video1.frame_final_end}) and '{video2.name}' ({video2.frame_start}-{video2.frame_final_end})")

    crossfade_effect = create_crossfade(
        seq_editor=seq_editor,
        strip1=video1,
        strip2=video2,
        transition_duration=transition_duration,
        channel=max(video1.channel, video2.channel) + 1 # Example channel placement
    )
    if crossfade_effect:
        print(f"Successfully created crossfade effect '{crossfade_effect.name}'")
    else:
        print(f"Failed to create crossfade between '{video1.name}' and '{video2.name}'. Check logs for errors.")

    if audio1 and audio2:
        if create_audio_crossfade(audio1, audio2, transition_duration):
            print(f"Successfully created audio crossfade for '{audio1.name}' and '{audio2.name}'")
        else:
            print(f"Failed to create audio crossfade for '{audio1.name}' and '{audio2.name}'.")
    else:
        print("Skipping audio crossfade due to missing audio strips.")

def demonstrate_gamma_crossfade(seq_editor: bpy.types.SequenceEditor, clips: list, transition_duration: int = 24):
    """Demonstrate gamma crossfade using transition_utils."""
    print("\n--- Demonstrating Gamma Crossfades (using transition_utils) ---")
    if len(clips) < 3: # Using clips 2 and 3 for this demo
        print("Need at least 3 clips for this gamma crossfade demo. Skipping.")
        return

    video2, audio2 = clips[1]
    video3, audio3 = clips[2]

    if not video2 or not video3:
        print("Error: Missing video strips for gamma crossfade. Skipping.")
        return
        
    # Adjust video3 start for overlap
    required_video3_start = ensure_integer_frame(video2.frame_final_end - transition_duration)
    if video3.frame_start > required_video3_start:
        print(f"Adjusting '{video3.name}' start from {video3.frame_start} to {required_video3_start} for overlap.")
        video3.frame_start = required_video3_start
        if audio3: audio3.frame_start = required_video3_start
        
    print(f"Preparing gamma crossfade between '{video2.name}' ({video2.frame_start}-{video2.frame_final_end}) and '{video3.name}' ({video3.frame_start}-{video3.frame_final_end})")

    gamma_effect = create_gamma_crossfade(
        seq_editor=seq_editor,
        strip1=video2,
        strip2=video3,
        transition_duration=transition_duration,
        channel=max(video2.channel, video3.channel) + 1
    )
    if gamma_effect:
        print(f"Successfully created gamma crossfade effect '{gamma_effect.name}'")
    else:
        print(f"Failed to create gamma crossfade between '{video2.name}' and '{video3.name}'.")

    if audio2 and audio3:
        if create_audio_crossfade(audio2, audio3, transition_duration):
            print(f"Successfully created audio crossfade for '{audio2.name}' and '{audio3.name}'")
        else:
            print(f"Failed to create audio crossfade for '{audio2.name}' and '{audio3.name}'.")
    else:
        print("Skipping audio crossfade due to missing audio strips for gamma demo.")

def demonstrate_wipe_transitions(seq_editor: bpy.types.SequenceEditor, clips: list, transition_duration: int = 60):
    """Demonstrate wipe transition using transition_utils."""
    print("\n--- Demonstrating Wipe Transitions (using transition_utils) ---")
    if not clips:
        print("Need at least 1 clip to create a wipe target. Skipping.")
        return

    video_orig, audio_orig = clips[0] # Source for duplication
    if not video_orig:
        print("Error: Missing original video strip for wipe demo. Skipping.")
        return

    # Create a color strip to wipe from
    color_strip_channel = video_orig.channel # Place color on same channel as video for this demo layout
    color_strip_start_frame = ensure_integer_frame(video_orig.frame_final_end + 30) # Place after original clips
    color_strip_duration = ensure_integer_frame(transition_duration + 20) # Make it long enough
    color_strip_end_frame = ensure_integer_frame(color_strip_start_frame + color_strip_duration)
    
    try:
        color_strip = seq_editor.strips.new_effect(
            name="BlueWipeSource", type='COLOR',
            channel=color_strip_channel,
            frame_start=color_strip_start_frame,
            frame_end=color_strip_end_frame
        )
        color_strip.color = (0.1, 0.2, 0.8) # Bluish
        print(f"Created color strip '{color_strip.name}' for wipe demo.")
    except Exception as e:
        print(f"Error creating color strip for wipe: {e}")
        traceback.print_exc()
        return

    # Create a duplicate of the first video clip to wipe into
    # Place it overlapping with the color strip for the wipe transition
    video_target_start_frame = ensure_integer_frame(color_strip.frame_start + (color_strip_duration - transition_duration) // 2) # Center transition somewhat
    
    video_target = add_movie_strip(
        seq_editor, video_orig.filepath, 
        channel=video_orig.channel, # Same channel, will be covered by wipe effect
        frame_start=video_target_start_frame, 
        name="WipeTargetVideo"
    )
    if not video_target:
        print("Error: Failed to create target video strip for wipe. Skipping.")
        safe_remove_strips(seq_editor) # Clean up color strip if target fails
        return

    audio_target = None
    if audio_orig and hasattr(audio_orig, 'filepath') and audio_orig.filepath: # Check if audio_orig has filepath
        audio_target = add_sound_strip(
            seq_editor, audio_orig.filepath, 
            channel=audio_orig.channel, 
            frame_start=video_target_start_frame, 
            name="WipeTargetAudio"
        )
    
    print(f"Preparing wipe from '{color_strip.name}' to '{video_target.name}'")

    wipe_effect = create_wipe(
        seq_editor=seq_editor,
        strip_being_wiped_away=color_strip, 
        strip_being_wiped_in=video_target,
        transition_duration=transition_duration,
        wipe_type='CLOCK',
        angle=radians(45),
        channel=max(color_strip.channel, video_target.channel) + 1
    )

    if wipe_effect:
        print(f"Successfully created wipe effect '{wipe_effect.name}'")
    else:
        print(f"Failed to create wipe from '{color_strip.name}' to '{video_target.name}'.")

    if audio_target:
        if create_audio_fade(audio_target, fade_type='IN', duration_frames=transition_duration):
            print(f"Successfully created audio fade-in for '{audio_target.name}'")
        else:
            print(f"Failed to create audio fade-in for '{audio_target.name}'.")
    else:
        print("Skipping audio fade for wipe target as audio_target is None.")

def demonstrate_fade_to_from_black(seq_editor: bpy.types.SequenceEditor, clips: list, fade_duration: int = 30):
    """Demonstrate fading to/from black using transition_utils."""
    print("\n--- Demonstrating Fades to/from Black (using transition_utils) ---")
    if not clips:
        print("Need at least 1 clip for fade to/from black. Skipping.")
        return

    # Fade in first clip
    video1, audio1 = clips[0]
    if video1:
        print(f"Preparing fade-in for '{video1.name}'")
        fade_in_effect = create_fade_to_color(
            seq_editor=seq_editor,
            strip=video1,
            fade_duration=fade_duration,
            fade_type='IN',
            color=(0,0,0) # Black
            # Channel will be auto-assigned by create_fade_to_color
        )
        if fade_in_effect:
            print(f"Successfully created fade-in effect '{fade_in_effect.name}' for '{video1.name}'")
        else:
            print(f"Failed to create fade-in for '{video1.name}'.")
        
        if audio1:
            if create_audio_fade(audio1, fade_type='IN', duration_frames=fade_duration):
                print(f"Successfully created audio fade-in for '{audio1.name}'")
            else:
                print(f"Failed to create audio fade-in for '{audio1.name}'.")
    else:
        print("Skipping fade-in demo due to missing first video strip.")

    # Fade out last clip
    if len(clips) > 0:
        video_last, audio_last = clips[-1]
        if video_last:
            print(f"Preparing fade-out for '{video_last.name}'")
            fade_out_effect = create_fade_to_color(
                seq_editor=seq_editor,
                strip=video_last,
                fade_duration=fade_duration,
                fade_type='OUT',
                color=(0,0,0) # Black
            )
            if fade_out_effect:
                print(f"Successfully created fade-out effect '{fade_out_effect.name}' for '{video_last.name}'")
            else:
                print(f"Failed to create fade-out for '{video_last.name}'.")
            
            if audio_last:
                if create_audio_fade(audio_last, fade_type='OUT', duration_frames=fade_duration):
                    print(f"Successfully created audio fade-out for '{audio_last.name}'")
                else:
                    print(f"Failed to create audio fade-out for '{audio_last.name}'.")
        else:
            print("Skipping fade-out demo due to missing last video strip.")

def main():
    """
    Main function demonstrating Chapter 4 concepts: Transitions and Fades.
    This function sets up a test sequence and then demonstrates various transition types.
    """
    print("Blender VSE Python API Test - Chapter 4: Transitions and Fades")
    print("===========================================================\n")
    
    try:
        scene = get_active_scene()
        if not scene:
            print("Fatal Error: Could not get active scene. Exiting.")
            return
            
        seq_editor = ensure_sequence_editor(scene)
        if not seq_editor:
            print(f"Fatal Error: Could not ensure sequence editor for scene '{scene.name}'. Exiting.")
            return
            
        print(f"Operating on scene: '{scene.name}' with Sequence Editor: {seq_editor}")
        
        test_media_dir = find_test_media_dir()
        if not test_media_dir:
            print("Fatal Error: Test media directory not found. Please check VSE_TEST_MEDIA_DIR env var or script config. Exiting.")
            return
            
        print(f"Using test media from: {test_media_dir}")

        # --- 1. Set up a basic sequence with slightly overlapping clips ---   
        clips = setup_test_sequence(seq_editor, test_media_dir)
        if not clips:
            print("Error: No clips were added during setup_test_sequence. Exiting demo.")
            return
        
        # --- 2. Demonstrate standard crossfade ---
        demonstrate_crossfades(seq_editor, clips, transition_duration=24)
        
        # --- 3. Demonstrate gamma crossfade ---
        demonstrate_gamma_crossfade(seq_editor, clips, transition_duration=24)
        
        # --- 4. Demonstrate wipe transition ---
        demonstrate_wipe_transitions(seq_editor, clips, transition_duration=60)
        
        # --- 5. Demonstrate fade to/from black ---
        demonstrate_fade_to_from_black(seq_editor, clips, fade_duration=30)
        
        print("\n--- Final Sequence State After All Demonstrations ---")
        print_sequence_info(seq_editor, "Final Sequence State")
        
        print("\nTest for Chapter 4 completed! Review the VSE timeline and console output.")

    except ImportError as e:
        print(f"Fatal ImportError during main execution: {e}. This usually means a critical utility script is missing or sys.path is incorrect.")
        traceback.print_exc()
    except Exception as e:
        print(f"An unexpected error occurred in main: {e}")
        traceback.print_exc()

# Run the main function to demonstrate Chapter 4 concepts
# This check is standard for Python scripts but less critical when run from Blender's Text Editor via "Run Script"
# However, it's good practice if the script might be imported or run in other contexts.
if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"Critical error running main() from __main__ guard: {e}")
        traceback.print_exc()