---
Title: Animal Website - Analysis, API Sketch, and UI Mockups
Ticket: 001-ANIMAL-PAGE
Status: draft
Topics:
    - golang
    - webapp
    - sqlite
    - htmx
    - templ
    - csv
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../experiments/animal-website
      Note: Implementation root (new experiment)
    - Path: ../../../../../../go.mod
      Note: Module dependencies (must use go-go-labs/go.mod)
ExternalSources: []
Summary: >
  Design and API sketch for a minimal Go + SQLite web app that uploads a CSV of animal names and
  renders the persisted animal list (htmx + bootstrap + templ).
LastUpdated: 2025-12-24T00:00:00Z
---

# Animal Website - Analysis, API Sketch, and UI Mockups

## Executive Summary

We’ll implement a small Go web server in `go-go-labs/experiments/animal-website` that:

- Serves an **Animals** page showing the current animal list from SQLite.
- Serves an **Upload CSV** page that posts a CSV file; the server parses it and **persists animal names in SQLite**.
- Uses **templ** for HTML and **htmx** to make upload and “refresh list” interactions smooth.

SQLite driver choice:

- Prefer **`modernc.org/sqlite`** (pure Go, no CGO) via `database/sql` driver name `sqlite`.
- Alternative is `github.com/mattn/go-sqlite3` (CGO) via driver name `sqlite3`.

## Problem Statement

We want a tiny “animal list” webapp where the source of truth is a user-uploaded CSV, not hardcoded data. The app must:

- Accept a CSV upload with animal names.
- Persist parsed names into a database.
- Render the current list in a web page.

Constraints:

- Go code must use the existing module `go-go-labs/go.mod` (no new `go.mod`).
- Implementation location: `go-go-labs/experiments/animal-website`.
- Database: SQLite.
- Web stack preference: **htmx + bootstrap + templ**.

## UI Sketch (ASCII)

### Animals Page (`GET /animals`)

```
+----------------------------------------------------------------------------------+
| Animal Website                                                                   |
|----------------------------------------------------------------------------------|
| [ Upload CSV ]  [ Clear list ]                                                   |
|----------------------------------------------------------------------------------|
| Animals (N)                                                                      |
|----------------------------------------------------------------------------------|
| #   | Name                                                                       |
|-----+----------------------------------------------------------------------------|
| 1   | cat                                                                        |
| 2   | dog                                                                        |
| 3   | capybara                                                                   |
| ... | ...                                                                        |
+----------------------------------------------------------------------------------+
```

Notes:

- “Upload CSV” navigates to `/upload`.
- “Clear list” triggers `POST /animals/clear` (htmx) and returns updated list fragment (or redirects).

### Upload Page (`GET /upload`)

```
+----------------------------------------------------------------------------------+
| Upload animals CSV                                                               |
|----------------------------------------------------------------------------------|
| Choose a CSV file: [__________________________] (Browse...)                      |
|                                                                                  |
| CSV format:                                                                      |
| - One animal name per line, OR                                                   |
| - First column is the animal name                                                |
|                                                                                  |
| Import mode: ( ) Replace list   ( ) Append                                       |
|                                                                                  |
| [ Upload ]                                                                       |
|----------------------------------------------------------------------------------|
| After upload: show success/error + link back to Animals                           |
+----------------------------------------------------------------------------------+
```

## Data Model / Schema (SQLite)

We’ll store a normalized list of animals. Minimal schema:

```sql
CREATE TABLE IF NOT EXISTS animals (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  name        TEXT NOT NULL,
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_animals_name_unique ON animals(name);
```

Import semantics (choose one):

- **Replace**: `DELETE FROM animals;` then insert parsed names.
- **Append**: insert parsed names (duplicates ignored by `INSERT OR IGNORE`).

The UI can offer a radio button; default to **Replace** (simplest mental model).

## CSV Parsing Rules

CSV ingestion should be forgiving:

- Accept `text/csv` uploads via `multipart/form-data`.
- Parse with `encoding/csv`.
- For each row:
  - Take first column as candidate name.
  - `strings.TrimSpace`.
  - Skip empty names.
  - Optionally `strings.ToLower` (decision) to normalize; or preserve original but enforce case-insensitive uniqueness (harder).

Suggested default: **trim + preserve case**, and uniqueness is exact match; document that in UI/help text.

## API Sketch

HTML-first endpoints (templ output), with optional htmx partials.

### Pages

- `GET /` → redirect to `/animals`
- `GET /animals`
  - returns full HTML page (layout + list)
- `GET /upload`
  - returns full HTML page with upload form

### Actions

- `POST /upload` (multipart/form-data)
  - inputs:
    - `file`: CSV file
    - `mode`: `replace|append` (optional; default `replace`)
  - behavior:
    - parse CSV
    - insert into SQLite according to mode
  - response:
    - non-htmx: `303 See Other` to `/animals`
    - htmx: return updated list fragment (and maybe a small “Imported X animals” banner)

- `POST /animals/clear`
  - behavior: delete all rows
  - response:
    - non-htmx: `303` to `/animals`
    - htmx: return empty list fragment

### Suggested htmx wiring

- On Animals page:
  - Clear button: `hx-post="/animals/clear" hx-target="#animals-list" hx-swap="outerHTML"`
  - List container has `id="animals-list"` so server can return just that fragment.

- On Upload page:
  - Form posts to `/upload`.
  - Option A (simple): normal POST + redirect.
  - Option B (htmx): `hx-post="/upload" hx-encoding="multipart/form-data" hx-target="#upload-result"`.

## Proposed Implementation Structure (Go packages)

Under `go-go-labs/experiments/animal-website/`:

```
animal-website/
  cmd/animal-website/          # main package
  internal/
    app/                       # wiring: routes, server, config
    db/                        # open DB + migrations
    animals/                   # repo + service (import/list/clear)
    httpui/                    # handlers, htmx helpers
    ui/                        # templ components
  static/
    app.css
    app.js
```

Notes:

- Use `github.com/rs/zerolog` and add `--log-level`.
- Serve static under `/static/` using `go:embed`.
- Use `github.com/a-h/templ` for templates, assume `templ generate -watch` is running.

## Design Decisions

- **HTML-first**: server renders HTML; no JSON API necessary for MVP.
- **SQLite**: single-file DB for easy local development and deployment.
- **htmx**: optional progressive enhancement; core flows still work without JS.
- **Import mode default**: Replace (deterministic “the CSV is the truth”).

## Alternatives Considered

- **JSON API + SPA**: overkill for “list + upload” and slower to build.
- **Store the CSV blob**: unnecessary unless we need provenance/versioning.
- **Pure file storage**: violates “save to a db” requirement.

## Implementation Plan (high-level)

- Scaffold minimal server + templ layout.
- Add SQLite schema + repo/service.
- Implement upload parsing + import.
- Render animals page + upload page; wire htmx optional enhancements.
- Add basic tests for CSV parsing + repo operations.

## Open Questions

- Should animal name uniqueness be case-insensitive (e.g. “Cat” == “cat”)?
- Should upload always **replace** without offering append, to keep UX simple?
- Do we want to show “invalid rows” feedback (line numbers) on upload failure?

## References

- Go `encoding/csv` docs (standard library)
- templ: `github.com/a-h/templ`
- htmx: `https://htmx.org/` (use in docs as a reference only)


