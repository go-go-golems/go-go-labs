# Transition Utilities for Blender VSE
# Helper functions for Chapter 4: Transitions and Fades

import bpy # type: ignore
import os
import traceback
from typing import Any, Optional # Added Optional and Any

# Common shortcuts for Blender data and context
D = bpy.data
C = bpy.context

def ensure_integer_frame(value: Any) -> int:
    """
    Safely converts a value to an integer frame number.
    Tries string and float conversion first to handle various input types like "1.0".
    Args:
        value: The value to convert.
    Returns:
        int: The integer frame number.
    Raises:
        ValueError: If the value cannot be converted to an integer.
    """
    try:
        # Convert to string first to handle Blender properties that might not be direct numbers
        return int(float(str(value)))
    except (TypeError, ValueError) as e:
        raise ValueError(f"Invalid frame value: '{value}' (type: {type(value).__name__}) cannot be converted to an integer. Error: {e}")

def create_crossfade(seq_editor: bpy.types.SequenceEditor,
                     strip1: bpy.types.Sequence,
                     strip2: bpy.types.Sequence,
                     transition_duration: int,
                     channel: Optional[int] = None) -> Optional[bpy.types.Sequence]:
    """
    Create a crossfade transition between two strips.
    The caller is responsible for ensuring strips overlap correctly for the duration.
    Transition typically starts at strip2.frame_start.

    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor.
        strip1 (bpy.types.Sequence): The first strip (being faded out).
        strip2 (bpy.types.Sequence): The second strip (being faded in).
        transition_duration (int): Duration of the transition in frames.
        channel (Optional[int]): Channel for the effect. Defaults to a channel
                                 above both input strips.

    Returns:
        Optional[bpy.types.Sequence]: The created transition strip or None on failure.
    """
    try:
        if not all(isinstance(s, bpy.types.Sequence) for s in [strip1, strip2]):
             print(f"Error: Invalid strip objects provided for crossfade ({strip1.name}, {strip2.name}). Expected bpy.types.Sequence.")
             return None
        if not isinstance(seq_editor, bpy.types.SequenceEditor):
            print("Error: Invalid sequence editor object provided for create_crossfade.")
            return None
        if transition_duration <= 0:
            print(f"Error: transition_duration ({transition_duration}) must be positive.")
            return None

        s1_start = ensure_integer_frame(strip1.frame_start)
        s1_final_end = ensure_integer_frame(strip1.frame_final_end)
        s2_start = ensure_integer_frame(strip2.frame_start)
        
        # Transition effect will start at s2_start and last for transition_duration
        trans_start = s2_start
        trans_end = ensure_integer_frame(trans_start + transition_duration)

        # Check for sufficient overlap:
        # Strip1 must effectively end at or after trans_end.
        # Strip2 must start such that strip1 can fade out over it for transition_duration.
        # More simply, the overlapping period between strip1 and strip2 must be >= transition_duration,
        # and the transition must occur within this overlap.
        # The effect is placed based on strip2's start.
        # So, strip1 must cover [s2_start, s2_start + transition_duration]
        # and strip2 must cover [s2_start, s2_start + transition_duration]
        
        # Condition from original logic: strip2 must start by (strip1.frame_final_end - transition_duration)
        # If s2_start > s1_final_end - transition_duration, then s1_final_end < s2_start + transition_duration
        if s2_start > ensure_integer_frame(s1_final_end - transition_duration):
            print(f"Error: Crossfade ({strip1.name} to {strip2.name}) - Strips do not overlap sufficiently. "
                  f"'{strip2.name}' (starts {s2_start}) must start on or before frame "
                  f"{ensure_integer_frame(s1_final_end - transition_duration)} (i.e., '{strip1.name}' end {s1_final_end} - duration {transition_duration}) "
                  f"to allow for a {transition_duration}-frame crossfade.")
            return None

        if channel is None:
            channel_num = ensure_integer_frame(max(strip1.channel, strip2.channel) + 1)
        else:
            channel_num = ensure_integer_frame(channel)

        transition = seq_editor.strips.new_effect(
            name=f"Cross_{strip1.name}_{strip2.name}",
            type='CROSS',
            channel=channel_num,
            frame_start=trans_start,
            frame_end=trans_end,
            seq1=strip1,
            seq2=strip2
        )

        print(f"Successfully created crossfade: {transition.name} (frames {trans_start}-{trans_end}) on channel {channel_num}")
        return transition

    except ValueError as ve:
        print(f"Error creating crossfade due to invalid frame value: {ve}")
        traceback.print_exc()
        return None
    except Exception as e:
        print(f"Error creating crossfade for '{strip1.name}' and '{strip2.name}': {e}")
        traceback.print_exc()
        return None

