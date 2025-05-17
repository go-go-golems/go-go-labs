# context.area: SEQUENCE_EDITOR
# Don't remove the comment above. It's important.

"""
Blender VSE Python API Test Script - Chapter 3: Trimming and Splitting Clips

This script demonstrates core concepts from Chapter 3 of the Blender VSE Python API guide,
focusing on programmatically trimming, cutting, and manipulating video clips in the VSE.

Key concepts covered:
1.  **Trimming Clips**: Adjusting in and out points using:
    -   `frame_offset_start` and `frame_offset_end` to trim from start/end.
    -   `frame_final_start` and `frame_final_end` to set explicit trim points.
2.  **Splitting Clips**: Using `bpy.ops.sequencer.split()` to cut strips into multiple parts.
3.  **Moving Strips**: Changing position by adjusting `frame_start` and `channel`.
4.  **Removing Segments**: Cutting out sections and closing gaps.
5.  **Example Operations**: Practical editing scenarios like removing a middle section.
6.  **Slip Operations**: Adjusting content while keeping timeline position fixed.


Usage:
    -   Run this script from Blender's Text Editor or Python Console.
    -   Ensure you are in the Video Editing workspace.
    -   The script creates a basic sequence and then demonstrates trimming/cutting.
"""

import bpy # type: ignore
import os
from mathutils import * # type: ignore

# Common shortcuts for Blender data and context
D = bpy.data
C = bpy.context

# --- Import core utilities from chapter 2 ---

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

# --- Chapter 3 Functions ---

def trim_strip_start(strip, frames_to_trim):
    """
    Trim frames from the beginning of a strip.
    
    As explained in Chapter 3, trimming the start means increasing the
    frame_offset_start value to skip frames from the source.
    
    Args:
        strip (bpy.types.Strip): The strip to trim.
        frames_to_trim (int): Number of frames to trim from beginning.
        
    Returns:
        tuple: (original_start_frame, new_start_frame) for verification.
    """
    original_start = strip.frame_final_start
    
    # Method 1: Using frame_offset_start
    strip.frame_offset_start += frames_to_trim
    
    print(f"Trimmed {frames_to_trim} frames from start of '{strip.name}'")
    print(f"  Original start frame: {original_start}")
    print(f"  New start frame: {strip.frame_final_start}")
    print(f"  Offset start is now: {strip.frame_offset_start}")
    
    return (original_start, strip.frame_final_start)

def trim_strip_end(strip, frames_to_trim):
    """
    Trim frames from the end of a strip.
    
    Trimming the end involves increasing the frame_offset_end
    to skip frames from the end of the source.
    
    Args:
        strip (bpy.types.Strip): The strip to trim.
        frames_to_trim (int): Number of frames to trim from end.
        
    Returns:
        tuple: (original_end_frame, new_end_frame) for verification.
    """
    original_end = strip.frame_final_end
    
    # Method 1: Using frame_offset_end
    strip.frame_offset_end += frames_to_trim
    
    print(f"Trimmed {frames_to_trim} frames from end of '{strip.name}'")
    print(f"  Original end frame: {original_end}")
    print(f"  New end frame: {strip.frame_final_end}")
    print(f"  Offset end is now: {strip.frame_offset_end}")
    
    return (original_end, strip.frame_final_end)

