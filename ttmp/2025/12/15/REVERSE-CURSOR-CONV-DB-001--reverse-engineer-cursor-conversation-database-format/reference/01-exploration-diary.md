---
Title: Exploration Diary
Ticket: REVERSE-CURSOR-CONV-DB-001
Status: active
Topics:
    - reverse-engineering
    - data-analysis
    - exploration
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-15T09:22:09.591880626-05:00
---

# Exploration Diary: Reverse Engineering Cursor Conversation Database

**Conversation UUID Anchor:** `aa8ad79b-e459-4989-93dd-5fbf136d08d0`

## Goal

To understand how Cursor stores conversation history, agent interactions, and related metadata. This exploration seeks to map the complete data architecture of Cursor's conversation persistence layer, documenting file structures, database schemas, and data relationships.

---

## Step 1: Initial Reconnaissance - Mapping the Territory

Like Darwin setting foot on the Galápagos, I began by surveying the landscape. The expedition started at `~/.cursor`, the primary habitat where Cursor stores its data.

### What I did

- Listed the top-level directory structure of `~/.cursor`
- Searched for database files (`*.db`, `*.sqlite*`)
- Examined configuration files (`hooks.json`, `ide_state.json`)
- Explored the `projects/` directory structure

### Why

To establish a baseline understanding of where Cursor stores data before diving into specific storage mechanisms.

### What I found

**Primary Directory Structure (`~/.cursor`):**
- `hooks.db` - SQLite database (1.1MB, 287 pages) - **This looks promising!**
- `hooks.json` - Hook configuration (references `beforeSubmitPrompt` hook)
- `ide_state.json` - Recently viewed files (JSON)
- `projects/` - Per-workspace directories with structure:
  - `agent-tools/` - UUID-named `.txt` files containing tool execution logs
  - `terminals/` - Terminal output files
  - `repo.json` - Workspace metadata
  - `worker.log` - Worker process logs
- `commands/` - Custom command definitions (markdown files)
- `extensions/` - VS Code extension cache
- `worktrees/` - Git worktree data

**Key Discovery:** The `hooks.db` file appears to be the central repository for agent interaction data. It's a SQLite 3.x database written with SQLite version 3045001.

### What I learned

Cursor uses a multi-layered storage approach:
1. **Global storage** (`~/.cursor`) for hooks and IDE state
2. **Per-workspace storage** (`~/.cursor/projects/`) for workspace-specific data
3. **VS Code-compatible storage** (`~/.config/Cursor/`) for editor state

### Search Log

| Search Query | Purpose | Result |
|-------------|---------|--------|
| `find ~/.cursor -name "*.db"` | Find database files | Found `hooks.db` and some analysis databases in worktrees |
| `find ~/.cursor -name "*conversation*"` | Direct conversation storage | No direct matches - conversations likely in database |
| `ls -la ~/.cursor` | Directory overview | Revealed `hooks.db`, `projects/`, `commands/` structure |

---

## Step 2: Probing the hooks.db Database - The Central Repository

Having identified `hooks.db` as a likely candidate, I began a systematic examination of its structure.

### What I did

- Inspected database file type and version
- Listed all tables in the database
- Examined schema for each table
- Queried sample data to understand relationships

### Why

To understand the data model and how conversations are tracked across different event types.

### What I found

**Database Schema Overview:**

The `hooks.db` database contains **11 tables** tracking various aspects of agent interactions:

1. **`prompt_submissions`** - User prompts submitted to the agent
   - Fields: `id`, `hook_event_name`, `conversation_id`, `generation_id`, `model`, `prompt`, `attachments` (JSON), `timestamp`, `cursor_version`, `workspace_roots`, `user_email`
   - **Key insight:** `conversation_id` and `generation_id` link related events

2. **`agent_responses`** - Agent response text
   - Fields: Similar to prompt_submissions, but with `text` instead of `prompt`
   - Links to conversations via `conversation_id` and `generation_id`

3. **`agent_thoughts`** - Internal agent reasoning/thinking
   - Fields: Includes `duration_ms` for thought processing time
   - Also linked via `conversation_id` and `generation_id`

4. **`file_operations`** - File edits made by the agent
   - Fields: `file_path`, `content`, `edits` (JSON array)
   - Tracks what files were modified during a conversation

5. **`shell_executions`** - Terminal commands executed
   - Fields: `command`, `cwd`, `output`, `duration_ms`
   - Records all shell activity during agent interactions

6. **`mcp_executions`** - MCP (Model Context Protocol) tool calls
   - Fields: `tool_name`, `tool_input`, `result_json`, `url`, `command`
   - Tracks external tool integrations

7. **`tab_file_reads`** - Files read by the agent
   - Fields: `file_path`, `content`
   - Logs file access patterns

8. **`tab_file_edits`** - File edits with detailed change tracking
   - Fields: `file_path`, `edits` (JSON array with ranges, old_line, new_line)
   - More granular than `file_operations`

9. **`agent_stops`** - When agent execution stopped
   - Fields: `status`, `loop_count`
   - Tracks termination conditions

10. **`hook_invocations`** - Raw hook execution logs
    - Fields: `pwd`, `raw_json` (complete JSON input from Cursor)
    - Most detailed table - contains full event payloads

11. **`sqlite_sequence`** - SQLite internal sequence tracking

**Data Relationships:**
- All tables share common linking fields: `conversation_id` and `generation_id`
- A `conversation_id` represents a single chat session
- A `generation_id` represents a single request/response cycle within a conversation
- Multiple events can share the same `generation_id` (e.g., multiple thoughts, file edits, shell commands)

**Sample Statistics:**
- Total prompt submissions: 10
- Unique conversations: 3 (one conversation has 5 prompts, another has 3)
- Agent responses: 5
- Most conversations have sparse data (not all events are logged)

### What worked

The SQLite database structure is well-normalized and queryable. The `conversation_id` and `generation_id` fields provide clear linking between related events.

### What didn't work

Some conversations have missing data - not all prompt submissions have corresponding agent responses in the database. This suggests:
1. Responses might be stored elsewhere
2. Some conversations might be incomplete/interrupted
3. The hook system might not capture all events

### What I learned

**Conversation Structure:**
- Conversations are identified by UUID (`conversation_id`)
- Each user prompt gets a new `generation_id`
- Multiple events (thoughts, file edits, shell commands) share the same `generation_id`
- The database tracks the complete "paper trail" of agent activity

**Hook System:**
- Cursor uses a hook-based event system
- Hooks are configured in `~/.cursor/hooks.json`
- The database captures events fired by these hooks
- Events include: `beforeSubmitPrompt`, `afterAgentResponse`, `afterAgentThought`, `afterFileEdit`, `afterShellExecution`, etc.

### Search Log

