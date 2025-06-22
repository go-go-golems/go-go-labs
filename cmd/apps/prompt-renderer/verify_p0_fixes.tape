# VHS Script to verify P0 fixes for prompt-renderer
# Demonstrates bullet toggling, section rendering, keyboard navigation, and error handling

Output verify_p0_fixes.gif

Set FontSize 16
Set Width 1400
Set Height 900

# Start the application
Type "go run ."
Enter
Sleep 2s

# Navigate to Code Review with Context template (should be template 3)
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms

# Select the template
Enter
Sleep 2s

# Fill in the code snippet variable
Type 'func main() { fmt.Println("Hello World") }'
Sleep 1s
Enter
Sleep 1s

# Navigate to context_request section using keyboard navigation
Type "j"
Sleep 500ms

# Test Space key toggling (should enable context request)
Type " "
Sleep 1s

# Navigate down to check bullets if present
Type "j"
Sleep 500ms

# Test Tab navigation 
Type "Tab"
Sleep 500ms
Type "Tab"
Sleep 500ms

# Test Shift+Tab navigation
Type "shift+tab"
Sleep 500ms

# Test Enter key toggling on bullets/toggles
Type "Enter"
Sleep 1s

# Navigate around to show keyboard navigation works
Type "k"
Sleep 500ms
Type "j"
Sleep 500ms

# Copy the prompt to clipboard (test c key)
Type "c"
Sleep 2s

# Go back to template list
Type "Escape"
Sleep 1s

# Test different template to show section rendering works
Type "k"
Sleep 500ms
Enter
Sleep 2s

# Navigate to bullet section and test toggling
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms
Type "j"
Sleep 500ms

# Toggle some bullets
Type "Space"
Sleep 500ms
Type "j"
Sleep 500ms
Type "Space" 
Sleep 500ms
Type "j"
Sleep 500ms
Type "Enter"
Sleep 1s

# Show that preview updates correctly
Sleep 2s

# Exit the application
Type "Escape"
Sleep 1s
Type "q"
Sleep 1s
