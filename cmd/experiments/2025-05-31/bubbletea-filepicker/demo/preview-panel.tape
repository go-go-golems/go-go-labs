# VHS Script: Preview Panel Demo
# Demonstration of dual-panel layout with file preview

# Setup
Output demo/preview-panel.gif
Set FontSize 13
Set Width 1100
Set Height 700
Set Theme "Dracula"

# Launch with preview enabled by default
Type "./filepicker test-files"
Enter
Sleep 1s

# Navigate to text files to show preview
Down Down Down
Enter
Sleep 300ms
Down Down
Enter
Sleep 300ms

# Navigate through files to show different previews
Down Down  # Navigate to text files
Sleep 800ms
Down  # Show different file preview
Sleep 800ms

# Toggle preview off and on
Tab
Sleep 800ms
Tab
Sleep 800ms

# Navigate to show directory preview
Up Up
Enter
Sleep 500ms
Down Down  # Show directory info in preview
Sleep 800ms

# Exit
Escape
