# Demo: Search & Export
# Shows export functionality and demonstrates search capabilities

Output demos/demo-search.gif

Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Dracula"

# Setup
Type "cd cmd/apps/add-research"
Enter
Sleep 500ms

Type "echo '# Demo: Search & Export' && echo"
Enter
Sleep 1s

# First, create a few notes to have something to search and export
Type "echo 'Setting up test notes for search & export demo...'"
Enter
Sleep 1s

# Create test notes with different dates and content
Type "./add-research --title 'Docker Best Practices' --no-links --message 'Container optimization techniques: minimal base images, non-root users, vulnerability scanning, multi-stage builds, resource limits.'"
Enter
Sleep 1s

Type "./add-research --title 'Kubernetes Deployment' --no-links --message 'Pod management patterns: rolling updates, blue-green deployments, service discovery, load balancing, ingress controllers.'"
Enter
Sleep 1s

Type "./add-research --title 'Database Scaling' --date '2024-01-10' --no-links --message 'Scaling strategies: vertical scaling with CPU/memory increases, horizontal scaling with read replicas and sharding.'"
Enter
Sleep 1s

Type "echo '✅ Test notes created!'"
Enter
Sleep 1s

# Show current notes structure
Type "echo '' && echo '📁 Current notes structure:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*.md' -type f | grep -E '(Docker|Kubernetes|Database)' | head -5"
Enter
Sleep 2s

# Since interactive search doesn't work in VHS, demonstrate export instead
Type "echo '' && echo '📤 Demonstrating export functionality...'"
Enter
Sleep 1s

# Export all notes
Type "./add-research --export --export-path 'complete-research-export.md'"
Enter
Sleep 2s

Type "echo '✅ Complete export created!'"
Enter
Sleep 1s

# Show export preview
Type "echo '📄 Export preview (first 15 lines):'"
Enter
Type "head -15 complete-research-export.md"
Enter
Sleep 3s

# Export with date filtering
Type "echo '' && echo '📅 Filtered export (January 2024):'"
Enter
Sleep 1s

Type "./add-research --export --export-from '2024-01-01' --export-to '2024-01-31' --export-path 'january-2024-export.md'"
Enter
Sleep 2s

Type "echo '📄 January 2024 export preview:'"
Enter
Type "head -10 january-2024-export.md"
Enter
Sleep 2s

# Show export statistics
Type "echo '' && echo '📊 Export statistics:'"
Enter
Type "echo 'Complete export:' && wc -l complete-research-export.md"
Enter
Type "echo 'January 2024 export:' && wc -l january-2024-export.md"
Enter
Sleep 2s

# Demonstrate searching by showing file patterns (since interactive search won't work)
Type "echo '' && echo '🔍 Search capabilities (file-based):'"
Enter
Sleep 1s

Type "echo 'Finding notes about Docker:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Docker*' -type f"
Enter
Sleep 1s

Type "echo '' && echo 'Finding notes about databases:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Database*' -type f"
Enter
Sleep 1s

Type "echo '' && echo 'Content search (grep through files):'"
Enter
Type "grep -r 'scaling' ~/code/wesen/obsidian-vault/research/ --include='*.md' | head -2"
Enter
Sleep 2s

# Show export with today's filter
Type "echo '' && echo '📅 Today\\'s notes export:'"
Enter
Type "./add-research --export --export-from '$(date +%Y-%m-%d)' --export-to '$(date +%Y-%m-%d)' --export-path 'today-export.md'"
Enter
Sleep 2s

Type "echo '📄 Today\\'s export preview:'"
Enter
Type "head -8 today-export.md"
Enter
Sleep 2s

# Show different export formats by showing file sizes
Type "echo '' && echo '📈 Export file comparison:'"
Enter
Type "ls -lh *export.md"
Enter
Sleep 2s

# Show that we can combine search with export
Type "echo '' && echo '🎯 Search & Export workflow:'"
Enter
Type "echo '1. Create notes on different dates'"
Enter  
Type "echo '2. Use --export with date filters'"
Enter
Type "echo '3. Use grep to find content across notes'"
Enter
Type "echo '4. Interactive search works in terminal (TTY required)'"
Enter
Sleep 3s

# Clean up
Type "echo '' && echo '🧹 Cleaning up demo files...'"
Enter
Type "rm -f complete-research-export.md january-2024-export.md today-export.md"
Enter
Sleep 1s

Type "echo '✅ Search & Export demo completed!'"
Enter
Sleep 2s

Screenshot demos/demo-search.txt
