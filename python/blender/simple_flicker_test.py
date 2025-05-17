# Minimal example for running the flicker effect test

import bpy
import os
import sys

# Add necessary paths
project_dir = os.path.dirname(bpy.data.filepath) if bpy.data.filepath else '/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender'
scripts_dir = os.path.join(project_dir, 'scripts')
utils_dir = os.path.join(scripts_dir, 'utils')
tests_dir = os.path.join(scripts_dir, 'tests')

for path in [scripts_dir, utils_dir, tests_dir]:
    if path not in sys.path:
        sys.path.append(path)

# Import core modules with reload
import importlib

# First make sure we have a sequence editor context
for area in bpy.context.screen.areas:
    if area.type == 'SEQUENCE_EDITOR':
        break
else:
    if bpy.context.screen.areas:
        bpy.context.screen.areas[0].type = 'SEQUENCE_EDITOR'

# Import our modules
import vse_utils
importlib.reload(vse_utils)
import vse_effects
importlib.reload(vse_effects)

# Run the test
import test_flicker_effect
importlib.reload(test_flicker_effect)

print("\n==== RUNNING FLICKER EFFECT TEST ====\n")
result = test_flicker_effect.main()
print(f"\n==== TEST {'SUCCEEDED' if result else 'FAILED'} ====\n")