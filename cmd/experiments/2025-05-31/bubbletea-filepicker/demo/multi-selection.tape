# VHS Script: Multi-Selection Demo
# Demonstration of multi-selection with Space key

# Setup
Output demo/multi-selection.gif
Set FontSize 14
Set Width 800
Set Height 600
Set Theme "Dracula"

# Launch
Type "./filepicker test-files"
Enter
Sleep 1s

# Navigate to a directory with files
Down Down Down  # Navigate to documents
Enter
Sleep 300ms
Down Down  # Navigate to work
Enter
Sleep 300ms

# Demonstrate multi-selection
Down  # Skip .. and navigate to first file
Space  # Select first file (✓)
Sleep 400ms
Down
Space  # Select second file (✓)
Sleep 400ms
Down
Space  # Select third file (✓)
Sleep 800ms

# Show deselection
Up
Space  # Deselect middle file
Sleep 500ms

# Show select all files
Type "ctrl+a"
Sleep 800ms

# Show deselect all
Type "A"
Sleep 500ms

# Exit
Escape
