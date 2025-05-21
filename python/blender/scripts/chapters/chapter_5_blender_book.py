# context.area: SEQUENCE_EDITOR
# Don't remove the comment above. It's important.

"""
Blender VSE Python API Test Script - Chapter 5: Effects and Adjustments

This script demonstrates core concepts from Chapter 5 of the Blender VSE Python API guide,
focusing on programmatically applying effects and adjustments to clips in the VSE.

Key concepts covered:
1.  **Transform Effects**: Adjusting position, scale, and rotation of clips
2.  **Color Effects**: Applying color adjustments using modifiers
3.  **Speed Control**: Changing playback speed of clips
4.  **Text Overlays**: Adding text titles and captions
5.  **Picture-in-Picture**: Creating PIP compositions
6.  **Glow Effects**: Applying glow/blur effects to clips

Usage:
    -   Run this script from Blender's Text Editor or Python Console.
    -   Ensure you are in the Video Editing workspace.
    -   The script creates a basic sequence and then demonstrates various effects.
"""

import bpy # type: ignore
import os
import sys
from math import radians
import traceback

# Common shortcuts for Blender data and context
D = bpy.data
C = bpy.context

# Set up enhanced logging
VERBOSE_LOGGING = True

def log(message, level='INFO'):
    """Enhanced logging function with levels and formatting"""
    if not VERBOSE_LOGGING and level != 'ERROR':
        return
    
    prefix = {
        'INFO': '🔵 [INFO]',
        'WARN': '🟠 [WARN]',
        'ERROR': '🔴 [ERROR]',
        'DEBUG': '🟢 [DEBUG]',
        'SUCCESS': '✅ [SUCCESS]',
        'IMPORT': '📦 [IMPORT]',
        'FUNCTION': '🔧 [FUNCTION]',
        'EFFECT': '✨ [EFFECT]',
    }.get(level, '[LOG]')
    
    print(f"{prefix} {message}")

# Try to import utilities - first make sure our utilities path is in sys.path
log("Setting up import paths", "DEBUG")
scripts_dir = os.path.dirname(os.path.abspath(__file__))
utils_dir = os.path.join(os.path.dirname(scripts_dir), 'utils')
log(f"Scripts directory: {scripts_dir}", "DEBUG")
log(f"Utils directory: {utils_dir}", "DEBUG")

if utils_dir not in sys.path:
    log(f"Adding {utils_dir} to sys.path", "DEBUG")
    sys.path.append(utils_dir)

# Print out all paths in sys.path for debugging
log("Current sys.path:", "DEBUG")
for i, path in enumerate(sys.path):
    log(f"  {i}: {path}", "DEBUG")

# Import from vse_utils first (these are standard utilities)
try:
    log("Attempting to import vse_utils", "IMPORT")
    from vse_utils import (
        get_active_scene, ensure_sequence_editor, add_movie_strip, add_sound_strip,
        add_text_strip as create_text_overlay, add_transform_effect as apply_transform_effect,
        find_test_media_dir
    )
    log("Successfully imported VSE utilities", "SUCCESS")
except ImportError as e:
    log(f"VSE utilities import failed: {e}", "ERROR")
    log(f"Traceback: {traceback.format_exc()}", "DEBUG")

# Now try to import the effect utilities
try:
    log("Attempting to import effect_utils", "IMPORT")
    from effect_utils import (
        create_picture_in_picture, apply_speed_effect,
        apply_color_balance, apply_glow_effect
    )
    log("Successfully imported effect utilities", "SUCCESS")
