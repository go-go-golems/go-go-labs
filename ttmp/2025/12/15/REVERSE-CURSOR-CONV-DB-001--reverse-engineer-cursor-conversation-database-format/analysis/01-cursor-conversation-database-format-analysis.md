---
Title: Cursor Conversation Database Format Analysis
Ticket: REVERSE-CURSOR-CONV-DB-001
Status: active
Topics:
    - reverse-engineering
    - data-analysis
    - exploration
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/.config/Cursor/User/globalStorage/state.vscdb
      Note: Primary SQLite database containing Cursor conversation and UI state (2.1GB)
    - Path: /home/manuel/.config/Cursor/User/workspaceStorage/
      Note: Per-workspace state databases
    - Path: /home/manuel/.config/Cursor/User/globalStorage/sourcegraph.amp/
      Note: Sourcegraph AMP extension conversation storage (not native Cursor)
ExternalSources: []
Summary: Comprehensive analysis of Cursor's actual conversation storage architecture based on reverse engineering of ~/.config/Cursor directory structure. Documents VS Code-compatible storage system using state.vscdb key-value database.
LastUpdated: 2025-12-15T09:25:12.623187787-05:00
---

# Cursor Conversation Database Format Analysis

**Conversation UUID Anchor:** `aa8ad79b-e459-4989-93dd-5fbf136d08d0`  
**Analysis Date:** 2025-12-15  
**Exploration Method:** Reverse engineering of local storage directories  
**Correction Note:** Initial analysis incorrectly focused on `hooks.db` which was a user experiment. This document reflects actual Cursor storage.

## Executive Summary

Cursor stores conversation history using **VS Code's storage API** in SQLite key-value databases. **Conversations are stored per-workspace** in `workspaceStorage/{workspace-uuid}/state.vscdb`, not globally. Each conversation is identified by a `composerId` (UUID) stored in `composer.composerData.allComposers[]`. Prompts are stored in `aiService.generations[]` and `aiService.prompts[]` arrays, with each generation having a `generationUUID` that matches the `generation_id` in hook logs. The system uses VS Code-compatible storage mechanisms, allowing integration with VS Code's extension ecosystem.

**Important:** The UUID `aa8ad79b-e459-4989-93dd-5fbf136d08d0` used as a conversation anchor is **NOT** a conversation identifier - it's just text content in one of the prompts. The actual conversation UUID is `59f64e5e-70e9-4892-a9e3-69d3d7f4b42d` (the `composerId`).

**Key Findings:**
- **Per-workspace storage:** `~/.config/Cursor/User/workspaceStorage/{workspace-uuid}/state.vscdb` (primary conversation storage)
- **Global storage:** `~/.config/Cursor/User/globalStorage/state.vscdb` (2.1GB, UI state and settings)
- **Data model:** Key-value store with JSON values
- **Conversation keys:** `aiService.generations`, `aiService.prompts`, `composer.composerData`
- **Conversation UUID:** Stored as `composerId` in `composer.composerData.allComposers[]`
- **Generation UUID:** Stored as `generationUUID` in `aiService.generations[]`
- **View state:** `workbench.panel.composerChatViewPane.{uuid}.hidden` keys (1,368 entries in global storage)
- **Structure:** Per-workspace arrays of generations/prompts + metadata object

---

## Storage Architecture Overview

### Primary Storage Locations

**Per-Workspace Storage (Conversations):**
- **Path:** `~/.config/Cursor/User/workspaceStorage/{workspace-uuid}/state.vscdb`
- **Type:** SQLite 3.x database (VS Code storage format)
- **Purpose:** Stores conversation data for each workspace
- **Schema:** Key-value store (`ItemTable` with `key` TEXT and `value` BLOB columns)
- **Key Keys:** `aiService.generations`, `aiService.prompts`, `composer.composerData`

**Global Storage (UI State):**
- **Path:** `~/.config/Cursor/User/globalStorage/state.vscdb`
- **Type:** SQLite 3.x database (VS Code storage format)
- **Size:** ~2.1GB
- **SQLite Version:** 3039004
- **Encoding:** UTF-8
- **Schema:** Key-value store (`ItemTable` with `key` TEXT and `value` BLOB columns)
- **Purpose:** Stores UI state, settings, and view configurations

