# VHS Script: Copy & Paste Demo
# Demonstration of file copy and paste operations

# Setup
Output demo/copy-paste.gif
Set FontSize 14
Set Width 900
Set Height 600
Set Theme "Dracula"

# Launch
Type "./filepicker test-files"
Enter
Sleep 1s

# Navigate to documents/work
Down Down Down
Enter
Sleep 300ms
Down Down
Enter
Sleep 300ms

# Select files to copy
Down  # Skip ..
Space  # Select file 1
Sleep 300ms
Down
Space  # Select file 2
Sleep 500ms

# Copy files
Type "c"
Sleep 400ms

# Navigate to personal directory
Backspace
Sleep 300ms
Down
Enter
Sleep 300ms

# Paste files
Type "v"
Sleep 1s

# Show files were copied (navigate through them)
Down Down Down
Sleep 800ms

# Exit
Escape
