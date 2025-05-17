# Demo script for video flicker effect in Blender

import bpy
import os
import sys
import importlib

# Add scripts directory to path
scripts_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'scripts')
if scripts_dir not in sys.path:
    sys.path.append(scripts_dir)

# Add utils directory to path
utils_dir = os.path.join(scripts_dir, 'utils')
if utils_dir not in sys.path:
    sys.path.append(utils_dir)

# Add tests directory to path
tests_dir = os.path.join(scripts_dir, 'tests')
if tests_dir not in sys.path:
    sys.path.append(tests_dir)

# Make sure we're working with the sequencer
for area in bpy.context.screen.areas:
    if area.type == 'SEQUENCE_EDITOR':
        break
else:
    # If no sequencer is found, try to change one area to a sequencer
    if bpy.context.screen.areas:
        bpy.context.screen.areas[0].type = 'SEQUENCE_EDITOR'
    else:
        print("No areas available to change to Video Sequence Editor")

# Import our utility modules
try:
    # Clear out previous module imports to ensure fresh reload
    for module_name in ['vse_utils', 'vse_effects', 'test_flicker_effect']:
        if module_name in sys.modules:
            del sys.modules[module_name]
    
    # Import the test script
    sys.path.insert(0, tests_dir)  # Make sure it finds the test module first
    import test_flicker_effect
    importlib.reload(test_flicker_effect)
    
    # Run the test
    print("\n\n==== RUNNING FLICKER EFFECT DEMO ====\n")
    result = test_flicker_effect.main()
    
    if result:
        print("\n==== DEMO COMPLETED SUCCESSFULLY ====\n")
        print("The flicker effect has been created on channel 5.")
        print("Press Alt+A to play the animation and see the effect.")
    else:
        print("\n==== DEMO FAILED TO COMPLETE ====\n")
        print("Check the console for error messages.")
        
except Exception as e:
    print(f"Error running demo: {str(e)}")
    import traceback
    traceback.print_exc()