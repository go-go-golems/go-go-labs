# VHS Script: Test Copy/Paste with Screenshot Validation
# This script tests copy/paste functionality and takes screenshots

# Setup 
Output demo/test-copy-paste.txt
Set FontSize 14
Set Width 1000
Set Height 700
Set Theme "Dracula"

# Launch the file picker
Type "./filepicker test-files"
Enter
Sleep 2s

# Take initial screenshot
Screenshot demo/initial-state.txt

# Navigate to documents directory
Down Down Down
Enter
Sleep 1s

# Navigate to work directory
Down Down
Enter
Sleep 1s

# Select files to copy
Down  # Skip ..
Space  # Select first file
Sleep 300ms
Down
Space  # Select second file
Sleep 300ms

# Take screenshot showing selected files
Screenshot demo/files-selected.txt

# Copy files
Type "c"
Sleep 500ms

# Navigate back to documents
Backspace
Sleep 500ms

# Navigate to personal directory
Down
Enter
Sleep 1s

# Paste files
Type "v"
Sleep 1s

# Take screenshot showing pasted files
Screenshot demo/files-pasted.txt

# Exit
Escape
