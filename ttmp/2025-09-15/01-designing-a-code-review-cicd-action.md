Awesome—here’s a **Go‑based GitHub Action** that preserves the configuration flexibility and context‑passing you liked in Claude Code, but routes the PR context to **your own tool** (HTTP service or CLI) and then **posts results back** as review comments / a summary.

---

## ✅ Grab the starter repo

**[Download the starter Action (zip)](sandbox:/mnt/data/go-agent-action-starter.zip)**

This is a ready‑to‑build Docker‑based Action:

```
go-agent-action-starter/
├─ action.yml
├─ Dockerfile
├─ go.mod
├─ cmd/agent-action/main.go
├─ README.md
└─ examples/
   ├─ review.yml
   └─ mention.yml
```

---

## What it does (at a glance)

1. **Collects PR context**

   * PR metadata (title, body, base/head, labels, head SHA)
   * Changed files (optionally **unified diff** + optional **file contents** with size caps)
   * Optional guidelines file (defaults to `CLAUDE.md` for compatibility)
   * Optional extra repo files via glob patterns (e.g. `docs/**/*.md,.github/*.md`)
   * Event payload/trigger info (actor, event name, mention text when relevant)
   * Built using the GitHub REST API and the official Go client (`go-github`).

2. **Sends one JSON payload to your tool**

   * Mode **`http`** → POST to your service (with optional headers/token)
   * Mode **`cmd`** → run your CLI and pipe JSON on **stdin**, parse JSON on **stdout**

3. **Posts results back to GitHub**

   * **Review** with batched inline comments (and optional `APPROVE` / `REQUEST_CHANGES` / `COMMENT`)
   * **Issue/PR comment**
   * **Job summary** via `$GITHUB_STEP_SUMMARY` (rendered on the run page).

> The Action is packaged as a **Docker container action**, which is one of the three supported ways to author actions (JavaScript, Composite, Docker).
> Inputs are plumbed in per Docker‑action conventions (using metadata and args/env).

---

## Inputs you control

(Full list in `action.yml`)

* **Triggers**

  * `trigger_phrase` (default: `@agent`)
  * `label_trigger`, `assignee_trigger` (optional)

* **Context shaping**

  * `guidelines_path` (default: `CLAUDE.md`)
  * `include_patch` (default: `true`)
  * `include_file_contents` (default: `false`)
  * `include_repo_globs` (comma‑sep)
  * `max_file_bytes` (default: `200000`)
  * `max_changed_files` (default: `200`)

* **How to call your tool**

  * `tool_mode` = `http` | `cmd`
  * HTTP: `tool_url`, `tool_method`, `tool_headers_json`, `tool_token`
  * CLI: `tool_cmd`, `tool_args_json`, `working_directory`

* **Where to post results**

  * `output_mode` = `review` | `comment` | `summary` | `stdout` | `review+summary`
  * `max_comments` cap (default: `30`)
  * `github_token` (optional; uses `GITHUB_TOKEN` by default)

---

## The JSON your tool receives

Your service/CLI gets one object with rich PR context:

```json
{
  "owner": "acme",
  "repo": "web",
  "number": 123,
  "title": "feat: add cache",
  "body": "…",
  "base_ref": "main",
  "head_ref": "feature/cache",
  "head_sha": "abcdef…",
  "user_login": "alice",
  "labels": ["backend", "perf"],
  "changed_files": [
    {
      "path": "pkg/cache/cache.go",
      "status": "modified",
      "patch": "@@ -16,6 +16,12 @@ …",          // if include_patch=true
      "additions": 12,
      "deletions": 2,
      "blob_url": "https://github.com/…",
      "raw_url": "https://raw.githubusercontent.com/…",
      "contents_b64": "…"                   // if include_file_contents=true and size <= max_file_bytes
    }
  ],
  "guidelines_b64": "…",                    // contents of CLAUDE.md (or your path)
  "extra_files": [
    {"path": "docs/adr/0001.md", "contents_b64": "…"}
  ],
  "triggered_by": "alice",
  "event_name": "pull_request",
  "trigger_text": "@agent please review",   // mention text if it triggered
  "run_id": "987654321"
}
```

> Changed files are retrieved using **“List pull request files”**; the modern fields for **inline review comments** use `line`, `start_line`, `side` and `start_side`, which the REST API accepts when creating comments/reviews.

### The JSON your tool returns

````json
{
  "summary_markdown": "### Review Summary\n- 5 files checked\n- 2 suggestions\n",
  "review_decision": "comment",            // approve | request_changes | comment (optional)
  "review_body": "Automated review feedback",
  "issue_comment": "High level notes…",    // optional separate PR comment
  "comments": [
    {
      "path": "pkg/cache/cache.go",
      "body": "Prefer `context.WithTimeout` here.\n\n```suggestion\nctx, cancel := context.WithTimeout(ctx, 2*time.Second)\n```\n",
      "line": 42,
      "side": "RIGHT",
      "start_line": 41,
      "start_side": "RIGHT",
      "subject_type": "line"
    }
  ]
}
````

