# VHS Script: Test Table Format Layout
# Tests the new column-based table format

# Setup 
Output demo/test-table-format.txt
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Dracula"

# Launch the file picker
Type "./filepicker test-files"
Enter
Sleep 2s

# Navigate to show files with different sizes and types
Down Down Down
Enter
Sleep 1s

# Navigate to work directory with files
Down Down
Enter
Sleep 1s

# Take screenshot showing table format
Screenshot demo/table-format-work.txt

# Go back and try media directory
Backspace
Sleep 500ms
Backspace
Sleep 500ms

# Navigate to media directory
Down Down Down Down Down
Enter
Sleep 1s

# Navigate to images with different file sizes
Down Down Down
Enter
Sleep 1s

# Take screenshot showing table format with images
Screenshot demo/table-format-images.txt

# Test with longer filenames
Down Down Down
Sleep 1s

# Take final screenshot
Screenshot demo/table-format-final.txt

# Exit
Escape
