Output screenshot_template_config.txt

Set FontSize 12
Set Width 120
Set Height 120

Type "go run . --log-level warn"
Enter
Sleep 3s

# Enter template config
Enter
Sleep 2s

# Fill in some data
Enter
Type "func hello() { return 'world' }"
Enter
Sleep 500ms

# Next field
Type "j"
Enter  
Type "go"
Enter
Sleep 1s

# Navigate to bullets and select some
Type "j"
Type "j"
Type "j"
Type "j"
Space
Sleep 500ms
Type "j"
Space
Sleep 1s

# Take screenshot
Screenshot screenshot_template_config.txt
Sleep 1000ms

# Quit
Escape
Type "q"