### Database Structure

```sql
CREATE TABLE ItemTable (
    key TEXT PRIMARY KEY,
    value BLOB
)
```

**Key Characteristics:**
- Simple key-value store pattern
- Values stored as BLOB (typically JSON strings)
- Used by VS Code and Cursor for extension state persistence
- 1,664 total keys found in global storage

### Complementary Storage Locations

1. **`~/.config/Cursor/User/workspaceStorage/{workspace-uuid}/`** - Per-workspace storage
   - Each workspace has its own `state.vscdb` database
   - Contains workspace-specific state and settings
   - May contain workspace-specific conversation data

2. **`~/.config/Cursor/User/globalStorage/sourcegraph.amp/`** - Extension storage
   - `threads3-development/` - Sourcegraph AMP extension threads
   - Contains `T-{uuid}.json` files with conversation threads
   - **Note:** This is extension storage, not native Cursor storage

3. **`~/.config/Cursor/User/History/`** - File edit history
   - 9,600+ directories (one per file)
   - `entries.json` files tracking file edit history
   - Separate from conversation storage

4. **`~/.cursor/`** - Cursor-specific configuration
   - `hooks.json` - Hook configuration (user experiments)
   - `projects/` - Per-workspace execution logs
   - `ide_state.json` - Recently viewed files

---

## Conversation Storage Format

### Per-Workspace Storage (Primary)

**Location:** `~/.config/Cursor/User/workspaceStorage/{workspace-uuid}/state.vscdb`  
**Key Storage Pattern:** Multiple keys store different aspects of conversations

### Key Storage Keys

1. **`aiService.generations`** - Generation/prompt records
2. **`aiService.prompts`** - Simplified prompt objects  
3. **`composer.composerData`** - Conversation metadata
4. **`workbench.panel.composerChatViewPane.{viewId}`** - View state references

### aiService.generations

**Format:** JSON array  
**Structure:**

```json
[
  {
    "unixMs": 1765808517629,
    "generationUUID": "3bcfa52b-ec65-4a91-83e5-e7663e3cc005",
    "type": "composer",
    "textDescription": "prompt text here..."
  },
  ...
]
```

**Fields:**
- `unixMs`: Timestamp in milliseconds
- `generationUUID`: Unique identifier for this generation (matches `generation_id` in hook logs)
- `type`: Generation type (e.g., "composer", "chat")
- `textDescription`: Full prompt text

**Purpose:** Stores all prompts/generations submitted in this workspace.

### aiService.prompts

**Format:** JSON array  
**Structure:**

```json
[
  {
    "text": "prompt text here...",
    "commandType": 4
  },
  ...
]
```

**Fields:**
- `text`: Prompt text
- `commandType`: Numeric command type code

**Purpose:** Simplified prompt storage (parallel to generations).

### composer.composerData

**Format:** JSON object  
**Structure:**

```json
{
  "allComposers": [
    {
      "type": "head",
      "composerId": "59f64e5e-70e9-4892-a9e3-69d3d7f4b42d",
      "name": "REVERSE-CURSOR-001 - initial research",
      "lastUpdatedAt": 1765809716205,
      "createdAt": 1765808373121,
      "unifiedMode": "agent",
      "forceMode": "edit",
      "hasUnreadMessages": false,
      "contextUsagePercent": 70.77249908447266,
      "totalLinesAdded": 1349,
      "totalLinesRemoved": 2,
      "filesChangedCount": 2,
      "subtitle": "01-cursor-conversation-database-format-analysis.md, 01-exploration-diary.md",
      "hasBlockingPendingActions": false,
      "isArchived": false,
      "isDraft": false,
      "isWorktree": false,
      "isSpec": false,
      "isBestOfNSubcomposer": false,
      "numSubComposers": 0,
      "referencedPlans": []
    }
  ],
  "selectedComposerIds": ["59f64e5e-70e9-4892-a9e3-69d3d7f4b42d"],
  "lastFocusedComposerIds": ["59f64e5e-70e9-4892-a9e3-69d3d7f4b42d"],
  "hasMigratedComposerData": false,
  "hasMigratedMultipleComposers": true
}
```