except ImportError as e:
    log(f"Effect utilities import failed: {e}", "ERROR")
    log(f"Traceback: {traceback.format_exc()}", "DEBUG")
    log("Using built-in fallback implementations", "WARN")
    
    # --- Fallback utility implementations ---
    # apply_transform_effect is now imported from vse_utils
    
    def create_picture_in_picture(seq_editor, main_strip, pip_strip, pip_scale=0.3, position='top-right', channel=None):
        """Create a picture-in-picture effect"""
        if channel is None:
            channel = max(main_strip.channel, pip_strip.channel) + 1
        
        # Calculate position based on specified corner
        width = C.scene.render.resolution_x
        height = C.scene.render.resolution_y
        padding = 20
        
        if position == 'top-right':
            offset_x = width/2 - (width * pip_scale)/2 - padding
            offset_y = height/2 - (height * pip_scale)/2 - padding
        elif position == 'top-left':
            offset_x = -(width/2 - (width * pip_scale)/2 - padding)
            offset_y = height/2 - (height * pip_scale)/2 - padding
        elif position == 'bottom-right':
            offset_x = width/2 - (width * pip_scale)/2 - padding
            offset_y = -(height/2 - (height * pip_scale)/2 - padding)
        elif position == 'bottom-left':
            offset_x = -(width/2 - (width * pip_scale)/2 - padding)
            offset_y = -(height/2 - (height * pip_scale)/2 - padding)
        else:  # 'center'
            offset_x = 0
            offset_y = 0
        
        return apply_transform_effect(
            seq_editor=seq_editor,
            strip=pip_strip,
            offset_x=offset_x,
            offset_y=offset_y,
            scale_x=pip_scale,
            scale_y=pip_scale,
            channel=channel
        )
    
    def apply_speed_effect(seq_editor, strip, speed_factor, channel=None):
        """Apply a speed effect to a strip"""
        if channel is None:
            channel = strip.channel + 1
        
        # Calculate new duration based on speed factor
        original_duration = strip.frame_final_duration
        new_duration = int(original_duration / speed_factor)
        
        speed_effect = seq_editor.strips.new_effect(
            name=f"Speed_{strip.name}_{speed_factor}x",
            type='SPEED',
            channel=channel,
            frame_start=strip.frame_start,
            frame_end=strip.frame_start + new_duration,
            seq1=strip
        )
        
        # Set stretch to input strip length if the property exists
        if hasattr(speed_effect, 'use_scale_to_length'):
            speed_effect.use_scale_to_length = True
        
        return speed_effect
    
    # create_text_overlay is now imported from vse_utils as add_text_strip
    
    def apply_color_balance(seq_editor, strip, lift=(1,1,1), gamma=(1,1,1), gain=(1,1,1)):
        """Apply color balance adjustments to a strip"""
        try:
            mod = strip.modifiers.new(name="ColorBalance", type='COLOR_BALANCE')
            mod.color_balance.lift = lift
            mod.color_balance.gamma = gamma
            mod.color_balance.gain = gain
            return mod
        except Exception as e:
            print(f"Error applying color balance: {e}")
            return None
    
    def apply_glow_effect(seq_editor, strip, threshold=0.5, blur_radius=3.0, quality=0.5, channel=None):
        """Apply a glow effect to a strip"""
        if channel is None:
            channel = strip.channel + 1
        
        glow = seq_editor.strips.new_effect(
            name=f"Glow_{strip.name}",
            type='GLOW',
            channel=channel,
            frame_start=strip.frame_start,
            frame_end=strip.frame_final_end,
            seq1=strip
        )
        
        if hasattr(glow, 'threshold'):
            glow.threshold = threshold
        if hasattr(glow, 'blur_radius'):
            glow.blur_radius = blur_radius
        if hasattr(glow, 'quality'):
            glow.quality = quality
        
        return glow

# --- Core utility functions from Chapter 2 & 3 ---
# Note: These utilities are now imported from vse_utils.py at the top of the file
# Keeping these wrappers for backward compatibility, but they now use the imported functions

