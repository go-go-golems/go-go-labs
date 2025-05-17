# VSE Effects Utility functions for Blender Python scripts
# Contains specialized effects like interleaving/flickering videos

import bpy
import math
import vse_utils

def interleave_strips(seq_editor, strip1, strip2, interval_frames=5, channel_offset=2):
    """
    Create a flickering effect by interleaving two video strips at regular intervals.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        strip1 (bpy.types.Sequence): First video strip
        strip2 (bpy.types.Sequence): Second video strip 
        interval_frames (int): Number of frames before switching to the other video
        channel_offset (int): Channel spacing above the input strips
    
    Returns:
        list: The created interleaved strips
    """
    # Determine which strip starts first and align calculations
    start_frame = min(strip1.frame_final_start, strip2.frame_final_start)
    end_frame = min(strip1.frame_final_end, strip2.frame_final_end)
    duration = end_frame - start_frame
    
    if duration <= 0:
        print("Error: Videos don't overlap")
        return []
    
    # Calculate new channel positions
    output_channel = max(strip1.channel, strip2.channel) + channel_offset
    result_strips = []
    
    # Calculate number of segments needed
    num_segments = math.ceil(duration / interval_frames)
    
    # Create segments directly (no need for temporary strips)
    for i in range(num_segments):
        segment_start = start_frame + (i * interval_frames)
        segment_end = min(segment_start + interval_frames, end_frame)
        
        # Get the original strip for this segment based on alternating pattern
        source_strip = strip1 if i % 2 == 0 else strip2
        
        # Calculate source frame offsets
        source_offset = segment_start - source_strip.frame_final_start
        
        # Determine the type of strip and create appropriate new strip
        if source_strip.type == 'MOVIE':
            # Create a new movie strip using slice of the original
            new_strip = seq_editor.sequences.new_movie(
                name=f"{source_strip.name}_seg{i}",
                filepath=source_strip.filepath,
                channel=output_channel,
                frame_start=segment_start
            )
            
            # Set correct position within source movie
            new_strip.frame_offset_start = source_strip.frame_offset_start + source_offset
            # Ensure the final end is at the correct position
            new_strip.frame_final_end = segment_end
            # Force the duration to be correct
            duration = segment_end - segment_start
            # Update strip properties to ensure correct duration display
            new_strip.frame_final_duration = duration
            # Debugging
            print(f"  Segment {i}: {segment_start}-{segment_end} (duration={duration})")
            
        elif source_strip.type == 'SOUND':
            # Create a new sound strip using slice of the original
            new_strip = seq_editor.sequences.new_sound(
                name=f"{source_strip.name}_seg{i}",
                filepath=source_strip.filepath,
                channel=output_channel,
                frame_start=segment_start
            )
            
            # Set correct position within source sound
            new_strip.frame_offset_start = source_strip.frame_offset_start + source_offset
            # Ensure the final end is at the correct position
            new_strip.frame_final_end = segment_end
            # Force the duration to be correct
            duration = segment_end - segment_start
            # Update strip properties to ensure correct duration display
            new_strip.frame_final_duration = duration
            # Debugging
            print(f"  Segment {i}: {segment_start}-{segment_end} (duration={duration})")
        
        # Add to our result list
        result_strips.append(new_strip)
    
    return result_strips

def create_flicker_effect(seq_editor, strip1, strip2, interval_frames=5, output_channel=None):
    """
    Create a flickering effect between two strips with automatic channel placement.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        strip1 (bpy.types.Sequence): First video strip
        strip2 (bpy.types.Sequence): Second video strip
        interval_frames (int): Number of frames before switching videos
        output_channel (int, optional): Specific output channel, or auto-selected if None
    
    Returns:
        list: The created interleaved strips
    """
    # Find an appropriate output channel if not specified
    if output_channel is None:
        output_channel = max(strip1.channel, strip2.channel) + 2
    
    # Create the interleaved effect
    interleaved_strips = interleave_strips(
        seq_editor, strip1, strip2, interval_frames, 
        channel_offset=output_channel - max(strip1.channel, strip2.channel)
    )
    
    print(f"Created flicker effect with {len(interleaved_strips)} segments")
    return interleaved_strips