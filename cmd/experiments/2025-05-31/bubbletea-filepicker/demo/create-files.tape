# VHS Script: File Creation Demo
# Demonstration of creating new files and directories

# Setup
Output demo/create-files.gif
Set FontSize 14
Set Width 800
Set Height 600
Set Theme "Dracula"

# Launch
Type "./filepicker test-files"
Enter
Sleep 1s

# Navigate to a directory
Down Down Down
Enter
Sleep 300ms

# Create new file
Type "n"
Sleep 300ms
Type "new-document.txt"
Enter
Sleep 800ms

# Create new directory
Type "m"
Sleep 300ms
Type "new-folder"
Enter
Sleep 800ms

# Show created items (navigate to see them)
Down Down Down
Sleep 500ms

# Enter the new directory
Enter
Sleep 500ms

# Go back
Backspace
Sleep 300ms

# Exit
Escape