def create_gamma_crossfade(seq_editor: bpy.types.SequenceEditor,
                           strip1: bpy.types.Sequence,
                           strip2: bpy.types.Sequence,
                           transition_duration: int,
                           channel: Optional[int] = None) -> Optional[bpy.types.Sequence]:
    """
    Create a gamma-corrected crossfade transition between two strips.
    The caller is responsible for ensuring strips overlap correctly.

    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor.
        strip1 (bpy.types.Sequence): The first strip (being faded out).
        strip2 (bpy.types.Sequence): The second strip (being faded in).
        transition_duration (int): Duration of the transition in frames.
        channel (Optional[int]): Channel for the effect. Defaults to a channel
                                 above both input strips.

    Returns:
        Optional[bpy.types.Sequence]: The created transition strip or None on failure.
    """
    try:
        if not all(isinstance(s, bpy.types.Sequence) for s in [strip1, strip2]):
             print(f"Error: Invalid strip objects for gamma_crossfade ({strip1.name}, {strip2.name}). Expected bpy.types.Sequence.")
             return None
        if not isinstance(seq_editor, bpy.types.SequenceEditor):
            print("Error: Invalid sequence editor object for create_gamma_crossfade.")
            return None
        if transition_duration <= 0:
            print(f"Error: transition_duration ({transition_duration}) must be positive.")
            return None

        s1_final_end = ensure_integer_frame(strip1.frame_final_end)
        s2_start = ensure_integer_frame(strip2.frame_start)

        trans_start = s2_start
        trans_end = ensure_integer_frame(trans_start + transition_duration)

        if s2_start > ensure_integer_frame(s1_final_end - transition_duration):
            print(f"Error: Gamma Crossfade ({strip1.name} to {strip2.name}) - Strips do not overlap sufficiently. "
                  f"'{strip2.name}' (starts {s2_start}) must start on or before frame "
                  f"{ensure_integer_frame(s1_final_end - transition_duration)} "
                  f"for a {transition_duration}-frame gamma crossfade.")
            return None

        if channel is None:
            channel_num = ensure_integer_frame(max(strip1.channel, strip2.channel) + 1)
        else:
            channel_num = ensure_integer_frame(channel)

        transition = seq_editor.strips.new_effect(
            name=f"GammaCross_{strip1.name}_{strip2.name}",
            type='GAMMA_CROSS',
            channel=channel_num,
            frame_start=trans_start,
            frame_end=trans_end,
            seq1=strip1,
            seq2=strip2
        )

        print(f"Successfully created gamma crossfade: {transition.name} (frames {trans_start}-{trans_end}) on channel {channel_num}")
        return transition

    except ValueError as ve:
        print(f"Error creating gamma crossfade due to invalid frame value: {ve}")
        traceback.print_exc()
        return None
    except Exception as e:
        print(f"Error creating gamma crossfade for '{strip1.name}' and '{strip2.name}': {e}")
        traceback.print_exc()
        return None

