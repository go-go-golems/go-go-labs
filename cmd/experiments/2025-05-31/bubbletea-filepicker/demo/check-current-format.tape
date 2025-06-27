# VHS Script: Check Current Format Issues
# Identify spacing and selected row truncation issues

# Setup for wide rendering
Output demo/check-current-format.txt
Set FontSize 13
Set Width 1400
Set Height 700
Set Theme "Dracula"

# Launch the file picker
Type "./filepicker test-files"
Enter
Sleep 2s

# Take screenshot of initial state
Screenshot demo/current-wide-format.txt

# Navigate to show selection
Down Down Down
Sleep 500ms

# Take screenshot with selection
Screenshot demo/current-selected-row.txt

# Multi-select to see the issue
Space
Sleep 300ms
Down
Space
Sleep 300ms

# Take screenshot with multi-selection
Screenshot demo/current-multi-selected.txt

# Exit
Type "q"
