# VSE Utilities for Blender's Video Sequence Editor
# Core utilities that can be imported by any VSE script

import bpy # type: ignore
import os

# Common shortcuts for Blender data and context
D = bpy.data
C = bpy.context

def get_active_scene():
    """Get the currently active scene."""
    return C.scene

def ensure_sequence_editor(scene=None):
    """Ensure a scene has a sequence editor, creating one if it doesn't exist.
    
    Args:
        scene (bpy.types.Scene, optional): The scene to check/modify. 
                                         Defaults to active scene.
    
    Returns:
        bpy.types.SequenceEditor: The sequence editor
    """
    if scene is None:
        scene = get_active_scene()
    
    if not scene.sequence_editor:
        scene.sequence_editor_create()
    
    return scene.sequence_editor

def add_video_strip(seq_editor, filepath, channel=1, frame_start=1, name=None):
    """Add a video strip to the sequence editor.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        filepath (str): Path to the video file
        channel (int, optional): Channel to place the strip. Defaults to 1.
        frame_start (int, optional): Frame to start the strip. Defaults to 1.
        name (str, optional): Name for the strip. Defaults to filename without extension.
    
    Returns:
        bpy.types.Strip: The created movie strip
    """
    if name is None:
        name = os.path.splitext(os.path.basename(filepath))[0]
    
    movie_strip = seq_editor.strips.new_movie(
        name=name,
        filepath=filepath,
        channel=channel,
        frame_start=frame_start
    )
    
    return movie_strip

def add_audio_strip(seq_editor, filepath, channel=2, frame_start=1, name=None):
    """Add an audio strip to the sequence editor.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        filepath (str): Path to the audio file
        channel (int, optional): Channel to place the strip. Defaults to 2.
        frame_start (int, optional): Frame to start the strip. Defaults to 1.
        name (str, optional): Name for the strip. Defaults to filename without extension.
    
    Returns:
        bpy.types.Strip: The created sound strip
    """
    if name is None:
        name = os.path.splitext(os.path.basename(filepath))[0]
    
    sound_strip = seq_editor.strips.new_sound(
        name=name,
        filepath=filepath,
        channel=channel,
        frame_start=frame_start
    )
    
    return sound_strip

def add_image_strip(seq_editor, filepath, channel=1, frame_start=1, duration=25, name=None):
    """Add an image strip to the sequence editor.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        filepath (str): Path to the image file
        channel (int, optional): Channel to place the strip. Defaults to 1.
        frame_start (int, optional): Frame to start the strip. Defaults to 1.
        duration (int, optional): Duration of the strip in frames. Defaults to 25.
        name (str, optional): Name for the strip. Defaults to filename without extension.
    
    Returns:
        bpy.types.Strip: The created image strip
    """
    if name is None:
        name = os.path.splitext(os.path.basename(filepath))[0]
    
    image_strip = seq_editor.strips.new_image(
        name=name,
        filepath=filepath,
        channel=channel,
        frame_start=frame_start
    )
    
    # Set the strip duration (default is often too short)
    image_strip.frame_final_duration = duration
    
    return image_strip

def add_color_strip(seq_editor, frame_start, frame_end, channel=1, name=None, color=(0,0,0)):
    """Add a solid color strip to the sequence editor.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        frame_start (int): Frame to start the strip
        frame_end (int): Frame to end the strip
        channel (int, optional): Channel to place the strip. Defaults to 1.
        name (str, optional): Name for the strip. Defaults to "Color".
        color (tuple, optional): RGB color values (0-1). Defaults to black (0,0,0).
    
    Returns:
        bpy.types.Strip: The created color strip
    """
    if name is None:
        name = "Color"
    
    color_strip = seq_editor.strips.new_effect(
        name=name,
        type='COLOR',
        channel=channel,
        frame_start=frame_start,
        frame_end=frame_end
    )
    
    # Set the color
    color_strip.color = color
    
    return color_strip