def create_wipe(seq_editor: bpy.types.SequenceEditor,
                strip_being_wiped_away: bpy.types.Sequence,
                strip_being_wiped_in: bpy.types.Sequence,
                transition_duration: int,
                wipe_type: str = 'SINGLE',
                angle: float = 0.0, # Radians
                channel: Optional[int] = None) -> Optional[bpy.types.Sequence]:
    """
    Create a wipe transition between two strips.
    The caller is responsible for ensuring strips overlap correctly.

    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor.
        strip_being_wiped_away (bpy.types.Sequence): The strip being wiped away.
        strip_being_wiped_in (bpy.types.Sequence): The strip being wiped in.
        transition_duration (int): Duration of the transition in frames.
        wipe_type (str): Type of wipe (e.g., 'SINGLE', 'DOUBLE', 'IRIS', 'CLOCK').
        angle (float): Angle of the wipe in radians.
        channel (Optional[int]): Channel for the effect. Defaults to a channel
                                 above both input strips.

    Returns:
        Optional[bpy.types.Sequence]: The created transition strip or None on failure.
    """
    try:
        if not all(isinstance(s, bpy.types.Sequence) for s in [strip_being_wiped_away, strip_being_wiped_in]):
             print(f"Error: Invalid strip objects for wipe. Expected bpy.types.Sequence.")
             return None
        if not isinstance(seq_editor, bpy.types.SequenceEditor):
            print("Error: Invalid sequence editor object for create_wipe.")
            return None
        if transition_duration <= 0:
            print(f"Error: transition_duration ({transition_duration}) must be positive.")
            return None
            
        s_away_final_end = ensure_integer_frame(strip_being_wiped_away.frame_final_end)
        s_in_start = ensure_integer_frame(strip_being_wiped_in.frame_start)

        trans_start = s_in_start
        trans_end = ensure_integer_frame(trans_start + transition_duration)

        if s_in_start > ensure_integer_frame(s_away_final_end - transition_duration):
            print(f"Error: Wipe ({strip_being_wiped_away.name} to {strip_being_wiped_in.name}) - Strips do not overlap sufficiently. "
                  f"'{strip_being_wiped_in.name}' (starts {s_in_start}) must start on or before frame "
                  f"{ensure_integer_frame(s_away_final_end - transition_duration)} "
                  f"for a {transition_duration}-frame wipe.")
            return None

        if channel is None:
            channel_num = ensure_integer_frame(max(strip_being_wiped_away.channel, strip_being_wiped_in.channel) + 1)
        else:
            channel_num = ensure_integer_frame(channel)

        transition = seq_editor.strips.new_effect(
            name=f"Wipe_{strip_being_wiped_away.name}_{strip_being_wiped_in.name}",
            type='WIPE',
            channel=channel_num,
            frame_start=trans_start,
            frame_end=trans_end,
            seq1=strip_being_wiped_away,
            seq2=strip_being_wiped_in
        )

        transition.transition_type = wipe_type
        transition.direction = 'IN' if angle == 0.0 else 'OUT' # This logic might need refinement based on Blender's exact wipe 'direction' behavior with 'angle'
        transition.angle = angle

        print(f"Successfully created wipe: {transition.name} (frames {trans_start}-{trans_end}) on channel {channel_num}, type: {wipe_type}")
        return transition

    except ValueError as ve:
        print(f"Error creating wipe due to invalid frame value: {ve}")
        traceback.print_exc()
        return None
    except Exception as e:
        print(f"Error creating wipe for '{strip_being_wiped_away.name}' and '{strip_being_wiped_in.name}': {e}")
        traceback.print_exc()
        return None

