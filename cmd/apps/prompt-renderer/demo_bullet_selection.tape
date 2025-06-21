Output demo_bullet_selection.txt

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
Type "package main"
Enter
Sleep 300ms

Type "j"
Enter  
Type "go"
Enter
Sleep 300ms

# Navigate to bullets
Type "j"
Type "j"
Type "j"
Type "j"
Sleep 500ms

# Select first bullet group
Space
Sleep 800ms

# Select second bullet group  
Type "j"
Space
Sleep 800ms

# Select third bullet group
Type "j"
Space
Sleep 1000ms

# Take screenshot showing all selected
Screenshot demo_bullet_selection.txt
Sleep 1000ms

# Quit
Escape
Type "q"
