# VHS Script: File Icons and Types Demo
# Demonstration of file type icons and size display

# Setup
Output demo/file-icons.gif
Set FontSize 14
Set Width 900
Set Height 700
Set Theme "Dracula"

# Launch filepicker
Type "./filepicker test-files"
Enter
Sleep 1s

# Navigate to media directory to show image icons
Down Down Down Down Down  # Navigate to media
Enter
Sleep 500ms

# Enter images to show 🖼️ icons
Down Down  # Navigate to images
Enter
Sleep 500ms
Down Down Down  # Show image files
Sleep 1s

# Go back and check projects for code files
Backspace
Sleep 300ms
Backspace
Sleep 300ms

# Navigate to projects for 💻 icons
Down Down Down Down Down Down  # Navigate to projects
Enter
Sleep 300ms
Down Down  # Enter go directory
Enter
Sleep 500ms
Down Down  # Show code files
Sleep 1s

# Go back to show archive files
Backspace
Sleep 300ms
Backspace
Sleep 300ms

# Show archive files with 📦 icons
Down  # Navigate to backups
Enter
Sleep 500ms
Down Down  # Show archive files
Sleep 1s

# Exit
Escape
