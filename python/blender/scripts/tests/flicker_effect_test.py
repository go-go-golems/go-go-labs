# context.area: SEQUENCE_EDITOR
# Don't remove the comment above. It's important.

"""
Flicker Effect Demo for Blender VSE

This script creates a vintage film flicker effect by randomly
altering the brightness of frames in a video clip.

Run this script from Blender's Text Editor with a video clip selected in the VSE.
"""

import bpy # type: ignore
import os
import sys
import random

# Add utilities path to import our modules
scripts_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
utils_dir = os.path.join(scripts_dir, 'utils')
if utils_dir not in sys.path:
    sys.path.append(utils_dir)

# Try to import VSE utilities
try:
    from vse_utils import (
        get_active_scene, ensure_sequence_editor, add_video_strip, 
        add_transform_effect, find_strips_at_frame
    )
    print("Successfully imported VSE utilities")
except ImportError as e:
    print(f"Warning: Could not import vse_utils: {e}")
    print("Using built-in implementations")
    
    # --- Fallback implementations ---
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
    
    def add_transform_effect(seq_editor, strip, offset_x=0, offset_y=0, scale_x=1.0, scale_y=1.0, 
                            rotation=0.0, channel=None):
        if channel is None:
            channel = strip.channel + 1
        transform = seq_editor.strips.new_effect(
            name=f"Transform_{strip.name}", type='TRANSFORM', channel=channel,
            frame_start=strip.frame_start, frame_end=strip.frame_final_end, seq1=strip)
        transform.transform.offset_x = offset_x
        transform.transform.offset_y = offset_y
        transform.transform.scale_x = scale_x
        transform.transform.scale_y = scale_y
        transform.transform.rotation = rotation
        return transform
    
    def find_strips_at_frame(seq_editor, frame, channel=None, strip_type=None):
        matching_strips = []
        for strip in seq_editor.strips_all:
            if strip.frame_final_start <= frame < strip.frame_final_end:
                if channel is not None and strip.channel != channel:
                    continue
                if strip_type is not None and strip.type != strip_type:
                    continue
                matching_strips.append(strip)
        return matching_strips

# --- Flicker Effect Functions ---

def add_flicker_effect(strip, intensity=0.2, frequency=0.25, channel=None):
    """
    Add a vintage film flicker effect to a video strip by creating multiple
    transform strips with random brightness adjustments.
    
    Args:
        strip (bpy.types.Strip): The video strip to add flicker to
        intensity (float): Strength of the flicker effect (0.0-1.0)
        frequency (float): How often the flicker occurs (0.0-1.0)
        channel (int, optional): Channel for the effect strips
    
    Returns:
        list: The transform effect strips created
    """
    if channel is None:
        channel = strip.channel + 1
    
    # Get sequence editor
    seq_editor = ensure_sequence_editor()
    
    # Determine frame range
    start_frame = strip.frame_start
    end_frame = strip.frame_final_end
    duration = end_frame - start_frame
    
    # Create transform effects at random frames
    transform_strips = []
    
    # Calculate how many flicker effects to add
    num_flickers = int(duration * frequency)
    
    print(f"Adding {num_flickers} flicker effects to '{strip.name}'")
    print(f"  Intensity: {intensity}, Frequency: {frequency}")
    print(f"  Duration: {duration} frames ({start_frame}-{end_frame})")
    
    # Create a list of random frames to apply flicker
    flicker_frames = random.sample(range(start_frame, end_frame), num_flickers)
    flicker_frames.sort()  # Sort in ascending order
    
    # Add transform effects for each flicker point
    for i, frame in enumerate(flicker_frames):
        # Decide how long this flicker lasts (1-3 frames typically)
        flicker_duration = random.randint(1, 3)
        
        # Calculate random brightness adjustment
        # Sometimes brighter, sometimes darker
        brightness = 1.0 + random.uniform(-intensity, intensity)
        
        # Create transform effect for this flicker segment
        transform = seq_editor.strips.new_effect(
            name=f"Flicker{i+1}",
            type='TRANSFORM',
            channel=channel,
            frame_start=frame,
            frame_end=min(frame + flicker_duration, end_frame),
            seq1=strip
        )
        
        # Set transform properties - only change brightness via scale
        transform.transform.scale_x = brightness
        transform.transform.scale_y = brightness
        
        transform_strips.append(transform)
    
    return transform_strips

def create_film_grain_overlay(seq_editor, video_strip, intensity=0.1, channel=None):
    """
    Create a film grain effect overlay.
    This is a placeholder - in a real implementation, you'd need a grain texture.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        video_strip (bpy.types.Strip): The video strip to match
        intensity (float): Grain intensity
        channel (int, optional): Channel for the grain overlay
    
    Returns:
        bpy.types.Strip: The film grain strip (typically would be an adjustment layer)
    """
    if channel is None:
        channel = video_strip.channel + 2
    
    # For a real implementation, you would add a noise texture here
    # This is just a placeholder - a slightly transparent adjustment layer
    print("Note: Film grain effect requires a noise texture. This is a placeholder.")
    
    # Create a gray color strip
    grain_strip = seq_editor.strips.new_effect(
        name="FilmGrain",
        type='COLOR',
        channel=channel,
        frame_start=video_strip.frame_start,
        frame_end=video_strip.frame_final_end
    )
    
    # Set to a dark gray
    grain_strip.color = (0.2, 0.2, 0.2)
    
    # In a real implementation, you would add a noise texture
    # and adjust its blend mode, opacity, etc.
    grain_strip.blend_type = 'OVERLAY'
    grain_strip.blend_alpha = intensity
    
    return grain_strip

