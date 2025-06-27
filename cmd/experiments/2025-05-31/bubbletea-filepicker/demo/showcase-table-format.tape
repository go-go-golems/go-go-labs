# VHS Script: Showcase Final Table Format
# Demonstrates the excellent table format with proper column spacing

# Setup for very wide rendering to showcase spacing
Output demo/showcase-table-format.gif
Set FontSize 12
Set Width 1600
Set Height 800
Set Theme "Dracula"

# Launch the file picker
Type "./filepicker test-files"
Enter
Sleep 2s

# Navigate through different areas to show table format
Down
Sleep 800ms
Down
Sleep 800ms
Down
Sleep 800ms

# Multi-select some items
Space
Sleep 500ms
Down
Space
Sleep 500ms
Down

# Show the beautiful table format
Sleep 1s

# Navigate through different file types
Down Down
Sleep 800ms

# Navigate into a directory to show more files
Up Up Up
Enter
Sleep 1s

# Navigate to work directory with different file types
Down Down
Enter
Sleep 1s

# Show different file types in table format
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms

# Go back to show overall layout
Backspace
Sleep 1s
Backspace
Sleep 1s

# Final showcase of the layout
Sleep 2s

# Exit
Escape
