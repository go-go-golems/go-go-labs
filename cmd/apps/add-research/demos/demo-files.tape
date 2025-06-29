# Demo: File Attachments
# Shows attaching files with -f flag and file browser functionality

Output demos/demo-files.gif

Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Dracula"

# Setup
Type "cd cmd/apps/add-research"
Enter
Sleep 500ms

Type "echo '# Demo: File Attachments' && echo"
Enter
Sleep 1s

# Create some sample files to attach
Type "echo 'Creating sample files to attach...'"
Enter

Type "echo 'server:\n  host: localhost\n  port: 8080\ndatabase:\n  driver: postgres' > config.yaml"
Enter

Type "echo 'type User {\n  id: ID!\n  name: String!\n  email: String!\n}' > schema.graphql"
Enter

Type "echo 'package main' > example.go"
Enter
Type "echo 'import fmt' >> example.go"
Enter  
Type "echo 'func main() { fmt.Println(Hello) }' >> example.go"
Enter

Sleep 1s

Type "echo '📁 Files created. Now attaching them to a research note...'"
Enter
Sleep 1s

# Attach files using -f flag
Type "./add-research --title 'Configuration Analysis' --file config.yaml --file schema.graphql --message 'Reviewing configuration files and GraphQL schema'"
Enter
Sleep 2s

# Links (can skip)
Type ""
Enter
Sleep 1s

Type "echo '✅ Note with attachments created!'"
Enter
Sleep 1s

# Show the result
Type "echo '📄 Preview of attached files in note:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Configuration-Analysis*' -exec cat {} \\; | head -20"
Enter
Sleep 3s

Type "echo '' && echo '🎯 Now trying interactive file browser...'"
Enter
Sleep 1s

# Demo file browser (if practical - this might need modification based on actual behavior)
Type "./add-research --browse --title 'Project Files Review' --no-links"
Enter
Sleep 1s

# Simulate file selection in browser (this might not work exactly as shown)
Type "example.go"
Enter
Sleep 500ms

Type "y"
Enter
Sleep 1s

# Add some content
Type "Reviewed Go example file structure and patterns."
Ctrl+D
Sleep 1s

Type "echo '✅ File browser demo completed!'"
Enter
Sleep 2s

# Show syntax highlighting preview
Type "echo '🎨 File attachments include syntax highlighting:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Project-Files-Review*' -exec head -15 {} \\;"
Enter
Sleep 3s

Screenshot demos/demo-files.txt