def add_frame_jump_effect(video_strip, jump_frequency=0.05, channel=None):
    """
    Simulate film frame jumps by creating small vertical offsets at random frames.
    
    Args:
        video_strip (bpy.types.Strip): The video strip to add jumps to
        jump_frequency (float): How often jumps occur (0.0-1.0)
        channel (int, optional): Channel for the transform effects
    
    Returns:
        list: The transform effect strips created for jumps
    """
    if channel is None:
        channel = video_strip.channel + 3
    
    # Get sequence editor
    seq_editor = ensure_sequence_editor()
    
    # Determine frame range
    start_frame = video_strip.frame_start
    end_frame = video_strip.frame_final_end
    duration = end_frame - start_frame
    
    # Calculate how many jump effects to add
    num_jumps = int(duration * jump_frequency)
    
    print(f"Adding {num_jumps} frame jump effects to '{video_strip.name}'")
    print(f"  Jump Frequency: {jump_frequency}")
    
    # Create a list of random frames for jumps
    jump_frames = random.sample(range(start_frame, end_frame), num_jumps)
    jump_frames.sort()  # Sort in ascending order
    
    jump_strips = []
    
    # Add transform effects for each jump point
    for i, frame in enumerate(jump_frames):
        # Jumps are typically very short (1-2 frames)
        jump_duration = 1
        
        # Calculate random vertical offset (small jumps)
        offset_y = random.choice([-20, -15, -10, 10, 15, 20])
        
        # Create transform effect for this jump
        transform = seq_editor.strips.new_effect(
            name=f"Jump{i+1}",
            type='TRANSFORM',
            channel=channel,
            frame_start=frame,
            frame_end=frame + jump_duration,
            seq1=video_strip
        )
        
        # Set transform to only offset vertically
        transform.transform.offset_y = offset_y
        
        jump_strips.append(transform)
    
    return jump_strips

def create_vintage_film_effect(video_strip, flicker_intensity=0.15, jump_frequency=0.03, 
                              grain_intensity=0.08):
    """
    Apply a complete vintage film effect to a video strip.
    
    Args:
        video_strip (bpy.types.Strip): The video strip to apply effects to
        flicker_intensity (float): Intensity of light flicker (0.0-1.0)
        jump_frequency (float): Frequency of frame jumps (0.0-1.0) 
        grain_intensity (float): Intensity of film grain effect (0.0-1.0)
    
    Returns:
        tuple: (flicker_strips, jump_strips, grain_strip)
    """
    seq_editor = ensure_sequence_editor()
    
    print(f"Creating vintage film effect for '{video_strip.name}'...")
    
    # Add flickering effect (brightness variations)
    flicker_strips = add_flicker_effect(
        strip=video_strip, 
        intensity=flicker_intensity, 
        frequency=0.2, 
        channel=video_strip.channel + 1
    )
    
    # Add frame jumps (vertical position shifts)
    jump_strips = add_frame_jump_effect(
        video_strip=video_strip, 
        jump_frequency=jump_frequency, 
        channel=video_strip.channel + 2
    )
    
    # Add film grain overlay
    grain_strip = create_film_grain_overlay(
        seq_editor=seq_editor, 
        video_strip=video_strip, 
        intensity=grain_intensity, 
        channel=video_strip.channel + 3
    )
    
    print(f"Vintage film effect applied with {len(flicker_strips)} flicker points and {len(jump_strips)} jumps")
    
    return (flicker_strips, jump_strips, grain_strip)

def main():
    """
    Main function that sets up and applies flicker effects to a video strip.
    If a strip is selected, uses that; otherwise loads a sample video.
    """
    print("Flicker Effect Demo for Blender VSE")
    print("=================================\n")
    
    scene = get_active_scene()
    seq_editor = ensure_sequence_editor(scene)
    
    print(f"Operating on scene: '{scene.name}' with Sequence Editor: {seq_editor}")
    
    # Try to find a selected video strip
    selected_strips = [strip for strip in seq_editor.strips_all if strip.select and strip.type == 'MOVIE']
    
    video_strip = None
    
    if selected_strips:
        # Use the first selected video strip
        video_strip = selected_strips[0]
        print(f"Using selected video strip: '{video_strip.name}'")
    else:
        # No video strip selected, try to find one already in the editor
        movie_strips = [strip for strip in seq_editor.strips_all if strip.type == 'MOVIE']
        
        if movie_strips:
            video_strip = movie_strips[0]
            print(f"Using existing video strip: '{video_strip.name}'")
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
                video_strip = add_video_strip(
                    seq_editor=seq_editor,
                    filepath=video_file,
                    channel=1,
                    frame_start=1
                )
                print(f"Added video strip: '{video_strip.name}'")
            else:
                print("Error: No video files found. Please select a video strip or add one to the VSE.")
                return
    
    # Now apply the vintage film effect to the video strip
    if video_strip:
        create_vintage_film_effect(
            video_strip=video_strip,
            flicker_intensity=0.15,  # Light flicker intensity
            jump_frequency=0.03,     # Frame jumps frequency 
            grain_intensity=0.08     # Film grain overlay intensity
        )
        
        print("\nVintage film effect has been applied!")
        print("Review the VSE timeline to see the created effects.")
    else:
        print("Error: No video strip available to apply effects to.")

# Run the main function when this script is executed
if __name__ == "__main__":
    main()