# Redefine movie/sound strip functions to include the console output the original had
def add_movie_strip_with_logging(seq_editor, filepath, channel=1, frame_start=1, name=None):
    """Add a movie strip to the sequence editor with detailed logging."""
    try:
        log(f"Adding movie strip from '{filepath}' on channel {channel} at frame {frame_start}", "FUNCTION")
        if name:
            log(f"Using custom name: {name}", "DEBUG")
        
        movie_strip = add_movie_strip(seq_editor, filepath, channel, frame_start, name)
        
        log(f"Added movie strip: {movie_strip.name} (Type: {movie_strip.type})")
        log(f"  File: {movie_strip.filepath}", "DEBUG")
        log(f"  Duration: {movie_strip.frame_duration} frames", "DEBUG")
        log(f"  Resolution: {movie_strip.elements[0].orig_width}x{movie_strip.elements[0].orig_height} px", "DEBUG")
        log(f"  Timeline: Channel {movie_strip.channel}, Start Frame {movie_strip.frame_start}, End Frame {movie_strip.frame_final_end}", "DEBUG")
        
        return movie_strip
    except Exception as e:
        log(f"Failed to add movie strip from '{filepath}': {e}", "ERROR")
        log(f"Traceback: {traceback.format_exc()}", "DEBUG")
        raise

def add_sound_strip_with_logging(seq_editor, filepath, channel=1, frame_start=1, name=None):
    """Add a sound strip to the sequence editor with detailed logging."""
    try:
        log(f"Adding sound strip from '{filepath}' on channel {channel} at frame {frame_start}", "FUNCTION")
        if name:
            log(f"Using custom name: {name}", "DEBUG")
            
        sound_strip = add_sound_strip(seq_editor, filepath, channel, frame_start, name)
        
        log(f"Added sound strip: {sound_strip.name} (Type: {sound_strip.type})")
        if hasattr(sound_strip, 'sound') and sound_strip.sound and hasattr(sound_strip.sound, 'filepath'):
            log(f"  File: {sound_strip.sound.filepath}", "DEBUG")
        else:
            log(f"  File: {filepath}", "DEBUG")
        log(f"  Duration: {sound_strip.frame_duration} frames", "DEBUG")
        log(f"  Volume: {sound_strip.volume}, Pan: {sound_strip.pan}", "DEBUG")
        log(f"  Timeline: Channel {sound_strip.channel}, Start Frame {sound_strip.frame_start}, End Frame {sound_strip.frame_final_end}", "DEBUG")
        
        return sound_strip
    except Exception as e:
        log(f"Failed to add sound strip from '{filepath}': {e}", "ERROR")
        log(f"Traceback: {traceback.format_exc()}", "DEBUG")
        raise

def check_and_set_fps(seq_editor, scene):
    """Check FPS of strips and ensure it matches the scene FPS."""
    log("Checking and validating FPS settings", "FUNCTION")
    
    if not seq_editor or not seq_editor.sequences_all:
        log("No sequences to check for FPS", "WARN")
        return True, "No sequences to check"
    
    # Collect FPS information from movie strips
    fps_counts = {}
    strip_fps = {}
    
    log(f"Inspecting {len(seq_editor.sequences_all)} strips for FPS information", "DEBUG")
    for strip in seq_editor.sequences_all:
        if strip.type == 'MOVIE':
            # Get source FPS using the new Blender 4.4 method
            src_fps = None
            log(f"Checking FPS for movie strip: {strip.name}", "DEBUG")
            
            if hasattr(strip.elements[0], 'orig_fps') and strip.elements[0].orig_fps:
                src_fps = strip.elements[0].orig_fps
                log(f"  Found FPS from strip.elements[0].orig_fps: {src_fps}", "DEBUG")
            elif hasattr(strip, 'fps'):
                src_fps = strip.fps
                log(f"  Found FPS from strip.fps: {src_fps}", "DEBUG")
            else:
                log(f"  Could not determine FPS for {strip.name}", "WARN")
                
            if src_fps:
                strip_fps[strip.name] = src_fps
                fps_counts[src_fps] = fps_counts.get(src_fps, 0) + 1
    
    if not fps_counts:
        log("No movie strips found with FPS information", "WARN")
        return True, "No movie strips found"
    
    # Find most common FPS
    most_common_fps = max(fps_counts.items(), key=lambda x: x[1])[0]
    log(f"Most common FPS in strips: {most_common_fps}", "DEBUG")
    
    # Check if all strips have the same FPS
    if len(fps_counts) > 1:
        log("WARNING: Different FPS detected in strips:", "WARN")
        for strip_name, fps in strip_fps.items():
            log(f"  • {strip_name}: {fps} fps", "WARN")
        log(f"Using most common FPS: {most_common_fps}", "WARN")
    
    # Check scene FPS
    scene_fps = scene.render.fps / scene.render.fps_base
    log(f"Current scene FPS: {scene_fps}", "DEBUG")
    
    if abs(scene_fps - most_common_fps) > 0.01:  # Allow small floating-point differences
        log(f"Adjusting scene FPS from {scene_fps} to {most_common_fps}", "WARN")
        
        # Use set_scene_fps from vse_utils if available, otherwise set directly
        try:
            log("Attempting to use vse_utils.set_scene_fps", "DEBUG")
            from vse_utils import set_scene_fps
            set_scene_fps(scene, most_common_fps)
            log("Successfully updated scene FPS using vse_utils", "SUCCESS")
        except (ImportError, AttributeError) as e:
            log(f"Falling back to direct FPS setting: {e}", "DEBUG")
            scene.render.fps = int(most_common_fps)
            scene.render.fps_base = 1.0
            log(f"Directly set scene.render.fps={scene.render.fps}, fps_base={scene.render.fps_base}", "DEBUG")
        
        return True, f"Scene FPS adjusted to {most_common_fps}"
    
    log(f"FPS check completed - all good with {scene_fps} fps", "SUCCESS")
    return True, f"All good - Scene and strips using {scene_fps} fps"

