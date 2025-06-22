Output demo_full_workflow.gif

Set FontSize 14
Set Width 1200
Set Height 800
Set PlaybackSpeed 0.8

Type "go run . --log-level warn"
Enter
Sleep 3s

# Show template selection
Sleep 2s

# Navigate and select Code Review template
Type "j"
Sleep 1s
Type "k" 
Sleep 1s
Enter
Sleep 2s

# Fill in code snippet variable
Enter
Type "package main"
Enter
Type "import 'fmt'"
Enter
Type "func fibonacci(n int) int {"
Enter  
Type "    if n <= 1 { return n }"
Enter
Type "    return fibonacci(n-1) + fibonacci(n-2)"
Enter
Type "}"
Enter
Sleep 1s

# Fill in language variable
Type "j"
Enter
Type "Go"
Enter
Sleep 1s

# Navigate to section variants
Type "j"
Sleep 500ms

# Change to detailed variant
Enter
Sleep 1s

# Navigate through bullet groups
Type "j"
Sleep 500ms

# Select quality bullet group
Space
Sleep 500ms

# Select security bullet group
Type "j" 
Space
Sleep 500ms

# Select structure bullet group  
Type "j"
Space
Sleep 1s

# Show the preview updating
Sleep 2s

# Copy to clipboard
Type "c"
Sleep 2s

# Save the configuration
Type "s" 
Sleep 1s

# Test with second template
Escape
Sleep 1s
Type "j"
Enter
Sleep 2s

# Fill summarization template
Enter
Type "AI has revolutionized industries through automation and machine learning."
Enter
Sleep 1s

Type "j"
Enter
Type "medium"
Enter
Sleep 1s

# Navigate and select bullet groups
Type "j"
Type "j"
Type "j"
Space
Sleep 500ms
Type "j"
Space
Sleep 2s

# Copy this prompt too
Type "c"
Sleep 1s

# Return to main menu
Escape
Sleep 1s

# Quit application
Type "q"
Sleep 1s