| Search Query | Purpose | Result |
|-------------|---------|--------|
| `sqlite3 hooks.db ".tables"` | List all tables | Found 11 tables |
| `sqlite3 hooks.db "PRAGMA table_info(prompt_submissions)"` | Understand prompt schema | Revealed conversation_id/generation_id linking |
| `sqlite3 hooks.db "SELECT COUNT(*) FROM prompt_submissions"` | Data volume | 10 prompt submissions total |
| `sqlite3 hooks.db "SELECT conversation_id, COUNT(*) FROM prompt_submissions GROUP BY conversation_id"` | Conversation distribution | Found 3 unique conversations |

---

## Step 3: Exploring Alternative Storage Locations - The Quest for Full Conversations

Suspecting that `hooks.db` might only contain metadata or partial data, I expanded the search to other potential storage locations.

### What I did

- Explored `~/.config/Cursor/` directory structure
- Examined workspace storage databases (`state.vscdb`)
- Searched for conversation-related JSON files
- Checked Local Storage and Session Storage (LevelDB)
- Investigated `anysphere.cursor-retrieval` directories

### Why

The `hooks.db` database seems to track events but might not contain the full conversation context. Cursor might store complete conversation history elsewhere, possibly in VS Code-compatible storage or browser-like storage mechanisms.

### What I found

**VS Code-Compatible Storage (`~/.config/Cursor/`):**

1. **Workspace Storage (`User/workspaceStorage/`):**
   - Each workspace has a UUID-named directory
   - Contains `state.vscdb` (SQLite database) with `ItemTable` (key-value store)
   - Keys include: `aiService.prompts`, `workbench.panel.aichat`
   - **Finding:** `aiService.prompts` exists but is very small (2 bytes) - likely just a marker

2. **History Storage (`User/History/`):**
   - Contains 9600+ directories (one per file)
   - Each directory has `entries.json` tracking file edit history
   - **Not conversation storage** - this is file edit undo/redo history

3. **Retrieval Storage (`anysphere.cursor-retrieval/`):**
   - Contains `high_level_folder_description.txt` and `embeddable_files.txt`
   - Used for codebase indexing/retrieval, not conversation storage

4. **Local Storage (`Local Storage/leveldb/`):**
   - LevelDB database (browser-like storage)
   - Contains application state but not conversation data

**Key Discovery:** The `state.vscdb` databases contain workspace state but `aiService.prompts` is essentially empty. This suggests conversations might be:
1. Stored in a different location
2. Stored remotely (cloud sync)
3. Stored in memory only
4. Stored in a format I haven't identified yet

### What worked

Found the VS Code-compatible storage structure, which gives insight into how Cursor integrates with VS Code's storage mechanisms.

### What didn't work

No clear location for full conversation text storage found. The `hooks.db` appears to be the primary (and possibly only) local storage for conversation data.

### What I learned

