# Test Runner for Chapter 3 Demonstrations

"""
This script runs individual or all demonstrations from Chapter 3 of the
Blender VSE Python API guide, allowing for isolated testing of specific
functionality.

Usage in Blender:
1. Open Blender Text Editor
2. Load and run this script
3. Choose a specific test in the UI or run all tests

Or run programmatically with:
    import chapter_3_test_runner
    chapter_3_test_runner.run_test('trim')  # Or 'split', 'remove', 'slip', 'move', 'all'
"""

import bpy
import os
import sys
import importlib

# Add script directory to path if needed
script_dir = os.path.dirname(os.path.abspath(__file__))
if script_dir not in sys.path:
    sys.path.append(script_dir)

# Import our modules with reload for development
import vse_utils
importlib.reload(vse_utils)

import chapter_3_demos
importlib.reload(chapter_3_demos)

# Mapping of test types to functions
test_functions = {
    'trim': chapter_3_demos.demonstrate_trimming,
    'split': chapter_3_demos.demonstrate_splitting,
    'remove': chapter_3_demos.demonstrate_segment_removal,
    'slip': chapter_3_demos.demonstrate_slip_edit,
    'move': chapter_3_demos.demonstrate_moving_strip
}

def run_test(test_type='all'):
    """
    Run a specific test or all tests.
    
    Args:
        test_type (str): One of 'trim', 'split', 'remove', 'slip', 'move', or 'all'
    
    Returns:
        bool: True if test(s) completed, False if invalid test type
    """
    # Validate test type
    if test_type != 'all' and test_type not in test_functions:
        print(f"Error: Unknown test type '{test_type}'. " +
              f"Valid types are: {', '.join(list(test_functions.keys()) + ['all'])}")
        return False
    
    # Find test media directory once to use for all tests
    test_media_dir = vse_utils.find_test_media_dir()
    
    print("\n" + "=" * 60)
    print("BLENDER VSE PYTHON API TESTS - CHAPTER 3")
    print("=" * 60)
    
    # Run a single test
    if test_type != 'all':
        return test_functions[test_type](test_media_dir)
    
    # Run all tests
    print("\nRunning all Chapter 3 tests\n")
    
    # Store current scene state
    original_scene = bpy.context.scene
    test_results = {}
    
    # Create a separate scene for each test
    for test_name, test_func in test_functions.items():
        # Create a new scene for this test
        test_scene_name = f"Test_{test_name}"
        test_scene = bpy.data.scenes.new(test_scene_name)
        
        # Switch to this scene
        bpy.context.window.scene = test_scene
        
        # Run the test
        print(f"\n{'=' * 20} Running test: {test_name} {'=' * 20}")
        try:
            result = test_func(test_media_dir)
            test_results[test_name] = result
        except Exception as e:
            print(f"Error in {test_name} test: {e}")
            test_results[test_name] = False
    
    # Switch back to the original scene
    bpy.context.window.scene = original_scene
    
    # Print summary
    print("\n" + "=" * 60)
    print("TEST RESULTS SUMMARY")
    print("=" * 60)
    
    for test_name, result in test_results.items():
        status = "PASSED" if result else "FAILED"
        print(f"{test_name.ljust(10)}: {status}")
    
    # Return True if all tests passed
    return all(test_results.values())

def main():
    """
    Main function for running the script directly.
    
    Checks command line arguments to determine which test to run.
    If no arguments provided, runs all tests.
    """
    test_to_run = 'all'
    
    # Try to get test type from command line (not available in Blender Text Editor)
    try:
        if len(sys.argv) > 4:  # In Blender, first args are blender, script path, etc.
            test_to_run = sys.argv[4]
    except:
        # Not running from command line or couldn't parse args
        pass
    
    # Run the specified test
    run_test(test_to_run)

# Run the script if executed directly
if __name__ == "__main__":
    main()