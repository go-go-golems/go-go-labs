Output demo_template_config.gif

Set FontSize 16
Set Width 1200
Set Height 700

Type "go run . --log-level warn"
Enter
Sleep 2s

# Select first template
Enter
Sleep 2s

# Navigate through form fields
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms

# Edit a variable
Type "k"
Sleep 500ms
Enter
Type "func main() {\n    fmt.Println(\"Hello, World!\")\n}"
Enter
Sleep 1s

# Navigate to next variable
Type "j"
Enter
Type "go"
Enter
Sleep 1s

# Navigate to sections
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms

# Change section variant
Enter
Sleep 1s

# Navigate to bullet groups
Type "j"
Sleep 500ms

# Toggle bullet groups
Space
Sleep 500ms
Type "j"
Space
Sleep 500ms
Type "j"
Space
Sleep 1s

# Copy to clipboard
Type "c"
Sleep 2s

# Save selection
Type "s"
Sleep 1s

# Go back
Escape
Sleep 1s

# Quit
Type "q"
