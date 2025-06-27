# VHS Script: Showcase Clean Layout Without Permissions
Set FontSize 12
Set Width 1400
Set Height 700
Set Theme "Dracula"
Output demo/showcase-no-permissions.gif

Type "./filepicker test-files"
Enter
Sleep 2s

# Show the clean layout without permissions
Sleep 1s

# Navigate through files to show different sizes
Down Down Down Down Down Down Down
Sleep 1s

# Show multi-selection works perfectly
Space
Sleep 500ms
Up
Space
Sleep 500ms
Down Down
Space
Sleep 1s

# Final showcase of the clean, focused layout
Sleep 2s

# Exit
Escape
