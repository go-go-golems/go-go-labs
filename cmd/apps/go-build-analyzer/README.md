## Go Build Analyzer

A toolexec wrapper and CLI that logs every Go tool invocation (compile, asm, pack, link, etc.) to SQLite and provides Glazed-based commands to explore runs, invocations, and performance stats.

This README gives you a fast start and a tour of the features with runnable examples.

### What you get

- **Toolexec wrapper**: Transparent, best-effort logging without changing build outcomes.
- **SQLite backend (pure Go)**: Uses `modernc.org/sqlite`; no CGO required.
- **Runs and invocations**: Group invocations per build “run” with optional comments.
- **Rich metadata**: Package `-p`, outputs, importcfg, buildid, Go/lang versions, OS/arch, CWD, concurrency, flags JSON, elapsed time, status.
- **Glazed CLI**: Structured output with `--output table|json|yaml|csv`, `--fields`, `--sort-columns`.

---

## Prerequisites

- Go 1.24+ (workspace uses a `go.work` with 1.24.3)
- Linux/macOS shell with `bash`, `sed`, `jq` (for script demos)

Optional:
- `sqlite3` CLI for ad‑hoc SQL queries

---

## Install / Build

From the repo root:

```bash
cd go-go-labs/cmd/apps/go-build-analyzer
go build -o ./go-build-analyzer .
```

You can also run the helper:

```bash
./scripts/01-build-binary.sh
```

The binary path printed by the script can be used with `-toolexec`.

---

## Quick Start

1) Create a run and export environment variables:

```bash
./scripts/02-new-run-and-export.sh
# prints TOOLEXEC_DB=... and TOOLEXEC_RUN_ID=...
```

- `TOOLEXEC_DB` (optional): path to `build_times.db` (defaults to repo root if unset)
- `TOOLEXEC_RUN_ID`: groups subsequent invocations under a single run

2) Perform a clean, instrumented build across both modules:

```bash
./scripts/03-instrumented-build.sh
```

3) Explore data:

```bash
# List runs
./go-build-analyzer runs-list --output table

# Top packages by compile time (latest run)
./scripts/11-top-packages.sh

# Recent compile invocations (table + json)
./scripts/12-invocations-sample.sh

# Filter by package
./scripts/13-invocations-by-pkg.sh github.com/go-go-golems/glazed/pkg/cmds/parameters

# Export JSONL for downstream tools
./scripts/14-export-jsonl.sh
```

At any time, switch output formats:

```bash
./go-build-analyzer stats-packages --run-id 3 --tool compile --limit 30 --output json
./go-build-analyzer invocations-list --run-id 3 --tool compile --limit 10 --output csv
```

---

## Command Tour

All commands support Glazed output flags such as `--output`, `--fields`, and `--sort-columns`.

- `runs-new`
  - Create a new run. Use `--comment` to annotate and `--print-env` to echo an `export TOOLEXEC_RUN_ID=...` snippet.
  - Example:
    ```bash
    ./go-build-analyzer runs-new --comment "full rebuild" --print-env --output table
    ```

- `runs-list`
  - List runs with `run_id`, timestamp, and comment.
  - Example:
    ```bash
    ./go-build-analyzer runs-list --output table
    ```

- `invocations-list`
  - List tool invocations with rich metadata. Filters: `--run-id`, `--tool`, `--pkg`, `--limit`.
  - Useful fields include: `tool`, `tool_path`, `pkg`, `status`, `elapsed_ms`, `os`, `arch`, `cwd`, `out`, `importcfg`, `embedcfg`, `buildid`, `goversion`, `lang`, `concurrency`, `complete`, `pack`, `source_count`, `flags_json`, `args`.
  - Example:
    ```bash
    ./go-build-analyzer invocations-list --run-id 3 --tool compile --limit 20 --output table
    ```

- `stats-packages`
  - Aggregate elapsed times by package for a tool (default `compile`).
  - Filters: `--run-id`, `--tool`, `--limit`.
  - Example:
    ```bash
    ./go-build-analyzer stats-packages --run-id 3 --tool compile --limit 30 --output table
    ```

---

## Data Model

Two tables are created in SQLite:

### `runs`
- `id` (PK), `ts_unix`, `comment`

### `invocations`
- Core: `id`, `run_id`, `ts_unix`, `tool`, `status`, `elapsed_ms`, `args`
- Identity: `tool_path`, `pkg`
- Platform/context: `os`, `arch`, `cwd`
- Compile/link flags: `out`, `importcfg`, `embedcfg`, `buildid`, `goversion`, `lang`, `concurrency`, `complete`, `pack`, `source_count`
- Raw parsed flags as JSON: `flags_json`

Foreign keys and indexes make querying by `run_id`, `tool`, `pkg`, and time efficient.

---

## Using with `go build -toolexec`

You can run the analyzer directly as the toolexec program. Ensure `TOOLEXEC_DB` and (optionally) `TOOLEXEC_RUN_ID` are set.

