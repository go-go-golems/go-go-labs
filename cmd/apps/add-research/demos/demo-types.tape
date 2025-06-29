# Demo: Types & Organization
# Shows different note types, custom dates, and vault organization

Output demos/demo-types.gif

Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Dracula"

# Setup
Type "cd cmd/apps/add-research"
Enter
Sleep 500ms

Type "echo '# Demo: Note Types & Organization' && echo"
Enter
Sleep 1s

# Show different note types
Type "echo '🏷️  Creating notes with different types...'"
Enter
Sleep 1s

# Research note (default)
Type "./add-research --type 'research' --title 'Machine Learning Study' --message 'Deep learning architectures and training methodologies' --no-links"
Enter
Sleep 1s

# Ideas note
Type "./add-research --type 'ideas' --title 'Mobile App Concept' --message 'Cross-platform productivity app with offline sync capabilities' --no-links"
Enter
Sleep 1s

# Meeting notes
Type "./add-research --type 'meetings' --title 'Team Sync January' --message 'Sprint planning and architecture decisions for Q1' --no-links"
Enter
Sleep 1s

# Technical note with custom date
Type "./add-research --type 'technical' --date '2024-01-05' --title 'Database Migration Strategy' --message 'Zero-downtime migration approach for production database' --no-links"
Enter
Sleep 1s

Type "echo '✅ Different note types created!'"
Enter
Sleep 1s

# Show vault organization structure
Type "echo '' && echo '📁 Vault organization structure:'"
Enter
Sleep 1s

Type "echo 'Research notes:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -type f -name '*.md' | head -3"
Enter
Sleep 1s

Type "echo '' && echo 'Ideas notes:'"
Enter
Type "find ~/code/wesen/obsidian-vault/ideas -type f -name '*.md' | head -3"
Enter
Sleep 1s

Type "echo '' && echo 'Meeting notes:'"
Enter
Type "find ~/code/wesen/obsidian-vault/meetings -type f -name '*.md' | head -3"
Enter
Sleep 1s

Type "echo '' && echo 'Technical notes:'"
Enter
Type "find ~/code/wesen/obsidian-vault/technical -type f -name '*.md' | head -3"
Enter
Sleep 2s

# Show metadata example
Type "echo '' && echo '📊 Creating note with metadata...'"
Enter
Sleep 1s

Type "./add-research --type 'research' --metadata --title 'Important Research' --message 'Critical findings that need structured metadata' --no-links"
Enter
Sleep 1s

Type "echo '📄 Metadata example:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Important-Research*' -exec head -15 {} \\;"
Enter
Sleep 3s

# Show date organization
Type "echo '' && echo '📅 Date-based organization:'"
Enter
Sleep 1s

Type "echo 'Today\\'s notes:'"
Enter
Type "find ~/code/wesen/obsidian-vault -name '*.md' -path '*$(date +%Y-%m-%d)*' | wc -l"
Enter
Sleep 1s

Type "echo 'January 5th notes:'"
Enter
Type "find ~/code/wesen/obsidian-vault -name '*.md' -path '*2024-01-05*'"
Enter
Sleep 2s

# Demonstrate append mode
Type "echo '' && echo '📝 Demonstrating append mode...'"
Enter
Sleep 1s

Type "./add-research --append --title 'Machine Learning Study' --message 'Additional findings: Transfer learning shows 40% improvement' --no-links"
Enter
Sleep 1s

Type "echo '📄 Appended content preview:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Machine-Learning-Study*' -exec tail -10 {} \\;"
Enter
Sleep 3s

# Show full directory tree structure
Type "echo '' && echo '🌳 Complete vault structure:'"
Enter
Type "find ~/code/wesen/obsidian-vault -type f -name '*.md' | head -10"
Enter
Sleep 3s

Type "echo '✅ Types & Organization demo completed!'"
Enter
Sleep 2s

Screenshot demos/demo-types.txt
