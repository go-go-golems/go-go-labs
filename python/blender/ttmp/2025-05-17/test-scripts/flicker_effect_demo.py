# Demo script for video flicker effect in Blender

import bpy
import os
import sys

# Bootstrap the import helper
scripts_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'scripts')
utils_dir = os.path.join(scripts_dir, 'utils')

if utils_dir not in sys.path:
    sys.path.append(utils_dir)

# Use the import helper
import import_helper
import_helper.setup_paths()

# Make sure we're in the sequencer
for area in bpy.context.screen.areas:
    if area.type == 'SEQUENCE_EDITOR':
        break
else:
    if bpy.context.screen.areas:
        bpy.context.screen.areas[0].type = 'SEQUENCE_EDITOR'

# Run the test script
print("\n\n==== RUNNING FLICKER EFFECT DEMO ====\n")
tests_dir = os.path.join(scripts_dir, 'tests')
result = import_helper.run_script(os.path.join(tests_dir, 'test_flicker_effect.py'))

if result:
    print("\n==== DEMO COMPLETED SUCCESSFULLY ====\n")
    print("The flicker effect has been created on channel 5.")
    print("Press Alt+A to play the animation and see the effect.")
else:
    print("\n==== DEMO FAILED TO COMPLETE ====\n")