```bash
export TOOLEXEC_DB="$(pwd)/build_times.db"
export TOOLEXEC_RUN_ID=123
go clean -cache
go build -a -toolexec="$(pwd)/go-build-analyzer" ./...
```

The wrapper will:
- Run the real tool binary
- Measure wall time
- Parse flags (e.g., `-p`, `-o`, `-importcfg`, `-buildid`, `-lang`, `-goversion`, `-c`)
- Record a row in `invocations` without affecting the exit code

---

## Using with `go test`

You can benchmark build time incurred by `go test`. The wrapper logs compile/link steps triggered by tests (including dependencies). Test execution time itself is not recorded.

### Minimal workflow

```bash
# 1) Choose DB and create a run; export RUN_ID for this shell
export TOOLEXEC_DB="$(pwd)/build_times.db"
eval "$(./go-build-analyzer runs-new --comment 'go test build' --print-env | grep '^export ')"

# 2) Avoid caches masking build work
go clean -cache -testcache

# 3a) Compile test binaries only (build-time focus)
go test -count=1 -c ./... -toolexec="$(pwd)/go-build-analyzer"

# 3b) Or run tests (still records the build steps)
go test -count=1 ./... -toolexec="$(pwd)/go-build-analyzer"
```

### Useful queries

```bash
# Top compile-time packages for this run
./go-build-analyzer stats-packages --run-id "$TOOLEXEC_RUN_ID" --tool compile --limit 50 --output table

# Recent compile/link steps
./go-build-analyzer invocations-list --run-id "$TOOLEXEC_RUN_ID" --tool compile --limit 50 --output table
./go-build-analyzer invocations-list --run-id "$TOOLEXEC_RUN_ID" --tool link --limit 50 --output table

# Approximate test-only binaries (linked test mains often end with _test)
./go-build-analyzer invocations-list --run-id "$TOOLEXEC_RUN_ID" --tool link --limit 500 --output json \
| jq -r '.[] | select(.out | test("_test$")) | {out, pkg, elapsed_ms, args}'

# Export everything for offline analysis
./go-build-analyzer invocations-list --run-id "$TOOLEXEC_RUN_ID" --limit 100000 --output json \
| jq -c '.[]' > "invocations-run-$TOOLEXEC_RUN_ID.jsonl"
```

### CI tips

- Use `-count=1` and `go clean -cache -testcache` to minimize cache effects.
- Prefer `go test -c` when you want to exclude test execution time and focus purely on build cost.
- Separate first-party vs third-party by filtering `pkg` prefixes in queries.

## Demo Scripts

Scripts live in `scripts/` and are safe to run repeatedly:

- `01-build-binary.sh`: Compile the analyzer.
- `02-new-run-and-export.sh`: Create a new run and export `TOOLEXEC_*` env vars.
- `03-instrumented-build.sh`: Clean cache and build both modules with `-toolexec` (continues on failures to capture logs).
- `10-query-runs.sh`: Show runs in table and JSON.
- `11-top-packages.sh`: Top packages by compile time for the latest run.
- `12-invocations-sample.sh`: Recent compile invocations (table + JSON).
- `13-invocations-by-pkg.sh <pkg>`: Filter invocations by package.
- `14-export-jsonl.sh`: Export invocations for the latest run as JSONL.

Example one‑liner tour:

```bash
./scripts/01-build-binary.sh && \
./scripts/02-new-run-and-export.sh && \
./scripts/03-instrumented-build.sh && \
./scripts/10-query-runs.sh && \
./scripts/11-top-packages.sh && \
./scripts/12-invocations-sample.sh && \
./scripts/14-export-jsonl.sh
```

---

## Ad‑hoc SQL

You can run custom SQL against the DB for dashboards and notebooks.

```sql
-- Top 30 packages by compile time for a given run
SELECT pkg, SUM(elapsed_ms) AS total_ms, COUNT(*) AS n
FROM invocations
WHERE tool = 'compile' AND run_id = :run
GROUP BY pkg
ORDER BY total_ms DESC
LIMIT 30;

-- Per-tool breakdown for the run
SELECT tool, SUM(elapsed_ms) AS total_ms, COUNT(*) AS n
FROM invocations
WHERE run_id = :run
GROUP BY tool
ORDER BY total_ms DESC;
```

---

## Tips, Caveats, Troubleshooting

- This tool logs best‑effort; it must never change build behavior. If logging fails, builds continue.
- Some experimental packages may fail to build (e.g., missing generated object files); logs are still recorded.
- `flags_json` contains all parsed flags for post‑hoc analysis. Consider extracting dimensions you need for dashboards.
- Use Glazed options to tailor output:
  - `--fields pkg,total_ms`
  - `--sort-columns -total_ms`
  - `--output json|yaml|csv|table`

---

## Learn More (Style & Help)

- Documentation style: see `glazed/pkg/doc/topics/how-to-write-good-documentation-pages.md`
- Glazed help: if you embed this in a larger app with help, you can follow patterns like:
  ```bash
  glaze help commands-reference
  glaze help layers-guide
  ```