**Key Fields:**
- `composerId`: **This is the conversation UUID!** (e.g., `59f64e5e-70e9-4892-a9e3-69d3d7f4b42d`)
- `name`: Conversation name/title
- `createdAt`, `lastUpdatedAt`: Timestamps
- `totalLinesAdded`, `totalLinesRemoved`: Edit statistics
- `filesChangedCount`: Number of files changed
- `subtitle`: Summary of changed files
- `contextUsagePercent`: Context usage percentage
- `unifiedMode`: Mode ("agent" or "chat")
- `forceMode`: Force mode setting
- Status flags: `isArchived`, `isDraft`, etc.

**Purpose:** Stores conversation metadata and statistics. The `composerId` is the actual conversation UUID.

### Global Chat Data Key

**Key:** `workbench.panel.aichat.view.aichat.chatdata` (in globalStorage)  
**Format:** JSON object  
**Structure:**

```json
{
  "tabs": [
    {
      "tabId": "uuid",
      "tabState": "chat",
      "bubbles": [
        {
          "type": "user",
          "id": "uuid",
          "messageType": 2,
          "rawText": "",
          "selections": [],
          "isFocused": true,
          "contextCacheTimestamp": 1721229087235,
          "waitingForContext": false,
          "longFilesFitTimestamp": 1721229087232,
          "waitingForLongFilesFit": true,
          "dropdownAdvancedCodebaseSearchBehavior": "embeddings",
          "delegate": {
            "a": "",
            "c": {},
            "d": {},
            "e": {},
            "f": {},
            "h": {}
          },
          "followup": false
        }
      ],
      "longContextModeEnabled": false,
      "lastFocusedBubble": "uuid"
    }
  ],
  "codeInterpreterTabs": [],
  "selectedTabId": "uuid",
  "displayTabs": false,
  "editorContext": {
    "isNotebook": false,
    "hasNonemptySelection": false
  },
  "debugPromptVisible": false,
  "pinnedContexts": {
    "fileSelections": [],
    "codeSelections": []
  }
}
```

### Data Model

### Per-Workspace Conversation Model

**Conversation Structure:**
- **Composer Metadata** (`composer.composerData.allComposers[]`):
  - `composerId`: Conversation UUID (primary identifier)
  - `name`: Conversation name/title
  - `createdAt`, `lastUpdatedAt`: Timestamps
  - `totalLinesAdded`, `totalLinesRemoved`: Edit statistics
  - `filesChangedCount`: Number of files changed
  - `subtitle`: Summary of changed files
  - `contextUsagePercent`: Context usage
  - `unifiedMode`: Mode ("agent" or "chat")
  - `forceMode`: Force mode setting
  - Status flags: `isArchived`, `isDraft`, etc.

- **Generations** (`aiService.generations[]`):
  - `generationUUID`: Unique identifier for this generation
  - `type`: Generation type ("composer", "chat")
  - `textDescription`: Full prompt text
  - `unixMs`: Timestamp

- **Prompts** (`aiService.prompts[]`):
  - `text`: Prompt text
  - `commandType`: Numeric command type

**Linking:**
- `generationUUID` in `aiService.generations` matches `generation_id` in hook logs
- `composerId` in `composer.composerData` matches `conversation_id` in hook logs
- Generations and prompts are stored in parallel arrays (same length, same order)

### Global Chat Data Model (UI State)

**Conversation Structure:**
- **Tabs:** Array of conversation tabs
  - `tabId`: Unique identifier (UUID) for the tab
  - `tabState`: State of the tab (e.g., "chat")
  - `bubbles`: Array of message bubbles
  - `lastFocusedBubble`: ID of the currently focused bubble
  - `longContextModeEnabled`: Boolean flag

- **Bubbles (Messages):**
  - `type`: Message type (e.g., "user", "assistant")
  - `id`: Unique identifier (UUID) for the bubble
  - `messageType`: Numeric message type code
  - `rawText`: Message text content
  - `selections`: Array of code selections
  - `isFocused`: Whether this bubble is currently focused
  - `contextCacheTimestamp`: Timestamp for context caching
  - `waitingForContext`: Boolean flag
  - `longFilesFitTimestamp`: Timestamp for long file fitting
  - `waitingForLongFilesFit`: Boolean flag
  - `dropdownAdvancedCodebaseSearchBehavior`: Search behavior setting
  - `delegate`: Delegate object with various properties
  - `followup`: Boolean flag for follow-up messages

