# Transition Utilities for Blender VSE
# Helper functions for Chapter 4: Transitions and Fades

import bpy # type: ignore
import os

# Common shortcuts for Blender data and context
D = bpy.data
C = bpy.context

def create_crossfade(seq_editor, strip1, strip2, transition_duration, channel=None):
    """
    Create a crossfade transition between two strips.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor to add the transition to
        strip1 (bpy.types.Strip): The first strip (being faded out)
        strip2 (bpy.types.Strip): The second strip (being faded in)
        transition_duration (int): Duration of the transition in frames
        channel (int, optional): Channel to place the transition on. If None, places
                               above both input strips.
    
    Returns:
        bpy.types.Strip: The created transition strip
    """
    # Make sure strips overlap by at least the transition duration
    if strip2.frame_start > strip1.frame_final_end - transition_duration:
        print(f"Warning: Strips don't overlap enough for {transition_duration} frame transition")
        # Adjust strip2 to start earlier to create required overlap
        strip2.frame_start = strip1.frame_final_end - transition_duration
        print(f"  Adjusted strip2.frame_start to {strip2.frame_start}")
    
    # Calculate transition start and end frames
    trans_start = int(strip2.frame_start)
    trans_end = int(trans_start + transition_duration)
    
    # Choose channel if not specified
    if channel is None:
        channel = max(strip1.channel, strip2.channel) + 1
    
    # Create the crossfade effect
    transition = seq_editor.strips.new_effect(
        name=f"Cross_{strip1.name}_{strip2.name}",
        type='CROSS',
        channel=channel,
        frame_start=trans_start,
        frame_end=trans_end,
        seq1=strip1,
        seq2=strip2
    )
    
    print(f"Created crossfade from '{strip1.name}' to '{strip2.name}'")
    print(f"  Duration: {transition_duration} frames ({trans_start}-{trans_end})")
    print(f"  Channel: {channel}")
    
    return transition

def create_gamma_crossfade(seq_editor, strip1, strip2, transition_duration, channel=None):
    """
    Create a gamma-corrected crossfade transition between two strips.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor to add the transition to
        strip1 (bpy.types.Strip): The first strip (being faded out)
        strip2 (bpy.types.Strip): The second strip (being faded in)
        transition_duration (int): Duration of the transition in frames
        channel (int, optional): Channel to place the transition on. If None, places
                               above both input strips.
    
    Returns:
        bpy.types.Strip: The created transition strip
    """
    # Make sure strips overlap by at least the transition duration
    if strip2.frame_start > strip1.frame_final_end - transition_duration:
        print(f"Warning: Strips don't overlap enough for {transition_duration} frame transition")
        # Adjust strip2 to start earlier to create required overlap
        strip2.frame_start = strip1.frame_final_end - transition_duration
        print(f"  Adjusted strip2.frame_start to {strip2.frame_start}")
    
    # Calculate transition start and end frames
    trans_start = int(strip2.frame_start)
    trans_end = int(trans_start + transition_duration)
    
    # Choose channel if not specified
    if channel is None:
        channel = max(strip1.channel, strip2.channel) + 1
    
    # Create the gamma crossfade effect
    transition = seq_editor.strips.new_effect(
        name=f"GammaCross_{strip1.name}_{strip2.name}",
        type='GAMMA_CROSS',
        channel=channel,
        frame_start=trans_start,
        frame_end=trans_end,
        seq1=strip1,
        seq2=strip2
    )
    
    print(f"Created gamma crossfade from '{strip1.name}' to '{strip2.name}'")
    print(f"  Duration: {transition_duration} frames ({trans_start}-{trans_end})")
    print(f"  Channel: {channel}")
    
    return transition

