---
Title: Animal Website (CSV upload → SQLite → Web UI)
Ticket: 001-ANIMAL-PAGE
Status: draft
Topics:
    - golang
    - webapp
    - sqlite
    - htmx
    - templ
    - csv
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../experiments/animal-website
      Note: Implementation root (new experiment)
    - Path: ../../../../../go.mod
      Note: Module dependencies (must use go-go-labs/go.mod)
    - Path: ../../../../../go.sum
      Note: Module lockfile
    - Path: ../../../../../ttmp/_guidelines/design-doc.md
      Note: Template/guidelines for design docs
ExternalSources: []
Summary: >
  Build a small Go web app that accepts a CSV upload of animal names, persists them in SQLite,
  and renders the resulting animal list in a browser (htmx + bootstrap + templ).
LastUpdated: 2025-12-24T00:00:00Z
---

# Animal Website (CSV upload → SQLite → Web UI)

Document workspace for `001-ANIMAL-PAGE`.

## Overview

Build an app (under `go-go-labs/experiments/animal-website`) that:

- Accepts a CSV upload containing animal names
- Saves parsed animal names into SQLite
- Renders the current list of animals in a web page

## Key Links

- **Analysis**: `analysis/01-animal-website-analysis.md`
- **Tasks**: `tasks.md`
- **Changelog**: `changelog.md`

## Status

Current status: **draft**

## Topics

- Go HTTP server + templ templates
- htmx-driven CSV upload UX
- SQLite schema + import semantics (replace vs append)

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.


