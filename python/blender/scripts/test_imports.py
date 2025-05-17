# Sample script to test importing our utility modules

import sys
import os
import importlib

print("\nBlender VSE Python Utilities Import Test")
print("========================================")

# Add script directories to path
scripts_dir = os.path.dirname(os.path.abspath(__file__))
utils_dir = os.path.join(scripts_dir, 'utils')

print(f"Scripts directory: {scripts_dir}")
print(f"Utils directory: {utils_dir}")

# Add to sys.path if not already there
for path in [scripts_dir, utils_dir]:
    if path not in sys.path:
        sys.path.append(path)
        print(f"Added to sys.path: {path}")

# Import and reload utility modules to ensure changes are reflected
try:
    import vse_utils
    importlib.reload(vse_utils)
    print("✓ Successfully imported vse_utils")
    
    # Test a function from the module
    print(f"  Available functions: {', '.join([f for f in dir(vse_utils) if not f.startswith('_') and callable(getattr(vse_utils, f))][:5])}...")
    
    # Import and test transition_utils
    import transition_utils
    importlib.reload(transition_utils)
    print("✓ Successfully imported transition_utils")
    
    # Import and test effect_utils
    import effect_utils
    importlib.reload(effect_utils)
    print("✓ Successfully imported effect_utils")
    
    print("\nAll utility modules imported successfully!")
    print("Run this script whenever you need to reload the modules after making changes.")
    
except ImportError as e:
    print(f"❌ Error importing modules: {e}")
    print("\nTroubleshooting:")
    print("1. Make sure the utility scripts exist in the correct location")
    print("2. Check that the scripts directory is in your sys.path")
    print("3. Check for syntax errors in the utility modules")
    
    print("\nCurrent sys.path:")
    for i, path in enumerate(sys.path):
        print(f"  {i}: {path}")