def add_text_strip(seq_editor, text, frame_start, frame_end, channel=10, position='center', 
                 size=50, color=(1,1,1)):
    """Add a text strip to the sequence editor.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        text (str): The text content to display
        frame_start (int): Frame to start the strip
        frame_end (int): Frame to end the strip
        channel (int, optional): Channel to place the strip. Defaults to 10.
        position (str, optional): Position of text: 'center', 'top', 'bottom'. Defaults to 'center'.
        size (int, optional): Font size. Defaults to 50.
        color (tuple, optional): RGB color values (0-1). Defaults to white (1,1,1).
    
    Returns:
        bpy.types.Strip: The created text strip
    """
    # Create the text strip
    text_strip = seq_editor.strips.new_effect(
        name=f"Text_{text[:10]}",
        type='TEXT',
        channel=channel,
        frame_start=frame_start,
        frame_end=frame_end
    )
    
    # Set the text strip properties
    text_strip.text = text
    text_strip.font_size = size
    text_strip.color = color
    
    # Position the text
    if position == 'center':
        text_strip.location = (0.5, 0.5)
    elif position == 'top':
        text_strip.location = (0.5, 0.8)
    elif position == 'bottom':
        text_strip.location = (0.5, 0.2)
    
    # Set alignment if the property exists
    if hasattr(text_strip, 'align_x'):
        text_strip.align_x = 'CENTER'
    if hasattr(text_strip, 'align_y'):
        if position == 'bottom':
            text_strip.align_y = 'TOP'
        elif position == 'top':
            text_strip.align_y = 'BOTTOM'
        else:
            text_strip.align_y = 'CENTER'
    
    return text_strip

def trim_strip_start(strip, frames_to_trim):
    """Trim frames from the beginning of a strip.
    
    Args:
        strip (bpy.types.Strip): The strip to trim
        frames_to_trim (int): Number of frames to trim from beginning
    
    Returns:
        tuple: (original_start_frame, new_start_frame) for verification
    """
    original_start = strip.frame_final_start
    
    # Method 1: Using frame_offset_start
    strip.frame_offset_start += frames_to_trim
    
    return (original_start, strip.frame_final_start)

def trim_strip_end(strip, frames_to_trim):
    """Trim frames from the end of a strip.
    
    Args:
        strip (bpy.types.Strip): The strip to trim
        frames_to_trim (int): Number of frames to trim from end
    
    Returns:
        tuple: (original_end_frame, new_end_frame) for verification
    """
    original_end = strip.frame_final_end
    
    # Method 1: Using frame_offset_end
    strip.frame_offset_end += frames_to_trim
    
    return (original_end, strip.frame_final_end)

def split_strip(strip, frame):
    """Split a strip at the specified frame.
    
    Args:
        strip (bpy.types.Strip): The strip to split
        frame (int): Frame at which to make the cut
    
    Returns:
        tuple: (left_part, right_part) - the two resulting strips
    """
    # Convert to integer frame number
    frame = int(frame)
    
    # Check if frame is within strip bounds
    if frame <= strip.frame_start or frame >= strip.frame_final_end:
        print(f"Warning: Split frame {frame} is outside '{strip.name}' bounds")
        return (strip, None)
    
    seq_editor = C.scene.sequence_editor
    
    # Record existing strips before the split
    pre_split_ids = {s.as_pointer() for s in seq_editor.strips_all}
    
    # Select only the strip we want to split
    for s in seq_editor.strips_all:
        s.select = (s == strip)
    
    # Set current frame and perform the split
    C.scene.frame_current = frame
    try:
        bpy.ops.sequencer.split(frame=frame, channel=strip.channel, type='SOFT')
    except RuntimeError as e:
        print(f"Error: Split failed on '{strip.name}': {e}")
        return (strip, None)
    
    # Find the left and right parts
    left_part = right_part = None
    for s in seq_editor.strips_all:
        if s.channel == strip.channel:
            if s.frame_final_end == frame:
                left_part = s
            elif s.frame_start == frame:
                right_part = s
    
    return (left_part, right_part)

