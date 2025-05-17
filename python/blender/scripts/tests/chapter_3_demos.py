# Demonstration functions for Chapter 3: Trimming and Splitting Clips

"""
This file contains individual test functions for each demonstration
from Chapter 3 of the Blender VSE Python API guide, each in its own
isolated environment for proper testing.
"""

import bpy
import os
import sys

# Import our utilities
script_dir = os.path.dirname(os.path.abspath(__file__))
if script_dir not in sys.path:
    sys.path.append(script_dir)

import vse_utils

# ----- Core Operations -----

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

    seq = bpy.context.scene.sequence_editor

    # Record existing strips (by pointer) before the split so we can detect new ones
    pre_split_ids = {s.as_pointer() for s in seq.strips_all}

    # Select only the strip we want to split
    for s in seq.strips_all:
        s.select = (s == strip)

    # Set current frame to the split point
    bpy.context.scene.frame_current = frame

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
    seq = bpy.context.scene.sequence_editor
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

# ----- Demonstration Functions -----

def demonstrate_trimming(test_media_dir=None):
    """
    Demonstrate trimming clips by adjusting their start and end offsets.
    
    This sets up a fresh scene with one video clip and demonstrates trimming
    frames from both the start and end of the clip.
    """
    print("\n=== DEMONSTRATION: TRIMMING CLIPS ===\n")
    
    # Setup
    scene = bpy.context.scene
    seq_editor = vse_utils.ensure_sequence_editor(scene)
    
    # Find test media directory
    if test_media_dir is None:
        test_media_dir = vse_utils.find_test_media_dir()
    
    # Set up a test scene with just one clip
    video_files = ["SampleVideo_1280x720_2mb.mp4"]
    clips = vse_utils.setup_test_sequence(seq_editor, test_media_dir, video_files)
    
    if not clips:
        print("No clips were added. Make sure test_media_dir contains valid video files.")
        return False
    
    # Get the video and audio strips
    video_strip, audio_strip = clips[0]
    
    # Print the initial state
    vse_utils.print_sequence_info(seq_editor, "Before Trimming")
    
    print(f"\nOriginal clip '{video_strip.name}' duration: {video_strip.frame_duration} frames")
    
    # Trim 24 frames from the start of both video and audio
    trim_strip_start(video_strip, 24)
    trim_strip_start(audio_strip, 24)
    
    # Trim 24 frames from the end of both video and audio
    trim_strip_end(video_strip, 24)
    trim_strip_end(audio_strip, 24)
    
    print(f"After trimming, '{video_strip.name}' final duration: {video_strip.frame_final_duration} frames")
    
    # Print the final state
    vse_utils.print_sequence_info(seq_editor, "After Trimming")
    
    return True

