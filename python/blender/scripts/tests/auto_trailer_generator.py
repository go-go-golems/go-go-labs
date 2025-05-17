# context.area: SEQUENCE_EDITOR
# Don't remove the comment above. It's important.

"""
Automatic Trailer Generator for Blender VSE

This script takes a longer video and automatically creates a trailer by
sampling clips from throughout the video and assembling them with transitions.

Run this script from Blender's Text Editor with a video in the VSE or no video (it will prompt).
"""

import bpy # type: ignore
import os
import sys
import random
from math import ceil

# Add utilities path to import our modules
scripts_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
utils_dir = os.path.join(scripts_dir, 'utils')
if utils_dir not in sys.path:
    sys.path.append(utils_dir)

# Try to import utilities
try:
    from vse_utils import (
        get_active_scene, ensure_sequence_editor, add_video_strip, add_audio_strip,
        add_text_strip, add_crossfade, split_strip, clear_all_strips
    )
    from transition_utils import create_crossfade, create_audio_crossfade, create_fade_to_color
    print("Successfully imported VSE utilities")
except ImportError as e:
    print(f"Warning: Could not import utilities: {e}")
    print("Using built-in implementations")
    
    # --- Fallback implementations of required functions ---
    def get_active_scene():
        return bpy.context.scene
    
    def ensure_sequence_editor(scene=None):
        if scene is None:
            scene = get_active_scene()
        if not scene.sequence_editor:
            scene.sequence_editor_create()
        return scene.sequence_editor
    
    def add_video_strip(seq_editor, filepath, channel=1, frame_start=1, name=None):
        if name is None:
            name = os.path.splitext(os.path.basename(filepath))[0]
        return seq_editor.strips.new_movie(
            name=name, filepath=filepath, channel=channel, frame_start=frame_start)
    
    def add_audio_strip(seq_editor, filepath, channel=2, frame_start=1, name=None):
        if name is None:
            name = os.path.splitext(os.path.basename(filepath))[0]
        return seq_editor.strips.new_sound(
            name=name, filepath=filepath, channel=channel, frame_start=frame_start)
    
    def add_text_strip(seq_editor, text, frame_start, frame_end, channel=10, 
                      position='center', size=50, color=(1,1,1)):
        text_strip = seq_editor.strips.new_effect(
            name=f"Text_{text[:10]}", type='TEXT', channel=channel,
            frame_start=frame_start, frame_end=frame_end)
        text_strip.text = text
        text_strip.font_size = size
        text_strip.color = color
        if position == 'center':
            text_strip.location = (0.5, 0.5)
        elif position == 'top':
            text_strip.location = (0.5, 0.8)
        elif position == 'bottom':
            text_strip.location = (0.5, 0.2)
        return text_strip
    
    def create_crossfade(seq_editor, strip1, strip2, transition_duration, channel=None):
        if channel is None:
            channel = max(strip1.channel, strip2.channel) + 1
        return seq_editor.strips.new_effect(
            name=f"Cross_{strip1.name}_{strip2.name}", type='CROSS',
            channel=channel, frame_start=strip2.frame_start,
            frame_end=strip2.frame_start + transition_duration,
            seq1=strip1, seq2=strip2)
    
    def create_audio_crossfade(sound1, sound2, overlap_frames=24):
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
        if channel is None:
            channel = strip.channel + 1
        if fade_type == 'IN':
            color_start = strip.frame_start
            color_end = color_start + fade_duration
        else:  # fade_type == 'OUT'
            color_end = strip.frame_final_end
            color_start = color_end - fade_duration
        
        color_strip = seq_editor.strips.new_effect(
            name=f"FadeColor_{strip.name}", type='COLOR',
            channel=channel, frame_start=color_start, frame_end=color_end)
        color_strip.color = color
        
        if fade_type == 'IN':
            return seq_editor.strips.new_effect(
                name=f"FadeIn_{strip.name}", type='CROSS',
                channel=channel + 1, frame_start=color_start, frame_end=color_end,
                seq1=color_strip, seq2=strip)
        else:  # fade_type == 'OUT'
            return seq_editor.strips.new_effect(
                name=f"FadeOut_{strip.name}", type='CROSS',
                channel=channel + 1, frame_start=color_start, frame_end=color_end,
                seq1=strip, seq2=color_strip)
    
    def split_strip(strip, frame):
        frame = int(frame)
        if frame <= strip.frame_start or frame >= strip.frame_final_end:
            return (strip, None)
        
        seq_editor = bpy.context.scene.sequence_editor
        pre_split_ids = {s.as_pointer() for s in seq_editor.strips_all}
        
        for s in seq_editor.strips_all:
            s.select = (s == strip)
        
        bpy.context.scene.frame_current = frame
        try:
            bpy.ops.sequencer.split(frame=frame, channel=strip.channel, type='SOFT')
        except RuntimeError:
            return (strip, None)
        
        left_part = right_part = None
        for s in seq_editor.strips_all:
            if s.channel == strip.channel:
                if s.frame_final_end == frame:
                    left_part = s
                elif s.frame_start == frame:
                    right_part = s
        
        return (left_part, right_part)
    
    def clear_all_strips(seq_editor):
        strip_count = len(seq_editor.strips_all)
        if strip_count > 0:
            strips_to_remove = list(seq_editor.strips_all)
            for strip in strips_to_remove:
                seq_editor.strips.remove(strip)
        return strip_count