def setup_test_sequence(seq_editor, test_media_dir):
    """Set up a test sequence with video clips for demonstration."""
    log("Setting up test sequence with video clips", "FUNCTION")
    
    # Clear all existing strips first
    if seq_editor.sequences_all:
        strip_count = len(seq_editor.sequences_all)
        log(f"Clearing {strip_count} existing strips", "DEBUG")
        
        # Try to use clear_all_strips from vse_utils if available
        try:
            log("Attempting to use vse_utils.clear_all_strips", "DEBUG")
            from vse_utils import clear_all_strips
            clear_all_strips(seq_editor)
            log("Successfully cleared strips using vse_utils", "SUCCESS")
        except (ImportError, AttributeError) as e:
            log(f"Falling back to manual strip removal: {e}", "DEBUG")
            log(f"Removing {strip_count} existing strips...", "WARN")
            
            try:
                for strip in seq_editor.sequences_all:
                    log(f"  Removing strip: {strip.name} ({strip.type})", "DEBUG")
                    seq_editor.sequences.remove(strip)
                log("All existing strips removed successfully", "SUCCESS")
            except Exception as e:
                log(f"Error removing strips: {e}", "ERROR")
                log(traceback.format_exc(), "DEBUG")
    else:
        log("No existing strips to clear", "DEBUG")
    
    # List of test video files to use
    video_files = [
        "SampleVideo_1280x720_2mb.mp4",
        "SampleVideo_1280x720_1mb.mp4",
        "SampleVideo_1280x720_5mb.mp4"
    ]
    log(f"Using {len(video_files)} test video files", "DEBUG")
    
    # Track channels for video and audio
    video_channel = 1
    audio_channel = 2
    
    # Start with frame 1 and add spacing between clips
    current_frame = 1
    frame_padding = 10  # Space between clips
    log(f"Starting at frame {current_frame} with {frame_padding} frames padding between clips", "DEBUG")
    
    # Name the channels
    try:
        if video_channel < len(seq_editor.channels):
            seq_editor.channels[video_channel-1].name = "Video"
            log(f"Named channel {video_channel} as 'Video'", "DEBUG")
        if audio_channel < len(seq_editor.channels):
            seq_editor.channels[audio_channel-1].name = "Audio"
            log(f"Named channel {audio_channel} as 'Audio'", "DEBUG")
    except Exception as e:
        log(f"Error naming channels: {e}", "WARN")
    
    # Add clips and store them for later use
    added_clips = []
    
    for i, video_filename in enumerate(video_files):
        video_path = os.path.join(test_media_dir, video_filename)
        log(f"Processing video file {i+1}/{len(video_files)}: {video_filename}", "DEBUG")
        log(f"  Full path: {video_path}", "DEBUG")
        
        if os.path.exists(video_path):
            log(f"  File exists, adding to sequence", "DEBUG")
            
            try:
                # Add video strip with logging
                video_strip = add_movie_strip_with_logging(
                    seq_editor, 
                    video_path, 
                    channel=video_channel, 
                    frame_start=current_frame,
                    name=f"Clip{i+1}"
                )
                
                # Add matching audio strip with logging
                audio_strip = add_sound_strip_with_logging(
                    seq_editor, 
                    video_path, 
                    channel=audio_channel, 
                    frame_start=current_frame,
                    name=f"Audio{i+1}"
                )
                
                # Store the pair
                added_clips.append((video_strip, audio_strip))
                log(f"  Added clip pair (video: {video_strip.name}, audio: {audio_strip.name})", "SUCCESS")
                
                # Update frame position for next clip
                prev_frame = current_frame
                current_frame = video_strip.frame_final_end + frame_padding
                log(f"  Next clip will start at frame {current_frame} (prev: {prev_frame}, duration: {video_strip.frame_final_duration})", "DEBUG")
            except Exception as e:
                log(f"  Error adding clip {video_filename}: {e}", "ERROR")
                log(traceback.format_exc(), "DEBUG")
        else:
            log(f"  Video file not found: {video_path}", "ERROR")
    
    log(f"Added {len(added_clips)} clips to the sequence", "SUCCESS")
    
    # Check and set FPS after adding all clips
    try:
        success, message = check_and_set_fps(seq_editor, C.scene)
        log(f"FPS Check: {message}", "INFO")
    except Exception as e:
        log(f"Error during FPS check: {e}", "ERROR")
        log(traceback.format_exc(), "DEBUG")
    
    return added_clips

