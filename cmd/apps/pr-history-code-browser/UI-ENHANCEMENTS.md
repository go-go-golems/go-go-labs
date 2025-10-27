# UI Enhancements for Cross-Referenced Data

This document describes the frontend enhancements that leverage the rich cross-referencing capabilities of the backend.

## Overview

The UI has been significantly enhanced to display and navigate between related entities (commits, PRs, files, and notes). Users can now see the full context of their work and easily navigate between related items.

## Enhanced Pages

### 1. PR Detail Page (`/prs/:id`)

**New Features:**
- **Enriched Changelog Entries**: Each changelog entry now displays:
  - Action badge (port, docs, refactor, etc.)
  - Action details
  - **Referenced Commit** (if present):
    - Clickable commit hash (links to commit detail)
    - Commit subject line
    - Author name and date
    - Styled with a blue-bordered box
  - **Referenced File** (if present):
    - File path displayed in monospace
    - Green color coding with file emoji
  - Timestamp of the changelog entry

- **Enhanced Analysis Notes**: Notes now show:
  - Note type and content
  - **Related Commit** (if present):
    - Clickable commit hash
    - Commit subject
  - **Related File** (if present):
    - File path in monospace
  - Tags with badge styling
  - Timestamp

**Example View:**
```
PR03-tool-executor
├── Changelog
│   ├── Action: port
│   │   ├── Details: "Brought over BaseToolExecutor..."
│   │   ├── Commit: b21e6f91 - Add tool executor
│   │   │   by John Doe • 2025-10-15
│   │   └── File: pkg/inference/tools/base.go
│   └── ...
└── Notes
    ├── Implementation note
    │   ├── Related to commit: b21e6f91 - Add tool executor
    │   └── Tags: [architecture, tools]
    └── ...
```

### 2. Commit Detail Page (`/commits/:hash`)

**New Features:**
- **Related PRs Section**: Shows all PRs that reference this commit
  - PR name (clickable - navigates to PR detail)
  - Action badge (how the PR used this commit)
  - Blue-bordered, clickable boxes
  - Click to navigate to PR detail page

- **Analysis Notes Section**: Shows notes linked to this commit
  - Note type and content
  - Tags with badge styling
  - Timestamp

**Example View:**
```
Commit: b21e6f91abc123
├── Subject: Add BaseToolExecutor implementation
├── Related PRs (2)
│   ├── PR03-tool-executor [port] ← Click to view PR
│   └── PR05-docs-update [docs] ← Click to view PR
├── Analysis Notes (1)
│   └── "Initial implementation of tool executor pattern"
│       Tags: [architecture, tools]
├── Files Changed (3)
│   └── ...
└── Symbols (5)
    └── ...
```

### 3. NEW: File Detail Page (`/files/:id`)

**Completely New Page** showing:

- **File Overview**:
  - Full file path (monospace with file emoji)
  - Total commit count

- **Recent Commits**: List of recent commits that modified this file
  - Clickable commit hash
  - Commit subject
  - Author and date
  - Links to commit detail page

- **Files Often Changed Together**: Co-change analysis
  - Shows files frequently modified in the same commits
  - Change count (number of co-occurrences)
  - Helps identify related code areas
  - Useful for impact analysis

- **Analysis Notes**: Notes specific to this file
  - Note type and content
  - Related commit (if present) with clickable link
  - Tags
  - Timestamp

**Example View:**
```
File: pkg/events/chat-events.go
├── Total commits: 45
├── Recent Commits
│   ├── a1b2c3d4 - Add message queuing
│   ├── e5f6g7h8 - Refactor event handlers
│   └── ...
├── Files Often Changed Together
│   ├── pkg/events/registry.go (32 co-changes)
│   ├── pkg/chat/handler.go (18 co-changes)
│   └── pkg/models/message.go (12 co-changes)
└── Analysis Notes
    └── "Central event definitions for chat system"
```

### 4. Files List Page (`/files`)

**Enhanced Features:**
- File paths are now **clickable**
- Click any file to navigate to its detail page
- Retains existing search/filter functionality

## Navigation Flow Examples

### Finding What Went Into a PR

1. Navigate to **PRs** page
2. Click on a PR (e.g., "PR03-tool-executor")
3. See **Changelog** with:
   - All commits ported into this PR
   - Click any commit hash to see full commit details
   - See which files were referenced in changelog entries
4. View **Analysis Notes** with their related commits/files

### Tracking a Commit's Usage

