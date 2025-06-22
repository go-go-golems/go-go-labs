Output test_new_features.txt

Set FontSize 12
Set Width 120
Set Height 120

Type "go run . --log-level warn"
Enter
Sleep 3s

# Navigate to template with toggle features
Type "j"
Type "j"
Enter
Sleep 2s

# Fill in variable
Enter
Type "func hello() { return 'world' }"
Enter
Sleep 500ms

# Navigate to toggle and enable it
Type "j"
Space
Sleep 1000ms

# Take screenshot showing toggle enabled
Screenshot test_new_features.txt
Sleep 1000ms

# Test individual bullet selection on first template
Escape
Sleep 1s
Enter
Sleep 2s

# Skip to bullets
Type "j"
Type "j"
Type "j"
Type "j"
Sleep 500ms

# Select individual bullets
Space
Sleep 500ms
Type "j"
Space
Sleep 500ms
Type "j"
Type "j"
Space
Sleep 1000ms

# Take screenshot showing individual bullet selection
Screenshot test_individual_bullets.txt
Sleep 1000ms

# Quit
Escape
Type "q"
