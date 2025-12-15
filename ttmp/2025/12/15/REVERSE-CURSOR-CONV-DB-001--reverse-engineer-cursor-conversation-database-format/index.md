---
Title: Reverse Engineer Cursor Conversation Database Format
Ticket: REVERSE-CURSOR-CONV-DB-001
Status: active
Topics:
    - reverse-engineering
    - data-analysis
    - exploration
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../.config/Cursor/User/globalStorage/state.vscdb
      Note: Primary Cursor conversation storage database (2.1GB SQLite key-value store)
    - Path: ../../../../../../../../.config/Cursor/User/workspaceStorage
      Note: Per-workspace state databases
    - Path: ../../../../../../../../.config/Cursor/User/workspaceStorage/94bae793ba109d83fb8934a587a6c719/state.vscdb
      Note: Per-workspace conversation storage containing aiService.generations
    - Path: ../../../../../../../../.config/Cursor/logs/20251212T173311/window8/output_20251215T091934/cursor.hooks.log
      Note: Hook execution logs containing conversation_id and generation_id matching stored data
    - Path: ../../../../../../../../.cursor/hooks.db
      Note: Primary SQLite database containing conversation event data
    - Path: ../../../../../../../../.cursor/hooks.json
      Note: Hook configuration file
    - Path: ../../../../../../../../.cursor/projects
      Note: Per-workspace storage directories
    - Path: ttmp/2025/12/15/REVERSE-CURSOR-CONV-DB-001--reverse-engineer-cursor-conversation-database-format/reference/02-workspace-state-vscdb-itemtable-export.md
      Note: Full ItemTable export of our workspace state.vscdb (73 keys)
    - Path: ttmp/2025/12/15/REVERSE-CURSOR-CONV-DB-001--reverse-engineer-cursor-conversation-database-format/reference/03-global-state-vscdb-aiCodeTrackingLines-export.md
      Note: ItemTable export (filtered) of global aiCodeTrackingLines
    - Path: ttmp/2025/12/15/REVERSE-CURSOR-CONV-DB-001--reverse-engineer-cursor-conversation-database-format/reference/04-global-state-vscdb-composer-pane-ecc9-export.md
      Note: ItemTable export of the composerChatViewPane entry for our view id
    - Path: ttmp/2025/12/15/REVERSE-CURSOR-CONV-DB-001--reverse-engineer-cursor-conversation-database-format/reference/05-global-state-vscdb-cursorDiskKV-composer-59f64e5e-export.md
      Note: Exported cursorDiskKV composerData/bubbles/checkpoints for our conversation
    - Path: ttmp/2025/12/15/REVERSE-CURSOR-CONV-DB-001--reverse-engineer-cursor-conversation-database-format/scripts/artifact-agent-tools-7206d086-264b-46da-93b3-384c8fd90eb4.txt
      Note: Captured output artifact from earlier exploration command
    - Path: ttmp/2025/12/15/REVERSE-CURSOR-CONV-DB-001--reverse-engineer-cursor-conversation-database-format/scripts/dump_workspace_conversation.py
      Note: Script to dump per-workspace composer/aiService conversation storage
    - Path: ttmp/2025/12/15/REVERSE-CURSOR-CONV-DB-001--reverse-engineer-cursor-conversation-database-format/scripts/export_cursordiskkv_conversation_to_markdown.py
      Note: Export Composer/Agent conversation from global state.vscdb cursorDiskKV to Markdown
    - Path: ttmp/2025/12/15/REVERSE-CURSOR-CONV-DB-001--reverse-engineer-cursor-conversation-database-format/scripts/export_itemtable_to_markdown.py
      Note: Export state.vscdb ItemTable to Markdown for manual inspection
    - Path: ttmp/2025/12/15/REVERSE-CURSOR-CONV-DB-001--reverse-engineer-cursor-conversation-database-format/scripts/extract_aiCodeTrackingLines.py
      Note: Extract aiCodeTrackingLines entries for a composerId
    - Path: ttmp/2025/12/15/REVERSE-CURSOR-CONV-DB-001--reverse-engineer-cursor-conversation-database-format/scripts/grep_cursor_logs_for_conversation.sh
      Note: Search logs for conversation_id/generationUUIDs/phrases
    - Path: ttmp/2025/12/15/REVERSE-CURSOR-CONV-DB-001--reverse-engineer-cursor-conversation-database-format/scripts/list_workspace_keys.sh
      Note: List key sizes in workspace state.vscdb
    - Path: ttmp/2025/12/15/REVERSE-CURSOR-CONV-DB-001--reverse-engineer-cursor-conversation-database-format/scripts/scan_global_state_for_needles.py
      Note: Byte-level scan of global state.vscdb blobs for UUIDs/phrases
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-15T09:22:07.537722466-05:00
---







# Reverse Engineer Cursor Conversation Database Format

Document workspace for REVERSE-CURSOR-CONV-DB-001.
