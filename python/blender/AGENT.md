# AGENT.md for Blender Python Project

## Build/Test Commands
- Interactive testing (requires user interaction): Open script in Blender's Text Editor and click Run Script
- Automated testing: 
  - send python code with execute_blender_code tool
  - load entire files (see below)

## File Naming & Organization
- Use underscores instead of hyphens in all Python filenames (e.g., `my_module.py`, not `my-module.py`)
- Module names must follow Python variable naming rules
- Organize code into subdirectories:
  - `utils/`: Reusable utilities and helper functions
  - `chapters/`: Chapter-specific implementation scripts
  - `tests/`: Test scripts and demo functions
  - `investigations/`: Exploratory analysis scripts

## Loading files / importing modules over blender MCP

### Using execute_blender_code Tool

For executing individual code blocks:

1. **Basic Scene Setup**:
```python
import bpy
import os
import sys

# Clear any existing scene data
scene = bpy.context.scene
seq_editor = scene.sequence_editor_create()
```

2. **Setting Up Import Paths**:
```python
# Set up absolute paths - NEVER use __file__
scripts_dir = '/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender/scripts'
utils_dir = os.path.join(scripts_dir, 'utils')
if utils_dir not in sys.path:
    sys.path.append(utils_dir)
```

3. **Importing Modules**:
```python
# Import utilities with error handling
try:
    from utils import vse_utils
    from utils import transition_utils
    print("Successfully imported utilities")
except ImportError as e:
    print(f"Warning: Could not import utilities: {e}")
```

4. **Running Functions**:
```python
# Always wrap function calls in try-except
try:
    result = vse_utils.some_function()
    print(f"Function executed successfully: {result}")
except Exception as e:
    print(f"Error executing function: {e}")
```

Key Points:
- NEVER use `__file__` or relative imports
- Always use absolute paths
- Break complex operations into smaller chunks
- Include proper error handling
- Print status messages for debugging

### Executing Specific Files

When asked to execute a specific file (e.g., "@some_file.py"), follow this pattern:

1. **Setup Environment**:
```python
import bpy
import os
import sys

# Set up absolute paths - NEVER use __file__
scripts_dir = '/home/manuel/code/wesen/corporate-headquarters/go-go-labs/python/blender/scripts'
utils_dir = os.path.join(scripts_dir, 'utils')
if utils_dir not in sys.path:
    sys.path.append(utils_dir)

# Add the directory containing the target file
target_dir = os.path.dirname(os.path.join(scripts_dir, 'path/to/target/file'))
if target_dir not in sys.path:
    sys.path.append(target_dir)
```

2. **Import and Execute**:
```python
# Import the module (assuming file is my_script.py)
import my_script
import importlib
importlib.reload(my_script)  # Ensure we have the latest version

# Execute the main function if it exists
if hasattr(my_script, 'main'):
    my_script.main()
```

Key Points:
- Always import the file as a module rather than executing it directly
- Use importlib.reload() to ensure you have the latest version
- Add the file's directory to sys.path before importing
- Use absolute paths throughout
- Handle imports and execution separately

## Code Style Guidelines
- Follow Blender 4.4 API conventions (avoid deprecated functions)
- Use proper docstrings with Args/Returns sections
- Import style: `import bpy` followed by module-specific imports
- Common aliases: `D = bpy.data`, `C = bpy.context`
- Error handling: Use try/except blocks and return tuple (success, message)
- Function names: Use snake_case
- Type annotations: Use docstring return types (bpy.types.X)

## Handling bpy Module in Development
- The `bpy` module only exists within Blender's Python environment
- For VS Code/IDE development, you may see "Import 'bpy' could not be resolved" warnings
- Options for handling this:
  - Add `# type: ignore` comment after bpy imports to suppress warnings
  - For functions using Blender types: `def my_func(obj): # type: ignore`
  - Create type stubs for bpy with `stub-generator` (advanced)
  - Ignore these warnings during development as they won't affect runtime in Blender

## Project Conventions
- Include context.area comments for workspace-specific scripts
- Document scripts with usage instructions in docstrings
- Use main() function pattern for script entry points
- Scripts are not automatically executed with `__name__ == "__main__"` - they must be triggered explicitly

## Utility Reference
- For details on available VSE utilities, see `python-blender-utilities-api.md`
