# Demo: Basic Note Creation
# Shows creating a simple research note with interactive title input

Output demos/demo-basic.gif

Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Dracula"

# Setup: ensure we're in the right directory and have a clean slate
Type "cd cmd/apps/add-research"
Enter
Sleep 500ms

Type "echo '# Demo: Basic Note Creation' && echo"
Enter
Sleep 1s

# Show help briefly
Type "./add-research --help | head -8"
Enter
Sleep 2s

Type "clear"
Enter
Sleep 500ms

# Create a basic note with command line options (avoid interactive mode for VHS)
Type "echo 'Creating a basic research note...'"
Enter
Sleep 1s

# Use command line arguments to avoid interactive prompts
Type "./add-research --title 'API Documentation Review' --no-links --message 'Key findings from REST API research: API design patterns, performance optimization, and next steps for implementation.'"
Enter
Sleep 2s

Type "echo '✅ Note created! Let me show you the file structure:'"
Enter
Sleep 1s

# Show the created file structure
Type "find ~/code/wesen/obsidian-vault/research -name '*API-Documentation-Review*' -type f -exec ls -la {} \\;"
Enter
Sleep 2s

# Show content of the created file  
Type "echo '📄 Content preview:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*API-Documentation-Review*' -exec head -15 {} \\;"
Enter
Sleep 3s

# Show that we can add links afterwards
Type "echo '🔗 Adding more content with append mode:'"
Enter
Type "./add-research --append --title 'API Documentation Review' --no-links --message 'Additional research findings: OAuth 2.0 implementation details and GraphQL comparison completed.'"
Enter
Sleep 2s

Type "echo '📄 Updated note:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*API-Documentation-Review*' -exec tail -8 {} \\;"
Enter
Sleep 3s

Screenshot demos/demo-basic.txt
