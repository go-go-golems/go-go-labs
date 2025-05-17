# Importing and Working with Blender Python Code through MCP

This document provides a technical overview of how to elegantly import Python code modules into Blender through the Multi-Channel Protocol (MCP) interface. We'll focus specifically on implementing and using the sequence editor import function as an example.

## Overview

When using Blender through MCP in external editor environments, repeated code execution becomes inefficient, especially for complex operations. Rather than sending large blocks of Python code directly, a more maintainable approach is to:

1. Create Python modules with well-defined functions
2. Load these modules into Blender
3. Call specific functions as needed

This pattern aligns with software engineering best practices and significantly improves code maintainability.

## Bootstrapping: First-Time Import

Before you can import your module, you need to initially load it into Blender. There are two primary approaches to bootstrapping your Python module imports:

### Method 1: Direct Execution

The simplest approach uses Python's `exec()` function to execute the module directly:

```python
import bpy
import os

# Get the current file path
file_path = os.path.join(os.getcwd(), 'test.py')
print(f"Loading script from: {file_path}")

# Execute the external script
exec(compile(open(file_path).read(), file_path, 'exec'))
```

This method:
1. Locates your script file in the current working directory
2. Compiles it into executable code
3. Runs it directly in the current Python environment

**Advantages:**
- Simple one-liner approach
- No need to modify sys.path
- Works well for initial loading of simple scripts

**Disadvantages:**
- Executes the entire script, including any main code block
- Doesn't provide clean import semantics (no module namespace)
- Limited error tracebacks

### Method 2: Module Path Setup (Preferred)

For more robust and maintainable code, it's better to set up proper module importing:

```python
import sys
import os
import importlib

# Add current directory to Python path
sys.path.append(os.getcwd())

# First time import
import test

# Subsequent imports should use reload
importlib.reload(test)
```

This method:
1. Adds your current directory to Python's module search path
2. Imports your module normally using Python's import system
3. Uses importlib.reload() for subsequent updates

**When to use which method:**
- Use direct execution (Method 1) for one-off scripts or initial bootstrapping
- Use proper importing (Method 2) for ongoing development and modular code

## Implementation Details

### Creating Importable Modules

Our example (`test.py` and `test2.py`) demonstrates how to structure Python modules for Blender:

1. **Standard Python modules** with `import bpy` and other dependencies
2. **Well-defined functions** with clear interfaces and docstrings
3. **Conditional execution** using `if __name__ == "__main__"` for direct execution vs import

### The Import Mechanism

Here's how you implement a simple file import function in your module:

```python
import bpy
import os

def import_file(filepath, channel=1, start_frame=1, name=None):
    """
    Import a file into the sequence editor.
    
    Args:
        filepath (str): The absolute path to the file
        channel (int, optional): The channel to place the strip on. Defaults to 1.
        start_frame (int, optional): The frame to start the strip at. Defaults to 1.
        name (str, optional): Custom name for the strip. Defaults to None (uses filename).
        
    Returns:
        The newly created strip object or None if failed
    """
    if not os.path.exists(filepath):
        print(f"Error: File {filepath} not found")
        return None
        
    scene = bpy.context.scene
    
    # Ensure sequence editor exists
    if not scene.sequence_editor:
        scene.sequence_editor_create()
    
    # Get file extension
    _, ext = os.path.splitext(filepath)
    ext = ext.lower()
    
    # Add the appropriate strip type
    if ext in ['.mp4', '.avi', '.mov', '.mkv', '.flv', '.webm']:
        strip = scene.sequence_editor.sequences.new_movie(
            name=name or os.path.basename(filepath),
            filepath=filepath,
            channel=channel,
            frame_start=start_frame
        )
        print(f"Added movie strip: {strip.name}")
    elif ext in ['.mp3', '.wav', '.ogg', '.flac']:
        strip = scene.sequence_editor.sequences.new_sound(
            name=name or os.path.basename(filepath),
            filepath=filepath,
            channel=channel,
            frame_start=start_frame
        )
        print(f"Added sound strip: {strip.name}")
    elif ext in ['.png', '.jpg', '.jpeg', '.tiff', '.bmp']:
        strip = scene.sequence_editor.sequences.new_image(
            name=name or os.path.basename(filepath),
            filepath=filepath,
            channel=channel,
            frame_start=start_frame
        )
        print(f"Added image strip: {strip.name}")
    else:
        print(f"Unsupported file type: {ext}")
        return None
        
    return strip
```

## How to Use This Function

### Step 1: Import the module