# --- Chapter 5 Demo Functions ---

def demonstrate_transform_effects(seq_editor, clips):
    """
    Demonstrate applying transform effects to clips.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
    """
    log("Demonstrating Transform Effects", "EFFECT")
    
    # Need at least one clip
    if not clips:
        log("Need at least 1 clip to demonstrate transform effects. Skipping.", "WARN")
        return
    
    # Get the first clip
    video1, _ = clips[0]
    log(f"Using video clip '{video1.name}' for transform effect", "DEBUG")
    
    try:
        # Apply the transform effect - scale down slightly and move up
        # Note: apply_transform_effect is already imported from vse_utils
        log(f"Applying transform effect to '{video1.name}'", "DEBUG")
        transform = apply_transform_effect(
            seq_editor=seq_editor,
            strip=video1,
            offset_x=0,  # No horizontal change
            offset_y=50,  # Move up 50 pixels
            scale_x=0.9,  # Scale to 90% width
            scale_y=0.9,  # Scale to 90% height
            rotation=5.0,  # Rotate 5 degrees
            channel=5     # Place on channel 5
        )
        
        log(f"Applied transform effect to '{video1.name}'")
        log(f"  Positioned at offset (0, 50) pixels", "DEBUG")
        log(f"  Scaled to 90% of original size", "DEBUG")
        log(f"  Rotated 5 degrees", "DEBUG")
        log(f"  Effect placed on channel {transform.channel}", "DEBUG")
        log(f"Transform effect created successfully", "SUCCESS")
    except Exception as e:
        log(f"Error applying transform effect: {e}", "ERROR")
        log(traceback.format_exc(), "DEBUG")

