# Effect Utilities for Blender VSE
# Helper functions for Chapter 5: Applying Effects and Adjustments

import bpy # type: ignore
import os
from math import radians

# Common shortcuts for Blender data and context
D = bpy.data
C = bpy.context

# Helper function to ensure frame numbers are integers
def ensure_integer_frame(value):
    """Helper function to ensure frame numbers are integers."""
    try:
        return int(float(value))
    except (TypeError, ValueError):
        return value

def apply_transform_effect(seq_editor, strip, offset_x=0, offset_y=0, scale_x=1.0, scale_y=1.0, rotation=0.0, channel=None):
    """
    Apply a transform effect to a strip to adjust position, scale, and rotation.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        strip (bpy.types.Strip): The strip to transform
        offset_x (float): Horizontal offset in pixels
        offset_y (float): Vertical offset in pixels
        scale_x (float): Horizontal scale factor (1.0 = 100%)
        scale_y (float): Vertical scale factor (1.0 = 100%)
        rotation (float): Rotation in degrees
        channel (int, optional): Channel for the effect. If None, uses strip.channel + 1
    
    Returns:
        bpy.types.Strip: The created transform effect strip
    """
    if channel is None:
        channel = strip.channel + 1
    
    # Create the transform effect with integer frame values
    transform = seq_editor.strips.new_effect(
        name=f"Transform_{strip.name}",
        type='TRANSFORM',
        channel=channel,
        frame_start=ensure_integer_frame(strip.frame_start),
        frame_end=ensure_integer_frame(strip.frame_final_end),
        seq1=strip
    )
    
    # Apply the transform properties
    transform.transform.offset_x = offset_x
    transform.transform.offset_y = offset_y
    transform.transform.scale_x = scale_x
    transform.transform.scale_y = scale_y
    transform.transform.rotation = radians(rotation)  # Convert degrees to radians
    
    print(f"Applied transform effect to '{strip.name}'")
    print(f"  Position: ({offset_x}, {offset_y}), Scale: ({scale_x}, {scale_y}), Rotation: {rotation}°")
    
    return transform

def create_picture_in_picture(seq_editor, main_strip, pip_strip, pip_scale=0.3, position='top-right', channel=None):
    """
    Create a picture-in-picture effect with one strip inside another.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        main_strip (bpy.types.Strip): The main (background) strip
        pip_strip (bpy.types.Strip): The strip to show as picture-in-picture
        pip_scale (float): Scale factor for the PIP (0.3 = 30% of original size)
        position (str): Position of PIP: 'top-right', 'top-left', 'bottom-right', 'bottom-left', 'center'
        channel (int, optional): Channel for the transform effect
    
    Returns:
        bpy.types.Strip: The created transform effect strip
    """
    if channel is None:
        channel = max(main_strip.channel, pip_strip.channel) + 1
    
    # Calculate position based on specified corner
    # Assume 1920x1080 as base resolution (adjust if needed)
    width = C.scene.render.resolution_x
    height = C.scene.render.resolution_y
    padding = 20  # Padding from the edge
    
    # Calculate offsets based on position parameter
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
    else:  # 'center' or default
        offset_x = 0
        offset_y = 0
    
    # Apply transform to PIP strip
    transform = apply_transform_effect(
        seq_editor=seq_editor,
        strip=pip_strip,
        offset_x=offset_x,
        offset_y=offset_y,
        scale_x=pip_scale,
        scale_y=pip_scale,
        channel=channel
    )
    
    print(f"Created picture-in-picture effect at {position}")
    
    return transform

def apply_speed_effect(seq_editor, strip, speed_factor, channel=None):
    """
    Apply a speed effect to a strip to make it play faster or slower.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        strip (bpy.types.Strip): The strip to adjust speed
        speed_factor (float): Speed factor (2.0 = 2x speed, 0.5 = half speed)
        channel (int, optional): Channel for the effect. If None, uses strip.channel + 1
    
    Returns:
        bpy.types.Strip: The created speed effect strip
    """
    if channel is None:
        channel = strip.channel + 1
    
    # Calculate new duration based on speed factor
    original_duration = strip.frame_final_duration
    new_duration = int(original_duration / speed_factor)
    
    # Create the speed effect with integer frame values
    frame_start = ensure_integer_frame(strip.frame_start)
    speed_effect = seq_editor.strips.new_effect(
        name=f"Speed_{strip.name}_{speed_factor}x",
        type='SPEED',
        channel=channel,
        frame_start=frame_start,
        frame_end=frame_start + new_duration,
        seq1=strip
    )
    
    # Set stretch to input strip length
    if hasattr(speed_effect, 'use_scale_to_length'):
        speed_effect.use_scale_to_length = True
    
    # Note: in Blender 4.x, the approach might be different depending on API
    # This is a typical approach but may need adjustments based on Blender version
    
    print(f"Applied speed effect to '{strip.name}'")
    print(f"  Speed Factor: {speed_factor}x")
    print(f"  Original Duration: {original_duration} frames")
    print(f"  New Duration: {new_duration} frames")
    
    return speed_effect