def clear_all_strips(seq_editor):
    """Remove all strips from the sequence editor.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor to clear
    
    Returns:
        int: Number of strips removed
    """
    strip_count = len(seq_editor.strips_all)
    
    if strip_count > 0:
        print(f"\n=== Clearing {strip_count} strips ===")
        
        # First, print all strips and their dependencies
        print("\nStrip dependencies before removal:")
        for strip in seq_editor.strips_all:
            deps = []
            if hasattr(strip, 'seq1') and strip.seq1:
                deps.append(f"seq1={strip.seq1.name}")
            if hasattr(strip, 'seq2') and strip.seq2:
                deps.append(f"seq2={strip.seq2.name}")
            dep_str = f" (Dependencies: {', '.join(deps)})" if deps else ""
            print(f"  {strip.name} (Type: {strip.type}){dep_str}")
        
        # Select all strips
        for strip in seq_editor.sequences_all:
            strip.select = True
        
        # Use Blender's operator to delete them (this handles dependencies correctly)
        print("\nDeleting all strips using Blender operator")
        bpy.ops.sequencer.delete()
    
    remaining = len(seq_editor.strips_all)
    print(f"\nStrips removed: {strip_count}, Remaining: {remaining}")
    return strip_count

def find_strips_at_frame(seq_editor, frame, channel=None, strip_type=None):
    """Find all strips at a specific frame, optionally filtering by channel or type.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor to search
        frame (int): The frame to check
        channel (int, optional): Channel to filter by. If None, check all channels.
        strip_type (str, optional): Strip type to filter by ('MOVIE', 'SOUND', etc.). 
                                   If None, include all types.
    
    Returns:
        list: Strips at the specified frame matching the criteria
    """
    matching_strips = []
    
    for strip in seq_editor.strips_all:
        # Check if strip contains the frame
        if strip.frame_final_start <= frame < strip.frame_final_end:
            # Apply channel filter if specified
            if channel is not None and strip.channel != channel:
                continue
            
            # Apply type filter if specified
            if strip_type is not None and strip.type != strip_type:
                continue
            
            matching_strips.append(strip)
    
    return matching_strips

def set_scene_fps(scene, fps):
    """Set the scene frame rate.
    
    Args:
        scene (bpy.types.Scene): The scene to modify
        fps (float): Frames per second to set
    
    Returns:
        float: The previous FPS setting
    """
    current_fps = scene.render.fps / scene.render.fps_base
    
    # Set the new FPS
    scene.render.fps = int(fps)
    scene.render.fps_base = 1.0
    
    return current_fps

def add_crossfade(seq_editor, strip1, strip2, duration_frames, channel=None):
    """Add a crossfade transition between two strips.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        strip1 (bpy.types.Strip): First strip (fading out)
        strip2 (bpy.types.Strip): Second strip (fading in)
        duration_frames (int): Duration of crossfade in frames
        channel (int, optional): Channel for the transition. If None, uses max(strip1.channel, strip2.channel) + 1
    
    Returns:
        bpy.types.Strip: The created crossfade strip
    """
    # Ensure the strips overlap
    if strip2.frame_start > strip1.frame_final_end - duration_frames:
        # Adjust strip2 position to create required overlap
        strip2.frame_start = strip1.frame_final_end - duration_frames
    
    # Set transition channel if not specified
    if channel is None:
        channel = max(strip1.channel, strip2.channel) + 1
    
    # Create the crossfade
    crossfade = seq_editor.strips.new_effect(
        name=f"Cross_{strip1.name}_{strip2.name}",
        type='CROSS',
        channel=channel,
        frame_start=strip2.frame_start,
        frame_end=strip2.frame_start + duration_frames,
        seq1=strip1,
        seq2=strip2
    )
    
    return crossfade

def add_transform_effect(seq_editor, strip, offset_x=0, offset_y=0, scale_x=1.0, scale_y=1.0, 
                       rotation=0.0, channel=None):
    """Add a transform effect to a strip.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        strip (bpy.types.Strip): The strip to transform
        offset_x (float, optional): Horizontal offset in pixels. Defaults to 0.
        offset_y (float, optional): Vertical offset in pixels. Defaults to 0.
        scale_x (float, optional): Horizontal scale factor. Defaults to 1.0.
        scale_y (float, optional): Vertical scale factor. Defaults to 1.0.
        rotation (float, optional): Rotation in degrees. Defaults to 0.0.
        channel (int, optional): Channel for the effect. If None, uses strip.channel + 1.
    
    Returns:
        bpy.types.Strip: The created transform effect strip
    """
    from math import radians
    
    if channel is None:
        channel = strip.channel + 1
    
    # Create the transform effect
    transform = seq_editor.strips.new_effect(
        name=f"Transform_{strip.name}",
        type='TRANSFORM',
        channel=channel,
        frame_start=strip.frame_start,
        frame_end=strip.frame_final_end,
        seq1=strip
    )
    
    # Set the transform properties
    transform.transform.offset_x = offset_x
    transform.transform.offset_y = offset_y
    transform.transform.scale_x = scale_x
    transform.transform.scale_y = scale_y
    transform.transform.rotation = radians(rotation)  # Convert degrees to radians
    
    return transform

