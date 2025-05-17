# Simple script to run test files with minimal imports

import bpy
import os
import sys

# Get the scripts directory
try:
    scripts_dir = os.path.dirname(os.path.abspath(__file__))
except NameError:
    # __file__ may not be defined when executing directly via execute_blender_code
    scripts_dir = os.path.join(os.path.dirname(bpy.data.filepath) if bpy.data.filepath else '/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender', 'scripts')

# Add utils directory to path
utils_dir = os.path.join(scripts_dir, 'utils')
if utils_dir not in sys.path:
    sys.path.append(utils_dir)

# Import run_file utility
from run_file import run_file

def run_test(test_name):
    """Run a test script by name."""
    tests_dir = os.path.join(scripts_dir, 'tests')
    test_path = os.path.join(tests_dir, f"{test_name}.py")
    
    print(f"\n==== RUNNING TEST: {test_name} ====\n")
    result = run_file(test_path, in_sequencer=True)
    
    if result:
        print(f"\n==== TEST COMPLETED SUCCESSFULLY: {test_name} ====\n")
    else:
        print(f"\n==== TEST FAILED: {test_name} ====\n")
    
    return result

# Example usage with execute_blender_code:
# from scripts.run_test import run_test
# run_test('test_flicker_effect')