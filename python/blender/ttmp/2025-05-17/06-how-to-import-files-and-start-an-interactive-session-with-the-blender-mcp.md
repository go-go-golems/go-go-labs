# How to Import Files and Start an Interactive Session with the Blender MCP

This document provides comprehensive guidance for working with Python in Blender, focusing on module imports and interactive development workflows.

## Import Patterns

### Basic Import Pattern

The standard pattern for importing modules in Blender scripts involves setting up the Python path and using reload for development:

```python
import sys, os, importlib

# Add scripts directory to path
scripts_dir = os.path.dirname(os.path.abspath(__file__))
if scripts_dir not in sys.path:
    sys.path.append(scripts_dir)

# Import and reload modules
import my_module
importlib.reload(my_module)  # For development - refreshes code changes
```

### Handling Path Issues

When using tools like `execute_blender_code`, `__file__` might not be defined. Use this pattern instead:

```python
import bpy, os, sys

# Handle path with or without __file__
try:
    scripts_dir = os.path.dirname(os.path.abspath(__file__))
except NameError:
    # Fallback when __file__ is not defined
    project_dir = os.path.dirname(bpy.data.filepath) if bpy.data.filepath else '/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender'
    scripts_dir = os.path.join(project_dir, 'scripts')
```

## Quick Execution Methods

### Direct File Execution

The fastest way to run a script file with `execute_blender_code`:

```python
# Single line execution
exec(open('/absolute/path/to/script.py').read())

# With path building
import os
proj_dir = '/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender'
exec(open(os.path.join(proj_dir, 'my_script.py')).read())
```

### Using Helper Utilities

The project includes utility modules to simplify execution:

#### With import_helper.py

```python
import sys, os
utils_dir = '/path/to/project/scripts/utils'
if utils_dir not in sys.path:
    sys.path.append(utils_dir)

import import_helper
import_helper.setup_paths()

# Now import any project module
import vse_utils
```

#### With run_file.py

```python
import sys, os
utils_dir = '/path/to/project/scripts/utils'
if utils_dir not in sys.path:
    sys.path.append(utils_dir)

from run_file import run_file

# Run a script with correct context
run_file('/path/to/script.py', in_sequencer=True)
```

## Setting Up the Editor Context

Many Blender operations require the correct editor context:

```python
# Switch to Sequence Editor context
for area in bpy.context.screen.areas:
    if area.type == 'SEQUENCE_EDITOR':
        break
else:
    if bpy.context.screen.areas:
        bpy.context.screen.areas[0].type = 'SEQUENCE_EDITOR'
        print("Switched to Sequence Editor context")
```

## Bootstrap Pattern

For quick project setup, create a `bootstrap.py` file:

```python
# bootstrap.py
import bpy, os, sys, importlib

def setup_environment():
    """Set up project environment and paths"""
    # Get project directory
    project_dir = os.path.dirname(bpy.data.filepath) if bpy.data.filepath else '/path/to/project'
    scripts_dir = os.path.join(project_dir, 'scripts')
    utils_dir = os.path.join(scripts_dir, 'utils')
    
    # Add to path
    for path in [scripts_dir, utils_dir]:
        if path not in sys.path:
            sys.path.append(path)
    
    # Switch to appropriate context if needed
    for area in bpy.context.screen.areas:
        if area.type == 'SEQUENCE_EDITOR':
            break
    else:
        if bpy.context.screen.areas:
            bpy.context.screen.areas[0].type = 'SEQUENCE_EDITOR'
    
    return project_dir

# Usage: from bootstrap import setup_environment
```

## Interactive Development Workflow

1. **Setup Phase**:
   - Create bootstrap script with environment setup
   - Ensure proper module paths are configured

2. **Development Loop**:
   - Edit code in external editor
   - Use `execute_blender_code` to run updated code
   - Or use `exec(open())` pattern for quick file execution
   - Use `importlib.reload()` to refresh modules

3. **Testing**:
   - Create test scripts in `scripts/tests/` directory
   - Run with context-aware utilities like `run_file.py`
   - Use print statements for debugging

## Common Issues & Solutions

### Path Issues

- **Problem**: Module imports fail with `ModuleNotFoundError`
- **Solution**: Ensure proper paths are added to `sys.path`

### Context Issues

- **Problem**: Blender operations fail with "context is incorrect"
- **Solution**: Use context setup code to switch to the appropriate editor type

### Module Reload Issues

- **Problem**: Code changes not reflected after edits
- **Solution**: Use `importlib.reload(module)` after imports

## Utilities Reference

### import_helper.py

Handle module imports and path setup:

```python
# Key functions
setup_paths()      # Configure all project paths
import_module()    # Import and reload modules
```

### run_file.py

Execute Python files with proper context:

```python
# Key function
run_file(filepath, in_sequencer=False)  # Run a file with optional context setup
```

### Simple Execution

For the most compact testing approach:

```python
# In execute_blender_code
import os
proj_dir = '/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender'
exec(open(os.path.join(proj_dir, 'simple_flicker_test.py')).read())
```