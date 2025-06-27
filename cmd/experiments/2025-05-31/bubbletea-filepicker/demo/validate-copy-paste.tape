# VHS Script: Validate Copy & Paste Functionality
# Creates txt screenshots to validate copy/paste operations

# Setup for validation
Output demo/validate-copy-paste.txt
Set FontSize 14
Set Width 800
Set Height 600
Set Theme "Dracula"

# Launch
Type "./filepicker test-files"
Enter
Sleep 2s

# Navigate to documents/work to find files to copy
Down Down Down Down  # Navigate to documents
Enter
Sleep 1s
Down Down  # Skip .. and navigate to work
Enter
Sleep 1s

# Select multiple files for copy
Down  # Skip ..
Space  # Select first file
Sleep 500ms
Down
Space  # Select second file
Sleep 500ms

# Copy files (c key)
Type "c"
Sleep 1s

# Take screenshot to show copy operation
Screenshot demo/copy-operation.txt

# Navigate back to parent directory
Backspace
Sleep 500ms

# Navigate to personal directory to paste
Down  # Navigate to personal
Enter
Sleep 1s

# Paste files (v key)
Type "v"
Sleep 2s

# Take screenshot to show paste result
Screenshot demo/paste-result.txt

# Exit
Escape
