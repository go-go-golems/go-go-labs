# Testing Guidelines for Blender Python VSE Utilities

## Step-by-Step Testing Process

1. **Preparation**: Read the file being tested to understand each function's purpose and parameters. Don't modify the file yet.

2. **Test File Creation**: Create a dedicated test file in the tests/ directory with functions that test each feature independently:
   - Each test function should focus on a single functionality
   - Include setup/teardown that completely clears the scene
   - Test basic, edge case, and error scenarios 

3. **Test Scene Management**:
   - Create a robust setup function that:
     - Completely clears any existing sequence editor (`scene.sequence_editor_clear()`)
     - Creates a fresh sequence editor 
     - Sets up necessary test media
   - This function should be called at the start of EVERY test

4. **Incremental Testing**: Test one function at a time via Blender MCP:
   ```python
   # Clear any existing scene data
   import bpy
   scene = bpy.context.scene
   if scene.sequence_editor:
       scene.sequence_editor_clear()
   seq_editor = scene.sequence_editor_create()
   
   # Import utility modules with absolute paths
   import sys, os
   scripts_dir = '/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender/scripts'
   utils_dir = os.path.join(scripts_dir, 'utils')
   if utils_dir not in sys.path: sys.path.append(utils_dir)
   
   # Import and test specific function
   from utils import my_utils_module
   
   # Run the specific test
   # Call test function directly, with proper error handling
   ```

5. **Validation & Iteration**:
   - After each test, inspect the scene to verify expected changes
   - Print detailed scene state before and after operations using vse_utils.print_sequence_info()
   - Fix issues in the utility file and repeat testing until all functions work
   - Ensure all parameter values are properly converted to expected types (int for frames, etc.)

6. **Documentation**: Once tests pass, document all required fixes in a `$FILENAME-things-I-had-to-fix.md` file, including:
   - API compatibility issues
   - Type conversion requirements
   - Error handling improvements
   - Logic corrections

7. **Progress Tracking**: Only move to the next utility file after thoroughly testing all functions in the current file.

## Special Considerations for Blender MCP

- Never rely on `__file__` or relative imports - use absolute paths
- Use integer values for all frame positions and durations
- Clear scene state completely between tests 
- Check for proper data types in function parameters
- Implement robust strip tracking when operations modify strips (use as_pointer())
- For animation data, check scene.animation_data rather than strip.animation_data
- Be aware that strip.type values are strings like 'MOVIE', 'SOUND', 'CROSS', etc.

## Common Issues & Solutions

### Float vs. Integer Type Errors
Blender's API strictly requires integers for frame positions. Always convert frames to integers:
```python
frame_start = int(strip.frame_start)
```

### Scene Cleanup
Always clear the scene completely between tests:
```python
def clear_scene():
    scene = bpy.context.scene
    if scene.sequence_editor:
        scene.sequence_editor_clear()  # Completely removes and recreates
    return scene.sequence_editor_create()
```

### Animation Data Access
Strip animation data (like volume keyframes) is stored in the scene, not on the strip:
```python
# Wrong approach
if strip.animation_data:  # Will likely be None
    # ...

# Correct approach
if scene.animation_data and scene.animation_data.action:
    for fc in scene.animation_data.action.fcurves:
        if fc.data_path.startswith('sequence_editor.sequences_all[') and \
           fc.data_path.endswith('].volume'):
            # Check if this fcurve belongs to our strip
            if strip.name in fc.data_path:
                # Found volume keyframe for this strip
```

### Strip Tracking After Operations
Track strips across operations using pointers:
```python
# Before operation
pre_op_ids = {s.as_pointer() for s in seq_editor.strips_all}

# After operation - find new strips
new_strips = [s for s in seq_editor.strips_all if s.as_pointer() not in pre_op_ids]
```

By following these guidelines, testing will be more methodical, reliable, and less prone to state pollution issues.