def create_audio_fade(sound_strip: bpy.types.SoundSequence, # More specific type
                      fade_type: str = 'IN',
                      duration_frames: int = 24) -> bool:
    """
    Create a volume fade for a sound strip by keyframing its volume.

    Args:
        sound_strip (bpy.types.SoundSequence): The sound strip to fade.
        fade_type (str): Type of fade ('IN' or 'OUT').
        duration_frames (int): Duration of the fade in frames.

    Returns:
        bool: True on success, False on failure.
    """
    try:
        if not isinstance(sound_strip, bpy.types.SoundSequence) or sound_strip.type != 'SOUND':
            print(f"Error: Invalid sound strip '{sound_strip.name if hasattr(sound_strip, 'name') else 'Unknown'}' provided for audio fade. Expected SoundSequence.")
            return False
        if fade_type not in ['IN', 'OUT']:
            print(f"Error: Invalid fade_type '{fade_type}'. Must be 'IN' or 'OUT'.")
            return False
        if duration_frames <= 0:
            print(f"Error: duration_frames ({duration_frames}) must be positive.")
            return False

        strip_start = ensure_integer_frame(sound_strip.frame_start)
        strip_final_end = ensure_integer_frame(sound_strip.frame_final_end) # frame_final_end is exclusive, so length is final_end - start
        strip_duration = ensure_integer_frame(sound_strip.frame_duration)


        if fade_type == 'IN':
            keyframe_start_frame = strip_start
            keyframe_end_frame = ensure_integer_frame(strip_start + duration_frames)
            # Ensure fade doesn't exceed strip duration
            if keyframe_end_frame > strip_start + strip_duration:
                keyframe_end_frame = ensure_integer_frame(strip_start + strip_duration)
                print(f"Warning: Audio fade-in duration for '{sound_strip.name}' was clamped to strip duration ({strip_duration} frames).")
            start_vol, end_vol = 0.0, 1.0
        else: # fade_type == 'OUT'
            keyframe_end_frame = strip_final_end # Keyframe at the very end of the strip content
            keyframe_start_frame = ensure_integer_frame(keyframe_end_frame - duration_frames)
             # Ensure fade doesn't start before strip beginning
            if keyframe_start_frame < strip_start:
                keyframe_start_frame = strip_start
                print(f"Warning: Audio fade-out duration for '{sound_strip.name}' was clamped to strip duration.")
            start_vol, end_vol = 1.0, 0.0
        
        # It's important to set the volume *before* inserting the keyframe for that frame
        sound_strip.volume = start_vol
        sound_strip.keyframe_insert(data_path="volume", frame=keyframe_start_frame)
        
        sound_strip.volume = end_vol
        sound_strip.keyframe_insert(data_path="volume", frame=keyframe_end_frame)

        print(f"Successfully created audio {fade_type} fade for '{sound_strip.name}' (frames {keyframe_start_frame}-{keyframe_end_frame}), vol {start_vol}->{end_vol}")
        return True

    except ValueError as ve:
        print(f"Error creating audio fade for '{sound_strip.name if hasattr(sound_strip, 'name') else 'Unknown'}' due to invalid frame value: {ve}")
        traceback.print_exc()
        return False
    except Exception as e:
        print(f"Error creating audio fade for '{sound_strip.name if hasattr(sound_strip, 'name') else 'Unknown'}': {e}")
        traceback.print_exc()
        return False

