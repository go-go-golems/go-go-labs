# Cross-Referencing Enhancements

This document describes the rich SQL cross-referencing features added to the PR History & Code Browser.

## Overview

The application now leverages extensive SQL JOINs and cross-references to provide deep insights into the codebase history. Instead of just showing raw data, it now reveals relationships between commits, files, PRs, symbols, and analysis notes.

## Enhanced Data Models

### 1. **Commits with PR Associations** (`CommitWithRefsAndPRs`)
- Shows which PRs reference each commit
- Includes analysis notes linked to the commit
- Displays PR names and actions (port, docs, refactor, etc.)

**API:** `GET /api/commits/{hash}`

**Response includes:**
```json
{
  "commit": {...},
  "files": [...],
  "symbols": [...],
  "pr_associations": [
    {"pr_id": 3, "pr_name": "PR03-tool-executor", "action": "port"}
  ],
  "notes": [...]
}
```

### 2. **PRs with Referenced Commits and Files** (`PRWithDetails`)
- Changelog entries now include full commit and file objects
- Analysis notes include their referenced commits and files
- Clickable links to related entities

**API:** `GET /api/prs/{id}`

**Changelog entries now include:**
```json
{
  "action": "port",
  "details": "Brought over BaseToolExecutor...",
  "commit": {
    "hash": "b21e6f91",
    "subject": "Add tool executor",
    "author_name": "...",
    ...
  },
  "file": {
    "path": "pkg/inference/tools/base.go"
  }
}
```

### 3. **Files with History and Related Files** (`FileWithHistory`)
- Complete commit history for each file
- Files often changed together (co-change analysis)
- Analysis notes for the file
- Total commit count

**API:** `GET /api/files/{id}/details`

**Response includes:**
```json
{
  "id": 123,
  "path": "pkg/events/chat-events.go",
  "commit_count": 45,
  "recent_commits": [...],
  "related_files": [
    {"path": "pkg/events/registry.go", "change_count": 32},
    {"path": "pkg/chat/handler.go", "change_count": 18}
  ],
  "notes": [...]
}
```

### 4. **Symbol History** (`SymbolHistory`)
- Track how symbols (functions, types, classes) evolve
- See all commits that touched a specific symbol
- View symbols across different files

**API:** `GET /api/symbols/history?symbol=DebugTap&limit=50`

**Response:**
```json
[
  {
    "symbol_name": "DebugTap",
    "symbol_kind": "type",
    "file_path": "pkg/debug/tap.go",
    "commits": [...]
  }
]
```

### 5. **Symbol Search**
- Find symbols by pattern matching
- Discover where symbols are defined

**API:** `GET /api/symbols/search?q=Handler&limit=100`

## SQL Query Enhancements

### Co-Change Analysis
```sql
-- Files often changed together with a specific file
SELECT f.id, f.path, COUNT(DISTINCT cf1.commit_id) as change_count
FROM commit_files cf1
JOIN commit_files cf2 ON cf1.commit_id = cf2.commit_id AND cf2.file_id = ?
JOIN files f ON cf1.file_id = f.id
WHERE cf1.file_id != ?
GROUP BY f.id, f.path
ORDER BY change_count DESC
```

### PR-Commit Associations
```sql
-- Find all PRs that reference a commit
SELECT DISTINCT p.id, p.name, pcl.action
FROM pr_changelog pcl
JOIN prs p ON pcl.pr_id = p.id
WHERE pcl.commit_id = ?
```

### Symbol Evolution
```sql
-- Track symbol changes across commits
SELECT c.*
FROM commits c
JOIN commit_symbols cs ON c.id = cs.commit_id
JOIN files f ON cs.file_id = f.id
WHERE cs.symbol_name = ? AND f.path = ?
ORDER BY c.committed_at DESC
```

### Enriched Changelog
```sql
-- PR changelog with full commit and file details
SELECT pcl.*, c.hash, c.subject, c.author_name, f.path
FROM pr_changelog pcl
LEFT JOIN commits c ON pcl.commit_id = c.id
LEFT JOIN files f ON pcl.file_id = f.id
WHERE pcl.pr_id = ?
```

## UI Enhancements

### Commit Detail View
- **NEW:** Related PRs section showing which PRs used this commit
- **NEW:** Analysis notes section with tags and timestamps
- **Clickable PR badges** that navigate to PR detail pages

### PR Detail View
- **ENHANCED:** Changelog entries show commit hashes (clickable)
- **ENHANCED:** File paths in changelog are highlighted
- **ENHANCED:** Notes show referenced commit and file information

### File Detail View (Coming Soon)
- Commit history timeline
- Related files that change together
- Analysis notes specific to the file

## Benefits

1. **Traceability**: Follow work from commit → PR → documentation
2. **Impact Analysis**: See which files are often changed together
3. **Code Evolution**: Track how symbols and APIs evolve over time
4. **Context**: Understand why changes were made through linked notes
5. **Navigation**: Click through related entities easily

## Example Workflows

### "Where was this feature implemented?"
1. Search for symbol name (e.g., `ToolExecutor`)
2. View symbol history to see all commits
3. Click commit to see PR associations
4. Navigate to PR to see full context and notes

### "What files should I check when modifying X?"
1. Go to file detail for X
2. View "related files" section
3. See files often changed together
4. Review their recent commits

### "What commits went into this PR?"
1. View PR detail
2. See enriched changelog with commit references
3. Click commit hashes to view full details
4. See files changed in each commit

## Technical Details

### Performance Considerations
- All queries use indexed columns (commit_id, file_id, pr_id)
- LEFT JOINs handle missing references gracefully
- Pagination limits prevent large result sets
- Read-only database mode for safety

### Data Integrity
- Nullable foreign keys (commit_id, file_id) allow flexible associations
- SQL.NullInt64/NullString handle missing data
- JSON serialization handles nested structures

## Future Enhancements

Potential additions:
- **Timeline view**: Visual timeline of commits, PRs, and notes
- **Graph view**: Dependency graph of related files
- **Diff view**: Show actual code changes inline
- **Symbol search**: Full-text search across all symbols
- **PR templates**: Generate PR descriptions from commits
- **Conflict prediction**: Warn about files likely to conflict

## API Reference

| Endpoint | Description | Cross-References |
|----------|-------------|------------------|
| `GET /api/commits/{hash}` | Commit details | PR associations, notes |
| `GET /api/prs/{id}` | PR details | Commits, files, notes |
| `GET /api/files/{id}/details` | File details | History, related files, notes |
| `GET /api/symbols/history` | Symbol evolution | Commits, files |
| `GET /api/symbols/search` | Search symbols | Files |

All endpoints return JSON with rich nested objects where appropriate.

