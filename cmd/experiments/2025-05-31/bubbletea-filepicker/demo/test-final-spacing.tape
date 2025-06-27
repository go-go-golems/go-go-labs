# VHS Script: Test Final Column Spacing
# Final validation of column spacing and selected row display

# Setup for wide rendering
Output demo/test-final-spacing.txt
Set FontSize 13
Set Width 1400
Set Height 700
Set Theme "Dracula"

# Launch the file picker
Type "./filepicker test-files"
Enter
Sleep 2s

# Take screenshot of improved spacing
Screenshot demo/final-wide-format.txt

# Select and multi-select to test all scenarios
Down Down Down
Space
Sleep 300ms
Down
Space
Sleep 300ms

# Take screenshot showing multi-selection with full permissions
Screenshot demo/final-multi-selection.txt

# Navigate to see more file types
Down Down Down
Sleep 500ms

# Take final screenshot
Screenshot demo/final-file-types.txt

# Exit
Type "q"