# --- Trailer Generator Functions ---

def extract_clip_segment(source_strip, start_offset, segment_duration, target_channel, target_frame):
    """
    Extract a segment from a source strip and place it at the target position.
    
    Args:
        source_strip (bpy.types.Strip): Source strip to extract from
        start_offset (int): Offset from source start to begin extraction
        segment_duration (int): Duration of segment to extract
        target_channel (int): Channel to place the extracted segment
        target_frame (int): Frame position to place the extracted segment
    
    Returns:
        tuple: (video_strip, audio_strip) for the extracted segment
    """
    seq_editor = ensure_sequence_editor()
    
    # Calculate the actual source frame to start from
    source_start = source_strip.frame_start + start_offset
    
    # Make sure we don't exceed the source clip length
    if source_start + segment_duration > source_strip.frame_final_end:
        segment_duration = source_strip.frame_final_end - source_start
        if segment_duration <= 0:
            print(f"Warning: Invalid segment - start offset too large: {start_offset}")
            return (None, None)
    
    # Create a new video strip from the same file
    video_strip = add_video_strip(
        seq_editor=seq_editor,
        filepath=source_strip.filepath,
        channel=target_channel,
        frame_start=target_frame,
        name=f"Trailer_Clip_{target_frame}"
    )
    
    # Add a matching audio strip
    audio_strip = add_audio_strip(
        seq_editor=seq_editor,
        filepath=source_strip.filepath,
        channel=target_channel + 1,
        frame_start=target_frame,
        name=f"Trailer_Audio_{target_frame}"
    )
    
    # Set the offsets to trim to the desired segment
    start_trim = start_offset
    end_trim = video_strip.frame_duration - (start_offset + segment_duration)
    
    if start_trim > 0:
        video_strip.frame_offset_start = start_trim
        audio_strip.frame_offset_start = start_trim
    
    if end_trim > 0:
        video_strip.frame_offset_end = end_trim
        audio_strip.frame_offset_end = end_trim
    
    print(f"Extracted segment from frame {source_start} for {segment_duration} frames")
    print(f"  Placed at frame {target_frame} on channel {target_channel}")
    
    return (video_strip, audio_strip)

