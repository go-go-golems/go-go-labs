# context.area: SEQUENCE_EDITOR
# Don't remove the comment above. It's important.

"""
Blender VSE DNA Explorer Script

This script is both a tool and a learning resource for exploring Blender's Video Sequence Editor (VSE) data structures ("DNA") and internals.

It demonstrates how Blender's data-block system, DNA, and RNA work in practice, and how the VSE fits into Blender's architecture. The script:
- Explores and prints the structure of the Sequence Editor, strips, channels, effects, and modifiers
- Logs all properties, methods, and attributes of these objects, including their types, descriptions, and options
- Explains the concepts of DNA (data layout), RNA (runtime API), and IDs (data-blocks) as they relate to the VSE
- Shows how the Python API (bpy) is generated from Blender's internal DNA/RNA definitions
- Provides commentary and references to help you understand Blender's architecture and scripting model

Usage:
    Run this script in Blender's Text Editor or Python Console while in the Video Editing workspace.
    The output will be printed to the console. Read the comments and docstrings for learning!

Background:
    - Blender's core data is organized as ID data-blocks (e.g. Scene, Object, Mesh, SequenceEditor, etc.), each defined in C structs (DNA) and exposed to Python via RNA.
    - The VSE (Video Sequence Editor) is a domain within Blender, with its own data-blocks (strips, channels, etc.) and operators.
      - A domain in Blender represents an isolated functional area with its own data structures, code, and logic. Domains help organize the codebase by grouping related functionality (e.g. objects, meshes, materials, nodes, etc.) into separate modules that can evolve independently.
      - Each domain typically has its own data-blocks, operators, and UI code, but follows common patterns like using DNA/RNA for data management.
    - The Python API (bpy) is auto-generated from RNA, which wraps the DNA structs and provides property metadata, type info, and documentation.
    - This script introspects these APIs, showing you the real structure and options available for scripting and add-on development.
    - For more, see the companion doc: 02-blender-internals-api.md
"""

import bpy # type: ignore   
import inspect
from pprint import pformat
from typing import Any, Dict, List, Optional, Set, Tuple

# Common shortcuts for Blender data and context
D = bpy.data
C = bpy.context

def print_blender_architecture_overview():
    """
    Print a high-level overview of Blender's architecture, focusing on DNA, RNA, IDs, and the VSE domain.
    """
    print("""
==================== BLENDER ARCHITECTURE OVERVIEW ====================
Blender is built on a layered architecture:
- DNA: C struct definitions for all data-blocks (the schema for .blend files)
- RNA: Reflection/introspection API, auto-generated from DNA, powers the Python API and UI
- ID data-blocks: All persistent data (scenes, objects, meshes, sequence editors, etc.)
- Main database: Holds all IDs in memory
- Operators: Tools and actions (Controller in MVC)
- Editors: UI views (like the Video Sequence Editor)

The Video Sequence Editor (VSE) is a domain in Blender for non-linear video editing. Its core data-blocks are:
- SequenceEditor: Holds strips, channels, and settings for a scene's VSE
- Strip: Represents a video, image, sound, or effect on the timeline
- Channel: Timeline tracks for organizing strips
- EffectStrip: Special strips for transitions/effects
- Modifiers: Non-destructive adjustments to strips

The Python API (bpy) exposes all of this via RNA. This script introspects these APIs, showing you the real structure and options available for scripting and add-on development.
=======================================================================
""")

def get_property_info(obj: Any, prop_name: str) -> Dict[str, Any]:
    """
    Get detailed information about a property from Blender's RNA system.
    
    Blender's RNA (Runtime Navigable API) wraps the DNA (C struct) fields and provides metadata for scripting and UI.
    This function extracts type, description, and options for a property.
    
    Args:
        obj: The object containing the property (must have bl_rna)
        prop_name: Name of the property
    Returns:
        Dictionary containing property information (type, description, enum values, etc.)
    """
    prop = obj.bl_rna.properties[prop_name]
    info = {
        'name': prop.name,
        'type': prop.type,
        'description': prop.description,
        'is_readonly': prop.is_readonly,
        'is_animatable': prop.is_animatable,
        'is_overridable': prop.is_overridable,
    }
    # ENUMs: list possible values
    if prop.type == 'ENUM':
        info['enum_items'] = [(item.identifier, item.name, item.description) 
                            for item in prop.enum_items]
    elif prop.type == 'POINTER':
        info['pointer_type'] = prop.fixed_type.identifier
    elif prop.type == 'COLLECTION':
        info['collection_type'] = prop.fixed_type.identifier
    return info

def explore_object(obj: Any, path: str = "", visited: Optional[Set[int]] = None) -> Dict[str, Any]:
    """
    Recursively explore an object's properties and structure using Blender's RNA system.
    This is a practical demonstration of how RNA exposes DNA-defined data-blocks to Python.
    
    Args:
        obj: The object to explore
        path: Current path in the object hierarchy (for logging)
        visited: Set of object IDs already visited (to prevent cycles)
    Returns:
        Dictionary containing object information (type, path, properties, methods, attributes)
    """
    if visited is None:
        visited = set()
    obj_id = id(obj)
    if obj_id in visited:
        return {'type': 'cycle', 'path': path}
    visited.add(obj_id)
    info = {
        'type': type(obj).__name__,
        'path': path,
        'properties': {},
        'methods': [],
        'attributes': {}
    }
    # Properties (from RNA)
    if hasattr(obj, 'bl_rna'):
        for prop_name in obj.bl_rna.properties.keys():
            try:
                info['properties'][prop_name] = get_property_info(obj, prop_name)  # type: ignore
            except Exception as e:
                info['properties'][prop_name] = {'error': str(e)}  # type: ignore
    # Methods (from Python class)
    for name, method in inspect.getmembers(obj, predicate=inspect.ismethod):
        if not name.startswith('_'):
            info['methods'].append(name)  # type: ignore
    # Attributes (from Python dir)
    for name in dir(obj):
        if not name.startswith('_'):
            try:
                value = getattr(obj, name)
                if not inspect.ismethod(value):
                    info['attributes'][name] = type(value).__name__  # type: ignore
            except Exception as e:
                info['attributes'][name] = f'Error: {str(e)}'  # type: ignore
    return info

