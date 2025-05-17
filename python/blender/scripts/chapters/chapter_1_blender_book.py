# context.area: SEQUENCE_EDITOR
# Don't remove the comment above. It's important.

"""
Blender VSE Python API Test Script - Chapter 1 Concepts

This script demonstrates basic concepts from Chapter 1 of the Blender VSE Python API guide.
It shows how to:
1. Access the VSE and its components
2. Create and manage sequence editors
3. Work with basic VSE data structures
4. Handle scene and context

Usage:
    Run this script from Blender's Text Editor or Python Console while in the Video Editing workspace.
"""

import bpy
from mathutils import * # type: ignore

# Common shortcuts for Blender data and context
D = bpy.data
C = bpy.context

def get_active_scene():
    """
    Get the currently active scene.
    
    Returns:
        bpy.types.Scene: The active scene object
    """
    return C.scene

def ensure_sequence_editor(scene=None):
    """
    Ensure a scene has a sequence editor, creating one if it doesn't exist.
    
    Args:
        scene (bpy.types.Scene, optional): The scene to check. If None, uses active scene.
        
    Returns:
        bpy.types.SequenceEditor: The sequence editor for the scene
    """
    if scene is None:
        scene = get_active_scene()
    
    # Create sequence editor if it doesn't exist
    if not scene.sequence_editor:
        scene.sequence_editor_create()
    
    return scene.sequence_editor

def check_and_set_fps(seq_editor, scene):
    """
    Check FPS of all strips and ensure they match the scene FPS.
    If strips have different FPS, use the most common one and set scene to match.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor to check
        scene (bpy.types.Scene): The scene to check/modify
        
    Returns:
        tuple: (bool, str) - (success, message)
    """
    if not seq_editor or not seq_editor.sequences_all:
        return True, "No sequences to check"
    
    # Collect FPS information from movie strips
    fps_counts = {}
    strip_fps = {}
    
    for strip in seq_editor.sequences_all:
        if strip.type == 'MOVIE':
            # Get source FPS using the new Blender 4.4 method
            src_fps = None
            if hasattr(strip.elements[0], 'orig_fps') and strip.elements[0].orig_fps:
                src_fps = strip.elements[0].orig_fps
            elif hasattr(strip, 'fps'):
                src_fps = strip.fps
                
            if src_fps:
                strip_fps[strip.name] = src_fps
                fps_counts[src_fps] = fps_counts.get(src_fps, 0) + 1
    
    if not fps_counts:
        return True, "No movie strips found"
    
    # Find most common FPS
    most_common_fps = max(fps_counts.items(), key=lambda x: x[1])[0]
    
    # Check if all strips have the same FPS
    if len(fps_counts) > 1:
        print("\nWARNING: Different FPS detected in strips:")
        for strip_name, fps in strip_fps.items():
            print(f"  • {strip_name}: {fps} fps")
        print(f"\nUsing most common FPS: {most_common_fps}")
    
    # Check scene FPS
    scene_fps = scene.render.fps / scene.render.fps_base
    
    if abs(scene_fps - most_common_fps) > 0.01:  # Allow small floating-point differences
        print(f"\nAdjusting scene FPS from {scene_fps} to {most_common_fps}")
        scene.render.fps = int(most_common_fps)
        scene.render.fps_base = 1.0
        return True, f"Scene FPS adjusted to {most_common_fps}"
    
    return True, f"All good - Scene and strips using {scene_fps} fps"

def print_sequence_info(seq_editor):
    """
    Print information about the sequence editor and its contents.
    
    Args:
        seq_editor (bpy.types.SequenceEditor): The sequence editor to analyze
    """
    print("\nSequence Editor Information:")
    print(f"Number of strips: {len(seq_editor.strips)}")
    print(f"Number of channels: {len(seq_editor.channels)}")
    
    # Print info about each strip
    print("\nStrips:")
    for strip in seq_editor.strips:
        print(f"- {strip.name} (Type: {strip.type}, Channel: {strip.channel})")

def list_available_scenes():
    """
    List all available scenes in the blend file.
    
    Returns:
        list: List of scene names
    """
    scenes = [scene.name for scene in D.scenes]
    print("\nAvailable Scenes:")
    for scene in scenes:
        print(f"- {scene}")
    return scenes

def main():
    """
    Main function demonstrating Chapter 1 concepts.
    """
    print("Blender VSE Python API Test - Chapter 1")
    print("=====================================")
    
    # Get active scene
    scene = get_active_scene()
    print(f"\nActive Scene: {scene.name}")
    
    # Ensure sequence editor exists
    seq_editor = ensure_sequence_editor(scene)
    print(f"\nSequence Editor created/accessed successfully")
    
    # Check and set FPS
    success, message = check_and_set_fps(seq_editor, scene)
    print(f"\nFPS Check: {message}")
    
    # Print sequence information
    print_sequence_info(seq_editor)
    
    # List available scenes
    list_available_scenes()
    
    print("\nTest completed successfully!")

main()