def create_text_overlay(seq_editor, text, frame_start, frame_end, channel=10,
                       position='center', size=50, color=(1,1,1,1)):
    """
    Create a text overlay strip with specified properties.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        text (str): The text content to display
        frame_start (int): Start frame for the text
        frame_end (int): End frame for the text
        channel (int): Channel to place the text strip
        position (str): Position of text: 'center', 'top', 'bottom'
        size (int): Font size
        color (tuple): RGBA color values (0-1 for each component)
    
    Returns:
        bpy.types.Strip: The created text strip
    """
    # Create the text strip with integer frame values
    text_strip = seq_editor.strips.new_effect(
        name=f"Text_{text[:10]}",
        type='TEXT',
        channel=channel,
        frame_start=ensure_integer_frame(frame_start),
        frame_end=ensure_integer_frame(frame_end)
    )
    
    # Set the text strip properties
    text_strip.text = text
    text_strip.font_size = size
    text_strip.color = color[:3]  # RGB only, alpha might be separate
    
    # Position the text
    if position == 'center':
        text_strip.location = (0.5, 0.5)
    elif position == 'top':
        text_strip.location = (0.5, 0.8)
    elif position == 'bottom':
        text_strip.location = (0.5, 0.2)
    
    # Set alignment
    if hasattr(text_strip, 'align_x'):
        text_strip.align_x = 'CENTER'
    if hasattr(text_strip, 'align_y'):
        if position == 'bottom':
            text_strip.align_y = 'TOP'
        elif position == 'top':
            text_strip.align_y = 'BOTTOM'
        else:
            text_strip.align_y = 'CENTER'
    
    print(f"Created text overlay: '{text}'")
    print(f"  Position: {position}, Size: {size}")
    print(f"  Frames: {frame_start} to {frame_end}")
    
    return text_strip

def apply_color_balance(seq_editor, strip, lift=(1,1,1), gamma=(1,1,1), gain=(1,1,1)):
    """
    Apply color balance adjustments to a strip using a modifier.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        strip (bpy.types.Strip): The strip to adjust
        lift (tuple): RGB values for shadows adjustment (1,1,1 = no change)
        gamma (tuple): RGB values for midtones adjustment (1,1,1 = no change)
        gain (tuple): RGB values for highlights adjustment (1,1,1 = no change)
    
    Returns:
        bpy.types.StripModifier: The created modifier or None on failure
    """
    try:
        # Create the color balance modifier
        mod = strip.modifiers.new(name="ColorBalance", type='COLOR_BALANCE')
        
        # Set the color balance values
        mod.color_balance.lift = lift
        mod.color_balance.gamma = gamma
        mod.color_balance.gain = gain
        
        print(f"Applied color balance to '{strip.name}'")
        print(f"  Lift (shadows): {lift}")
        print(f"  Gamma (midtones): {gamma}")
        print(f"  Gain (highlights): {gain}")
        
        return mod
    except Exception as e:
        print(f"Error applying color balance: {e}")
        return None

def apply_glow_effect(seq_editor, strip, threshold=0.5, blur_radius=3.0, quality=5, channel=None):
    """
    Apply a glow effect to a strip.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor
        strip (bpy.types.Strip): The strip to apply glow to
        threshold (float): Threshold for bright areas (0.0-1.0)
        blur_radius (float): Radius of the glow effect
        quality (int): Quality of the glow effect (1-10, higher is better)
        channel (int, optional): Channel for the effect. If None, uses strip.channel + 1
    
    Returns:
        bpy.types.Strip: The created glow effect strip
    """
    if channel is None:
        channel = strip.channel + 1
    
    # Ensure frame values are integers
    frame_start = ensure_integer_frame(strip.frame_start)
    frame_end = ensure_integer_frame(strip.frame_final_end)
    
    # Create the glow effect
    glow = seq_editor.strips.new_effect(
        name=f"Glow_{strip.name}",
        type='GLOW',
        channel=channel,
        frame_start=frame_start,
        frame_end=frame_end,
        seq1=strip
    )
    
    # Set glow properties
    if hasattr(glow, 'threshold'):
        glow.threshold = threshold
    if hasattr(glow, 'blur_radius'):
        glow.blur_radius = blur_radius
    if hasattr(glow, 'quality'):
        # Ensure quality is an integer - Blender 4.4 requires this
        quality_int = int(quality) if isinstance(quality, float) else quality
        glow.quality = quality_int
    
    print(f"Applied glow effect to '{strip.name}'")
    print(f"  Threshold: {threshold}, Blur Radius: {blur_radius}, Quality: {quality}")
    
    return glow