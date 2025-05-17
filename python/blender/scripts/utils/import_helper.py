# Helper functions for importing modules in Blender scripts

import bpy
import os
import sys
import importlib

def setup_paths():
    """Set up all necessary paths for imports."""
    # Get script directory (works in both script editor and execute_blender_code)
    project_dir = os.path.dirname(bpy.data.filepath) if bpy.data.filepath else os.path.dirname(__file__)
    
    # Navigate to project root if we're in a subdirectory
    if os.path.basename(project_dir) == 'utils':
        project_dir = os.path.dirname(os.path.dirname(project_dir))
    
    # Define key paths
    scripts_dir = os.path.join(project_dir, 'scripts')
    utils_dir = os.path.join(scripts_dir, 'utils')
    tests_dir = os.path.join(scripts_dir, 'tests')
    
    # Add paths if needed
    for path in [scripts_dir, utils_dir, tests_dir]:
        if path not in sys.path:
            sys.path.append(path)
    
    return project_dir, scripts_dir, utils_dir, tests_dir

def import_module(module_name, force_reload=True):
    """Import a module with optional force reload."""
    # Import the module
    module = __import__(module_name)
    
    # Force reload if requested
    if force_reload:
        importlib.reload(module)
    
    return module

def run_script(script_path):
    """Run a script from a file path."""
    # Make sure paths are set up
    setup_paths()
    
    # Get the directory and filename
    script_dir = os.path.dirname(script_path)
    script_name = os.path.splitext(os.path.basename(script_path))[0]
    
    # Add the script directory to path if not already there
    if script_dir not in sys.path:
        sys.path.append(script_dir)
    
    # Clear the module if it's already loaded
    if script_name in sys.modules:
        del sys.modules[script_name]
    
    # Import and run the script
    script_module = __import__(script_name)
    importlib.reload(script_module)
    
    # Run main() function if it exists
    if hasattr(script_module, 'main'):
        return script_module.main()
    
    return script_module