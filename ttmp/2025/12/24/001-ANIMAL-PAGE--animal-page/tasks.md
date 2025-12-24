---
Title: 001-ANIMAL-PAGE Task List
Ticket: 001-ANIMAL-PAGE
Status: draft
Topics:
    - golang
    - webapp
    - sqlite
    - htmx
    - templ
    - csv
DocType: task-list
Intent: short-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../experiments/animal-website
      Note: Implementation root (new experiment)
ExternalSources: []
Summary: >
  Implementation checklist for the animal CSV upload + SQLite persistence + web UI experiment.
LastUpdated: 2025-12-24T00:00:00Z
---

# 001-ANIMAL-PAGE Task List

## Tasks

- [ ] **Scaffold experiment**: create `go-go-labs/experiments/animal-website` with a `cmd/animal-website` entrypoint (use module `go-go-labs/go.mod`)
- [ ] **Logging + flags**: add zerolog + `--log-level`, `--listen-addr`, `--db-path`
- [ ] **DB init**: open SQLite db and run schema migration on startup (create `animals` table + unique index)
- [ ] **Animals repository**: implement `List(ctx)`, `Clear(ctx)`, `InsertMany(ctx, names, mode)`
- [ ] **CSV import**: parse uploaded CSV (`encoding/csv`), trim/validate names, decide replace vs append behavior
- [ ] **HTTP routes**:
  - [ ] `GET /` → redirect to `/animals`
  - [ ] `GET /animals` → full page render (templ)
  - [ ] `GET /upload` → upload form page (templ)
  - [ ] `POST /upload` → handle multipart upload, persist, redirect or return htmx fragment
  - [ ] `POST /animals/clear` → clear db, redirect or return htmx fragment
- [ ] **templ UI**:
  - [ ] Layout + navbar
  - [ ] Animals list component (usable as full page section and as htmx fragment)
  - [ ] Upload page form component (supports multipart)
- [ ] **Static assets**: add `static/app.css` (+ optional `static/app.js`) and serve under `/static/` via `go:embed`
- [ ] **htmx wiring**: enhance clear/upload flows with htmx (`hx-post`, `hx-target`, `hx-encoding="multipart/form-data"`)
- [ ] **Bootstrap styling**: use bootstrap classes for a clean layout (CDN or local, but keep custom CSS in `/static/`)
- [ ] **Tests**:
  - [ ] CSV parsing tests (edge cases: empty lines, whitespace, single-column/multi-column)
  - [ ] Repo tests against SQLite (in-memory or temp file)
- [ ] **Docs**: add a small README under `experiments/animal-website` describing how to run + CSV format expectations

## Completed

- [x] Ticket workspace created with analysis + tasks + changelog
- [x] Scaffold experiment structure: create directories and main.go entrypoint
- [x] Add logging (zerolog) and CLI flags (--log-level, --listen-addr, --db-path)
- [x] Create SQLite schema and migration logic
- [x] Implement animals repository (List, Clear, InsertMany)
- [x] Create templ templates (layout, animals list, upload form)
- [x] Implement HTTP handlers (GET /animals, GET /upload, POST /upload, POST /animals/clear)
- [x] Add CSV parsing and import logic
- [x] Add static assets (CSS) and serve via go:embed
- [x] Wire up htmx attributes for progressive enhancement
- [x] Add Bootstrap styling
- [x] Create README with usage instructions

## Notes

- Default import mode suggestion is **replace** (CSV becomes source of truth).