def generate_trailer(source_strip, trailer_duration=600, transition_duration=15, num_clips=None):
    """
    Generate a trailer from a longer source video.
    
    Args:
        source_strip (bpy.types.Strip): Source video strip
        trailer_duration (int): Desired trailer duration in frames
        transition_duration (int): Duration of crossfades between clips
        num_clips (int, optional): Number of clips to use. If None, calculated automatically.
    
    Returns:
        list: Generated clip pairs [(video1, audio1), (video2, audio2), ...]
    """
    seq_editor = ensure_sequence_editor()
    
    # Clear existing strips if any
    clear_all_strips(seq_editor)
    
    # Add the source strip back for reference
    source_video = add_video_strip(
        seq_editor=seq_editor,
        filepath=source_strip.filepath,
        channel=5,  # Use a higher channel to keep it separate
        frame_start=1,
        name="Source_Video"
    )
    source_video.mute = True  # Mute it as we'll extract segments from it
    
    source_duration = source_video.frame_final_duration
    print(f"Source video duration: {source_duration} frames")
    
    # Calculate number of clips if not specified
    if num_clips is None:
        # A rough heuristic: 1 clip per 5 seconds of trailer (at 24fps)
        num_clips = max(3, ceil(trailer_duration / 120))  # At least 3 clips
    
    print(f"Generating trailer with {num_clips} clips and {transition_duration} frame transitions")
    
    # Calculate non-transition duration (actual content)
    content_duration = trailer_duration - (transition_duration * (num_clips - 1))
    
    # Calculate individual clip duration
    clip_duration = content_duration // num_clips
    
    print(f"Each clip will be approximately {clip_duration} frames")
    
    # Generate sample points throughout the source video
    # We'll skip the first and last 10% of the source for better content
    usable_start = int(source_duration * 0.1)
    usable_end = int(source_duration * 0.9)
    usable_duration = usable_end - usable_start
    
    # Create evenly spaced sample points with some randomness
    sample_interval = usable_duration // num_clips
    sample_points = []
    
    for i in range(num_clips):
        # Base point plus some randomness within the interval
        base = usable_start + (i * sample_interval)
        jitter = random.randint(-sample_interval//4, sample_interval//4)
        point = max(0, min(source_duration - clip_duration, base + jitter))
        sample_points.append(point)
    
    # Sort points to use in sequential order
    # This is optional - for a more jumbled trailer, could skip sorting
    sample_points.sort()
    
    print(f"Sample points selected: {sample_points}")
    
    # Extract segments and place them in sequence
    current_frame = 1
    clip_pairs = []
    
    for i, offset in enumerate(sample_points):
        # Extract a segment and place it at current_frame
        video, audio = extract_clip_segment(
            source_strip=source_video,
            start_offset=offset,
            segment_duration=clip_duration,
            target_channel=1,  # Video channel
            target_frame=current_frame
        )
        
        if video and audio:
            clip_pairs.append((video, audio))
            current_frame = video.frame_final_end - transition_duration
    
    # Add transitions between clips
    for i in range(len(clip_pairs) - 1):
        video1, audio1 = clip_pairs[i]
        video2, audio2 = clip_pairs[i + 1]
        
        # Create video crossfade
        create_crossfade(
            seq_editor=seq_editor,
            strip1=video1,
            strip2=video2,
            transition_duration=transition_duration,
            channel=3  # Above both video channels
        )
        
        # Create audio crossfade
        create_audio_crossfade(
            sound1=audio1,
            sound2=audio2,
            overlap_frames=transition_duration
        )
    
    # Add fade in/out to the first and last clips
    if clip_pairs:
        first_video, first_audio = clip_pairs[0]
        last_video, last_audio = clip_pairs[-1]
        
        # Create fade in from black for first clip
        create_fade_to_color(
            seq_editor=seq_editor,
            strip=first_video,
            fade_duration=24,  # 1 second at 24fps
            fade_type='IN',
            color=(0, 0, 0),  # Black
            channel=4
        )
        
        # Create fade out to black for last clip
        create_fade_to_color(
            seq_editor=seq_editor,
            strip=last_video,
            fade_duration=36,  # 1.5 seconds at 24fps
            fade_type='OUT',
            color=(0, 0, 0),  # Black
            channel=4
        )
    
    return clip_pairs

def add_trailer_titles(seq_editor, clip_pairs, source_title="Video Title"):
    """
    Add titles and text overlays to a trailer.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        clip_pairs (list): List of (video, audio) pairs from generate_trailer
        source_title (str): Title of the original video/movie
    
    Returns:
        list: The created text strips
    """
    if not clip_pairs:
        return []
    
    text_strips = []
    
    # Get first and last video clips to determine frame ranges
    first_video, _ = clip_pairs[0]
    last_video, _ = clip_pairs[-1]
    
    # Calculate the total trailer duration
    trailer_start = first_video.frame_start
    trailer_end = last_video.frame_final_end
    
    # Add main title
    main_title = add_text_strip(
        seq_editor=seq_editor,
        text=f"{source_title} - TRAILER",
        frame_start=trailer_start + 48,  # After initial fade-in
        frame_end=trailer_start + 120,   # Duration about 3 seconds
        channel=10,
        position='center',
        size=80,
        color=(1, 0.8, 0.2)  # Gold color
    )
    text_strips.append(main_title)
    
    # Add a "Coming Soon" at the end
    end_text = add_text_strip(
        seq_editor=seq_editor,
        text="COMING SOON",
        frame_start=trailer_end - 120,
        frame_end=trailer_end - 24,  # End before final fade-out
        channel=10,
        position='center',
        size=100,
        color=(1, 1, 1)  # White
    )
    text_strips.append(end_text)
    
    # You could add more text like credits, dates, etc.
    
    return text_strips

def main():
    """
    Main function to run the trailer generator.
    If a video strip is selected, uses that, otherwise attempts to load a sample.
    """
    print("Automatic Trailer Generator for Blender VSE")
    print("===========================================\n")
    
    scene = get_active_scene()
    seq_editor = ensure_sequence_editor(scene)
    
    print(f"Operating on scene: '{scene.name}' with Sequence Editor: {seq_editor}")
    
    # Try to find a selected video strip
    selected_strips = [strip for strip in seq_editor.strips_all if strip.select and strip.type == 'MOVIE']
    
    source_strip = None
    
    if selected_strips:
        # Use the first selected video strip
        source_strip = selected_strips[0]
        print(f"Using selected video strip: '{source_strip.name}'")
    else:
        # No video strip selected, try to find one already in the editor
        movie_strips = [strip for strip in seq_editor.strips_all if strip.type == 'MOVIE']
        
        if movie_strips:
            source_strip = movie_strips[0]
            print(f"Using existing video strip: '{source_strip.name}'")
        else:
            # No video strips found, try to load a sample
            print("No video strips found. Attempting to load a sample video...")
            
            # Path to test media - try to find it in common locations
            test_media_dir = "/home/manuel/Movies/blender-movie-editor"
            
            # Check if path exists, if not look for alternative locations
            if not os.path.exists(test_media_dir):
                # Try relative path from current script
                script_dir = os.path.dirname(os.path.abspath(__file__))
                alternative_paths = [
                    os.path.join(os.path.dirname(script_dir), "media"),
                    "/tmp/blender-test-media"
                ]
                
                for path in alternative_paths:
                    if os.path.exists(path):
                        test_media_dir = path
                        print(f"Using media path: {test_media_dir}")
                        break
            
            # Try to find a video file
            video_file = None
            if os.path.exists(test_media_dir):
                for filename in os.listdir(test_media_dir):
                    if filename.lower().endswith(('.mp4', '.avi', '.mov', '.mkv')):
                        video_file = os.path.join(test_media_dir, filename)
                        break
            
            if video_file:
                print(f"Found video file: {video_file}")
                source_strip = add_video_strip(
                    seq_editor=seq_editor,
                    filepath=video_file,
                    channel=1,
                    frame_start=1
                )
                print(f"Added video strip: '{source_strip.name}'")
            else:
                print("Error: No video files found. Please select a video strip or add one to the VSE.")
                return
    
    # Now generate a trailer from the source strip
    if source_strip:
        # Customize these parameters as needed
        trailer_duration = min(600, source_strip.frame_final_duration // 2)  # Max 25 seconds at 24fps or half the source
        transition_duration = 15  # 15 frames between clips
        num_clips = 5  # Number of clips to extract
        
        # Extract the title from the filename for use in titles
        source_title = os.path.splitext(os.path.basename(source_strip.filepath))[0].replace('_', ' ').title()
        
        # Generate the trailer
        clip_pairs = generate_trailer(
            source_strip=source_strip,
            trailer_duration=trailer_duration,
            transition_duration=transition_duration,
            num_clips=num_clips
        )
        
        # Add titles
        if clip_pairs:
            add_trailer_titles(
                seq_editor=seq_editor,
                clip_pairs=clip_pairs,
                source_title=source_title
            )
            
            print("\nTrailer generation complete!")
            print(f"Created trailer with {len(clip_pairs)} clips and duration of approximately {trailer_duration} frames")
            print("Review the VSE timeline to see the created trailer.")
        else:
            print("Error: Failed to generate trailer clips.")
    else:
        print("Error: No source video available to generate trailer from.")

# Run the main function when this script is executed
if __name__ == "__main__":
    main()