def demonstrate_splitting(test_media_dir=None):
    """
    Demonstrate splitting a clip at a specific frame.
    
    This sets up a fresh scene with one video clip and demonstrates 
    splitting it into two separate clips at a specified frame.
    """
    print("\n=== DEMONSTRATION: SPLITTING CLIPS ===\n")
    
    # Setup
    scene = bpy.context.scene
    seq_editor = vse_utils.ensure_sequence_editor(scene)
    
    # Find test media directory
    if test_media_dir is None:
        test_media_dir = vse_utils.find_test_media_dir()
    
    # Set up a test scene with just one clip
    video_files = ["SampleVideo_1280x720_2mb.mp4"]
    clips = vse_utils.setup_test_sequence(seq_editor, test_media_dir, video_files)
    
    if not clips:
        print("No clips were added. Make sure test_media_dir contains valid video files.")
        return False
    
    # Get the video and audio strips
    video_strip, audio_strip = clips[0]
    
    # Print the initial state
    vse_utils.print_sequence_info(seq_editor, "Before Splitting")
    
    # Calculate a split point near the middle
    middle_frame = int(video_strip.frame_start + (video_strip.frame_final_duration // 2))
    print(f"\nSplitting '{video_strip.name}' at frame {middle_frame}")
    
    # Split the video and audio strips
    video_parts = split_strip(video_strip, middle_frame)
    audio_parts = split_strip(audio_strip, middle_frame)
    
    if video_parts[0] and video_parts[1]:
        print(f"Successfully split video into two parts: '{video_parts[0].name}' and '{video_parts[1].name}'")
    else:
        print("Warning: Video split operation did not produce two valid parts")
    
    if audio_parts[0] and audio_parts[1]:
        print(f"Successfully split audio into two parts: '{audio_parts[0].name}' and '{audio_parts[1].name}'")
    else:
        print("Warning: Audio split operation did not produce two valid parts")
    
    # Print the final state
    vse_utils.print_sequence_info(seq_editor, "After Splitting")
    
    return True

def demonstrate_segment_removal(test_media_dir=None):
    """
    Demonstrate removing a segment from the middle of a clip.
    
    This sets up a fresh scene with one longer video clip, removes a 
    segment from the middle, and closes the gap.
    """
    print("\n=== DEMONSTRATION: REMOVING SEGMENTS ===\n")
    
    # Setup
    scene = bpy.context.scene
    seq_editor = vse_utils.ensure_sequence_editor(scene)
    
    # Find test media directory
    if test_media_dir is None:
        test_media_dir = vse_utils.find_test_media_dir()
    
    # Set up a test scene with one longer clip
    video_files = ["SampleVideo_1280x720_5mb.mp4"]  # Using the longer 5mb sample
    clips = vse_utils.setup_test_sequence(seq_editor, test_media_dir, video_files)
    
    if not clips:
        print("No clips were added. Make sure test_media_dir contains valid video files.")
        return False
    
    # Get the video and audio strips
    video_strip, audio_strip = clips[0]
    
    # Print the initial state
    vse_utils.print_sequence_info(seq_editor, "Before Segment Removal")
    
    # Define a segment to remove (e.g., 1/4 through to 1/2 through)
    segment_start = video_strip.frame_start + (video_strip.frame_final_duration // 4)
    segment_end = video_strip.frame_start + (video_strip.frame_final_duration // 2)
    
    print(f"\nRemoving middle segment from '{video_strip.name}'")
    print(f"  Original duration: {video_strip.frame_final_duration} frames")
    print(f"  Removing segment from frame {segment_start} to {segment_end}")
    
    # Remove segment from video and audio
    video_result = remove_segment(video_strip, segment_start, segment_end)
    audio_result = remove_segment(audio_strip, segment_start, segment_end)
    
    # Check results
    if video_result[0] and video_result[1]:
        # Calculate new duration
        new_duration = video_result[0].frame_final_duration + video_result[1].frame_final_duration
        print(f"  New combined video duration: {new_duration} frames")
        print(f"  Removed {video_strip.frame_final_duration - new_duration} frames")
    else:
        print("  Warning: Segment removal did not produce expected video parts")
    
    # Print the final state
    vse_utils.print_sequence_info(seq_editor, "After Segment Removal")
    
    return True

def demonstrate_slip_edit(test_media_dir=None):
    """
    Demonstrate slip editing a clip to change content without changing duration.
    
    This sets up a fresh scene with one video clip and demonstrates 
    slip editing to adjust which portion of the source content is displayed
    without changing the clip's position or duration on the timeline.
    """
    print("\n=== DEMONSTRATION: SLIP EDITING ===\n")
    
    # Setup
    scene = bpy.context.scene
    seq_editor = vse_utils.ensure_sequence_editor(scene)
    
    # Find test media directory
    if test_media_dir is None:
        test_media_dir = vse_utils.find_test_media_dir()
    
    # Set up a test scene with just one clip
    video_files = ["SampleVideo_1280x720_2mb.mp4"]
    clips = vse_utils.setup_test_sequence(seq_editor, test_media_dir, video_files)
    
    if not clips:
        print("No clips were added. Make sure test_media_dir contains valid video files.")
        return False
    
    # Get the video strip (we'll only slip the video in this demo)
    video_strip = clips[0][0]
    
    # Print the initial state
    vse_utils.print_sequence_info(seq_editor, "Before Slip Edit")
    
    print(f"\nPerforming slip edit on '{video_strip.name}'")
    print(f"  Before slip: offset_start={video_strip.frame_offset_start}, offset_end={video_strip.frame_offset_end}")
    print(f"  Timeline position: {video_strip.frame_start}-{video_strip.frame_final_end} (unchanged by slip)")
    
    # Slip by 12 frames (show later content)
    slip_strip(video_strip, 12)
    
    print(f"  After slip: offset_start={video_strip.frame_offset_start}, offset_end={video_strip.frame_offset_end}")
    print(f"  Timeline position: {video_strip.frame_start}-{video_strip.frame_final_end} (still same position)")
    
    # Print the final state
    vse_utils.print_sequence_info(seq_editor, "After Slip Edit")
    
    return True

def demonstrate_moving_strip(test_media_dir=None):
    """
    Demonstrate moving a strip to a different time or channel.
    
    This sets up a fresh scene with one video clip and demonstrates
    moving it to a different channel and/or position on the timeline.
    """
    print("\n=== DEMONSTRATION: MOVING STRIPS ===\n")
    
    # Setup
    scene = bpy.context.scene
    seq_editor = vse_utils.ensure_sequence_editor(scene)
    
    # Find test media directory
    if test_media_dir is None:
        test_media_dir = vse_utils.find_test_media_dir()
    
    # Set up a test scene with two clips for better visualization of movement
    video_files = ["SampleVideo_1280x720_1mb.mp4", "SampleVideo_1280x720_2mb.mp4"]
    clips = vse_utils.setup_test_sequence(seq_editor, test_media_dir, video_files)
    
    if not clips or len(clips) < 2:
        print("Not enough clips were added. Make sure test_media_dir contains valid video files.")
        return False
    
    # Get the first video strip to move
    video_strip = clips[0][0]
    
    # Print the initial state
    vse_utils.print_sequence_info(seq_editor, "Before Moving")
    
    print(f"\nMoving '{video_strip.name}'")
    print(f"  Original position: Channel {video_strip.channel}, Start Frame {video_strip.frame_start}")
    
    # Move to a higher channel and earlier position
    new_channel = 4
    new_frame = video_strip.frame_start + 10  # Move 10 frames later
    
    move_strip(video_strip, new_frame_start=new_frame, new_channel=new_channel)
    
    print(f"  New position: Channel {video_strip.channel}, Start Frame {video_strip.frame_start}")
    
    # Print the final state
    vse_utils.print_sequence_info(seq_editor, "After Moving")
    
    return True