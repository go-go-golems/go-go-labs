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
- Preferred method: Add module path and use importlib.reload
```python
import sys, os, importlib

# Add the main scripts directory
scripts_dir = os.path.dirname(os.path.abspath(__file__))
if scripts_dir not in sys.path:
    sys.path.append(scripts_dir)

# For importing from subdirectories
utils_dir = os.path.join(scripts_dir, 'utils')
if utils_dir not in sys.path:
    sys.path.append(utils_dir)

# Now import and reload for development
import module_name
importlib.reload(module_name)  # For subsequent updates

# For subdirectory imports
from utils import vse_utils
importlib.reload(vse_utils)
```

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
