# Test script to verify imports work correctly with new directory structure

import bpy
import sys
import os
import importlib

# Set up import paths - handle both direct execution and MCP execution
try:
    # When run as a file
    scripts_dir = os.path.dirname(os.path.abspath(__file__))
except NameError:
    # When run through MCP or pasted code
    # Use a relative path from current directory if possible
    try:
        scripts_dir = os.path.join(os.getcwd(), 'scripts')
        if not os.path.exists(scripts_dir):
            scripts_dir = os.getcwd()  # Current directory might be scripts already
    except:
        # Fallback to hardcoded path
        scripts_dir = '/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender/scripts'

if scripts_dir not in sys.path:
    sys.path.append(scripts_dir)

# Add subdirectories to path
utils_dir = os.path.join(scripts_dir, 'utils')
tests_dir = os.path.join(scripts_dir, 'tests')
chapters_dir = os.path.join(scripts_dir, 'chapters')
investigations_dir = os.path.join(scripts_dir, 'investigations')

# Add each directory to path
for directory in [utils_dir, tests_dir, chapters_dir, investigations_dir]:
    if directory not in sys.path:
        sys.path.append(directory)

# Report path setup
print("\n===== PYTHON PATH SETUP =====\n")
print(f"Main scripts directory: {scripts_dir}")
print("\nPython sys.path now contains:")
for i, path in enumerate(sys.path):
    print(f"  {i}: {path}")

# Test imports from various subdirectories
print("\n===== TESTING IMPORTS =====\n")

try:
    # Import from utils
    print("Importing from utils...")
    from utils import vse_utils
    importlib.reload(vse_utils)
    # Get module file path safely
    vse_utils_path = getattr(vse_utils, '__file__', 'Module in memory')
    print(f"  SUCCESS: Imported vse_utils from {vse_utils_path}")
    
    # Test a function from vse_utils
    print("  Testing function from vse_utils...")
    scene = bpy.context.scene
    seq_editor = vse_utils.ensure_sequence_editor(scene)
    print(f"  Result: {seq_editor}")
    
    # Import from tests
    print("\nImporting from tests...")
    from tests import chapter_3_demos
    importlib.reload(chapter_3_demos)
    # Get module file path safely
    chapter_3_demos_path = getattr(chapter_3_demos, '__file__', 'Module in memory')
    print(f"  SUCCESS: Imported chapter_3_demos from {chapter_3_demos_path}")
    
    # Import from chapters
    print("\nImporting from chapters...")
    from chapters import chapter_3_blender_book
    importlib.reload(chapter_3_blender_book)
    # Get module file path safely
    chapter_3_book_path = getattr(chapter_3_blender_book, '__file__', 'Module in memory')
    print(f"  SUCCESS: Imported chapter_3_blender_book from {chapter_3_book_path}")
    
    # Import from investigations
    print("\nImporting from investigations...")
    from investigations import segment_removal_investigation
    importlib.reload(segment_removal_investigation)
    # Get module file path safely
    segment_removal_path = getattr(segment_removal_investigation, '__file__', 'Module in memory')
    print(f"  SUCCESS: Imported segment_removal_investigation from {segment_removal_path}")
    
    print("\nAll imports successful!")
    
except ImportError as e:
    print(f"  FAILED: Import error: {e}")
except Exception as e:
    print(f"  ERROR: {e}")

print("\n===== IMPORT TEST COMPLETE =====\n")