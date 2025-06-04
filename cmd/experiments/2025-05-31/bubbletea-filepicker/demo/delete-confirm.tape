# VHS Script: Delete Confirmation Demo
# Demonstration of delete operation with confirmation dialog

# Setup
Output demo/delete-confirm.gif
Set FontSize 14
Set Width 800
Set Height 600
Set Theme "Dracula"

# Launch
Type "./filepicker test-files"
Enter
Sleep 1s

# Navigate to a directory with files
Down Down Down
Enter
Sleep 300ms
Down Down
Enter
Sleep 300ms

# Select files to delete
Down Down  # Navigate to files
Space  # Select file 1
Sleep 300ms
Down
Space  # Select file 2
Sleep 500ms

# Initiate delete
Type "d"
Sleep 1s

# Show confirmation dialog, then cancel
Type "n"
Sleep 500ms

# Try delete again but confirm this time
Down  # Navigate to another file
Type "d"
Sleep 800ms

# Confirm deletion
Type "y"
Sleep 800ms

# Show file is gone
Sleep 500ms

# Exit
Escape
