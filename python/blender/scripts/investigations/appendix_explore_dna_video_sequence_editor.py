# Appendix: Exploring Blender VSE Split Operation Behavior

"""
This script specifically demonstrates and explains the behavior of Blender's 
split operation in the Video Sequence Editor. It visualizes how Blender handles
splitting strips, which explains the issues we encountered in our implementation.

The key finding is that Blender's split operation doesn't behave in the intuitive
way of creating two independent strips with separate timeline positions. Instead,
it often maintains the original strip's position information while adjusting offset 
values to simulate the split.
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

# Set up a clean environment
def setup_test_environment():
    """Set up a clean test environment with a single clip."""
    scene = bpy.context.scene
    seq_editor = vse_utils.ensure_sequence_editor(scene)
    
    # Clear any existing strips
    vse_utils.clear_all_strips(seq_editor)
    
    # Find test media directory
    test_media_dir = vse_utils.find_test_media_dir()
    
    # Add a single video clip
    video_files = ["SampleVideo_1280x720_2mb.mp4"]
    clips = vse_utils.setup_test_sequence(seq_editor, test_media_dir, video_files)
    
    if not clips:
        print("Failed to add test clip. Make sure test media is available.")
        return None, None
        
    return scene, seq_editor

# Print detailed information about a strip
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

# Perform a split operation and analyze the results
def analyze_split_operation(strip, frame):
    """Split a strip and analyze the resulting strips in detail."""
    print("\n=== SPLIT OPERATION ANALYSIS ===\n")
    
    # Before split
    print_strip_details(strip, "BEFORE SPLIT")
    
    # Record all strips before the split
    seq_editor = bpy.context.scene.sequence_editor
    pre_split_ids = {s.as_pointer(): s.name for s in seq_editor.strips_all}
    
    # Ensure the strip is selected
    for s in seq_editor.strips_all:
        s.select = (s == strip)
    
    # Set the current frame and perform the split
    bpy.context.scene.frame_current = frame
    print(f"\nSplitting at frame {frame}...")
    
    try:
        bpy.ops.sequencer.split(frame=frame, channel=strip.channel, type='SOFT')
        print("Split operation completed.")
    except Exception as e:
        print(f"Split operation failed: {e}")
        return
    
    # Analyze the results
    print("\n=== RESULTS ===\n")
    print("Strips after split operation:")
    
    # Collect all strips after the split
    post_split_strips = {}
    for s in seq_editor.strips_all:
        post_split_strips[s.as_pointer()] = s
        
    # Identify which strips are new vs. modified
    print("\nSTRIP IDENTIFICATION:")
    for ptr, s in post_split_strips.items():
        if ptr in pre_split_ids:
            status = "MODIFIED (original strip)"
        else:
            status = "NEW (created by split)"
        print(f"  {s.name}: {status}")
    
    # Analyze each strip in detail
    print("\nDETAILED STRIP ANALYSIS:")
    for s in seq_editor.strips_all:
        print_strip_details(s)
        
    # Explain the key observation
    print("\n=== KEY OBSERVATION ===\n")
    print("Contrary to the expected behavior where splitting would create two separate strips")
    print("with distinct timeline positions (e.g., one at 1-100 and one at 100-200), Blender often:")
    print("  1. Keeps the original strip in its original position (e.g., 1-200)")
    print("  2. Creates a new strip ALSO starting at the original position (e.g., 1-200)")
    print("  3. Uses frame_offset_start/end on both strips to control which parts of the")
    print("     source content are actually shown")
    print("\nThis explains why our identification based on frame_final_end == split_frame")
    print("did not reliably identify the left and right parts after splitting.")
    
# Main demonstration
def main():
    """Run the demonstration of split operation behavior."""
    print("\nDEMONSTRATION: UNDERSTANDING BLENDER VSE SPLIT OPERATION\n")
    
    # Set up environment
    scene, seq_editor = setup_test_environment()
    if not seq_editor:
        return
    
    # Get the first (and only) clip
    video_strip = None
    for strip in seq_editor.strips_all:
        if strip.type == 'MOVIE':
            video_strip = strip
            break
    
    if not video_strip:
        print("No video strip found!")
        return
    
    # Calculate a split point near the middle
    split_frame = int(video_strip.frame_start + (video_strip.frame_final_duration // 2))
    
    # Analyze what happens during a split operation
    analyze_split_operation(video_strip, split_frame)
    
    print("\nDEMONSTRATION COMPLETE")
    print("This explains why working with split strips requires careful tracking")
    print("of the strips using as_pointer() and multiple identification heuristics.")

# Run the demonstration
if __name__ == "__main__":
    main()