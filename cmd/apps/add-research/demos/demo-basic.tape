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
Type "./add-research --help | head -10"
Enter
Sleep 2s

Type "clear"
Enter
Sleep 500ms

# Create a basic note with interactive title
Type "echo 'Creating a basic research note...'"
Enter
Sleep 1s

Type "./add-research"
Enter
Sleep 1s

# Title input
Type "API Documentation Review"
Enter
Sleep 500ms

# Content input
Type "## Key Findings"
Enter
Type ""
Enter
Type "- REST API best practices discovered"
Enter
Type "- GraphQL vs REST comparison completed"
Enter
Type "- Authentication patterns documented"
Enter
Type ""
Enter
Type "## Next Steps"
Enter
Type "- Implement OAuth 2.0 flow"
Enter
Type "- Review API versioning strategy"
Enter

# Finish content input (Ctrl+D)
Ctrl+D
Sleep 1s

# Links input - add a few relevant links
Type "https://restfulapi.net/rest-api-design-tutorial-with-example/"
Enter
Type "https://graphql.org/learn/best-practices/"
Enter
Type ""
Enter
Sleep 1s

# Show the created file structure
Type "echo '✅ Note created! Let me show you the file structure:'"
Enter
Sleep 1s

Type "find ~/code/wesen/obsidian-vault/research -name '*.md' -type f -exec ls -la {} \\; | tail -3"
Enter
Sleep 2s

# Show content of the created file
Type "echo '📄 Content preview:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*API-Documentation-Review*' -exec head -10 {} \\;"
Enter
Sleep 3s

Screenshot demos/demo-basic.txt