- **Metadata:**
  - `selectedTabId`: Currently selected tab ID
  - `displayTabs`: Whether tabs are displayed
  - `editorContext`: Editor context information
  - `debugPromptVisible`: Debug prompt visibility flag
  - `pinnedContexts`: Pinned context selections

### Composer Chat View Panes

**Key Pattern:** `workbench.panel.composerChatViewPane.{uuid}.hidden`  
**Count:** 1,368 entries found  
**Format:** JSON array  
**Purpose:** Tracks view state for composer chat panes

**Structure:**
```json
[
  {
    "id": "workbench.panel.aichat.view.{uuid}",
    "isHidden": false
  },
  ...
]
```

**Observations:**
- Each composerChatViewPane entry contains an array of `aichat.view` references
- References point to `workbench.panel.aichat.view.{uuid}` IDs
- These IDs don't exist as separate keys in the database
- Likely used for UI state management and view tracking

### Background Composer

**Key:** `workbench.backgroundComposer.persistentData`  
**Format:** JSON object  
**Structure:**
```json
{
  "showControlPanel": false,
  "dataVersion": 1,
  "lastOpenedBcIds": {},
  "archivedBcIds": [],
  "selectedTab": "personal",
  "isBackgroundComposerEnabled": true
}
```

**Purpose:** Stores background composer state and settings.

---

## Storage Patterns

### Key Naming Conventions

1. **Chat Data:**
   - `workbench.panel.aichat.view.aichat.chatdata` - Main chat data
   - `workbench.panel.aichat.hidden` - Chat panel visibility

2. **Composer View Panes:**
   - `workbench.panel.composerChatViewPane.{uuid}.hidden` - View pane state
   - `workbench.panel.composerChatViewPane.hidden` - General pane visibility

3. **Background Composer:**
   - `workbench.backgroundComposer.persistentData` - Persistent data
   - `backgroundComposer.windowBcMapping` - Window mappings

4. **Settings:**
   - `composer.hasReopenedOnce` - Composer state flag
   - `cursor/composerAutocompleteHeuristicsAutoApplied` - Autocomplete settings
   - `cursor/composerAutocompleteHeuristicsEnabled` - Autocomplete enable flag

### Data Persistence

**Storage Mechanism:**
- Uses VS Code's storage API (`vscode.StateStorage`)
- Data persisted to SQLite database
- Key-value pattern allows flexible extension storage
- Values stored as JSON strings (BLOB in database)

**Persistence Timing:**
- Data written when state changes
- May be cached in memory before persistence
- Active conversations may not be immediately persisted

---

## Querying Conversations

### Finding Conversations in Workspace Storage

```sql
-- Get conversation metadata
SELECT value FROM ItemTable 
WHERE key = 'composer.composerData';

-- Get all generations/prompts
SELECT value FROM ItemTable 
WHERE key = 'aiService.generations';

-- Get simplified prompts
SELECT value FROM ItemTable 
WHERE key = 'aiService.prompts';
```

### Finding a Specific Conversation

```python
import sqlite3
import json

workspace_uuid = "94bae793ba109d83fb8934a587a6c719"
conversation_id = "59f64e5e-70e9-4892-a9e3-69d3d7f4b42d"

db_path = f"~/.config/Cursor/User/workspaceStorage/{workspace_uuid}/state.vscdb"
conn = sqlite3.connect(db_path)
cursor = conn.cursor()

# Get composer data
cursor.execute("SELECT value FROM ItemTable WHERE key = 'composer.composerData'")
composer_data = json.loads(cursor.fetchone()[0])

# Find conversation
conversation = next(
    (c for c in composer_data['allComposers'] if c['composerId'] == conversation_id),
    None
)

# Get generations
cursor.execute("SELECT value FROM ItemTable WHERE key = 'aiService.generations'")
generations = json.loads(cursor.fetchone()[0])

# Get prompts
cursor.execute("SELECT value FROM ItemTable WHERE key = 'aiService.prompts'")
prompts = json.loads(cursor.fetchone()[0])

print(f"Conversation: {conversation['name']}")
print(f"Generations: {len(generations)}")
print(f"Prompts: {len(prompts)}")
```

