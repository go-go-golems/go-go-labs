# VHS Script: Demo History Navigation Features
Set FontSize 12
Set Width 1400
Set Height 700
Set Theme "Dracula"
Output demo/demo-history-navigation.gif

Type "./filepicker test-files"
Enter
Sleep 2s

# Navigate into documents directory (2nd down from cursor position)
Down Down
Enter
Sleep 1s

# Navigate into personal subdirectory
Down Down
Enter
Sleep 1s

# Navigate into work subdirectory (go back to documents first)
Backspace
Sleep 500ms
Down Down Down  # Navigate to work
Enter
Sleep 1s

# Demonstrate back navigation with 'h'
Type "h"
Sleep 1s

# Go back to root with 'h'
Type "h"
Sleep 1s

# Demonstrate forward navigation with 'l'
Type "l"
Sleep 1s

# Go forward to work with 'l'
Type "l"
Sleep 1s

# Show that we can navigate around and use history
Down
Sleep 500ms
Up
Sleep 500ms

# Final demo of the history working
Type "h"
Sleep 1s
Type "l"
Sleep 1s

# Exit
Escape
