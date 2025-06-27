# VHS Script: Demonstrate Table Format
# Shows the new column-based table format

# Setup 
Output demo/demo-table-format.gif
Set FontSize 13
Set Width 1200
Set Height 700
Set Theme "Dracula"

# Launch the file picker
Type "./filepicker test-files"
Enter
Sleep 2s

# Navigate through different directories to show table format
Down
Sleep 300ms
Down
Sleep 300ms
Up
Sleep 300ms

# Enable detailed view if not already
Type "F3"
Sleep 1s

# Show the table format with aligned columns
Down Down Down Down
Sleep 1s

# Navigate to show different file types
Up Up
Sleep 800ms

# Show how table handles different content
Down Down Down Down Down Down
Sleep 800ms

# Toggle detailed view to show difference
Type "F3"
Sleep 1s
Type "F3"  # Toggle back to detailed
Sleep 1s

# Exit
Escape
