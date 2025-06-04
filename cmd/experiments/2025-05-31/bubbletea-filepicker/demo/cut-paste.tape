# VHS Script: Cut & Paste Demo
# Demonstration of file cut (move) and paste operations

# Setup
Output demo/cut-paste.gif
Set FontSize 14
Set Width 800
Set Height 600
Set Theme "Dracula"

# Launch
Type "./filepicker test-files"
Enter
Sleep 1s

# Navigate to source directory
Down Down Down
Enter
Sleep 300ms
Down Down
Enter
Sleep 300ms

# Select files to cut (move)
Down Down  # Navigate to files
Space  # Select file 1
Sleep 300ms
Down
Space  # Select file 2
Sleep 500ms

# Cut files (x)
Type "x"
Sleep 400ms

# Navigate to destination
Backspace
Sleep 300ms
Down
Enter
Sleep 300ms

# Paste files (they should move, not copy)
Type "v"
Sleep 1s

# Show files were moved
Down Down
Sleep 500ms

# Go back to source to show files are gone
Backspace
Sleep 300ms
Down Down
Enter
Sleep 500ms

# Exit
Escape