def split_strip(strip, frame, select_right=True):
    """
    Split a strip at the specified frame.
    
    This uses the sequencer.split operator to cut a strip into two parts.
    The split is made at the specified frame, and by default the right
    part remains selected after the cut.
    
    Args:
        strip (bpy.types.Strip): The strip to split.
        frame (int or float): Frame at which to make the cut (will be converted to int).
        select_right (bool): Whether to select the right part after cutting.
        
    Returns:
        tuple: (left_part, right_part) - the two resulting strips. Either can be None
               if the split operation failed or produced only one side (e.g., frame
               equals strip boundary).
    """
    # Ensure we have an integer frame number
    frame = int(frame)

    # If the frame is outside the strip bounds, do nothing and warn
    if frame <= strip.frame_start or frame >= strip.frame_final_end:
        print(f"  [WARN] Split frame {frame} is outside '{strip.name}' bounds (" \
              f"{strip.frame_start}-{strip.frame_final_end}). No split performed.")
        return (strip, None)

    seq = C.scene.sequence_editor

    # Record existing strips (by pointer) before the split so we can detect new ones
    pre_split_ids = {s.as_pointer() for s in seq.strips_all}

    # Select only the strip we want to split
    for s in seq.strips_all:
        s.select = (s == strip)

    # Set current frame to the split point
    C.scene.frame_current = frame

    # Perform the split – if it fails (e.g. strip locked) catch the exception
    try:
        bpy.ops.sequencer.split(frame=frame, channel=strip.channel, type='SOFT')
    except RuntimeError as e:
        print(f"  [ERROR] Split failed on '{strip.name}': {e}")
        return (strip, None)

    # Identify strips after split
    left_part = None
    right_part = None
    for s in seq.strips_all:
        if s.as_pointer() in pre_split_ids:
            # Existing strip – may have updated range; determine if it's left or right
            if s.channel == strip.channel:
                if s.frame_final_end == frame:
                    left_part = s
                elif s.frame_start == frame:
                    right_part = s
        else:
            # New strip created by split – decide if left or right
            if s.channel == strip.channel:
                if s.frame_final_end == frame:
                    left_part = s
                elif s.frame_start == frame:
                    right_part = s

    # Fallback heuristics if one side still missing
    if not left_part or not right_part:
        for s in seq.strips_all:
            if s.channel != strip.channel:
                continue
            if s.frame_final_end <= frame and (not left_part):
                left_part = s
            elif s.frame_start >= frame and (not right_part):
                right_part = s

    print(f"Split strip '{strip.name}' at frame {frame}")
    if left_part:
        print(f"  Left part: '{left_part.name}', frames {left_part.frame_start}-{left_part.frame_final_end}")
    else:
        print("  Left part: Not created or not found")
    if right_part:
        print(f"  Right part: '{right_part.name}', frames {right_part.frame_start}-{right_part.frame_final_end}")
    else:
        print("  Right part: Not created or not found (e.g., split at end of content)")

    # Manage selection state
    if left_part and right_part:
        if select_right:
            left_part.select = False
            right_part.select = True
        else:
            left_part.select = True
            right_part.select = False

    return (left_part, right_part)

def remove_segment(strip, start_frame, end_frame):
    """
    Remove a segment from a strip and close the gap.
    
    If the splits fail (e.g., frame outside strip bounds), the function will
    abort gracefully and return (strip, None).
    """
    print(f"\nRemoving segment from {start_frame} to {end_frame} from '{strip.name}'")

    # Split at the start of segment
    part_a, part_b = split_strip(strip, start_frame, select_right=True)
    if not part_b:
        print("  [WARN] Unable to create middle/right part at first split – aborting segment removal.")
        return (part_a, None)

    # Split at the end of segment (working on part_b)
    part_b, part_c = split_strip(part_b, end_frame, select_right=True)
    if not part_b:
        print("  [WARN] Second split produced no middle segment – aborting segment removal.")
        return (part_a, part_c)

    # Remove the middle segment (part_b)
    seq = C.scene.sequence_editor
    if part_b and part_b in seq.strips_all:
        seq.strips.remove(part_b)
        print(f"  Removed segment '{part_b.name}'")

    # Move part_c to start at the end of part_a
    if part_c and part_a:
        part_c.frame_start = part_a.frame_final_end
        print(f"  Moved '{part_c.name}' to frame {part_c.frame_start} (closing the gap)")

    return (part_a, part_c)

def slip_strip(strip, offset):
    """
    Perform a slip edit, adjusting the content shown without changing strip length.
    
    A slip edit preserves the duration and timeline position of a strip,
    but shifts what portion of the source is displayed. It's equivalent to 
    adjusting both offsets equally but in opposite directions.
    
    Args:
        strip (bpy.types.Strip): The strip to slip.
        offset (int): Frames to slip by (positive = later content, negative = earlier).
        
    Returns:
        None
    """
    # Save original values for reporting
    orig_start_offset = strip.frame_offset_start
    orig_end_offset = strip.frame_offset_end
    
    # Select the strip and use the slip operator
    strip.select = True
    bpy.ops.sequencer.slip(offset=offset)
    
    print(f"Slip edit on '{strip.name}' by {offset} frames")
    print(f"  Start offset changed: {orig_start_offset} → {strip.frame_offset_start}")
    print(f"  End offset changed: {orig_end_offset} → {strip.frame_offset_end}")

