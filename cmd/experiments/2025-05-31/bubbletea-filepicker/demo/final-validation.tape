# VHS Script: Final Validation of Table Format
# Creates a clean screenshot for before/after comparison

# Setup for wide rendering
Output demo/final-validation.txt
Set FontSize 13
Set Width 1400
Set Height 700
Set Theme "Dracula"

# Launch the file picker
Type "./filepicker test-files"
Enter
Sleep 2s

# Take final validation screenshot
Screenshot demo/table-format-final.txt

# Exit
Type "q"
