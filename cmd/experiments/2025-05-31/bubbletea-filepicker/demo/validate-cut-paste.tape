# VHS Script: Validate Cut & Paste (Move) Functionality
# Creates txt screenshots to validate cut/paste operations

# Setup for validation
Output demo/validate-cut-paste.txt
Set FontSize 14
Set Width 800
Set Height 600
Set Theme "Dracula"

# Launch
Type "./filepicker test-files"
Enter
Sleep 2s

# Navigate to media/audio to find files to move
Down Down Down Down Down  # Navigate to media
Enter
Sleep 1s
Down  # Navigate to audio
Enter
Sleep 1s

# Select files for cut (move)
Down  # Skip ..
Space  # Select first file
Sleep 500ms
Down
Space  # Select second file
Sleep 500ms

# Cut files (x key)
Type "x"
Sleep 1s

# Take screenshot to show cut operation
Screenshot demo/cut-operation.txt

# Navigate back and then to video directory
Backspace
Sleep 500ms
Down Down  # Navigate to video
Enter
Sleep 1s

# Paste files (v key) - they should move
Type "v"
Sleep 2s

# Take screenshot to show paste result
Screenshot demo/cut-paste-result.txt

# Go back to audio directory to verify files were moved (not copied)
Backspace
Sleep 500ms
Up  # Back to audio
Enter
Sleep 1s

# Take screenshot to show original location is empty
Screenshot demo/original-location-after-move.txt

# Exit
Escape
