# VHS Script: Basic Navigation Demo
# Simple demonstration of launching and basic navigation

# Setup
Output demo/basic-navigation.gif
Set FontSize 14
Set Width 800
Set Height 600
Set Theme "Dracula"

# Setup and launch
Type "./filepicker test-files"
Enter
Sleep 1s

# Basic navigation
Down Down Down
Sleep 300ms
Up Up  
Sleep 300ms

# Enter directory
Enter
Sleep 500ms

# Navigate in subdirectory
Down Down
Sleep 300ms

# Go back
Backspace
Sleep 500ms

# Exit
Escape