* GitHub supports **suggested changes** inside review comments using fenced blocks with `suggestion`, including multi‑line.

---

## How to wire it up in your workflows

The zip includes two working examples (see `examples/`):

### 1) Review every PR automatically

```yaml
name: agent-review
on:
  pull_request:
    types: [opened, synchronize, reopened]

permissions:
  contents: read
  pull-requests: write

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          persist-credentials: false

      - name: Agent review
        uses: ./.  # after publishing, use your-org/agent-action@v1
        with:
          tool_mode: http
          tool_url: https://agent.internal.example/review
          tool_token: ${{ secrets.AGENT_TOKEN }}
          include_patch: true
          include_file_contents: false
          include_repo_globs: "README.md,.github/*.md"
          output_mode: review+summary
```

### 2) On‑demand via mention (e.g., `@agent`)

```yaml
name: agent-mention
on:
  issue_comment:
    types: [created]

permissions:
  contents: read
  pull-requests: write

jobs:
  run-on-mention:
    if: ${{ github.event.issue.pull_request && contains(github.event.comment.body, '@agent') }}
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          persist-credentials: false

      - name: Agent on mention
        uses: ./. 
        with:
          tool_mode: http
          tool_url: https://agent.internal.example/mention
          tool_token: ${{ secrets.AGENT_TOKEN }}
          output_mode: review+summary
```

> The Action posts **review comments** by creating a **review** (which can batch many draft comments) and optional state (`APPROVE`/`REQUEST_CHANGES`/`COMMENT`). For single PR‑timeline comments, it uses the issues API.
> Job summaries are written to `$GITHUB_STEP_SUMMARY` and shown on the run page.

---

## How it’s implemented (Go)

* **Action type:** Docker container action with `action.yml` & `Dockerfile`.
* **GitHub API client:** `github.com/google/go-github/v66` with token auth.
* **PR context gathering:**

  * Reads event name/path from `GITHUB_EVENT_NAME`/`GITHUB_EVENT_PATH` (standard runner env).
  * Determines PR number from `pull_request` / `issue_comment` payloads.
  * Pull request + files via REST (“List pull request files”).
  * Optional workspace reads (e.g., `CLAUDE.md`) and extra globbed files from the mounted workspace (`actions/checkout` mounts it at `/github/workspace`).
* **Output posting:**

  * Creates a **pull request review** with batched comments (modern `line`/`start_line` fields).
  * Writes a **job summary** to `$GITHUB_STEP_SUMMARY`.
  * (If you later need checks UI: the **Checks API** requires a GitHub App; otherwise use commit statuses. This starter sticks to reviews & comments.)

---

## Two ways to plug in “your tool”

### A) HTTP service (recommended)

* Set: `tool_mode: http`, `tool_url`, optional `tool_headers_json`, `tool_token`
* The Action **POSTs** JSON to your endpoint and expects the **JSON response** shown above.

Minimal test endpoint (pseudo):

```bash
# echo server example (Node/Go/…):
# parse req JSON -> compute -> return ToolResponse JSON
```

### B) CLI process

* Set: `tool_mode: cmd`, `tool_cmd`, `tool_args_json`
* The Action writes PR JSON to **stdin**, reads ToolResponse JSON from **stdout**.
* `working_directory` defaults to `/github/workspace`.
* The runtime image is **Alpine** (has `bash`), so shell wrapping is possible if you need it.

---

## Publishing the Action

1. Create a new GitHub repo, add these files, and push.
2. Create a release and tag, e.g. `v1`.
3. In downstream repos, reference it: `uses: your-org/agent-action@v1`.

> If you want to add a **prebuilt** image instead of building on the fly, you can publish the image and point `action.yml` at it; but the default “Dockerfile build per version” is the documented, common pattern.

---

## Notes on parity with “Claude Code action”

* **Trigger phrase** (`@agent`), **label/assignee triggers**, **guidelines file**, **diff + optional file contents**, **extra repo docs via globs**, **review comments + suggestions**, **summary** → covered here.
* If you later want advanced features from Claude Code (e.g., MCP servers, provider routing, or GitHub‑App checks UI), we can extend this skeleton—its JSON boundary to your tool is designed to be stable so you can evolve the backend without re‑wiring Actions.

---

## References used while building this

* **Create Docker actions** & **metadata syntax** for inputs/args.
* **Docker actions & inputs via args/INPUT\_**.
* **Pull request files & review comments API** (modern comment fields and review batching).
* **Suggested changes syntax** in PR reviews.
* **Job summaries** with `$GITHUB_STEP_SUMMARY`.
* **Go GitHub client** (`go-github`) usage.

---

## Next steps you can take

* Drop your service URL (or CLI) behind `tool_*` inputs and run the included example workflows.
* If you want me to mirror **additional** Claude Code inputs (e.g., more triggers or context knobs) one‑for‑one, tell me which ones—I'll fold them into this skeleton cleanly.

If you want me to publish this into a GitHub repo with a `v1` tag (and tweak naming, org, or defaults), I can generate that repo bundle for you right now.
