# Simple script to run transition utility tests
import sys
import os
import importlib

# Add script directories to path
scripts_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
utils_dir = os.path.join(scripts_dir, 'utils')
tests_dir = os.path.dirname(os.path.abspath(__file__))

for path in [scripts_dir, utils_dir, tests_dir]:
    if path not in sys.path:
        sys.path.append(path)

# Import and reload the test module
from tests import test_transition_utils
importlib.reload(test_transition_utils)

# Function to run a specific test
def run_test(test_name):
    test_transition_utils.run_file(test_name)

# If run directly, show available tests
if __name__ == "__main__":
    print("\nAvailable tests:")
    print("  - crossfade")
    print("  - gamma_crossfade")
    print("  - wipe")
    print("  - audio_fade")
    print("  - audio_crossfade")
    print("  - fade_to_color")
    print("  - all (runs all tests)\n")
    print("Run a test using: run_test('test_name')")