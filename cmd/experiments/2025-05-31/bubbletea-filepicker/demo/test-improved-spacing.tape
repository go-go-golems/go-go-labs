# VHS Script: Test Improved Column Spacing
# Verify the improved spacing and fix for selected row truncation

# Setup for wide rendering
Output demo/test-improved-spacing.txt
Set FontSize 13
Set Width 1400
Set Height 700
Set Theme "Dracula"

# Launch the file picker
Type "./filepicker test-files"
Enter
Sleep 2s

# Take screenshot of initial state with improved spacing
Screenshot demo/improved-wide-format.txt

# Navigate to show selection
Down Down Down
Sleep 500ms

# Take screenshot with selection - should show full permissions
Screenshot demo/improved-selected-row.txt

# Multi-select to test both selection indicators
Space
Sleep 300ms
Down
Space
Sleep 300ms
Up
Sleep 300ms

# Take screenshot with multi-selection
Screenshot demo/improved-multi-selected.txt

# Navigate to a file with longer name to test truncation
Down Down Down Down
Sleep 500ms

# Take screenshot with longer filename
Screenshot demo/improved-long-filename.txt

# Exit
Type "q"