```python
import sys
import os

# Add current directory to Python path
sys.path.append(os.getcwd())

# Import our test module
import importlib
import test  # Or test2 for the version with funny names
importlib.reload(test)  # Reload to get any changes
```

### Step 2: Call the function

```python
# Basic usage
new_strip = test.import_file("/path/to/video.mp4")

# Advanced usage with all parameters
new_strip = test.import_file(
    filepath="/path/to/audio.wav",
    channel=3,
    start_frame=25,
    name="Custom Strip Name"
)

# Using the funny version from test2.py
new_strip = test2.clip_teleporter_5000("/path/to/image.jpg")
```

## Technical Deep Dive: Blender's Import Mechanics

### Python Module System Integration

Blender's Python interpreter is a standard CPython environment with a few special considerations:

1. **Module Path Management**: When we use `sys.path.append(os.getcwd())`, we're extending Blender's Python module search path to include the current working directory. This is necessary because Blender's default module search paths don't typically include your project directory.

2. **Module Reloading**: The `importlib.reload(test)` call ensures that if we've made changes to the module after initially importing it, those changes are available to the current session. Without this reload step, Python would use the cached version of the module.

### Blender Data Interactions

The import function interacts with several key parts of Blender's architecture:

1. **Data-Block System**: When adding strips, we're creating new data within Blender's scene data-blocks. Specifically, we're adding sequences to the `scene.sequence_editor.sequences` collection.

2. **Operator-Free Approach**: This implementation uses direct data manipulation rather than operators (`bpy.ops`). This approach is more reliable for scripting because:
   - It doesn't depend on context (which can be problematic in background processing)
   - It provides more explicit control over the creation process
   - It returns direct references to the created objects

3. **Custom Property Handling**: The function employs parameter defaults and type checking to ensure reasonable behavior when parameters are omitted.

### Sequence Editor Under the Hood

The `import_file` function interfaces with Blender's sequence editor at a relatively low level:

1. **Creating the Editor**: `scene.sequence_editor_create()` ensures the sequence editor data structure exists. This is equivalent to opening the sequence editor for the first time in the UI.

2. **Direct Sequence Creation**: Methods like `sequences.new_movie()` create the appropriate strip type based on file extension. These methods map directly to C functions in Blender's core that create the underlying sequence data structures.

3. **Sequence Types**: Blender's sequence editor distinguishes between several strip types:
   - `MOVIE`: Video files
   - `SOUND`: Audio files
   - `IMAGE`: Single images
   - `META`, `SCENE`, `ADJUSTMENT`, etc.: Other special types

## Alternative Approaches and Extensions

### Funny Function Names

The `test2.py` module demonstrates how to make the interface more entertaining with appropriately named functions and entertaining output messages:

```python
def clip_teleporter_5000(filepath, channel=1, start_frame=1, name=None):
    """Teleport a media file into the sequence editor universe."""
    # Implementation similar to import_file but with fun messages
    # ...
```

### Command Batching

For more complex operations, you might want to batch multiple commands:

```python
def import_and_transition(filepath1, filepath2, transition_frames=10):
    """Import two files and create a transition between them."""
    strip1 = import_file(filepath1, channel=1)
    strip2 = import_file(filepath2, channel=2, start_frame=strip1.frame_final_end)
    
    # Create a transition
    effect = scene.sequence_editor.sequences.new_effect(
        name=f"Transition_{strip1.name}_{strip2.name}",
        type="CROSS",
        channel=3,
        frame_start=strip1.frame_final_end - transition_frames,
        frame_end=strip2.frame_start + transition_frames
    )
    effect.seq1 = strip1
    effect.seq2 = strip2
    
    return effect
```

## Best Practices

1. **Error Handling**: Always check if files exist and handle errors gracefully
2. **Parameter Validation**: Validate inputs before performing operations
3. **Documentation**: Include docstrings that explain parameters and return values
4. **Functional Design**: Keep functions focused on a single responsibility
5. **Return Values**: Return useful objects that can be used in further operations
6. **Module Reloading**: Always use importlib.reload() when iteratively developing

## Conclusion

Importing and working with Python code in Blender through MCP becomes much more manageable when you organize your code into well-structured modules. This approach offers several benefits:

- **Reduced redundancy**: Write code once, import everywhere
- **Better organization**: Separate functionality into logical modules
- **Easier maintenance**: Fix bugs or add features in one place
- **Improved development workflow**: Test incrementally without reloading everything

The sequence editor import function demonstrates these principles by providing a clean, reusable interface to a common operation, while hiding the complexity of file type detection and proper sequence editor initialization. 