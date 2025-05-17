# Test script for video interleaving/flickering effect

import bpy
import os
import sys
import importlib

# Add the scripts directory to the path for imports
scripts_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if scripts_dir not in sys.path:
    sys.path.append(scripts_dir)
    
# Make sure utils directory is also in path
utils_dir = os.path.join(scripts_dir, 'utils')
if utils_dir not in sys.path:
    sys.path.append(utils_dir)
    
# Import our utility modules
import vse_utils
importlib.reload(vse_utils)
import vse_effects
importlib.reload(vse_effects)

def main():
    """Set up a test scene and demonstrate the flicker effect."""
    # Get active scene and ensure we have a sequence editor
    scene = vse_utils.get_active_scene()
    seq_editor = vse_utils.ensure_sequence_editor(scene)
    
    # Clear any existing strips
    vse_utils.clear_all_strips(seq_editor)
    
    # Find test media directory
    test_media_dir = vse_utils.find_test_media_dir()
    print(f"Using media directory: {test_media_dir}")
    
    # Set up two video clips side by side instead of sequential
    video_files = [
        "SampleVideo_1280x720_2mb.mp4",
        "SampleVideo_1280x720_5mb.mp4"
    ]
    
    # Place videos on separate channels (1 and 3) for clarity
    video1_path = os.path.join(test_media_dir, video_files[0])
    video2_path = os.path.join(test_media_dir, video_files[1])
    
    if not (os.path.exists(video1_path) and os.path.exists(video2_path)):
        print("Test videos not found. Please ensure test media directory contains sample videos.")
        return False
    
    # Add first video
    strip1 = vse_utils.add_movie_strip(
        seq_editor, 
        video1_path, 
        channel=1, 
        frame_start=1,
        name="Video1"
    )
    
    # Add second video
    strip2 = vse_utils.add_movie_strip(
        seq_editor, 
        video2_path, 
        channel=3, 
        frame_start=1,
        name="Video2"
    )
    
    # Apply the flicker effect
    print("\nCreating flicker effect...")
    interval_frames = 10  # Switch videos every 10 frames
    output_channel = 5    # Place output on channel 5
    
    flicker_strips = vse_effects.create_flicker_effect(
        seq_editor, strip1, strip2, 
        interval_frames=interval_frames, 
        output_channel=output_channel
    )
    
    # Print information about the scene
    vse_utils.print_sequence_info(seq_editor, "Final Sequence with Flicker Effect")
    
    # Set end frame to match our content
    max_frame = max(strip.frame_final_end for strip in seq_editor.sequences_all)
    scene.frame_end = max_frame
    
    # Position timeline at the start
    scene.frame_current = 1
    
    print("\nFlicker effect created!")
    print(f"- Flickering between clips every {interval_frames} frames")
    print(f"- Original clips on channels 1 and 3")
    print(f"- Interleaved result on channel {output_channel}")
    print("\nPlay the sequence to see the effect!")
    
    return True

# Run the main function
if __name__ == "__main__":
    main()