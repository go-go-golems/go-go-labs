# VHS Script: Screenshot Table Format
# Takes a screenshot of the table format

# Setup 
Output demo/screenshot-table.txt
Set FontSize 14
Set Width 1200
Set Height 700
Set Theme "Dracula"

# Launch the file picker
Type "./filepicker test-files"
Enter
Sleep 2s

# Take screenshot showing table format
Screenshot demo/table-layout.txt

# Navigate down to see more files
Down Down Down Down
Sleep 1s

# Take another screenshot
Screenshot demo/table-with-selection.txt

# Exit without selecting
Type "q"
