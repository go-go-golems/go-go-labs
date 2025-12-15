## Scripts (REVERSE-CURSOR-CONV-DB-001)

All scripts in this folder are **read-only**: they only read from `~/.config/Cursor` (and optionally `~/.cursor`) and print results to stdout.

### Quick index

- **`dump_workspace_conversation.py`**: inspect per-workspace conversation persistence (`composer.composerData`, `aiService.generations`, `aiService.prompts`).
- **`scan_global_state_for_needles.py`**: scan `globalStorage/state.vscdb` for byte-level needles (works even when values are BLOBs).
- **`extract_aiCodeTrackingLines.py`**: parse `aiCodeTrackingLines` and filter records for a specific `composerId`.
- **`extract_aichat_chatdata.py`**: extract a readable transcript from `workbench.panel.aichat.view.aichat.chatdata`.
- **`list_workspace_keys.sh`**: list keys + sizes in a workspace `state.vscdb` (focus on composer / aiService).
- **`grep_cursor_logs_for_conversation.sh`**: grep Cursor logs for a `conversation_id` / `generation_id` / selected phrases.
- **`list_global_keys.py`**: list top keys by size in global `state.vscdb` (helps find likely transcript blobs).
- **`dump_state_key.py`**: dump a single `ItemTable` key (with light redaction for secret-looking JSON).
- **`scan_dir_for_needles.py`**: scan binary-ish storage dirs (LevelDB / Session Storage / etc.) for needles.
- **`scan_workspace_storage_for_keys.py`**: quickly survey many workspace DBs for keys like `chatdata`/`aiService`.
- **`export_itemtable_to_markdown.py`**: export `ItemTable` (all keys or filtered by `--key-like`) into a Markdown file for manual inspection.
- **`locate_sqlite_binary_matches.py`**: find raw byte offsets of a phrase in a SQLite file and map to page numbers (for “binary file matches” forensics).
- **`map_sqlite_pages_with_dbstat.py`**: map page numbers to owning sqlite objects via `dbstat`.
- **`export_cursordiskkv_conversation_to_markdown.py`**: export a Composer/Agent conversation from global `cursorDiskKV` (bubbleId/checkpointId keys) to Markdown.

### Example usage (our current conversation)

```bash
python3 scripts/dump_workspace_conversation.py \
  --workspace-db /home/manuel/.config/Cursor/User/workspaceStorage/94bae793ba109d83fb8934a587a6c719/state.vscdb \
  --composer-id 59f64e5e-70e9-4892-a9e3-69d3d7f4b42d

python3 scripts/extract_aiCodeTrackingLines.py \
  --global-db /home/manuel/.config/Cursor/User/globalStorage/state.vscdb \
  --composer-id 59f64e5e-70e9-4892-a9e3-69d3d7f4b42d \
  --limit 20

python3 scripts/scan_global_state_for_needles.py \
  --db /home/manuel/.config/Cursor/User/globalStorage/state.vscdb \
  --needle 59f64e5e-70e9-4892-a9e3-69d3d7f4b42d \
  --needle "hooks.db is actually not relevant"

bash scripts/list_workspace_keys.sh \
  /home/manuel/.config/Cursor/User/workspaceStorage/94bae793ba109d83fb8934a587a6c719/state.vscdb

bash scripts/grep_cursor_logs_for_conversation.sh \
  --logs-root /home/manuel/.config/Cursor/logs \
  --conversation-id 59f64e5e-70e9-4892-a9e3-69d3d7f4b42d
```

### Artifacts

During the expedition we sometimes ran commands that produced large outputs. Those are copied in here as:
- `artifact-agent-tools-*.txt`