**Storage Architecture:**
- Cursor uses VS Code's storage APIs (`state.vscdb` for workspace state)
- Conversation data appears to be primarily in `hooks.db` (hook-based event logging)
- File edit history is separate (VS Code's native history system)
- Codebase indexing is separate (`anysphere.cursor-retrieval`)

**Implication:** The `hooks.db` database might be the **complete** local conversation storage, but it's event-based rather than message-based. To reconstruct a conversation, one would need to:
1. Query all events for a `conversation_id`
2. Group by `generation_id` to get request/response pairs
3. Reconstruct the conversation flow from events

### Search Log

| Search Query | Purpose | Result |
|-------------|---------|--------|
| `find ~/.config/Cursor -name "*.db"` | Find VS Code databases | Found many `state.vscdb` files |
| `sqlite3 state.vscdb "SELECT key FROM ItemTable"` | Check workspace storage keys | Found `aiService.prompts` (empty) |
| `find ~/.config/Cursor -name "*conversation*"` | Direct conversation files | No matches |
| `ls -la ~/.config/Cursor/User/History/` | Check history storage | 9600+ directories (file edit history) |

---

## Step 4: Examining Agent Tools and Project Storage - The Workspace Trail

Curious about the `agent-tools/` directories in projects, I investigated these UUID-named files to understand their relationship to conversations.

### What I did

- Examined sample `agent-tools/` files
- Checked their content and format
- Compared UUIDs to `generation_id` values in `hooks.db`
- Explored `repo.json` structure

### Why

The `agent-tools/` directory contains UUID-named files that might be related to conversation tracking or tool execution logs.

### What I found

**Agent Tools Files:**
- Located in `~/.cursor/projects/{workspace-hash}/agent-tools/`
- Files named with UUIDs (e.g., `02db10da-afde-4dc0-8dfd-971ea50e5865.txt`)
- Content: Tool execution logs showing docmgr commands, file operations, etc.
- **Format:** Plain text logs of agent tool invocations

**Sample Content:**
```
[ok] Reset to /tmp/docmgr-scenario-local-2025-12-12
[ok] Mock codebase created at /tmp/docmgr-scenario-local-2025-12-12/acme-chat-app
Docs root initialized at /tmp/docmgr-scenario-local-2025-12-12/acme-chat-app/ttmp
...
```

**Relationship to Conversations:**
- These appear to be execution logs, not conversation storage
- They capture the *output* of tool executions, not the conversation context
- UUIDs don't directly match `generation_id` values (different purpose)

**Project Structure:**
- Each workspace gets a directory: `home-manuel-workspaces-{date}-{name}/`
- Contains: `agent-tools/`, `terminals/`, `repo.json`, `worker.log`
- `repo.json`: Workspace metadata (if present)

### What worked

Confirmed that `agent-tools/` files are execution logs, providing a complementary view of agent activity alongside `hooks.db`.

### What I learned

**Multi-Layer Logging:**
- `hooks.db`: Event-based logging (structured, queryable)
- `agent-tools/`: Tool execution output (unstructured logs)
- `terminals/`: Terminal output snapshots
- Together, these provide a comprehensive audit trail

**Workspace Isolation:**
- Each workspace has its own storage directory
- Workspace identification uses path-based hashing
- This allows per-workspace data isolation

### Search Log

| Search Query | Purpose | Result |
|-------------|---------|--------|
| `read_file agent-tools/*.txt` | Examine tool logs | Found execution logs |
| `ls ~/.cursor/projects/*/agent-tools/` | List tool files | Found UUID-named .txt files |

---

## Step 5: Analyzing Database Relationships - Reconstructing Conversations

With a complete understanding of the database schema, I attempted to reconstruct a full conversation from the database to understand the data model.

### What I did

- Selected a conversation with multiple events (`conversation_id: 1d439f68-3a1d-44e7-8ffd-930a1e7f08dd`)
- Queried all events for this conversation
- Examined the relationship between `generation_id` values
- Attempted to reconstruct conversation flow

### Why

To understand how to piece together a complete conversation from the event-based storage model.

### What I found

**Conversation Reconstruction:**

For conversation `1d439f68-3a1d-44e7-8ffd-930a1e7f08dd`:

**Generation 1 (`204947fa-a126-47a5-9772-5f100f34f010`):**
- Prompt: "again"
- Response: (empty in database)
- Thoughts: 0 thoughts recorded
- File operations: 1 edit (removed trailing comma from JSON)
- Shell executions: None recorded

**Generation 2 (`09323ff7-0e08-4d39-a03c-267efd7e2857`):**
- Prompt: "run ls and store in file, and think"
- Response: "Ran `ls` and saved the output to `ls_output3.txt`..."
- Thoughts: Not recorded
- File operations: None recorded
- Shell executions: Likely present but not queried

**Generation 3 (`ef613c6f-6f97-4ad0-a1ff-ebe8162f53b2`):**
- Prompt: "edit the file by hand"
- Response: "Removed the trailing comma on line 14..."
- Thoughts: Not recorded
- File operations: 1 edit (JSON fix)
- Shell executions: None recorded

**Key Observations:**
1. Not all generations have responses recorded
2. Thoughts are rarely captured (might be opt-in or filtered)
3. File operations are well-tracked with full edit diffs
4. The `attachments` field in `prompt_submissions` contains JSON (likely file references)

### What worked

Successfully reconstructed conversation flow using `conversation_id` and `generation_id` linking.

### What didn't work

Some data appears missing (responses, thoughts). This could be:
1. Hooks not configured to capture all events
2. Events filtered before storage
3. Data stored elsewhere (cloud?)

### What I learned

**Conversation Model:**
- **Conversation** = Sequence of user prompts and agent responses
- **Generation** = Single request/response cycle
- **Events** = Individual actions within a generation (thoughts, file edits, shell commands)

**Data Completeness:**
- Prompts: Well-captured (via `beforeSubmitPrompt` hook)
- Responses: Sometimes captured (via `afterAgentResponse` hook)
- Thoughts: Rarely captured (via `afterAgentThought` hook)
- File operations: Well-captured (via `afterFileEdit` hook)
- Shell executions: Captured (via `afterShellExecution` hook)

**Reconstruction Strategy:**
To reconstruct a conversation:
1. Query `prompt_submissions` for a `conversation_id`
2. For each `generation_id`, query:
   - `agent_responses` for the response text
   - `agent_thoughts` for reasoning steps
   - `file_operations` for file changes
   - `shell_executions` for terminal activity
   - `mcp_executions` for tool calls
3. Order by `timestamp` to reconstruct chronological flow

### Search Log

| Search Query | Purpose | Result |
|-------------|---------|--------|
| `SELECT * FROM prompt_submissions WHERE conversation_id = '...'` | Get conversation prompts | Found 3 prompts |
| `SELECT * FROM agent_responses WHERE conversation_id = '...'` | Get responses | Found 2 responses (1 missing) |
| `SELECT * FROM file_operations WHERE conversation_id = '...'` | Get file edits | Found 2 file edits |

---

## Current Understanding: The Cursor Conversation Database Architecture

Based on my exploration, here is the current understanding of how Cursor stores conversations:

### Primary Storage: `~/.cursor/hooks.db`

**Database Type:** SQLite 3.x  
**Purpose:** Event-based logging of all agent interactions  
**Schema:** 11 tables tracking different event types

### Key Tables:

1. **`prompt_submissions`** - User prompts (input)
2. **`agent_responses`** - Agent responses (output)
3. **`agent_thoughts`** - Internal reasoning (if captured)
4. **`file_operations`** - File edits with full diffs
5. **`shell_executions`** - Terminal commands and output
6. **`mcp_executions`** - External tool calls
7. **`tab_file_reads`** - Files read by agent
8. **`tab_file_edits`** - Detailed file edit tracking
9. **`agent_stops`** - Termination events
10. **`hook_invocations`** - Raw hook event payloads

### Linking Mechanism:

- **`conversation_id`** (UUID): Groups all events in a single chat session
- **`generation_id`** (UUID): Groups events for a single request/response cycle
- **`timestamp`**: Chronological ordering

### Data Model:

```
Conversation (conversation_id)
  └── Generation 1 (generation_id_1)
      ├── Prompt (prompt_submissions)
      ├── Response (agent_responses)
      ├── Thoughts (agent_thoughts) [optional]
      ├── File Edits (file_operations)
      ├── Shell Commands (shell_executions)
      └── Tool Calls (mcp_executions)
  └── Generation 2 (generation_id_2)
      └── ...
```

### Complementary Storage:

- **`~/.cursor/projects/`**: Per-workspace execution logs
- **`~/.config/Cursor/`**: VS Code-compatible workspace state
- **`~/.config/Cursor/User/History/`**: File edit history (separate system)

---

## Next Steps

1. **Examine `hook_invocations` table** - Contains raw JSON, might have complete conversation data
2. **Check `attachments` field** - JSON array might contain file references or context
3. **Investigate cloud storage** - Cursor might sync conversations to cloud
4. **Analyze hook configuration** - Understand what events are captured vs. filtered
5. **Write analysis document** - Comprehensive format documentation

---

## Field Notes

**Interesting Discoveries:**
- The database uses event sourcing pattern (events, not state)
- Conversations are reconstructed from events, not stored as messages
- Some data appears missing (responses, thoughts) - might be hook configuration dependent
- File edits are stored with full content and diffs (very detailed)
- The `hook_invocations` table contains raw JSON - might be the most complete source

**Questions Remaining:**
- Where are full conversation transcripts stored? (if anywhere)
- Why are some responses missing from `agent_responses`?
- Are thoughts filtered or opt-in?
- Is there cloud sync for conversations?
- What does the `attachments` JSON contain?

**Hypotheses:**
1. `hooks.db` is the primary (and possibly only) local conversation storage
2. Conversations are event-based, not message-based
3. Missing data might be due to hook configuration or filtering
4. The `hook_invocations.raw_json` field might contain complete event data

---

## Step 6: Deep Dive into hook_invocations - The Complete Event Archive

The `hook_invocations` table caught my attention as it contains `raw_json` - potentially the most complete event data. I investigated this table to understand its role in conversation storage.

### What I did

- Counted total hook invocations
- Examined sample `raw_json` content
- Analyzed event type distribution
- Checked JSON payload sizes
- Compared hook_invocations to other tables

### Why

The `raw_json` field might contain complete event payloads that aren't captured in the normalized tables, making it a potential source for full conversation reconstruction.

### What I found

**hook_invocations Statistics:**
- Total entries: **269** (much more than other tables!)
- Contains complete JSON payloads from Cursor
- Largest JSON payload: 36,899 bytes (very detailed!)
- All 3 conversations have entries in this table

**Event Distribution:**
- `afterShellExecution`: 62 events (most common)
- `beforeShellExecution`: 62 events
- `afterAgentThought`: 51 events
- `afterFileEdit`: 36 events
- `beforeReadFile`: 28 events
- `stop`: 10 events
- `beforeSubmitPrompt`: 9 events
- `afterAgentResponse`: 8 events
- `beforeTabFileRead`: 3 events

**Key Discovery:** The `hook_invocations` table captures **all** hook events, not just the ones that get parsed into normalized tables. This makes it the most complete source of conversation data.

**Sample raw_json Content:**
- Contains complete event payloads with all fields
- Includes file paths, content, edits, command output, etc.
- JSON structure matches the normalized table schemas but with additional context

**attachments Field Format:**
```json
[{"type":"file","file_path":"/path/to/file.sh"}]
```
- JSON array of file references
- Used to track files attached to prompts

**edits Field Format:**
```json
[{"old_string":"...","new_string":"..."}]
```
- JSON array of edit diffs
- Contains full before/after content for file changes

### What worked

The `hook_invocations` table provides the most complete event archive, with full JSON payloads preserving all event data.

### What I learned

**Data Completeness Hierarchy:**
1. **`hook_invocations`** - Most complete (raw JSON, all events)
2. **Normalized tables** - Structured, queryable, but may miss some events
3. **agent-tools files** - Execution logs, complementary view

**Hook Configuration:**
- Only `beforeSubmitPrompt` hook is configured in `hooks.json`
- Other hooks must be configured per-workspace (in `.cursor/hooks.json` within projects)
- This explains why some events are captured and others aren't

**Event Capture Pattern:**
- Shell executions are most frequently captured (124 events)
- Agent thoughts are well-captured (51 events)
- File edits are captured (36 events)
- Prompt submissions are less frequent (9 events) - likely because only one hook is globally configured

### Search Log

| Search Query | Purpose | Result |
|-------------|---------|--------|
| `SELECT COUNT(*) FROM hook_invocations` | Total events | 269 entries |
| `SELECT attachments FROM prompt_submissions` | Attachment format | JSON array with file references |
| `SELECT edits FROM file_operations` | Edit format | JSON array with old_string/new_string |
| `SELECT hook_event_name, COUNT(*) FROM hook_invocations GROUP BY hook_event_name` | Event distribution | Shell executions most common |

---

## Step 7: Final Synthesis - The Complete Picture

Having explored all major storage locations, I synthesized my findings into a complete understanding of Cursor's conversation storage architecture.

### What I did

- Reviewed all findings across exploration steps
- Identified patterns and relationships
- Documented the complete data model
- Identified gaps and limitations

### Why

To create a comprehensive understanding that can guide future analysis or tooling development.

### What I found

**Complete Storage Architecture:**

1. **Primary Storage: `~/.cursor/hooks.db`**
   - Event-based conversation logging
   - 11 normalized tables + `hook_invocations` with raw JSON
   - Most complete local conversation storage

2. **Complementary Storage:**
   - `~/.cursor/projects/`: Per-workspace execution logs
   - `~/.config/Cursor/`: VS Code workspace state (minimal conversation data)
   - `~/.config/Cursor/User/History/`: File edit history (separate system)

3. **Data Model:**
   - Conversations identified by `conversation_id` (UUID)
   - Generations identified by `generation_id` (UUID)
   - Events linked via these IDs
   - Chronological ordering via `timestamp`

4. **Data Completeness:**
   - Prompts: Well-captured (if hooks configured)
   - Responses: Sometimes captured (hook-dependent)
   - Thoughts: Frequently captured (51 events found)
   - File edits: Well-captured with full diffs
   - Shell executions: Very well-captured (124 events)
   - Tool calls: Captured via MCP executions

### What worked

The exploration revealed a clear, queryable data model with good event coverage for most interaction types.

### What didn't work

Some data appears missing due to hook configuration. The system is hook-dependent, so completeness varies by workspace configuration.

### What I learned

**Key Insights:**
1. Cursor uses **event sourcing** for conversation storage
2. Conversations are **reconstructed** from events, not stored as messages
3. **Hook configuration** determines what gets captured
4. **hook_invocations** is the most complete data source
5. The system is **workspace-aware** (per-workspace storage)

**Limitations:**
- Data completeness depends on hook configuration
- Some responses may be missing if hooks aren't configured
- Cloud sync (if any) wasn't discovered in local storage
- No direct "conversation transcript" storage found

**Reconstruction Strategy:**
To reconstruct a complete conversation:
1. Query `hook_invocations` for `conversation_id` (most complete)
2. Or query normalized tables grouped by `generation_id`
3. Order events by `timestamp`
4. Reconstruct flow: Prompt → Thoughts → File Edits → Shell Commands → Response

---

## Step 8: Critical Correction - hooks.db Was User Experiment

**Important Discovery:** The `hooks.db` database I analyzed extensively was actually a user experiment, not Cursor's native conversation storage! This requires a complete revision of my findings.

### What I did

- User informed me that `hooks.db` was their experiment
- Restarted exploration focusing on actual Cursor storage
- Searched `~/.config/Cursor/` more thoroughly
- Examined `state.vscdb` database (2.1GB!)

### Why

To find the actual Cursor conversation storage, not user experiments.

### What I found

**Actual Storage Locations:**

1. **`~/.config/Cursor/User/globalStorage/state.vscdb`** - **2.1GB SQLite database**
   - Contains `workbench.panel.aichat.view.aichat.chatdata` key
   - Contains thousands of `workbench.panel.composerChatViewPane.{uuid}.hidden` keys
   - This appears to be the primary conversation storage!

2. **Chat Data Structure:**
   - Key: `workbench.panel.aichat.view.aichat.chatdata`
   - Format: JSON with `tabs` array
   - Each tab has: `tabId`, `tabState`, `bubbles` array
   - Bubbles contain messages with `type`, `id`, `messageType`, etc.

3. **Composer Chat View Panes:**
   - Keys: `workbench.panel.composerChatViewPane.{uuid}.hidden`
   - Thousands of these UUIDs
   - Values appear to be JSON arrays with view state

4. **Sourcegraph AMP Extension:**
   - `~/.config/Cursor/User/globalStorage/sourcegraph.amp/threads3-development/`
   - Contains thread JSON files (T-{uuid}.json)
   - These are from an extension, not native Cursor storage

### What worked

Found the actual storage location: `state.vscdb` with chat data keys.

### What didn't work

- UUID `aa8ad79b-e459-4989-93dd-5fbf136d08d0` not found yet in searches
- Database is locked when Cursor is running (2.1GB size suggests heavy usage)
- Need to explore the chatdata structure more deeply

### What I learned

**Key Insight:** Cursor uses VS Code's storage API (`state.vscdb`) for conversation persistence, not a custom database like `hooks.db`.

**Storage Pattern:**
- Global storage: `~/.config/Cursor/User/globalStorage/state.vscdb`
- Workspace storage: `~/.config/Cursor/User/workspaceStorage/{workspace-uuid}/state.vscdb`
- Key-value store pattern (ItemTable with key/value columns)

**Next Steps:**
- Explore `chatdata` structure in detail
- Search for UUID in the large database
- Understand composerChatViewPane structure
- Check if conversations are stored per-workspace or globally

### Search Log

| Search Query | Purpose | Result |
|-------------|---------|--------|
| `find ~/.config/Cursor -name "*.db"` | Find databases | Found state.vscdb (2.1GB) |
| `sqlite3 state.vscdb "SELECT key FROM ItemTable WHERE key LIKE '%chat%'"` | Find chat keys | Found chatdata and composerChatViewPane keys |
| `sqlite3 state.vscdb "SELECT value FROM ItemTable WHERE key = 'workbench.panel.aichat.view.aichat.chatdata'"` | Get chat data | Found JSON with tabs structure |
| `grep -r "aa8ad79b" ~/.config/Cursor` | Find UUID | No matches yet |

---

## Step 9: Searching for Conversation UUID

Attempted to locate the conversation UUID `aa8ad79b-e459-4989-93dd-5fbf136d08d0` in the actual Cursor storage.

### What I did

- Searched `state.vscdb` for keys containing the UUID
- Searched values in `state.vscdb` for the UUID string
- Examined `chatdata` structure for tab IDs matching the UUID
- Checked composerChatViewPane entries for UUID references
- Searched JSON files in globalStorage

### Why

To verify where and how conversation IDs are stored, and to find this specific conversation.

### What I found

**UUID Search Results:**
- No keys found containing `aa8ad79b-e459-4989-93dd-5fbf136d08d0`
- No values found containing the UUID string
- Current `chatdata` contains only 1 tab with ID `2a279cd5-ab52-43e1-8f48-f62cd7848f10` (different UUID)
- ComposerChatViewPane entries reference `workbench.panel.aichat.view.{uuid}` IDs, but those keys don't exist as separate entries

**Possible Explanations:**
1. **Conversation not yet persisted** - This is an active conversation, may be stored in memory or not yet written to disk
2. **Different storage location** - Conversation might be stored per-workspace or in a different database
3. **UUID format difference** - The stored format might differ from the anchor UUID
4. **Cloud storage** - Conversations might be synced to cloud rather than stored locally

### What worked

Confirmed the storage structure and data format, even if this specific UUID isn't found yet.

### What I learned

**Storage Pattern:**
- Conversations appear to be stored in `chatdata` as tabs
- Each tab has a `tabId` (UUID)
- Tabs contain `bubbles` array with messages
- ComposerChatViewPane entries track view state, not conversation data directly

**Current State:**
- Only 1 active tab in chatdata
- Tab ID doesn't match search UUID
- Suggests either conversation not persisted yet or stored elsewhere

### Search Log

| Search Query | Purpose | Result |
|-------------|---------|--------|
| `SELECT key FROM ItemTable WHERE key LIKE '%aa8ad79b%'` | Find UUID in keys | No matches |
| `SELECT value FROM ItemTable WHERE value LIKE '%aa8ad79b%'` | Find UUID in values | No matches |
| `grep -r "aa8ad79b" ~/.config/Cursor` | Search all files | No matches |
| Examined chatdata tabs | Check tab IDs | Found 1 tab with different UUID |

---

## Step 10: Found It! The Actual Conversation Storage

After the user pointed out that our conversation text appears in the databases, I discovered the actual conversation storage structure.

### What I did

- Searched for conversation text "hooks.db is actually not relevant" in databases
- Found it in workspace storage: `94bae793ba109d83fb8934a587a6c719/state.vscdb`
- Examined `aiService.generations` and `aiService.prompts` keys
- Found `composer.composerData` containing conversation metadata
- Located the actual conversation UUID from hook logs

### Why

To understand where conversations are actually stored and how they're structured.

### What I found

**Actual Conversation Storage:**

1. **`aiService.generations`** - Array of generation records
   - Each generation has: `generationUUID`, `type`, `textDescription`, `unixMs`
   - Contains all prompts submitted in this workspace
   - 6 generations found for this conversation

2. **`aiService.prompts`** - Array of prompt objects
   - Each prompt has: `text`, `commandType`
   - Simpler structure than generations
   - 6 prompts found (matches generations)

3. **`composer.composerData`** - Composer metadata
   - `allComposers`: Array of composer/conversation metadata
   - Each composer has:
     - `composerId`: The actual conversation UUID! (`59f64e5e-70e9-4892-a9e3-69d3d7f4b42d`)
     - `name`: Conversation name ("REVERSE-CURSOR-001 - initial research")
     - `createdAt`, `lastUpdatedAt`: Timestamps
     - `totalLinesAdded`, `totalLinesRemoved`: Edit statistics
     - `filesChangedCount`: Number of files changed
     - `subtitle`: Files changed summary
     - `contextUsagePercent`: Context usage
     - `unifiedMode`: Mode ("agent" or "chat")
     - `forceMode`: Force mode setting
     - `isArchived`, `isDraft`: Status flags

4. **Hook Logs** - `cursor.hooks.log` contains:
   - `conversation_id`: `59f64e5e-70e9-4892-a9e3-69d3d7f4b42d` (matches composerId!)
   - `generation_id`: Matches `generationUUID` in aiService.generations
   - Full prompt text and metadata

**Key Discovery:**
- The UUID `aa8ad79b-e459-4989-93dd-5fbf136d08d0` I generated is **NOT** a conversation UUID
- It's just text content in one of the prompts (generation `b7e5b0b1-45f5-4c1d-bdfe-79250a20ff91`)
- The **actual conversation UUID** is `59f64e5e-70e9-4892-a9e3-69d3d7f4b42d` (the composerId)
- Conversations are stored **per-workspace** in workspaceStorage, not globally

**Storage Structure:**
```
Workspace Storage (94bae793ba109d83fb8934a587a6c719/state.vscdb)
  ├── aiService.generations: [generation records with generationUUID]
  ├── aiService.prompts: [prompt objects]
  ├── composer.composerData: {
  │     allComposers: [{
  │       composerId: "59f64e5e-70e9-4892-a9e3-69d3d7f4b42d" (conversation UUID!)
  │       name: "REVERSE-CURSOR-001 - initial research"
  │       ...
  │     }]
  │   }
  └── workbench.panel.composerChatViewPane.{viewId}: [view references]
```

### What worked

Found the complete conversation storage structure! Conversations are stored per-workspace with:
- Metadata in `composer.composerData`
- Prompts in `aiService.prompts`
- Generations in `aiService.generations`
- View state in `composerChatViewPane` keys

### What I learned

**Conversation Storage Model:**
- **Conversation UUID** = `composerId` in `composer.composerData.allComposers[]`
- **Generation UUID** = `generationUUID` in `aiService.generations[]`
- **Storage Location** = Per-workspace (`workspaceStorage/{workspace-uuid}/state.vscdb`)
- **Metadata** = Stored separately from messages (in composer.composerData)
- **Messages** = Stored as generations and prompts arrays

**Why UUID Not Found Initially:**
- I was searching for `aa8ad79b-e459-4989-93dd-5fbf136d08d0` which is just text content
- The actual conversation UUID is `59f64e5e-70e9-4892-a9e3-69d3d7f4b42d`
- Conversations are stored per-workspace, not globally
- Need to check workspaceStorage, not just globalStorage

**Storage Pattern:**
- Per-workspace storage for conversation data
- Global storage for UI state and settings
- Composer metadata tracks conversation statistics
- Generations track individual prompts/responses

### Search Log

| Search Query | Purpose | Result |
|-------------|---------|--------|
| `grep -r "hooks.db is actually not relevant"` | Find conversation text | Found in logs and workspaceStorage |
| `SELECT key FROM ItemTable WHERE value LIKE '%hooks.db%'` | Find storage key | Found `aiService.generations` and `aiService.prompts` |
| `SELECT value FROM ItemTable WHERE key = 'composer.composerData'` | Get composer data | Found conversation metadata with composerId |
| `grep "conversation_id" cursor.hooks.log` | Find conversation UUID | Found `59f64e5e-70e9-4892-a9e3-69d3d7f4b42d` |

---

*Exploration complete! Found actual conversation storage structure.*

---

## Step 11: Consolidation — Per-workspace DB confirmed, and a first pass at “where are assistant responses?”

This step tightened the map: I verified the *exact* on-disk workspace DB that holds our composer conversation, enumerated the relevant keys and their sizes, and then tried to locate where Cursor persists the **assistant side** of the conversation (beyond just the user prompts). I also hit repeated JSON parsing failures when piping `sqlite3` output into Python, so I’m pausing further “debug-the-parser” attempts per the “don’t fix errors more than twice in a row” rule.

### What I did
- Listed the workspace storage directory for this conversation’s workspace:
  - `/home/manuel/.config/Cursor/User/workspaceStorage/94bae793ba109d83fb8934a587a6c719/`
- Confirmed the DB file and its rough size:
  - `state.vscdb` is **~98KB** (small enough to inspect exhaustively)
- Enumerated conversation-related keys and measured their value sizes:
  - `aiService.generations` (~2518 bytes)
  - `aiService.prompts` (~1874 bytes)
  - `composer.composerData` (~1163 bytes)
  - `workbench.panel.composerChatViewPane.ecc9c419-097d-49e4-b521-95d311edd307` (~117 bytes)
  - `workbench.panel.aichat.ecc9c419-097d-49e4-b521-95d311edd307.numberOfVisibleViews` (~1 byte)
- Searched the workspace DB for additional keys likely to store assistant responses (patterns: `assistant`, `response`, `completion`, `message`, `tool`, etc.)
- Attempted to parse `aiService.generations` with a `sqlite3 | python3` pipeline to introspect all keys present in each generation record, but hit JSON decode errors (stdin was empty / non-JSON).

### Why
- To confirm we’re looking at the *right* persistence layer: per-workspace `state.vscdb`, not global UI state.
- To identify the minimal set of keys that constitute a “conversation record” and learn where the assistant responses are persisted.

### What worked
- The per-workspace database is clearly bounded and contains exactly the expected conversation scaffolding:
  - prompts/generations (`aiService.*`)
  - conversation metadata (`composer.composerData`, including `composerId = conversation_id`)
  - a view-pane reference mapping (`workbench.panel.composerChatViewPane.*`)
- Key sizes are small and stable, suggesting the per-workspace DB stores **metadata + prompt summaries**, not full transcripts (or full transcripts are stored elsewhere and referenced here).

### What didn’t work
- The direct pipeline `sqlite3 ... "SELECT value FROM ItemTable WHERE key='aiService.generations';" | python3 -c 'json.load(...)'` intermittently produced **empty stdin** (resulting in `JSONDecodeError: Expecting value: line 1 column 1`).
  - This likely means: `sqlite3` printed nothing due to an error on stderr (or due to BLOB/text mode issues), and the Python side attempted to parse an empty stream.
  - Per user rule, I’m not going to keep iterating on fixing this in an ad-hoc way right now.

### What I learned
- The conversation is **definitely** persisted per-workspace in `workspaceStorage/{workspace-uuid}/state.vscdb`.
- The “anchor UUID” we generated (`aa8ad79b-e459-4989-93dd-5fbf136d08d0`) is not a structural ID; it’s content that can appear inside `aiService.generations[*].textDescription`.
- The workspace DB currently exposes prompt-side artifacts and composer metadata, but does not obviously expose assistant messages via similarly named keys—so assistant responses may live in:
  - global storage (UI transcript cache),
  - a different workspace key not matched by my current grep patterns,
  - log files (e.g., composer/AI logs),
  - or a separate local/remote store.

### What warrants a second pair of eyes
- Whether `aiService.generations`/`aiService.prompts` are **only user prompts** (as observed) or whether they can include assistant turns under different `type` values / schemas.
- Whether `composer.composerData` is the authoritative “conversation index” and where it references the actual message transcript.

### Next steps
- Search `~/.config/Cursor/logs/...` for the **conversation_id** (`59f64e5e-70e9-4892-a9e3-69d3d7f4b42d`) and **generationUUIDs** to locate persisted assistant output in logs.
- Inspect the *values* of `history.entries` in the workspace DB; sometimes VS Code stores larger structured blobs under generic keys.
- Re-run the JSON extraction for `aiService.generations` using a more robust extraction method (e.g., `SELECT quote(value)` or `hex(value)` and decode) — but only after you confirm you want me to proceed given the “two error fixes” rule.

### Search log (commands + results)
- **List workspace storage directory**:
  - `/home/manuel/.config/Cursor/User/workspaceStorage/94bae793ba109d83fb8934a587a6c719/` contains `state.vscdb` (~98KB) and `workspace.json`
- **Key size scan**:
  - Found the five relevant keys above; no additional obvious `assistant/response/message` keys surfaced in this DB.

---

## Step 12: Script Preservation — capture the expedition tools as reproducible scripts

This step turns the “field improvisations” into durable tooling. The goal is that a future reader can reproduce the investigation by running scripts from the ticket folder, without re-deriving the ad-hoc one-liners.

### What I did
- Added the following scripts to the ticket’s `scripts/` directory:
  - `scripts/dump_workspace_conversation.py`
  - `scripts/scan_global_state_for_needles.py`
  - `scripts/extract_aiCodeTrackingLines.py`
  - `scripts/list_workspace_keys.sh`
  - `scripts/grep_cursor_logs_for_conversation.sh`
  - `scripts/README.md`
- Copied an “interesting” artifact produced during exploration into the ticket:
  - `scripts/artifact-agent-tools-7206d086-264b-46da-93b3-384c8fd90eb4.txt`
- Related these scripts in the ticket index via `docmgr doc relate` so they’re discoverable from `index.md`.

### Why
- Keeps this work reproducible and reviewable.
- Prevents the “knowledge evaporates into terminal scrollback” failure mode.
- Ensures we follow your rule: *store all scripts we do in the ticket scripts folder*.

### What worked
- Scripts are now versioned alongside the analysis and diary, with absolute-path examples in `scripts/README.md`.

### What should be done next
- From now on, prefer updating/adding scripts here instead of running long one-off pipelines.

---

## Step 13: Repro Runs + New Findings — where responses *actually* show up (chat vs composer)

This step re-ran the investigation using the ticket scripts, and it clarified an important split: **Cursor “AI Chat” history (`aichat.chatdata`) contains full assistant messages**, while **Composer/Agent mode in our current workspace persists prompts + metadata, but not obviously the assistant response transcript** (at least not in the small per-workspace DB).

### What I did
- Ran `scripts/dump_workspace_conversation.py` against our workspace DB:
  - `/home/manuel/.config/Cursor/User/workspaceStorage/94bae793ba109d83fb8934a587a6c719/state.vscdb`
  - Confirmed `composerId` conversation UUID: `59f64e5e-70e9-4892-a9e3-69d3d7f4b42d`
  - Observed `aiService.generations.count` grew (we kept prompting “continue”)
- Ran `scripts/list_workspace_keys.sh` on the same DB:
  - Confirmed only a handful of relevant keys exist, all **TEXT** (no BLOBs)
- Ran `scripts/scan_global_state_for_needles.py`:
  - Found `composerId` in global `aiCodeTrackingLines` and in a `composerChatViewPane.*.hidden` entry
  - Did **not** find our prompt phrases in global storage (they live per-workspace)
- Inspected global `aiCodeTrackingLines` and confirmed it is a **JSON array** with entries keyed by:
  - `metadata.source=composer`, `metadata.composerId`, `metadata.fileName`, `metadata.invocationID`, `metadata.timestamp`
- Probed other browser-like storage (LevelDB / Session Storage / Service Worker DB) for our conversation_id / generationUUIDs:
  - Found **no hits**
- Surveyed other workspaces and found some have `workbench.panel.aichat.view.aichat.chatdata` stored in their workspace DBs.
- Pulled one such workspace (`072b61c4b8bc1f706965d8cbddf2d100`) and dumped its `aichat.chatdata`:
  - Confirmed it includes an AI bubble with `rawText`, `modelType`, `requestId`
  - Wrote `scripts/extract_aichat_chatdata.py` to extract readable transcripts from that structure

### Why
- To determine where assistant responses are persisted, and whether the storage differs between “AI Chat” and “Composer/Agent”.

### What worked
- **AI Chat transcripts** are straightforward:
  - `workbench.panel.aichat.view.aichat.chatdata` contains AI bubbles with `rawText` (full assistant message text).
- **Composer tracking** is strong at the metadata level:
  - `composer.composerData` has conversation stats (files changed, lines added/removed, context usage).
  - `aiCodeTrackingLines` links composerId + fileName + invocationID (helps correlate edits with prompts).

### What didn’t work / what’s still missing
- For our current conversation’s **Composer/Agent** mode, the per-workspace DB contains:
  - prompts (`aiService.generations`, `aiService.prompts`)
  - metadata (`composer.composerData`)
  - pane references (`workbench.panel.composerChatViewPane.*`)
  - but **no obvious assistant response transcript key** (no `rawText`, `role`, `assistant`, etc.).
- `cursor.hooks.log` is a hook-runner log; we do not have hooks configured for `afterAgentResponse` / `afterAgentThought`, so it does not contain the assistant response payloads for this conversation.

### What I learned
- Cursor appears to persist two different “conversation-like” stores:
  - **AI Chat**: full transcript lives in `aichat.chatdata` (bubbles include assistant `rawText`)
  - **Composer/Agent**: workspace DB stores prompts + metadata; assistant responses likely live elsewhere (or are stored in a structure we haven’t found yet).

### Artifacts captured
- Copied additional `agent-tools/*.txt` outputs into the ticket `scripts/` folder as `artifact-agent-tools-*.txt` for reproducibility.

---

## Step 14: Negative Space Mapping — “assistant text isn’t on disk” (for Composer/Agent)

This step is about proving a *negative*: I attempted to locate any durable on-disk store that contains the assistant’s replies for our current Composer/Agent conversation by searching for distinctive assistant-only phrases. The result was consistently **NO HITS** across all obvious local persistence layers.

### What I did
- Queried global storage (`/home/manuel/.config/Cursor/User/globalStorage/state.vscdb`) for keys that might correspond to our Composer view id:
  - Looked for `workbench.panel.aichat.view.<composerId>`-style keys and for keys containing the composerId in their *names*.
  - Result: no such keys exist (only `workbench.panel.aichat.view.aichat.chatdata` exists globally, and it’s small).
- Inspected large `workbench.panel.composerChatViewPane.*.hidden` entries:
  - These contain lists of `"id": "workbench.panel.aichat.view.<uuid>"` entries and `isHidden` booleans.
  - They do **not** contain message text (`rawText`/`role`/etc.).
- Searched for distinctive assistant phrases (“Got it:”, “I’ll first append”, “Plan:”) via:
  - `scripts/scan_global_state_for_needles.py` on both global and our workspace DB
  - `scripts/grep_cursor_logs_for_conversation.sh` over `~/.config/Cursor/logs`
  - Result: no assistant phrase hits anywhere (only the *commands we ran* show up in `cursor.hooks.log`, which is expected).

### Why
- To determine whether Composer/Agent responses are actually persisted locally (and where), or whether only prompts/metadata are persisted and the rest is reconstructed from a remote store.

### What I learned
- **AI Chat** and **Composer/Agent** appear to have materially different persistence characteristics:
  - **AI Chat**: assistant responses are persisted as plain text (`rawText`) in `workbench.panel.aichat.view.aichat.chatdata` (confirmed in other workspace DBs).
  - **Composer/Agent** (our conversation): prompts are persisted (`aiService.generations`, `aiService.prompts`) and metadata is persisted (`composer.composerData`, `aiCodeTrackingLines`), but assistant responses are **not discoverable** as plain text in:
    - our workspace `state.vscdb`
    - global `state.vscdb`
    - Cursor logs (beyond prompt + tool executions)
    - Electron-ish stores (Local Storage LevelDB / Session Storage / Service Worker DB)
- Working hypothesis: **Composer/Agent assistant transcript is not stored locally as plain text** (or is stored in an encoded/encrypted/remote-only form that does not preserve easily-searchable plaintext).

### Search log (high-signal probes)
- `scan_global_state_for_needles.py --needle "Got it:" --needle "I’ll first append" --needle "Plan:"` (global + workspace DB): **NO HITS**
- dump of `workbench.panel.composerChatViewPane.*.hidden`: lists view IDs only, no transcript fields

---

## Step 15: Reproducing the “binary file matches” grep — and exporting the SQLite contents to Markdown

You observed:

> `grep -r "hooks.db is actually not relevant" *`  
> … `state.vscdb: binary file matches` (workspaceStorage + globalStorage)

This step replays that observation in a more *structured* way:
- confirm the phrase is present in the **workspace** `state.vscdb` *as actual ItemTable values*
- attempt to locate it in **global** `state.vscdb` (it does *not* appear in ItemTable values there)
- export the relevant SQLite contents to Markdown so you can manually inspect what’s in the DB(s)

### What I did
- Built a reusable exporter script:
  - `scripts/export_itemtable_to_markdown.py`
- Exported our workspace DB (small; fully exportable):
  - **DB**: `/home/manuel/.config/Cursor/User/workspaceStorage/94bae793ba109d83fb8934a587a6c719/state.vscdb`
  - **Output**: `reference/02-workspace-state-vscdb-itemtable-export.md`
  - This export includes **all 73 keys**, and dumps each key’s JSON (pretty-printed) where possible.
- Exported targeted slices of global DB (huge; only export filtered keys):
  - `reference/03-global-state-vscdb-aiCodeTrackingLines-export.md` (key = `aiCodeTrackingLines`)
  - `reference/04-global-state-vscdb-composer-pane-ecc9-export.md` (key = `workbench.panel.composerChatViewPane.ecc9c419-097d-49e4-b521-95d311edd307.hidden`)

### Why
- `grep` against a `.vscdb` reports “binary file matches” without telling us which logical key/value contains the match.
- Exporting the ItemTable gives a human-auditable, line-oriented artifact that answers “what is actually in the DB?”.

### What I learned
- In the **workspace** DB export, the phrase appears as part of the prompt text stored under:
  - `aiService.generations`
  - `aiService.prompts`
- In the **global** DB export, we still do not see that phrase in ItemTable keys/values. That means your earlier `grep` “binary match” for global storage may be:
  - a match in some other table (not `ItemTable`), or
  - a match in a WAL/free page/unallocated region that SQLite no longer references via `ItemTable`.

### Where to look now (if you want to go deeper)
- If you want to reconcile “grep says binary match” vs “ItemTable export has no match”, the next step is a *raw byte scan* of `globalStorage/state.vscdb` (streaming, because it’s large) to find offsets of that phrase and then correlate offsets to pages/tables.

---

## Step 16: Deep dive — identifying the “binary match” owner: `cursorDiskKV` (and exporting it)

This is the breakthrough: the “binary file matches” in `globalStorage/state.vscdb` do **not** come from `ItemTable`. They come from a *second* table, `cursorDiskKV`, where values are stored as large BLOB/TEXT payloads, often spilling into overflow pages (which is why grep calls it “binary”).

### What I did (and why)

1) **Raw byte scan → page numbers**
- I scanned the raw SQLite file bytes for the exact phrase:
  - needle: `"hooks.db is actually not relevant"`
  - file: `/home/manuel/.config/Cursor/User/globalStorage/state.vscdb`
