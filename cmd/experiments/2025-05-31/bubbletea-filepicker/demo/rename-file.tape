# VHS Script: Rename Operation Demo
# Demonstration of renaming files and directories

# Setup
Output demo/rename-file.gif
Set FontSize 14
Set Width 800
Set Height 600
Set Theme "Dracula"

# Launch
Type "./filepicker test-files"
Enter
Sleep 1s

# Navigate to a file
Down Down Down
Enter
Sleep 300ms
Down Down
Enter
Sleep 300ms
Down Down  # Navigate to a file

# Start rename operation
Type "r"
Sleep 500ms

# Clear current name and type new name
Ctrl+A
Type "renamed-document.txt"
Enter
Sleep 800ms

# Show the file was renamed (it should appear in new position due to sorting)
Down Down
Sleep 500ms

# Exit
Escape
