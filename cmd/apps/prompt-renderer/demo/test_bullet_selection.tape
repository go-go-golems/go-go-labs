Output test_bullet_selection.txt

Set FontSize 12
Set Width 120
Set Height 120

Type "go run . --log-level warn"
Enter
Sleep 3s

# Enter template config
Enter
Sleep 2s

# Skip variables by using Escape to clear and just putting simple values
Enter
Type "func test() { return 42 }"
Enter
Sleep 300ms

Type "j"
Enter  
Type "go"
Enter
Sleep 500ms

# Navigate to bullet section
Type "j"
Type "j"
Type "j"
Type "j"
Sleep 500ms

# Select quality bullet (should show ☑)
Space
Sleep 800ms

# Move to security and select it  
Type "j"
Space
Sleep 800ms

# Take screenshot showing selected bullets
Screenshot test_bullet_selection.txt
Sleep 1000ms

# Quit
Escape
Type "q"