def create_wipe(seq_editor, strip1, strip2, transition_duration, wipe_type='SINGLE', angle=0.0, channel=None):
    """
    Create a wipe transition between two strips.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor to add the transition to
        strip1 (bpy.types.Strip): The first strip (being wiped away)
        strip2 (bpy.types.Strip): The second strip (being wiped in)
        transition_duration (int): Duration of the transition in frames
        wipe_type (str): The type of wipe ('SINGLE', 'DOUBLE', 'IRIS', 'CLOCK')
        angle (float): Angle of the wipe in radians (0.0 is right to left)
        channel (int, optional): Channel to place the transition on. If None, places
                               above both input strips.
    
    Returns:
        bpy.types.Strip: The created transition strip
    """
    # Make sure strips overlap by at least the transition duration
    if strip2.frame_start > strip1.frame_final_end - transition_duration:
        print(f"Warning: Strips don't overlap enough for {transition_duration} frame transition")
        # Adjust strip2 to start earlier to create required overlap
        strip2.frame_start = strip1.frame_final_end - transition_duration
        print(f"  Adjusted strip2.frame_start to {strip2.frame_start}")
    
    # Calculate transition start and end frames
    trans_start = int(strip2.frame_start)
    trans_end = int(trans_start + transition_duration)
    
    # Choose channel if not specified
    if channel is None:
        channel = max(strip1.channel, strip2.channel) + 1
    
    # Create the wipe effect
    transition = seq_editor.strips.new_effect(
        name=f"Wipe_{strip1.name}_{strip2.name}",
        type='WIPE',
        channel=channel,
        frame_start=trans_start,
        frame_end=trans_end,
        seq1=strip1,
        seq2=strip2
    )
    
    # Configure wipe properties
    transition.transition_type = wipe_type
    transition.direction = 'IN' if angle == 0.0 else 'OUT'
    transition.angle = angle
    
    print(f"Created wipe transition from '{strip1.name}' to '{strip2.name}'")
    print(f"  Type: {wipe_type}, Angle: {angle} radians")
    print(f"  Duration: {transition_duration} frames ({trans_start}-{trans_end})")
    print(f"  Channel: {channel}")
    
    return transition

def create_audio_fade(sound_strip, fade_type='IN', duration_frames=24):
    """
    Create a volume fade for a sound strip.
    
    Args:
        sound_strip (bpy.types.SoundStrip): The sound strip to fade
        fade_type (str): The type of fade ('IN' or 'OUT')
        duration_frames (int): Duration of the fade in frames
    
    Returns:
        bool: Success or failure
    """
    if not sound_strip or sound_strip.type != 'SOUND':
        print("Error: Invalid sound strip provided")
        return False
    
    try:
        # Calculate start and end frames based on fade type
        if fade_type == 'IN':
            start_frame = int(sound_strip.frame_start)
            end_frame = int(start_frame + duration_frames)
            start_vol = 0.0
            end_vol = 1.0
        elif fade_type == 'OUT':
            end_frame = int(sound_strip.frame_final_end)
            start_frame = int(end_frame - duration_frames)
            start_vol = 1.0
            end_vol = 0.0
        else:
            print(f"Error: Invalid fade type '{fade_type}'. Use 'IN' or 'OUT'.")
            return False
            
        # Set and keyframe the volume at start frame
        sound_strip.volume = start_vol
        sound_strip.keyframe_insert("volume", frame=start_frame)
        
        # Set and keyframe the volume at end frame
        sound_strip.volume = end_vol
        sound_strip.keyframe_insert("volume", frame=end_frame)
        
        print(f"Created audio {fade_type} fade for '{sound_strip.name}'")
        print(f"  Duration: {duration_frames} frames ({start_frame}-{end_frame})")
        print(f"  Volume: {start_vol} to {end_vol}")
        
        return True
    except Exception as e:
        print(f"Error creating audio fade: {e}")
        return False