def print_sequence_info(seq_editor, title="Current Sequence State"):
    """Print information about all strips in the sequence editor.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor to analyze
        title (str, optional): Title for the output. Defaults to "Current Sequence State".
    """
    print(f"\n----- {title} -----")
    print(f"Total strips: {len(seq_editor.strips_all)}")
    
    # Group strips by channel
    strips_by_channel = {}
    for strip in sorted(seq_editor.strips_all, key=lambda s: (s.channel, s.frame_start)):
        if strip.channel not in strips_by_channel:
            strips_by_channel[strip.channel] = []
        strips_by_channel[strip.channel].append(strip)
    
    # Print strips by channel
    for channel in sorted(strips_by_channel.keys()):
        print(f"\nChannel {channel}:")
        for strip in strips_by_channel[channel]:
            print(f"  {strip.name} (Type: {strip.type})")
            print(f"    Frames: {strip.frame_start}-{strip.frame_final_end}, Duration: {strip.frame_final_duration}")
            
            # Print type-specific info
            if strip.type == 'MOVIE':
                print(f"    File: {strip.filepath}")
                print(f"    Offsets: start={strip.frame_offset_start}, end={strip.frame_offset_end}")
            elif strip.type == 'SOUND':
                print(f"    Volume: {strip.volume}")
                # Check if strip has animation data for volume
                has_volume_keyframes = False
                try:
                    # Check the scene's animation data since strip keyframes are stored there
                    if bpy.context.scene.animation_data and bpy.context.scene.animation_data.action:
                        for fc in bpy.context.scene.animation_data.action.fcurves:
                            if fc.data_path.startswith('sequence_editor.sequences_all[') and fc.data_path.endswith('].volume'):
                                # This is a volume fcurve for some strip - check if it's for this strip
                                if strip.name in fc.data_path:
                                    has_volume_keyframes = True
                                    break
                    if has_volume_keyframes:
                        print(f"    Has volume keyframes: Yes")
                except Exception as e:
                    print(f"    Error checking volume keyframes: {e}")
            elif strip.type in {'CROSS', 'GAMMA_CROSS', 'WIPE'}:
                if hasattr(strip, 'seq1') and strip.seq1:
                    print(f"    Input 1: {strip.seq1.name}")
                if hasattr(strip, 'seq2') and strip.seq2:
                    print(f"    Input 2: {strip.seq2.name}")
                if strip.type == 'WIPE' and hasattr(strip, 'transition_type'):
                    print(f"    Wipe Type: {strip.transition_type}")
            elif strip.type == 'COLOR':
                print(f"    Color: RGB{strip.color}")

def find_test_media_dir(default_path="/home/manuel/Movies/blender-movie-editor"):
    """Find a directory with test media files.
    
    Args:
        default_path (str, optional): Default path to look for test media. 
                                      Defaults to "/home/manuel/Movies/blender-movie-editor".
    
    Returns:
        str: Path to a directory with test media files
    """
    # First check if the default path exists
    if os.path.exists(default_path):
        return default_path
    
    # Try alternative locations
    script_dir = os.path.dirname(os.path.abspath(__file__))
    project_dir = os.path.dirname(os.path.dirname(script_dir))  # Go up two levels
    
    alternative_paths = [
        os.path.join(project_dir, "media"),
        os.path.join(script_dir, "media"),
        "/tmp/blender-test-media"
    ]
    
    for path in alternative_paths:
        if os.path.exists(path):
            return path
    
    print(f"Warning: Could not find test media directory. Using default: {default_path}")
    return default_path

# Alias functions to match the book's functions
add_movie_strip = add_video_strip
add_sound_strip = add_audio_strip