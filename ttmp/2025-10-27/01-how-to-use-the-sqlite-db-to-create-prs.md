# How to Use the SQLite History DB to Slice Clean PRs

This note summarises what worked (and what did not) while turning the messy `task/add-gpt5-responses-to-geppetto` branch into a tidier PR stack. Keep it next to the other helpers in `geppetto/ttmp/2025-10-23/`:

- `git-history-and-code-index.db` — the SQLite database described here.
- `git-history-index-guide.md` — mechanics for rebuilding the DB or running common queries.
- `feature-history-timeline.md` — narrative walkthrough of the original branch.
- `pr-extraction-guide.md` — proposed PR ordering and general cherry-picking tactics.

The workflow below assumes you have those files in place.

---

## 1. Understand the Schema

The DB captures both git history and our manual analysis. Key tables:

| Table | Purpose |
| ----- | ------- |
| `commits` | One row per commit (hash, subject, author, committed_at). |
| `files` | One row per tracked path. |
| `commit_files` | Many-to-many link: which file changed in which commit. |
| `commit_symbols` | Symbol-level data (function/type names) when available. |
| `analysis_notes` | Manual annotations we add while dissecting the branch. |
| `prs` | Candidate PR slices with name, description, status. |
| `pr_changelog` | Fine-grained actions (port, docs, etc.) tied to a PR/file/commit. |

Load the DB in read-only mode when you only need info:

```bash
sqlite3 -readonly geppetto/ttmp/2025-10-23/git-history-and-code-index.db
```

For quick reference, run `.schema <table_name>` to see column layouts.

---

## 2. Locate the Commits/Files You Need

Use the DB to scope each PR before touching the clean repo (`geppetto-clean/`).

Examples:

```sql
-- List commits touching tool executor code near the end of the branch
SELECT substr(c.hash,1,8) AS hash, c.subject, c.committed_at
FROM commit_files cf
JOIN commits c ON cf.commit_id = c.id
JOIN files f ON cf.file_id = f.id
WHERE f.path LIKE 'pkg/inference/tools/%'
  AND c.committed_at BETWEEN '2025-10-21' AND '2025-10-24'
ORDER BY c.committed_at;

-- Inspect which symbols changed in a noisy file
SELECT substr(c.hash,1,8), f.path, cs.symbol_name, cs.symbol_kind
FROM commit_symbols cs
JOIN commits c ON cs.commit_id = c.id
JOIN files f ON cs.file_id = f.id
WHERE f.path = 'pkg/events/chat-events.go';
```

Combine these with the narrative from `feature-history-timeline.md` to decide what belongs together.

---

## 3. Record Decisions While You Work

The `analysis_notes`, `prs`, and `pr_changelog` tables become the long-term memory that we kept missing in earlier attempts.

*When you start a new PR slice*:

```sql
INSERT INTO prs (name, description, status, updated_at)
VALUES ('PR03-tool-executor', 'Generic tool executor abstraction with context-aware tools', 'in-progress', datetime('now'));
```

*When you port something or intentionally skip it*, log it:

```sql
INSERT INTO pr_changelog (pr_id, commit_id, file_id, action, details)
VALUES (3, 799, 695, 'port', 'Brought over BaseToolExecutor from commit b21e6f91.');
```

*When you discover something noteworthy*, add a manual note:

```sql
INSERT INTO analysis_notes (commit_id, file_id, note_type, note, tags)
VALUES (799, 695, 'manual-review', 'BaseToolExecutor supplies the retry + event hooks; no debug tap dependency.', 'PR03,tool-executor');
```

Updating these tables as you go made later reconciliation far easier:

- We could sanity-check whether the clean PR still matched the messy branch (`diff` vs. the files mentioned in `pr_changelog`).
- We avoided redoing research because the `analysis_notes` already told us *why* a file mattered.

---

## 4. Build the Clean Slice in `geppetto-clean/`

With queries in hand, create a `pr/<topic>` branch in `geppetto-clean/` and port only what the DB says you need:

1. `git checkout -b pr/tool-executor main`
2. Use `git show <commit> -- <paths> | git apply` or `git cherry-pick <commit> -- <paths>` based on the guidance from `pr-extraction-guide.md`.
3. Run `gofmt`, `GOWORK=off go test ./...`, and `LEFTHOOK=0 make lint` as needed.
4. Stage & commit once the slice stands alone.
5. Update the DB (`pr_changelog`, `analysis_notes`, `prs.status`) before moving on.

When documentation pieces existed already (e.g., the event/ tools topics), copying them from the messy repo kept mainline docs synchronised without rewriting them by hand.

---

## 5. Lessons Learned

### What worked well
- **DB-driven scoping**: Queries made it obvious which files belonged to each feature and prevented us from over- or under-including changes.
- **Manual annotations** (`analysis_notes`, `pr_changelog`): investing a minute per change paid off later when reconciling branches or re-running lint/tests.
- **Document reuse**: copying the existing markdown from the messy repo (after verifying it matched the new PR scope) avoided drift in docs.
- **Turning off Lefthook when necessary**: `LEFTHOOK=0` kept commits flowing even when tests needed extra setup.

### What should be improved next time
- **Staying on the right branch**: amending commits on `main` after rebases caused confusion. Double-check `git branch --show-current` before amending.
- **Bulk copying code**: we briefly overwrote event types with a slimmed down version from the clean repo, which broke tool-event consumers. Always diff against the target branch before copying large files.
- **Tracking untracked directories**: note the extra directories (`e2e-responses-runner/`, `llm-runner/`, `t/`, `ttmp/...`) before merging so they do not surprise you later.

### Suggested future enhancements
- Automate the “port & log” steps via a helper script that records inserts into `pr_changelog`.
- Add a view that joins `prs` with their latest notes to see at a glance what is left for each PR.
- Capture test commands in `pr_changelog` (action `test`) so reviewers know what has already been run.

---

## 6. Reference Commands

To keep everything discoverable, the snippet below lists commands we used most often:

```bash
# table list
sqlite3 geppetto/ttmp/2025-10-23/git-history-and-code-index.db ".tables"

# walls of text go to markdown notes (e.g., feature-history-timeline.md)

# diff the clean branch against messy branch
diff -u geppetto/pkg/events/chat-events.go geppetto-clean/pkg/events/chat-events.go

# run tests without triggering Lefthook
LEFTHOOK=0 go test ./...

# record manual log
sqlite3 geppetto/ttmp/2025-10-23/git-history-and-code-index.db \
"INSERT INTO analysis_notes (...) VALUES (...);"
```

Follow the pattern: query first, port second, document third. That rhythm let us build PR02 (event registry) and PR03 (tool executor) cleanly, and will help with the remaining slices.***
