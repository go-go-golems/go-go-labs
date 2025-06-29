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
Type "./add-research --type 'research' --title 'Machine Learning Study' --no-links --message 'Deep learning research: neural networks, training strategies, optimization techniques for AI applications.'"
Enter
Sleep 1s

# Ideas note
Type "./add-research --type 'ideas' --title 'Mobile App Concept' --no-links --message 'Cross-platform productivity app idea: task management, calendar integration, offline sync capabilities.'"
Enter
Sleep 1s

# Meeting notes
Type "./add-research --type 'meetings' --title 'Team Sync January' --no-links --message 'Sprint planning meeting: authentication system, database optimization, monitoring infrastructure deployment.'"
Enter
Sleep 1s

# Technical note with custom date
Type "./add-research --type 'technical' --date '2024-01-05' --title 'Database Migration Strategy' --no-links --message 'Zero-downtime migration plan: read replicas, dual-write pattern, data validation steps.'"
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

Type "./add-research --type 'research' --metadata --title 'Important Research Project' --no-links --message 'Critical findings with structured metadata: performance optimization, security practices, scalability patterns.'"
Enter
Sleep 2s

Type "echo '📄 Metadata example:'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Important-Research-Project*' -exec head -15 {} \\;"
Enter
Sleep 3s

# Show date organization
Type "echo '' && echo '📅 Date-based organization:'"
Enter
Sleep 1s

Type "echo 'Today\\'s notes count:'"
Enter
Type "find ~/code/wesen/obsidian-vault -name '*.md' -path '*$(date +%Y-%m-%d)*' | wc -l"
Enter
Sleep 1s

Type "echo 'January 5th, 2024 notes:'"
Enter
Type "find ~/code/wesen/obsidian-vault -name '*.md' -path '*2024-01-05*'"
Enter
Sleep 2s

# Demonstrate append mode
Type "echo '' && echo '📝 Demonstrating append mode...'"
Enter
Sleep 1s

Type "./add-research --append --title 'Machine Learning Study' --no-links --message 'Additional findings: transfer learning 40% improvement, preprocessing optimization, ensemble techniques boost accuracy.'"
Enter
Sleep 2s

Type "echo '📄 Appended content preview (last 8 lines):'"
Enter
Type "find ~/code/wesen/obsidian-vault/research -name '*Machine-Learning-Study*' -exec tail -8 {} \\;"
Enter
Sleep 3s

# Show different types with custom organization
Type "echo '' && echo '🗂️  Type-based organization benefits:'"
Enter
Sleep 1s

Type "echo '• Research: Structured investigations and findings'"
Enter
Type "echo '• Ideas: Creative concepts and future projects'"
Enter  
Type "echo '• Meetings: Collaboration records and decisions'"
Enter
Type "echo '• Technical: Implementation details and procedures'"
Enter
Sleep 2s

# Show complete directory structure overview
Type "echo '' && echo '🌳 Vault structure overview:'"
Enter
Type "find ~/code/wesen/obsidian-vault -type f -name '*.md' | head -8 | wc -l"
Enter
Type "echo 'Total notes created in demo'"
Enter
Sleep 2s

# Show how to work with different types  
Type "echo '' && echo '⚡ Quick type-based workflows:'"
Enter
Type "echo 'Research: add-research --type research --metadata'"
Enter
Type "echo 'Ideas: add-research --type ideas --title New-Concept'"
Enter
Type "echo 'Meetings: add-research --type meetings --date 2024-01-15'"
Enter
Type "echo 'Technical: add-research --type technical --append'"
Enter
Sleep 3s

# Show organization by date and type
Type "echo '' && echo '📊 Organization summary:'"
Enter
Type "echo 'Notes are organized by:'"
Enter
Type "echo '1. Type (research, ideas, meetings, technical)'"
Enter
Type "echo '2. Date (YYYY-MM-DD folders)'"
Enter
Type "echo '3. Sequential numbering within each day'"
Enter
Type "echo '4. Metadata for enhanced searchability'"
Enter
Sleep 3s

Type "echo '✅ Types & Organization demo completed!'"
Enter
Sleep 2s

Screenshot demos/demo-types.txt