1. Navigate to **Commits** page
2. Click on a commit hash
3. See **Related PRs** section showing:
   - Which PRs included this commit
   - What action they took (port, docs, refactor)
4. Click any PR to see the full PR context

### Understanding File Changes

1. Navigate to **Files** page
2. Search or browse for a file
3. Click the file path
4. See:
   - Complete commit history
   - **Related files** that change together
   - Analysis notes about the file

### Impact Analysis Workflow

1. Go to file detail for file you're about to modify
2. Check **"Files Often Changed Together"** section
3. Review those files for potential impact
4. Check their recent commits
5. Read analysis notes for context

## Visual Styling

### Color Coding
- **Blue** (#3498db): Commit references and PR associations
- **Green** (#27ae60): File paths and file-related content
- **Orange/Amber**: Action badges
- **Gray** (#7f8c8d): Secondary information (dates, metadata)
- **Light Gray** (#f8f9fa): Background for nested content

### Interactive Elements
- Commit hashes are clickable links (blue, monospace)
- PR names are clickable (styled as badges)
- File items are clickable (hover effect)
- Related PR boxes are clickable with hover pointer

### Information Hierarchy
- Primary content (subjects, titles) in larger, bold text
- Referenced entities in bordered boxes
- Metadata in smaller, gray text
- Tags as small colored badges

## Technical Implementation

### Type Definitions (`frontend/src/types/index.ts`)
```typescript
// Enriched changelog with commit and file objects
export interface PRChangelog {
  // ... existing fields ...
  commit?: Commit;  // Full commit object
  file?: File;      // Full file object
}

// Enriched notes with commit and file objects
export interface AnalysisNote {
  // ... existing fields ...
  commit?: Commit;  // Full commit object
  file?: File;      // Full file object
}

// New: File with complete history and relationships
export interface FileWithHistory {
  id: number;
  path: string;
  commit_count: number;
  recent_commits: Commit[];
  related_files: RelatedFile[];
  notes: AnalysisNote[];
}

export interface RelatedFile {
  path: string;
  change_count: number;
}

// Commit details with PR associations
export interface CommitDetails {
  commit: Commit;
  files: FileChange[];
  symbols: CommitSymbol[];
  pr_associations?: PRAssociation[];  // New
  notes?: AnalysisNote[];             // New
}
```

### API Client (`frontend/src/api/client.ts`)
```typescript
// New endpoint
async getFileDetails(fileId: number): Promise<FileWithHistory> {
  return fetchJSON<FileWithHistory>(`${API_BASE}/files/${fileId}/details`);
}
```

### Routing (`frontend/src/App.tsx`)
```typescript
<Route path="files/:id" element={<FileDetailPage />} />  // New route
```

## Benefits

1. **Traceability**: Follow code changes from commit → PR → notes
2. **Context**: Understand *why* changes were made
3. **Navigation**: Quickly jump between related entities
4. **Impact Analysis**: See which files change together
5. **Discoverability**: Find related work easily
6. **Documentation**: View analysis notes in context

## User Scenarios

### Scenario 1: "What PRs used this commit?"
- Navigate to commit detail
- See "Related PRs" section
- Click any PR to view full context

### Scenario 2: "What commits are in this PR?"
- Navigate to PR detail
- View changelog entries
- Each entry shows the commit hash, subject, and author
- Click commit hash to see full diff and changes

### Scenario 3: "What files should I check when modifying X?"
- Navigate to file X detail page
- Check "Files Often Changed Together"
- Review recent commits for those files
- Check analysis notes for warnings/patterns

### Scenario 4: "Where is symbol Y used?"
- Search commits for symbol Y (existing feature)
- View commit detail
- See which PRs referenced this commit
- Navigate to PRs to understand usage context

## Future Enhancements

Potential additions based on this foundation:
- Symbol search page with results showing symbol history
- Timeline visualization of commits, PRs, and notes
- Graph view of file relationships
- Inline code diffs in commit detail
- PR template generation from commit history
- Conflict prediction based on file co-change patterns

## Testing Checklist

- [ ] PR detail page shows enriched changelog with commits and files
- [ ] Commit hashes in PR changelog are clickable
- [ ] Commit detail page shows related PRs section
- [ ] PR badges in commit detail are clickable
- [ ] Files list has clickable file paths
- [ ] File detail page shows complete information
- [ ] Co-changed files display correctly
- [ ] All navigation links work correctly
- [ ] Styling is consistent across pages
- [ ] Loading states work properly
- [ ] Error states are handled gracefully