### Finding Chat Data (Global Storage)

```sql
-- Get main chat data (global)
SELECT value FROM ItemTable 
WHERE key = 'workbench.panel.aichat.view.aichat.chatdata';
```

### Finding Composer View Panes

```sql
-- List all composer chat view panes
SELECT key FROM ItemTable 
WHERE key LIKE 'workbench.panel.composerChatViewPane.%'
ORDER BY key;
```

### Counting Conversations

**Per-Workspace:**
```sql
-- Count conversations in a workspace
SELECT value FROM ItemTable WHERE key = 'composer.composerData';
-- Then count allComposers array length
```

**Global Storage:**
```sql
-- Count composer chat view panes (proxy for conversation count)
SELECT COUNT(*) FROM ItemTable 
WHERE key LIKE 'workbench.panel.composerChatViewPane.%' 
AND key LIKE '%.hidden';
```

**Result:** 1,368 composer chat view panes found in global storage (suggests many conversations across all workspaces)

### Extracting Tab Information

Using Python to parse chatdata:

```python
import sqlite3
import json

conn = sqlite3.connect('~/.config/Cursor/User/globalStorage/state.vscdb')
cursor = conn.cursor()

cursor.execute("SELECT value FROM ItemTable WHERE key = 'workbench.panel.aichat.view.aichat.chatdata'")
row = cursor.fetchone()

if row:
    data = json.loads(row[0])
    tabs = data.get('tabs', [])
    print(f"Total tabs: {len(tabs)}")
    for i, tab in enumerate(tabs):
        print(f"Tab {i}: id={tab.get('tabId')}, bubbles={len(tab.get('bubbles', []))}")
```

---

## Workspace Storage

### Per-Workspace Databases

Each workspace has its own storage database:
- **Path:** `~/.config/Cursor/User/workspaceStorage/{workspace-uuid}/state.vscdb`
- **Structure:** Same key-value pattern as global storage
- **Purpose:** Workspace-specific state and settings

### Workspace Identification

Workspaces are identified by UUIDs derived from workspace paths:
- Hash-based UUID generation
- Consistent across sessions
- Allows per-workspace data isolation

---

## Extension Storage

### Sourcegraph AMP Extension

**Location:** `~/.config/Cursor/User/globalStorage/sourcegraph.amp/threads3-development/`  
**Format:** JSON files named `T-{uuid}.json`  
**Structure:**

```json
{
  "v": 493,
  "id": "T-3559cdf1-d074-4a69-9776-b201e7d67e4f",
  "created": 1754518600080,
  "messages": [
    {
      "role": "user",
      "content": [
        {
          "type": "text",
          "text": "..."
        }
      ],
      "userState": {
        "currentlyVisibleFiles": [...],
        "runningTerminalCommands": [],
        "activeEditor": "...",
        "cursorLocation": {...}
      },
      "meta": {
        "sentAt": 1754518600421
      }
    },
    {
      "role": "assistant",
      "content": [...],
      "state": {
        "type": "complete",
        "stopReason": "tool_use"
      },
      "usage": {...}
    }
  ]
}
```

**Note:** This is extension storage, not native Cursor storage. Cursor may use similar patterns internally.

---

## UUID Search Results

**Search Target:** `aa8ad79b-e459-4989-93dd-5fbf136d08d0` (anchor UUID)

**Discovery:** This UUID is **NOT** a conversation identifier! It's just text content in one of the prompts.

**Actual Conversation UUID:** `59f64e5e-70e9-4892-a9e3-69d3d7f4b42d` (the `composerId`)

**Search Methods:**
1. Key search: `SELECT key FROM ItemTable WHERE key LIKE '%aa8ad79b%'`
2. Value search: `SELECT value FROM ItemTable WHERE value LIKE '%aa8ad79b%'`
3. File system search: `grep -r "aa8ad79b" ~/.config/Cursor`
4. Found in: `aiService.generations[3].textDescription` (as text content)

**Results:**
- UUID found as text content in generation `b7e5b0b1-45f5-4c1d-bdfe-79250a20ff91`
- Not used as a conversation identifier
- Actual conversation identified by `composerId`: `59f64e5e-70e9-4892-a9e3-69d3d7f4b42d`

