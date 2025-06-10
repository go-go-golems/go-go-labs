# VHS Script: Help System Demo
# Demonstration of the built-in help system

# Setup
Output demo/help-system.gif
Set FontSize 13
Set Width 1000
Set Height 700
Set Theme "Dracula"

# Launch
Type "./filepicker test-files"
Enter
Sleep 1s

# Show normal interface first
Down Down
Sleep 300ms

# Toggle help system
Type "?"
Sleep 2s

# Scroll through help sections (showing all key bindings)
Sleep 1s

# Toggle help off
Type "?"
Sleep 500ms

# Show that we're back to normal interface
Down
Sleep 300ms

# Toggle help back on briefly
Type "?"
Sleep 1s

# Turn off and exit
Type "?"
Sleep 300ms
Escape
