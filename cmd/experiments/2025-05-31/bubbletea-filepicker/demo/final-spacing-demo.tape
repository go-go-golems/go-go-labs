# VHS Script: Final Demo of Improved Column Spacing
Set FontSize 12
Set Width 1400
Set Height 700
Set Theme "Dracula"
Output demo/final-spacing-demo.gif

Type "./filepicker test-files"
Enter
Sleep 2s

# Show the improved spacing
Sleep 1s

# Navigate through some files to show different sizes
Down Down Down Down Down Down
Sleep 1s

# Navigate to show the large file we created
Down
Sleep 1s

# Show multi-selection still works with new spacing
Space
Sleep 500ms
Up
Space
Sleep 500ms

# Final showcase
Sleep 2s

# Exit
Escape