def move_strip(strip, new_frame_start=None, new_channel=None):
    """
    Move a strip to a new position in time or to a different channel.
    
    This simple function demonstrates how to reposition strips by
    directly setting their properties.
    
    Args:
        strip (bpy.types.Strip): The strip to move.
        new_frame_start (int, optional): New starting frame.
        new_channel (int, optional): New channel number.
        
    Returns:
        None
    """
    old_start = strip.frame_start
    old_channel = strip.channel
    
    if new_frame_start is not None:
        strip.frame_start = new_frame_start
    
    if new_channel is not None:
        strip.channel = new_channel
    
    changes = []
    if new_frame_start is not None:
        changes.append(f"frame_start: {old_start} → {strip.frame_start}")
    if new_channel is not None:
        changes.append(f"channel: {old_channel} → {strip.channel}")
    
    print(f"Moved strip '{strip.name}': {', '.join(changes)}")

def check_and_set_fps(seq_editor, scene):
    """
    Check FPS of all strips and ensure they match the scene FPS.
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
    """
    Set up a test sequence with a few video clips for demonstration.
    This reuses logic from Chapter 2's main function but in a more focused way.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor to add strips to.
        test_media_dir (str): Directory containing test media files.
        
    Returns:
        list: A list of (video_strip, audio_strip) tuples for the added clips.
    """
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
    current_frame = 1
    frame_padding = 0  # No padding - we'll place clips back-to-back
    
    # Name the channels
    if video_channel < len(seq_editor.channels):
        seq_editor.channels[video_channel].name = "Video"
    if audio_channel < len(seq_editor.channels):
        seq_editor.channels[audio_channel].name = "Audio"
    
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
def main():
    """
    Main function demonstrating Chapter 3 concepts: Trimming and Splitting Clips.
    This function sets up a test sequence and then demonstrates various editing operations.
    """
    print("Blender VSE Python API Test - Chapter 3: Trimming and Splitting Clips")
    print("===================================================================")
    
    scene = get_active_scene()
    seq_editor = ensure_sequence_editor(scene)
    
    print(f"\nOperating on scene: '{scene.name}' with Sequence Editor: {seq_editor}")
    
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
    
    # --- 1. Set up a basic sequence ---
    clips = setup_test_sequence(seq_editor, test_media_dir)
    
    if not clips:
        print("No clips were added. Make sure test_media_dir is correct.")
        return

    # --- 2. Demonstrate Trimming ---
    print("\n--- Demonstration: Trimming ---")
    # Let's trim the first clip: 24 frames from start, 24 frames from end
    if clips and len(clips) > 0:
        video1, audio1 = clips[0]
        print(f"\nOriginal clip '{video1.name}' duration: {video1.frame_duration} frames")
        
        # Trim 24 frames (1 second at 24fps) from the start
        trim_strip_start(video1, 24)
        trim_strip_start(audio1, 24)  # Keep audio in sync
        
        # Trim 24 frames from the end
        trim_strip_end(video1, 24)
        trim_strip_end(audio1, 24)  # Keep audio in sync
        
        print(f"After trimming, '{video1.name}' final duration: {video1.frame_final_duration} frames")
    
    # --- 3. Demonstrate Splitting ---
    print("\n--- Demonstration: Splitting ---")
    if clips and len(clips) > 1:
        video2, audio2 = clips[1]
        
        # Calculate a split point near the middle and ensure it's an integer
        middle_frame = int(video2.frame_start + (video2.frame_final_duration // 2))
        print(f"\nSplitting '{video2.name}' at frame {middle_frame}")
        
        # Split the video and audio
        video_parts = split_strip(video2, middle_frame)
        audio_parts = split_strip(audio2, middle_frame)
        
        print(f"Original clip split into: '{video_parts[0].name}' and '{video_parts[1].name}'")
    
    # --- 4. Demonstrate Removing a Segment ---
    print("\n--- Demonstration: Removing a Segment ---")
    # Get a fresh clip for this demo since previous operations may have modified original clips
    # We'll try to find a clip with enough duration for our test
    video3 = None
    audio3 = None
    
    # Find the longest video and audio strips to use
    for strip in seq_editor.strips_all:
        if strip.type == 'MOVIE' and (video3 is None or strip.frame_final_duration > video3.frame_final_duration):
            video3 = strip
        elif strip.type == 'SOUND' and (audio3 is None or strip.frame_final_duration > audio3.frame_final_duration):
            audio3 = strip
    
    # Check if we have valid strips to work with
    if video3 and audio3 and video3.frame_final_duration > 100:  # Ensure enough frames for the demo
        # Define a segment to remove (e.g., 1/4 through to 1/2 through)
        # Use a smaller segment to ensure we're well within the clip bounds
        segment_start = video3.frame_start + (video3.frame_final_duration // 4)
        segment_end = video3.frame_start + (video3.frame_final_duration // 2)
        
        print(f"\nRemoving middle segment from '{video3.name}'")
        print(f"  Original duration: {video3.frame_final_duration} frames")
        
        # Remove segment from video
        video_result = remove_segment(video3, segment_start, segment_end)
        # Also remove from audio to keep sync
        audio_result = remove_segment(audio3, segment_start, segment_end)
        
        # Final check - should have two parts that together are shorter
        if 'video_result' in locals() and video_result[0] and video_result[1]:
            new_duration = video_result[0].frame_final_duration + video_result[1].frame_final_duration
            print(f"  New combined duration: {new_duration} frames")
    else:
        print("\nSkipping segment removal demo: No suitable strips found.")
    
    # --- 5. Demonstrate Slip Edit ---
    print("\n--- Demonstration: Slip Edit ---")
    # First check if we have video_parts from the earlier splitting demo
    slip_target = None
    
    if 'video_parts' in locals() and video_parts and len(video_parts) > 0:
        # Use one of the parts from our earlier split
        slip_target = video_parts[0]
    elif clips and len(clips) > 0:
        # If splitting wasn't done or failed, use the first clip
        slip_target = clips[0][0]  # First clip's video part
    
    if slip_target:
        print(f"\nPerforming slip edit on '{slip_target.name}'")
        print(f"  Before slip: offset_start={slip_target.frame_offset_start}, offset_end={slip_target.frame_offset_end}")
        
        # Slip by 12 frames (show later content)
        slip_strip(slip_target, 12)
    else:
        print("\nCannot perform slip edit: No suitable target strip found.")
    
    # --- 6. Demonstrate Moving a Strip ---
    print("\n--- Demonstration: Moving a Strip ---")
    # First check if we have video_parts from the earlier splitting demo
    move_target = None
    
    if 'video_parts' in locals() and video_parts and len(video_parts) > 1:
        move_target = video_parts[1]
    elif clips and len(clips) > 1:
        # If splitting wasn't done or failed, use the second clip
        move_target = clips[1][0]  # Second clip's video part
    
    if move_target:
        print(f"\nMoving '{move_target.name}'")
        # Move to a higher channel
        move_strip(move_target, new_channel=4)
        # Also move it earlier by 10 frames
        move_strip(move_target, new_frame_start=move_target.frame_start - 10)
    else:
        print("\nCannot perform move operation: No suitable target strip found.")
    
    # --- Recap the Final State ---
    print("\n--- Final State of Strips ---")
    for strip in seq_editor.strips_all:
        print(f"- '{strip.name}' (Type: {strip.type}): Channel {strip.channel}, "
              f"Frames {strip.frame_start}-{strip.frame_final_end}, "
              f"Duration: {strip.frame_final_duration}, "
              f"Offsets: start={strip.frame_offset_start}, end={strip.frame_offset_end}")
    
    print("\nTest for Chapter 3 completed! Review the VSE timeline and console output.")

# Run the main function to demonstrate Chapter 3 concepts
main() 