def create_audio_crossfade(sound1, sound2, overlap_frames=24):
    """
    Create a crossfade between two audio strips by keyframing their volumes.
    
    Args:
        sound1 (bpy.types.SoundStrip): The first sound strip (fading out)
        sound2 (bpy.types.SoundStrip): The second sound strip (fading in)
        overlap_frames (int): Duration of the overlap/crossfade in frames
    
    Returns:
        bool: Success or failure
    """
    if not sound1 or not sound2 or sound1.type != 'SOUND' or sound2.type != 'SOUND':
        print("Error: Invalid sound strips provided")
        return False
    
    try:
        # Ensure the strips are overlapping by the required amount
        if sound2.frame_start > sound1.frame_final_end - overlap_frames:
            print(f"Warning: Sound strips don't overlap enough for {overlap_frames} frame crossfade")
            # Adjust sound2 to start earlier
            sound2.frame_start = sound1.frame_final_end - overlap_frames
            print(f"  Adjusted sound2.frame_start to {sound2.frame_start}")
        
        # Calculate crossfade start and end
        fade_start = int(sound2.frame_start)
        fade_end = int(fade_start + overlap_frames)
        
        # Keyframe sound1 volume (fading out)
        sound1.volume = 1.0
        sound1.keyframe_insert("volume", frame=fade_start)
        sound1.volume = 0.0
        sound1.keyframe_insert("volume", frame=fade_end)
        
        # Keyframe sound2 volume (fading in)
        sound2.volume = 0.0
        sound2.keyframe_insert("volume", frame=fade_start)
        sound2.volume = 1.0
        sound2.keyframe_insert("volume", frame=fade_end)
        
        print(f"Created audio crossfade from '{sound1.name}' to '{sound2.name}'")
        print(f"  Duration: {overlap_frames} frames ({fade_start}-{fade_end})")
        
        return True
    except Exception as e:
        print(f"Error creating audio crossfade: {e}")
        return False

def create_fade_to_color(seq_editor, strip, fade_duration, fade_type='IN', color=(0,0,0), channel=None):
    """
    Create a fade from/to a solid color (often black) for a strip.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        strip (bpy.types.Strip): The strip to fade
        fade_duration (int): Duration of the fade in frames
        fade_type (str): Type of fade ('IN' = from color to clip, 'OUT' = from clip to color)
        color (tuple): RGB color values (0-1 for each component)
        channel (int, optional): Channel for the effect. If None, uses strip.channel + 1
    
    Returns:
        bpy.types.Strip: The created effect strip or None on failure
    """
    if not channel:
        channel = strip.channel + 1
    
    try:
        # Create a color strip with the specified color
        if fade_type == 'IN':
            # For fade in, color strip overlaps start of original strip
            color_start = int(strip.frame_start)
            color_end = int(color_start + fade_duration)
        else:  # fade_type == 'OUT'
            # For fade out, color strip overlaps end of original strip
            color_end = int(strip.frame_final_end)
            color_start = int(color_end - fade_duration)
        
        # Create the color strip
        color_strip = seq_editor.strips.new_effect(
            name=f"FadeColor_{strip.name}",
            type='COLOR',
            channel=channel,
            frame_start=color_start,
            frame_end=color_end
        )
        
        # Set the color
        color_strip.color = color
        
        # Create a crossfade between the strips
        if fade_type == 'IN':
            transition = seq_editor.strips.new_effect(
                name=f"FadeIn_{strip.name}",
                type='CROSS',
                channel=channel + 1,
                frame_start=color_start,
                frame_end=color_end,
                seq1=color_strip,
                seq2=strip
            )
        else:  # fade_type == 'OUT'
            transition = seq_editor.strips.new_effect(
                name=f"FadeOut_{strip.name}",
                type='CROSS',
                channel=channel + 1,
                frame_start=color_start,
                frame_end=color_end,
                seq1=strip,
                seq2=color_strip
            )
        
        print(f"Created fade {fade_type} to/from color for '{strip.name}'")
        print(f"  Duration: {fade_duration} frames ({color_start}-{color_end})")
        
        return transition
    except Exception as e:
        print(f"Error creating fade to color: {e}")
        return None