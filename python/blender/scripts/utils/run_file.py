# Utility for running Python files in Blender
# This is a helper for running scripts from Blender's Text Editor or the Python console

import bpy # type: ignore
import os

def run_file(filepath, in_sequencer=False):
    """
    Run any Python file in Blender.
    
    Args:
        filepath (str): Path to the Python file to run
        in_sequencer (bool): Set to True if the script should run in the Sequencer context
    
    Returns:
        bool: Success status
    """
    if not os.path.exists(filepath):
        print(f"Error: File not found: {filepath}")
        return False
    
    original_area_type = None
    
    # If in_sequencer is True, try to switch to Sequencer context
    if in_sequencer:
        # Store current area type for restoration
        if bpy.context.area:
            original_area_type = bpy.context.area.type
            try:
                bpy.context.area.type = 'SEQUENCE_EDITOR'
                print(f"Switched context to SEQUENCE_EDITOR for script execution")
            except Exception as e:
                print(f"Warning: Could not switch to SEQUENCE_EDITOR: {e}")
    
    try:
        # Execute the file contents
        with open(filepath, 'r') as f:
            code = f.read()
        
        print(f"Running file: {filepath}")
        exec(code)
        print(f"Completed execution of {os.path.basename(filepath)}")
        success = True
    except Exception as e:
        print(f"Error executing file: {e}")
        success = False
    
    # Restore original area type if changed
    if original_area_type:
        try:
            bpy.context.area.type = original_area_type
            print(f"Restored context to {original_area_type}")
        except Exception as e:
            print(f"Warning: Could not restore original context: {e}")
    
    return success

# Example usage
if __name__ == "__main__":
    # This can be used directly in Blender's Python console or Text Editor
    script_path = "/path/to/your/script.py"  # Change this to your script path
    run_file(script_path, in_sequencer=True)