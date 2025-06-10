# VHS Script: Quick Overview Demo
# Quick demonstration of main features

# Setup
Output demo/overview.gif
Set FontSize 14
Set Width 900
Set Height 700
Set Theme "Dracula"

# Launch
Type "./filepicker test-files"
Enter
Sleep 1s

# Basic navigation
Down Down Down
Enter
Sleep 300ms

# Multi-selection
Down Down
Space  # Select file
Sleep 300ms
Down
Space  # Select another
Sleep 300ms

# Copy operation
Type "c"
Sleep 300ms

# Navigate and paste
Backspace
Sleep 300ms
Down
Enter
Sleep 300ms
Type "v"
Sleep 500ms

# Show help briefly
Type "?"
Sleep 1s
Type "?"
Sleep 300ms

# Exit
Escape