def explore_strip_types() -> Dict[str, Any]:
    """
    Explore all available strip types in the VSE.
    In Blender's DNA/RNA, each strip type (MovieStrip, ImageStrip, SoundStrip, EffectStrip, etc.)
    is a subclass of the base Strip type, with its own properties and options.
    This function introspects all types ending with 'Strip' in bpy.types.
    Returns:
        Dictionary containing information about strip types
    """
    info = {}
    for name in dir(bpy.types):
        if name.endswith('Strip'):
            strip_type = getattr(bpy.types, name)
            info[name] = explore_object(strip_type, f"bpy.types.{name}")
    return info

def explore_sequence_editor() -> Dict[str, Any]:
    """
    Explore the sequence editor structure for the current scene.
    The SequenceEditor is the main data-block for the VSE, holding strips, channels, and settings.
    Returns:
        Dictionary containing sequence editor information
    """
    scene = C.scene
    if not scene.sequence_editor:
        return {'error': 'No sequence editor found'}
    seq_editor = scene.sequence_editor
    return explore_object(seq_editor, "scene.sequence_editor")

def explore_channels() -> Dict[str, Any]:
    """
    Explore channel properties and structure.
    Channels are timeline tracks in the VSE, each holding strips. Channels are managed as a collection in the SequenceEditor.
    Returns:
        Dictionary containing channel information
    """
    scene = C.scene
    if not scene.sequence_editor:
        return {'error': 'No sequence editor found'}
    info = {}
    for i, channel in enumerate(scene.sequence_editor.channels):
        info[f"channel_{i}"] = explore_object(channel, f"scene.sequence_editor.channels[{i}]")
    return info

def explore_effects() -> Dict[str, Any]:
    """
    Explore effect types and their properties.
    Effect strips (e.g. Cross, GammaCross, Wipe, etc.) are special strip types for transitions and effects.
    This function introspects all types ending with 'EffectStrip' in bpy.types.
    Returns:
        Dictionary containing effect information
    """
    info = {}
    for name in dir(bpy.types):
        if name.endswith('EffectStrip'):
            effect_type = getattr(bpy.types, name)
            info[name] = explore_object(effect_type, f"bpy.types.{name}")
    return info

def explore_modifiers() -> Dict[str, Any]:
    """
    Explore modifier types and their properties.
    Modifiers are non-destructive adjustments to strips (e.g. color balance, curves, transform, etc.).
    This function introspects all types ending with 'Modifier' in bpy.types.
    Returns:
        Dictionary containing modifier information
    """
    info = {}
    for name in dir(bpy.types):
        if name.endswith('Modifier'):
            modifier_type = getattr(bpy.types, name)
            info[name] = explore_object(modifier_type, f"bpy.types.{name}")
    return info

def print_section(title: str, data: Dict[str, Any], indent: int = 0) -> None:
    """
    Print a section of the exploration results, with a title and pretty formatting.
    Args:
        title: Section title
        data: Section data
        indent: Indentation level
    """
    indent_str = "  " * indent
    print(f"\n{indent_str}{'=' * 20}")
    print(f"{indent_str}{title}")
    print(f"{indent_str}{'=' * 20}")
    print(pformat(data, indent=2, width=120))

def print_dna_rna_id_summary():
    """
    Print a summary of Blender's DNA, RNA, and ID system, with a focus on how the VSE fits in.
    """
    print("""
-------------------- BLENDER DNA / RNA / ID SYSTEM --------------------
- DNA: Blender's internal schema for all data-blocks, written in C headers (makesdna)
- RNA: Reflection API, auto-generated from DNA, exposes properties, types, and docs to Python and UI
- ID: All persistent data in Blender is an ID data-block (Scene, Object, Mesh, SequenceEditor, etc.)
- Main: The in-memory database holding all IDs
- VSE: The Video Sequence Editor is a domain with its own data-blocks (SequenceEditor, Strip, Channel, etc.)
- Python API: The bpy module is generated from RNA, so all properties and types you see here are defined in Blender's C code and exposed via RNA
- Operators: Actions (like adding strips, cutting, rendering) are implemented as operators, which act on these data-blocks
----------------------------------------------------------------------
""")

def main():
    """
    Main function to explore VSE DNA and internals, with educational commentary.
    """
    print("Blender VSE DNA Explorer - Internals Edition")
    print("=============================================")
    print_blender_architecture_overview()
    print_dna_rna_id_summary()
    # Explore sequence editor
    print_section("Sequence Editor Structure (scene.sequence_editor)", explore_sequence_editor())
    # Explore strip types
    print_section("Strip Types (bpy.types.*Strip)", explore_strip_types())
    # Explore channels
    print_section("Channel Properties (scene.sequence_editor.channels)", explore_channels())
    # Explore effects
    print_section("Effect Types (bpy.types.*EffectStrip)", explore_effects())
    # Explore modifiers
    print_section("Modifier Types (bpy.types.*Modifier)", explore_modifiers())
    print("\nExploration completed!\n")
    print("For more, see the companion doc: 02-blender-internals-api.md and the official Blender API docs.")

main() 