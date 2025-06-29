# Demo: Content Sources
# Shows clipboard integration and piped input

Output demos/demo-content.gif

Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Dracula"

# Setup
Type "cd cmd/apps/add-research"
Enter
Sleep 500ms

Type "echo '# Demo: Content Sources' && echo"
Enter
Sleep 1s

# Demonstrate clipboard integration
Type "echo 'Setting up clipboard content...'"
Enter

# Simulate clipboard content (using echo to simulate)
Type "echo 'Key research insights:\n- Microservices architecture benefits\n- Container orchestration with Kubernetes\n- CI/CD pipeline optimization strategies' | tee /tmp/clipboard_content"
Enter
Sleep 1s

# Copy to clipboard (simulate)
Type "cat /tmp/clipboard_content | head -4"
Enter
Sleep 1s

Type "echo '📋 Clipboard content ready. Creating note from clipboard...'"
Enter
Sleep 1s

# Create note from clipboard
Type "cat /tmp/clipboard_content | ./add-research --clip --title 'Microservices Research'"
Enter
Sleep 1s

# Add links
Type "https://microservices.io/"
Enter
Type "https://kubernetes.io/docs/concepts/"
Enter
Type ""
Enter
Sleep 1s

Type "echo '✅ Note created from clipboard!'"
Enter
Sleep 1s

# Show piped input
Type "echo '' && echo '📥 Now demonstrating piped input...'"
Enter
Sleep 1s

# Create content to pipe
Type "echo 'Performance Benchmarks:\n\n## Database Queries\n- SELECT operations: 150ms avg\n- INSERT operations: 45ms avg\n- Complex joins: 300ms avg\n\n## API Response Times\n- GET /users: 25ms\n- POST /orders: 80ms' > /tmp/performance_data"
Enter

# Pipe content to add-research
Type "cat /tmp/performance_data | ./add-research --title 'Performance Analysis' --no-links"
Enter
Sleep 2s

Type "echo '✅ Note created from piped input!'"
Enter
Sleep 1s

# Show combined sources
Type "echo '' && echo '🔄 Combining multiple content sources...'"
Enter
Sleep 1s

# Create note with message + file + clipboard simulation
Type "echo 'Additional context from command line' | ./add-research --title 'Multi-Source Research' --message 'Base findings from investigation' --file config.yaml"
Enter
Sleep 1s

# Add links
Type "https://example.com/research"
Enter
Type ""
Enter
Sleep 1s

# Show results
Type "echo '📊 Results summary:'"
Enter
Sleep 1s

Type "echo '1. Clipboard-based note:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Microservices-Research*' -exec head -8 {} \\;"
Enter
Sleep 2s

Type "echo '' && echo '2. Piped input note:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Performance-Analysis*' -exec head -8 {} \\;"
Enter
Sleep 2s

Type "echo '' && echo '3. Multi-source note:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Multi-Source-Research*' -exec head -8 {} \\;"
Enter
Sleep 3s

Screenshot demos/demo-content.txt