def create_audio_crossfade(sound1: bpy.types.SoundSequence,
                           sound2: bpy.types.SoundSequence,
                           overlap_frames: int = 24) -> bool:
    """
    Create a crossfade between two audio strips by keyframing their volumes.
    Caller must ensure strips are positioned for overlap.

    Args:
        sound1 (bpy.types.SoundSequence): The first sound strip (fading out).
        sound2 (bpy.types.SoundSequence): The second sound strip (fading in).
        overlap_frames (int): Duration of the overlap/crossfade in frames.

    Returns:
        bool: True on success, False on failure.
    """
    try:
        if not isinstance(sound1, bpy.types.SoundSequence) or sound1.type != 'SOUND':
            print(f"Error: Invalid sound strip sound1 ('{sound1.name if hasattr(sound1, 'name') else 'Unknown'}') for audio crossfade.")
            return False
        if not isinstance(sound2, bpy.types.SoundSequence) or sound2.type != 'SOUND':
            print(f"Error: Invalid sound strip sound2 ('{sound2.name if hasattr(sound2, 'name') else 'Unknown'}') for audio crossfade.")
            return False
        if overlap_frames <= 0:
            print(f"Error: overlap_frames ({overlap_frames}) must be positive.")
            return False

        s1_start = ensure_integer_frame(sound1.frame_start)
        s1_final_end = ensure_integer_frame(sound1.frame_final_end)
        s2_start = ensure_integer_frame(sound2.frame_start)
        s2_final_end = ensure_integer_frame(sound2.frame_final_end)

        # Crossfade occurs over the overlap period, typically starting at sound2.frame_start
        # and lasting for overlap_frames.
        fade_start_abs = s2_start # Start of fade corresponds to start of sound2 for this setup
        fade_end_abs = ensure_integer_frame(fade_start_abs + overlap_frames)

        # Check if sound1 actually overlaps with this period
        if s1_final_end < fade_end_abs or s1_start > fade_start_abs :
             print(f"Warning: Audio Crossfade - '{sound1.name}' (frames {s1_start}-{s1_final_end}) may not fully cover the "
                  f"crossfade period ({fade_start_abs}-{fade_end_abs}) relative to '{sound2.name}'.")
        # Check if sound2 starts early enough for sound1 to fade over it
        if s2_start > ensure_integer_frame(s1_final_end - overlap_frames):
            print(f"Error: Audio Crossfade ({sound1.name} to {sound2.name}) - Strips do not overlap sufficiently. "
                  f"'{sound2.name}' (starts {s2_start}) must start on or before frame "
                  f"{ensure_integer_frame(s1_final_end - overlap_frames)} "
                  f"for a {overlap_frames}-frame crossfade.")
            return False
        
        # Keyframe sound1 volume (fading out)
        sound1.volume = 1.0 # Assume it's at full volume before fade
        sound1.keyframe_insert(data_path="volume", frame=fade_start_abs)
        sound1.volume = 0.0
        sound1.keyframe_insert(data_path="volume", frame=fade_end_abs)
        
        # Keyframe sound2 volume (fading in)
        sound2.volume = 0.0 # Assume it's at zero volume before fade
        sound2.keyframe_insert(data_path="volume", frame=fade_start_abs)
        sound2.volume = 1.0
        sound2.keyframe_insert(data_path="volume", frame=fade_end_abs)

        print(f"Successfully created audio crossfade for '{sound1.name}' and '{sound2.name}' (frames {fade_start_abs}-{fade_end_abs})")
        return True

    except ValueError as ve:
        print(f"Error creating audio crossfade due to invalid frame value: {ve}")
        traceback.print_exc()
        return False
    except Exception as e:
        print(f"Error creating audio crossfade for '{sound1.name if hasattr(sound1, 'name') else 'Unknown'}' and '{sound2.name if hasattr(sound2, 'name') else 'Unknown'}': {e}")
        traceback.print_exc()
        return False

