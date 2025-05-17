# Investigation of Module Import Issues

"""
This script investigates why importing a module with hyphens in the name fails,
and demonstrates the proper way to import Python modules.

Python module names must follow variable naming rules, which means they should:
- Start with a letter or underscore
- Contain only letters, numbers, and underscores
- NOT contain hyphens, spaces, or other special characters

This is why 'appendix-explore-dna-video-sequence-editor' cannot be imported directly.
"""

import bpy
import os
import sys
import importlib

# Add script directory to path if needed
script_dir = '/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender/scripts'
if script_dir not in sys.path:
    sys.path.append(script_dir)

# Debug function to check paths and module loading
def debug_module_loading():
    """Print debugging information about module loading."""
    print("\n===== MODULE IMPORT DEBUGGING =====\n")
    print(f"Current working directory: {os.getcwd()}")
    print(f"Script directory: {script_dir}")
    print("\nPython sys.path contains:")
    for i, path in enumerate(sys.path):
        print(f"  {i}: {path}")
    
    print("\nListing .py files in script directory:")
    try:
        for file in os.listdir(script_dir):
            if file.endswith(".py"):
                print(f"  {file}")
    except Exception as e:
        print(f"  Error listing files: {e}")
    
    print("\nTrying different import approaches:")
    
    # Try different import approaches
    modules_to_try = [
        "vse_utils",
        "appendix_explore_dna_video_sequence_editor",  # Using underscores
        "appendix-explore-dna-video-sequence-editor"   # Using hyphens (should fail)
    ]
    
    for module_name in modules_to_try:
        try:
            print(f"  Attempting to import {module_name}...")
            # Use __import__ for more control
            module = __import__(module_name)
            print(f"    SUCCESS! Module loaded: {module.__name__}")
        except ImportError as e:
            print(f"    FAILED: {e}")
        except Exception as e:
            print(f"    ERROR: {e}")
    
    print("\nChecking file existence directly:")
    files_to_check = [
        "vse_utils.py",
        "appendix_explore_dna_video_sequence_editor.py",
        "appendix-explore-dna-video-sequence-editor.py"
    ]
    
    for filename in files_to_check:
        file_path = os.path.join(script_dir, filename)
        exists = os.path.exists(file_path)
        print(f"  {filename}: {'EXISTS' if exists else 'NOT FOUND'} at {file_path}")
    
    print("\n=== END MODULE IMPORT DEBUGGING ===\n")
    
# Demonstrate the correct way to import a script with hyphens
def demonstrate_correct_import():
    """Demonstrate the correct way to import a script with hyphens in the name."""
    print("\n===== DEMONSTRATING CORRECT IMPORT =====\n")
    
    # Option 1: Rename the file first (recommended)
    print("Option 1: Rename the file to use underscores instead of hyphens")
    print("  Original: appendix-explore-dna-video-sequence-editor.py")
    print("  Correct:  appendix_explore_dna_video_sequence_editor.py")
    
    # Option 2: Import by loading the source directly
    print("\nOption 2: Import by loading and executing the source directly")
    try:
        hyphen_filename = "appendix-explore-dna-video-sequence-editor.py"
        hyphen_path = os.path.join(script_dir, hyphen_filename)
        
        if os.path.exists(hyphen_path):
            print(f"  Loading {hyphen_filename} using exec()...")
            namespace = {}
            with open(hyphen_path, 'r') as f:
                code = f.read()
                # Execute the code in a namespace
                exec(code, namespace)
            print("  Success! File loaded via exec()")
            
            # Access functions defined in the file
            if 'main' in namespace:
                print("  Found 'main' function in the loaded file")
            else:
                print("  Could not find 'main' function in the loaded file")
        else:
            print(f"  File not found: {hyphen_path}")
    except Exception as e:
        print(f"  Error loading file: {e}")
    
    print("\nCONCLUSION:")
    print("1. Python modules MUST use underscores instead of hyphens")
    print("2. Import will fail if the file has hyphens in its name")
    print("3. Best practice: Always name Python files using underscores")
    print("   Example: my_module.py NOT my-module.py")
    
    print("\n=== END DEMONSTRATION ===\n")

# Main function
def main():
    """Run the investigation."""
    print("\nINVESTIGATION: MODULE IMPORT ISSUES\n")
    
    # Run the debugging
    debug_module_loading()
    
    # Demonstrate the correct way to import
    demonstrate_correct_import()

# Run the investigation
if __name__ == "__main__":
    main()