- I recorded byte offsets and mapped them to SQLite pages (\(page = \lfloor offset / pageSize \rfloor + 1\)).
- The contexts around those hits looked like our investigation text and tool invocation JSON, not `ItemTable` JSON blobs.

2) **Page ownership mapping (`dbstat`)**
- I used `dbstat` to map the hit pages to owning objects.
- Result: every hit page belonged to **`cursorDiskKV`**, mostly as **overflow** pages.

3) **Schema confirmation**
- I enumerated tables and found:
  - `ItemTable`
  - `cursorDiskKV`
- Verified schema:
  - `CREATE TABLE cursorDiskKV (key TEXT UNIQUE ON CONFLICT REPLACE, value BLOB)`

4) **Proving it contains our conversation**
- Queried `cursorDiskKV` for keys containing our conversation UUID:
  - Found many keys shaped like:
    - `composerData:<composerId>`
    - `bubbleId:<composerId>:<bubbleId>`
    - `checkpointId:<composerId>:<checkpointId>`
    - `codeBlockDiff:<composerId>:<uuid>`
- Confirmed that one `bubbleId:<composerId>:<bubbleId>` entry contains our original prompt text and a `requestId` equal to the generation UUID.
- Confirmed at least one large bubble payload contains nested fields like `toolFormerData.result` and `codeBlocks[0].content` holding our diary content and diffs (i.e. the “binary match” source).

