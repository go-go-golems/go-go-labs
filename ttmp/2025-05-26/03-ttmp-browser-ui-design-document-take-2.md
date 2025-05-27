---
title: "TTMP Browser – Strategic Design Overview"
date: "2025-05-26"
author: "manuel"
status: "concept"
tags: ["strategy", "ttmp", "information-architecture", "product"]
---

# TTMP Browser – Strategic Design Overview

## 1. Why build this?

Our `ttmp/` directory has grown into an invaluable knowledge garden—ideas, research notes, specs, and snippets captured in markdown files. But finding, connecting, and extending this knowledge has become painful:

* Too many files to memorise paths or names
* No consistent tagging or metadata leads to rediscovery problems
* Manual `grep` workflows slow down creative flow

The **TTMP Browser** turns the directory into a searchable, self-curating library so that inspiration, not friction, drives our work.

## 2. Guiding Principles

1. **Zero lock-in** – Files remain plain markdown+YAML; no database migrations, no proprietary formats.
2. **Speed matters** – Instant search results reinforce habit-forming use.
3. **Progressive enhancement** – Basic CLI and web view first; polish later.
4. **Local-first** – All processing happens on disk; network optional.
5. **Extensible by design** – Clear API boundaries so future plugins (e.g. AI summariser, git versioning) can slot in without rewrites.

## 3. North-Star Experience

> "In one keystroke I open the browser, type '_'ebpf'_' and within 100 ms I'm reading my notes from last month, tagged correctly, with a permalink I can share."

Key user stories:

* **Search** – As a developer, I can fuzzy-search titles, tags, and full text to locate any note in under a second.
* **Capture** – From the CLI I can `ttmp new "idea"` and immediately start writing with a pre-filled YAML header in the ttmp/YYYY-MM-DD/NN-IDEA.md file.
* **Organise** – I can batch-edit tags or move notes to new dates without breaking links.
* **Reflect** – I can star/favourite notes and add lightweight personal annotations.


## 5. Capability Map (High-Level)

| Capability | How we'll approach it | Later? |
|------------|----------------------|--------|
| Indexing   | bleve golang search engine | |
| Search     | Simple TF-IDF + fuzzy title matching | ML/semantic search |
| Metadata   | Mandatory YAML preamble enforced by tooling | ... |
| UI         | Bootstrap + vanilla JS, rendered via templ | ... |
| Automation | Cobra CLI (`ttmp new/search/tag`) | Browser/VSCode extensions |

## 6. Architectural Snapshot

```
+------------+        REST/JSON         +--------------+
|   Web UI   |  <-------------------->  |   Go Server  |
+------------+                          +--------------+
        ^                                       |
        | Static files                          |
        |                                       v
+------------------+      fsnotify      +----------------+
|  ttmp Directory  | <----------------> |  Index Engine  |
+------------------+                    +----------------+
```

*Single binary serves HTTP & CLI commands.* No separate services.


---
**In one sentence:** TTMP Browser is our lightweight, local-first knowledge base that turns a messy directory into a discoverable library without tying us to any new platform.* 