def create_fade_to_color(seq_editor: bpy.types.SequenceEditor,
                         strip: bpy.types.Sequence,
                         fade_duration: int,
                         fade_type: str = 'IN', # 'IN' = from color to strip, 'OUT' = from strip to color
                         color: tuple[float, float, float] = (0.0, 0.0, 0.0), # RGB
                         channel: Optional[int] = None) -> Optional[bpy.types.Sequence]:
    """
    Create a fade from/to a solid color for a strip using a color strip and a crossfade.

    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor.
        strip (bpy.types.Sequence): The strip to fade.
        fade_duration (int): Duration of the fade in frames.
        fade_type (str): Type of fade ('IN' or 'OUT').
        color (tuple[float, float, float]): RGB color values (0-1 for each component).
        channel (Optional[int]): Channel for the color strip. The crossfade will be
                                 placed on channel + 1. If None, uses strip.channel + 1
                                 for color strip.

    Returns:
        Optional[bpy.types.Sequence]: The created crossfade effect strip or None on failure.
    """
    try:
        if not isinstance(strip, bpy.types.Sequence):
             print(f"Error: Invalid strip object provided for fade_to_color. Expected bpy.types.Sequence.")
             return None
        if not isinstance(seq_editor, bpy.types.SequenceEditor):
            print("Error: Invalid sequence editor object for create_fade_to_color.")
            return None
        if fade_duration <= 0:
            print(f"Error: fade_duration ({fade_duration}) must be positive.")
            return None
        if fade_type not in ['IN', 'OUT']:
            print(f"Error: Invalid fade_type '{fade_type}'. Must be 'IN' or 'OUT'.")
            return False

        base_channel = ensure_integer_frame(channel if channel is not None else strip.channel + 1)
        crossfade_channel = ensure_integer_frame(base_channel + 1) # Cross effect on channel above color strip

        strip_start_abs = ensure_integer_frame(strip.frame_start)
        strip_final_end_abs = ensure_integer_frame(strip.frame_final_end)

        if fade_type == 'IN':
            # Color strip starts at strip_start_abs and lasts for fade_duration
            color_start_frame = strip_start_abs
            color_end_frame = ensure_integer_frame(strip_start_abs + fade_duration)
        else: # fade_type == 'OUT'
            # Color strip ends at strip_final_end_abs and starts fade_duration before it
            color_end_frame = strip_final_end_abs
            color_start_frame = ensure_integer_frame(strip_final_end_abs - fade_duration)
        
        if color_start_frame < 0: # Safety for very short strips or long fades
            print(f"Warning: Calculated color_start_frame ({color_start_frame}) is negative. Clamping to 0.")
            color_start_frame = 0
            if fade_type == 'OUT': # Adjust end if start was clamped
                 color_end_frame = ensure_integer_frame(fade_duration)
            elif fade_type == 'IN': # Adjust end if start was clamped (should not happen for IN if strip_start_abs >=0)
                 color_end_frame = ensure_integer_frame(fade_duration)


        # Create the color strip
        color_strip = seq_editor.strips.new_effect(
            name=f"ColorForFade_{strip.name}",
            type='COLOR',
            channel=base_channel,
            frame_start=color_start_frame,
            frame_end=color_end_frame
        )
        color_strip.color = color # color is tuple (r,g,b)

        # Create a crossfade between the color strip and the original strip
        if fade_type == 'IN':
            # Fading from color_strip to strip
            seq1, seq2 = color_strip, strip
            effect_name = f"FadeInFromColor_{strip.name}"
        else: # fade_type == 'OUT'
            # Fading from strip to color_strip
            seq1, seq2 = strip, color_strip
            effect_name = f"FadeOutToColor_{strip.name}"

        # The crossfade transition should have the same start/end as the color strip
        # as it defines the duration of the fade.
        transition_effect = seq_editor.strips.new_effect(
            name=effect_name,
            type='CROSS',
            channel=crossfade_channel,
            frame_start=color_start_frame, # Crossfade aligns with color strip
            frame_end=color_end_frame,     # Crossfade aligns with color strip
            seq1=seq1,
            seq2=seq2
        )
        
        print(f"Successfully created fade {fade_type} for '{strip.name}' using color {color} "
              f"(color strip: {color_strip.name} frames {color_start_frame}-{color_end_frame} on ch {base_channel}, "
              f"crossfade: {transition_effect.name} on ch {crossfade_channel})")
        return transition_effect

    except ValueError as ve:
        print(f"Error creating fade to color for '{strip.name}' due to invalid frame value: {ve}")
        traceback.print_exc()
        return None
    except Exception as e:
        print(f"Error creating fade to color for '{strip.name}': {e}")
        traceback.print_exc()
        return None