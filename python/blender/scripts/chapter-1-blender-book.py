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
    Run this script in Blender's Text Editor or Python Console while in the Video Editing workspace.
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
    
    # Print sequence information
    print_sequence_info(seq_editor)
    
    # List available scenes
    list_available_scenes()
    
    print("\nTest completed successfully!")

main()