**Conversation Storage Location:**
- **Workspace:** `94bae793ba109d83fb8934a587a6c719`
- **Database:** `workspaceStorage/94bae793ba109d83fb8934a587a6c719/state.vscdb`
- **Keys:**
  - `aiService.generations` - Contains 6 generations
  - `aiService.prompts` - Contains 6 prompts
  - `composer.composerData` - Contains composer metadata with `composerId = 59f64e5e-70e9-4892-a9e3-69d3d7f4b42d`
  - `workbench.panel.composerChatViewPane.ecc9c419-097d-49e4-b521-95d311edd307` - View state

**Key Insight:** Conversations are stored **per-workspace**, not globally. The conversation UUID is the `composerId` in `composer.composerData.allComposers[]`.

---

## Comparison with VS Code Storage

### Similarities

- Uses same SQLite key-value store pattern
- Compatible with VS Code extension storage API
- Same database structure (`ItemTable`)
- Similar key naming conventions (`workbench.*`)

### Differences

- Cursor-specific keys: `workbench.panel.aichat.*`, `workbench.panel.composerChatViewPane.*`
- Chat-specific data structures (tabs, bubbles)
- Composer-specific state management
- Background composer integration

---

## Security and Privacy Considerations

### Data Storage

- All data stored locally in `~/.config/Cursor/`
- Database is unencrypted SQLite
- File paths may be stored in conversation context
- Code selections and file contents may be included

### User Identification

- No explicit user email fields found in chatdata
- Workspace paths may reveal user directory structure
- Conversation content may contain sensitive information

### Recommendations

- Treat `state.vscdb` as sensitive (contains conversation history)
- Review conversation content before sharing databases
- Consider encryption for sensitive workspaces
- Be aware of file paths and code content in conversations

---

## Limitations and Unknowns

### Unanswered Questions

1. **Cloud Sync:** Does Cursor sync conversations to cloud? Where?
2. **Conversation Persistence:** When are conversations written to disk?
3. **UUID Storage:** Where are conversation UUIDs stored?
4. **Message History:** How are full message histories stored?
5. **Per-Workspace Storage:** Are conversations stored per-workspace or globally?

### Data Gaps

- Current chatdata shows only 1 active tab
- Many composerChatViewPane entries but unclear relationship to conversations
- No direct conversation transcript storage found
- Message content structure not fully explored

---

## Tools and Utilities

### Useful SQL Queries

**List all chat-related keys:**
```sql
SELECT key FROM ItemTable 
WHERE key LIKE '%chat%' OR key LIKE '%composer%' OR key LIKE '%aichat%'
ORDER BY key;
```

**Get chat data:**
```sql
SELECT value FROM ItemTable 
WHERE key = 'workbench.panel.aichat.view.aichat.chatdata';
```

**Count composer view panes:**
```sql
SELECT COUNT(*) FROM ItemTable 
WHERE key LIKE 'workbench.panel.composerChatViewPane.%';
```

**Find largest entries:**
```sql
SELECT key, LENGTH(value) as size FROM ItemTable 
WHERE key LIKE 'workbench.panel.composerChatViewPane.%'
ORDER BY size DESC
LIMIT 10;
```

### Exporting Data

**Export chatdata to JSON:**
```sql
.mode json
.output chatdata.json
SELECT value FROM ItemTable WHERE key = 'workbench.panel.aichat.view.aichat.chatdata';
```

**Export all composer panes:**
```sql
.mode csv
.headers on
.output composer_panes.csv
SELECT key, LENGTH(value) as size FROM ItemTable 
WHERE key LIKE 'workbench.panel.composerChatViewPane.%';
```

---

## Future Research Directions

### Potential Enhancements

1. **Conversation Export Tool:** Script to extract and format conversations
2. **Database Browser:** GUI tool for exploring state.vscdb
3. **Search Interface:** Full-text search across conversations
4. **Analytics:** Conversation statistics and insights
5. **Backup/Restore:** Tools for backing up conversation history

### Research Questions

