# Demo: File Attachments
# Shows attaching files with -f flag and syntax highlighting

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

# Create some simple sample files to attach
Type "echo 'Creating sample files to attach...'"
Enter
Sleep 1s

# Create simple files without complex escaping
Type "echo 'host: localhost' > config.yaml"
Enter
Type "echo 'port: 8080' >> config.yaml"
Enter

Type "echo 'version=1.0' > app.properties"
Enter
Type "echo 'debug=true' >> app.properties"
Enter

Type "echo 'README content here' > README.md"
Enter

Sleep 1s

Type "echo '📁 Files created. Attaching them to a research note...'"
Enter
Sleep 1s

# Attach multiple files using -f flag
Type "./add-research --title 'Configuration Files Review' --file config.yaml --file app.properties --file README.md --no-links --message 'Reviewing project configuration files.'"
Enter
Sleep 3s

Type "echo '✅ Note with multiple file attachments created!'"
Enter
Sleep 1s

# Show the result
Type "echo '📄 Preview of attached files in note:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Configuration-Files-Review*' -exec head -20 {} \\;"
Enter
Sleep 4s

Type "echo '' && echo '📊 File info:'"
Enter
Type "ls -la config.yaml app.properties README.md"
Enter
Sleep 2s

# Demonstrate that existing files can also be attached
Type "echo '' && echo '📝 Attaching existing project files:'"
Enter
Sleep 1s

# Use existing files from the project
Type "./add-research --title 'Project Files Analysis' --file go.mod --file README.md --no-links --message 'Analysis of existing project structure and dependencies.'"
Enter
Sleep 2s

Type "echo '📄 Project files note preview:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Project-Files-Analysis*' -exec head -15 {} \\;"
Enter
Sleep 3s

# Show file contents are preserved with syntax highlighting
Type "echo '' && echo '🎨 Files are embedded with syntax highlighting:'"
Enter
Type "echo '• YAML files show structured data'"
Enter
Type "echo '• Properties files show key-value pairs'"
Enter
Type "echo '• Markdown files preserve formatting'"
Enter
Type "echo '• Go mod files show dependencies'"
Enter
Sleep 2s

# Show the complete note structure
Type "echo '' && echo '📋 Complete note with attachments:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Configuration-Files-Review*' -exec wc -l {} \\;"
Enter
Sleep 1s

# Clean up sample files
Type "echo '' && echo '🧹 Cleaning up sample files...'"
Enter
Type "rm -f config.yaml app.properties"
Enter
Sleep 1s

Type "echo '✅ File attachments demo completed!'"
Enter
Sleep 2s

Screenshot demos/demo-files.txt