def demonstrate_picture_in_picture(seq_editor, clips):
    """
    Demonstrate creating a picture-in-picture effect with two clips.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
    """
    print("\n--- Demonstrating Picture-in-Picture ---")
    
    # Need at least two clips
    if len(clips) < 2:
        print("Need at least 2 clips to demonstrate picture-in-picture. Skipping.")
        return
    
    # Get two clips
    video1, _ = clips[0]  # Main clip
    video2, _ = clips[1]  # PIP clip
    
    # Make sure video2 is at least as long as video1 for simplicity
    if video2.frame_final_end < video1.frame_final_end:
        print(f"Note: Adjusting second clip to match first clip's duration")
        # Extend video2's end to match video1
        # This is just for demo purposes, in real usage might need more care
        video2.frame_start = video1.frame_start
    
    # Create the picture-in-picture effect
    pip_effect = create_picture_in_picture(
        seq_editor=seq_editor,
        main_strip=video1,
        pip_strip=video2,
        pip_scale=0.3,   # 30% of original size
        position='top-right',
        channel=6
    )
    
    print(f"Created picture-in-picture effect")
    print(f"  Main clip: '{video1.name}'")
    print(f"  PIP clip: '{video2.name}' scaled to 30% in top-right corner")
    print(f"  Effect placed on channel {pip_effect.channel}")

def demonstrate_speed_effects(seq_editor, clips):
    """
    Demonstrate speed control effects.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
    """
    print("\n--- Demonstrating Speed Effects ---")
    
    # Need at least one clip
    if not clips:
        print("Need at least 1 clip to demonstrate speed effects. Skipping.")
        return
    
    # We'll need a fresh clip for this to avoid affecting other demos
    # Let's duplicate a clip and add it at the end
    last_video, last_audio = clips[-1]
    start_frame = last_video.frame_final_end + 50  # Add some spacing
    
    # Create a new strip using the same file
    slow_video = add_movie_strip(
        seq_editor=seq_editor,
        filepath=last_video.filepath,
        channel=last_video.channel,
        frame_start=start_frame,
        name="SlowClip"
    )
    
    # Also add a new strip for fast motion demo
    fast_video = add_movie_strip(
        seq_editor=seq_editor,
        filepath=last_video.filepath,
        channel=last_video.channel,
        frame_start=slow_video.frame_final_end + 50,  # After the slow clip
        name="FastClip"
    )
    
    # Create a slow-motion effect (0.5x speed)
    slow_effect = apply_speed_effect(
        seq_editor=seq_editor,
        strip=slow_video,
        speed_factor=0.5,  # Half speed (slow motion)
        channel=7
    )
    
    print(f"Applied slow-motion effect (0.5x) to '{slow_video.name}'")
    print(f"  Original duration: {slow_video.frame_final_duration} frames")
    print(f"  New duration: {slow_effect.frame_final_duration} frames")
    print(f"  Effect placed on channel {slow_effect.channel}")
    
    # Create a fast-motion effect (2x speed)
    fast_effect = apply_speed_effect(
        seq_editor=seq_editor,
        strip=fast_video,
        speed_factor=2.0,  # Double speed
        channel=7
    )
    
    print(f"Applied fast-motion effect (2.0x) to '{fast_video.name}'")
    print(f"  Original duration: {fast_video.frame_final_duration} frames")
    print(f"  New duration: {fast_effect.frame_final_duration} frames")
    print(f"  Effect placed on channel {fast_effect.channel}")

def demonstrate_text_overlays(seq_editor, clips):
    """
    Demonstrate text overlays and titles.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
    """
    print("\n--- Demonstrating Text Overlays ---")
    
    # Need at least one clip as reference for timeline
    if not clips:
        print("Need at least 1 clip as reference for text placement. Skipping.")
        return
    
    # Get reference to first and last clip for timeline reference
    first_video, _ = clips[0]
    last_video, _ = clips[-1]
    
    # Create a main title at the beginning
    # Note: create_text_overlay is already imported from vse_utils as add_text_strip
    title = create_text_overlay(
        seq_editor=seq_editor,
        text="Chapter 5: Effects and Adjustments",
        frame_start=first_video.frame_start,
        frame_end=first_video.frame_start + 90,  # 3-4 second title at 24-30fps
        channel=10,
        position='center',
        size=70,
        color=(1, 0.8, 0.2, 1)  # Golden yellow
    )
    
    print(f"Created main title overlay")
    print(f"  Text: '{title.text}'")
    print(f"  Duration: {title.frame_final_duration} frames")
    print(f"  Placed on channel {title.channel}")
    
    # Create a lower third subtitle for each clip
    for i, (video, _) in enumerate(clips):
        subtitle = create_text_overlay(
            seq_editor=seq_editor,
            text=f"Video Clip {i+1}",
            frame_start=video.frame_start + 10,  # Start a bit after clip starts
            frame_end=video.frame_start + 60,    # Short duration
            channel=9,
            position='bottom',
            size=40,
            color=(1, 1, 1, 1)  # White
        )
        
        print(f"Created lower third subtitle for '{video.name}'")
        print(f"  Text: '{subtitle.text}'")
    
    # Create an end credit
    credits = create_text_overlay(
        seq_editor=seq_editor,
        text="Created with Blender VSE Python API",
        frame_start=last_video.frame_final_end - 60,
        frame_end=last_video.frame_final_end,
        channel=10,
        position='bottom',
        size=50,
        color=(0.9, 0.9, 0.9, 1)  # Light gray
    )
    
    print(f"Created end credits overlay")
    print(f"  Text: '{credits.text}'")