1. How are conversations linked to workspaces?
2. What is the relationship between tabs and composerChatViewPane entries?
3. How are message bubbles structured in detail?
4. Is there cloud sync for conversations?
5. How are conversation UUIDs generated and stored?

---

## Conclusion

Cursor uses VS Code's storage API for conversation persistence, storing chat data in a key-value SQLite database. The primary storage location is `state.vscdb` (2.1GB) containing chatdata and view state. Conversations are organized as tabs with message bubbles, stored as JSON structures. The system integrates with VS Code's extension ecosystem while providing Cursor-specific chat functionality.

**Key Takeaways:**
- **Per-workspace storage** for conversation data (primary)
- **Global storage** for UI state and settings
- Key-value store pattern with JSON values
- **Conversation UUID** = `composerId` in `composer.composerData.allComposers[]`
- **Generation UUID** = `generationUUID` in `aiService.generations[]`
- Conversations stored as arrays: `aiService.generations` and `aiService.prompts`
- Metadata stored separately in `composer.composerData`
- View state tracked in composerChatViewPane entries (global storage)
- Extension storage separate from native storage

**For Developers:**
- Use VS Code storage API patterns
- Query `state.vscdb` for conversation data
- Parse JSON structures for chat content
- Be aware of workspace-specific storage
- Consider cloud sync implications

---

## References

- Exploration Diary: `reference/01-exploration-diary.md`
- Database File: `~/.config/Cursor/User/globalStorage/state.vscdb`
- VS Code Storage API: VS Code extension API documentation
- Workspace Storage: `~/.config/Cursor/User/workspaceStorage/`

---

## Summary: Complete Conversation Storage Model

### Storage Hierarchy

```
Per-Workspace Storage (workspaceStorage/{uuid}/state.vscdb)
  ├── aiService.generations: [{
  │     generationUUID: "uuid" (matches generation_id in hooks)
  │     type: "composer" | "chat"
  │     textDescription: "prompt text"
  │     unixMs: timestamp
  │   }]
  ├── aiService.prompts: [{
  │     text: "prompt text"
  │     commandType: number
  │   }]
  ├── composer.composerData: {
  │     allComposers: [{
  │       composerId: "uuid" (THIS IS THE CONVERSATION UUID!)
  │       name: "conversation name"
  │       createdAt, lastUpdatedAt: timestamps
  │       totalLinesAdded, totalLinesRemoved: stats
  │       filesChangedCount: number
  │       ...
  │     }]
  │   }
  └── workbench.panel.composerChatViewPane.{viewId}: {
        "workbench.panel.aichat.view.{composerId}": {
          collapsed, isHidden, size
        }
      }

Global Storage (globalStorage/state.vscdb)
  ├── workbench.panel.aichat.view.aichat.chatdata: {
  │     tabs: [{ tabId, bubbles: [...] }]
  │   }
  └── workbench.panel.composerChatViewPane.{uuid}.hidden: [
        { id: "workbench.panel.aichat.view.{uuid}", isHidden: bool }
      ]
```

### UUID Identification

**Conversation UUID:**
- Stored as `composerId` in `composer.composerData.allComposers[]`
- Example: `59f64e5e-70e9-4892-a9e3-69d3d7f4b42d`
- Matches `conversation_id` in hook logs

**Generation UUID:**
- Stored as `generationUUID` in `aiService.generations[]`
- Example: `3bcfa52b-ec65-4a91-83e5-e7663e3cc005`
- Matches `generation_id` in hook logs

**Anchor UUID (Not a Conversation ID):**
- `aa8ad79b-e459-4989-93dd-5fbf136d08d0` is **NOT** a conversation identifier
- It's just text content in one of the prompts
- Found in `aiService.generations[3].textDescription`

### Key Insights

1. **Per-Workspace Storage:** Conversations are stored per-workspace, not globally
2. **Dual Storage:** Prompts stored in both `aiService.generations` and `aiService.prompts`
3. **Metadata Separation:** Conversation metadata stored separately from prompts
4. **View State:** View state tracked separately in composerChatViewPane keys
5. **UUID Confusion:** Text content UUIDs are not conversation identifiers

---

*Analysis completed: 2025-12-15*  
*Corrected: Removed hooks.db references (user experiment)*  
*Updated: Found actual per-workspace conversation storage structure*
