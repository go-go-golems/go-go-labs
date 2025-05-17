# AGENT.md for Blender Python Project

## Build/Test Commands
- Interactive testing (requires user interaction): Open script in Blender's Text Editor and click Run Script
- Automated testing: 
  - send python code with execute_blender_code tool
  - load entire files (see below)

## Loading files / importing modules over blender MCP
- Preferred method: Add module path and use importlib.reload
```python
import sys, os, importlib
sys.path.append(os.getcwd())
import module_name
importlib.reload(module_name)  # For subsequent updates
```

## Code Style Guidelines
- Follow Blender 4.4 API conventions (avoid deprecated functions)
- Use proper docstrings with Args/Returns sections
- Import style: `import bpy` followed by module-specific imports
- Common aliases: `D = bpy.data`, `C = bpy.context`
- Error handling: Use try/except blocks and return tuple (success, message)
- Function names: Use snake_case
- Type annotations: Use docstring return types (bpy.types.X)

## Project Conventions
- Include context.area comments for workspace-specific scripts
- Document scripts with usage instructions in docstrings
- Use main() function pattern for script entry points
