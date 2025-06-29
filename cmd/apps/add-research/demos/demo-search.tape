# Demo: Search & Export
# Shows search functionality and export with date filtering

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

# First, create a few notes to have something to search
Type "echo 'Setting up test notes for search demo...'"
Enter
Sleep 1s

# Create first test note
Type "./add-research --title 'Docker Best Practices' --message 'Container optimization techniques and security guidelines' --no-links"
Enter
Sleep 1s

# Create second test note 
Type "./add-research --title 'Kubernetes Deployment' --message 'Pod management and service discovery patterns' --no-links"
Enter
Sleep 1s

# Create third test note with different date
Type "./add-research --title 'Database Scaling' --date '2024-01-10' --message 'Horizontal vs vertical scaling strategies for PostgreSQL' --no-links"
Enter
Sleep 1s

Type "echo '✅ Test notes created!'"
Enter
Sleep 1s

# Demonstrate search functionality
Type "echo '' && echo '🔍 Searching existing notes...'"
Enter
Sleep 1s

Type "./add-research --search"
Enter
Sleep 1s

# Simulate search selection (this might need adjustment based on actual interface)
Type "Docker"
Enter
Sleep 1s

# Show search results and potentially select one
Type "1"
Enter
Sleep 1s

# Copy to clipboard option
Type "y"
Enter
Sleep 1s

Type "echo '✅ Search completed with clipboard copy!'"
Enter
Sleep 1s

# Demonstrate export functionality
Type "echo '' && echo '📤 Exporting notes...'"
Enter
Sleep 1s

# Export all notes
Type "./add-research --export --export-path 'all-research-notes.md'"
Enter
Sleep 2s

Type "echo '✅ Export created!'"
Enter
Sleep 1s

# Show export preview
Type "echo '📄 Export preview:'"
Enter
Type "head -20 all-research-notes.md"
Enter
Sleep 3s

# Export with date filtering
Type "echo '' && echo '📅 Exporting with date filter...'"
Enter
Sleep 1s

Type "./add-research --export --export-from '2024-01-10' --export-to '2024-01-11' --export-path 'filtered-export.md'"
Enter
Sleep 2s

Type "echo '📄 Filtered export preview:'"
Enter
Type "head -15 filtered-export.md"
Enter
Sleep 2s

# Show file sizes and counts
Type "echo '' && echo '📊 Export statistics:'"
Enter
Type "echo 'All notes export:' && wc -l all-research-notes.md"
Enter
Type "echo 'Filtered export:' && wc -l filtered-export.md"
Enter
Sleep 2s

# Clean up
Type "echo '' && echo '🧹 Cleaning up demo files...'"
Enter
Type "rm -f all-research-notes.md filtered-export.md"
Enter
Sleep 1s

Type "echo '✅ Search & Export demo completed!'"
Enter
Sleep 2s

Screenshot demos/demo-search.txt
