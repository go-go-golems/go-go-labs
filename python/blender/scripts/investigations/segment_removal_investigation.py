# Investigation of Segment Removal Issues

"""
This script investigates why segment removal (cutting out a middle part of a strip)
was failing in our original implementation. It analyzes the behavior of two consecutive
split operations and explains the cause of the issue.

Segment removal involves:
1. Splitting a strip at the start of the segment to remove
2. Splitting the right part at the end of the segment
3. Removing the middle part
4. Closing the gap by moving the end part

The issue in our original implementation was related to how Blender positions strips
after split operations, which we'll examine in detail here.
"""

import bpy
import os
import sys

# Add script directory to path if needed
script_dir = '/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender/scripts'
if script_dir not in sys.path:
    sys.path.append(script_dir)

# Import utilities
import vse_utils

# ----- Split and Segment Removal Functions -----

def split_strip(strip, frame, select_right=True):
    """Split a strip at the specified frame and return both parts."""
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

    # Fallback heuristics if one side still missing - this part helped fix the issue
    if not left_part or not right_part:
        for s in seq.strips_all:
            if s.channel != strip.channel:
                continue
            if s.frame_final_end <= frame and (not left_part):
                left_part = s
            elif s.frame_start >= frame and (not right_part):
                right_part = s

    # Another fallback - look at frame_final_start vs the split point
    if not left_part or not right_part:
        for s in seq.strips_all:
            if s.channel != strip.channel:
                continue
            if s.frame_final_start < frame and (not left_part):
                left_part = s
            elif s.frame_final_start >= frame and (not right_part):
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

def remove_segment_original(strip, start_frame, end_frame):
    """Original version of remove_segment that had issues."""
    print(f"\nORIGINAL METHOD: Removing segment from {start_frame} to {end_frame} from '{strip.name}'")

    # Split at the start of segment
    part_a, part_b = split_strip(strip, start_frame, select_right=True)
    if not part_b:
        print("  [WARN] First split failed to create a right part - aborting")
        return (part_a, None)

    # Split at the end of segment (working on part_b)
    middle_part, part_c = split_strip(part_b, end_frame, select_right=True)
    if not middle_part:
        print("  [WARN] Second split produced no middle segment - aborting")
        return (part_a, part_c)

    # Remove the middle segment
    seq = bpy.context.scene.sequence_editor
    if middle_part and middle_part in seq.strips_all:
        seq.strips.remove(middle_part)
        print(f"  Removed segment '{middle_part.name}'")

    # Move part_c to close the gap
    if part_c and part_a:
        original_start = part_c.frame_start
        part_c.frame_start = part_a.frame_final_end
        print(f"  Moved '{part_c.name}' from frame {original_start} to {part_c.frame_start} (closing gap)")

    return (part_a, part_c)

def remove_segment_improved(strip, start_frame, end_frame):
    """Improved version of remove_segment with better strip identification."""
    print(f"\nIMPROVED METHOD: Removing segment from {start_frame} to {end_frame} from '{strip.name}'")

    # First split - get left and middle+right portions
    part_a, remainder = split_strip(strip, start_frame, select_right=True)
    
    if not remainder:
        print("  [WARN] First split failed to create a remainder part - aborting")
        return (part_a, None)
    
    # Show the state between first and second splits
    print("\n  STATE AFTER FIRST SPLIT:")
    seq = bpy.context.scene.sequence_editor
    strip_ids = {}
    for s in seq.strips_all:
        vse_utils.print_strip_details(s, f"  Strip {s.name}")
        strip_ids[s.name] = s.as_pointer()
    
    # Important: Record which strip is our remainder for the second split
    remainder_ptr = remainder.as_pointer()
    print(f"\n  Using strip '{remainder.name}' (ptr: {remainder_ptr}) for second split")
    
    # Second split - divide remainder into middle and right portions
    middle, right = split_strip(remainder, end_frame, select_right=True)
    
    if not right:
        print("  [WARN] Second split didn't produce a right part - using fallbacks")
        # Try to find the right part by checking all strips
        for s in seq.strips_all:
            # Skip the strip we already identified as left part
            if s.as_pointer() == part_a.as_pointer():
                continue
            # Skip the strip we used for the second split (middle)
            if s.as_pointer() == remainder_ptr:
                continue
            # This must be the newly created strip
            right = s
            print(f"  Found potential right part: {right.name} (ptr: {right.as_pointer()})")
            break
    
    if not middle:
        print("  [WARN] Second split didn't maintain the middle part properly")
        middle = remainder  # Use the remainder as middle if we can't find it
    
    print("\n  IDENTIFIED PARTS AFTER SPLITS:")
    print(f"  Left part: {part_a.name if part_a else 'None'}")
    print(f"  Middle part: {middle.name if middle else 'None'}")
    print(f"  Right part: {right.name if right else 'None'}")
    
    # Remove the middle segment
    if middle and middle in seq.strips_all:
        try:
            seq.strips.remove(middle)
            print(f"  Removed segment '{middle.name}'")
        except Exception as e:
            print(f"  Error removing middle segment: {e}")
    
    # Move right part to close the gap
    if right and part_a:
        try:
            original_start = right.frame_start
            right.frame_start = part_a.frame_final_end
            print(f"  Moved '{right.name}' from frame {original_start} to {right.frame_start} (closing gap)")
        except Exception as e:
            print(f"  Error moving right part: {e}")
    
    return (part_a, right)

