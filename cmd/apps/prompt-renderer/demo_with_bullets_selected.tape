Output demo_with_bullets_selected.txt

Set FontSize 12
Set Width 120
Set Height 120

Type "go run . --log-level warn"
Enter
Sleep 3s

# Enter template config
Enter
Sleep 2s

# Fill in sample data quickly
Enter
Type "func fibonacci(n int) int { return n }"
Enter
Sleep 300ms

Type "j"
Enter  
Type "go"
Enter
Sleep 500ms

# Navigate to bullets
Type "j"
Type "j"
Type "j"
Type "j"
Sleep 500ms

# Select quality bullet group
Space
Sleep 500ms

# Move to security and select
Type "j"
Space
Sleep 500ms

# Move back to see both selected
Type "k"
Sleep 1000ms

# Take screenshot
Screenshot demo_with_bullets_selected.txt
Sleep 1000ms

# Quit
Escape
Type "q"