def demonstrate_color_effects(seq_editor, clips):
    """
    Demonstrate color grading and effects.
    
    Args:
        seq_editor: The sequence editor
        clips: List of (video_strip, audio_strip) tuples
    """
    print("\n--- Demonstrating Color Effects ---")
    
    # Need at least one clip
    if not clips:
        print("Need at least 1 clip to demonstrate color effects. Skipping.")
        return
    
    # Apply color balance to the first clip
    video1, _ = clips[0]
    color_mod = apply_color_balance(
        seq_editor=seq_editor,
        strip=video1,
        lift=(1.1, 0.9, 0.9),    # Slightly reddish shadows
        gamma=(0.95, 1.05, 1.1),  # Slightly blue/green midtones
        gain=(1.0, 1.0, 1.1)      # Slightly blueish highlights
    )
    
    if color_mod:
        print(f"Applied color balance to '{video1.name}'")
        print(f"  Lift (shadows): (1.1, 0.9, 0.9) - reddish")
        print(f"  Gamma (midtones): (0.95, 1.05, 1.1) - blue/green tint")
        print(f"  Gain (highlights): (1.0, 1.0, 1.1) - slight blue tint")
    
    # If we have a second clip, apply a glow effect to it
    if len(clips) > 1:
        video2, _ = clips[1]
        glow = apply_glow_effect(
            seq_editor=seq_editor,
            strip=video2,
            threshold=0.7,     # Only brightest areas glow
            blur_radius=5.0,   # Moderate blur
            quality=0.8,       # High quality
            channel=8
        )
        
        print(f"Applied glow effect to '{video2.name}'")
        print(f"  Threshold: 0.7, Blur Radius: 5.0")
        print(f"  Effect placed on channel {glow.channel}")