# ----- Test Framework -----

def setup_test_environment():
    """Set up a clean test environment with a single clip."""
    scene = bpy.context.scene
    seq_editor = vse_utils.ensure_sequence_editor(scene)
    
    # Clear any existing strips
    vse_utils.clear_all_strips(seq_editor)
    
    # Find test media directory
    test_media_dir = vse_utils.find_test_media_dir()
    
    # Add a single long video clip
    video_files = ["SampleVideo_1280x720_5mb.mp4"]  # Use the longer 5mb sample
    clips = vse_utils.setup_test_sequence(seq_editor, test_media_dir, video_files)
    
    if not clips:
        print("Failed to add test clip. Make sure test media is available.")
        return None, None
        
    return scene, seq_editor

def compare_methods():
    """Compare original and improved segment removal methods."""
    print("\n=== COMPARISON: ORIGINAL vs IMPROVED SEGMENT REMOVAL ===\n")
    
    # Set up environment
    scene, seq_editor = setup_test_environment()
    if not seq_editor:
        return
    
    # Get the video strip
    video_strip = None
    for strip in seq_editor.strips_all:
        if strip.type == 'MOVIE':
            video_strip = strip
            break
    
    if not video_strip:
        print("No video strip found!")
        return
    
    # Print initial state
    vse_utils.print_sequence_info(seq_editor, "Initial State")
    
    # Define segment to remove (1/4 to 1/2 through the clip)
    segment_start = video_strip.frame_start + (video_strip.frame_final_duration // 4)
    segment_end = video_strip.frame_start + (video_strip.frame_final_duration // 2)
    
    # Get a clean copy of the original strip for the second test
    original_strip_props = {
        'name': video_strip.name,
        'filepath': video_strip.filepath,
        'channel': video_strip.channel,
        'frame_start': video_strip.frame_start
    }
    
    # Test the original method
    print("\n--- TESTING ORIGINAL METHOD ---")
    vse_utils.print_strip_details(video_strip, "Before removal")
    result_original = remove_segment_original(video_strip, segment_start, segment_end)
    
    # Print result
    print("\nResult of original method:")
    if result_original[0] and result_original[1]:
        print(f"SUCCESS - Created two parts: {result_original[0].name} and {result_original[1].name}")
    else:
        print(f"FAILURE - Did not create two valid parts: {result_original[0]} and {result_original[1]}")
    
    vse_utils.print_sequence_info(seq_editor, "After Original Method")
    
    # Clear and set up again for the improved method
    vse_utils.clear_all_strips(seq_editor)
    video_strip = seq_editor.strips.new_movie(
        name=original_strip_props['name'],
        filepath=original_strip_props['filepath'],
        channel=original_strip_props['channel'],
        frame_start=original_strip_props['frame_start']
    )
    
    # Test the improved method
    print("\n--- TESTING IMPROVED METHOD ---")
    vse_utils.print_strip_details(video_strip, "Before removal")
    result_improved = remove_segment_improved(video_strip, segment_start, segment_end)
    
    # Print result
    print("\nResult of improved method:")
    if result_improved[0] and result_improved[1]:
        print(f"SUCCESS - Created two parts: {result_improved[0].name} and {result_improved[1].name}")
    else:
        print(f"FAILURE - Did not create two valid parts: {result_improved[0]} and {result_improved[1]}")
    
    vse_utils.print_sequence_info(seq_editor, "After Improved Method")
    
    # Final analysis
    print("\n=== ANALYSIS ===\n")
    print("Root causes of segment removal issues in original implementation:")
    print("1. Split operation doesn't position strips as intuitively expected")
    print("2. The strip returned from first split might not be identified properly for second split")
    print("3. Frame offsets rather than actual positions determine what content is shown")
    print("\nOur improved implementation fixes these issues by:")
    print("1. Tracking strip identity across operations using as_pointer()")
    print("2. Using multiple fallback heuristics to identify parts after splitting")
    print("3. Adding defensive code to handle unexpected strip states")

# Main function
def main():
    """Run the investigation."""
    print("\nINVESTIGATION: SEGMENT REMOVAL ISSUES\n")
    
    # Compare methods
    compare_methods()
    
    print("\nINVESTIGATION COMPLETE")

# Run the investigation
if __name__ == "__main__":
    main()