5) **Export for manual inspection**
- I restored/added a dedicated exporter script (storing it in ticket `scripts/`, per your rule):
  - `scripts/export_cursordiskkv_conversation_to_markdown.py`
- Then I exported our conversation’s `cursorDiskKV` view to Markdown:
  - `reference/05-global-state-vscdb-cursorDiskKV-composer-59f64e5e-export.md`
  - Includes:
    - `composerData:59f64e5e-...`
    - first 120 bubbles (by header list order)
    - a checkpoint key index (first 40)
    - “large string path” extraction to surface tool outputs/diffs embedded in bubble payloads

### Why this matters
- `cursorDiskKV` is almost certainly the missing persistence layer for **Composer/Agent** mode.
- It explains the earlier mismatch:
  - `grep` sees the bytes (in `cursorDiskKV` overflow pages),
  - while `SELECT ... FROM ItemTable` does not.

### Open questions (next probes)
- What do `type=1` vs `type=2` bubble payloads correspond to, exactly (user vs assistant vs tool events)?
- Where is the full assistant natural-language response stored (some type=2 bubbles have empty `text`, but contain large nested results/diffs)?
- Do `checkpointId:*` entries contain a complete transcript snapshot, or only state diffs?

---

## Step 17: Migration + Safety — non-destructive rsync into `go-go-labs/ttmp` and committing frequently