def main():
    """
    Main function demonstrating Chapter 5 concepts: Effects and Adjustments.
    This function sets up a test sequence and then demonstrates various effects.
    """
    log("Blender VSE Python API Test - Chapter 5: Effects and Adjustments", "INFO")
    log("=============================================================", "INFO")
    
    # Get scene and ensure we have a sequence editor
    log("Initializing scene and sequence editor", "FUNCTION")
    try:
        scene = get_active_scene()
        log(f"Active scene: {scene.name}", "DEBUG")
        
        seq_editor = ensure_sequence_editor(scene)
        log(f"Sequence editor initialized: {seq_editor}", "DEBUG")
        
        # Log render settings
        resolution_x = scene.render.resolution_x
        resolution_y = scene.render.resolution_y
        fps = scene.render.fps / scene.render.fps_base
        log(f"Scene settings - Resolution: {resolution_x}x{resolution_y}, FPS: {fps}", "DEBUG")
    except Exception as e:
        log(f"Error initializing scene/sequence editor: {e}", "ERROR")
        log(traceback.format_exc(), "DEBUG")
        raise
    
    # --- Configuration ---
    # Use find_test_media_dir from vse_utils if available
    log("Looking for test media directory", "FUNCTION")
    try:
        test_media_dir = find_test_media_dir("/home/manuel/Movies/blender-movie-editor")
        log(f"Found test media directory using utility: {test_media_dir}", "SUCCESS")
    except (NameError, AttributeError) as e:
        # Fallback implementation if utility not available
        log(f"Could not use find_test_media_dir utility: {e}", "WARN")
        test_media_dir = "/home/manuel/Movies/blender-movie-editor"
        
        # Check if path exists, if not look for alternative locations
        if not os.path.exists(test_media_dir):
            log(f"Default path {test_media_dir} not found, trying alternatives", "WARN")
            # Try relative path from current script
            script_dir = os.path.dirname(os.path.abspath(__file__))
            alternative_paths = [
                os.path.join(script_dir, "../media"),
                os.path.join(script_dir, "media"),
                "/tmp/blender-test-media"
            ]
            
            for path in alternative_paths:
                log(f"Checking alternative path: {path}", "DEBUG")
                if os.path.exists(path):
                    test_media_dir = path
                    log(f"Using alternative media path: {test_media_dir}", "SUCCESS")
                    break
            
            if not os.path.exists(test_media_dir):
                log(f"Warning: Media directory not found. Please download test videos to: {test_media_dir}", "ERROR")
                log("Sample videos can be downloaded from: https://sample-videos.com/", "INFO")
    
    # --- 1. Set up a basic sequence with clips ---
    log("Starting to set up test sequence", "FUNCTION")
    try:
        clips = setup_test_sequence(seq_editor, test_media_dir)
        
        if not clips:
            log("No clips were added. Make sure test_media_dir is correct.", "ERROR")
            return
        
        log(f"Successfully set up test sequence with {len(clips)} clips", "SUCCESS")
    except Exception as e:
        log(f"Failed to set up test sequence: {e}", "ERROR")
        log(traceback.format_exc(), "DEBUG")
        return
    
    # Run each demo function with proper error handling
    try:
        # --- 2. Demonstrate transform effects ---
        log("Starting transform effects demonstration", "FUNCTION")
        demonstrate_transform_effects(seq_editor, clips)
        
        # --- 3. Demonstrate picture-in-picture ---
        log("Starting picture-in-picture demonstration", "FUNCTION")
        demonstrate_picture_in_picture(seq_editor, clips)
        
        # --- 4. Demonstrate speed effects ---
        log("Starting speed effects demonstration", "FUNCTION")
        demonstrate_speed_effects(seq_editor, clips)
        
        # --- 5. Demonstrate text overlays ---
        log("Starting text overlays demonstration", "FUNCTION")
        demonstrate_text_overlays(seq_editor, clips)
        
        # --- 6. Demonstrate color effects ---
        log("Starting color effects demonstration", "FUNCTION")
        demonstrate_color_effects(seq_editor, clips)
    except Exception as e:
        log(f"Error during effect demonstrations: {e}", "ERROR")
        log(traceback.format_exc(), "DEBUG")
    
    # --- Recap the Final State ---
    log("Final State of Strips", "INFO")
    try:
        strip_count = 0
        for strip in seq_editor.strips_all:
            strip_count += 1
            if strip_count <= 10:  # Only show first 10 to avoid overwhelming output
                log(f"- '{strip.name}' (Type: {strip.type}): Channel {strip.channel}, Frames {strip.frame_start}-{strip.frame_final_end}")
        
        if strip_count > 10:
            log(f"... and {strip_count - 10} more strips", "INFO")
        
        log(f"Total strips in final sequence: {strip_count}", "SUCCESS")
    except Exception as e:
        log(f"Error while printing final state: {e}", "ERROR")
    
    log("Test for Chapter 5 completed! Review the VSE timeline and console output.", "SUCCESS")

# Run the main function to demonstrate Chapter 5 concepts
if __name__ == "__main__":
    try:
        log("Starting Chapter 5 demonstration script", "INFO")
        main()
        log("Script execution completed successfully", "SUCCESS")
    except Exception as e:
        log(f"Unhandled error in main execution: {e}", "ERROR")
        log(traceback.format_exc(), "DEBUG")
        print(f"\n\nERROR: Script failed with error: {e}")
    finally:
        log("Script execution finished", "INFO")