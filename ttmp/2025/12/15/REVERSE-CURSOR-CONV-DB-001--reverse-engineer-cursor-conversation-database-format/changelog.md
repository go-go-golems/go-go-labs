# Changelog

## 2025-12-15

- Initial workspace created


## 2025-12-15

Completed comprehensive exploration and analysis of Cursor conversation database format. Created exploration diary and detailed analysis document documenting database schema, data model, and storage architecture.


## 2025-12-15

Corrected exploration: hooks.db was user experiment, not Cursor native storage. Found actual storage in state.vscdb (2.1GB) with chatdata and composerChatViewPane keys. Continuing search for UUID.


## 2025-12-15

Rewrote analysis document with corrected findings. Removed all hooks.db references (user experiment). Documented actual Cursor storage: state.vscdb (2.1GB) with chatdata key-value structure. UUID search inconclusive - conversation may not be persisted yet.


## 2025-12-15

Found actual conversation storage! Conversations stored per-workspace in workspaceStorage/{uuid}/state.vscdb. Keys: aiService.generations (prompts), aiService.prompts (simpler), composer.composerData (metadata with composerId=conversation UUID). Actual conversation UUID: 59f64e5e-70e9-4892-a9e3-69d3d7f4b42d. The anchor UUID aa8ad79b... is just text content, not a conversation ID.


## 2025-12-15

Found complete conversation storage structure! Conversations stored per-workspace in workspaceStorage/{uuid}/state.vscdb. Keys: aiService.generations (prompts with generationUUID), aiService.prompts (simpler), composer.composerData (metadata with composerId=conversation UUID). Actual conversation UUID: 59f64e5e-70e9-4892-a9e3-69d3d7f4b42d. Anchor UUID aa8ad79b... is just text content. Updated analysis with complete storage model.


## 2025-12-15

Completed analysis. Found complete conversation storage: per-workspace in workspaceStorage/{uuid}/state.vscdb with aiService.generations, aiService.prompts, and composer.composerData keys. Conversation UUID = composerId. Anchor UUID is just text content. Documented full storage model.


## 2025-12-15

Added reproducible scripts under ticket scripts/ (workspace dump, global blob needle scan, aiCodeTrackingLines extractor, log grep). Copied prior exploration artifact into scripts/.


## 2025-12-15

Diary: added Step 13 (repro runs + findings about AI Chat transcripts vs Composer prompts-only persistence). Scripts: extended README, added more artifact-agent-tools captures.


## 2025-12-15

Diary: added Step 14 (negative search proving Composer/Agent assistant replies not found in local stores; contrasts with AI Chat chatdata rawText persistence).


## 2025-12-15

Added ItemTable->Markdown exporter script and exported workspace/global state.vscdb subsets for manual inspection (references/02-04). Diary: added Step 15 explaining grep 'binary matches' vs ItemTable evidence.


## 2025-12-15

Deep dive: binary match traced to global state.vscdb cursorDiskKV table (via raw byte offsets + dbstat page ownership). Added exporter script + produced reference/05 cursorDiskKV conversation export. Diary: added Step 16.