At this point the investigation produced a lot of artifacts (diary, analysis, scripts, exports). To avoid losing work and to align with your workflow, I moved the entire ticket folder into your real git repo and committed it immediately.

### What I did
- Located a real git repo containing `ttmp/`:
  - `/home/manuel/code/wesen/corporate-headquarters/go-go-labs`
- Per your instruction, migrated our `ttmp/` content using **non-destructive** rsync:
  - **Source**: `/home/manuel/workspaces/2025-12-15/inspect-agent-conversations/ttmp/`
  - **Dest**: `/home/manuel/code/wesen/corporate-headquarters/go-go-labs/ttmp/`
  - Used `--backup` with a timestamp suffix so if anything already existed at the destination it would be preserved with a `.bak-<timestamp>` suffix.
  - Did a `--dry-run` first, then the real sync.
- Staged and committed only the ticket folder inside `go-go-labs`:
  - `ttmp/2025/12/15/REVERSE-CURSOR-CONV-DB-001--reverse-engineer-cursor-conversation-database-format/`
  - Commit: `876a677`

### Why
- The original workspace folder was not a git repo; committing wasn’t possible there.
- Moving into `go-go-labs` ensures the work is durable and lets us commit more often going forward.

### What I learned
- `go-go-labs` already had many unrelated untracked/modified files, so being surgical about staging only our ticket folder was important.




