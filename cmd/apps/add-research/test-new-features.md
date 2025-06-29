# Testing New Features of add-research

## Test Cases

### 1. Test Enhanced Link Behavior
```bash
# Default behavior - should ask for links interactively
echo "Test content for default links behavior" | ./add-research --title "Test Default Links" --log-level debug

# Skip link prompting by providing links
echo "Test content with provided links" | ./add-research --title "Test Provided Links" --links "https://example.com" "https://github.com/test/repo" --log-level debug

# Disable links entirely
echo "Test content without links" | ./add-research --title "Test No Links" --no-links --log-level debug
```

### 2. Test Enhanced Metadata
```bash
# Create note with enhanced metadata
echo "Test content for metadata" | ./add-research --title "Metadata Test Note" --metadata --no-links --log-level debug
```

### 3. Test Export Functionality
```bash
# Export all notes
./add-research --export --log-level debug

# Export with date range
./add-research --export --export-from "2024-12-01" --export-to "2024-12-31" --export-path "december-notes.md" --log-level debug
```

### 4. Test Enhanced Search
```bash
# Search with enhanced information
./add-research --search --log-level debug
```

## Expected Results

### Link Behavior Changes:
1. **Default**: Should prompt for links interactively
2. **With --links**: Should skip prompting and use provided links
3. **With --no-links**: Should skip links entirely

### Enhanced Metadata:
- Auto-generated ID/slug
- Word count (initially 0, can be updated)
- Source information
- Enhanced YAML frontmatter

### Export Features:
- Combined markdown file with all notes
- Date range filtering
- Proper header with generation info

### Search Improvements:
- File size and word count in results
- Content preview (first few lines)
- Better formatting

## Test Notes

All tests should be run from the cmd/apps/add-research directory with the vault structure:
```
~/code/wesen/obsidian-vault/research/YYYY-MM-DD/NNN-title.md
```
